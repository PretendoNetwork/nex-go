package nex

import (
	"crypto/rc4"
	"net"
)

// Client represents generic NEX/PRUDP client
type Client struct {
	_UDPConn                  *net.UDPAddr
	Server                    *Server
	Cipher                    *rc4.Cipher
	Decipher                  *rc4.Cipher
	State                     int
	SignatureKey              string
	SignatureBase             int
	SecureKey                 []byte
	ServerConnectionSignature []byte
	ClientConnectionSignature []byte
	SessionID                 int
	Packets                   []Packet
	PacketQueue               map[string]Packet
	SequenceIDIn              Counter
	SequenceIDOut             Counter
	RMCCallID                 uint32
}

// SetCipher sets the client RC4 Cipher
func (client *Client) SetCipher(key string) {
	client.SecureKey = []byte(key)
	cipher, _ := rc4.NewCipher([]byte(key))
	client.Cipher = cipher
}

// NewClient returns a new generic client
func NewClient(addr *net.UDPAddr, server *Server) Client {
	cipher, _ := rc4.NewCipher([]byte("CD&ML"))
	decipher, _ := rc4.NewCipher([]byte("CD&ML"))

	client := Client{
		_UDPConn:    addr,
		Server:      server,
		Cipher:      cipher,
		Decipher:    decipher,
		State:       0,
		SessionID:   0,
		PacketQueue: make(map[string]Packet),
	}

	return client
}
