package nex

import "net"

// TODO - We can also breakout the decoding/encoding functions here too, but that would require getters and setters for all packet fields

// PRUDPV1Settings defines settings for how to handle aspects of PRUDPv1 packets
type PRUDPV1Settings struct {
	LegacyConnectionSignature     bool
	ConnectionSignatureCalculator func(packet *PRUDPPacketV1, addr net.Addr) ([]byte, error)
	SignatureCalculator           func(packet *PRUDPPacketV1, sessionKey, connectionSignature []byte) []byte
}

// NewPRUDPV1Settings returns a new PRUDPV1Settings
func NewPRUDPV1Settings() *PRUDPV1Settings {
	return &PRUDPV1Settings{
		LegacyConnectionSignature:     false,
		ConnectionSignatureCalculator: defaultPRUDPv1ConnectionSignature,
		SignatureCalculator:           defaultPRUDPv1CalculateSignature,
	}
}
