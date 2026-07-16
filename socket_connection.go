package nex

import (
	"net"

	"github.com/lxzan/gws"
)

// SocketConnection represents a single open socket.
// A single socket may have many PRUDP connections open on it.
type SocketConnection struct {
	Server              *PRUDPServer // * PRUDP server the socket is connected to
	ProxyAddress        net.Addr     // * Address of the proxy server. When not proxied, same as SocketConnection.Address
	Address             net.Addr     // * Address of the real client
	WebSocketConnection *gws.Conn    // * Only used in PRUDPLite
}

// NewSocketConnection creates a new SocketConnection
func NewSocketConnection(server *PRUDPServer, address net.Addr, webSocketConnection *gws.Conn) *SocketConnection {
	return &SocketConnection{
		Server:              server,
		ProxyAddress:        cloneAddr(address), // * Need to make a copy of the net.Addr so it can be worked with independently
		Address:             address,
		WebSocketConnection: webSocketConnection,
	}
}

// TODO - This is sort of a hack, replace this with our own type that implements net.Addr and adds this functionality natively?
func cloneAddr(addr net.Addr) net.Addr {
	switch v := addr.(type) {
	case *net.TCPAddr:
		return &net.TCPAddr{
			IP:   append([]byte(nil), v.IP...),
			Port: v.Port,
			Zone: v.Zone,
		}
	case *net.UDPAddr:
		return &net.UDPAddr{
			IP:   append([]byte(nil), v.IP...),
			Port: v.Port,
			Zone: v.Zone,
		}
	}

	// TODO - Maybe not safe?
	return nil
}
