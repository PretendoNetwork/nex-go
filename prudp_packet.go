package nex

import (
	"crypto/rc4"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
)

// PRUDPPacket holds all the fields each packet should have in all PRUDP versions
type PRUDPPacket struct {
	server                 *PRUDPServer
	sender                 *PRUDPConnection
	readStream             *ByteStreamIn
	version                uint8
	sourceVirtualPort      VirtualPort
	destinationVirtualPort VirtualPort
	packetType             uint16
	flags                  uint16
	sessionID              uint8
	substreamID            uint8
	signature              []byte
	sequenceID             uint16
	connectionSignature    []byte
	fragmentID             uint8
	payload                []byte
	message                *RMCMessage
	sendCount              uint32
	sentAt                 time.Time
	timeout                *Timeout
}

// SetSender sets the Client who sent the packet
func (p *PRUDPPacket) SetSender(sender ConnectionInterface) {
	p.sender = sender.(*PRUDPConnection)
}

// Sender returns the Client who sent the packet
func (p *PRUDPPacket) Sender() ConnectionInterface {
	return p.sender
}

// Flags returns the packet flags
func (p *PRUDPPacket) Flags() uint16 {
	return p.flags
}

// HasFlag checks if the packet has the given flag
func (p *PRUDPPacket) HasFlag(flag uint16) bool {
	return p.flags&flag != 0
}

// AddFlag adds the given flag to the packet flag bitmask
func (p *PRUDPPacket) AddFlag(flag uint16) {
	p.flags |= flag
}

// SetType sets the packets type
func (p *PRUDPPacket) SetType(packetType uint16) {
	p.packetType = packetType
}

// Type returns the packets type
func (p *PRUDPPacket) Type() uint16 {
	return p.packetType
}

// SetSourceVirtualPortStreamType sets the packets source VirtualPort StreamType
func (p *PRUDPPacket) SetSourceVirtualPortStreamType(streamType constants.StreamType) {
	p.sourceVirtualPort.SetStreamType(streamType)
}

// SourceVirtualPortStreamType returns the packets source VirtualPort StreamType
func (p *PRUDPPacket) SourceVirtualPortStreamType() constants.StreamType {
	return p.sourceVirtualPort.StreamType()
}

// SetSourceVirtualPortStreamID sets the packets source VirtualPort port number
func (p *PRUDPPacket) SetSourceVirtualPortStreamID(port uint8) {
	p.sourceVirtualPort.SetStreamID(port)
}

// SourceVirtualPortStreamID returns the packets source VirtualPort port number
func (p *PRUDPPacket) SourceVirtualPortStreamID() uint8 {
	return p.sourceVirtualPort.StreamID()
}

// SetDestinationVirtualPortStreamType sets the packets destination VirtualPort StreamType
func (p *PRUDPPacket) SetDestinationVirtualPortStreamType(streamType constants.StreamType) {
	p.destinationVirtualPort.SetStreamType(streamType)
}

// DestinationVirtualPortStreamType returns the packets destination VirtualPort StreamType
func (p *PRUDPPacket) DestinationVirtualPortStreamType() constants.StreamType {
	return p.destinationVirtualPort.StreamType()
}

// SetDestinationVirtualPortStreamID sets the packets destination VirtualPort port number
func (p *PRUDPPacket) SetDestinationVirtualPortStreamID(port uint8) {
	p.destinationVirtualPort.SetStreamID(port)
}

// DestinationVirtualPortStreamID returns the packets destination VirtualPort port number
func (p *PRUDPPacket) DestinationVirtualPortStreamID() uint8 {
	return p.destinationVirtualPort.StreamID()
}

// SessionID returns the packets session ID
func (p *PRUDPPacket) SessionID() uint8 {
	return p.sessionID
}

// SetSessionID sets the packets session ID
func (p *PRUDPPacket) SetSessionID(sessionID uint8) {
	p.sessionID = sessionID
}

// SubstreamID returns the packets substream ID
func (p *PRUDPPacket) SubstreamID() uint8 {
	return p.substreamID
}

// SetSubstreamID sets the packets substream ID
func (p *PRUDPPacket) SetSubstreamID(substreamID uint8) {
	p.substreamID = substreamID
}

func (p *PRUDPPacket) setSignature(signature []byte) {
	p.signature = signature
}

// SequenceID returns the packets sequenc ID
func (p *PRUDPPacket) SequenceID() uint16 {
	return p.sequenceID
}

// SetSequenceID sets the packets sequenc ID
func (p *PRUDPPacket) SetSequenceID(sequenceID uint16) {
	p.sequenceID = sequenceID
}

// Payload returns the packets payload
func (p *PRUDPPacket) Payload() []byte {
	return p.payload
}

// SetPayload sets the packets payload
func (p *PRUDPPacket) SetPayload(payload []byte) {
	p.payload = payload
}

func (p *PRUDPPacket) decryptPayload() []byte {
	payload := p.payload

	// TODO - This assumes a reliable DATA packet. Handle unreliable here? Or do that in a different method?
	if p.packetType == constants.DataPacket {
		slidingWindow := p.sender.SlidingWindow(p.SubstreamID())

		payload, _ = slidingWindow.streamSettings.EncryptionAlgorithm.Decrypt(payload)
	}

	return payload
}

func (p *PRUDPPacket) getConnectionSignature() []byte {
	return p.connectionSignature
}

func (p *PRUDPPacket) setConnectionSignature(connectionSignature []byte) {
	p.connectionSignature = connectionSignature
}

func (p *PRUDPPacket) getFragmentID() uint8 {
	return p.fragmentID
}

func (p *PRUDPPacket) setFragmentID(fragmentID uint8) {
	p.fragmentID = fragmentID
}

// RMCMessage returns the packets RMC Message
func (p *PRUDPPacket) RMCMessage() *RMCMessage {
	return p.message
}

// SetRMCMessage sets the packets RMC Message
func (p *PRUDPPacket) SetRMCMessage(message *RMCMessage) {
	p.message = message
}

// SendCount returns the number of times this packet has been sent
func (p *PRUDPPacket) SendCount() uint32 {
	return p.sendCount
}

func (p *PRUDPPacket) incrementSendCount() {
	p.sendCount++
}

// SentAt returns the latest time that this packet has been sent
func (p *PRUDPPacket) SentAt() time.Time {
	return p.sentAt
}

func (p *PRUDPPacket) setSentAt(time time.Time) {
	p.sentAt = time
}

func (p *PRUDPPacket) getTimeout() *Timeout {
	return p.timeout
}

func (p *PRUDPPacket) setTimeout(timeout *Timeout) {
	p.timeout = timeout
}

func (p *PRUDPPacket) processUnreliableCrypto() []byte {
	// * Since unreliable DATA packets can come in out of
	// * order, each packet uses a dedicated RC4 stream
	uniqueKey := p.sender.UnreliablePacketBaseKey[:]
	uniqueKey[0] = byte((uint16(uniqueKey[0]) + p.sequenceID) & 0xFF)
	uniqueKey[1] = byte((uint16(uniqueKey[1]) + (p.sequenceID >> 8)) & 0xFF)
	uniqueKey[31] = byte((uniqueKey[31] + p.sessionID) & 0xFF)

	cipher, _ := rc4.NewCipher(uniqueKey)
	ciphered := make([]byte, len(p.payload))

	cipher.XORKeyStream(ciphered, p.payload)

	return ciphered
}
