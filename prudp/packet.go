package prudp

import (
	"bytes"
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

/*
func main() {
	packet_hex := "afa16200966ae88231030000704a3f4fd43b79f560cd38ac27e957788338d16af252ca44e42d21956dcc6c08b433aa3fc486ac07b8eacecf12fca15a1200509c0e6bed36a6348276bd0a0c4275b60f624df8354ef6a90060d448f096db1798e4a1b8fad05e8c6d5e09237ac4e13f44bb26172aff00ee9239389f0e7a6b7c2dba46230f968507cb31c4291a4510f1474680d6b9c1dd3382684ffdf3add55a607729bcb6f9ebd8c705ff90a9f799c2c8a559701fabb59825f5481dbe14a5998e50dcd99747b39245e92d308a05185bbc47865abc716e5c7c7d62002190b27dd05514210dc47b32ab6b3b6c9ae1dac5751519"
	data, _ := hex.DecodeString(packet_hex)

	var PRUDPPacket Packet
	PRUDPPacket.FromBytes(data)

	fmt.Println("PRUDPPacket.Version:------->", PRUDPPacket.Version)
	fmt.Println("PRUDPPacket.Magic:--------->", PRUDPPacket.Magic)
	fmt.Println("PRUDPPacket.Source:-------->", PRUDPPacket.Source)
	fmt.Println("PRUDPPacket.Destination:--->", PRUDPPacket.Destination)
	fmt.Println("PRUDPPacket.TypeFlags:----->", PRUDPPacket.TypeFlags)
	fmt.Println("PRUDPPacket.SessionID:----->", PRUDPPacket.SessionID)
	fmt.Println("PRUDPPacket.Signature:----->", PRUDPPacket.Signature)
	fmt.Println("PRUDPPacket.SequenceID:---->", PRUDPPacket.SequenceID)
	fmt.Println("PRUDPPacket.MetaData:------>", PRUDPPacket.MetaData)
	fmt.Println("PRUDPPacket.Payload:------->", PRUDPPacket.Payload)
	fmt.Println("PRUDPPacket.MultiAckVersion:", PRUDPPacket.MultiAckVersion)
	fmt.Println("PRUDPPacket.Type:---------->", PRUDPPacket.Type)
	fmt.Println("PRUDPPacket.Flags:--------->", PRUDPPacket.Flags)
}
*/
