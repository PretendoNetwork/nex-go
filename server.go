package nex

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
)

// Server represents a PRUDP server
type Server struct {
	socket                *net.UDPConn
	compressionScheme     CompressionScheme
	clients               map[string]*Client
	genericEventHandles   map[string][]func(PacketInterface)
	prudpV0EventHandles   map[string][]func(*PacketV0)
	prudpV1EventHandles   map[string][]func(*PacketV1)
	accessKey             string
	prudpVersion          int
	nexMinorVersion       int
	fragmentSize          int16
	resendTimeout         float32
	pingTimeout           int
	signatureVersion      int
	flagsVersion          int
	checksumVersion       int
	kerberosKeySize       int
	kerberosKeyDerivation int
	serverVersion         int
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

	panic(err)

	quit <- struct{}{}
}

func (server *Server) handleSocketMessage() error {
	var buffer [64000]byte

	socket := server.GetSocket()

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

	if server.GetPrudpVersion() == 0 {
		packet, err = NewPacketV0(client, data)
	} else {
		packet, err = NewPacketV1(client, data)
	}

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if packet.HasFlag(FlagAck) || packet.HasFlag(FlagMultiAck) {
		return nil
	}

	if packet.HasFlag(FlagNeedsAck) {
		if packet.GetType() != ConnectPacket || (packet.GetType() == ConnectPacket && len(packet.GetPayload()) <= 0) {
			go server.AcknowledgePacket(packet, nil)
		}
	}

	switch packet.GetType() {
	case SynPacket:
		client.Reset()
		server.Emit("Syn", packet)
	case ConnectPacket:
		packet.GetSender().SetClientConnectionSignature(packet.GetConnectionSignature())

		server.Emit("Connect", packet)
	case DataPacket:
		server.Emit("Data", packet)
	case DisconnectPacket:
		server.Kick(client)
		server.Emit("Disconnect", packet)
	case PingPacket:
		server.SendPing(client)
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
	discriminator := client.GetAddress().String()

	_, connected := server.clients[discriminator]

	return connected
}

// Kick removes a client from the server
func (server *Server) Kick(client *Client) {
	discriminator := client.GetAddress().String()

	if _, ok := server.clients[discriminator]; ok {
		delete(server.clients, discriminator)
		fmt.Println("Kicked user", discriminator)
	}
}

// SendPing sends a ping packet to the given client
func (server *Server) SendPing(client *Client) {
	var pingPacket PacketInterface

	if server.GetPrudpVersion() == 0 {
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
	sender := packet.GetSender()

	var ackPacket PacketInterface

	if server.GetPrudpVersion() == 0 {
		ackPacket, _ = NewPacketV0(sender, nil)
	} else {
		ackPacket, _ = NewPacketV1(sender, nil)
	}

	ackPacket.SetSource(packet.GetDestination())
	ackPacket.SetDestination(packet.GetSource())
	ackPacket.SetType(packet.GetType())
	ackPacket.SetSequenceID(packet.GetSequenceID())
	ackPacket.SetFragmentID(packet.GetFragmentID())
	ackPacket.AddFlag(FlagAck)

	if payload != nil {
		ackPacket.SetPayload(payload)
	}

	if server.GetPrudpVersion() == 1 {
		packet := packet.(*PacketV1)
		ackPacket := ackPacket.(*PacketV1)

		ackPacket.SetVersion(1)
		ackPacket.SetSubstreamID(0)
		ackPacket.AddFlag(FlagHasSize)

		if packet.GetType() == SynPacket {
			serverConnectionSignature := make([]byte, 16)
			rand.Read(serverConnectionSignature)

			ackPacket.GetSender().SetServerConnectionSignature(serverConnectionSignature)

			ackPacket.SetSupportedFunctions(packet.GetSupportedFunctions())
			ackPacket.SetMaximumSubstreamID(0)

			ackPacket.SetConnectionSignature(serverConnectionSignature)
		}

		if packet.GetType() == ConnectPacket {

			ackPacket.SetConnectionSignature(make([]byte, 16))

			ackPacket.SetSupportedFunctions(packet.GetSupportedFunctions())

			ackPacket.SetInitialSequenceID(10000)

			ackPacket.SetMaximumSubstreamID(0)
		}

		if packet.GetType() == DataPacket {
			// Aggregate acknowledgement
			ackPacket.ClearFlag(FlagAck)
			ackPacket.AddFlag(FlagMultiAck)

			payloadStream := NewStreamOut(server)

			// New version
			if server.GetNexMinorVersion() >= 2 {
				ackPacket.SetSequenceID(0)
				ackPacket.SetSubstreamID(1)

				// I'm lazy so just ack one packet
				payloadStream.WriteUInt8(0)                         // substream ID
				payloadStream.WriteUInt8(0)                         // length of additional sequence ids
				payloadStream.WriteUInt16LE(packet.GetSequenceID()) // Sequence id
			}

			ackPacket.SetPayload(payloadStream.Bytes())
		}
	}

	data := ackPacket.Bytes()

	server.SendRaw(sender.GetAddress(), data)
}

// GetSocket returns the underlying server UDP socket
func (server *Server) GetSocket() *net.UDPConn {
	return server.socket
}

// SetSocket sets the underlying UDP socket
func (server *Server) SetSocket(socket *net.UDPConn) {
	server.socket = socket
}

// GetPrudpVersion returns the server PRUDP version
func (server *Server) GetPrudpVersion() int {
	return server.prudpVersion
}

// SetPrudpVersion sets the server PRUDP version
func (server *Server) SetPrudpVersion(prudpVersion int) {
	server.prudpVersion = prudpVersion
}

// GetNexMinorVersion returns the server NEX version
func (server *Server) GetNexMinorVersion() int {
	return server.nexMinorVersion
}

// SetNexMinorVersion sets the server NEX version
func (server *Server) SetNexMinorVersion(nexMinorVersion int) {
	server.nexMinorVersion = nexMinorVersion
}

// GetChecksumVersion returns the server packet checksum version
func (server *Server) GetChecksumVersion() int {
	return server.checksumVersion
}

// SetChecksumVersion sets the server packet checksum version
func (server *Server) SetChecksumVersion(checksumVersion int) {
	server.checksumVersion = checksumVersion
}

// GetFlagsVersion returns the server packet flags version
func (server *Server) GetFlagsVersion() int {
	return server.flagsVersion
}

// SetFlagsVersion sets the server packet flags version
func (server *Server) SetFlagsVersion(flagsVersion int) {
	server.flagsVersion = flagsVersion
}

// GetAccessKey returns the server access key
func (server *Server) GetAccessKey() string {
	return server.accessKey
}

// SetAccessKey sets the server access key
func (server *Server) SetAccessKey(accessKey string) {
	server.accessKey = accessKey
}

// GetSignatureVersion returns the server packet signature version
func (server *Server) GetSignatureVersion() int {
	return server.signatureVersion
}

// SetSignatureVersion sets the server packet signature version
func (server *Server) SetSignatureVersion(signatureVersion int) {
	server.signatureVersion = signatureVersion
}

// GetKerberosKeySize returns the server kerberos key size
func (server *Server) GetKerberosKeySize() int {
	return server.kerberosKeySize
}

// SetKerberosKeySize sets the server kerberos key size
func (server *Server) SetKerberosKeySize(kerberosKeySize int) {
	server.kerberosKeySize = kerberosKeySize
}

// UsePacketCompression sets the compression scheme used
func (server *Server) UsePacketCompression(scheme CompressionScheme) {
	server.compressionScheme = scheme
}

// Send writes data to client
func (server *Server) Send(packet PacketInterface) {
	data := packet.GetPayload()
	fragments := int(int16(len(data)) / server.fragmentSize)

	fragmentID := 1
	for i := 0; i <= fragments; i++ {
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
func (server *Server) SendFragment(packet PacketInterface, fragmentID int) {
	data, err := server.compressionScheme.Compress(packet.GetPayload())
	if err != nil {
		panic(err) // temporary, will replace with a proper logging and handling method later
	}

	client := packet.GetSender()

	packet.SetPayload(data)
	packet.SetSequenceID(uint16(client.GetSequenceIDCounterOut().Increment()))

	encodedPacket := packet.Bytes()

	server.SendRaw(client.GetAddress(), encodedPacket)
}

// SendRaw writes raw packet data to the client socket
func (server *Server) SendRaw(conn *net.UDPAddr, data []byte) {
	server.GetSocket().WriteToUDP(data, conn)
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
		compressionScheme:     NewDummyCompression(),
	}

	return server
}
