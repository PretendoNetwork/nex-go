package prudp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Packet represents a generic PRUDP packet if any type
type Packet struct {
	Version         int
	Magic           []byte
	Source          uint8
	Destination     uint8
	TypeFlags       uint16
	SessionID       uint8
	Signature       []byte
	SequenceID      uint16
	MetaData        []byte
	Payload         []byte
	MultiAckVersion uint8 // only present in v1
	Type            uint16
	Flags           uint16
}

// Go doesn't allow byte arrays, even with fixed sizes, to be type `const` :L
// Casting to `var` instead
var (
	PRUDPV0ClientMagic = []byte{0xAF, 0xA1}
	PRUDPV0ServerMagic = []byte{0xA1, 0xAF}
	PRUDPV1Magic       = []byte{0xEA, 0xD0}
)

// FromBytes converts a byte array to a generic PRUDP packet
func (PRUDPPacket *Packet) FromBytes(Data []byte) {
	magic := Data[:2]

	PRUDPPacket.Magic = magic

	if bytes.Equal(magic, PRUDPV0ClientMagic) || bytes.Equal(magic, PRUDPV0ServerMagic) {
		var V0Packet PacketV0
		V0Packet.FromBytes(Data)

		PRUDPPacket.Version = 0

		PRUDPPacket.Source = V0Packet.Header.Source
		PRUDPPacket.Destination = V0Packet.Header.Destination
		PRUDPPacket.TypeFlags = V0Packet.Header.TypeFlags
		PRUDPPacket.SessionID = V0Packet.Header.SessionID
		PRUDPPacket.Signature = V0Packet.Header.Signature[:]
		PRUDPPacket.SequenceID = V0Packet.Header.SequenceID
		PRUDPPacket.MetaData = V0Packet.MetaData
		PRUDPPacket.Payload = V0Packet.Payload
		PRUDPPacket.Type = V0Packet.Type
		PRUDPPacket.Flags = V0Packet.Flags
	} else if bytes.Equal(magic, PRUDPV1Magic) {
		var V1Packet PacketV1
		V1Packet.FromBytes(Data)

		PRUDPPacket.Version = 1

		PRUDPPacket.Source = V1Packet.Header.Source
		PRUDPPacket.Destination = V1Packet.Header.Destination
		PRUDPPacket.TypeFlags = V1Packet.Header.TypeFlags
		PRUDPPacket.SessionID = V1Packet.Header.SessionID
		PRUDPPacket.MultiAckVersion = V1Packet.Header.MultiAckVersion
		PRUDPPacket.SequenceID = V1Packet.Header.SequenceID
		PRUDPPacket.Signature = V1Packet.Signature[:]
		PRUDPPacket.MetaData = V1Packet.MetaData
		PRUDPPacket.Payload = V1Packet.Payload
		PRUDPPacket.Type = V1Packet.Type
		PRUDPPacket.Flags = V1Packet.Flags
	} else {
		fmt.Println("[ERROR] UNKNOWN PRUDP PACKET TYPE:", magic)
	}
}

// Bytes converts a PRUDP packet to a byte array
func (PRUDPPacket *Packet) Bytes() []byte {
	PacketSize := 0

	switch PRUDPPacket.Version {
	case 0:
		PacketSize += 12
	case 1:
		PacketSize += 30
	}

	PacketSize += len(PRUDPPacket.MetaData)
	PacketSize += len(PRUDPPacket.Payload)

	data := bytes.NewBuffer(make([]byte, 0, PacketSize))

	switch PRUDPPacket.Version {
	case 0:
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.Source))
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.Destination))
		binary.Write(data, binary.LittleEndian, uint16((int(PRUDPPacket.Flags)<<4)|int(PRUDPPacket.Type)))
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.SessionID))
		binary.Write(data, binary.LittleEndian, PRUDPPacket.Signature)
		binary.Write(data, binary.LittleEndian, uint16(PRUDPPacket.SequenceID))
		binary.Write(data, binary.LittleEndian, PRUDPPacket.MetaData)
		binary.Write(data, binary.LittleEndian, PRUDPPacket.Payload)
	case 1:
		binary.Write(data, binary.LittleEndian, [2]byte{0xEA, 0xD0})
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.Version))
		binary.Write(data, binary.LittleEndian, uint8(len(PRUDPPacket.MetaData)))
		binary.Write(data, binary.LittleEndian, uint16(len(PRUDPPacket.Payload)))
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.Source))
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.Destination))
		binary.Write(data, binary.LittleEndian, uint16((int(PRUDPPacket.Flags)<<4)|int(PRUDPPacket.Type)))
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.SessionID))
		binary.Write(data, binary.LittleEndian, uint8(PRUDPPacket.MultiAckVersion))
		binary.Write(data, binary.LittleEndian, uint16(PRUDPPacket.SequenceID))
		binary.Write(data, binary.LittleEndian, PRUDPPacket.Signature)
		binary.Write(data, binary.LittleEndian, PRUDPPacket.MetaData)
		binary.Write(data, binary.LittleEndian, PRUDPPacket.Payload)
	}

	return data.Bytes()
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

func (PRUDPPacket *Packet) SetSignature(Signature []byte) {
	PRUDPPacket.Signature = Signature
}

// SetSequenceID sets the packets SequenceID property
func (PRUDPPacket *Packet) SetSequenceID(SequenceID uint16) {
	PRUDPPacket.SequenceID = SequenceID
}

// SetMetaData sets the packets MetaData property
func (PRUDPPacket *Packet) SetMetaData(MetaData []byte) {
	PRUDPPacket.MetaData = MetaData
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
func (PRUDPPacket *Packet) SetType(Type int) {
	PRUDPPacket.Type = uint16(Type)
}

// AddFlag adds a packet flag to the packet
func (PRUDPPacket *Packet) AddFlag(Flag int) {
	PRUDPPacket.Flags |= uint16(Flag)
}

// ClearFlag removes a packet flag from the packet
func (PRUDPPacket *Packet) ClearFlag(Flag int) {
	PRUDPPacket.Flags &^= uint16(Flag)
}

// NewPacket returns a new PRUDP packet generic
func NewPacket() *Packet {
	return &Packet{
		Flags: 0,
	}
}

/*
func main() {
	packet_hex := "ead001032501afa1e200e0003600368ab97371f1b20d9eaef91a2dee89c802010080fdd5e239fbb18501721aedd4d097e285f9a435d0e79d565ab65e95224bf40441097976db3dcf9fc835bca9ab26c1706f63a50b136124a131c329632bbcb85c31c9dc50695a3f3dec213b1aba9a26e40f5963688861a40220bbb461feb7713186e6b25569852fcfa6bdd723a22178a99d66947cc0d0b5e3a78c6ed410132f2f504f5c4805a2a0c53669a8a32e80c885b9584cf9d74aef67dcde00631c3db4f581ee3936abc345a42b60a2b36d7a872eb3fd000d9a3c91b7cda9e78cd23801570f588af8136ea10db4bb9cef67b6653c04d8495ab3dd4fac0a3c54474e2172b2d6d8ed8b343d04f7ca9a1ac9fe78a0db397f2bfbd1882834eb467246d17c60a162914a54b98f23b56c5e7d9de39199d7df2d74bf09137ec2ba68442de0d6d9c2fb04fccbcd"
	data, _ := hex.DecodeString(packet_hex)

	var Packet1 Packet
	Packet1.FromBytes(data)

	Packet2 := NewPacket()

	Packet2.SetVersion(Packet1.Version)
	Packet2.SetSource(Packet1.Source)
	Packet2.SetDestination(Packet1.Destination)
	Packet2.SetSessionID(Packet1.SessionID)
	Packet2.SetSignature(Packet1.Signature)
	Packet2.SetSequenceID(Packet1.SequenceID)
	Packet2.SetMultiAckVersion(Packet1.MultiAckVersion)
	Packet2.SetType(int(Packet1.Type))

	fmt.Println("Input data:\n", data)
	fmt.Println("\n")
	fmt.Println("Packet.Version:--------->", Packet1.Version)
	fmt.Println("Packet.Magic:----------->", Packet1.Magic)
	fmt.Println("Packet.Source:---------->", Packet1.Source)
	fmt.Println("Packet.Destination:----->", Packet1.Destination)
	fmt.Println("Packet.TypeFlags:------->", Packet1.TypeFlags)
	fmt.Println("Packet.SessionID:------->", Packet1.SessionID)
	fmt.Println("Packet.Signature:------->", Packet1.Signature)
	fmt.Println("Packet.SequenceID:------>", Packet1.SequenceID)
	fmt.Println("Packet.MultiAckVersion:->", Packet1.MultiAckVersion)
	fmt.Println("Packet.Type:------------>", Packet1.Type)
	fmt.Println("Packet.Flags:----------->", Packet1.Flags)
	fmt.Println("\n")
	fmt.Println("Packet.Bytes():\n", Packet1.Bytes())
}
*/
