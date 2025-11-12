package proxy

import "net"

type ProxyProtocol interface {
	// HeaderSize returns the size of the proxy protocol header
	// attached to the start of all packets
	HeaderSize() int

	// Parse extracts proxy header from the packet. The client address and proxy address
	// are updated in this function. Returns the real packet payload after the proxy header
	Parse(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error)

	// Encode wraps a payload with the proxy header for the given addresses
	Encode(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error)
}
