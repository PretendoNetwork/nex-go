package proxy

import (
	"encoding/binary"
	"net"
)

const PRUDP_SIMPLE_PROXY_PROTOCOL_VERSION = 0

// PRUDPSimpleProxyProtocol implements a custom proxy header for PRUDP. Should only be used when
// one of the other protocols cannot be used
type PRUDPSimpleProxyProtocol struct{}

func (pspp *PRUDPSimpleProxyProtocol) HeaderSize() int {
	return 7
}

func (pspp *PRUDPSimpleProxyProtocol) Parse(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	realClientIP := net.IP(packet[1:5])
	realClientPort := int(binary.BigEndian.Uint16(packet[5:7]))

	switch v := clientAddress.(type) {
	case *net.TCPAddr:
		v.IP = realClientIP
		v.Port = realClientPort
	case *net.UDPAddr:
		v.IP = realClientIP
		v.Port = realClientPort
	}

	return packet[pspp.HeaderSize():], nil
}

func (pspp *PRUDPSimpleProxyProtocol) Encode(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	var ipv4 net.IP
	var port int

	switch v := clientAddress.(type) {
	case *net.TCPAddr:
		ipv4 = v.IP.To4()
		port = v.Port
	case *net.UDPAddr:
		ipv4 = v.IP.To4()
		port = v.Port
	}

	newData := make([]byte, pspp.HeaderSize()+len(packet))
	newData[0] = PRUDP_SIMPLE_PROXY_PROTOCOL_VERSION
	copy(newData[1:5], ipv4)
	binary.BigEndian.PutUint16(newData[5:7], uint16(port))
	copy(newData[7:], packet)

	return newData, nil
}
