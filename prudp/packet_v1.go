package prudp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// PacketV1 represents a v1 PRUDP packet
type PacketV1 struct {
	Header    PacketV1Header
	Signature []byte
	MetaData  []byte
	Payload   []byte
	Type      uint16
	Flags     uint16
}

// PacketV1Header represents a v1 PRUDP packet header
type PacketV1Header struct {
	Version         uint8
	MetaDataSize    uint8
	PayloadSize     uint16
	Source          uint8
	Destination     uint8
	TypeFlags       uint16
	SessionID       uint8
	MultiAckVersion uint8
	SequenceID      uint16
}

// FromBytes converts a byte array to a PRUDP Packet struct
func (Packet *PacketV1) FromBytes(Data []byte) *PacketV1 {
	buffer := bytes.NewReader(Data[2:]) // Ignore the MAGIC. This is set in packet.go

	var PacketHeader PacketV1Header

	if err := binary.Read(buffer, binary.LittleEndian, &PacketHeader); err != nil {
		fmt.Println(err)
	}

	Flags := PacketHeader.TypeFlags >> 4
	Type := PacketHeader.TypeFlags & 0xF

	Signature := make([]byte, 16)
	if _, err := buffer.Read(Signature); err != nil {
		fmt.Println(err)
	}

	MetaData := make([]byte, PacketHeader.MetaDataSize)
	if _, err := buffer.Read(MetaData); err != nil {
		fmt.Println(err)
	}

	Packet.Header = PacketHeader
	Packet.Signature = Signature
	Packet.MetaData = MetaData
	Packet.Payload = Data[30+PacketHeader.MetaDataSize:]
	Packet.Type = Type
	Packet.Flags = Flags

	return Packet
}

// Bytes converts a PRUDP packet to a byte array and returns it
func (Packet *PacketV1) Bytes() []byte {
	// 2 = packet MAGIC (fixed length)
	// 12  = packet header (fixed length)
	// 16  = packet signature (fixed length)
	length := 2 + 12 + 16 + int(Packet.Header.MetaDataSize) + int(Packet.Header.PayloadSize)

	data := bytes.NewBuffer(make([]byte, 0, length))

	binary.Write(data, binary.LittleEndian, [2]byte{0xEA, 0xD0}) // MAGIC
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.Version))
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.MetaDataSize))
	binary.Write(data, binary.LittleEndian, uint16(Packet.Header.PayloadSize))
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.Source))
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.Destination))
	binary.Write(data, binary.LittleEndian, uint16(Packet.Header.TypeFlags))
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.SessionID))
	binary.Write(data, binary.LittleEndian, uint8(Packet.Header.MultiAckVersion))
	binary.Write(data, binary.LittleEndian, uint16(Packet.Header.SequenceID))
	binary.Write(data, binary.LittleEndian, Packet.Signature)
	binary.Write(data, binary.LittleEndian, Packet.MetaData)
	binary.Write(data, binary.LittleEndian, Packet.Payload)

	return data.Bytes()
}

// NewV1Packet returns a new PRUDP v1 packet
func NewV1Packet() *PacketV1 {
	return &PacketV1{}
}

/*
func main() {
	packet_hex := "ead001032501afa1e200e0003600368ab97371f1b20d9eaef91a2dee89c802010080fdd5e239fbb18501721aedd4d097e285f9a435d0e79d565ab65e95224bf40441097976db3dcf9fc835bca9ab26c1706f63a50b136124a131c329632bbcb85c31c9dc50695a3f3dec213b1aba9a26e40f5963688861a40220bbb461feb7713186e6b25569852fcfa6bdd723a22178a99d66947cc0d0b5e3a78c6ed410132f2f504f5c4805a2a0c53669a8a32e80c885b9584cf9d74aef67dcde00631c3db4f581ee3936abc345a42b60a2b36d7a872eb3fd000d9a3c91b7cda9e78cd23801570f588af8136ea10db4bb9cef67b6653c04d8495ab3dd4fac0a3c54474e2172b2d6d8ed8b343d04f7ca9a1ac9fe78a0db397f2bfbd1882834eb467246d17c60a162914a54b98f23b56c5e7d9de39199d7df2d74bf09137ec2ba68442de0d6d9c2fb04fccbcd"
	data, _ := hex.DecodeString(packet_hex)

	var Packet PacketV1
	_ = Packet.FromBytes(data)

	fmt.Println("Initial data bytes", data, "\n")
	fmt.Println("Struct ---------->", Packet, "\n")
	fmt.Println("Packet.Bytes() -->", Packet.Bytes())

	fmt.Println()
}
*/
