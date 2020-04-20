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

// Checksum returns the packet checksum
func (packet *PacketV0) Checksum() uint32 {
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

	if packet.Sender().Server().ChecksumVersion() == 0 {
		checksumSize = 4
	} else {
		checksumSize = 1
	}

	stream := NewStreamIn(packet.Data(), packet.Sender().Server())

	packet.SetSource(stream.ReadUInt8())
	packet.SetDestination(stream.ReadUInt8())

	typeFlags = stream.ReadUInt16LE()

	packet.SetSessionID(stream.ReadUInt8())
	packet.SetSignature(stream.ReadBytesNext(4))
	packet.SetSequenceID(stream.ReadUInt16LE())

	if packet.Sender().Server().FlagsVersion() == 0 {
		packet.SetType(typeFlags & 7)
		packet.SetFlags(typeFlags >> 3)
	} else {
		packet.SetType(typeFlags & 0xF)
		packet.SetFlags(typeFlags >> 4)
	}

	if _, ok := validTypes[packet.Type()]; !ok {
		return errors.New("[PRUDPv0] Packet type not valid type")
	}

	if packet.Type() == SynPacket || packet.Type() == ConnectPacket {
		if len(packet.Data()[stream.ByteOffset():]) < 4 {
			return errors.New("[PRUDPv0] Packet specific data not large enough for connection signature")
		}

		packet.SetConnectionSignature(stream.ReadBytesNext(4))
	}

	if packet.Type() == DataPacket {
		if len(packet.Data()[stream.ByteOffset():]) < 1 {
			return errors.New("[PRUDPv0] Packet specific data not large enough for fragment ID")
		}

		packet.SetFragmentID(stream.ReadUInt8())
	}

	if packet.HasFlag(FlagHasSize) {
		if len(packet.Data()[stream.ByteOffset():]) < 2 {
			return errors.New("[PRUDPv0] Packet specific data not large enough for payload size")
		}

		payloadSize = stream.ReadUInt16LE()
	} else {
		payloadSize = uint16(len(packet.data) - int(stream.ByteOffset()) - checksumSize)
	}

	if payloadSize > 0 {
		if len(packet.Data()[stream.ByteOffset():]) < int(payloadSize) {
			return errors.New("[PRUDPv0] Packet data length less than payload length")
		}

		payloadCrypted := stream.ReadBytesNext(int64(payloadSize))

		packet.SetPayload(payloadCrypted)

		if packet.Type() == DataPacket {
			ciphered := make([]byte, payloadSize)
			packet.Sender().Decipher().XORKeyStream(ciphered, payloadCrypted)

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
		packet.SetChecksum(uint32(stream.ReadUInt8()))
	} else {
		packet.SetChecksum(stream.ReadUInt32LE())
	}

	packetBody := stream.Bytes()

	calculatedChecksum := packet.calculateChecksum(packetBody[:len(packetBody)-checksumSize])

	if calculatedChecksum != packet.Checksum() {
		fmt.Println("[ERROR] Calculated checksum did not match")
	}

	return nil
}

// Bytes encodes the packet and returns a byte array
func (packet *PacketV0) Bytes() []byte {
	if packet.Type() == DataPacket {

		if packet.HasFlag(FlagAck) {
			packet.SetPayload([]byte{})
		} else {
			payload := packet.Payload()

			if payload != nil || len(payload) > 0 {
				payloadSize := len(payload)

				encrypted := make([]byte, payloadSize)
				packet.Sender().Cipher().XORKeyStream(encrypted, payload)

				packet.SetPayload(encrypted)
			}
		}

		if !packet.HasFlag(FlagHasSize) {
			packet.AddFlag(FlagHasSize)
		}
	}

	var typeFlags uint16
	if packet.Sender().Server().FlagsVersion() == 0 {
		typeFlags = packet.Type() | packet.Flags()<<3
	} else {
		typeFlags = packet.Type() | packet.Flags()<<4
	}

	stream := NewStreamOut(packet.Sender().Server())
	packetSignature := packet.calculateSignature()

	stream.WriteUInt8(packet.Source())
	stream.WriteUInt8(packet.Destination())
	stream.WriteUInt16LE(typeFlags)
	stream.WriteUInt8(packet.SessionID())
	stream.Grow(int64(len(packetSignature)))
	stream.WriteBytesNext(packetSignature)
	stream.WriteUInt16LE(packet.SequenceID())

	options := packet.encodeOptions()
	optionsLength := len(options)

	if optionsLength > 0 {
		stream.Grow(int64(optionsLength))
		stream.WriteBytesNext(options)
	}

	payload := packet.Payload()

	if payload != nil && len(payload) > 0 {
		stream.Grow(int64(len(payload)))
		stream.WriteBytesNext(payload)
	}

	checksum := packet.calculateChecksum(stream.Bytes())

	if packet.Sender().Server().ChecksumVersion() == 0 {
		stream.WriteUInt32LE(checksum)
	} else {
		stream.WriteUInt8(uint8(checksum))
	}

	return stream.Bytes()
}

func (packet *PacketV0) calculateSignature() []byte {
	// Friends server handles signatures differently, so check for the Friends server access key
	if packet.Sender().Server().AccessKey() == "ridfebb9" {
		if packet.Type() == DataPacket {
			payload := packet.Payload()

			if payload == nil || len(payload) <= 0 {
				signature := NewStreamOut(packet.Sender().Server())
				signature.WriteUInt32LE(0x12345678)

				return signature.Bytes()
			}

			key := packet.Sender().SignatureKey()
			cipher := hmac.New(md5.New, key)
			cipher.Write(payload)

			return cipher.Sum(nil)[:4]
		}

		clientConnectionSignature := packet.Sender().ClientConnectionSignature()

		if clientConnectionSignature != nil {
			return clientConnectionSignature
		}

		return []byte{0x0, 0x0, 0x0, 0x0}
	}

	// Normal signature handling

	return []byte{}
}

func (packet *PacketV0) encodeOptions() []byte {
	stream := NewStreamOut(packet.Sender().Server())

	if packet.Type() == SynPacket {
		stream.Grow(4)
		stream.WriteBytesNext(packet.Sender().ServerConnectionSignature())
	}

	if packet.Type() == ConnectPacket {
		stream.Grow(4)
		stream.WriteBytesNext(packet.Sender().ClientConnectionSignature())
	}

	if packet.Type() == DataPacket {
		stream.WriteUInt8(packet.FragmentID())
	}

	if packet.HasFlag(FlagHasSize) {
		payload := packet.Payload()

		if payload != nil {
			stream.WriteUInt16LE(uint16(len(payload)))
		} else {
			stream.WriteUInt16LE(0)
		}
	}

	return stream.Bytes()
}

func (packet *PacketV0) calculateChecksum(data []byte) uint32 {
	signatureBase := packet.Sender().SignatureBase()
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
