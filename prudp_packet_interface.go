package nex

import (
	"net"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
)

// PRUDPPacketInterface defines all the methods a PRUDP packet should have
type PRUDPPacketInterface interface {
	Copy() PRUDPPacketInterface
	Version() int
	Bytes() []byte
	SetSender(sender ConnectionInterface)
	Sender() ConnectionInterface
	Flags() uint16
	HasFlag(flag uint16) bool
	AddFlag(flag uint16)
	SetType(packetType uint16)
	Type() uint16
	SetSourceVirtualPortStreamType(streamType constants.StreamType)
	SourceVirtualPortStreamType() constants.StreamType
	SetSourceVirtualPortStreamID(port uint8)
	SourceVirtualPortStreamID() uint8
	SetDestinationVirtualPortStreamType(streamType constants.StreamType)
	DestinationVirtualPortStreamType() constants.StreamType
	SetDestinationVirtualPortStreamID(port uint8)
	DestinationVirtualPortStreamID() uint8
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
	SendCount() uint32
	incrementSendCount()
	SentAt() time.Time
	setSentAt(time time.Time)
	getTimeout() *Timeout
	setTimeout(timeout *Timeout)
	decode() error
	setSignature(signature []byte)
	calculateConnectionSignature(addr net.Addr) ([]byte, error)
	calculateSignature(sessionKey, connectionSignature []byte) []byte
	decryptPayload() []byte
	getConnectionSignature() []byte
	setConnectionSignature(connectionSignature []byte)
	getFragmentID() uint8
	setFragmentID(fragmentID uint8)
	processUnreliableCrypto() []byte
}
