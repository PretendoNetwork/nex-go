package nex

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
)

// PacketV0 reresents a PRUDPv0 packet
type PacketV0 struct {
	Packet
	checksum uint32
}

// SetChecksum sets the packet checksum
func (packet *PacketV0) SetChecksum(checksum uint32) {
	packet.checksum = checksum
}

// GetChecksum returns the packet checksum
func (packet *PacketV0) GetChecksum() uint32 {
	return packet.checksum
}

// Decode decodes the packet
func (packet *PacketV0) Decode() error {

	if len(packet.Data()) < 9 {
		return errors.New("[PRUDPv0] Packet length less than header minimum")
	}

	var checksumSize int
	var payloadSize uint16
	var typeFlags uint16

	if packet.GetSender().GetServer().GetChecksumVersion() == 0 {
		checksumSize = 4
	} else {
		checksumSize = 1
	}

	stream := NewStreamIn(packet.Data(), packet.GetSender().GetServer())

	packet.SetSource(uint8(stream.ReadByteNext()))
	packet.SetDestination(uint8(stream.ReadByteNext()))

	typeFlags = stream.ReadU16LENext(1)[0]

	packet.SetSessionID(uint8(stream.ReadByteNext()))
	packet.SetSignature(stream.ReadBytesNext(4))
	packet.SetSequenceID(stream.ReadU16LENext(1)[0])

	if packet.GetSender().GetServer().GetFlagsVersion() == 0 {
		packet.SetType(typeFlags & 7)
		packet.SetFlags(typeFlags >> 3)
	} else {
		packet.SetType(typeFlags & 0xF)
		packet.SetFlags(typeFlags >> 4)
	}

	if _, ok := validTypes[packet.GetType()]; !ok {
		return errors.New("[PRUDPv0] Packet type not valid type")
	}

	if packet.GetType() == SynPacket || packet.GetType() == ConnectPacket {
		if len(packet.Data()[stream.ByteOffset():]) < 4 {
			return errors.New("[PRUDPv0] Packet specific data not large enough for connection signature")
		}

		packet.SetConnectionSignature(stream.ReadBytesNext(4))
	}

	if packet.GetType() == DataPacket {
		if len(packet.Data()[stream.ByteOffset():]) < 1 {
			return errors.New("[PRUDPv0] Packet specific data not large enough for fragment ID")
		}

		packet.SetFragmentID(uint8(stream.ReadByteNext()))
	}

	if packet.HasFlag(FlagHasSize) {
		if len(packet.Data()[stream.ByteOffset():]) < 2 {
			return errors.New("[PRUDPv0] Packet specific data not large enough for payload size")
		}

		payloadSize = stream.ReadU16LENext(1)[0]
	} else {
		payloadSize = uint16(len(packet.data) - int(stream.ByteOffset()) - checksumSize)
	}

	if payloadSize > 0 {
		if len(packet.Data()[stream.ByteOffset():]) < int(payloadSize) {
			return errors.New("[PRUDPv0] Packet data length less than payload length")
		}

		payloadCrypted := stream.ReadBytesNext(int64(payloadSize))

		packet.SetPayload(payloadCrypted)

		if packet.GetType() == DataPacket {
			ciphered := make([]byte, payloadSize)
			packet.GetSender().GetDecipher().XORKeyStream(ciphered, payloadCrypted)

			request, err := NewRMCRequest(ciphered)

			if err != nil {
				return errors.New("[PRUDPv0] Error parsing RMC request: " + err.Error())
			}

			packet.rmcRequest = request
		}
	}

	if len(packet.Data()[stream.ByteOffset():]) < int(checksumSize) {
		return errors.New("[PRUDPv0] Packet data length less than checksum length")
	}

	if checksumSize == 1 {
		packet.SetChecksum(uint32(stream.ReadByteNext()))
	} else {
		packet.SetChecksum(stream.ReadU32LENext(1)[0])
	}

	packetBody := stream.Bytes()

	calculatedChecksum := packet.calculateChecksum(packetBody[:len(packetBody)-checksumSize])

	if calculatedChecksum != packet.GetChecksum() {
		fmt.Println("[ERROR] Calculated checksum did not match")
	}

	return nil
}

// Bytes encodes the packet and returns a byte array
func (packet *PacketV0) Bytes() []byte {
	if packet.GetType() == DataPacket {

		if packet.HasFlag(FlagAck) {
			packet.SetPayload([]byte{})
		} else {
			payload := packet.GetPayload()

			if payload != nil || len(payload) > 0 {
				payloadSize := len(payload)

				encrypted := make([]byte, payloadSize)
				packet.GetSender().GetCipher().XORKeyStream(encrypted, payload)

				packet.SetPayload(encrypted)
			}
		}

		if !packet.HasFlag(FlagHasSize) {
			packet.AddFlag(FlagHasSize)
		}
	}

	var typeFlags uint16
	if packet.GetSender().GetServer().GetFlagsVersion() == 0 {
		typeFlags = packet.GetType() | packet.GetFlags()<<3
	} else {
		typeFlags = packet.GetType() | packet.GetFlags()<<4
	}

	stream := NewStreamOut(packet.GetSender().GetServer())

	packetSize := 11

	stream.Grow(int64(packetSize))
	stream.WriteByteNext(packet.GetSource())
	stream.WriteByteNext(packet.GetDestination())
	stream.WriteU16LENext([]uint16{typeFlags})
	stream.WriteByteNext(packet.GetSessionID())
	stream.WriteBytesNext(packet.calculateSignature())
	stream.WriteU16LENext([]uint16{packet.GetSequenceID()})

	options := packet.encodeOptions()
	optionsLength := len(options)

	if optionsLength > 0 {
		stream.Grow(int64(optionsLength))
		stream.WriteBytesNext(options)
	}

	payload := packet.GetPayload()

	if payload != nil && len(payload) > 0 {
		stream.Grow(int64(len(payload)))
		stream.WriteBytesNext(payload)
	}

	checksum := packet.calculateChecksum(stream.Bytes())

	if packet.GetSender().GetServer().GetChecksumVersion() == 0 {
		stream.Grow(4)
		stream.WriteU32LENext([]uint32{checksum})
	} else {
		stream.Grow(1)
		stream.WriteByteNext(byte(checksum))
	}

	return stream.Bytes()
}

func (packet *PacketV0) calculateSignature() []byte {
	// Friends server handles signatures differently, so check for the Friends server access key
	if packet.GetSender().GetServer().GetAccessKey() == "ridfebb9" {
		if packet.GetType() == DataPacket {
			payload := packet.GetPayload()

			if payload == nil || len(payload) <= 0 {
				signature := NewStreamIn(make([]byte, 4), packet.GetSender().GetServer())
				signature.WriteU32LENext([]uint32{0x12345678})

				return signature.Bytes()
			}

			key := packet.GetSender().GetSignatureKey()
			cipher := hmac.New(md5.New, key)
			cipher.Write(payload)

			return cipher.Sum(nil)[:4]
		}

		clientConnectionSignature := packet.GetSender().GetClientConnectionSignature()

		if clientConnectionSignature != nil {
			return clientConnectionSignature
		}

		return []byte{0x0, 0x0, 0x0, 0x0}
	}

	// Normal signature handling

	return []byte{}
}

func (packet *PacketV0) encodeOptions() []byte {
	stream := NewStreamOut(packet.GetSender().GetServer())

	if packet.GetType() == SynPacket {
		stream.Grow(4)
		stream.WriteBytesNext(packet.GetSender().GetServerConnectionSignature())
	}

	if packet.GetType() == ConnectPacket {
		stream.Grow(4)
		stream.WriteBytesNext(packet.GetSender().GetClientConnectionSignature())
	}

	if packet.GetType() == DataPacket {
		stream.Grow(1)
		stream.WriteByteNext(byte(packet.GetFragmentID()))
	}

	if packet.HasFlag(FlagHasSize) {
		stream.Grow(2)
		payload := packet.GetPayload()

		if payload != nil {
			stream.WriteU16LENext([]uint16{uint16(len(payload))})
		} else {
			stream.WriteU16LENext([]uint16{0})
		}
	}

	return stream.Bytes()
}

func (packet *PacketV0) calculateChecksum(data []byte) uint32 {
	signatureBase := packet.GetSender().GetSignatureBase()
	steps := len(data) / 4
	var temp uint32

	for i := 0; i < steps; i++ {
		offset := i * 4
		temp += binary.LittleEndian.Uint32(data[offset : offset+4])
	}

	temp &= 0xFFFFFFFF

	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, temp)

	checksum := signatureBase
	checksum += sum(data[len(data) & ^3:])
	checksum += sum(buff)

	return uint32(checksum & 0xFF)
}

// NewPacketV0 returns a new PRUDPv0 packet
func NewPacketV0(client *Client, data []byte) (*PacketV0, error) {
	packet := NewPacket(client, data)
	packetv0 := PacketV0{Packet: packet}

	if data != nil {
		err := packetv0.Decode()

		if err != nil {
			return &PacketV0{}, errors.New("[PRUDPv0] Error decoding packet data: " + err.Error())
		}
	}

	return &packetv0, nil
}
