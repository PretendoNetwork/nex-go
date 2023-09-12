// Package nex implements an API for creating bare-bones
// NEX servers and clients and provides the underlying
// PRUDP implementation
//
// No NEX protocols are implemented in this package. For
// NEX protocols see https://github.com/PretendoNetwork/nex-protocols-go
//
// No PIA code is implemented in this package
package nex

import (
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"golang.org/x/exp/slices"
)

// Server represents a PRUDP server
type Server struct {
	socket                      *net.UDPConn
	clients                     *MutexMap[string, *Client]
	genericEventHandles         map[string][]func(PacketInterface)
	prudpV0EventHandles         map[string][]func(*PacketV0)
	prudpV1EventHandles         map[string][]func(*PacketV1)
	hppEventHandles             map[string][]func(*HPPPacket)
	hppClientResponses          map[*Client](chan []byte)
	passwordFromPIDHandler      func(pid uint32) (string, uint32)
	useNewMultiACK              bool
	accessKey                   string
	prudpVersion                int
	prudpProtocolMinorVersion   int
	supportedFunctions          int
	fragmentSize                int16
	resendTimeout               time.Duration
	resendMaxIterations         int
	pingTimeout                 int
	kerberosPassword            string
	kerberosKeySize             int
	kerberosKeyDerivation       int
	kerberosTicketVersion       int
	connectionIDCounter         *Counter
	nexVersion                  *NEXVersion
	datastoreProtocolVersion    *NEXVersion
	matchMakingProtocolVersion  *NEXVersion
	rankingProtocolVersion      *NEXVersion
	ranking2ProtocolVersion     *NEXVersion
	messagingProtocolVersion    *NEXVersion
	utilityProtocolVersion      *NEXVersion
	natTraversalProtocolVersion *NEXVersion
	emuSendPacketDropPercent    int
	emuRecvPacketDropPercent    int
}

// Listen starts a NEX server on a given address
func (server *Server) Listen(address string) {
	protocol := "udp"

	udpAddress, err := net.ResolveUDPAddr(protocol, address)
	if err != nil {
		panic(err)
	}

	socket, err := net.ListenUDP(protocol, udpAddress)
	if err != nil {
		panic(err)
	}

	server.SetSocket(socket)

	quit := make(chan struct{})

	for i := 0; i < runtime.NumCPU(); i++ {
		go server.listenDatagram(quit)
	}

	logger.Success(fmt.Sprintf("PRUDP server listening on address - %s", udpAddress.String()))

	server.Emit("Listening", nil)

	<-quit
}

func (server *Server) listenDatagram(quit chan struct{}) {
	err := error(nil)

	for err == nil {
		err = server.handleSocketMessage()
	}

	quit <- struct{}{}

	panic(err)
}

func (server *Server) handleSocketMessage() error {
	var buffer [64000]byte

	socket := server.Socket()

	length, addr, err := socket.ReadFromUDP(buffer[0:])
	if err != nil {
		return err
	}

	if server.shouldDropPacket(true) {
		// Emulate packet drop for debugging
		return nil
	}

	discriminator := addr.String()

	client, ok := server.clients.Get(discriminator)

	if !ok {
		client = NewClient(addr, server)

		server.clients.Set(discriminator, client)
	}

	data := buffer[0:length]

	var packet PacketInterface

	if server.PRUDPVersion() == 0 {
		packet, err = NewPacketV0(client, data)
	} else {
		packet, err = NewPacketV1(client, data)
	}

	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
		return nil
	}

	client.IncreasePingTimeoutTime(server.PingTimeout())

	if packet.HasFlag(FlagAck) || packet.HasFlag(FlagMultiAck) {
		// TODO - Should this return an error?
		server.handleAcknowledgement(packet)
		return nil
	}

	// TODO - Make a better API in client to access incomingPacketManager?
	client.incomingPacketManager.Push(packet)

	// TODO - Make this API smarter. Only track missing packets and not all packets?
	// * Keep processing packets so long as the next one is in the pool,
	// * this way if several packets came in out of order they all get
	// * processed at once the moment the correct next packet comes in
	for next := client.incomingPacketManager.Next(); next != nil; {
		err := server.processPacket(next)
		if err != nil {
			// TODO - Should this return the error too?
			logger.Error(err.Error())
			return nil
		}

		next = client.incomingPacketManager.Next()
	}

	return nil
}

func (server *Server) processPacket(packet PacketInterface) error {
	err := packet.DecryptPayload()
	if err != nil {
		return err
	}

	client := packet.Sender()

	if packet.HasFlag(FlagAck) || packet.HasFlag(FlagMultiAck) {
		return nil
	}

	if packet.HasFlag(FlagNeedsAck) {
		if packet.Type() != ConnectPacket || (packet.Type() == ConnectPacket && len(packet.Payload()) <= 0) {
			go server.AcknowledgePacket(packet, nil)
		}

		if packet.Type() == DisconnectPacket {
			go server.AcknowledgePacket(packet, nil)
			go server.AcknowledgePacket(packet, nil)
		}
	}

	switch packet.Type() {
	case SynPacket:
		// * PID should always be 0 when a fresh connection is made
		if client.PID() != 0 {
			// * Was connected before on the same device, using a different account
			server.Emit("Disconnect", packet) // * Disconnect the old connection
		}
		err := client.Reset()
		if err != nil {
			// TODO - Should this return the error too?
			logger.Error(err.Error())
			return nil
		}

		client.SetConnected(true)
		client.StartTimeoutTimer()
		// TODO - Don't make this part suck ass?
		// * Manually incrementing because the original manager gets destroyed in the reset
		// * but we need to still track the SYN packet was sent
		client.incomingPacketManager.currentSequenceID.Increment()
		server.Emit("Syn", packet)
	case ConnectPacket:
		packet.Sender().SetClientConnectionSignature(packet.ConnectionSignature())

		server.Emit("Connect", packet)
	case DataPacket:
		server.Emit("Data", packet)
	case DisconnectPacket:
		server.Emit("Disconnect", packet)
		server.GracefulKick(client)
	case PingPacket:
		//server.SendPing(client)
		server.Emit("Ping", packet)
	}

	server.Emit("Packet", packet)

	return nil
}

func (server *Server) handleAcknowledgement(packet PacketInterface) {
	if packet.Version() == 0 {
		// * PRUDPv0 does not have aggregate acknowledgement
		packet.Sender().outgoingResendManager.Remove(packet.SequenceID())
	} else {
		// TODO - Validate the aggregate packet is valid and can be processed
		sequenceIDs := make([]uint16, 0)
		stream := NewStreamIn(packet.Payload(), server)
		var baseSequenceID uint16

		// TODO - We should probably handle these errors lol
		if server.useNewMultiACK {
			_, _ = stream.ReadUInt8() // * Substream ID. NEX always uses 0
			additionalIDsCount, _ := stream.ReadUInt8()
			baseSequenceID, _ = stream.ReadUInt16LE()

			for i := 0; i < int(additionalIDsCount); i++ {
				additionalID, _ := stream.ReadUInt16LE()
				sequenceIDs = append(sequenceIDs, additionalID)
			}
		} else {
			baseSequenceID = packet.SequenceID()

			for remaining := stream.Remaining(); remaining != 0; {
				additionalID, _ := stream.ReadUInt16LE()
				sequenceIDs = append(sequenceIDs, additionalID)
				remaining = stream.Remaining()
			}
		}

		// * MutexMap.Each locks the mutex, can't remove while reading
		// * Have to just loop again
		packet.Sender().outgoingResendManager.pending.Each(func(sequenceID uint16, pending *PendingPacket) {
			if sequenceID <= baseSequenceID && !slices.Contains(sequenceIDs, sequenceID) {
				sequenceIDs = append(sequenceIDs, sequenceID)
			}
		})

		// * Actually remove the packets from the pool
		for _, sequenceID := range sequenceIDs {
			packet.Sender().outgoingResendManager.Remove(sequenceID)
		}
	}
}

// HPPListen starts a NEX HPP server on a given address
func (server *Server) HPPListen(address string) {
	hppHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		pidValue := req.Header.Get("pid")
		if pidValue == "" {
			logger.Error("[HPP] PID is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token := req.Header.Get("token")
		if token == "" {
			logger.Error("[HPP] Token is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		accessKeySignature := req.Header.Get("signature1")
		if accessKeySignature == "" {
			logger.Error("[HPP] Access key signature is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		passwordSignature := req.Header.Get("signature2")
		if passwordSignature == "" {
			logger.Error("[HPP] Password signature is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		pid, err := strconv.Atoi(pidValue)
		if err != nil {
			// TODO - Should this return the error too?
			logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rmcRequestString := req.FormValue("file")

		rmcRequestBytes := []byte(rmcRequestString)

		client := NewClient(nil, server)
		client.SetPID(uint32(pid))

		hppPacket, err := NewHPPPacket(client, rmcRequestBytes)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create new HPPPacket instance. %s", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hppPacket.SetAccessKeySignature(accessKeySignature)
		hppPacket.SetPasswordSignature(passwordSignature)

		err = hppPacket.ValidateAccessKey()
		if err != nil {
			// TODO - Should this return the error too?
			logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = hppPacket.ValidatePassword()
		if err != nil {
			logger.Error(err.Error())
			rmcRequest := hppPacket.RMCRequest()
			callID := rmcRequest.CallID()

			errorResponse := NewRMCResponse(0, callID)
			// * HPP returns PythonCore::ValidationError if password is missing or invalid
			errorResponse.SetError(Errors.PythonCore.ValidationError)

			_, err = w.Write(errorResponse.Bytes())
			if err != nil {
				// TODO - Should this return the error too?
				logger.Error(err.Error())
			}

			return
		}

		server.hppClientResponses[client] = make(chan []byte)

		server.Emit("Data", hppPacket)

		rmcResponseBytes := <-server.hppClientResponses[client]

		if len(rmcResponseBytes) > 0 {
			_, err = w.Write(rmcResponseBytes)
			if err != nil {
				// TODO - Should this return the error too?
				logger.Error(err.Error())
			}
		}

		delete(server.hppClientResponses, client)
	}

	http.HandleFunc("/hpp/", hppHandler)

	quit := make(chan struct{})

	go server.handleHTTP(address, quit)

	logger.Success(fmt.Sprintf("HPP server listening on address - %s", address))

	<-quit
}

func (server *Server) handleHTTP(address string, quit chan struct{}) {
	err := http.ListenAndServe(address, nil)

	quit <- struct{}{}

	panic(err)
}

// On sets the data event handler
func (server *Server) On(event string, handler interface{}) {
	// Check if the handler type matches one of the allowed types, and store the handler in it's allowed property
	// Need to cast the handler to the correct function type before storing
	switch handler := handler.(type) {
	case func(PacketInterface):
		server.genericEventHandles[event] = append(server.genericEventHandles[event], handler)
	case func(*PacketV0):
		server.prudpV0EventHandles[event] = append(server.prudpV0EventHandles[event], handler)
	case func(*PacketV1):
		server.prudpV1EventHandles[event] = append(server.prudpV1EventHandles[event], handler)
	case func(*HPPPacket):
		server.hppEventHandles[event] = append(server.hppEventHandles[event], handler)
	}
}

// Emit runs the given event handle
func (server *Server) Emit(event string, packet interface{}) {

	eventName := server.genericEventHandles[event]
	for i := 0; i < len(eventName); i++ {
		handler := eventName[i]
		packet := packet.(PacketInterface)
		go handler(packet)
	}

	// Check if the packet type matches one of the allowed types and run the given handler

	switch packet := packet.(type) {
	case *PacketV0:
		eventName := server.prudpV0EventHandles[event]
		for i := 0; i < len(eventName); i++ {
			handler := eventName[i]
			go handler(packet)
		}
	case *PacketV1:
		eventName := server.prudpV1EventHandles[event]
		for i := 0; i < len(eventName); i++ {
			handler := eventName[i]
			go handler(packet)
		}
	case *HPPPacket:
		eventName := server.hppEventHandles[event]
		for i := 0; i < len(eventName); i++ {
			handler := eventName[i]
			go handler(packet)
		}
	}
}

// ClientConnected checks if a given client is stored on the server
func (server *Server) ClientConnected(client *Client) bool {
	discriminator := client.Address().String()

	_, connected := server.clients.Get(discriminator)

	return connected
}

// TimeoutKick removes a client from the server for inactivity
func (server *Server) TimeoutKick(client *Client) {
	var packet PacketInterface
	var err error

	if server.PRUDPVersion() == 0 {
		packet, err = NewPacketV0(client, nil)
		packet.SetVersion(0)
	} else {
		packet, err = NewPacketV1(client, nil)
		packet.SetVersion(1)
	}

	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
		return
	}

	packet.SetSource(0xA1)
	packet.SetDestination(0xAF)
	packet.SetType(DisconnectPacket)

	server.Send(packet)

	server.Emit("Kick", packet)
	client.SetConnected(false)
	discriminator := client.Address().String()

	client.outgoingResendManager.Clear()
	server.clients.Delete(discriminator)
}

// GracefulKick removes an active client from the server
func (server *Server) GracefulKick(client *Client) {
	var packet PacketInterface
	var err error

	if server.PRUDPVersion() == 0 {
		packet, err = NewPacketV0(client, nil)
		packet.SetVersion(0)
	} else {
		packet, err = NewPacketV1(client, nil)
		packet.SetVersion(1)
	}

	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
		return
	}

	packet.SetSource(0xA1)
	packet.SetDestination(0xAF)
	packet.SetType(DisconnectPacket)

	packet.AddFlag(FlagReliable)

	server.Send(packet)

	server.Emit("Kick", packet)
	client.SetConnected(false)
	client.StopTimeoutTimer()
	discriminator := client.Address().String()

	client.outgoingResendManager.Clear()
	server.clients.Delete(discriminator)
}

// GracefulKickAll removes all clients from the server
func (server *Server) GracefulKickAll() {
	// * https://stackoverflow.com/a/40456170
	server.clients.RLock()
	defer server.clients.RUnlock()
	// TODO - MAKE A BETTER API FOR RANGING OVER THIS DATA INSIDE MutexMap!
	for _, client := range server.clients.real {
		server.clients.RUnlock()

		var packet PacketInterface
		var err error
		if server.PRUDPVersion() == 0 {
			packet, err = NewPacketV0(client, nil)
			packet.SetVersion(0)
		} else {
			packet, err = NewPacketV1(client, nil)
			packet.SetVersion(1)
		}

		if err != nil {
			// TODO - Should this return the error too?
			logger.Error(err.Error())
			server.clients.RLock()
			continue
		}

		packet.SetSource(0xA1)
		packet.SetDestination(0xAF)
		packet.SetType(DisconnectPacket)

		packet.AddFlag(FlagReliable)

		server.Send(packet)

		server.Emit("Kick", packet)
		client.SetConnected(false)
		discriminator := client.Address().String()

		client.outgoingResendManager.Clear()
		server.clients.Delete(discriminator)

		server.clients.RLock()
	}
}

// SendPing sends a ping packet to the given client
func (server *Server) SendPing(client *Client) {
	var pingPacket PacketInterface
	var err error

	if server.PRUDPVersion() == 0 {
		pingPacket, err = NewPacketV0(client, nil)
	} else {
		pingPacket, err = NewPacketV1(client, nil)
	}

	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
		return
	}

	pingPacket.SetSource(0xA1)
	pingPacket.SetDestination(0xAF)
	pingPacket.SetType(PingPacket)
	pingPacket.AddFlag(FlagNeedsAck)
	pingPacket.AddFlag(FlagReliable)

	server.Send(pingPacket)
}

// AcknowledgePacket acknowledges that the given packet was recieved
func (server *Server) AcknowledgePacket(packet PacketInterface, payload []byte) {
	sender := packet.Sender()

	var ackPacket PacketInterface
	var err error

	if server.PRUDPVersion() == 0 {
		ackPacket, err = NewPacketV0(sender, nil)
	} else {
		ackPacket, err = NewPacketV1(sender, nil)
	}

	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
		return
	}

	ackPacket.SetSource(packet.Destination())
	ackPacket.SetDestination(packet.Source())
	ackPacket.SetType(packet.Type())
	ackPacket.SetSequenceID(packet.SequenceID())
	ackPacket.SetFragmentID(packet.FragmentID())
	ackPacket.AddFlag(FlagAck)
	ackPacket.AddFlag(FlagHasSize)

	if payload != nil {
		ackPacket.SetPayload(payload)
	}

	if server.PRUDPVersion() == 1 {
		packet := packet.(*PacketV1)
		ackPacket := ackPacket.(*PacketV1)

		ackPacket.SetVersion(1)
		ackPacket.SetSubstreamID(0)
		ackPacket.AddFlag(FlagHasSize)

		if packet.Type() == SynPacket || packet.Type() == ConnectPacket {
			ackPacket.SetPRUDPProtocolMinorVersion(packet.sender.PRUDPProtocolMinorVersion())
			//Going to leave this note here in case this causes issues later on, but for now, the below line breaks Splatoon and Minecraft Wii U (and probs other later games).
			//ackPacket.SetSupportedFunctions(packet.sender.SupportedFunctions())
			ackPacket.SetMaximumSubstreamID(0)
		}

		if packet.Type() == SynPacket {
			serverConnectionSignature := make([]byte, 16)
			_, err := rand.Read(serverConnectionSignature)
			if err != nil {
				// TODO - Should this return the error too?
				logger.Error(err.Error())
				return
			}

			ackPacket.Sender().SetServerConnectionSignature(serverConnectionSignature)
			ackPacket.SetConnectionSignature(serverConnectionSignature)
		}

		if packet.Type() == ConnectPacket {
			ackPacket.SetConnectionSignature(make([]byte, 16))
			ackPacket.SetInitialSequenceID(10000)
		}

		if packet.Type() == DataPacket {
			// Aggregate acknowledgement
			ackPacket.ClearFlag(FlagAck)
			ackPacket.AddFlag(FlagMultiAck)

			payloadStream := NewStreamOut(server)

			// New version
			if server.PRUDPProtocolMinorVersion() >= 2 {
				ackPacket.SetSequenceID(0)
				ackPacket.SetSubstreamID(1)

				// I'm lazy so just ack one packet
				payloadStream.WriteUInt8(0)                      // substream ID
				payloadStream.WriteUInt8(0)                      // length of additional sequence ids
				payloadStream.WriteUInt16LE(packet.SequenceID()) // Sequence id
			}

			ackPacket.SetPayload(payloadStream.Bytes())
		}
	}

	data := ackPacket.Bytes()

	server.SendRaw(sender.Address(), data)
}

// Socket returns the underlying server UDP socket
func (server *Server) Socket() *net.UDPConn {
	return server.socket
}

// SetSocket sets the underlying UDP socket
func (server *Server) SetSocket(socket *net.UDPConn) {
	server.socket = socket
}

// UseNewMultiACK checks if the server uses the new FLAG_MULTI_ACK encoding
func (server *Server) UseNewMultiACK() bool {
	return server.useNewMultiACK
}

// SetUseNewMultiACK sets whether the server uses the new FLAG_MULTI_ACK encoding
func (server *Server) SetUseNewMultiACK(useNewMultiACK bool) {
	server.useNewMultiACK = useNewMultiACK
}

// PRUDPVersion returns the server PRUDP version
func (server *Server) PRUDPVersion() int {
	return server.prudpVersion
}

// SetPRUDPVersion sets the server PRUDP version
func (server *Server) SetPRUDPVersion(prudpVersion int) {
	server.prudpVersion = prudpVersion
}

// PRUDPProtocolMinorVersion returns the server PRUDP minor version
func (server *Server) PRUDPProtocolMinorVersion() int {
	return server.prudpProtocolMinorVersion
}

// SetPRUDPProtocolMinorVersion sets the server PRUDP minor
func (server *Server) SetPRUDPProtocolMinorVersion(prudpProtocolMinorVersion int) {
	server.prudpProtocolMinorVersion = prudpProtocolMinorVersion
}

// NEXVersion returns the server NEX version
func (server *Server) NEXVersion() *NEXVersion {
	return server.nexVersion
}

// SetDefaultNEXVersion sets the default NEX protocol versions
func (server *Server) SetDefaultNEXVersion(nexVersion *NEXVersion) {
	server.nexVersion = nexVersion
	server.datastoreProtocolVersion = nexVersion.Copy()
	server.matchMakingProtocolVersion = nexVersion.Copy()
	server.rankingProtocolVersion = nexVersion.Copy()
	server.ranking2ProtocolVersion = nexVersion.Copy()
	server.messagingProtocolVersion = nexVersion.Copy()
	server.utilityProtocolVersion = nexVersion.Copy()
	server.natTraversalProtocolVersion = nexVersion.Copy()
}

// DataStoreProtocolVersion returns the servers DataStore protocol version
func (server *Server) DataStoreProtocolVersion() *NEXVersion {
	return server.datastoreProtocolVersion
}

// SetDataStoreProtocolVersion sets the servers DataStore protocol version
func (server *Server) SetDataStoreProtocolVersion(nexVersion *NEXVersion) {
	server.datastoreProtocolVersion = nexVersion
}

// MatchMakingProtocolVersion returns the servers MatchMaking protocol version
func (server *Server) MatchMakingProtocolVersion() *NEXVersion {
	return server.matchMakingProtocolVersion
}

// SetMatchMakingProtocolVersion sets the servers MatchMaking protocol version
func (server *Server) SetMatchMakingProtocolVersion(nexVersion *NEXVersion) {
	server.matchMakingProtocolVersion = nexVersion
}

// RankingProtocolVersion returns the servers Ranking protocol version
func (server *Server) RankingProtocolVersion() *NEXVersion {
	return server.rankingProtocolVersion
}

// SetRankingProtocolVersion sets the servers Ranking protocol version
func (server *Server) SetRankingProtocolVersion(nexVersion *NEXVersion) {
	server.rankingProtocolVersion = nexVersion
}

// Ranking2ProtocolVersion returns the servers Ranking2 protocol version
func (server *Server) Ranking2ProtocolVersion() *NEXVersion {
	return server.ranking2ProtocolVersion
}

// SetRanking2ProtocolVersion sets the servers Ranking2 protocol version
func (server *Server) SetRanking2ProtocolVersion(nexVersion *NEXVersion) {
	server.ranking2ProtocolVersion = nexVersion
}

// MessagingProtocolVersion returns the servers Messaging protocol version
func (server *Server) MessagingProtocolVersion() *NEXVersion {
	return server.messagingProtocolVersion
}

// SetMessagingProtocolVersion sets the servers Messaging protocol version
func (server *Server) SetMessagingProtocolVersion(nexVersion *NEXVersion) {
	server.messagingProtocolVersion = nexVersion
}

// UtilityProtocolVersion returns the servers Utility protocol version
func (server *Server) UtilityProtocolVersion() *NEXVersion {
	return server.utilityProtocolVersion
}

// SetUtilityProtocolVersion sets the servers Utility protocol version
func (server *Server) SetUtilityProtocolVersion(nexVersion *NEXVersion) {
	server.utilityProtocolVersion = nexVersion
}

// SetNATTraversalProtocolVersion sets the servers NAT Traversal protocol version
func (server *Server) SetNATTraversalProtocolVersion(nexVersion *NEXVersion) {
	server.natTraversalProtocolVersion = nexVersion
}

// NATTraversalProtocolVersion returns the servers NAT Traversal protocol version
func (server *Server) NATTraversalProtocolVersion() *NEXVersion {
	return server.natTraversalProtocolVersion
}

// SupportedFunctions returns the supported PRUDP functions by the server
func (server *Server) SupportedFunctions() int {
	return server.supportedFunctions
}

// SetSupportedFunctions sets the supported PRUDP functions by the server
func (server *Server) SetSupportedFunctions(supportedFunctions int) {
	server.supportedFunctions = supportedFunctions
}

// AccessKey returns the server access key
func (server *Server) AccessKey() string {
	return server.accessKey
}

// SetAccessKey sets the server access key
func (server *Server) SetAccessKey(accessKey string) {
	server.accessKey = accessKey
}

// KerberosPassword returns the server kerberos password
func (server *Server) KerberosPassword() string {
	return server.kerberosPassword
}

// SetKerberosPassword sets the server kerberos password
func (server *Server) SetKerberosPassword(kerberosPassword string) {
	server.kerberosPassword = kerberosPassword
}

// KerberosKeySize returns the server kerberos key size
func (server *Server) KerberosKeySize() int {
	return server.kerberosKeySize
}

// SetKerberosKeySize sets the server kerberos key size
func (server *Server) SetKerberosKeySize(kerberosKeySize int) {
	server.kerberosKeySize = kerberosKeySize
}

// KerberosTicketVersion returns the server kerberos ticket contents version
func (server *Server) KerberosTicketVersion() int {
	return server.kerberosTicketVersion
}

// SetKerberosTicketVersion sets the server kerberos ticket contents version
func (server *Server) SetKerberosTicketVersion(ticketVersion int) {
	server.kerberosTicketVersion = ticketVersion
}

// PingTimeout returns the server ping timeout time in seconds
func (server *Server) PingTimeout() int {
	return server.pingTimeout
}

// SetPingTimeout sets the server ping timeout time in seconds
func (server *Server) SetPingTimeout(pingTimeout int) {
	server.pingTimeout = pingTimeout
}

// SetFragmentSize sets the packet fragment size
func (server *Server) SetFragmentSize(fragmentSize int16) {
	server.fragmentSize = fragmentSize
}

// SetResendTimeout sets the time that a packet should wait before resending to the client
func (server *Server) SetResendTimeout(resendTimeout time.Duration) {
	server.resendTimeout = resendTimeout
}

// SetResendMaxIterations sets the max number of times a packet can try to resend before assuming the client is dead
func (server *Server) SetResendMaxIterations(resendMaxIterations int) {
	server.resendMaxIterations = resendMaxIterations
}

// ConnectionIDCounter gets the server connection ID counter
func (server *Server) ConnectionIDCounter() *Counter {
	return server.connectionIDCounter
}

// FindClientFromPID finds a client by their PID
func (server *Server) FindClientFromPID(pid uint32) *Client {
	// * https://stackoverflow.com/a/40456170
	// TODO - MAKE A BETTER API FOR RANGING OVER THIS DATA INSIDE MutexMap!
	server.clients.RLock()
	for _, client := range server.clients.real {
		server.clients.RUnlock()
		if client.pid == pid {
			return client
		}
		server.clients.RLock()
	}

	server.clients.RUnlock()

	return nil
}

// FindClientFromConnectionID finds a client by their Connection ID
func (server *Server) FindClientFromConnectionID(rvcid uint32) *Client {
	// * https://stackoverflow.com/a/40456170
	// TODO - MAKE A BETTER API FOR RANGING OVER THIS DATA INSIDE MutexMap!
	server.clients.RLock()
	for _, client := range server.clients.real {
		server.clients.RUnlock()
		if client.connectionID == rvcid {
			return client
		}
		server.clients.RLock()
	}

	server.clients.RUnlock()

	return nil
}

// SetPasswordFromPIDFunction sets the function for HPP or the auth server to get a NEX password using the PID
func (server *Server) SetPasswordFromPIDFunction(handler func(pid uint32) (string, uint32)) {
	server.passwordFromPIDHandler = handler
}

// PasswordFromPIDFunction returns the function for HPP or the auth server to get a NEX password using the PID
func (server *Server) PasswordFromPIDFunction() func(pid uint32) (string, uint32) {
	return server.passwordFromPIDHandler
}

// Send writes data to client
func (server *Server) Send(packet PacketInterface) {
	switch packet := packet.(type) {
	case *HPPPacket:
		client := packet.Sender()
		payload := packet.Payload()
		server.hppClientResponses[client] <- payload
	default:
		data := packet.Payload()
		fragments := int(int16(len(data)) / server.fragmentSize)

		var fragmentID uint8 = 1
		for i := 0; i <= fragments; i++ {
			time.Sleep(time.Second / 2)
			if int16(len(data)) < server.fragmentSize {
				packet.SetPayload(data)
				server.SendFragment(packet, 0)
			} else {
				packet.SetPayload(data[:server.fragmentSize])
				server.SendFragment(packet, fragmentID)

				data = data[server.fragmentSize:]
				fragmentID++
			}
		}
	}

}

// SendFragment sends a packet fragment to the client
func (server *Server) SendFragment(packet PacketInterface, fragmentID uint8) {
	client := packet.Sender()
	payload := packet.Payload()

	if packet.Type() == DataPacket {
		if packet.Version() == 0 && packet.HasFlag(FlagAck) {
			// * v0 ACK payloads empty, ensure this
			payload = []byte{}
		} else if !packet.HasFlag(FlagMultiAck) {
			if payload != nil || len(payload) > 0 {
				payloadSize := len(payload)

				encrypted := make([]byte, payloadSize)
				packet.Sender().Cipher().XORKeyStream(encrypted, payload)

				payload = encrypted
			}
		}

		// * Only add the HAS_SIZE flag if the payload exists
		if !packet.HasFlag(FlagHasSize) && len(payload) > 0 {
			packet.AddFlag(FlagHasSize)
		}
	}

	packet.SetFragmentID(fragmentID)

	packet.SetPayload(payload)
	packet.SetSequenceID(uint16(client.SequenceIDOutManager().Next(packet)))

	encodedPacket := packet.Bytes()

	server.SendRaw(client.Address(), encodedPacket)

	if (packet.HasFlag(FlagReliable) || packet.Type() == SynPacket) && packet.HasFlag(FlagNeedsAck) {
		packet.Sender().outgoingResendManager.Add(packet)
	}
}

// SendRaw writes raw packet data to the client socket
func (server *Server) SendRaw(conn *net.UDPAddr, data []byte) {
	if server.shouldDropPacket(false) {
		// Emulate packet drop for debugging
		return
	}

	_, err := server.Socket().WriteToUDP(data, conn)
	if err != nil {
		// TODO - Should this return the error too?
		logger.Error(err.Error())
	}
}

func (server *Server) shouldDropPacket(isRecv bool) bool {
	if isRecv {
		return server.emuRecvPacketDropPercent != 0 && mrand.Intn(100) < server.emuRecvPacketDropPercent
	} else {
		return server.emuSendPacketDropPercent != 0 && mrand.Intn(100) < server.emuSendPacketDropPercent
	}
}

// SetEmulatedPacketDropPercent sets the percentage of emulated sent and received dropped packets
func (server *Server) SetEmulatedPacketDropPercent(forRecv bool, percent int) {
	if forRecv {
		server.emuRecvPacketDropPercent = percent
	} else {
		server.emuSendPacketDropPercent = percent
	}
}

// NewServer returns a new NEX server
func NewServer() *Server {
	server := &Server{
		genericEventHandles:      make(map[string][]func(PacketInterface)),
		prudpV0EventHandles:      make(map[string][]func(*PacketV0)),
		prudpV1EventHandles:      make(map[string][]func(*PacketV1)),
		hppEventHandles:          make(map[string][]func(*HPPPacket)),
		hppClientResponses:       make(map[*Client](chan []byte)),
		clients:                  NewMutexMap[string, *Client](),
		useNewMultiACK:           false,
		prudpVersion:             1,
		fragmentSize:             1300,
		resendTimeout:            time.Second * 2,
		resendMaxIterations:      5,
		pingTimeout:              5,
		kerberosKeySize:          32,
		kerberosKeyDerivation:    0,
		connectionIDCounter:      NewCounter(10),
		emuSendPacketDropPercent: 0,
		emuRecvPacketDropPercent: 0,
	}

	server.SetDefaultNEXVersion(NewNEXVersion(0, 0, 0))

	return server
}
