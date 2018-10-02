package nex

import (
	"bytes"
	"fmt"
)

// Packet represents a generic PRUDP packet of any type
type Packet struct {
	Sender          *Client
	Data            []byte
	Version         int
	Type            uint16
	Flags           uint16
	Source          uint8
	Destination     uint8
	SourceType      uint8
	SourcePort      uint8
	DestinationType uint8
	DestinationPort uint8
	SessionID       uint8
	Signature       []byte
	SequenceID      uint16
	FragmentID      int
	Payload         []byte
	RMCRequest      RMCRequest // only present in Data packets
	MultiAckVersion uint8      // only present in v1
	Checksum        int        // only present in v0
}

// Go doesn't allow byte arrays, even with fixed sizes, to be type `const` :L
// Casting to `var` instead
var (
	PRUDPV1Magic   = []byte{0xEA, 0xD0}
	PRUDPLiteMagic = []byte{0x80}
)

// FromBytes converts a byte array to a generic PRUDP packet
func (PRUDPPacket *Packet) FromBytes(Data []byte) {
	PRUDPPacket.Data = Data
	//header := PRUDPPacket.decodeHeader()

	var decoded map[string]interface{}

	if bytes.Equal(PRUDPPacket.Data[:2], PRUDPV1Magic) {
		PRUDPPacket.Version = 1
		decoded = decodePacketV1(PRUDPPacket)
		PRUDPPacket.MultiAckVersion = decoded["MultiAckVersion"].(uint8)
	} else if bytes.Equal(PRUDPPacket.Data[:1], PRUDPLiteMagic) {
		PRUDPPacket.Version = 2
		// handle Lite?
		// we aren't doing Switch stuff but I dunno why not have a check anyway!
	} else {
		// assume v0?
		// v0 doesn't have a dedicated magic, and the first 2 bytes aren't really static
		PRUDPPacket.Version = 0
		decoded = decodePacketV0(PRUDPPacket)
		PRUDPPacket.Checksum = decoded["Checksum"].(int)
	}

	PRUDPPacket.Type = decoded["Type"].(uint16)
	PRUDPPacket.Flags = decoded["Flags"].(uint16)
	PRUDPPacket.SourceType = decoded["SourceType"].(uint8)
	PRUDPPacket.SourcePort = decoded["SourcePort"].(uint8)
	PRUDPPacket.DestinationType = decoded["DestinationType"].(uint8)
	PRUDPPacket.DestinationPort = decoded["DestinationPort"].(uint8)
	PRUDPPacket.FragmentID = decoded["FragmentID"].(int)

	PRUDPPacket.Source = decoded["Source"].(uint8)
	PRUDPPacket.Destination = decoded["Destination"].(uint8)
	PRUDPPacket.SequenceID = decoded["SequenceID"].(uint16)
	PRUDPPacket.SessionID = decoded["SessionID"].(uint8)
	PRUDPPacket.Signature = decoded["Signature"].([]byte)

	PRUDPPacket.Payload = decoded["Payload"].([]byte)

	if decoded["RMCRequest"] != nil {
		PRUDPPacket.RMCRequest = decoded["RMCRequest"].(RMCRequest)
	}
}

// Bytes converts a PRUDP packet to a byte array
func (PRUDPPacket *Packet) Bytes() []byte {

	var encoded []byte
	switch PRUDPPacket.Version {
	case 0:
		encoded = encodePacketV0(PRUDPPacket)
	case 1:
		encoded = encodePacketV1(PRUDPPacket)
	case 2: // Lite (Switch)
		encoded = []byte{}
	default:
		fmt.Println("Unknown PRUDP version:", PRUDPPacket.Version)
		encoded = []byte{}
	}

	return encoded
}

// SetVersion sets the packets Version property
func (PRUDPPacket *Packet) SetVersion(Version int) {
	PRUDPPacket.Version = Version
}

// SetSource sets the packets Source property
func (PRUDPPacket *Packet) SetSource(Source uint8) {
	PRUDPPacket.Source = Source
}

// SetDestination sets the packets Destination property
func (PRUDPPacket *Packet) SetDestination(Destination uint8) {
	PRUDPPacket.Destination = Destination
}

// SetSessionID sets the packets SessionID property
func (PRUDPPacket *Packet) SetSessionID(SessionID uint8) {
	PRUDPPacket.SessionID = SessionID
}

// SetSignature sets the packets Signature property
func (PRUDPPacket *Packet) SetSignature(Signature []byte) {
	PRUDPPacket.Signature = Signature
}

// SetSequenceID sets the packets SequenceID property
func (PRUDPPacket *Packet) SetSequenceID(SequenceID uint16) {
	PRUDPPacket.SequenceID = SequenceID
}

// SetPayload sets the packets Payload property
func (PRUDPPacket *Packet) SetPayload(Payload []byte) {
	PRUDPPacket.Payload = Payload
}

// SetMultiAckVersion sets the packets MultiAckVersion property
func (PRUDPPacket *Packet) SetMultiAckVersion(MultiAckVersion uint8) {
	PRUDPPacket.MultiAckVersion = MultiAckVersion
}

// SetType sets the packets Type property
func (PRUDPPacket *Packet) SetType(Type uint16) {
	PRUDPPacket.Type = Type
}

// AddFlag adds a packet flag to the packet
func (PRUDPPacket *Packet) AddFlag(Flag uint16) {
	PRUDPPacket.Flags |= Flag
}

// ClearFlag removes a packet flag from the packet
func (PRUDPPacket *Packet) ClearFlag(Flag uint16) {
	PRUDPPacket.Flags &^= Flag
}

// HasFlag checks if the packet has a flag
func (PRUDPPacket *Packet) HasFlag(Flag uint16) bool {
	return PRUDPPacket.Flags&Flag != 0
}

// NewPacket returns a new PRUDP packet generic
func NewPacket(client *Client) Packet {
	return Packet{
		Sender:          client,
		Data:            []byte{},
		Version:         0,
		Type:            0,
		Flags:           0,
		Source:          0,
		Destination:     0,
		SourceType:      0,
		SourcePort:      0,
		DestinationType: 0,
		DestinationPort: 0,
		SessionID:       0,
		Signature:       []byte{},
		SequenceID:      0,
		FragmentID:      0,
		Payload:         []byte{},
		MultiAckVersion: 0,
		Checksum:        0,
	}
}
