package nex

// PRUDPPacket holds all the fields each packet should have in all PRUDP versions
type PRUDPPacket struct {
	sender                *PRUDPClient
	readStream            *StreamIn
	sourceStreamType      uint8
	sourcePort            uint8
	destinationStreamType uint8
	destinationPort       uint8
	packetType            uint16
	flags                 uint16
	sessionID             uint8
	substreamID           uint8
	signature             []byte
	sequenceID            uint16
	connectionSignature   []byte
	fragmentID            uint8
	payload               []byte
	message               *RMCMessage
}

// Sender returns the Client who sent the packet
func (p *PRUDPPacket) Sender() ClientInterface {
	return p.sender
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

// SetSourceStreamType sets the packet virtual source stream type
func (p *PRUDPPacket) SetSourceStreamType(sourceStreamType uint8) {
	p.sourceStreamType = sourceStreamType
}

// SourceStreamType returns the packet virtual source stream type
func (p *PRUDPPacket) SourceStreamType() uint8 {
	return p.sourceStreamType
}

// SetSourcePort sets the packet virtual source stream type
func (p *PRUDPPacket) SetSourcePort(sourcePort uint8) {
	p.sourcePort = sourcePort
}

// SourcePort returns the packet virtual source stream type
func (p *PRUDPPacket) SourcePort() uint8 {
	return p.sourcePort
}

// SetDestinationStreamType sets the packet virtual destination stream type
func (p *PRUDPPacket) SetDestinationStreamType(destinationStreamType uint8) {
	p.destinationStreamType = destinationStreamType
}

// DestinationStreamType returns the packet virtual destination stream type
func (p *PRUDPPacket) DestinationStreamType() uint8 {
	return p.destinationStreamType
}

// SetDestinationPort sets the packet virtual destination port
func (p *PRUDPPacket) SetDestinationPort(destinationPort uint8) {
	p.destinationPort = destinationPort
}

// DestinationPort returns the packet virtual destination port
func (p *PRUDPPacket) DestinationPort() uint8 {
	return p.destinationPort
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
	if p.packetType == DataPacket {
		substream := p.sender.reliableSubstream(p.SubstreamID())

		payload = substream.Decrypt(payload)
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
