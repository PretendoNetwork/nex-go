package nex

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Settings is a port of the settings handler in Kinnay's NintendoClients repo
type Settings struct {
	PrudpTransport          int
	PrudpVersion            int
	PrudpStreamType         int
	PrudpFragmentSize       int
	PrudpResendTimeout      float32
	PrudpPingTimeout        int
	PrudpSilenceTimeout     float32
	PrudpCompression        int
	PrudpV0SignatureVersion int
	PrudpV0FlagsVersion     int
	PrudpV0ChecksumVersion  int
	KerberosKeySize         int
	KerberosKeyDerivation   int
	IntSize                 int
	Version                 int
	AccessKey               string
}

// NewSettings returns a set of default server settings
func NewSettings() Settings {
	// Defaults
	return Settings{
		PrudpTransport:          0,
		PrudpVersion:            1,
		PrudpStreamType:         10,
		PrudpFragmentSize:       1300,
		PrudpResendTimeout:      1.5,
		PrudpPingTimeout:        4,
		PrudpSilenceTimeout:     7.5,
		PrudpCompression:        0,
		PrudpV0SignatureVersion: 0,
		PrudpV0FlagsVersion:     1,
		PrudpV0ChecksumVersion:  1,
		KerberosKeySize:         32,
		KerberosKeyDerivation:   0,
		IntSize:                 4,
		Version:                 0,
	}
}

// Server represents generic NEX server
type Server struct {
	_UDPServer       *net.UDPConn
	Settings         Settings
	CompressPacket   func([]byte) []byte
	DecompressPacket func([]byte) []byte
	Clients          map[string]*Client
	Handlers         map[string]func(*Client, *Packet)
}

// NewServer returns a new NEX server
func NewServer(settings Settings) *Server {
	server := &Server{
		Settings: settings,
		Handlers: make(map[string]func(*Client, *Packet)),
		Clients:  make(map[string]*Client),
	}

	if settings.PrudpCompression == 0 {
		compression := DummyCompression{}
		server.CompressPacket = compression.Compress
	} else {
		compression := ZLibCompression{}
		server.CompressPacket = compression.Compress
	}

	return server
}

// Listen starts a NEX server on a given port
func (server *Server) Listen(port string) {

	protocol := "udp"

	address, _ := net.ResolveUDPAddr(protocol, port)
	UDPServer, _ := net.ListenUDP(protocol, address)

	server._UDPServer = UDPServer

	fmt.Println("NEX server listening on port", port)

	for {
		readPacket(server)
	}
}

// On defines a datagram event handler
func (server *Server) On(event string, handler func(*Client, *Packet)) {
	server.Handlers[event] = handler
}

// Kick removes a client from the server
func (server *Server) Kick(client Client) {
	discriminator := client._UDPConn.String()

	if _, ok := server.Clients[discriminator]; ok {
		delete(server.Clients, discriminator)
		fmt.Println("Kicked user", discriminator)
	}
}

func (server *Server) createAckPacket(Packet *Packet) Packet {
	ack := NewPacket(Packet.Sender)

	if Packet.Type == Types["Syn"] {
		Packet.Sender.ServerConnectionSignature = Packet.Signature
		ack.SetSignature(Packet.Signature)
	}

	if Packet.Type == Types["Connect"] {
		Packet.Sender.ClientConnectionSignature = Packet.Signature
		ack.SetSignature(Packet.Signature)
	}

	ack.SetVersion(Packet.Version)
	ack.SetDestination(Packet.Source)
	ack.SetSource(Packet.Destination)
	ack.SetType(Packet.Type)
	ack.AddFlag(Flags["Ack"])
	ack.FragmentID = Packet.FragmentID
	ack.SequenceID = Packet.SequenceID
	return ack
}

// AcknowledgeWithPayload creates an acknowledgement packet based on the input packet
// with an additional payload.
func (server *Server) AcknowledgeWithPayload(Packet *Packet, Payload []byte) {
	ack := server.createAckPacket(Packet)
	ack.Payload = Payload
	server.SendRaw(Packet.Sender._UDPConn, ack.Bytes())
}

// Acknowledge creates an acknowledgement packet based on the input packet
func (server *Server) Acknowledge(Packet *Packet) {
	ack := server.createAckPacket(Packet)
	server.SendRaw(Packet.Sender._UDPConn, ack.Bytes())
}

// Send writes data to client
func (server *Server) Send(client *Client, packet *Packet) {
	data := packet.Payload
	times := len(data) / server.Settings.PrudpFragmentSize

	fragmentID := 1
	for i := 0; i <= times; i++ {
		if len(data) < server.Settings.PrudpFragmentSize {
			packet.SetPayload(data)
			server.SendFragment(client, packet, 0)
		} else {
			packet.SetPayload(data[:server.Settings.PrudpFragmentSize])
			server.SendFragment(client, packet, fragmentID)

			data = data[server.Settings.PrudpFragmentSize:]
			fragmentID++
		}
	}
}

// SendFragment sends a packet fragment to the client
func (server *Server) SendFragment(client *Client, packet *Packet, fragmentID int) {
	data := packet.Payload
	packet.SetPayload(server.CompressPacket(data))
	packet.SequenceID = client.SequenceIDOut.Increment()
	server.SendRaw(client._UDPConn, packet.Bytes())
}

// SendRaw writes raw packet data to the client socket
func (server *Server) SendRaw(conn *net.UDPAddr, data []byte) {
	server._UDPServer.WriteToUDP(data, conn)
}

func readPacket(server *Server) {

	var buffer [64000]byte
	len, addr, _ := server._UDPServer.ReadFromUDP(buffer[0:])

	discriminator := addr.String()

	if _, ok := server.Clients[discriminator]; !ok {
		newClient := NewClient(addr, server)
		newClient.SignatureBase = sum([]byte(server.Settings.AccessKey))
		newClient.SignatureKey = hex.EncodeToString(MD5Hash([]byte(server.Settings.AccessKey)))

		server.Clients[discriminator] = &newClient
	}

	client := server.Clients[discriminator]

	data := buffer[0:len]

	Packet := NewPacket(client)
	Packet.FromBytes(data)

	if server.Handlers["Packet"] != nil {
		server.Handlers["Packet"](client, &Packet)
	}

	switch Packet.Type {
	case 0:
		client.SetKey("CD&ML")
		handler := server.Handlers["Syn"]
		if handler != nil {
			handler(client, &Packet)
		}
	case 1:
		rand.Seed(time.Now().UnixNano())
		client.SessionID = rand.Intn(0xFF)

		handler := server.Handlers["Connect"]
		if handler != nil {
			handler(client, &Packet)
		}
	case 2:
		handler := server.Handlers["Data"]
		if handler != nil {
			handler(client, &Packet)
		}
	case 3:
		handler := server.Handlers["Disconnect"]
		if handler != nil {
			handler(client, &Packet)
		}
	case 4:
		handler := server.Handlers["Ping"]
		if handler != nil {
			handler(client, &Packet)
		}
	default:
		fmt.Println("UNKNOWN TYPE", Packet.Type)
	}
}
