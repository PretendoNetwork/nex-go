package nex

import (
	"fmt"
	"net"

	PRUDP "github.com/PretendoNetwork/nex-go/prudp"
)

// Server represents generic NEX server
// _UDPServer: The underlying UDP server the NEX server uses
// Clients   : A list of connected clients
// Handlers  : NEX server packet event handles
type Server struct {
	_UDPServer *net.UDPConn
	Clients    map[string]Client
	Handlers   map[string]func(Client, PRUDP.Packet)
}

// NewServer returns a new NEX server
func NewServer() *Server {
	return &Server{
		Handlers: make(map[string]func(Client, PRUDP.Packet)),
		Clients:  make(map[string]Client),
	}
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
func (server *Server) On(event string, handler func(Client, PRUDP.Packet)) {
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
func (server *Server) Send(client Client, Packet interface{}) {
	switch Packet.(type) {
	case []byte:
		buffer := Packet.([]byte)
		server._UDPServer.WriteToUDP(buffer, client._UDPConn)
	case PRUDP.Packet:
		buffer := Packet.(PRUDP.Packet).Bytes()
		server._UDPServer.WriteToUDP(buffer, client._UDPConn)
	}
}

func readPacket(server *Server) {

	var buffer [64000]byte
	len, addr, _ := server._UDPServer.ReadFromUDP(buffer[0:])

	discriminator := addr.String()

	if _, ok := server.Clients[discriminator]; !ok {
		server.Clients[discriminator] = NewClient(addr)
	}

	client := server.Clients[discriminator]

	data := buffer[0:len]

	Packet := PRUDP.NewPacket()
	Packet.FromBytes(data)

	switch Packet.Type {
	case 0:
		handler := server.Handlers["Syn"]
		if handler != nil {
			server.Handlers["Syn"](client, Packet)
		}
	case 1:
		handler := server.Handlers["Connect"]
		if handler != nil {
			server.Handlers["Connect"](client, Packet)
		}
	case 2:
		handler := server.Handlers["Data"]
		if handler != nil {
			server.Handlers["Data"](client, Packet)
		}
	case 3:
		handler := server.Handlers["Disconnect"]
		if handler != nil {
			server.Handlers["Disconnect"](client, Packet)
		}
	case 4:
		handler := server.Handlers["Ping"]
		if handler != nil {
			server.Handlers["Ping"](client, Packet)
		}
	default:
		fmt.Println("UNKNOWN TYPE", Packet.Type)
	}
}
