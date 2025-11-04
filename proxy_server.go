package nex

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type ProxyServer struct {
	mappings map[int]*ProxyMapping
	wg       sync.WaitGroup
}

type ProxyMapping struct {
	listenPort int
	targetAddr *net.UDPAddr
	conn       *net.UDPConn
}

func (ps *ProxyServer) handlePort(mapping *ProxyMapping) {
	defer ps.wg.Done()

	buffer := make([]byte, 64000)

	for {
		read, addr, err := mapping.conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Errorf("Read error on port %d: %s\n", mapping.listenPort, err)
			return
		}

		data := make([]byte, read)
		copy(data, buffer[:read])

		if addr.String() == mapping.targetAddr.String() {
			ps.handleServerPacket(mapping, data)
		} else {
			ps.handleClientPacket(mapping, data, addr)
		}
	}
}

// handleClientPacket prepends client IP:port and forwards to upstream server
func (ps *ProxyServer) handleClientPacket(mapping *ProxyMapping, data []byte, clientAddr *net.UDPAddr) {
	ip4 := clientAddr.IP.To4()
	if ip4 == nil {
		logger.Warningf("Warning: non-IPv4 address: %s\n", clientAddr.String())
		return
	}

	packet := make([]byte, 6+len(data))

	copy(packet[0:4], ip4)
	binary.BigEndian.PutUint16(packet[4:6], uint16(clientAddr.Port))
	copy(packet[6:], data)

	_, err := mapping.conn.WriteToUDP(packet, mapping.targetAddr)
	if err != nil {
		logger.Errorf("Error forwarding to server: %s\n", err)
	}
}

// handleServerPacket extracts destination address and forwards payload to client
func (ps *ProxyServer) handleServerPacket(mapping *ProxyMapping, data []byte) {
	if len(data) < 6 {
		logger.Warningf("Warning: packet too short (%d bytes)\n", len(data))
		return
	}

	clientIP := net.IPv4(data[0], data[1], data[2], data[3])
	clientPort := int(binary.BigEndian.Uint16(data[4:6]))
	clientAddr := &net.UDPAddr{
		IP:   clientIP,
		Port: clientPort,
	}

	payload := data[6:]

	_, err := mapping.conn.WriteToUDP(payload, clientAddr)
	if err != nil {
		logger.Errorf("Error forwarding to client: %s\n", err)
	}
}

// AddMapping adds a port mapping (listenPort -> targetAddr)
func (ps *ProxyServer) AddMapping(listenPort int, targetAddr string) error {
	target, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve target address %s: %w", targetAddr, err)
	}

	listen := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: listenPort,
	}

	conn, err := net.ListenUDP("udp", listen)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", listenPort, err)
	}

	mapping := &ProxyMapping{
		listenPort: listenPort,
		targetAddr: target,
		conn:       conn,
	}

	ps.mappings[listenPort] = mapping

	return nil
}

// Start begins listening on all mapped ports
func (ps *ProxyServer) Start() {
	for _, mapping := range ps.mappings {
		ps.wg.Add(1)
		go ps.handlePort(mapping)
	}
	ps.wg.Wait()
}

// NewProxyServer creates a new UDP proxy server
func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		mappings: make(map[int]*ProxyMapping),
	}
}
