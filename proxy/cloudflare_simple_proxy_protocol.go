package proxy

import (
	"encoding/binary"
	"net"
)

const CLOUDFLARE_SIMPLE_PROXY_PROTOCOL_MAGIC = 0x56EC

// CloudflareSimpleProxyProtocol implements Cloudflares Simple Proxy Protocol
// https://developers.cloudflare.com/spectrum/reference/simple-proxy-protocol-header/
type CloudflareSimpleProxyProtocol struct{}

func (cspp *CloudflareSimpleProxyProtocol) HeaderSize() int {
	return 38
}

func (cspp *CloudflareSimpleProxyProtocol) Parse(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	realClientIP := net.IP(packet[2:18])
	realProxyIP := net.IP(packet[18:34])
	realClientPort := int(binary.BigEndian.Uint16(packet[34:36]))
	realProxyPort := int(binary.BigEndian.Uint16(packet[36:38]))

	switch v := clientAddress.(type) {
	case *net.TCPAddr:
		v.IP = realClientIP
		v.Port = realClientPort
	case *net.UDPAddr:
		v.IP = realClientIP
		v.Port = realClientPort
	}

	switch v := proxyAddress.(type) {
	case *net.TCPAddr:
		v.IP = realProxyIP
		v.Port = realProxyPort
	case *net.UDPAddr:
		v.IP = realProxyIP
		v.Port = realProxyPort
	}

	return packet[cspp.HeaderSize():], nil
}

func (cspp *CloudflareSimpleProxyProtocol) Encode(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	var clientIP net.IP
	var clientPort int
	var proxyIP net.IP
	var proxyPort int

	switch v := clientAddress.(type) {
	case *net.TCPAddr:
		clientIP = v.IP
		clientPort = v.Port
	case *net.UDPAddr:
		clientIP = v.IP
		clientPort = v.Port
	}

	switch v := proxyAddress.(type) {
	case *net.TCPAddr:
		proxyIP = v.IP
		proxyPort = v.Port
	case *net.UDPAddr:
		proxyIP = v.IP
		proxyPort = v.Port
	}

	if clientIP.To4() != nil {
		clientIP = clientIP.To16()
	}

	if proxyIP.To4() != nil {
		proxyIP = proxyIP.To16()
	}

	newData := make([]byte, cspp.HeaderSize()+len(packet))
	binary.BigEndian.PutUint16(newData[0:2], CLOUDFLARE_SIMPLE_PROXY_PROTOCOL_MAGIC)
	copy(newData[2:18], clientIP)
	copy(newData[18:34], proxyIP)
	binary.BigEndian.PutUint16(newData[34:36], uint16(clientPort))
	binary.BigEndian.PutUint16(newData[36:38], uint16(proxyPort))
	copy(newData[38:], packet)

	return newData, nil
}
