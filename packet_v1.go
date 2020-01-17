package nex

// PacketV1 reresents a PRUDPv1 packet
type PacketV1 struct {
	Packet
	magic              [2]byte
	substreamID        int8
	supportedFunctions int32
	initialSequenceID  int16
	maximumSubstreamID int8
}

// Decode decodes the packet
func (packet *PacketV1) Decode() {}

// Bytes encodes the packet and returns a byte array
func (packet *PacketV1) Bytes() []byte {
	return []byte{}
}

// NewPacketV1 returns a new PRUDPv1 packet
func NewPacketV1(client *Client, data []byte) *PacketV1 {
	packet := NewPacket(client, data)
	packetv1 := PacketV1{Packet: packet}

	if data != nil {
		packetv1.Decode()
	}

	return &packetv1
}
