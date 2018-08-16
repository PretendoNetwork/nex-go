package prudp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// PacketV0 represents a v0 PRUDP packet
type PacketV0 struct {
	Header   PacketV0Header
	MetaData []byte
	Payload  []byte
	Checksum uint8
	Type     uint16
	Flags    uint16
}

// PacketV0Header represents a v0 PRUDP packet header
type PacketV0Header struct {
	Source      uint8
	Destination uint8
	TypeFlags   uint16
	SessionID   uint8
	Signature   [4]byte
	SequenceID  uint16
}

// FromBytes converts a byte array to a PRUDP Packet struct
func (Packet *PacketV0) FromBytes(Data []byte) *PacketV0 {
	buffer := bytes.NewReader(Data)

	var PacketHeader PacketV0Header
	var MetaDataSize int

	if err := binary.Read(buffer, binary.LittleEndian, &PacketHeader); err != nil {
		fmt.Println(err)
	}

	Flags := PacketHeader.TypeFlags >> 4
	Type := PacketHeader.TypeFlags & 0xF

	switch Type {
	case 0:
		fallthrough
	case 1:
		MetaDataSize = 4
	case 2:
		MetaDataSize = 1
	}

	if Flags&0x008 != 0 {
		MetaDataSize = 2
	}

	MetaData := make([]byte, MetaDataSize)
	if _, err := buffer.Read(MetaData); err != nil {
		fmt.Println(err)
	}

	Packet.Header = PacketHeader
	Packet.MetaData = MetaData
	Packet.Payload = Data[11+MetaDataSize : len(Data)-1]
	Packet.Checksum = Data[len(Data)-1]
	Packet.Type = Type
	Packet.Flags = Flags

	return Packet
}

// Bytes converts a PRUDP packet to a byte array and returns it
func (Packet *PacketV0) Bytes() []byte {
	// 11 = packet header size (fixed length)
	// 1  = checksum byte
	length := 11 + len(Packet.MetaData) + len(Packet.Payload) + 1

	data := bytes.NewBuffer(make([]byte, 0, length))

	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.Source))
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.Destination))
	binary.Write(data, binary.LittleEndian, uint16(Packet.Header.TypeFlags))
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.SessionID))
	binary.Write(data, binary.LittleEndian, Packet.Header.Signature)
	binary.Write(data, binary.LittleEndian, uint16(Packet.Header.SequenceID))
	binary.Write(data, binary.LittleEndian, Packet.MetaData)
	binary.Write(data, binary.LittleEndian, Packet.Payload)

	return data.Bytes()
}

// NewV0Packet returns a new PRUDP v0 packet
func NewV0Packet() *PacketV0 {
	return &PacketV0{}
}

/*
func main() {
	data := []byte{0xAF, 0xA1, 0x62, 0x00, 0x95, 0xB7, 0x62, 0x5A, 0xAB, 0x02, 0x00, 0x00, 0x19, 0x87, 0x44, 0xDB, 0x99, 0xF8, 0x2C, 0x50, 0x05, 0xA3, 0x61, 0xFD, 0x2A, 0x1D, 0xF2, 0x80, 0xC8, 0x63, 0xD6, 0xC5, 0xAF, 0x61, 0x9B, 0x28, 0x6A, 0xEF}

	var Packet PacketV0
	_ = Packet.FromBytes(data)

	fmt.Println("TypeFlags", ToHEX(Packet.Header.TypeFlags))
	fmt.Println("Type", ToHEX(Packet.Type))
	fmt.Println("Flags", ToHEX(Packet.Flags))
	TypeFlags := (int(Packet.Flags) << 4) | int(Packet.Type)
	fmt.Println("TypeFlags (computed)", ToHEX(uint16(TypeFlags)))

	//Flags := PacketHeader.TypeFlags >> 4
	//Type := PacketHeader.TypeFlags & 0xF
}

func ToHEX(data uint16) string {
	return "0x" + hex.EncodeToString([]byte{byte(data)})
}
*/
