package nex

// Packet represents a generic PRUDP packet
type Packet struct {
	sender              *Client
	data                []byte
	version             uint8
	source              uint8
	destination         uint8
	packetType          uint16
	flags               uint16
	sessionID           uint8
	signature           []byte
	sequenceID          uint16
	connectionSignature []byte
	fragmentID          uint8
	payload             []byte
	rmcRequest          RMCRequest
	PacketInterface
}

// Data returns bytes used to create the packet (this is not the same as Bytes())
func (packet *Packet) Data() []byte {
	return packet.data
}

// Sender returns the packet sender
func (packet *Packet) Sender() *Client {
	return packet.sender
}

// SetVersion sets the packet PRUDP version
func (packet *Packet) SetVersion(version uint8) {
	packet.version = version
}

// Version gets the packet PRUDP version
func (packet *Packet) Version() uint8 {
	return packet.version
}

// SetSource sets the packet source
func (packet *Packet) SetSource(source uint8) {
	packet.source = source
}

// Source returns the packet source
func (packet *Packet) Source() uint8 {
	return packet.source
}

// SetDestination sets the packet destination
func (packet *Packet) SetDestination(destination uint8) {
	packet.destination = destination
}

// Destination returns the packet destination
func (packet *Packet) Destination() uint8 {
	return packet.destination
}

// SetType sets the packet type
func (packet *Packet) SetType(packetType uint16) {
	packet.packetType = packetType
}

// Type returns the packet type
func (packet *Packet) Type() uint16 {
	return packet.packetType
}

// SetFlags sets the packet flag bitmask
func (packet *Packet) SetFlags(bitmask uint16) {
	packet.flags = bitmask
}

// Flags returns the packet flag bitmask
func (packet *Packet) Flags() uint16 {
	return packet.flags
}

// HasFlag checks if the packet has the given flag
func (packet *Packet) HasFlag(flag uint16) bool {
	return packet.flags&flag != 0
}

// AddFlag adds the given flag to the packet flag bitmask
func (packet *Packet) AddFlag(flag uint16) {
	packet.flags |= flag
}

// ClearFlag removes the given flag from the packet bitmask
func (packet *Packet) ClearFlag(flag uint16) {
	packet.flags &^= flag
}

// SetSessionID sets the packet sessionID
func (packet *Packet) SetSessionID(sessionID uint8) {
	packet.sessionID = sessionID
}

// SessionID returns the packet sessionID
func (packet *Packet) SessionID() uint8 {
	return packet.sessionID
}

// SetSignature sets the packet signature
func (packet *Packet) SetSignature(signature []byte) {
	packet.signature = signature
}

// Signature returns the packet signature
func (packet *Packet) Signature() []byte {
	return packet.signature
}

// SetSequenceID sets the packet sequenceID
func (packet *Packet) SetSequenceID(sequenceID uint16) {
	packet.sequenceID = sequenceID
}

// SequenceID returns the packet sequenceID
func (packet *Packet) SequenceID() uint16 {
	return packet.sequenceID
}

// SetConnectionSignature sets the packet connection signature
func (packet *Packet) SetConnectionSignature(connectionSignature []byte) {
	packet.connectionSignature = connectionSignature
}

// ConnectionSignature returns the packet connection signature
func (packet *Packet) ConnectionSignature() []byte {
	return packet.connectionSignature
}

// SetFragmentID sets the packet fragmentID
func (packet *Packet) SetFragmentID(fragmentID uint8) {
	packet.fragmentID = fragmentID
}

// FragmentID returns the packet fragmentID
func (packet *Packet) FragmentID() uint8 {
	return packet.fragmentID
}

// SetPayload sets the packet payload
func (packet *Packet) SetPayload(payload []byte) {
	packet.payload = payload
}

// Payload returns the packet payload
func (packet *Packet) Payload() []byte {
	return packet.payload
}

// RMCRequest returns the packet RMC request
func (packet *Packet) RMCRequest() RMCRequest {
	return packet.rmcRequest
}

// NewPacket returns a new PRUDP packet generic
func NewPacket(client *Client, data []byte) Packet {
	packet := Packet{
		sender:  client,
		data:    data,
		payload: []byte{},
	}

	return packet
}
