package nex

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"runtime"
	"slices"
	"time"
)

// PRUDPServer represents a bare-bones PRUDP server
type PRUDPServer struct {
	udpSocket                     *net.UDPConn
	clients                       *MutexMap[string, *PRUDPClient]
	PRUDPVersion                  int
	PRUDPMinorVersion             uint32
	IsQuazalMode                  bool
	IsSecureServer                bool
	SupportedFunctions            uint32
	accessKey                     string
	kerberosPassword              []byte
	kerberosTicketVersion         int
	kerberosKeySize               int
	FragmentSize                  int
	version                       *LibraryVersion
	datastoreProtocolVersion      *LibraryVersion
	matchMakingProtocolVersion    *LibraryVersion
	rankingProtocolVersion        *LibraryVersion
	ranking2ProtocolVersion       *LibraryVersion
	messagingProtocolVersion      *LibraryVersion
	utilityProtocolVersion        *LibraryVersion
	natTraversalProtocolVersion   *LibraryVersion
	prudpEventHandlers            map[string][]func(packet PacketInterface)
	clientRemovedEventHandlers    []func(client *PRUDPClient)
	connectionIDCounter           *Counter[uint32]
	pingTimeout                   time.Duration
	PasswordFromPID               func(pid *PID) (string, uint32)
	PRUDPv1ConnectionSignatureKey []byte
}

// OnData adds an event handler which is fired when a new DATA packet is received
func (s *PRUDPServer) OnData(handler func(packet PacketInterface)) {
	s.on("data", handler)
}

// OnDisconnect adds an event handler which is fired when a new DISCONNECT packet is received
//
// To handle a client being removed from the server, see OnClientRemoved which fires on more cases
func (s *PRUDPServer) OnDisconnect(handler func(packet PacketInterface)) {
	s.on("disconnect", handler)
}

// OnClientRemoved adds an event handler which is fired when a client is removed from the server
//
// Fires both on a natural disconnect and from a timeout
func (s *PRUDPServer) OnClientRemoved(handler func(client *PRUDPClient)) {
	// * "removed" events are a special case, so handle them separately
	s.clientRemovedEventHandlers = append(s.clientRemovedEventHandlers, handler)
}

func (s *PRUDPServer) on(name string, handler func(packet PacketInterface)) {
	if _, ok := s.prudpEventHandlers[name]; !ok {
		s.prudpEventHandlers[name] = make([]func(packet PacketInterface), 0)
	}

	s.prudpEventHandlers[name] = append(s.prudpEventHandlers[name], handler)
}

func (s *PRUDPServer) emit(name string, packet PRUDPPacketInterface) {
	if handlers, ok := s.prudpEventHandlers[name]; ok {
		for _, handler := range handlers {
			go handler(packet)
		}
	}
}

func (s *PRUDPServer) emitRemoved(client *PRUDPClient) {
	for _, handler := range s.clientRemovedEventHandlers {
		go handler(client)
	}
}

// Listen starts a PRUDP server on a given port
func (s *PRUDPServer) Listen(port int) {
	// * Ensure the server has a key for PRUDPv1 connection signatures
	if len(s.PRUDPv1ConnectionSignatureKey) != 16 {
		s.PRUDPv1ConnectionSignatureKey = make([]byte, 16)
		_, err := rand.Read(s.PRUDPv1ConnectionSignatureKey)
		if err != nil {
			panic(err)
		}
	}

	udpAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	socket, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		panic(err)
	}

	s.udpSocket = socket

	quit := make(chan struct{})

	for i := 0; i < runtime.NumCPU(); i++ {
		go s.listenDatagram(quit)
	}

	<-quit
}

func (s *PRUDPServer) listenDatagram(quit chan struct{}) {
	err := error(nil)

	for err == nil {
		err = s.handleSocketMessage()
	}

	quit <- struct{}{}

	panic(err)
}

func (s *PRUDPServer) handleSocketMessage() error {
	buffer := make([]byte, 64000)

	read, addr, err := s.udpSocket.ReadFromUDP(buffer)
	if err != nil {
		return err
	}

	discriminator := addr.String()

	client, ok := s.clients.Get(discriminator)

	if !ok {
		client = NewPRUDPClient(addr, s)
		client.startHeartbeat()

		// * Fail-safe. If the server reboots, then
		// * s.clients has no record of old clients.
		// * An existing client which has not killed
		// * the connection on it's end MAY still send
		// * DATA packets once the server is back
		// * online, assuming it reboots fast enough.
		// * Since the client did NOT redo the SYN
		// * and CONNECT packets, it's reliable
		// * substreams never got remade. This is put
		// * in place to ensure there is always AT
		// * LEAST one substream in place, so the client
		// * can naturally error out due to the RC4
		// * errors.
		// *
		// * NOTE: THE CLIENT MAY NOT HAVE THE REAL
		// * CORRECT NUMBER OF SUBSTREAMS HERE. THIS
		// * IS ONLY DONE TO PREVENT A SERVER CRASH,
		// * NOT TO SAVE THE CLIENT. THE CLIENT IS
		// * EXPECTED TO NATURALLY DIE HERE
		client.createReliableSubstreams(0)

		s.clients.Set(discriminator, client)
	}

	packetData := buffer[:read]
	readStream := NewStreamIn(packetData, s)

	var packets []PRUDPPacketInterface

	// * Support any packet type the client sends and respond
	// * with that same type. Also keep reading from the stream
	// * until no more data is left, to account for multiple
	// * packets being sent at once
	if bytes.Equal(packetData[:2], []byte{0xEA, 0xD0}) {
		packets, _ = NewPRUDPPacketsV1(client, readStream)
	} else {
		packets, _ = NewPRUDPPacketsV0(client, readStream)
	}

	for _, packet := range packets {
		go s.processPacket(packet)
	}

	return nil
}

func (s *PRUDPServer) processPacket(packet PRUDPPacketInterface) {
	packet.Sender().(*PRUDPClient).resetHeartbeat()

	if packet.HasFlag(FlagAck) || packet.HasFlag(FlagMultiAck) {
		s.handleAcknowledgment(packet)
		return
	}

	switch packet.Type() {
	case SynPacket:
		s.handleSyn(packet)
	case ConnectPacket:
		s.handleConnect(packet)
	case DataPacket:
		s.handleData(packet)
	case DisconnectPacket:
		s.handleDisconnect(packet)
	case PingPacket:
		s.handlePing(packet)
	}
}

func (s *PRUDPServer) handleAcknowledgment(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagMultiAck) {
		s.handleMultiAcknowledgment(packet)
		return
	}

	client := packet.Sender().(*PRUDPClient)

	substream := client.reliableSubstream(packet.SubstreamID())
	substream.ResendScheduler.AcknowledgePacket(packet.SequenceID())
}

func (s *PRUDPServer) handleMultiAcknowledgment(packet PRUDPPacketInterface) {
	client := packet.Sender().(*PRUDPClient)
	stream := NewStreamIn(packet.Payload(), s)
	sequenceIDs := make([]uint16, 0)
	var baseSequenceID uint16
	var substream *ReliablePacketSubstreamManager

	if packet.SubstreamID() == 1 {
		// * New aggregate acknowledgment packets set this to 1
		// * and encode the real substream ID in in the payload
		substreamID, _ := stream.ReadUInt8()
		additionalIDsCount, _ := stream.ReadUInt8()
		baseSequenceID, _ = stream.ReadUInt16LE()
		substream = client.reliableSubstream(substreamID)

		for i := 0; i < int(additionalIDsCount); i++ {
			additionalID, _ := stream.ReadUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	} else {
		// TODO - This is how Kinnay's client handles this, but it doesn't make sense for QRV? Since it can have multiple reliable substreams?
		// * Old aggregate acknowledgment packets always use
		// * substream 0
		substream = client.reliableSubstream(0)
		baseSequenceID = packet.SequenceID()

		for stream.Remaining() > 0 {
			additionalID, _ := stream.ReadUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	}

	// * MutexMap.Each locks the mutex, can't remove while reading.
	// * Have to just loop again
	substream.ResendScheduler.packets.Each(func(sequenceID uint16, pending *PendingPacket) bool {
		if sequenceID <= baseSequenceID && !slices.Contains(sequenceIDs, sequenceID) {
			sequenceIDs = append(sequenceIDs, sequenceID)
		}

		return false
	})

	// * Actually remove the packets from the pool
	for _, sequenceID := range sequenceIDs {
		substream.ResendScheduler.AcknowledgePacket(sequenceID)
	}
}

func (s *PRUDPServer) handleSyn(packet PRUDPPacketInterface) {
	client := packet.Sender().(*PRUDPClient)

	var ack PRUDPPacketInterface

	if packet.Version() == 0 {
		ack, _ = NewPRUDPPacketV0(client, nil)
	} else {
		ack, _ = NewPRUDPPacketV1(client, nil)
	}

	connectionSignature, err := packet.calculateConnectionSignature(client.address)
	if err != nil {
		logger.Error(err.Error())
	}

	client.reset()
	client.clientConnectionSignature = connectionSignature
	client.SourceStreamType = packet.SourceStreamType()
	client.SourcePort = packet.SourcePort()
	client.DestinationStreamType = packet.DestinationStreamType()
	client.DestinationPort = packet.DestinationPort()

	ack.SetType(SynPacket)
	ack.AddFlag(FlagAck)
	ack.AddFlag(FlagHasSize)
	ack.SetSourceStreamType(packet.DestinationStreamType())
	ack.SetSourcePort(packet.DestinationPort())
	ack.SetDestinationStreamType(packet.SourceStreamType())
	ack.SetDestinationPort(packet.SourcePort())
	ack.setConnectionSignature(connectionSignature)
	ack.setSignature(ack.calculateSignature([]byte{}, []byte{}))

	if ack, ok := ack.(*PRUDPPacketV1); ok {
		// * Negotiate with the client what we support
		ack.maximumSubstreamID = packet.(*PRUDPPacketV1).maximumSubstreamID // * No change needed, we can just support what the client wants
		ack.minorVersion = packet.(*PRUDPPacketV1).minorVersion             // * No change needed, we can just support what the client wants
		ack.supportedFunctions = s.SupportedFunctions & packet.(*PRUDPPacketV1).supportedFunctions
	}

	s.emit("syn", ack)

	s.sendRaw(client.address, ack.Bytes())
}

func (s *PRUDPServer) handleConnect(packet PRUDPPacketInterface) {
	client := packet.Sender().(*PRUDPClient)

	var ack PRUDPPacketInterface

	if packet.Version() == 0 {
		ack, _ = NewPRUDPPacketV0(client, nil)
	} else {
		ack, _ = NewPRUDPPacketV1(client, nil)
	}

	client.serverConnectionSignature = packet.getConnectionSignature()

	connectionSignature, err := packet.calculateConnectionSignature(client.address)
	if err != nil {
		logger.Error(err.Error())
	}

	ack.SetType(ConnectPacket)
	ack.AddFlag(FlagAck)
	ack.AddFlag(FlagHasSize)
	ack.SetSourceStreamType(packet.DestinationStreamType())
	ack.SetSourcePort(packet.DestinationPort())
	ack.SetDestinationStreamType(packet.SourceStreamType())
	ack.SetDestinationPort(packet.SourcePort())
	ack.setConnectionSignature(make([]byte, len(connectionSignature)))
	ack.SetSessionID(0)
	ack.SetSequenceID(1)

	if ack, ok := ack.(*PRUDPPacketV1); ok {
		// * At this stage the client and server have already
		// * negotiated what they each can support, so configure
		// * the client now and just send the client back the
		// * negotiated configuration
		ack.maximumSubstreamID = packet.(*PRUDPPacketV1).maximumSubstreamID
		ack.minorVersion = packet.(*PRUDPPacketV1).minorVersion
		ack.supportedFunctions = packet.(*PRUDPPacketV1).supportedFunctions

		client.minorVersion = ack.minorVersion
		client.supportedFunctions = ack.supportedFunctions
		client.createReliableSubstreams(ack.maximumSubstreamID)
	} else {
		client.createReliableSubstreams(0)
	}

	var payload []byte

	if s.IsSecureServer {
		sessionKey, pid, checkValue, err := s.readKerberosTicket(packet.Payload())
		if err != nil {
			logger.Error(err.Error())
		}

		client.SetPID(pid)
		client.setSessionKey(sessionKey)

		stream := NewStreamOut(s)

		// * The response value is a Buffer whose data contains
		// * checkValue+1. This is just a lazy way of encoding
		// * a Buffer type
		stream.WriteUInt32LE(4)              // * Buffer length
		stream.WriteUInt32LE(checkValue + 1) // * Buffer data

		payload = stream.Bytes()
	} else {
		payload = make([]byte, 0)
	}

	ack.SetPayload(payload)
	ack.setSignature(ack.calculateSignature([]byte{}, packet.getConnectionSignature()))

	s.emit("connect", ack)

	s.sendRaw(client.address, ack.Bytes())
}

func (s *PRUDPServer) handleData(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagReliable) {
		s.handleReliable(packet)
	} else {
		s.handleUnreliable(packet)
	}
}

func (s *PRUDPServer) handleDisconnect(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		s.acknowledgePacket(packet)
	}

	client := packet.Sender().(*PRUDPClient)

	client.cleanup() // * "removed" event is dispatched here
	s.clients.Delete(client.address.String())

	s.emit("disconnect", packet)
}

func (s *PRUDPServer) handlePing(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		s.acknowledgePacket(packet)
	}
}

func (s *PRUDPServer) readKerberosTicket(payload []byte) ([]byte, *PID, uint32, error) {
	stream := NewStreamIn(payload, s)

	ticketData, err := stream.ReadBuffer()
	if err != nil {
		return nil, nil, 0, err
	}

	requestData, err := stream.ReadBuffer()
	if err != nil {
		return nil, nil, 0, err
	}

	serverKey := DeriveKerberosKey(NewPID[uint64](2), s.kerberosPassword)

	ticket := NewKerberosTicketInternalData()
	err = ticket.Decrypt(NewStreamIn(ticketData, s), serverKey)
	if err != nil {
		return nil, nil, 0, err
	}

	ticketTime := ticket.Issued.Standard()
	serverTime := time.Now().UTC()

	timeLimit := ticketTime.Add(time.Minute * 2)
	if serverTime.After(timeLimit) {
		return nil, nil, 0, errors.New("Kerberos ticket expired")
	}

	sessionKey := ticket.SessionKey
	kerberos := NewKerberosEncryption(sessionKey)

	decryptedRequestData, err := kerberos.Decrypt(requestData)
	if err != nil {
		return nil, nil, 0, err
	}

	checkDataStream := NewStreamIn(decryptedRequestData, s)

	userPID, err := checkDataStream.ReadPID()
	if err != nil {
		return nil, nil, 0, err
	}

	_, err = checkDataStream.ReadUInt32LE() // * CID of secure server station url
	if err != nil {
		return nil, nil, 0, err
	}

	responseCheck, err := checkDataStream.ReadUInt32LE()
	if err != nil {
		return nil, nil, 0, err
	}

	return sessionKey, userPID, responseCheck, nil
}

func (s *PRUDPServer) acknowledgePacket(packet PRUDPPacketInterface) {
	var ack PRUDPPacketInterface

	if packet.Version() == 0 {
		ack, _ = NewPRUDPPacketV0(packet.Sender().(*PRUDPClient), nil)
	} else {
		ack, _ = NewPRUDPPacketV1(packet.Sender().(*PRUDPClient), nil)
	}

	ack.SetType(packet.Type())
	ack.AddFlag(FlagAck)
	ack.SetSourceStreamType(packet.DestinationStreamType())
	ack.SetSourcePort(packet.DestinationPort())
	ack.SetDestinationStreamType(packet.SourceStreamType())
	ack.SetDestinationPort(packet.SourcePort())
	ack.SetSequenceID(packet.SequenceID())
	ack.setFragmentID(packet.getFragmentID())
	ack.SetSubstreamID(packet.SubstreamID())

	s.sendPacket(ack)

	// * Servers send the DISCONNECT ACK 3 times
	if packet.Type() == DisconnectPacket {
		s.sendPacket(ack)
		s.sendPacket(ack)
	}
}

func (s *PRUDPServer) handleReliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		s.acknowledgePacket(packet)
	}

	substream := packet.Sender().(*PRUDPClient).reliableSubstream(packet.SubstreamID())

	for _, pendingPacket := range substream.Update(packet) {
		if packet.Type() == DataPacket {
			payload := substream.AddFragment(pendingPacket.decryptPayload())

			if packet.getFragmentID() == 0 {
				message := NewRMCMessage()
				err := message.FromBytes(payload)
				if err != nil {
					// TODO - Should this return the error too?
					logger.Error(err.Error())
				}

				substream.ResetFragmentedPayload()

				packet.SetRMCMessage(message)

				s.emit("data", packet)
			}
		}
	}
}

func (s *PRUDPServer) handleUnreliable(packet PRUDPPacketInterface) {}

func (s *PRUDPServer) sendPing(client *PRUDPClient) {
	var ping PRUDPPacketInterface

	if s.PRUDPVersion == 0 {
		ping, _ = NewPRUDPPacketV0(client, nil)
	} else {
		ping, _ = NewPRUDPPacketV1(client, nil)
	}

	ping.SetType(PingPacket)
	ping.AddFlag(FlagNeedsAck)
	ping.SetSourceStreamType(client.DestinationStreamType)
	ping.SetSourcePort(client.DestinationPort)
	ping.SetDestinationStreamType(client.SourceStreamType)
	ping.SetDestinationPort(client.SourcePort)
	ping.SetSubstreamID(0)

	s.sendPacket(ping)
}

// Send sends the packet to the packets sender
func (s *PRUDPServer) Send(packet PacketInterface) {
	if packet, ok := packet.(PRUDPPacketInterface); ok {
		data := packet.Payload()
		fragments := int(len(data) / s.FragmentSize)

		var fragmentID uint8 = 1
		for i := 0; i <= fragments; i++ {
			if len(data) < s.FragmentSize {
				packet.SetPayload(data)
				packet.setFragmentID(0)
			} else {
				packet.SetPayload(data[:s.FragmentSize])
				packet.setFragmentID(fragmentID)

				data = data[s.FragmentSize:]
				fragmentID++
			}

			s.sendPacket(packet)
		}
	}
}

func (s *PRUDPServer) sendPacket(packet PRUDPPacketInterface) {
	// * PRUDPServer.Send will send fragments as the same packet,
	// * just with different fields. In order to prevent modifying
	// * multiple packets at once, due to the same pointer being
	// * reused, we must make a copy of the packet being sent
	packetCopy := packet.Copy()
	client := packetCopy.Sender().(*PRUDPClient)

	if !packetCopy.HasFlag(FlagAck) && !packetCopy.HasFlag(FlagMultiAck) {
		if packetCopy.HasFlag(FlagReliable) {
			substream := client.reliableSubstream(packetCopy.SubstreamID())
			packetCopy.SetSequenceID(substream.NextOutgoingSequenceID())
		} else if packetCopy.Type() == DataPacket {
			packetCopy.SetSequenceID(client.nextOutgoingUnreliableSequenceID())
		} else if packetCopy.Type() == PingPacket {
			packetCopy.SetSequenceID(client.nextOutgoingPingSequenceID())
		} else {
			packetCopy.SetSequenceID(0)
		}
	}

	if packetCopy.Type() == DataPacket && !packetCopy.HasFlag(FlagAck) && !packetCopy.HasFlag(FlagMultiAck) {
		if packetCopy.HasFlag(FlagReliable) {
			substream := client.reliableSubstream(packetCopy.SubstreamID())
			packetCopy.SetPayload(substream.Encrypt(packetCopy.Payload()))
		}
		// TODO - Unreliable crypto
	}

	packetCopy.setSignature(packetCopy.calculateSignature(client.sessionKey, client.serverConnectionSignature))

	if packetCopy.HasFlag(FlagReliable) && packetCopy.HasFlag(FlagNeedsAck) {
		substream := client.reliableSubstream(packetCopy.SubstreamID())
		substream.ResendScheduler.AddPacket(packetCopy)
	}

	s.sendRaw(packetCopy.Sender().Address(), packetCopy.Bytes())
}

// sendRaw will send the given address the provided packet
func (s *PRUDPServer) sendRaw(conn net.Addr, data []byte) {
	_, err := s.udpSocket.WriteToUDP(data, conn.(*net.UDPAddr))
	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
	}
}

// AccessKey returns the servers sandbox access key
func (s *PRUDPServer) AccessKey() string {
	return s.accessKey
}

// SetAccessKey sets the servers sandbox access key
func (s *PRUDPServer) SetAccessKey(accessKey string) {
	s.accessKey = accessKey
}

// KerberosPassword returns the server kerberos password
func (s *PRUDPServer) KerberosPassword() []byte {
	return s.kerberosPassword
}

// SetKerberosPassword sets the server kerberos password
func (s *PRUDPServer) SetKerberosPassword(kerberosPassword []byte) {
	s.kerberosPassword = kerberosPassword
}

// SetFragmentSize sets the max size for a packets payload
func (s *PRUDPServer) SetFragmentSize(fragmentSize int) {
	// TODO - Derive this value from the MTU
	// * From the wiki:
	// *
	// * The fragment size depends on the implementation.
	// * It is generally set to the MTU minus the packet overhead.
	// *
	// * In old NEX versions, which only support PRUDP v0, the MTU is
	// * hardcoded to 1000 and the maximum payload size seems to be 962 bytes.
	// *
	// * Later, the MTU was increased to 1364, and the maximum payload
	// * size is seems to be 1300 bytes, unless PRUDP v0 is used, in which case it’s 1264 bytes.
	s.FragmentSize = fragmentSize
}

// SetKerberosTicketVersion sets the version used when handling kerberos tickets
func (s *PRUDPServer) SetKerberosTicketVersion(kerberosTicketVersion int) {
	s.kerberosTicketVersion = kerberosTicketVersion
}

// KerberosKeySize gets the size for the kerberos session key
func (s *PRUDPServer) KerberosKeySize() int {
	return s.kerberosKeySize
}

// SetKerberosKeySize sets the size for the kerberos session key
func (s *PRUDPServer) SetKerberosKeySize(kerberosKeySize int) {
	s.kerberosKeySize = kerberosKeySize
}

// LibraryVersion returns the server NEX version
func (s *PRUDPServer) LibraryVersion() *LibraryVersion {
	return s.version
}

// SetDefaultLibraryVersion sets the default NEX protocol versions
func (s *PRUDPServer) SetDefaultLibraryVersion(version *LibraryVersion) {
	s.version = version
	s.datastoreProtocolVersion = version.Copy()
	s.matchMakingProtocolVersion = version.Copy()
	s.rankingProtocolVersion = version.Copy()
	s.ranking2ProtocolVersion = version.Copy()
	s.messagingProtocolVersion = version.Copy()
	s.utilityProtocolVersion = version.Copy()
	s.natTraversalProtocolVersion = version.Copy()
}

// DataStoreProtocolVersion returns the servers DataStore protocol version
func (s *PRUDPServer) DataStoreProtocolVersion() *LibraryVersion {
	return s.datastoreProtocolVersion
}

// SetDataStoreProtocolVersion sets the servers DataStore protocol version
func (s *PRUDPServer) SetDataStoreProtocolVersion(version *LibraryVersion) {
	s.datastoreProtocolVersion = version
}

// MatchMakingProtocolVersion returns the servers MatchMaking protocol version
func (s *PRUDPServer) MatchMakingProtocolVersion() *LibraryVersion {
	return s.matchMakingProtocolVersion
}

// SetMatchMakingProtocolVersion sets the servers MatchMaking protocol version
func (s *PRUDPServer) SetMatchMakingProtocolVersion(version *LibraryVersion) {
	s.matchMakingProtocolVersion = version
}

// RankingProtocolVersion returns the servers Ranking protocol version
func (s *PRUDPServer) RankingProtocolVersion() *LibraryVersion {
	return s.rankingProtocolVersion
}

// SetRankingProtocolVersion sets the servers Ranking protocol version
func (s *PRUDPServer) SetRankingProtocolVersion(version *LibraryVersion) {
	s.rankingProtocolVersion = version
}

// Ranking2ProtocolVersion returns the servers Ranking2 protocol version
func (s *PRUDPServer) Ranking2ProtocolVersion() *LibraryVersion {
	return s.ranking2ProtocolVersion
}

// SetRanking2ProtocolVersion sets the servers Ranking2 protocol version
func (s *PRUDPServer) SetRanking2ProtocolVersion(version *LibraryVersion) {
	s.ranking2ProtocolVersion = version
}

// MessagingProtocolVersion returns the servers Messaging protocol version
func (s *PRUDPServer) MessagingProtocolVersion() *LibraryVersion {
	return s.messagingProtocolVersion
}

// SetMessagingProtocolVersion sets the servers Messaging protocol version
func (s *PRUDPServer) SetMessagingProtocolVersion(version *LibraryVersion) {
	s.messagingProtocolVersion = version
}

// UtilityProtocolVersion returns the servers Utility protocol version
func (s *PRUDPServer) UtilityProtocolVersion() *LibraryVersion {
	return s.utilityProtocolVersion
}

// SetUtilityProtocolVersion sets the servers Utility protocol version
func (s *PRUDPServer) SetUtilityProtocolVersion(version *LibraryVersion) {
	s.utilityProtocolVersion = version
}

// SetNATTraversalProtocolVersion sets the servers NAT Traversal protocol version
func (s *PRUDPServer) SetNATTraversalProtocolVersion(version *LibraryVersion) {
	s.natTraversalProtocolVersion = version
}

// NATTraversalProtocolVersion returns the servers NAT Traversal protocol version
func (s *PRUDPServer) NATTraversalProtocolVersion() *LibraryVersion {
	return s.natTraversalProtocolVersion
}

// ConnectionIDCounter returns the servers CID counter
func (s *PRUDPServer) ConnectionIDCounter() *Counter[uint32] {
	return s.connectionIDCounter
}

// FindClientByConnectionID returns the PRUDP client connected with the given connection ID
func (s *PRUDPServer) FindClientByConnectionID(connectedID uint32) *PRUDPClient {
	var client *PRUDPClient

	s.clients.Each(func(discriminator string, c *PRUDPClient) bool {
		if c.ConnectionID == connectedID {
			client = c
			return true
		}

		return false
	})

	return client
}

// FindClientByPID returns the PRUDP client connected with the given PID
func (s *PRUDPServer) FindClientByPID(pid uint64) *PRUDPClient {
	var client *PRUDPClient

	s.clients.Each(func(discriminator string, c *PRUDPClient) bool {
		if c.pid.pid == pid {
			client = c
			return true
		}

		return false
	})

	return client
}

// NewPRUDPServer will return a new PRUDP server
func NewPRUDPServer() *PRUDPServer {
	return &PRUDPServer{
		clients:             NewMutexMap[string, *PRUDPClient](),
		IsQuazalMode:        false,
		kerberosKeySize:     32,
		FragmentSize:        1300,
		prudpEventHandlers:  make(map[string][]func(PacketInterface)),
		connectionIDCounter: NewCounter[uint32](10),
		pingTimeout:         time.Second * 15,
	}
}
