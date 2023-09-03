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
	checksum uint8
}

// SetChecksum sets the packet checksum
func (packet *PacketV0) SetChecksum(checksum uint8) {
	packet.checksum = checksum
}

// Checksum returns the packet checksum
func (packet *PacketV0) Checksum() uint8 {
	return packet.checksum
}

// Decode decodes the packet
func (packet *PacketV0) Decode() error {
	checksumSize := 1
	var payloadSize uint16
	var typeFlags uint16

	stream := NewStreamIn(packet.Data(), packet.Sender().Server())

	source, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 source. %s", err.Error())
	}

	packet.SetSource(source)

	destination, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 destination. %s", err.Error())
	}

	packet.SetDestination(destination)

	typeFlags, err = stream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 type-flags. %s", err.Error())
	}

	packet.SetType(typeFlags & 0xF)

	if _, ok := validTypes[packet.Type()]; !ok {
		return errors.New("Invalid PRUDP packet type")
	}

	packet.SetFlags(typeFlags >> 4)

	sessionID, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 session ID. %s", err.Error())
	}

	packet.SetSessionID(sessionID)

	if len(stream.Bytes()[stream.ByteOffset():]) < 4 {
		return errors.New("Failed to read PRUDPv0 packet signature. Not have enough data")
	}

	packet.SetSignature(stream.ReadBytesNext(4))

	sequenceID, err := stream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 sequence ID. %s", err.Error())
	}

	packet.SetSequenceID(sequenceID)

	if packet.Type() == SynPacket || packet.Type() == ConnectPacket {
		if len(packet.Data()[stream.ByteOffset():]) < 4 {
			return errors.New("Failed to read PRUDPv0 connection signature. Not enough data")
		}

		packet.SetConnectionSignature(stream.ReadBytesNext(4))
	}

	if packet.Type() == DataPacket {
		fragmentID, err := stream.ReadUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv0 connection signature. %s", err.Error())
		}

		packet.SetFragmentID(fragmentID)
	}

	if packet.HasFlag(FlagHasSize) {
		if len(packet.Data()[stream.ByteOffset():]) < 2 {
			return errors.New("[PRUDPv0] Packet specific data not large enough for payload size")
		}

		payloadSize, err = stream.ReadUInt16LE()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv0 payload size. %s", err.Error())
		}
	} else {
		payloadSize = uint16(len(packet.data) - int(stream.ByteOffset()) - checksumSize)
	}

	if payloadSize > 0 {
		if len(packet.Data()[stream.ByteOffset():]) < int(payloadSize) {
			return errors.New("[PRUDPv0] Packet data length less than payload length")
		}

		payloadCrypted := stream.ReadBytesNext(int64(payloadSize))

		packet.SetPayload(payloadCrypted)
	}

	if len(packet.Data()[stream.ByteOffset():]) < int(checksumSize) {
		return errors.New("[PRUDPv0] Packet data length less than checksum length")
	}

	checksum, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 packet checksum. %s", err.Error())
	}

	packet.SetChecksum(checksum)

	packetBody := stream.Bytes()

	calculatedChecksum := packet.calculateChecksum(packetBody[:len(packetBody)-checksumSize])

	if calculatedChecksum != packet.Checksum() {
		logger.Error("PRUDPv0 packet calculated checksum did not match")
	}

	return nil
}

// DecryptPayload decrypts the packets payload and sets the RMC request data
func (packet *PacketV0) DecryptPayload() error {
	if packet.Type() == DataPacket && !packet.HasFlag(FlagAck) {
		ciphered := make([]byte, len(packet.Payload()))

		packet.Sender().Decipher().XORKeyStream(ciphered, packet.Payload())

		request := NewRMCRequest()
		err := request.FromBytes(ciphered)
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv0 RMC request. %s", err.Error())
		}

		packet.rmcRequest = request
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

	var typeFlags uint16 = packet.Type() | packet.Flags()<<4

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

	if len(payload) > 0 {
		stream.Grow(int64(len(payload)))
		stream.WriteBytesNext(payload)
	}

	checksum := packet.calculateChecksum(stream.Bytes())

	stream.WriteUInt8(checksum)

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
	} else { // Normal signature handling
		if packet.Type() == DataPacket || packet.Type() == DisconnectPacket {
			payload := NewStreamOut(packet.Sender().Server())
			sessionKey := packet.Sender().SessionKey()
			if sessionKey != nil {
				payload.Grow(int64(len(sessionKey)))
				payload.WriteBytesNext(sessionKey)
			}
			payload.WriteUInt16LE(packet.sequenceID)
			payload.Grow(1)
			payload.WriteByteNext(packet.fragmentID)
			pktpay := packet.Payload()
			if len(pktpay) > 0 {
				payload.Grow(int64(len(pktpay)))
				payload.WriteBytesNext(pktpay)
			}

			key := packet.Sender().SignatureKey()
			cipher := hmac.New(md5.New, key)
			cipher.Write(payload.Bytes())

			return cipher.Sum(nil)[:4]
		} else {
			clientConnectionSignature := packet.Sender().ClientConnectionSignature()

			if clientConnectionSignature != nil {
				return clientConnectionSignature
			}
		}
	}

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
		stream.WriteBytesNext([]byte{0x00, 0x00, 0x00, 0x00})
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

func (packet *PacketV0) calculateChecksum(data []byte) uint8 {
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

	return uint8(checksum & 0xFF)
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
