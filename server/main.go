package nex

import (
	"fmt"
	"net"

	PRUDPLib "github.com/PretendoNetwork/prudplib"
	General "github.com/PretendoNetwork/prudplib/General"
)

// Server represents generic NEX server
// _UDPServer: The underlying UDP server the NEX server uses
// Clients   : A list of connected clients
// Handlers  : NEX server packet event handles
type Server struct {
	_UDPServer *net.UDPConn
	Clients    map[string]Client
	Handlers   map[string]func(Client, General.Packet)
}

// Client represents generic NEX/PRUDP client
// _UDPConn           : The address of the client
// State              : The clients connection state
// SignatureKey       : MD5 hash of the servers access key
// SignatureBase      : Packet checksum base (sum on the SignatureKey)
// SecureKey          : NEX server packet event handles
// ConnectionSignature: The clients unique connection signature
// SessionID          : Clients session ID
// Packets            : Packets sent to the server from the client
// PacketQueue        : Packet queue
// SequenceIDIn       : The sequence ID for client->server packets
// SequenceIDOut      : The sequence ID for server->client packets
type Client struct {
	_UDPConn            *net.UDPAddr
	State               int
	SignatureKey        string
	SignatureBase       int
	SecureKey           []byte
	ConnectionSignature []byte
	SessionID           int
	Packets             []General.Packet
	PacketQueue         map[string]General.Packet
	SequenceIDIn        int
	SequenceIDOut       int
}

// NewServer returns a new NEX server
func NewServer() *Server {
	return &Server{
		Handlers: make(map[string]func(Client, General.Packet)),
		Clients:  make(map[string]Client),
	}
}

// NewClient returns a new generic client
func NewClient(addr *net.UDPAddr) Client {

	client := Client{
		_UDPConn:    addr,
		State:       0,
		SessionID:   0,
		PacketQueue: make(map[string]General.Packet),
	}

	return client
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
func (server *Server) On(event string, handler func(Client, General.Packet)) {
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

// Send writes data to client
func (server *Server) Send(client Client, data interface{}) {
	switch data.(type) {
	case []byte:
		buffer := data.([]byte)
		server._UDPServer.WriteToUDP(buffer, client._UDPConn)
	case General.Packet:
		// TODO
	}
}

func readPacket(server *Server) {

	var buffer [64000]byte
	len, addr, _ := server._UDPServer.ReadFromUDP(buffer[0:])

	discriminator := addr.String()

	if _, ok := server.Clients[discriminator]; ok {
		fmt.Println("Stored User")
	} else {
		server.Clients[discriminator] = NewClient(addr)
	}

	client := server.Clients[discriminator]

	packet := buffer[0:len]

	PRUDPPacket, _ := PRUDPLib.NewPacket(packet)

	switch PRUDPPacket.Type {
	case 0:
		handler := server.Handlers["Syn"]
		if handler != nil {
			server.Handlers["Syn"](client, PRUDPPacket)
		}
	case 1:
		handler := server.Handlers["Connect"]
		if handler != nil {
			server.Handlers["Connect"](client, PRUDPPacket)
		}
	case 2:
		handler := server.Handlers["Data"]
		if handler != nil {
			server.Handlers["Data"](client, PRUDPPacket)
		}
	case 3:
		handler := server.Handlers["Disconnect"]
		if handler != nil {
			server.Handlers["Disconnect"](client, PRUDPPacket)
		}
	case 4:
		handler := server.Handlers["Ping"]
		if handler != nil {
			server.Handlers["Ping"](client, PRUDPPacket)
		}
	default:
		fmt.Println("UNKNOWN TYPE", PRUDPPacket.Type)
	}
}

/*
// Testing the package
// This will be removed in the release
func main() {
	server := NewServer()

	server.On("Syn", func(client Client, packet General.Packet) {
		server.Send(client, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	})

	server.Listen(":60000")
}
*/
