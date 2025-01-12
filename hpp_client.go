package nex

import (
	"net"

	"github.com/PretendoNetwork/nex-go/v2/types"
)

// HPPClient represents a single HPP client
type HPPClient struct {
	address  *net.TCPAddr
	endpoint *HPPServer
	pid      types.PID
}

// Endpoint returns the server the client is connecting to
func (c *HPPClient) Endpoint() EndpointInterface {
	return c.endpoint
}

// Address returns the clients address as a net.Addr
func (c *HPPClient) Address() net.Addr {
	return c.address
}

// PID returns the clients NEX PID
func (c *HPPClient) PID() types.PID {
	return c.pid
}

// SetPID sets the clients NEX PID
func (c *HPPClient) SetPID(pid types.PID) {
	c.pid = pid
}

// NewHPPClient creates and returns a new Client using the provided IP address and server
func NewHPPClient(address *net.TCPAddr, server *HPPServer) *HPPClient {
	return &HPPClient{
		address:  address,
		endpoint: server,
	}
}
