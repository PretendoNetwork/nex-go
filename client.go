package nex

import (
	"math/rand"
	"net"
	"time"
)

// Client represents generic NEX/PRUDP client
type Client struct {
	_UDPConn                  *net.UDPAddr
	Server                    *Server
	CipherKey                 string
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
}

// SetCipher sets the client RC4 Cipher
func (client *Client) SetCipherKey(key string) {
	client.CipherKey = key
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

	rand.Seed(time.Now().UnixNano())

	client := Client{
		_UDPConn:                  addr,
		Server:                    server,
		CipherKey:                 "CD&ML",
		ServerConnectionSignature: signature,
		State:       0,
		SessionID:   rand.Intn(0xFF),
		PacketQueue: make(map[string]Packet),
	}

	return client
}
