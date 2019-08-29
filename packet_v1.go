package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	
	"github.com/superwhiskers/crunch"
)

func decodePacketV1(PRUDPPacket *Packet) map[string]interface{} {
	stream := crunch.NewBuffer(PRUDPPacket.Data)
	stream.SeekByte(2, true)
	if stream.ReadByteNext() != 1 {
		// invalid PRUDP version number?
	}

	optionsSize := stream.ReadByteNext()
	payloadSize := stream.ReadU16LENext(1)[0]
	source := stream.ReadByteNext()
	destination := stream.ReadByteNext()
	typeFlags := stream.ReadU16LENext(1)[0]
	sessionID := stream.ReadByteNext()
	multiAckVersion := stream.ReadByteNext()
	sequenceID := stream.ReadU16LENext(1)[0]
	checksum := stream.ReadBytesNext(16) // signature
	optionsData := stream.ReadBytesNext(int64(optionsSize))
	payload := stream.ReadBytesNext(int64(payloadSize))

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

	signatureCheck := CalculateV1Signature(PRUDPPacket.Sender, stream.ReadBytes(2,14), optionsData, PRUDPPacket.Sender.ServerConnectionSignature, payload)
	if !bytes.Equal(signatureCheck, checksum) {
		fmt.Println("[ERROR] Calculated checksum does not match decoded checksum!")
	}

	if packetType == Types["Data"] && len(payload) > 0 {
		crypted := make([]byte, len(payload))
		PRUDPPacket.Sender.Decipher.XORKeyStream(crypted, payload)

		request := NewRMCRequest(crypted)

		decoded["RMCRequest"] = request
	}

	return decoded
}

func encodePacketV1(PRUDPPacket *Packet) []byte {
	if PRUDPPacket.Type == Types["Data"] && len(PRUDPPacket.Payload) > 0 {
		crypted := make([]byte, len(PRUDPPacket.Payload))
		PRUDPPacket.Sender.Cipher.XORKeyStream(crypted, PRUDPPacket.Payload)
		PRUDPPacket.Payload = crypted
	}

	options := encodeOptionsV1(PRUDPPacket)

	stream := crunch.NewBuffer()
	headerStream := crunch.NewBuffer()

	headerStream.WriteByteNext(1)
	headerStream.WriteByteNext(uint8(len(options)))
	headerStream.WriteU16LENext([]uint16{uint16(len(PRUDPPacket.Payload))})
	headerStream.WriteByteNext(PRUDPPacket.Source)
	headerStream.WriteByteNext(PRUDPPacket.Destination)
	headerStream.WriteU16LENext([]uint16{uint16(int(PRUDPPacket.Type) | int(PRUDPPacket.Flags)<<4)})
	headerStream.WriteByteNext(PRUDPPacket.SessionID)
	headerStream.WriteByteNext(PRUDPPacket.MultiAckVersion)
	headerStream.WriteU16LENext([]uint16{PRUDPPacket.SequenceID})

	stream.WriteBytesNext([]byte{0xEA, 0xD0})
	stream.WriteBytesNext(headerStream.Bytes())
	stream.WriteBytesNext(CalculateV1Signature(PRUDPPacket.Sender, headerStream.Bytes(), options, PRUDPPacket.Sender.ClientConnectionSignature, PRUDPPacket.Payload))
	stream.WriteBytesNext(options)
	stream.WriteBytesNext(PRUDPPacket.Payload)

	return stream.Bytes()
}

func decodeV1Options(data []byte) map[int]interface{} {
	stream := crunch.NewBuffer()
	options := make(map[int]interface{})

	for int(stream.ByteOffset()) < len(data) {
		if len(data)-int(stream.ByteOffset()) < 2 {
			fmt.Println("Line 86")
			return nil
		}

		optType := int(stream.ReadByteNext())
		length := int(stream.ReadByteNext())

		if len(data)-int(stream.ByteOffset()) < length {
			fmt.Println("Line 94")
			return nil
		}

		if optType == OptionsSupport {
			if length != 4 {
				fmt.Println("Line 100")
				return nil
			}
			options[optType] = stream.ReadU32LENext(1)[0]
		} else if optType == OptionsConnectionSignature {
			if length != 16 {
				fmt.Println("Line 106")
				return nil
			}
			options[optType] = stream.ReadBytesNext(16)
		} else if optType == OptionsFragment || optType == Options4 {
			if length != 1 {
				fmt.Println("Line 112")
				return nil
			}
			options[optType] = int(stream.ReadByteNext())
		} else if optType == Options3 {
			if length != 2 {
				fmt.Println("Line 118")
				return nil
			}
			options[optType] = stream.ReadU16LENext(1)[0]
		}
	}

	return options
}

func encodeOptionsV1(PRUDPPacket *Packet) []byte {
	stream := crunch.NewBuffer()

	if PRUDPPacket.Type == Types["Syn"] || PRUDPPacket.Type == Types["Connect"] {
		stream.WriteByteNext(uint8(OptionsSupport))
		stream.WriteByteNext(4)
		stream.WriteU32LENext([]uint32{uint32(OptionsAll)})

		stream.WriteByteNext(uint8(OptionsConnectionSignature))
		stream.WriteByteNext(16)
		stream.WriteBytesNext(PRUDPPacket.Signature)

		if PRUDPPacket.Type == Types["Connect"] {
			stream.WriteByteNext(uint8(Options3))
			stream.WriteByteNext(2)
			stream.WriteU16LENext([]uint16{0xFFFF}) // I dunno
		}

		stream.WriteByteNext(uint8(Options4))
		stream.WriteByteNext(1)
		stream.WriteByteNext(0)
	} else if PRUDPPacket.Type == Types["Data"] {
		stream.WriteByteNext(uint8(OptionsFragment))
		stream.WriteByteNext(1)
		stream.WriteByteNext(uint8(PRUDPPacket.FragmentID))
	}

	return stream.Bytes()
}

// CalculateV1Signature calculates the HMAC signature for a given packet
func CalculateV1Signature(Client *Client, header []byte, options []byte, signature []byte, payload []byte) []byte {

	signatureBase := crunch.NewBuffer()
	signatureBase.WriteU32LENext([]uint32{uint32(Client.SignatureBase)})

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
