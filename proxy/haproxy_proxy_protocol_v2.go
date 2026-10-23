package proxy

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	HAPROXY_V2_SIGNATURE = "\x0D\x0A\x0D\x0A\x00\x0D\x0A\x51\x55\x49\x54\x0A"

	// * Version and Command
	HAPROXY_V2_VERSION   = 0x2
	HAPROXY_V2_CMD_LOCAL = 0x0
	HAPROXY_V2_CMD_PROXY = 0x1

	// * Address families
	HAPROXY_V2_AF_UNSPEC = 0x0
	HAPROXY_V2_AF_INET   = 0x1
	HAPROXY_V2_AF_INET6  = 0x2
	HAPROXY_V2_AF_UNIX   = 0x3

	// * Transport protocols
	HAPROXY_V2_TRANSPORT_UNSPEC = 0x0
	HAPROXY_V2_TRANSPORT_STREAM = 0x1
	HAPROXY_V2_TRANSPORT_DGRAM  = 0x2

	// * Combined protocol bytes
	HAPROXY_V2_PROTO_UNSPEC      = 0x00
	HAPROXY_V2_PROTO_TCP4        = 0x11
	HAPROXY_V2_PROTO_UDP4        = 0x12
	HAPROXY_V2_PROTO_TCP6        = 0x21
	HAPROXY_V2_PROTO_UDP6        = 0x22
	HAPROXY_V2_PROTO_UNIX_STREAM = 0x31
	HAPROXY_V2_PROTO_UNIX_DGRAM  = 0x32
)

// HAProxyProxyProtocolV2 implements HAProxys PROXY protocol version 2 (binary format)
// https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt#:~:text=2.2.%20Binary%20header%20format%20(version%202)
type HAProxyProxyProtocolV2 struct{}

func (hpp *HAProxyProxyProtocolV2) HeaderSize() int {
	// * V2 is variable length (16 bytes minimum + address length)
	// * Return -1 to indicate variable length
	return -1
}

func (hpp *HAProxyProxyProtocolV2) Parse(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	// TODO - Add validation checks here

	versionCommand := packet[12]
	version := (versionCommand >> 4) & 0x0F
	command := versionCommand & 0x0F

	if version != HAPROXY_V2_VERSION {
		return nil, fmt.Errorf("unsupported PROXY protocol version: 0x%x", version)
	}

	family := packet[13]
	addressLengthgth := binary.BigEndian.Uint16(packet[14:16])
	totalHeaderLen := 16 + int(addressLengthgth)

	if command == HAPROXY_V2_CMD_LOCAL {
		return packet[totalHeaderLen:], nil
	}

	addressData := packet[16:totalHeaderLen]

	switch family {
	case HAPROXY_V2_PROTO_TCP4, HAPROXY_V2_PROTO_UDP4:
		realClientIP := net.IPv4(addressData[0], addressData[1], addressData[2], addressData[3])
		realClientPort := int(binary.BigEndian.Uint16(addressData[8:10]))

		switch v := clientAddress.(type) {
		case *net.TCPAddr:
			v.IP = realClientIP
			v.Port = realClientPort
		case *net.UDPAddr:
			v.IP = realClientIP
			v.Port = realClientPort
		}
	case HAPROXY_V2_PROTO_TCP6, HAPROXY_V2_PROTO_UDP6:
		realClientIP := net.IP(addressData[0:16])
		realClientPort := int(binary.BigEndian.Uint16(addressData[32:34]))

		switch v := clientAddress.(type) {
		case *net.TCPAddr:
			v.IP = realClientIP
			v.Port = realClientPort
		case *net.UDPAddr:
			v.IP = realClientIP
			v.Port = realClientPort
		}
	}

	return packet[totalHeaderLen:], nil
}

func (hpp *HAProxyProxyProtocolV2) Encode(clientAddress net.Addr, proxyAddress net.Addr, packet []byte) ([]byte, error) {
	var clientIP net.IP
	var clientPort int
	var proxyIP net.IP
	var proxyPort int
	var protocol byte
	var addressLength uint16

	switch v := clientAddress.(type) {
	case *net.TCPAddr:
		clientIP = v.IP
		clientPort = v.Port
	case *net.UDPAddr:
		clientIP = v.IP
		clientPort = v.Port
	}

	// TODO - I'm almost positive this is wrong, the PROXY protocol docs say this is the "destination" but it's unclear if that's the proxy server or ustream server
	switch v := proxyAddress.(type) {
	case *net.TCPAddr:
		proxyIP = v.IP
		proxyPort = v.Port
	case *net.UDPAddr:
		proxyIP = v.IP
		proxyPort = v.Port
	}

	var addressData []byte
	if clientIP.To4() != nil {
		clientIP = clientIP.To4()
		proxyIP = proxyIP.To4()

		switch clientAddress.(type) {
		case *net.TCPAddr:
			protocol = HAPROXY_V2_PROTO_TCP4
		case *net.UDPAddr:
			protocol = HAPROXY_V2_PROTO_UDP4
		}

		addressLength = 12
		addressData = make([]byte, 12)
		copy(addressData[0:4], clientIP)
		copy(addressData[4:8], proxyIP)
		binary.BigEndian.PutUint16(addressData[8:10], uint16(clientPort))
		binary.BigEndian.PutUint16(addressData[10:12], uint16(proxyPort))
	} else {
		clientIP = clientIP.To16()
		proxyIP = proxyIP.To16()

		switch clientAddress.(type) {
		case *net.TCPAddr:
			protocol = HAPROXY_V2_PROTO_TCP6
		case *net.UDPAddr:
			protocol = HAPROXY_V2_PROTO_UDP6
		}

		addressLength = 36
		addressData = make([]byte, 36)
		copy(addressData[0:16], clientIP)
		copy(addressData[16:32], proxyIP)
		binary.BigEndian.PutUint16(addressData[32:34], uint16(clientPort))
		binary.BigEndian.PutUint16(addressData[34:36], uint16(proxyPort))
	}

	header := make([]byte, 16)
	copy(header[0:12], []byte(HAPROXY_V2_SIGNATURE))
	header[12] = (HAPROXY_V2_VERSION << 4) | HAPROXY_V2_CMD_PROXY
	header[13] = protocol
	binary.BigEndian.PutUint16(header[14:16], addressLength)

	newData := make([]byte, 16+int(addressLength)+len(packet))
	copy(newData[0:16], header)
	copy(newData[16:16+addressLength], addressData)
	copy(newData[16+addressLength:], packet)

	return newData, nil
}
