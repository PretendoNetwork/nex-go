package nex

import "net"

// PRUDPPacketInterface defines all the methods a PRUDP packet should have
type PRUDPPacketInterface interface {
	Copy() PRUDPPacketInterface
	Version() int
	Bytes() []byte
	Sender() ClientInterface
	Flags() uint16
	HasFlag(flag uint16) bool
	AddFlag(flag uint16)
	SetType(packetType uint16)
	Type() uint16
	SetSourceStreamType(sourceStreamType uint8)
	SourceStreamType() uint8
	SetSourcePort(sourcePort uint8)
	SourcePort() uint8
	SetDestinationStreamType(destinationStreamType uint8)
	DestinationStreamType() uint8
	SetDestinationPort(destinationPort uint8)
	DestinationPort() uint8
	SessionID() uint8
	SetSessionID(sessionID uint8)
	SubstreamID() uint8
	SetSubstreamID(substreamID uint8)
	SequenceID() uint16
	SetSequenceID(sequenceID uint16)
	Payload() []byte
	SetPayload(payload []byte)
	RMCMessage() *RMCMessage
	SetRMCMessage(message *RMCMessage)
	decode() error
	setSignature(signature []byte)
	calculateConnectionSignature(addr net.Addr) ([]byte, error)
	calculateSignature(sessionKey, connectionSignature []byte) []byte
	decryptPayload() []byte
	getConnectionSignature() []byte
	setConnectionSignature(connectionSignature []byte)
	getFragmentID() uint8
	setFragmentID(fragmentID uint8)
}
