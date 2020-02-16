package nex

import (
	"crypto/rc4"
	"net"
)

// Client represents a connected or non-connected PRUDP client
type Client struct {
	address                   *net.UDPAddr
	server                    *Server
	cipher                    *rc4.Cipher
	decipher                  *rc4.Cipher
	signatureKey              []byte
	signatureBase             int
	secureKey                 []byte
	serverConnectionSignature []byte
	clientConnectionSignature []byte
	sessionID                 int
	sessionKey                []byte
	sequenceIDIn              *Counter
	sequenceIDOut             *Counter
}

// Reset resets the Client to default values
func (client *Client) Reset() {
	client.sequenceIDIn = NewCounter(0)
	client.sequenceIDOut = NewCounter(0)

	client.UpdateAccessKey(client.GetServer().GetAccessKey())
	client.UpdateRC4Key([]byte("CD&ML"))

	if client.GetServer().GetPrudpVersion() == 0 {
		client.SetServerConnectionSignature(make([]byte, 4))
		client.SetClientConnectionSignature(make([]byte, 4))
	} else {
		client.SetServerConnectionSignature([]byte{})
		client.SetClientConnectionSignature([]byte{})
	}
}

// GetAddress returns the clients UDP address
func (client *Client) GetAddress() *net.UDPAddr {
	return client.address
}

// GetServer returns the server the client is currently connected to
func (client *Client) GetServer() *Server {
	return client.server
}

// UpdateRC4Key sets the client RC4 stream key
func (client *Client) UpdateRC4Key(RC4Key []byte) {
	cipher, _ := rc4.NewCipher(RC4Key)
	client.cipher = cipher

	decipher, _ := rc4.NewCipher(RC4Key)
	client.decipher = decipher
}

// GetCipher returns the RC4 cipher stream for out-bound packets
func (client *Client) GetCipher() *rc4.Cipher {
	return client.cipher
}

// GetDecipher returns the RC4 cipher stream for in-bound packets
func (client *Client) GetDecipher() *rc4.Cipher {
	return client.decipher
}

// UpdateAccessKey sets the client signature base and signature key
func (client *Client) UpdateAccessKey(accessKey string) {
	client.signatureBase = sum([]byte(accessKey))
	client.signatureKey = MD5Hash([]byte(accessKey))
}

// GetSignatureBase returns the v0 checksum signature base
func (client *Client) GetSignatureBase() int {
	return client.signatureBase
}

// GetSignatureKey returns signature key
func (client *Client) GetSignatureKey() []byte {
	return client.signatureKey
}

// SetServerConnectionSignature sets the clients server-side connection signature
func (client *Client) SetServerConnectionSignature(serverConnectionSignature []byte) {
	client.serverConnectionSignature = serverConnectionSignature
}

// GetServerConnectionSignature returns the clients server-side connection signature
func (client *Client) GetServerConnectionSignature() []byte {
	return client.serverConnectionSignature
}

// SetClientConnectionSignature sets the clients client-side connection signature
func (client *Client) SetClientConnectionSignature(clientConnectionSignature []byte) {
	client.clientConnectionSignature = clientConnectionSignature
}

// GetClientConnectionSignature returns the clients client-side connection signature
func (client *Client) GetClientConnectionSignature() []byte {
	return client.clientConnectionSignature
}

// GetSequenceIDCounterOut returns the clients packet SequenceID counter for out-going packets
func (client *Client) GetSequenceIDCounterOut() *Counter {
	return client.sequenceIDOut
}

// GetSequenceIDCounterIn returns the clients packet SequenceID counter for incoming packets
func (client *Client) GetSequenceIDCounterIn() *Counter {
	return client.sequenceIDIn
}

// SetSessionKey sets the clients session key
func (client *Client) SetSessionKey(sessionKey []byte) {
	client.sessionKey = sessionKey
}

// GetSessionKey returns the clients session key
func (client *Client) GetSessionKey() []byte {
	return client.sessionKey
}

// NewClient returns a new PRUDP client
func NewClient(address *net.UDPAddr, server *Server) *Client {
	client := &Client{
		address: address,
		server:  server,
	}

	client.Reset()

	return client
}
