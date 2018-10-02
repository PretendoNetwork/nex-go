package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// PacketV1Header represents the start of a v1 PRUDP packet
type PacketV1Header struct {
	Magic           [2]byte
	Version         uint8
	OptionsSize     uint8
	PayloadSize     uint16
	Source          uint8
	Destination     uint8
	TypeFlags       uint16
	SessionID       uint8
	MultiAckVersion uint8
	SequenceID      uint16
	Signature       [16]byte
}

func decodeV1(PRUDPPacket *Packet) map[string]interface{} {
	stream := NewInputStream(PRUDPPacket.Data)
	stream.Skip(2) // Magic
	if stream.UInt8() != 1 {
		// invalid PRUDP version number?
	}

	optionsSize := stream.UInt8()
	payloadSize := stream.UInt16LE()
	source := stream.UInt8()
	destination := stream.UInt8()
	typeFlags := stream.UInt16LE()
	sessionID := stream.UInt8()
	multiAckVersion := stream.UInt8()
	sequenceID := stream.UInt16LE()
	checksum := stream.Bytes(16) // signature
	optionsData := stream.Bytes(int(optionsSize))
	payload := stream.Bytes(int(payloadSize))

	sourceType := source >> 4
	sourcePort := source & 0xF

	destinationType := destination >> 4
	destinationPort := destination & 0xF

	packetType := typeFlags & 0xF
	flags := typeFlags >> 4

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
		"Signature":       []byte{},
		"Payload":         payload,
		"MultiAckVersion": multiAckVersion,
	}

	options := decodeV1Options(optionsData)

	if packetType == Types["Syn"] || packetType == Types["Connect"] {
		if options[OptionsConnectionSignature] != nil {
			decoded["Signature"] = options[OptionsConnectionSignature]
		} else {
			// it's supposed to be, so I guess error out or something?
		}
	} else if packetType == Types["Data"] {
		if options[OptionsFragment] != nil {
			decoded["FragmentID"] = options[OptionsFragment]
		} else {
			// it's supposed to be, so I guess error out or something?
		}
	}

	//fmt.Println("CLIENT SIG", hex.EncodeToString(signature))

	signatureCheck := CalculateV1Signature(PRUDPPacket.Sender, stream.data[2:14], optionsData, PRUDPPacket.Sender.ServerConnectionSignature, payload)
	if !bytes.Equal(signatureCheck, checksum) {
		fmt.Println("[ERROR] Calculated checksum does not match decoded checksum!")
		//fmt.Println(hex.EncodeToString(PRUDPPacket.Sender.SecureKey))
		//fmt.Println(hex.EncodeToString(PRUDPPacket.Sender.ServerConnectionSignature))
		//fmt.Println(hex.EncodeToString(signatureCheck))
		//fmt.Println(hex.EncodeToString(signature))
	}

	if packetType == Types["Data"] && len(payload) > 0 {
		crypted := make([]byte, len(payload))
		PRUDPPacket.Sender.Decipher.XORKeyStream(crypted, payload)

		request := NewRMCRequest(crypted)

		decoded["RMCRequest"] = request
	}

	return decoded
}

func encodeV1(PRUDPPacket *Packet) []byte {
	if PRUDPPacket.Type == Types["Data"] && len(PRUDPPacket.Payload) > 0 {
		crypted := make([]byte, len(PRUDPPacket.Payload))
		PRUDPPacket.Sender.Cipher.XORKeyStream(crypted, PRUDPPacket.Payload)
		PRUDPPacket.Payload = crypted
	}

	options := encodeV1Options(PRUDPPacket)

	stream := NewOutputStream()
	headerStream := NewOutputStream()

	headerStream.UInt8(1)
	headerStream.UInt8(uint8(len(options)))
	headerStream.UInt16LE(uint16(len(PRUDPPacket.Payload)))
	headerStream.UInt8(PRUDPPacket.Source)
	headerStream.UInt8(PRUDPPacket.Destination)
	headerStream.UInt16LE(uint16(int(PRUDPPacket.Type) | int(PRUDPPacket.Flags)<<4))
	headerStream.UInt8(PRUDPPacket.SessionID)
	headerStream.UInt8(PRUDPPacket.MultiAckVersion)
	headerStream.UInt16LE(PRUDPPacket.SequenceID)

	stream.Write([]byte{0xEA, 0xD0})
	stream.Write(headerStream.Bytes())

	//stream.Write(PRUDPPacket.Signature)
	stream.Write(CalculateV1Signature(PRUDPPacket.Sender, headerStream.Bytes(), options, PRUDPPacket.Sender.ClientConnectionSignature, PRUDPPacket.Payload))
	stream.Write(options)
	stream.Write(PRUDPPacket.Payload)

	return stream.Bytes()
}

func decodeV1Options(data []byte) map[int]interface{} {
	stream := NewInputStream(data)
	options := make(map[int]interface{})

	for stream.pos < len(data) {
		if len(data)-stream.pos < 2 {
			fmt.Println("Line 86")
			return nil
		}

		optType := int(stream.UInt8())
		length := int(stream.UInt8())

		if len(data)-stream.pos < length {
			fmt.Println("Line 94")
			return nil
		}

		if optType == OptionsSupport {
			if length != 4 {
				fmt.Println("Line 100")
				return nil
			}
			options[optType] = stream.UInt32LE()
		} else if optType == OptionsConnectionSignature {
			if length != 16 {
				fmt.Println("Line 106")
				return nil
			}
			options[optType] = stream.Bytes(16)
		} else if optType == OptionsFragment || optType == Options4 {
			if length != 1 {
				fmt.Println("Line 112")
				return nil
			}
			options[optType] = int(stream.UInt8())
		} else if optType == Options3 {
			if length != 2 {
				fmt.Println("Line 118")
				return nil
			}
			options[optType] = stream.UInt16LE()
		}
	}

	return options
}

func encodeV1Options(PRUDPPacket *Packet) []byte {
	stream := NewOutputStream()

	if PRUDPPacket.Type == Types["Syn"] || PRUDPPacket.Type == Types["Connect"] {
		stream.UInt8(uint8(OptionsSupport))
		stream.UInt8(4)
		stream.UInt32LE(uint32(OptionsAll))

		stream.UInt8(uint8(OptionsConnectionSignature))
		stream.UInt8(16)
		stream.Write(PRUDPPacket.Signature)

		if PRUDPPacket.Type == Types["Connect"] {
			stream.UInt8(uint8(Options3))
			stream.UInt8(2)
			stream.UInt16LE(0xFFFF) // I dunno
		}

		stream.UInt8(uint8(Options4))
		stream.UInt8(1)
		stream.UInt8(0)
	} else if PRUDPPacket.Type == Types["Data"] {
		stream.UInt8(uint8(OptionsFragment))
		stream.UInt8(1)
		stream.UInt8(uint8(PRUDPPacket.FragmentID))
	}

	return stream.Bytes()
}

// CalculateV1Signature calculates the HMAC signature for a given packet
func CalculateV1Signature(Client *Client, header []byte, options []byte, signature []byte, payload []byte) []byte {

	signatureBase := bytes.NewBuffer(make([]byte, 0, 4))
	binary.Write(signatureBase, binary.LittleEndian, uint32(Client.SignatureBase))

	key, _ := hex.DecodeString(Client.SignatureKey)
	mac := hmac.New(md5.New, key)
	mac.Write(header[4:])
	mac.Write(Client.SecureKey)
	mac.Write(signatureBase.Bytes())
	mac.Write(signature)
	mac.Write(options)
	mac.Write(payload)
	return mac.Sum(nil)
}
