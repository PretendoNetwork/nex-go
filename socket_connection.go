package nex

import (
	"net"

	"github.com/lxzan/gws"
)

// SocketConnection represents a single open socket.
// A single socket may have many PRUDP connections open on it.
type SocketConnection struct {
	Server              *PRUDPServer                       // * PRUDP server the socket is connected to
	Address             net.Addr                           // * Sockets address
	WebSocketConnection *gws.Conn                          // * Only used in PRUDPLite
	Connections         *MutexMap[uint8, *PRUDPConnection] // * Open PRUDP connections separated by rdv::Stream ID, also called "port number"
}

// NewSocketConnection creates a new SocketConnection
func NewSocketConnection(server *PRUDPServer, address net.Addr, webSocketConnection *gws.Conn) *SocketConnection {
	return &SocketConnection{
		Server:              server,
		Address:             address,
		WebSocketConnection: webSocketConnection,
		Connections:         NewMutexMap[uint8, *PRUDPConnection](),
	}
}
