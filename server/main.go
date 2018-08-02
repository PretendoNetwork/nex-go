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
	Clients    []Client
	Handlers   map[string]func(*net.UDPAddr, General.Packet)
}

// Client represents generic NEX/PRUDP client
// State        : The clients connection state
// SignatureKey : MD5 hash of the servers access key
// SignatureBase: Packet checksum base
// SecureKey    : NEX server packet event handles
// Signature    : The clients signature
// CallID       : A list of connected clients
// SessionID    : Clients session ID
// Packets      : Client sent, server-bound packets
// PacketQueue  : Packet Queue
type Client struct {
	State         int
	SignatureKey  string
	SignatureBase int
	SecureKey     []byte
	Signature     []byte
	CallID        int
	SessionID     int
	Packets       []General.Packet
	PacketQueue   map[string]General.Packet
}

// NewServer returns a new NEX server
func NewServer() *Server {
	return &Server{
		Handlers: make(map[string]func(*net.UDPAddr, General.Packet)),
	}
}

// NewClient returns a new generic client
func NewClient() Client {

	client := Client{
		State:       0,
		CallID:      0,
		SessionID:   0,
		PacketQueue: make(map[string]General.Packet),
	}

	return client
}

// Listen starts a NEX server on a given port
func (NEXServer *Server) Listen(port string) {

	protocol := "udp"

	address, _ := net.ResolveUDPAddr(protocol, port)
	server, _ := net.ListenUDP(protocol, address)

	NEXServer._UDPServer = server

	fmt.Println("NEX server listening on port", port)

	for {
		readPacket(NEXServer)
	}
}

// On defines a datagram event handler
func (NEXServer *Server) On(event string, handler func(*net.UDPAddr, General.Packet)) {
	NEXServer.Handlers[event] = handler
}

func readPacket(NEXServer *Server) {

	var buffer [64000]byte
	len, addr, _ := NEXServer._UDPServer.ReadFromUDP(buffer[0:])

	packet := buffer[0:len]

	PRUDPPacket, _ := PRUDPLib.NewPacket(packet)

	switch PRUDPPacket.Type {
	case 0:
		handler := NEXServer.Handlers["Syn"]
		if handler != nil {
			NEXServer.Handlers["Syn"](addr, PRUDPPacket)
		}
	case 1:
		handler := NEXServer.Handlers["Connect"]
		if handler != nil {
			NEXServer.Handlers["Connect"](addr, PRUDPPacket)
		}
	case 2:
		handler := NEXServer.Handlers["Data"]
		if handler != nil {
			NEXServer.Handlers["Data"](addr, PRUDPPacket)
		}
	case 3:
		handler := NEXServer.Handlers["Disconnect"]
		if handler != nil {
			NEXServer.Handlers["Disconnect"](addr, PRUDPPacket)
		}
	case 4:
		handler := NEXServer.Handlers["Ping"]
		if handler != nil {
			NEXServer.Handlers["Ping"](addr, PRUDPPacket)
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

	server.On("Syn", func(client *net.UDPAddr, packet General.Packet) {
		fmt.Println("Handle PRUDP Packet")
	})

	server.On("Connect", func(client *net.UDPAddr, packet General.Packet) {
		fmt.Println("Handle PRUDP Packet")
	})

	server.On("Data", func(client *net.UDPAddr, packet General.Packet) {
		fmt.Println("Handle PRUDP Packet")
	})

	server.Listen(":60000")
}
*/
