package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
)

// OptionAllFunctions is used with OptionSupportedFunctions to support all methods
var OptionAllFunctions = 0xFFFFFFFF

// OptionSupportedFunctions is the ID for the Supported Functions option in PRUDP v1 packets
var OptionSupportedFunctions = 0

// OptionConnectionSignature is the ID for the Connection Signature option in PRUDP v1 packets
var OptionConnectionSignature = 1

// OptionFragmentID is the ID for the Fragment ID option in PRUDP v1 packets
var OptionFragmentID = 2

// OptionInitialSequenceID is the ID for the initial sequence ID option in PRUDP v1 packets
var OptionInitialSequenceID = 3

// OptionMaxSubstreamID is the ID for the max substream ID option in PRUDP v1 packets
var OptionMaxSubstreamID = 4

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

// GetSubstreamID returns the packet substream ID
func (packet *PacketV1) GetSubstreamID() uint8 {
	return packet.substreamID
}

func (packet *PacketV1) SetSupportedFunctions(supportedFunctions uint32) {
	packet.supportedFunctions = supportedFunctions
}

func (packet *PacketV1) GetSupportedFunctions() uint32 {
	return packet.supportedFunctions
}

func (packet *PacketV1) SetInitialSequenceID(initialSequenceID uint16) {
	packet.initialSequenceID = initialSequenceID
}

func (packet *PacketV1) GetInitialSequenceID() uint16 {
	return packet.initialSequenceID
}

func (packet *PacketV1) SetMaximumSubstreamID(maximumSubstreamID uint8) {
	packet.maximumSubstreamID = maximumSubstreamID
}

func (packet *PacketV1) GetMaximumSubstreamID() uint8 {
	return packet.maximumSubstreamID
}

// Decode decodes the packet
func (packet *PacketV1) Decode() error {
	if len(packet.Data()) < 30 { // magic + header + signature
		return errors.New("[PRUDPv1] Packet length less than minimum")
	}

	stream := NewStreamIn(packet.Data(), packet.GetSender().GetServer())

	packet.magic = stream.ReadBytesNext(2)

	if !bytes.Equal(packet.magic, []byte{0xEA, 0xD0}) {
		return errors.New("PRUDPv1 packet magic did not match")
	}

	packet.SetVersion(uint8(stream.ReadByteNext()))

	if packet.GetVersion() != 1 {
		return errors.New("PRUDPv1 version did not match")
	}

	optionsLength := stream.ReadByteNext()
	payloadSize := stream.ReadU16LENext(1)[0]

	packet.SetSource(uint8(stream.ReadByteNext()))
	packet.SetDestination(uint8(stream.ReadByteNext()))

	typeFlags := stream.ReadU16LENext(1)[0]

	if packet.GetSender().GetServer().GetFlagsVersion() == 0 {
		packet.SetType(typeFlags & 7)
		packet.SetFlags(typeFlags >> 3)
	} else {
		packet.SetType(typeFlags & 0xF)
		packet.SetFlags(typeFlags >> 4)
	}

	if _, ok := validTypes[packet.GetType()]; !ok {
		return errors.New("[PRUDPv1] Packet type not valid type")
	}

	packet.SetSessionID(uint8(stream.ReadByteNext()))
	packet.SetSubstreamID(uint8(stream.ReadByteNext()))
	packet.SetSequenceID(stream.ReadU16LENext(1)[0])

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

		if packet.GetType() == DataPacket && !packet.HasFlag(FlagMultiAck) {
			ciphered := make([]byte, payloadSize)

			packet.GetSender().GetDecipher().XORKeyStream(ciphered, payloadCrypted)

			request, err := NewRMCRequest(ciphered)

			if err != nil {
				return errors.New("[PRUDPv1] Error parsing RMC request: " + err.Error())
			}

			packet.rmcRequest = request
		}
	}

	calculatedSignature := packet.calculateSignature(packet.Data()[2:14], packet.GetSender().GetServerConnectionSignature(), options, packet.GetPayload())

	if !bytes.Equal(calculatedSignature, packet.GetSignature()) {
		fmt.Println("[ERROR] Calculated signature did not match")
	}

	return nil
}

// Bytes encodes the packet and returns a byte array
func (packet *PacketV1) Bytes() []byte {

	if packet.GetType() == DataPacket {

		if !packet.HasFlag(FlagMultiAck) {
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
	stream.Grow(30)

	stream.WriteBytesNext([]byte{0xEA, 0xD0})
	stream.WriteByteNext(1)

	options := packet.encodeOptions()
	optionsLength := len(options)

	stream.WriteByteNext(byte(optionsLength))
	stream.WriteU16LENext([]uint16{uint16(len(packet.GetPayload()))})
	stream.WriteByteNext(packet.GetSource())
	stream.WriteByteNext(packet.GetDestination())
	stream.WriteU16LENext([]uint16{typeFlags})
	stream.WriteByteNext(packet.GetSessionID())
	stream.WriteByteNext(packet.GetSubstreamID())
	stream.WriteU16LENext([]uint16{packet.GetSequenceID()})

	signature := packet.calculateSignature(stream.Bytes()[2:14], packet.GetSender().GetClientConnectionSignature(), options, packet.GetPayload())

	stream.WriteBytesNext(signature)

	if optionsLength > 0 {
		stream.Grow(int64(optionsLength))
		stream.WriteBytesNext(options)
	}

	payload := packet.GetPayload()
	payloadLength := len(payload)

	if payload != nil && payloadLength > 0 {
		stream.Grow(int64(payloadLength))
		stream.WriteBytesNext(payload)
	}

	return stream.Bytes()
}

func (packet *PacketV1) decodeOptions(options []byte) {
	optionsStream := NewStreamIn(options, packet.GetSender().GetServer())

	for optionsStream.ByteOffset() != optionsStream.ByteCapacity() {
		optionID := optionsStream.ReadByteNext()
		optionSize := optionsStream.ReadByteNext()

		switch int(optionID) {
		case OptionSupportedFunctions:
			// Only need the LSB
			lsb := optionsStream.ReadBytesNext(int64(optionSize))[0]
			packet.SetSupportedFunctions(uint32(lsb))
		case OptionConnectionSignature:
			packet.SetConnectionSignature(optionsStream.ReadBytesNext(int64(optionSize)))
		case OptionFragmentID:
			packet.SetFragmentID(uint8(optionsStream.ReadByteNext()))
		case OptionInitialSequenceID:
			packet.SetInitialSequenceID(optionsStream.ReadU16LENext(1)[0])
		case OptionMaxSubstreamID:
			packet.SetMaximumSubstreamID(uint8(optionsStream.ReadByteNext()))
		}
	}
}

func (packet *PacketV1) encodeOptions() []byte {
	stream := NewStreamOut(packet.GetSender().GetServer())

	if packet.GetType() == SynPacket || packet.GetType() == ConnectPacket {
		stream.Grow(6)
		stream.WriteByteNext(uint8(OptionSupportedFunctions))
		stream.WriteByteNext(4)
		stream.WriteU32LENext([]uint32{packet.supportedFunctions})

		stream.Grow(18)
		stream.WriteByteNext(uint8(OptionConnectionSignature))
		stream.WriteByteNext(16)
		stream.WriteBytesNext(packet.GetConnectionSignature())

		if packet.GetType() == ConnectPacket {
			stream.Grow(4)
			stream.WriteByteNext(uint8(OptionInitialSequenceID))
			stream.WriteByteNext(2)
			stream.WriteU16LENext([]uint16{packet.initialSequenceID})
		}

		stream.Grow(3)
		stream.WriteByteNext(uint8(OptionMaxSubstreamID))
		stream.WriteByteNext(1)
		stream.WriteByteNext(packet.maximumSubstreamID)
	} else if packet.GetType() == DataPacket {
		stream.Grow(3)
		stream.WriteByteNext(uint8(OptionFragmentID))
		stream.WriteByteNext(1)
		stream.WriteByteNext(packet.GetFragmentID())
	}

	return stream.Bytes()
}

func (packet *PacketV1) calculateSignature(header []byte, connectionSignature []byte, options []byte, payload []byte) []byte {
	key := packet.GetSender().GetSignatureKey()

	signatureBase := make([]byte, 4)
	binary.LittleEndian.PutUint32(signatureBase, uint32(packet.GetSender().GetSignatureBase()))

	mac := hmac.New(md5.New, key)

	mac.Write(header[4:])
	mac.Write(packet.GetSender().GetSessionKey())
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
