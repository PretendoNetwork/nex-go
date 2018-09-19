package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// PacketV0Header represents a v0 PRUDP packet header
type PacketV0Header struct {
	Source      uint8
	Destination uint8
	TypeFlags   uint16
	SessionID   uint8
	Signature   [4]byte
	SequenceID  uint16
}

func decodeV0(PRUDPPacket *Packet) map[string]interface{} {
	buffer := bytes.NewReader(PRUDPPacket.Data)
	rawPacket := PRUDPPacket.Data

	var PacketHeader PacketV0Header
	checksumVersion := PRUDPPacket.Sender.Server.Settings.PrudpV0ChecksumVersion
	checksumSize := 1
	if checksumVersion == 0 {
		checksumSize = 4
	}

	var payloadSize uint16

	if err := binary.Read(buffer, binary.LittleEndian, &PacketHeader); err != nil {
		fmt.Println("err")
	}

	PacketSource := PacketHeader.Source
	PacketDestination := PacketHeader.Destination

	PacketSourceType := PacketSource >> 4
	PacketSourcePort := PacketSource & 0xF
	PacketDestinationType := PacketDestination >> 4
	PacketDestinationPort := PacketDestination & 0xF

	var PacketFlags uint16
	var PacketType uint16

	if PRUDPPacket.Sender.Server.Settings.PrudpV0FlagsVersion == 0 {
		PacketFlags = PacketHeader.TypeFlags >> 3
		PacketType = PacketHeader.TypeFlags & 7
	} else {
		PacketFlags = PacketHeader.TypeFlags >> 4
		PacketType = PacketHeader.TypeFlags & 0xF
	}

	offset := 11

	PacketSignature := []byte{}
	if PacketType == Types["Syn"] || PacketType == Types["Connect"] {
		PacketSignature = rawPacket[offset : offset+4]
		offset += 4
	}

	PacketFragmentID := 0
	if PacketType == Types["Data"] {
		PacketFragmentID = int(rawPacket[offset])
		offset++
	}

	if PacketFlags&Flags["HasSize"] != 0 {
		payloadSizeTemp := make([]byte, 2)
		buffer.ReadAt(payloadSizeTemp, int64(offset))
		payloadSize = readUInt16(payloadSizeTemp, binary.LittleEndian)
		offset += 2
	} else {
		payloadSize = uint16(len(rawPacket) - offset - checksumSize)
	}

	checksumTemp := make([]byte, checksumSize)
	buffer.ReadAt(checksumTemp, int64(uint16(offset)+payloadSize))
	var checksum int

	if checksumSize == 1 {
		checksum = int(checksumTemp[0])
	} else {
		checksum = int(readUInt32(checksumTemp, binary.LittleEndian))
	}

	if CalculateV0Checksum(PRUDPPacket.Sender.SignatureBase, rawPacket[:uint16(offset)+payloadSize], checksumVersion) != checksum {
		fmt.Println("[ERROR] Calculated checksum does not match decoded checksum!")
	}

	payload := rawPacket[uint16(offset) : uint16(offset)+payloadSize]

	decoded := map[string]interface{}{
		"Source":          PacketSource,
		"Destination":     PacketDestination,
		"SourceType":      PacketSourceType,
		"SourcePort":      PacketSourcePort,
		"DestinationType": PacketDestinationType,
		"DestinationPort": PacketDestinationPort,
		"SequenceID":      PacketHeader.SequenceID,
		"SessionID":       PacketHeader.SessionID,
		"Flags":           PacketFlags,
		"Type":            PacketType,
		"Signature":       PacketSignature,
		"FragmentID":      PacketFragmentID,
		"Payload":         payload,
		"Checksum":        checksum,
	}

	var Payload []byte
	Payload = decoded["Payload"].([]byte)

	if PacketType == Types["Data"] && len(Payload) > 0 {
		crypted := make([]byte, len(Payload))
		PRUDPPacket.Sender.Decipher.XORKeyStream(crypted, Payload)

		request := NewRMCRequest(crypted)

		decoded["RMCRequest"] = request
	}

	return decoded
}

func encodeV0(PRUDPPacket *Packet) []byte {

	checksumVersion := PRUDPPacket.Sender.Server.Settings.PrudpV0ChecksumVersion
	checksumSize := 1
	if checksumVersion == 0 {
		checksumSize = 4
	}

	if PRUDPPacket.Type == Types["Data"] && len(PRUDPPacket.Payload) > 0 {
		crypted := make([]byte, len(PRUDPPacket.Payload))
		PRUDPPacket.Sender.Cipher.XORKeyStream(crypted, PRUDPPacket.Payload)
		PRUDPPacket.Payload = crypted
	}

	// 11 = packet header size (fixed length)
	length := 11 + checksumSize

	buffer := bytes.NewBuffer(make([]byte, 0, length))
	options := encodeOptions(PRUDPPacket)

	binary.Write(buffer, binary.LittleEndian, uint8(PRUDPPacket.Source))
	binary.Write(buffer, binary.LittleEndian, uint8(PRUDPPacket.Destination))
	binary.Write(buffer, binary.LittleEndian, uint16(int(PRUDPPacket.Type)|int(PRUDPPacket.Flags)<<4))
	binary.Write(buffer, binary.LittleEndian, uint8(PRUDPPacket.Sender.SessionID))
	binary.Write(buffer, binary.LittleEndian, CalculateV0Signature(PRUDPPacket))
	binary.Write(buffer, binary.LittleEndian, uint16(PRUDPPacket.SequenceID))

	binary.Write(buffer, binary.LittleEndian, options)
	binary.Write(buffer, binary.LittleEndian, PRUDPPacket.Payload)
	binary.Write(buffer, binary.LittleEndian, uint8(CalculateV0Checksum(PRUDPPacket.Sender.SignatureBase, buffer.Bytes(), checksumVersion)))

	return buffer.Bytes()
}

// CalculateV0Checksum calculates the checksum of a prudpv0 packet
func CalculateV0Checksum(checksum int, packet []byte, version int) int {
	// in the future we need to check the `version` here and change the alg accordingly
	pos := 0

	sections := len(packet) / 4
	chunks := make([]uint32, 0, sections)

	for i := 0; i < sections; i++ {
		chunk := binary.LittleEndian.Uint32(packet[pos : pos+4])
		chunks = append(chunks, chunk)

		pos += 4
	}

	temp1 := sum(chunks)
	temp := temp1 & 0xFFFFFFFF

	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(temp))

	tempSum := sum(packet[len(packet) & ^3:])

	checksum += tempSum
	tempSum = sum(buff)
	checksum += tempSum

	return (checksum & 0xFF)
}

func encodeOptions(PRUDPPacket *Packet) []byte {
	length := 0

	if PRUDPPacket.Type == Types["Syn"] || PRUDPPacket.Type == Types["Connect"] {
		length += len(PRUDPPacket.Signature)
	}

	if PRUDPPacket.Type == Types["Data"] {
		length++
	}

	if PRUDPPacket.HasFlag(Flags["HasSize"]) {
		length += 2
	}

	buffer := bytes.NewBuffer(make([]byte, 0, length))

	if PRUDPPacket.Type == Types["Syn"] || PRUDPPacket.Type == Types["Connect"] {
		binary.Write(buffer, binary.LittleEndian, PRUDPPacket.Signature)
	}

	if PRUDPPacket.Type == Types["Data"] {
		binary.Write(buffer, binary.LittleEndian, uint8(PRUDPPacket.FragmentID))
	}

	if PRUDPPacket.HasFlag(Flags["HasSize"]) {
		binary.Write(buffer, binary.LittleEndian, uint16(len(PRUDPPacket.Payload)))
	}

	return buffer.Bytes()
}

// CalculateV0Signature calculates the signature of a prudpv0 packet
func CalculateV0Signature(PRUDPPacket *Packet) []byte {
	if PRUDPPacket.Type == Types["Data"] || (PRUDPPacket.Type == Types["Disconnect"] && PRUDPPacket.Sender.Server.Settings.PrudpV0SignatureVersion == 0) {
		data := PRUDPPacket.Payload
		if PRUDPPacket.Sender.Server.Settings.PrudpV0SignatureVersion == 0 {
			length := len(PRUDPPacket.Sender.SecureKey) + 2 + 1 + len(data)
			buffer := bytes.NewBuffer(make([]byte, 0, length))

			binary.Write(buffer, binary.LittleEndian, PRUDPPacket.Sender.SecureKey)
			binary.Write(buffer, binary.LittleEndian, uint16(PRUDPPacket.SequenceID))
			binary.Write(buffer, binary.LittleEndian, uint8(PRUDPPacket.FragmentID))
			binary.Write(buffer, binary.LittleEndian, data)

			data = buffer.Bytes()
		}

		if len(data) > 0 {
			key, _ := hex.DecodeString(PRUDPPacket.Sender.SignatureKey)
			cipher := hmac.New(md5.New, key)
			cipher.Write(data)
			return cipher.Sum(nil)[:4]
		}

		buffer := bytes.NewBuffer(make([]byte, 0, 4))
		binary.Write(buffer, binary.LittleEndian, uint32(0x12345678))

		return buffer.Bytes()
	}

	if PRUDPPacket.Signature != nil {
		return PRUDPPacket.Signature
	} else {
		return make([]byte, 4)
	}
}
