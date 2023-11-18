package nex

import "net"

// HPPClient represents a single HPP client
type HPPClient struct {
	address *net.TCPAddr
	server  *HPPServer
	pid     *PID
}

// Server returns the server the client is connecting to
func (c *HPPClient) Server() ServerInterface {
	return c.server
}

// Address returns the clients address as a net.Addr
func (c *HPPClient) Address() net.Addr {
	return c.address
}

// PID returns the clients NEX PID
func (c *HPPClient) PID() *PID {
	return c.pid
}

// SetPID sets the clients NEX PID
func (c *HPPClient) SetPID(pid *PID) {
	c.pid = pid
}

// NewHPPClient creates and returns a new Client using the provided IP address and server
func NewHPPClient(address *net.TCPAddr, server *HPPServer) *HPPClient {
	return &HPPClient{
		address: address,
		server:  server,
	}
}
