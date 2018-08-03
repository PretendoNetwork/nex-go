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

// NewServer returns a new NEX server
func NewServer() *Server {
	return &Server{
		Handlers: make(map[string]func(Client, General.Packet)),
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
