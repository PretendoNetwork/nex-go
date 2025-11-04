package nex

import (
	"net"

	"github.com/lxzan/gws"
)

// SocketConnection represents a single open socket.
// A single socket may have many PRUDP connections open on it.
type SocketConnection struct {
	Server              *PRUDPServer // * PRUDP server the socket is connected to
	ProxyAddress        net.Addr     // * Address of the proxy server, when in proxied mode
	Address             net.Addr     // * Address of the real client
	WebSocketConnection *gws.Conn    // * Only used in PRUDPLite
}

// NewSocketConnection creates a new SocketConnection
func NewSocketConnection(server *PRUDPServer, address net.Addr, webSocketConnection *gws.Conn) *SocketConnection {
	return &SocketConnection{
		Server:              server,
		Address:             address,
		WebSocketConnection: webSocketConnection,
	}
}
