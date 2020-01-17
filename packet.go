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

// GetSender returns the packet sender
func (packet *Packet) GetSender() *Client {
	return packet.sender
}

// SetVersion sets the packet PRUDP version
func (packet *Packet) SetVersion(version uint8) {
	packet.version = version
}

// GetVersion gets the packet PRUDP version
func (packet *Packet) GetVersion() uint8 {
	return packet.version
}

// SetSource sets the packet source
func (packet *Packet) SetSource(source uint8) {
	packet.source = source
}

// GetSource returns the packet source
func (packet *Packet) GetSource() uint8 {
	return packet.source
}

// SetDestination sets the packet destination
func (packet *Packet) SetDestination(destination uint8) {
	packet.destination = destination
}

// GetDestination returns the packet destination
func (packet *Packet) GetDestination() uint8 {
	return packet.destination
}

// SetType sets the packet type
func (packet *Packet) SetType(packetType uint16) {
	packet.packetType = packetType
}

// GetType returns the packet type
func (packet *Packet) GetType() uint16 {
	return packet.packetType
}

// SetFlags sets the packet flag bitmask
func (packet *Packet) SetFlags(bitmask uint16) {
	packet.flags = bitmask
}

// GetFlags returns the packet flag bitmask
func (packet *Packet) GetFlags() uint16 {
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

// GetSessionID returns the packet sessionID
func (packet *Packet) GetSessionID() uint8 {
	return packet.sessionID
}

// SetSignature sets the packet signature
func (packet *Packet) SetSignature(signature []byte) {
	packet.signature = signature
}

// GetSignature returns the packet signature
func (packet *Packet) GetSignature() []byte {
	return packet.signature
}

// SetSequenceID sets the packet sequenceID
func (packet *Packet) SetSequenceID(sequenceID uint16) {
	packet.sequenceID = sequenceID
}

// GetSequenceID returns the packet sequenceID
func (packet *Packet) GetSequenceID() uint16 {
	return packet.sequenceID
}

// SetConnectionSignature sets the packet connection signature
func (packet *Packet) SetConnectionSignature(connectionSignature []byte) {
	packet.connectionSignature = connectionSignature
}

// GetConnectionSignature returns the packet connection signature
func (packet *Packet) GetConnectionSignature() []byte {
	return packet.connectionSignature
}

// SetFragmentID sets the packet fragmentID
func (packet *Packet) SetFragmentID(fragmentID uint8) {
	packet.fragmentID = fragmentID
}

// GetFragmentID returns the packet fragmentID
func (packet *Packet) GetFragmentID() uint8 {
	return packet.fragmentID
}

// SetPayload sets the packet payload
func (packet *Packet) SetPayload(payload []byte) {
	packet.payload = payload
}

// GetPayload returns the packet payload
func (packet *Packet) GetPayload() []byte {
	return packet.payload
}

// GetRMCRequest returns the packet RMC request
func (packet *Packet) GetRMCRequest() RMCRequest {
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
