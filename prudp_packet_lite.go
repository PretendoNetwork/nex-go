package nex

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/PretendoNetwork/nex-go/v2/constants"
)

// PRUDPPacketLite represents a PRUDPLite packet
type PRUDPPacketLite struct {
	PRUDPPacket
	sourceVirtualPortStreamType      constants.StreamType
	sourceVirtualPortStreamID        uint8
	destinationVirtualPortStreamType constants.StreamType
	destinationVirtualPortStreamID   uint8
	optionsLength                    uint8
	minorVersion                     uint32
	supportedFunctions               uint32
	maximumSubstreamID               uint8
	initialUnreliableSequenceID      uint16
	liteSignature                    []byte
}

// SetSourceVirtualPortStreamType sets the packets source VirtualPort StreamType
func (p *PRUDPPacketLite) SetSourceVirtualPortStreamType(streamType constants.StreamType) {
	p.sourceVirtualPortStreamType = streamType
}

// SourceVirtualPortStreamType returns the packets source VirtualPort StreamType
func (p *PRUDPPacketLite) SourceVirtualPortStreamType() constants.StreamType {
	return p.sourceVirtualPortStreamType
}

// SetSourceVirtualPortStreamID sets the packets source VirtualPort port number
func (p *PRUDPPacketLite) SetSourceVirtualPortStreamID(port uint8) {
	p.sourceVirtualPortStreamID = port
}

// SourceVirtualPortStreamID returns the packets source VirtualPort port number
func (p *PRUDPPacketLite) SourceVirtualPortStreamID() uint8 {
	return p.sourceVirtualPortStreamID
}

// SetDestinationVirtualPortStreamType sets the packets destination VirtualPort constants.StreamType
func (p *PRUDPPacketLite) SetDestinationVirtualPortStreamType(streamType constants.StreamType) {
	p.destinationVirtualPortStreamType = streamType
}

// DestinationVirtualPortStreamType returns the packets destination VirtualPort constants.StreamType
func (p *PRUDPPacketLite) DestinationVirtualPortStreamType() constants.StreamType {
	return p.destinationVirtualPortStreamType
}

// SetDestinationVirtualPortStreamID sets the packets destination VirtualPort port number
func (p *PRUDPPacketLite) SetDestinationVirtualPortStreamID(port uint8) {
	p.destinationVirtualPortStreamID = port
}

// DestinationVirtualPortStreamID returns the packets destination VirtualPort port number
func (p *PRUDPPacketLite) DestinationVirtualPortStreamID() uint8 {
	return p.destinationVirtualPortStreamID
}

// Copy copies the packet into a new PRUDPPacketLite
//
// Retains the same PRUDPConnection pointer
func (p *PRUDPPacketLite) Copy() PRUDPPacketInterface {
	copied, _ := NewPRUDPPacketLite(p.server, p.sender, nil)

	copied.server = p.server
	copied.sourceVirtualPortStreamType = p.sourceVirtualPortStreamType
	copied.sourceVirtualPortStreamID = p.sourceVirtualPortStreamID
	copied.destinationVirtualPortStreamType = p.destinationVirtualPortStreamType
	copied.destinationVirtualPortStreamID = p.destinationVirtualPortStreamID
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

	p.sourceVirtualPortStreamType = constants.StreamType(streamTypes >> 4)
	p.destinationVirtualPortStreamType = constants.StreamType(streamTypes & 0xF)

	p.sourceVirtualPortStreamID, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPLite virtual source port. %s", err.Error())
	}

	p.destinationVirtualPortStreamID, err = p.readStream.ReadUInt8()
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

	if p.readStream.Remaining() < uint64(payloadLength) {
		return errors.New("Failed to read PRUDPLite payload. Not have enough data")
	}

	p.payload = p.readStream.ReadBytesNext(int64(payloadLength))

	return nil
}

// Bytes encodes a PRUDPLite packet into a byte slice
func (p *PRUDPPacketLite) Bytes() []byte {
	options := p.encodeOptions()

	stream := NewByteStreamOut(p.server.LibraryVersions, p.server.ByteStreamSettings)

	stream.WriteUInt8(0x80)
	stream.WriteUInt8(uint8(len(options)))
	stream.WriteUInt16LE(uint16(len(p.payload)))
	stream.WriteUInt8(uint8((p.sourceVirtualPortStreamType << 4) | p.destinationVirtualPortStreamType))
	stream.WriteUInt8(p.sourceVirtualPortStreamID)
	stream.WriteUInt8(p.destinationVirtualPortStreamID)
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
	if p.readStream.Remaining() < uint64(p.optionsLength) {
		return errors.New("Not have enough data")
	}

	data := p.readStream.ReadBytesNext(int64(p.optionsLength))
	optionsStream := NewByteStreamIn(data, p.server.LibraryVersions, p.server.ByteStreamSettings)

	for optionsStream.Remaining() > 0 {
		optionID, err := optionsStream.ReadUInt8()
		if err != nil {
			return err
		}

		optionSize, err := optionsStream.ReadUInt8() // * Options size. We already know the size based on the ID, though
		if err != nil {
			return err
		}

		if p.packetType == constants.SynPacket || p.packetType == constants.ConnectPacket {
			if optionID == 0 {
				p.supportedFunctions, err = optionsStream.ReadUInt32LE()

				p.minorVersion = p.supportedFunctions & 0xFF
				p.supportedFunctions = p.supportedFunctions >> 8
			}

			if optionID == 1 {
				if optionsStream.Remaining() < uint64(optionSize) {
					err = errors.New("Failed to read connection signature. Not have enough data")
				} else {
					p.connectionSignature = optionsStream.ReadBytesNext(int64(optionSize))
				}
			}

			if optionID == 4 {
				p.maximumSubstreamID, err = optionsStream.ReadUInt8()
			}
		}

		if p.packetType == constants.ConnectPacket {
			if optionID == 3 {
				p.initialUnreliableSequenceID, err = optionsStream.ReadUInt16LE()
			}
		}

		if p.packetType == constants.DataPacket {
			if optionID == 2 {
				p.fragmentID, err = optionsStream.ReadUInt8()
			}
		}

		if p.packetType == constants.ConnectPacket && !p.HasFlag(constants.PacketFlagAck) {
			if optionID == 0x80 {
				if optionsStream.Remaining() < uint64(optionSize) {
					err = errors.New("Failed to read lite signature. Not have enough data")
				} else {
					p.liteSignature = optionsStream.ReadBytesNext(int64(optionSize))
				}
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
	optionsStream := NewByteStreamOut(p.server.LibraryVersions, p.server.ByteStreamSettings)

	if p.packetType == constants.SynPacket || p.packetType == constants.ConnectPacket {
		optionsStream.WriteUInt8(0)
		optionsStream.WriteUInt8(4)
		optionsStream.WriteUInt32LE(p.minorVersion | (p.supportedFunctions << 8))

		if p.packetType == constants.SynPacket && p.HasFlag(constants.PacketFlagAck) {
			optionsStream.WriteUInt8(1)
			optionsStream.WriteUInt8(16)
			optionsStream.Grow(16)
			optionsStream.WriteBytesNext(p.connectionSignature)
		}

		if p.packetType == constants.ConnectPacket && !p.HasFlag(constants.PacketFlagAck) {
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
func NewPRUDPPacketLite(server *PRUDPServer, connection *PRUDPConnection, readStream *ByteStreamIn) (*PRUDPPacketLite, error) {
	packet := &PRUDPPacketLite{
		PRUDPPacket: PRUDPPacket{
			sender:     connection,
			readStream: readStream,
		},
	}

	packet.server = server

	if readStream != nil {
		err := packet.decode()
		if err != nil {
			return nil, fmt.Errorf("Failed to decode PRUDPLite packet. %s", err.Error())
		}
	}

	return packet, nil
}

// NewPRUDPPacketsLite reads all possible PRUDPLite packets from the stream
func NewPRUDPPacketsLite(server *PRUDPServer, connection *PRUDPConnection, readStream *ByteStreamIn) ([]PRUDPPacketInterface, error) {
	packets := make([]PRUDPPacketInterface, 0)

	for readStream.Remaining() > 0 {
		packet, err := NewPRUDPPacketLite(server, connection, readStream)
		if err != nil {
			return packets, err
		}

		packets = append(packets, packet)
	}

	return packets, nil
}
