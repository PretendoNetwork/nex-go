package nex

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	
	"github.com/superwhiskers/crunch"
)

func decodePacketV0(PRUDPPacket *Packet) map[string]interface{} {
	checksumVersion := PRUDPPacket.Sender.Server.Settings.PrudpV0ChecksumVersion
	checksumSize := 1
	if checksumVersion == 0 {
		checksumSize = 4
	}

	/*stream := NewInputStream(PRUDPPacket.Data)
	source := stream.UInt8()
	destination := stream.UInt8()
	typeFlags := stream.UInt16LE()
	sessionID := stream.UInt8()
	signature := stream.Bytes(4)
	sequenceID := stream.UInt16LE()*/
	
	stream := crunch.NewBuffer(PRUDPPacket.Data)
	
	source := stream.ReadByteNext()
	destination := stream.ReadByteNext()
	typeFlags := stream.ReadU16LENext(1)[0]
	sessionID := stream.ReadByteNext()
	signature := stream.ReadBytesNext(4)
	sequenceID := stream.ReadU16LENext(1)[0]
	
	if sequenceID == PRUDPPacket.Sender.SequenceIDIn.value {
		PRUDPPacket.Sender.SequenceIDIn.Increment()
	} else {
		return nil
	}

	sourceType := source >> 4
	sourcePort := source & 0xF

	destinationType := destination >> 4
	destinationPort := destination & 0xF

	var packetType uint16
	var flags uint16

	if PRUDPPacket.Sender.Server.Settings.PrudpV0FlagsVersion == 0 {
		packetType = typeFlags & 7
		flags = typeFlags >> 3
	} else {
		packetType = typeFlags & 0xF
		flags = typeFlags >> 4
	}

	decoded := map[string]interface{}{
		"Source":          source,
		"Destination":     destination,
		"SourceType":      sourceType,
		"SourcePort":      sourcePort,
		"DestinationType": destinationType,
		"DestinationPort": destinationPort,
		"SequenceID":      sequenceID,
		"SessionID":       sessionID,
		"Flags":           flags,
		"Type":            packetType,
		"FragmentID":      0,
		"Signature":       signature,
		"Payload":         []byte{},
	}

	if packetType == Types["Syn"] || packetType == Types["Connect"] {
		decoded["Signature"] = stream.ReadBytesNext(4)
	}

	if packetType == Types["Data"] {
		decoded["FragmentID"] = int(stream.ReadByteNext())
	}

	var payloadSize uint16
	if flags&Flags["HasSize"] != 0 {
		payloadSize = stream.ReadU16LENext(1)[0]
	} else {
		payloadSize = uint16(len(stream.Bytes()) - int(stream.ByteOffset()) - checksumSize)
	}

	payload := stream.ReadBytesNext(int64(payloadSize))
	decoded["Payload"] = payload

	if packetType == Types["Data"] && len(payload) > 0 {
		crypted := make([]byte, len(payload))
		PRUDPPacket.Sender.Decipher.XORKeyStream(crypted, payload)

		request := NewRMCRequest(crypted)

		decoded["RMCRequest"] = request
	} else {
		decoded["RMCRequest"] = nil
	}

	var checksum int
	if checksumSize == 1 {
		checksum = int(stream.ReadByteNext())
	} else {
		checksum = int(stream.ReadU32LENext(1)[0])
	}

	decoded["Checksum"] = checksum

	if CalculateV0Checksum(PRUDPPacket.Sender.SignatureBase, stream.ReadBytes(int64(len(stream.Bytes())-checksumSize), int64(checksumSize)), checksumVersion) != checksum {
		fmt.Println("[ERROR] Calculated checksum does not match decoded checksum!")
	}

	return decoded
}

func encodePacketV0(PRUDPPacket *Packet) []byte {
	checksumVersion := PRUDPPacket.Sender.Server.Settings.PrudpV0ChecksumVersion

	if PRUDPPacket.Type == Types["Data"] && len(PRUDPPacket.Payload) > 0 {
		crypted := make([]byte, len(PRUDPPacket.Payload))
		PRUDPPacket.Sender.Cipher.XORKeyStream(crypted, PRUDPPacket.Payload)
		PRUDPPacket.Payload = crypted
	}

	// stream := NewOutputStream()
	stream := crunch.NewBuffer()
	options := encodeOptionsV0(PRUDPPacket)

	stream.WriteByteNext(PRUDPPacket.Source)
	stream.WriteByteNext(PRUDPPacket.Destination)
	stream.WriteU16LENext([]uint16{PRUDPPacket.Type | PRUDPPacket.Flags << 4})
	stream.WriteByteNext(uint8(PRUDPPacket.Sender.SessionID))
	stream.WriteBytesNext(CalculateV0Signature(PRUDPPacket))
	stream.WriteU16LENext([]uint16{PRUDPPacket.SequenceID})

	stream.WriteBytesNext(options)
	if PRUDPPacket.Type == Types["Data"] && len(PRUDPPacket.Payload) > 0 {
		crypted := make([]byte, len(PRUDPPacket.Payload))
		PRUDPPacket.Sender.Cipher.XORKeyStream(crypted, PRUDPPacket.Payload)

		stream.WriteBytesNext(PRUDPPacket.Payload)
	}
	stream.WriteByteNext(uint8(CalculateV0Checksum(PRUDPPacket.Sender.SignatureBase, stream.Bytes(), checksumVersion)))

	return stream.Bytes()
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

func encodeOptionsV0(PRUDPPacket *Packet) []byte {
	stream := crunch.NewBuffer()

	if PRUDPPacket.Type == Types["Syn"] || PRUDPPacket.Type == Types["Connect"] {
		stream.WriteBytesNext(PRUDPPacket.Signature)
	}

	if PRUDPPacket.Type == Types["Data"] {
		stream.WriteByteNext(uint8(PRUDPPacket.FragmentID))
	}

	if PRUDPPacket.HasFlag(Flags["HasSize"]) {
		stream.WriteU16LENext([]uint16{uint16(len(PRUDPPacket.Payload))})
	}

	return stream.Bytes()
}

// CalculateV0Signature calculates the signature of a prudpv0 packet
func CalculateV0Signature(PRUDPPacket *Packet) []byte {
	if PRUDPPacket.Type == Types["Data"] || (PRUDPPacket.Type == Types["Disconnect"] && PRUDPPacket.Sender.Server.Settings.PrudpV0SignatureVersion == 0) {
		data := PRUDPPacket.Payload
		if PRUDPPacket.Sender.Server.Settings.PrudpV0SignatureVersion == 0 {
			buffer := crunch.NewBuffer()
			
			buffer.WriteBytesNext(PRUDPPacket.Sender.SecureKey)
			buffer.WriteU16LENext([]uint16{PRUDPPacket.SequenceID})
			buffer.WriteByteNext(uint8(PRUDPPacket.FragmentID))
			buffer.WriteBytesNext(data)

			data = buffer.Bytes()
		}

		if len(data) > 0 {
			key, _ := hex.DecodeString(PRUDPPacket.Sender.SignatureKey)
			cipher := hmac.New(md5.New, key)
			cipher.Write(data)
			return cipher.Sum(nil)[:4]
		}

		buffer := crunch.NewBuffer()
		buffer.WriteU32LENext([]uint32{0x12345678})

		return buffer.Bytes()
	}

	if PRUDPPacket.Signature != nil {
		return PRUDPPacket.Signature
	}

	// todo: client/server?
	if PRUDPPacket.Sender.ClientConnectionSignature != nil {
		return PRUDPPacket.Sender.ClientConnectionSignature
	} else {
		return make([]byte, 4)
	}
}
