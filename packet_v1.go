package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
)

// Magic is the expected PRUDPv1 magic number
var Magic = []byte{0xEA, 0xD0}

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
	magic                     []byte
	substreamID               uint8
	prudpProtocolMinorVersion int
	supportedFunctions        int
	initialSequenceID         uint16
	maximumSubstreamID        uint8
}

// SetSubstreamID sets the packet substream ID
func (packet *PacketV1) SetSubstreamID(substreamID uint8) {
	packet.substreamID = substreamID
}

// SubstreamID returns the packet substream ID
func (packet *PacketV1) SubstreamID() uint8 {
	return packet.substreamID
}

// PRUDPProtocolMinorVersion returns the packet PRUDP minor version
func (packet *PacketV1) PRUDPProtocolMinorVersion() int {
	return packet.prudpProtocolMinorVersion
}

// SetPRUDPProtocolMinorVersion sets the packet PRUDP minor version
func (packet *PacketV1) SetPRUDPProtocolMinorVersion(prudpProtocolMinorVersion int) {
	packet.prudpProtocolMinorVersion = prudpProtocolMinorVersion
}

// SetSupportedFunctions sets the packet supported functions flags
func (packet *PacketV1) SetSupportedFunctions(supportedFunctions int) {
	packet.supportedFunctions = supportedFunctions
}

// SupportedFunctions returns the packet supported functions flags
func (packet *PacketV1) SupportedFunctions() int {
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
	stream := NewStreamIn(packet.Data(), packet.Sender().Server())

	if len(stream.Bytes()[stream.ByteOffset():]) < 2 {
		return errors.New("Failed to read PRUDPv1 magic. Not have enough data")
	}

	packet.magic = stream.ReadBytesNext(2)

	if !bytes.Equal(packet.magic, Magic) {
		return fmt.Errorf("Invalid PRUDPv1 magic. Expected %x, got %x", Magic, packet.magic)
	}

	version, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 version. %s", err.Error())
	}

	packet.SetVersion(version)

	if packet.Version() != 1 {
		return fmt.Errorf("Invalid PRUDPv1 version. Expected 1, got %d", packet.Version())
	}

	optionsLength, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 options length. %s", err.Error())
	}

	payloadSize, err := stream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 payload size. %s", err.Error())
	}

	source, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 source. %s", err.Error())
	}

	packet.SetSource(source)

	destination, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 destination. %s", err.Error())
	}

	packet.SetDestination(destination)

	typeFlags, err := stream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 type-flags. %s", err.Error())
	}

	packet.SetType(typeFlags & 0xF)

	if _, ok := validTypes[packet.Type()]; !ok {
		return errors.New("Invalid PRUDP packet type")
	}

	packet.SetFlags(typeFlags >> 4)

	sessionID, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 session ID. %s", err.Error())
	}

	packet.SetSessionID(sessionID)

	substreamID, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 substream ID. %s", err.Error())
	}

	packet.SetSubstreamID(substreamID)

	sequenceID, err := stream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 sequence ID. %s", err.Error())
	}

	packet.SetSequenceID(sequenceID)

	if len(stream.Bytes()[stream.ByteOffset():]) < 16 {
		return errors.New("Failed to read PRUDPv1 packet signature. Not have enough data")
	}

	packet.SetSignature(stream.ReadBytesNext(16))

	if len(packet.Data()[stream.ByteOffset():]) < int(optionsLength) {
		return errors.New("[PRUDPv1] Packet specific data size does not match")
	}

	options := stream.ReadBytesNext(int64(optionsLength))

	err = packet.decodeOptions(options)
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 options. %s", err.Error())
	}

	if payloadSize > 0 {
		if len(packet.Data()[stream.ByteOffset():]) < int(payloadSize) {
			return errors.New("Failed to read PRUDPv1 packet payload. Not enough data")
		}

		payloadCrypted := stream.ReadBytesNext(int64(payloadSize))

		packet.SetPayload(payloadCrypted)
	}

	calculatedSignature := packet.calculateSignature(packet.Data()[2:14], packet.Sender().ServerConnectionSignature(), options, packet.Payload())

	if !bytes.Equal(calculatedSignature, packet.Signature()) {
		logger.Error("PRUDPv1 calculated signature did not match")
	}

	return nil
}

// DecryptPayload decrypts the packets payload and sets the RMC request data
func (packet *PacketV1) DecryptPayload() error {
	if packet.Type() == DataPacket && !packet.HasFlag(FlagMultiAck) {
		ciphered := make([]byte, len(packet.Payload()))

		packet.Sender().Decipher().XORKeyStream(ciphered, packet.Payload())

		request := NewRMCRequest()
		err := request.FromBytes(ciphered)
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv1 RMC request. %s", err.Error())
		}

		packet.rmcRequest = request
	}

	return nil
}

// Bytes encodes the packet and returns a byte array
func (packet *PacketV1) Bytes() []byte {
	//if packet.Type() == DataPacket {
	//	if !packet.HasFlag(FlagMultiAck) {
	//		payload := packet.Payload()
	//
	//		if payload != nil || len(payload) > 0 {
	//			payloadSize := len(payload)
	//
	//			encrypted := make([]byte, payloadSize)
	//			packet.Sender().Cipher().XORKeyStream(encrypted, payload)
	//
	//			packet.SetPayload(encrypted)
	//		}
	//	}
	//
	//	if !packet.HasFlag(FlagHasSize) {
	//		packet.AddFlag(FlagHasSize)
	//	}
	//}

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

func (packet *PacketV1) decodeOptions(options []byte) error {
	optionsStream := NewStreamIn(options, packet.Sender().Server())

	for optionsStream.ByteOffset() != optionsStream.ByteCapacity() {
		optionID, err := optionsStream.ReadUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv1 option ID. %s", err.Error())
		}

		optionSize, err := optionsStream.ReadUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv1 option size for option ID %d. %s", optionID, err.Error())
		}

		switch optionID {
		case OptionSupportedFunctions:
			supportedFunctions, err := optionsStream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read PRUDPv1 option supported functions. %s", err.Error())
			}

			packet.sender.SetPRUDPProtocolMinorVersion(int(supportedFunctions & 0xFF))
			packet.sender.SetSupportedFunctions(int(supportedFunctions >> 8))
		case OptionConnectionSignature:
			packet.SetConnectionSignature(optionsStream.ReadBytesNext(int64(optionSize)))
		case OptionFragmentID:
			fragmentID, err := optionsStream.ReadUInt8()
			if err != nil {
				return fmt.Errorf("Failed to read PRUDPv1 option fragment ID. %s", err.Error())
			}

			packet.SetFragmentID(fragmentID)
		case OptionInitialSequenceID:
			sequenceID, err := optionsStream.ReadUInt16LE()
			if err != nil {
				return fmt.Errorf("Failed to read PRUDPv1 option sequence ID. %s", err.Error())
			}

			packet.SetInitialSequenceID(sequenceID)
		case OptionMaxSubstreamID:
			maximumSubstreamID, err := optionsStream.ReadUInt8()
			if err != nil {
				return fmt.Errorf("Failed to read PRUDPv1 option maximum substream ID. %s", err.Error())
			}

			packet.SetMaximumSubstreamID(maximumSubstreamID)
		}
	}

	return nil
}

func (packet *PacketV1) encodeOptions() []byte {
	stream := NewStreamOut(packet.Sender().Server())

	if packet.Type() == SynPacket || packet.Type() == ConnectPacket {
		stream.WriteUInt8(OptionSupportedFunctions)
		stream.WriteUInt8(4)
		stream.WriteUInt32LE(uint32(packet.prudpProtocolMinorVersion) | uint32(packet.supportedFunctions<<8))

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
	sessionKey := packet.Sender().SessionKey()

	signatureBase := make([]byte, 4)
	binary.LittleEndian.PutUint32(signatureBase, uint32(packet.Sender().SignatureBase()))

	mac := hmac.New(md5.New, key)

	mac.Write(header[4:])
	mac.Write(sessionKey)
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
			return &PacketV1{}, fmt.Errorf("Failed to decode PRUDPv1 packet. %s", err.Error())
		}
	}

	return &packetv1, nil
}
