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
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"time"
)

// Server represents a PRUDP server
type Server struct {
	socket                     *net.UDPConn
	clients                    map[string]*Client
	genericEventHandles        map[string][]func(PacketInterface)
	prudpV0EventHandles        map[string][]func(*PacketV0)
	prudpV1EventHandles        map[string][]func(*PacketV1)
	accessKey                  string
	prudpVersion               int
	prudpProtocolMinorVersion  int
	supportedFunctions         int
	fragmentSize               int16
	resendTimeout              float32
	pingTimeout                int
	kerberosPassword           string
	kerberosKeySize            int
	kerberosKeyDerivation      int
	kerberosTicketVersion      int
	connectionIDCounter        *Counter
	nexVersion                 *NEXVersion
	datastoreProtocolVersion   *NEXVersion
	matchMakingProtocolVersion *NEXVersion
	rankingProtocolVersion     *NEXVersion
	ranking2ProtocolVersion    *NEXVersion
	messagingProtocolVersion   *NEXVersion
	utilityProtocolVersion     *NEXVersion
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

	discriminator := addr.String()

	if _, ok := server.clients[discriminator]; !ok {
		newClient := NewClient(addr, server)
		server.clients[discriminator] = newClient
	}

	client := server.clients[discriminator]

	data := buffer[0:length]

	var packet PacketInterface

	if server.PRUDPVersion() == 0 {
		packet, err = NewPacketV0(client, data)
	} else {
		packet, err = NewPacketV1(client, data)
	}

	if err != nil {
		return nil
	}

	client.IncreasePingTimeoutTime(server.PingTimeout())

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
		client.Reset()
		client.SetConnected(true)
		client.StartTimeoutTimer()
		server.Emit("Syn", packet)
	case ConnectPacket:
		packet.Sender().SetClientConnectionSignature(packet.ConnectionSignature())

		server.Emit("Connect", packet)
	case DataPacket:
		server.Emit("Data", packet)
	case DisconnectPacket:
		server.Emit("Disconnect", packet)
		server.Kick(client)
	case PingPacket:
		//server.SendPing(client)
		server.Emit("Ping", packet)
	}

	server.Emit("Packet", packet)

	return nil
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
	}
}

// ClientConnected checks if a given client is stored on the server
func (server *Server) ClientConnected(client *Client) bool {
	discriminator := client.Address().String()

	_, connected := server.clients[discriminator]

	return connected
}

// Kick removes a client from the server
func (server *Server) Kick(client *Client) {
	var packet PacketInterface

	if server.PRUDPVersion() == 0 {
		packet, _ = NewPacketV0(client, nil)
	} else {
		packet, _ = NewPacketV1(client, nil)
	}
	packet.SetVersion(1)
	packet.SetSource(0xA1)
	packet.SetDestination(0xAF)
	packet.SetType(DisconnectPacket)

	packet.AddFlag(FlagReliable)

	server.Send(packet)

	server.Emit("Kick", packet)
	client.SetConnected(false)
	discriminator := client.Address().String()
	delete(server.clients, discriminator)
}

// Kick removes a client from the server
func (server *Server) KickAll() {
	for _, client := range server.clients {
		var packet PacketInterface

		if server.PRUDPVersion() == 0 {
			packet, _ = NewPacketV0(client, nil)
		} else {
			packet, _ = NewPacketV1(client, nil)
		}
		packet.SetVersion(1)
		packet.SetSource(0xA1)
		packet.SetDestination(0xAF)
		packet.SetType(DisconnectPacket)
	
		packet.AddFlag(FlagReliable)
	
		server.Send(packet)

		server.Emit("Kick", packet)
		client.SetConnected(false)
		discriminator := client.Address().String()
		delete(server.clients, discriminator)
	}
}

// SendPing sends a ping packet to the given client
func (server *Server) SendPing(client *Client) {
	var pingPacket PacketInterface

	if server.PRUDPVersion() == 0 {
		pingPacket, _ = NewPacketV0(client, nil)
	} else {
		pingPacket, _ = NewPacketV1(client, nil)
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

	if server.PRUDPVersion() == 0 {
		ackPacket, _ = NewPacketV0(sender, nil)
	} else {
		ackPacket, _ = NewPacketV1(sender, nil)
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
			rand.Read(serverConnectionSignature)

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
	server.datastoreProtocolVersion = nexVersion
	server.matchMakingProtocolVersion = nexVersion
	server.rankingProtocolVersion = nexVersion
	server.ranking2ProtocolVersion = nexVersion
	server.messagingProtocolVersion = nexVersion
	server.utilityProtocolVersion = nexVersion
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

// ConnectionIDCounter gets the server connection ID counter
func (server *Server) ConnectionIDCounter() *Counter {
	return server.connectionIDCounter
}

// FindClientFromPID finds a client by their PID
func (server *Server) FindClientFromPID(pid uint32) *Client {
	for _, client := range server.clients {
		if client.pid == pid {
			return client
		}
	}

	return nil
}

// FindClientFromConnectionID finds a client by their Connection ID
func (server *Server) FindClientFromConnectionID(rvcid uint32) *Client {
	for _, client := range server.clients {
		if client.connectionID == rvcid {
			return client
		}
	}

	return nil
}

// Send writes data to client
func (server *Server) Send(packet PacketInterface) {
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

// SendFragment sends a packet fragment to the client
func (server *Server) SendFragment(packet PacketInterface, fragmentID uint8) {
	data := packet.Payload()
	client := packet.Sender()

	packet.SetFragmentID(fragmentID)
	packet.SetPayload(data)
	packet.SetSequenceID(uint16(client.SequenceIDCounterOut().Increment()))

	encodedPacket := packet.Bytes()

	server.SendRaw(client.Address(), encodedPacket)
}

// SendRaw writes raw packet data to the client socket
func (server *Server) SendRaw(conn *net.UDPAddr, data []byte) {
	_, err := server.Socket().WriteToUDP(data, conn)
	if err != nil {
		logger.Error(err.Error())
	}
}

// NewServer returns a new NEX server
func NewServer() *Server {
	server := &Server{
		genericEventHandles:   make(map[string][]func(PacketInterface)),
		prudpV0EventHandles:   make(map[string][]func(*PacketV0)),
		prudpV1EventHandles:   make(map[string][]func(*PacketV1)),
		clients:               make(map[string]*Client),
		prudpVersion:          1,
		fragmentSize:          1300,
		resendTimeout:         1.5,
		pingTimeout:           5,
		kerberosKeySize:       32,
		kerberosKeyDerivation: 0,
		connectionIDCounter:   NewCounter(10),
	}

	return server
}
