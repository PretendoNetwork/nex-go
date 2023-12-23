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

	"github.com/PretendoNetwork/nex-go/compression"
	"github.com/PretendoNetwork/nex-go/types"
	"github.com/lxzan/gws"
)

// PRUDPServer represents a bare-bones PRUDP server
type PRUDPServer struct {
	udpSocket                       *net.UDPConn
	websocketServer                 *WebSocketServer
	PRUDPVersion                    int
	PRUDPMinorVersion               uint32
	virtualServers                  *MutexMap[uint8, *MutexMap[uint8, *MutexMap[string, *PRUDPClient]]]
	IsQuazalMode                    bool
	VirtualServerPorts              []uint8
	SecureVirtualServerPorts        []uint8
	SupportedFunctions              uint32
	accessKey                       string
	kerberosPassword                []byte
	kerberosTicketVersion           int
	kerberosKeySize                 int
	FragmentSize                    int
	version                         *LibraryVersion
	datastoreProtocolVersion        *LibraryVersion
	matchMakingProtocolVersion      *LibraryVersion
	rankingProtocolVersion          *LibraryVersion
	ranking2ProtocolVersion         *LibraryVersion
	messagingProtocolVersion        *LibraryVersion
	utilityProtocolVersion          *LibraryVersion
	natTraversalProtocolVersion     *LibraryVersion
	prudpEventHandlers              map[string][]func(packet PacketInterface)
	clientRemovedEventHandlers      []func(client *PRUDPClient)
	connectionIDCounter             *Counter[uint32]
	pingTimeout                     time.Duration
	passwordFromPIDHandler          func(pid *types.PID) (string, uint32)
	PRUDPv1ConnectionSignatureKey   []byte
	EnhancedChecksum                bool
	PRUDPv0CustomChecksumCalculator func(packet *PRUDPPacketV0, data []byte) uint32
	stringLengthSize                int
	CompressionAlgorithm            compression.Algorithm
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

// Listen is an alias of ListenUDP. Implemented to conform to the ServerInterface
func (s *PRUDPServer) Listen(port int) {
	s.ListenUDP(port)
}

// ListenUDP starts a PRUDP server on a given port using a UDP server
func (s *PRUDPServer) ListenUDP(port int) {
	s.initPRUDPv1ConnectionSignatureKey()
	s.initVirtualPorts()

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

// ListenWebSocket starts a PRUDP server on a given port using a WebSocket server
func (s *PRUDPServer) ListenWebSocket(port int) {
	s.initPRUDPv1ConnectionSignatureKey()
	s.initVirtualPorts()

	s.websocketServer = &WebSocketServer{
		prudpServer: s,
	}

	s.websocketServer.listen(port)
}

// ListenWebSocketSecure starts a PRUDP server on a given port using a secure (TLS) WebSocket server
func (s *PRUDPServer) ListenWebSocketSecure(port int, certFile, keyFile string) {
	s.initPRUDPv1ConnectionSignatureKey()
	s.initVirtualPorts()

	s.websocketServer = &WebSocketServer{
		prudpServer: s,
	}

	s.websocketServer.listenSecure(port, certFile, keyFile)
}

func (s *PRUDPServer) initPRUDPv1ConnectionSignatureKey() {
	// * Ensure the server has a key for PRUDPv1 connection signatures
	if len(s.PRUDPv1ConnectionSignatureKey) != 16 {
		s.PRUDPv1ConnectionSignatureKey = make([]byte, 16)
		_, err := rand.Read(s.PRUDPv1ConnectionSignatureKey)
		if err != nil {
			panic(err)
		}
	}
}

func (s *PRUDPServer) initVirtualPorts() {
	for _, port := range s.VirtualServerPorts {
		virtualServer := NewMutexMap[uint8, *MutexMap[string, *PRUDPClient]]()
		virtualServer.Set(VirtualStreamTypeDO, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeRV, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeOldRVSec, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeSBMGMT, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeNAT, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeSessionDiscovery, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeNATEcho, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeRouting, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeGame, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeRVSecure, NewMutexMap[string, *PRUDPClient]())
		virtualServer.Set(VirtualStreamTypeRelay, NewMutexMap[string, *PRUDPClient]())

		s.virtualServers.Set(port, virtualServer)
	}

	logger.Success("Virtual ports created")
}

func (s *PRUDPServer) listenDatagram(quit chan struct{}) {
	var err error

	for err == nil {
		buffer := make([]byte, 64000)
		var read int
		var addr *net.UDPAddr

		read, addr, err = s.udpSocket.ReadFromUDP(buffer)
		packetData := buffer[:read]

		err = s.handleSocketMessage(packetData, addr, nil)
	}

	quit <- struct{}{}

	panic(err)
}

func (s *PRUDPServer) handleSocketMessage(packetData []byte, address net.Addr, webSocketConnection *gws.Conn) error {
	readStream := NewStreamIn(packetData, s)

	var packets []PRUDPPacketInterface

	// * Support any packet type the client sends and respond
	// * with that same type. Also keep reading from the stream
	// * until no more data is left, to account for multiple
	// * packets being sent at once
	if s.websocketServer != nil && packetData[0] == 0x80 {
		packets, _ = NewPRUDPPacketsLite(nil, readStream)
	} else if bytes.Equal(packetData[:2], []byte{0xEA, 0xD0}) {
		packets, _ = NewPRUDPPacketsV1(nil, readStream)
	} else {
		packets, _ = NewPRUDPPacketsV0(nil, readStream)
	}

	for _, packet := range packets {
		go s.processPacket(packet, address, webSocketConnection)
	}

	return nil
}

func (s *PRUDPServer) processPacket(packet PRUDPPacketInterface, address net.Addr, webSocketConnection *gws.Conn) {
	if !slices.Contains(s.VirtualServerPorts, packet.DestinationPort()) {
		logger.Warningf("Client %s trying to connect to unbound server vport %d", address.String(), packet.DestinationPort())
		return
	}

	if packet.DestinationStreamType() > VirtualStreamTypeRelay {
		logger.Warningf("Client %s trying to use invalid to server stream type %d", address.String(), packet.DestinationStreamType())
		return
	}

	virtualServer, _ := s.virtualServers.Get(packet.DestinationPort())
	virtualServerStream, _ := virtualServer.Get(packet.DestinationStreamType())

	discriminator := fmt.Sprintf("%s-%d-%d", address.String(), packet.SourcePort(), packet.SourceStreamType())

	client, ok := virtualServerStream.Get(discriminator)

	if !ok {
		client = NewPRUDPClient(s, address, webSocketConnection)
		client.startHeartbeat()

		// * Fail-safe. If the server reboots, then
		// * clients has no record of old clients.
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

		virtualServerStream.Set(discriminator, client)
	}

	packet.SetSender(client)
	client.resetHeartbeat()

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
		substreamID, _ := stream.ReadPrimitiveUInt8()
		additionalIDsCount, _ := stream.ReadPrimitiveUInt8()
		baseSequenceID, _ = stream.ReadPrimitiveUInt16LE()
		substream = client.reliableSubstream(substreamID)

		for i := 0; i < int(additionalIDsCount); i++ {
			additionalID, _ := stream.ReadPrimitiveUInt16LE()
			sequenceIDs = append(sequenceIDs, additionalID)
		}
	} else {
		// TODO - This is how Kinnay's client handles this, but it doesn't make sense for QRV? Since it can have multiple reliable substreams?
		// * Old aggregate acknowledgment packets always use
		// * substream 0
		substream = client.reliableSubstream(0)
		baseSequenceID = packet.SequenceID()

		for stream.Remaining() > 0 {
			additionalID, _ := stream.ReadPrimitiveUInt16LE()
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

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(client, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(client, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(client, nil)
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

	s.sendRaw(client, ack.Bytes())
}

func (s *PRUDPServer) handleConnect(packet PRUDPPacketInterface) {
	client := packet.Sender().(*PRUDPClient)

	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(client, nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(client, nil)
	} else {
		ack, _ = NewPRUDPPacketV0(client, nil)
	}

	client.serverConnectionSignature = packet.getConnectionSignature()
	client.clientSessionID = packet.SessionID()

	connectionSignature, err := packet.calculateConnectionSignature(client.address)
	if err != nil {
		logger.Error(err.Error())
	}

	client.serverSessionID = packet.SessionID()

	ack.SetType(ConnectPacket)
	ack.AddFlag(FlagAck)
	ack.AddFlag(FlagHasSize)
	ack.SetSourceStreamType(packet.DestinationStreamType())
	ack.SetSourcePort(packet.DestinationPort())
	ack.SetDestinationStreamType(packet.SourceStreamType())
	ack.SetDestinationPort(packet.SourcePort())
	ack.setConnectionSignature(make([]byte, len(connectionSignature)))
	ack.SetSessionID(client.serverSessionID)
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
		client.outgoingUnreliableSequenceIDCounter = NewCounter[uint16](packet.(*PRUDPPacketV1).initialUnreliableSequenceID)
	} else {
		client.createReliableSubstreams(0)
	}

	var payload []byte

	if slices.Contains(s.SecureVirtualServerPorts, packet.DestinationPort()) {
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
		stream.WritePrimitiveUInt32LE(4)              // * Buffer length
		stream.WritePrimitiveUInt32LE(checkValue + 1) // * Buffer data

		payload = stream.Bytes()
	} else {
		payload = make([]byte, 0)
	}

	ack.SetPayload(payload)
	ack.setSignature(ack.calculateSignature([]byte{}, packet.getConnectionSignature()))

	s.emit("connect", ack)

	s.sendRaw(client, ack.Bytes())
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

	virtualServer, _ := s.virtualServers.Get(packet.DestinationPort())
	virtualServerStream, _ := virtualServer.Get(packet.DestinationStreamType())

	client := packet.Sender().(*PRUDPClient)
	discriminator := fmt.Sprintf("%s-%d-%d", client.address.String(), packet.SourcePort(), packet.SourceStreamType())

	client.cleanup() // * "removed" event is dispatched here
	virtualServerStream.Delete(discriminator)

	s.emit("disconnect", packet)
}

func (s *PRUDPServer) handlePing(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		s.acknowledgePacket(packet)
	}
}

func (s *PRUDPServer) readKerberosTicket(payload []byte) ([]byte, *types.PID, uint32, error) {
	stream := NewStreamIn(payload, s)

	ticketData := types.NewBuffer([]byte{})
	if err := ticketData.ExtractFrom(stream); err != nil {
		return nil, nil, 0, err
	}

	requestData := types.NewBuffer([]byte{})
	if err := requestData.ExtractFrom(stream); err != nil {
		return nil, nil, 0, err
	}

	serverKey := DeriveKerberosKey(types.NewPID(2), s.kerberosPassword)

	ticket := NewKerberosTicketInternalData()
	if err := ticket.Decrypt(NewStreamIn([]byte(*ticketData), s), serverKey); err != nil {
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

	decryptedRequestData, err := kerberos.Decrypt(*requestData)
	if err != nil {
		return nil, nil, 0, err
	}

	checkDataStream := NewStreamIn(decryptedRequestData, s)

	userPID := types.NewPID(0)
	if err := userPID.ExtractFrom(checkDataStream); err != nil {
		return nil, nil, 0, err
	}

	_, err = checkDataStream.ReadPrimitiveUInt32LE() // * CID of secure server station url
	if err != nil {
		return nil, nil, 0, err
	}

	responseCheck, err := checkDataStream.ReadPrimitiveUInt32LE()
	if err != nil {
		return nil, nil, 0, err
	}

	return sessionKey, userPID, responseCheck, nil
}

func (s *PRUDPServer) acknowledgePacket(packet PRUDPPacketInterface) {
	var ack PRUDPPacketInterface

	if packet.Version() == 2 {
		ack, _ = NewPRUDPPacketLite(packet.Sender().(*PRUDPClient), nil)
	} else if packet.Version() == 1 {
		ack, _ = NewPRUDPPacketV1(packet.Sender().(*PRUDPClient), nil)
	} else {
		ack, _ = NewPRUDPPacketV0(packet.Sender().(*PRUDPClient), nil)
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
			var decryptedPayload []byte

			if packet.Version() != 2 {
				decryptedPayload = pendingPacket.decryptPayload()
			} else {
				// * PRUDPLite does not encrypt payloads
				decryptedPayload = pendingPacket.Payload()
			}

			decompressedPayload, err := s.CompressionAlgorithm.Decompress(decryptedPayload)
			if err != nil {
				logger.Error(err.Error())
			}

			payload := substream.AddFragment(decompressedPayload)

			if packet.getFragmentID() == 0 {
				message := NewRMCMessage(s)
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

func (s *PRUDPServer) handleUnreliable(packet PRUDPPacketInterface) {
	if packet.HasFlag(FlagNeedsAck) {
		s.acknowledgePacket(packet)
	}

	// * Since unreliable DATA packets can in theory reach the
	// * server in any order, and they lack a substream, it's
	// * not actually possible to know what order they should
	// * be processed in for each request. So assume all packets
	// * MUST be fragment 0 (unreliable packets do not have frags)
	// *
	// * Example -
	// *
	// * Say there is 2 requests to the same protocol, methods 1
	// * and 2. The starting unreliable sequence ID is 10. If both
	// * method 1 and 2 are called at the same time, but method 1
	// * has a fragmented payload, the packets could, in theory, reach
	// * the server like so:
	// *
	// *	- Method1 - Sequence 10, Fragment 1
	// *	- Method1 - Sequence 13, Fragment 3
	// *	- Method2 - Sequence 12, Fragment 0
	// *	- Method1 - Sequence 11, Fragment 2
	// *	- Method1 - Sequence 14, Fragment 0
	// *
	// * If we reorder these to the proper order, like so:
	// *
	// *	- Method1 - Sequence 10, Fragment 1
	// *	- Method1 - Sequence 11, Fragment 2
	// *	- Method2 - Sequence 12, Fragment 0
	// *	- Method1 - Sequence 13, Fragment 3
	// *	- Method1 - Sequence 14, Fragment 0
	// *
	// * We still have a gap where Method2 was called. It's not
	// * possible to know if the packet with sequence ID 12 belongs
	// * to the Method1 calls or not. We don't even know which methods
	// * the packets are for at this stage yet, since the RMC data
	// * can't be checked until all the fragments are collected and
	// * the payload decrypted. In this case, we would see fragment 0
	// * and assume that's the end of fragments, losing the real last
	// * fragments and resulting in a bad decryption
	// TODO - Is this actually true? I'm just assuming, based on common sense, tbh. Kinnay also does not implement fragmented unreliable packets?
	if packet.getFragmentID() != 0 {
		logger.Warningf("Unexpected unreliable fragment ID. Expected 0, got %d", packet.getFragmentID())
		return
	}

	payload := packet.processUnreliableCrypto()

	message := NewRMCMessage(s)
	err := message.FromBytes(payload)
	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
	}

	packet.SetRMCMessage(message)

	s.emit("data", packet)
}

func (s *PRUDPServer) sendPing(client *PRUDPClient) {
	var ping PRUDPPacketInterface

	if s.websocketServer != nil {
		ping, _ = NewPRUDPPacketLite(client, nil)
	} else if s.PRUDPVersion == 0 {
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

	packetCopy.SetSessionID(client.serverSessionID)

	if packetCopy.Type() == DataPacket && !packetCopy.HasFlag(FlagAck) && !packetCopy.HasFlag(FlagMultiAck) {
		if packetCopy.HasFlag(FlagReliable) {
			payload := packetCopy.Payload()

			compressedPayload, err := s.CompressionAlgorithm.Compress(payload)
			if err != nil {
				logger.Error(err.Error())
			}

			substream := client.reliableSubstream(packetCopy.SubstreamID())

			// * According to other Quazal server implementations,
			// * the RC4 stream is always reset to the default key
			// * regardless if the client is connecting to a secure
			// * server (prudps) or not
			if s.IsQuazalMode {
				substream.SetCipherKey([]byte("CD&ML"))
			}

			// * PRUDPLite packet. No RC4
			if packetCopy.Version() != 2 {
				packetCopy.SetPayload(substream.Encrypt(compressedPayload))
			}
		} else {
			// * PRUDPLite packet. No RC4
			if packetCopy.Version() != 2 {
				packetCopy.SetPayload(packetCopy.processUnreliableCrypto())
			}
		}
	}

	packetCopy.setSignature(packetCopy.calculateSignature(client.sessionKey, client.serverConnectionSignature))

	if packetCopy.HasFlag(FlagReliable) && packetCopy.HasFlag(FlagNeedsAck) {
		substream := client.reliableSubstream(packetCopy.SubstreamID())
		substream.ResendScheduler.AddPacket(packetCopy)
	}

	s.sendRaw(packetCopy.Sender().(*PRUDPClient), packetCopy.Bytes())
}

// sendRaw will send the given client the provided packet
func (s *PRUDPServer) sendRaw(client *PRUDPClient, data []byte) {
	// TODO - Should this return the error too?

	var err error

	if s.udpSocket != nil {
		_, err = s.udpSocket.WriteToUDP(data, client.address.(*net.UDPAddr))
	} else if client.webSocketConnection != nil {
		err = client.webSocketConnection.WriteMessage(gws.OpcodeBinary, data)
	}

	if err != nil {
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
func (s *PRUDPServer) FindClientByConnectionID(serverPort, serverStreamType uint8, connectedID uint32) *PRUDPClient {
	var client *PRUDPClient

	virtualServer, _ := s.virtualServers.Get(serverPort)
	virtualServerStream, _ := virtualServer.Get(serverStreamType)

	virtualServerStream.Each(func(discriminator string, c *PRUDPClient) bool {
		if c.ConnectionID == connectedID {
			client = c
			return true
		}

		return false
	})

	return client
}

// FindClientByPID returns the PRUDP client connected with the given PID
func (s *PRUDPServer) FindClientByPID(serverPort, serverStreamType uint8, pid uint64) *PRUDPClient {
	var client *PRUDPClient

	virtualServer, _ := s.virtualServers.Get(serverPort)
	virtualServerStream, _ := virtualServer.Get(serverStreamType)

	virtualServerStream.Each(func(discriminator string, c *PRUDPClient) bool {
		if c.pid.Value() == pid {
			client = c
			return true
		}

		return false
	})

	return client
}

// PasswordFromPID calls the function set with SetPasswordFromPIDFunction and returns the result
func (s *PRUDPServer) PasswordFromPID(pid *types.PID) (string, uint32) {
	if s.passwordFromPIDHandler == nil {
		logger.Errorf("Missing PasswordFromPID handler. Set with SetPasswordFromPIDFunction")
		return "", Errors.Core.NotImplemented
	}

	return s.passwordFromPIDHandler(pid)
}

// SetPasswordFromPIDFunction sets the function for the auth server to get a NEX password using the PID
func (s *PRUDPServer) SetPasswordFromPIDFunction(handler func(pid *types.PID) (string, uint32)) {
	s.passwordFromPIDHandler = handler
}

// StringLengthSize returns the size of the length field used for Quazal::String types
func (s *PRUDPServer) StringLengthSize() int {
	return s.stringLengthSize
}

// SetStringLengthSize sets the size of the length field used for Quazal::String types
func (s *PRUDPServer) SetStringLengthSize(size int) {
	s.stringLengthSize = size
}

// NewPRUDPServer will return a new PRUDP server
func NewPRUDPServer() *PRUDPServer {
	return &PRUDPServer{
		VirtualServerPorts:       []uint8{1},
		SecureVirtualServerPorts: make([]uint8, 0),
		virtualServers:           NewMutexMap[uint8, *MutexMap[uint8, *MutexMap[string, *PRUDPClient]]](),
		IsQuazalMode:             false,
		kerberosKeySize:          32,
		FragmentSize:             1300,
		prudpEventHandlers:       make(map[string][]func(PacketInterface)),
		connectionIDCounter:      NewCounter[uint32](10),
		pingTimeout:              time.Second * 15,
		stringLengthSize:         2,
		CompressionAlgorithm:     compression.NewDummyCompression(),
	}
}
