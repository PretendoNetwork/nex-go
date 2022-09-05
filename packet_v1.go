package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
)

// OptionAllFunctions is used with OptionSupportedFunctions to support all methods
var OptionAllFunctions = 0xFFFFFFFF

// OptionSupportedFunctions is the ID for the Supported Functions option in PRUDP v1 packets
var OptionSupportedFunctions uint8 = 0

// OptionConnectionSignature is the ID for the Connection Signature option in PRUDP v1 packets
var OptionConnectionSignature uint8 = 1

// OptionFragmentID is the ID for the Fragment ID option in PRUDP v1 packets
var OptionFragmentID uint8 = 2

// OptionInitialSequenceID is the ID for the initial sequence ID option in PRUDP v1 packets
var OptionInitialSequenceID uint8 = 3

// OptionMaxSubstreamID is the ID for the max substream ID option in PRUDP v1 packets
var OptionMaxSubstreamID uint8 = 4

// PacketV1 reresents a PRUDPv1 packet
type PacketV1 struct {
	Packet
	magic              []byte
	substreamID        uint8
	supportedFunctions uint32
	initialSequenceID  uint16
	maximumSubstreamID uint8
}

// SetSubstreamID sets the packet substream ID
func (packet *PacketV1) SetSubstreamID(substreamID uint8) {
	packet.substreamID = substreamID
}

// SubstreamID returns the packet substream ID
func (packet *PacketV1) SubstreamID() uint8 {
	return packet.substreamID
}

// SetSupportedFunctions sets the packet supported functions flags
func (packet *PacketV1) SetSupportedFunctions(supportedFunctions uint32) {
	packet.supportedFunctions = supportedFunctions
}

// SupportedFunctions returns the packet supported functions flags
func (packet *PacketV1) SupportedFunctions() uint32 {
	return packet.supportedFunctions
}

// SetInitialSequenceID sets the packet initial sequence ID for unreliable packets
func (packet *PacketV1) SetInitialSequenceID(initialSequenceID uint16) {
	packet.initialSequenceID = initialSequenceID
}

// InitialSequenceID returns the packet initial sequence ID for unreliable packets
func (packet *PacketV1) InitialSequenceID() uint16 {
	return packet.initialSequenceID
}

// SetMaximumSubstreamID sets the packet maximum substream ID
func (packet *PacketV1) SetMaximumSubstreamID(maximumSubstreamID uint8) {
	packet.maximumSubstreamID = maximumSubstreamID
}

// MaximumSubstreamID returns the packet maximum substream ID
func (packet *PacketV1) MaximumSubstreamID() uint8 {
	return packet.maximumSubstreamID
}

// Decode decodes the packet
func (packet *PacketV1) Decode() error {
	if len(packet.Data()) < 30 { // magic + header + signature
		return errors.New("[PRUDPv1] Packet length less than minimum")
	}

	stream := NewStreamIn(packet.Data(), packet.Sender().Server())

	packet.magic = stream.ReadBytesNext(2)

	if !bytes.Equal(packet.magic, []byte{0xEA, 0xD0}) {
		return errors.New("PRUDPv1 packet magic did not match")
	}

	packet.SetVersion(stream.ReadUInt8())

	if packet.Version() != 1 {
		return errors.New("PRUDPv1 version did not match")
	}

	optionsLength := stream.ReadUInt8()
	payloadSize := stream.ReadUInt16LE()

	packet.SetSource(stream.ReadUInt8())
	packet.SetDestination(stream.ReadUInt8())

	typeFlags := stream.ReadUInt16LE()

	packet.SetType(typeFlags & 0xF)
	packet.SetFlags(typeFlags >> 4)

	if _, ok := validTypes[packet.Type()]; !ok {
		return errors.New("[PRUDPv1] Packet type not valid type")
	}

	packet.SetSessionID(stream.ReadUInt8())
	packet.SetSubstreamID(stream.ReadUInt8())
	packet.SetSequenceID(stream.ReadUInt16LE())

	packet.SetSignature(stream.ReadBytesNext(16))

	if len(packet.Data()[stream.ByteOffset():]) < int(optionsLength) {
		return errors.New("[PRUDPv1] Packet specific data size does not match")
	}

	options := stream.ReadBytesNext(int64(optionsLength))

	packet.decodeOptions(options)

	if payloadSize > 0 {
		if len(packet.Data()[stream.ByteOffset():]) < int(payloadSize) {
			return errors.New("[PRUDPv1] Packet data length less than payload length")
		}

		payloadCrypted := stream.ReadBytesNext(int64(payloadSize))

		packet.SetPayload(payloadCrypted)

		if packet.Type() == DataPacket && !packet.HasFlag(FlagMultiAck) {
			ciphered := make([]byte, payloadSize)

			packet.Sender().Decipher().XORKeyStream(ciphered, payloadCrypted)

			request := NewRMCRequest()
			err := request.FromBytes(ciphered)

			if err != nil {
				return errors.New("[PRUDPv1] Error parsing RMC request: " + err.Error())
			}

			packet.rmcRequest = request
		}
	}

	calculatedSignature := packet.calculateSignature(packet.Data()[2:14], packet.Sender().ServerConnectionSignature(), options, packet.Payload())

	if !bytes.Equal(calculatedSignature, packet.Signature()) {
		logger.Error("PRUDPv1 calculated signature did not match")
	}

	return nil
}

// Bytes encodes the packet and returns a byte array
func (packet *PacketV1) Bytes() []byte {

	if packet.Type() == DataPacket {

		if !packet.HasFlag(FlagMultiAck) {
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

	stream.WriteUInt16LE(0xD0EA) // v1 magic
	stream.WriteUInt8(1)

	options := packet.encodeOptions()
	optionsLength := len(options)

	stream.WriteUInt8(uint8(optionsLength))
	stream.WriteUInt16LE(uint16(len(packet.Payload())))
	stream.WriteUInt8(packet.Source())
	stream.WriteUInt8(packet.Destination())
	stream.WriteUInt16LE(typeFlags)
	stream.WriteUInt8(packet.SessionID())
	stream.WriteUInt8(packet.SubstreamID())
	stream.WriteUInt16LE(packet.SequenceID())

	signature := packet.calculateSignature(stream.Bytes()[2:14], packet.Sender().ClientConnectionSignature(), options, packet.Payload())

	stream.Grow(int64(len(signature)))
	stream.WriteBytesNext(signature)

	if optionsLength > 0 {
		stream.Grow(int64(optionsLength))
		stream.WriteBytesNext(options)
	}

	payload := packet.Payload()
	payloadLength := len(payload)

	if payload != nil && payloadLength > 0 {
		stream.Grow(int64(payloadLength))
		stream.WriteBytesNext(payload)
	}

	return stream.Bytes()
}

func (packet *PacketV1) decodeOptions(options []byte) {
	optionsStream := NewStreamIn(options, packet.Sender().Server())

	for optionsStream.ByteOffset() != optionsStream.ByteCapacity() {
		optionID := optionsStream.ReadUInt8()
		optionSize := optionsStream.ReadUInt8()

		switch optionID {
		case OptionSupportedFunctions:
			// Only need the LSB
			lsb := optionsStream.ReadBytesNext(int64(optionSize))[0]
			packet.sender.server.SetPRUDPProtocolMinorVersion(int(lsb))
			packet.SetSupportedFunctions(uint32(lsb))
		case OptionConnectionSignature:
			packet.SetConnectionSignature(optionsStream.ReadBytesNext(int64(optionSize)))
		case OptionFragmentID:
			packet.SetFragmentID(optionsStream.ReadUInt8())
		case OptionInitialSequenceID:
			packet.SetInitialSequenceID(optionsStream.ReadUInt16LE())
		case OptionMaxSubstreamID:
			packet.SetMaximumSubstreamID(optionsStream.ReadUInt8())
		}
	}
}

func (packet *PacketV1) encodeOptions() []byte {
	stream := NewStreamOut(packet.Sender().Server())

	if packet.Type() == SynPacket || packet.Type() == ConnectPacket {
		stream.WriteUInt8(OptionSupportedFunctions)
		stream.WriteUInt8(4)
		stream.WriteUInt32LE(packet.supportedFunctions)

		stream.WriteUInt8(OptionConnectionSignature)
		stream.WriteUInt8(16)
		stream.Grow(16)
		stream.WriteBytesNext(packet.ConnectionSignature())

		if packet.Type() == ConnectPacket {
			stream.WriteUInt8(OptionInitialSequenceID)
			stream.WriteUInt8(2)
			stream.WriteUInt16LE(packet.initialSequenceID)
		}

		stream.WriteUInt8(OptionMaxSubstreamID)
		stream.WriteUInt8(1)
		stream.WriteUInt8(packet.maximumSubstreamID)
	} else if packet.Type() == DataPacket {
		stream.WriteUInt8(OptionFragmentID)
		stream.WriteUInt8(1)
		stream.WriteUInt8(packet.FragmentID())
	}

	return stream.Bytes()
}

func (packet *PacketV1) calculateSignature(header []byte, connectionSignature []byte, options []byte, payload []byte) []byte {
	key := packet.Sender().SignatureKey()

	signatureBase := make([]byte, 4)
	binary.LittleEndian.PutUint32(signatureBase, uint32(packet.Sender().SignatureBase()))

	mac := hmac.New(md5.New, key)

	mac.Write(header[4:])
	mac.Write(packet.Sender().SessionKey())
	mac.Write(signatureBase)
	mac.Write(connectionSignature)
	mac.Write(options)
	mac.Write(payload)

	return mac.Sum(nil)
}

// NewPacketV1 returns a new PRUDPv1 packet
func NewPacketV1(client *Client, data []byte) (*PacketV1, error) {
	packet := NewPacket(client, data)
	packetv1 := PacketV1{Packet: packet}

	if data != nil {
		err := packetv1.Decode()

		if err != nil {
			return &PacketV1{}, errors.New("[PRUDPv1] Error decoding packet data: " + err.Error())
		}
	}

	return &packetv1, nil
}
