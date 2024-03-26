package nex

import "net"

// TODO - We can also breakout the decoding/encoding functions here too, but that would require getters and setters for all packet fields

// PRUDPV0Settings defines settings for how to handle aspects of PRUDPv0 packets
type PRUDPV0Settings struct {
	IsQuazalMode                  bool
	EncryptedConnect              bool
	LegacyConnectionSignature     bool
	UseEnhancedChecksum           bool
	ConnectionSignatureCalculator func(packet *PRUDPPacketV0, addr net.Addr) ([]byte, error)
	SignatureCalculator           func(packet *PRUDPPacketV0, sessionKey, connectionSignature []byte) []byte
	DataSignatureCalculator       func(packet *PRUDPPacketV0, sessionKey []byte) []byte
	ChecksumCalculator            func(packet *PRUDPPacketV0, data []byte) uint32
}

// NewPRUDPV0Settings returns a new PRUDPV0Settings
func NewPRUDPV0Settings() *PRUDPV0Settings {
	return &PRUDPV0Settings{
		IsQuazalMode:                  false,
		EncryptedConnect:              false,
		LegacyConnectionSignature:     false,
		UseEnhancedChecksum:           false,
		ConnectionSignatureCalculator: defaultPRUDPv0ConnectionSignature,
		SignatureCalculator:           defaultPRUDPv0CalculateSignature,
		DataSignatureCalculator:       defaultPRUDPv0CalculateDataSignature,
		ChecksumCalculator:            defaultPRUDPv0CalculateChecksum,
	}
}
