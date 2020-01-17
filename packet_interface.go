package nex

// PacketInterface implements all Packet methods
type PacketInterface interface {
	GetSender() *Client
	SetVersion(version uint8)
	GetVersion() uint8
	SetSource(source uint8)
	GetSource() uint8
	SetDestination(destination uint8)
	GetDestination() uint8
	SetType(packetType uint16)
	GetType() uint16
	SetFlags(bitmask uint16)
	GetFlags() uint16
	HasFlag(flag uint16) bool
	AddFlag(flag uint16)
	ClearFlag(flag uint16)
	SetSessionID(sessionID uint8)
	GetSessionID() uint8
	SetSignature(signature []byte)
	GetSignature() []byte
	SetSequenceID(sequenceID uint16)
	GetSequenceID() uint16
	SetConnectionSignature(connectionSignature []byte)
	GetConnectionSignature() []byte
	SetFragmentID(fragmentID uint8)
	GetFragmentID() uint8
	SetPayload(payload []byte)
	GetPayload() []byte
	GetRMCRequest() RMCRequest
	Bytes() []byte
}
