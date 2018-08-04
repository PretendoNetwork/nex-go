package nex

import (
	"net"

	General "github.com/PretendoNetwork/nex-go/prudp/General"
)

// Client represents generic NEX/PRUDP client
// _UDPConn           : The address of the client
// State              : The clients connection state
// SignatureKey       : MD5 hash of the servers access key
// SignatureBase      : Packet checksum base (sum on the SignatureKey)
// SecureKey          : NEX server packet event handles
// ConnectionSignature: The clients unique connection signature
// SessionID          : Clients session ID
// Packets            : Packets sent to the server from the client
// PacketQueue        : Packet queue
// SequenceIDIn       : The sequence ID for client->server packets
// SequenceIDOut      : The sequence ID for server->client packets
type Client struct {
	_UDPConn            *net.UDPAddr
	State               int
	SignatureKey        string
	SignatureBase       int
	SecureKey           []byte
	ConnectionSignature []byte
	SessionID           int
	Packets             []General.Packet
	PacketQueue         map[string]General.Packet
	SequenceIDIn        int
	SequenceIDOut       int
}

// NewClient returns a new generic client
func NewClient(addr *net.UDPAddr) Client {

	client := Client{
		_UDPConn:    addr,
		State:       0,
		SessionID:   0,
		PacketQueue: make(map[string]General.Packet),
	}

	return client
}
