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
	socket                *net.UDPConn
	compressPacket        func([]byte) []byte
	decompressPacket      func([]byte) []byte
	clients               map[string]*Client
	genericEventHandles   map[string][]func(PacketInterface)
	prudpV0EventHandles   map[string][]func(*PacketV0)
	prudpV1EventHandles   map[string][]func(*PacketV1)
	accessKey             string
	prudpVersion          int
	nexVersion            int
	fragmentSize          int16
	resendTimeout         float32
	usePacketCompression  bool
	pingTimeout           int
	signatureVersion      int
	flagsVersion          int
	checksumVersion       int
	kerberosKeySize       int
	kerberosKeyDerivation int
	serverVersion         int
	connectionIDCounter   *Counter
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

	fmt.Println("NEX server listening on address", udpAddress)

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

	if server.PrudpVersion() == 0 {
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
	}

	switch packet.Type() {
	case SynPacket:
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
	switch handler.(type) {
	case func(PacketInterface):
		server.genericEventHandles[event] = append(server.genericEventHandles[event], handler.(func(PacketInterface)))
	case func(*PacketV0):
		server.prudpV0EventHandles[event] = append(server.prudpV0EventHandles[event], handler.(func(*PacketV0)))
	case func(*PacketV1):
		server.prudpV1EventHandles[event] = append(server.prudpV1EventHandles[event], handler.(func(*PacketV1)))
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

	switch packet.(type) {
	case *PacketV0:
		eventName := server.prudpV0EventHandles[event]
		for i := 0; i < len(eventName); i++ {
			handler := eventName[i]
			go handler(packet.(*PacketV0))
		}
	case *PacketV1:
		eventName := server.prudpV1EventHandles[event]
		for i := 0; i < len(eventName); i++ {
			handler := eventName[i]
			go handler(packet.(*PacketV1))
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
	// Server events expect a packet to be passed, even though this isn't really a packet event
	var packet PacketInterface

	if server.PrudpVersion() == 0 {
		packet, _ = NewPacketV0(client, nil)
	} else {
		packet, _ = NewPacketV1(client, nil)
	}

	server.Emit("Kick", packet)
	client.SetConnected(false)
	discriminator := client.Address().String()
	delete(server.clients, discriminator)
}

// SendPing sends a ping packet to the given client
func (server *Server) SendPing(client *Client) {
	var pingPacket PacketInterface

	if server.PrudpVersion() == 0 {
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

	if server.PrudpVersion() == 0 {
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

	if server.PrudpVersion() == 1 {
		packet := packet.(*PacketV1)
		ackPacket := ackPacket.(*PacketV1)

		ackPacket.SetVersion(1)
		ackPacket.SetSubstreamID(0)
		ackPacket.AddFlag(FlagHasSize)

		if packet.Type() == SynPacket {
			serverConnectionSignature := make([]byte, 16)
			rand.Read(serverConnectionSignature)

			ackPacket.Sender().SetServerConnectionSignature(serverConnectionSignature)

			ackPacket.SetSupportedFunctions(packet.SupportedFunctions())
			ackPacket.SetMaximumSubstreamID(0)

			ackPacket.SetConnectionSignature(serverConnectionSignature)
		}

		if packet.Type() == ConnectPacket {

			ackPacket.SetConnectionSignature(make([]byte, 16))

			ackPacket.SetSupportedFunctions(packet.SupportedFunctions())

			ackPacket.SetInitialSequenceID(10000)

			ackPacket.SetMaximumSubstreamID(0)
		}

		if packet.Type() == DataPacket {
			// Aggregate acknowledgement
			ackPacket.ClearFlag(FlagAck)
			ackPacket.AddFlag(FlagMultiAck)

			payloadStream := NewStreamOut(server)

			// New version
			if server.NexVersion() >= 2 {
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

	//fmt.Println(hex.EncodeToString(data))

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

// PrudpVersion returns the server PRUDP version
func (server *Server) PrudpVersion() int {
	return server.prudpVersion
}

// SetPrudpVersion sets the server PRUDP version
func (server *Server) SetPrudpVersion(prudpVersion int) {
	server.prudpVersion = prudpVersion
}

// NexVersion returns the server NEX version
func (server *Server) NexVersion() int {
	return server.nexVersion
}

// SetNexVersion sets the server NEX version
func (server *Server) SetNexVersion(nexVersion int) {
	server.nexVersion = nexVersion
}

// ChecksumVersion returns the server packet checksum version
func (server *Server) ChecksumVersion() int {
	return server.checksumVersion
}

// SetChecksumVersion sets the server packet checksum version
func (server *Server) SetChecksumVersion(checksumVersion int) {
	server.checksumVersion = checksumVersion
}

// FlagsVersion returns the server packet flags version
func (server *Server) FlagsVersion() int {
	return server.flagsVersion
}

// SetFlagsVersion sets the server packet flags version
func (server *Server) SetFlagsVersion(flagsVersion int) {
	server.flagsVersion = flagsVersion
}

// AccessKey returns the server access key
func (server *Server) AccessKey() string {
	return server.accessKey
}

// SetAccessKey sets the server access key
func (server *Server) SetAccessKey(accessKey string) {
	server.accessKey = accessKey
}

// SignatureVersion returns the server packet signature version
func (server *Server) SignatureVersion() int {
	return server.signatureVersion
}

// SetSignatureVersion sets the server packet signature version
func (server *Server) SetSignatureVersion(signatureVersion int) {
	server.signatureVersion = signatureVersion
}

// KerberosKeySize returns the server kerberos key size
func (server *Server) KerberosKeySize() int {
	return server.kerberosKeySize
}

// SetKerberosKeySize sets the server kerberos key size
func (server *Server) SetKerberosKeySize(kerberosKeySize int) {
	server.kerberosKeySize = kerberosKeySize
}

// PingTimeout returns the server ping timeout time in seconds
func (server *Server) PingTimeout() int {
	return server.pingTimeout
}

// SetPingTimeout sets the server ping timeout time in seconds
func (server *Server) SetPingTimeout(pingTimeout int) {
	server.pingTimeout = pingTimeout
}

// UsePacketCompression enables or disables packet compression
func (server *Server) UsePacketCompression(usePacketCompression bool) {
	if usePacketCompression {
		compression := ZLibCompression{}
		server.SetPacketCompression(compression.Compress)
	} else {
		compression := DummyCompression{}
		server.SetPacketCompression(compression.Compress)
	}

	server.usePacketCompression = usePacketCompression
}

// SetPacketCompression sets the packet compression function
func (server *Server) SetPacketCompression(compression func([]byte) []byte) {
	server.compressPacket = compression
}

// SetFragmentSize sets the packet fragment size
func (server *Server) SetFragmentSize(fragmentSize int16) {
	server.fragmentSize = fragmentSize
}

// ConnectionIDCounter gets the server connection ID counter
func (server *Server) ConnectionIDCounter() *Counter {
	return server.connectionIDCounter
}

// ClientFromPID sets the packet compression function
func (server *Server) FindClientFromPID(pid uint32) *Client {
	for _, client := range server.clients {
		if client.pid == pid {
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
	packet.SetPayload(server.compressPacket(data))
	packet.SetSequenceID(uint16(client.SequenceIDCounterOut().Increment()))

	encodedPacket := packet.Bytes()

	server.SendRaw(client.Address(), encodedPacket)
}

// SendRaw writes raw packet data to the client socket
func (server *Server) SendRaw(conn *net.UDPAddr, data []byte) {
	server.Socket().WriteToUDP(data, conn)
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
		signatureVersion:      0,
		flagsVersion:          1,
		checksumVersion:       1,
		kerberosKeySize:       32,
		kerberosKeyDerivation: 0,
		connectionIDCounter:   NewCounter(10),
	}

	server.UsePacketCompression(false)

	return server
}
