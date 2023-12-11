package nex

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"net"
)

// PRUDPPacketLite represents a PRUDPLite packet
type PRUDPPacketLite struct {
	PRUDPPacket
	optionsLength               uint8
	minorVersion                uint32
	supportedFunctions          uint32
	maximumSubstreamID          uint8
	initialUnreliableSequenceID uint16
	liteSignature               []byte
}

// Copy copies the packet into a new PRUDPPacketLite
//
// Retains the same PRUDPClient pointer
func (p *PRUDPPacketLite) Copy() PRUDPPacketInterface {
	copied, _ := NewPRUDPPacketLite(p.sender, nil)

	copied.server = p.server
	copied.sourceStreamType = p.sourceStreamType
	copied.sourcePort = p.sourcePort
	copied.destinationStreamType = p.destinationStreamType
	copied.destinationPort = p.destinationPort
	copied.packetType = p.packetType
	copied.flags = p.flags
	copied.sessionID = p.sessionID
	copied.substreamID = p.substreamID

	if p.signature != nil {
		copied.signature = append([]byte(nil), p.signature...)
	}

	copied.sequenceID = p.sequenceID

	if p.connectionSignature != nil {
		copied.connectionSignature = append([]byte(nil), p.connectionSignature...)
	}

	copied.fragmentID = p.fragmentID

	if p.payload != nil {
		copied.payload = append([]byte(nil), p.payload...)
	}

	if p.message != nil {
		copied.message = p.message.Copy()
	}

	copied.optionsLength = p.optionsLength
	copied.minorVersion = p.minorVersion
	copied.supportedFunctions = p.supportedFunctions
	copied.maximumSubstreamID = p.maximumSubstreamID
	copied.initialUnreliableSequenceID = p.initialUnreliableSequenceID

	return copied
}

// Version returns the packets PRUDP version
func (p *PRUDPPacketLite) Version() int {
	return 2
}

// decode parses the packets data
func (p *PRUDPPacketLite) decode() error {
	magic, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPLite magic. %s", err.Error())
	}

	if magic != 0x80 {
		return fmt.Errorf("Invalid PRUDPLite magic. Expected 0x80, got 0x%x", magic)
	}

	p.optionsLength, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite options length. %s", err.Error())
	}

	payloadLength, err := p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite payload length. %s", err.Error())
	}

	streamTypes, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite virtual ports stream types. %s", err.Error())
	}

	p.sourceStreamType = streamTypes >> 4
	p.destinationStreamType = streamTypes & 0xF

	p.sourcePort, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite virtual source port. %s", err.Error())
	}

	p.destinationPort, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite virtual destination port. %s", err.Error())
	}

	p.fragmentID, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite fragment ID. %s", err.Error())
	}

	typeAndFlags, err := p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPLite type and flags. %s", err.Error())
	}

	p.flags = typeAndFlags >> 4
	p.packetType = typeAndFlags & 0xF

	p.sequenceID, err = p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite sequence ID. %s", err.Error())
	}

	err = p.decodeOptions()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite options. %s", err.Error())
	}

	p.payload = p.readStream.ReadBytesNext(int64(payloadLength))

	return nil
}

// Bytes encodes a PRUDPLite packet into a byte slice
func (p *PRUDPPacketLite) Bytes() []byte {
	options := p.encodeOptions()

	stream := NewStreamOut(nil)

	stream.WriteUInt8(0x80)
	stream.WriteUInt8(uint8(len(options)))
	stream.WriteUInt16LE(uint16(len(p.payload)))
	stream.WriteUInt8((p.sourceStreamType << 4) | p.destinationStreamType)
	stream.WriteUInt8(p.sourcePort)
	stream.WriteUInt8(p.destinationPort)
	stream.WriteUInt8(p.fragmentID)
	stream.WriteUInt16LE(p.packetType | (p.flags << 4))
	stream.WriteUInt16LE(p.sequenceID)

	stream.Grow(int64(len(options)))
	stream.WriteBytesNext(options)

	stream.Grow(int64(len(p.payload)))
	stream.WriteBytesNext(p.payload)

	return stream.Bytes()
}

func (p *PRUDPPacketLite) decodeOptions() error {
	data := p.readStream.ReadBytesNext(int64(p.optionsLength))
	optionsStream := NewStreamIn(data, nil)

	for optionsStream.Remaining() > 0 {
		optionID, err := optionsStream.ReadUInt8()
		if err != nil {
			return err
		}

		optionSize, err := optionsStream.ReadUInt8() // * Options size. We already know the size based on the ID, though
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
				p.connectionSignature = optionsStream.ReadBytesNext(int64(optionSize))
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

		if p.packetType == ConnectPacket && !p.HasFlag(FlagAck) {
			if optionID == 0x80 {
				p.liteSignature = optionsStream.ReadBytesNext(int64(optionSize))
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

func (p *PRUDPPacketLite) encodeOptions() []byte {
	optionsStream := NewStreamOut(nil)

	if p.packetType == SynPacket || p.packetType == ConnectPacket {
		optionsStream.WriteUInt8(0)
		optionsStream.WriteUInt8(4)
		optionsStream.WriteUInt32LE(p.minorVersion | (p.supportedFunctions << 8))

		if p.packetType == SynPacket && p.HasFlag(FlagAck) {
			optionsStream.WriteUInt8(1)
			optionsStream.WriteUInt8(16)
			optionsStream.Grow(16)
			optionsStream.WriteBytesNext(p.connectionSignature)
		}

		if p.packetType == ConnectPacket && !p.HasFlag(FlagAck) {
			optionsStream.WriteUInt8(1)
			optionsStream.WriteUInt8(16)
			optionsStream.Grow(16)
			optionsStream.WriteBytesNext(p.liteSignature)
		}
	}

	return optionsStream.Bytes()
}

func (p *PRUDPPacketLite) calculateConnectionSignature(addr net.Addr) ([]byte, error) {
	var ip net.IP
	var port int

	switch v := addr.(type) {
	case *net.TCPAddr:
		ip = v.IP.To4()
		port = v.Port
	default:
		return nil, fmt.Errorf("Unsupported network type: %T", addr)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))

	data := append(ip, portBytes...)
	hash := hmac.New(md5.New, p.server.PRUDPv1ConnectionSignatureKey)
	hash.Write(data)

	return hash.Sum(nil), nil
}

func (p *PRUDPPacketLite) calculateSignature(sessionKey, connectionSignature []byte) []byte {
	// * PRUDPLite has no signature
	return make([]byte, 0)
}

// NewPRUDPPacketLite creates and returns a new PacketLite using the provided Client and stream
func NewPRUDPPacketLite(client *PRUDPClient, readStream *StreamIn) (*PRUDPPacketLite, error) {
	packet := &PRUDPPacketLite{
		PRUDPPacket: PRUDPPacket{
			sender:     client,
			readStream: readStream,
		},
	}

	if readStream != nil {
		packet.server = readStream.Server.(*PRUDPServer)
		err := packet.decode()
		if err != nil {
			return nil, fmt.Errorf("Failed to decode PRUDPLite packet. %s", err.Error())
		}
	}

	if client != nil {
		packet.server = client.server
	}

	return packet, nil
}

// NewPRUDPPacketsLite reads all possible PRUDPLite packets from the stream
func NewPRUDPPacketsLite(client *PRUDPClient, readStream *StreamIn) ([]PRUDPPacketInterface, error) {
	packets := make([]PRUDPPacketInterface, 0)

	for readStream.Remaining() > 0 {
		packet, err := NewPRUDPPacketLite(client, readStream)
		if err != nil {
			return packets, err
		}

		packets = append(packets, packet)
	}

	return packets, nil
}
