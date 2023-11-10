package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// PRUDPPacketV1 represents a PRUDPv1 packet
type PRUDPPacketV1 struct {
	PRUDPPacket
	optionsLength               uint8
	payloadLength               uint16
	minorVersion                uint32
	supportedFunctions          uint32
	maximumSubstreamID          uint8
	initialUnreliableSequenceID uint16
}

// Version returns the packets PRUDP version
func (p *PRUDPPacketV1) Version() int {
	return 1
}

// Decode parses the packets data
func (p *PRUDPPacketV1) decode() error {
	if p.readStream.Remaining() < 2 {
		return errors.New("Failed to read PRUDPv1 magic. Not have enough data")
	}

	magic := p.readStream.ReadBytesNext(2)

	if !bytes.Equal(magic, []byte{0xEA, 0xD0}) {
		return fmt.Errorf("Invalid PRUDPv1 magic. Expected 0xEAD0, got 0x%x", magic)
	}

	err := p.decodeHeader()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPv1 header. %s", err.Error())
	}

	p.signature = p.readStream.ReadBytesNext(16)

	err = p.decodeOptions()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPv1 options. %s", err.Error())
	}

	p.payload = p.readStream.ReadBytesNext(int64(p.payloadLength))

	return nil
}

// Bytes encodes a PRUDPv1 packet into a byte slice
func (p *PRUDPPacketV1) Bytes() []byte {
	options := p.encodeOptions()

	p.optionsLength = uint8(len(options))

	header := p.encodeHeader()

	stream := NewStreamOut(nil)

	stream.Grow(2)
	stream.WriteBytesNext([]byte{0xEA, 0xD0})

	stream.Grow(12)
	stream.WriteBytesNext(header)

	stream.Grow(16)
	stream.WriteBytesNext(p.signature)

	stream.Grow(int64(p.optionsLength))
	stream.WriteBytesNext(options)

	stream.Grow(int64(len(p.payload)))
	stream.WriteBytesNext(p.payload)

	return stream.Bytes()
}

func (p *PRUDPPacketV1) decodeHeader() error {
	if p.readStream.Remaining() < 12 {
		return errors.New("Failed to read PRUDPv1 magic. Not have enough data")
	}

	version, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPv1 version. %s", err.Error())
	}

	if version != 1 {
		return fmt.Errorf("Invalid PRUDPv1 version. Expected 1, got %d", version)
	}

	p.optionsLength, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPv1 options length. %s", err.Error())
	}

	p.payloadLength, err = p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPv1 payload length. %s", err.Error())
	}

	source, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 source. %s", err.Error())
	}

	destination, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 destination. %s", err.Error())
	}

	p.sourceStreamType = source >> 4
	p.sourcePort = source & 0xF
	p.destinationStreamType = destination >> 4
	p.destinationPort = destination & 0xF

	// TODO - Does QRV also encode it this way in PRUDPv1?
	typeAndFlags, err := p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 type and flags. %s", err.Error())
	}

	p.flags = typeAndFlags >> 4
	p.packetType = typeAndFlags & 0xF

	if _, ok := validPacketTypes[p.packetType]; !ok {
		return errors.New("Invalid PRUDPv1 packet type")
	}

	p.sessionID, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 session ID. %s", err.Error())
	}

	p.substreamID, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 substream ID. %s", err.Error())
	}

	p.sequenceID, err = p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 sequence ID. %s", err.Error())
	}

	return nil
}

func (p *PRUDPPacketV1) encodeHeader() []byte {
	stream := NewStreamOut(nil)

	stream.WriteUInt8(1) // * Version
	stream.WriteUInt8(p.optionsLength)
	stream.WriteUInt16LE(uint16(len(p.payload)))
	stream.WriteUInt8(p.sourcePort | (p.sourceStreamType << 4))
	stream.WriteUInt8(p.destinationPort | (p.destinationStreamType << 4))
	stream.WriteUInt16LE(p.packetType | (p.flags << 4)) // TODO - Does QRV also encode it this way in PRUDPv1?
	stream.WriteUInt8(p.sessionID)
	stream.WriteUInt8(p.substreamID)
	stream.WriteUInt16LE(p.sequenceID)

	return stream.Bytes()
}

func (p *PRUDPPacketV1) decodeOptions() error {
	data := p.readStream.ReadBytesNext(int64(p.optionsLength))
	optionsStream := NewStreamIn(data, nil)

	for optionsStream.Remaining() > 0 {
		optionID, err := optionsStream.ReadUInt8()
		if err != nil {
			return err
		}

		_, err = optionsStream.ReadUInt8() // * Options size. We already know the size based on the ID, though
		if err != nil {
			return err
		}

		if p.packetType == SynPacket || p.packetType == ConnectPacket {
			if optionID == 0 {
				p.supportedFunctions, err = optionsStream.ReadUInt32LE()

				p.minorVersion = p.supportedFunctions & 0xFF
				p.supportedFunctions = p.supportedFunctions >> 8
			}

			if optionID == 1 {
				p.connectionSignature = optionsStream.ReadBytesNext(16)
			}

			if optionID == 4 {
				p.maximumSubstreamID, err = optionsStream.ReadUInt8()
			}
		}

		if p.packetType == ConnectPacket {
			if optionID == 3 {
				p.initialUnreliableSequenceID, err = optionsStream.ReadUInt16LE()
			}
		}

		if p.packetType == DataPacket {
			if optionID == 2 {
				p.fragmentID, err = optionsStream.ReadUInt8()
			}
		}

		// * Only one option is processed at a time, so we can
		// * just check for errors here rather than after EVERY
		// * read
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PRUDPPacketV1) encodeOptions() []byte {
	optionsStream := NewStreamOut(nil)

	if p.packetType == SynPacket || p.packetType == ConnectPacket {
		optionsStream.WriteUInt8(0)
		optionsStream.WriteUInt8(4)
		optionsStream.WriteUInt32LE(p.minorVersion | (p.supportedFunctions << 8))

		optionsStream.WriteUInt8(1)
		optionsStream.WriteUInt8(16)
		optionsStream.Grow(16)
		optionsStream.WriteBytesNext(p.connectionSignature)

		// * Encoded here for NintendoClients compatibility.
		// * The order of these options should not matter,
		// * however when NintendoClients calculates the
		// * signature it does NOT use the original options
		// * section, and instead re-encodes the data in a
		// * specific order. Due to how this section is
		// * parsed, though, order REALLY doesn't matter.
		// * NintendoClients expects option 3 before 4, though
		if p.packetType == ConnectPacket {
			optionsStream.WriteUInt8(3)
			optionsStream.WriteUInt8(2)
			optionsStream.WriteUInt16LE(p.initialUnreliableSequenceID)
		}

		optionsStream.WriteUInt8(4)
		optionsStream.WriteUInt8(1)
		optionsStream.WriteUInt8(p.maximumSubstreamID)
	}

	if p.packetType == DataPacket {
		optionsStream.WriteUInt8(2)
		optionsStream.WriteUInt8(1)
		optionsStream.WriteUInt8(p.fragmentID)
	}

	return optionsStream.Bytes()
}

func (p *PRUDPPacketV1) calculateConnectionSignature(addr net.Addr) ([]byte, error) {
	var ip net.IP
	var port int

	switch v := addr.(type) {
	case *net.UDPAddr:
		ip = v.IP.To4()
		port = v.Port
	default:
		return nil, fmt.Errorf("Unsupported network type: %T", addr)
	}

	// * The real client seems to not care about this. The original
	// * server just used rand.Read here. This is done to implement
	// * compatibility with NintendoClients, as this is how it
	// * calculates PRUDPv1 connection signatures
	key := []byte{0x26, 0xc3, 0x1f, 0x38, 0x1e, 0x46, 0xd6, 0xeb, 0x38, 0xe1, 0xaf, 0x6a, 0xb7, 0x0d, 0x11}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))

	data := append(ip, portBytes...)
	hash := hmac.New(md5.New, key)
	hash.Write(data)

	return hash.Sum(nil), nil
}

func (p *PRUDPPacketV1) calculateSignature(sessionKey, connectionSignature []byte) []byte {
	accessKeyBytes := []byte(p.sender.server.accessKey)
	options := p.encodeOptions()
	header := p.encodeHeader()

	accessKeySum := sum[byte, uint32](accessKeyBytes)
	accessKeySumBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(accessKeySumBytes, accessKeySum)

	key := md5.Sum(accessKeyBytes)
	mac := hmac.New(md5.New, key[:])

	mac.Write(header[4:])
	mac.Write(sessionKey)
	mac.Write(accessKeySumBytes)
	mac.Write(connectionSignature)
	mac.Write(options)
	mac.Write(p.payload)

	return mac.Sum(nil)
}

// NewPRUDPPacketV1 creates and returns a new PacketV1 using the provided Client and stream
func NewPRUDPPacketV1(client *PRUDPClient, readStream *StreamIn) (*PRUDPPacketV1, error) {
	packet := &PRUDPPacketV1{
		PRUDPPacket: PRUDPPacket{
			sender:     client,
			readStream: readStream,
		},
	}

	if readStream != nil {
		err := packet.decode()
		if err != nil {
			return nil, fmt.Errorf("Failed to decode PRUDPv1 packet. %s", err.Error())
		}
	}

	return packet, nil
}

// NewPRUDPPacketsV1 reads all possible PRUDPv1 packets from the stream
func NewPRUDPPacketsV1(client *PRUDPClient, readStream *StreamIn) ([]PRUDPPacketInterface, error) {
	packets := make([]PRUDPPacketInterface, 0)

	for readStream.Remaining() > 0 {
		packet, err := NewPRUDPPacketV1(client, readStream)
		if err != nil {
			return packets, err
		}

		packets = append(packets, packet)
	}

	return packets, nil
}
