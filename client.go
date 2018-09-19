package nex

import (
	"crypto/rc4"
	"math/rand"
	"net"
	"time"
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
	cipher, _ := rc4.NewCipher([]byte(key))
	client.Cipher = cipher
}

// NewClient returns a new generic client
func NewClient(addr *net.UDPAddr, server *Server) Client {
	rand.Seed(time.Now().UnixNano())

	var signature []byte
	if server.Settings.PrudpVersion == 0 {
		signature = make([]byte, 4)
	} else {
		signature = make([]byte, 16)
	}

	rand.Read(signature)

	cipher, _ := rc4.NewCipher([]byte("CD&ML"))
	decipher, _ := rc4.NewCipher([]byte("CD&ML"))

	client := Client{
		_UDPConn:                  addr,
		Server:                    server,
		Cipher:                    cipher,
		Decipher:                  decipher,
		ServerConnectionSignature: signature,
		State:       0,
		SessionID:   0,
		PacketQueue: make(map[string]Packet),
	}

	return client
}
