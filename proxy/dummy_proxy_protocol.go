package proxy

import (
	"net"
)

// DummyProxyProtocol has no proxy header. Returns the data as-is
type DummyProxyProtocol struct{}

func (dpp *DummyProxyProtocol) HeaderSize() int {
	return 0
}

func (dpp *DummyProxyProtocol) Parse(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	return packet, nil
}

func (dpp *DummyProxyProtocol) Encode(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	return packet, nil
}

func NewDummyProxyProtocol() *DummyProxyProtocol {
	return &DummyProxyProtocol{}
}
