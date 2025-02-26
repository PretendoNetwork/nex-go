package nex

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/PretendoNetwork/nex-go/v2/constants"
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

// Copy copies the packet into a new PRUDPPacketV1
//
// Retains the same PRUDPConnection pointer
func (p *PRUDPPacketV1) Copy() PRUDPPacketInterface {
	copied, _ := NewPRUDPPacketV1(p.server, p.sender, nil)

	copied.server = p.server
	copied.sourceVirtualPort = p.sourceVirtualPort
	copied.destinationVirtualPort = p.destinationVirtualPort
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
	copied.payloadLength = p.payloadLength
	copied.minorVersion = p.minorVersion
	copied.supportedFunctions = p.supportedFunctions
	copied.maximumSubstreamID = p.maximumSubstreamID
	copied.initialUnreliableSequenceID = p.initialUnreliableSequenceID

	return copied
}

// Version returns the packets PRUDP version
func (p *PRUDPPacketV1) Version() int {
	return 1
}

// decode parses the packets data
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

	if p.readStream.Remaining() < 16 {
		return errors.New("Failed to read PRUDPv1 signature. Not have enough data")
	}

	p.signature = p.readStream.ReadBytesNext(16)

	err = p.decodeOptions()
	if err != nil {
		return fmt.Errorf("Failed to decode PRUDPv1 options. %s", err.Error())
	}

	if p.readStream.Remaining() < uint64(p.payloadLength) {
		return errors.New("Failed to read PRUDPv1 payload. Not have enough data")
	}

	p.payload = p.readStream.ReadBytesNext(int64(p.payloadLength))

	return nil
}

// Bytes encodes a PRUDPv1 packet into a byte slice
func (p *PRUDPPacketV1) Bytes() []byte {
	options := p.encodeOptions()

	p.optionsLength = uint8(len(options))

	header := p.encodeHeader()

	stream := NewByteStreamOut(p.server.LibraryVersions, p.server.ByteStreamSettings)

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

	p.sourceVirtualPort = VirtualPort(source)

	destination, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 destination. %s", err.Error())
	}

	p.destinationVirtualPort = VirtualPort(destination)

	// TODO - Does QRV also encode it this way in PRUDPv1?
	typeAndFlags, err := p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv1 type and flags. %s", err.Error())
	}

	p.flags = typeAndFlags >> 4
	p.packetType = typeAndFlags & 0xF

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
	stream := NewByteStreamOut(p.server.LibraryVersions, p.server.ByteStreamSettings)

	stream.WriteUInt8(1) // * Version
	stream.WriteUInt8(p.optionsLength)
	stream.WriteUInt16LE(uint16(len(p.payload)))
	stream.WriteUInt8(uint8(p.sourceVirtualPort))
	stream.WriteUInt8(uint8(p.destinationVirtualPort))
	stream.WriteUInt16LE(p.packetType | (p.flags << 4)) // TODO - Does QRV also encode it this way in PRUDPv1?
	stream.WriteUInt8(p.sessionID)
	stream.WriteUInt8(p.substreamID)
	stream.WriteUInt16LE(p.sequenceID)

	return stream.Bytes()
}

func (p *PRUDPPacketV1) decodeOptions() error {
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

		_, err = optionsStream.ReadUInt8() // * Options size. We already know the size based on the ID, though
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
				if optionsStream.Remaining() < 16 {
					err = errors.New("Not have enough data")
				} else {
					p.connectionSignature = optionsStream.ReadBytesNext(16)
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
	optionsStream := NewByteStreamOut(p.server.LibraryVersions, p.server.ByteStreamSettings)

	if p.packetType == constants.SynPacket || p.packetType == constants.ConnectPacket {
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
		if p.packetType == constants.ConnectPacket {
			optionsStream.WriteUInt8(3)
			optionsStream.WriteUInt8(2)
			optionsStream.WriteUInt16LE(p.initialUnreliableSequenceID)
		}

		optionsStream.WriteUInt8(4)
		optionsStream.WriteUInt8(1)
		optionsStream.WriteUInt8(p.maximumSubstreamID)
	}

	if p.packetType == constants.DataPacket {
		optionsStream.WriteUInt8(2)
		optionsStream.WriteUInt8(1)
		optionsStream.WriteUInt8(p.fragmentID)
	}

	return optionsStream.Bytes()
}

func (p *PRUDPPacketV1) calculateConnectionSignature(addr net.Addr) ([]byte, error) {
	return p.server.PRUDPV1Settings.ConnectionSignatureCalculator(p, addr)
}

func (p *PRUDPPacketV1) calculateSignature(sessionKey, connectionSignature []byte) []byte {
	return p.server.PRUDPV1Settings.SignatureCalculator(p, sessionKey, connectionSignature)
}

// NewPRUDPPacketV1 creates and returns a new PacketV1 using the provided Client and stream
func NewPRUDPPacketV1(server *PRUDPServer, connection *PRUDPConnection, readStream *ByteStreamIn) (*PRUDPPacketV1, error) {
	packet := &PRUDPPacketV1{
		PRUDPPacket: PRUDPPacket{
			sender:     connection,
			readStream: readStream,
			version:    1,
		},
	}

	packet.server = server

	if readStream != nil {
		err := packet.decode()
		if err != nil {
			return nil, fmt.Errorf("Failed to decode PRUDPv1 packet. %s", err.Error())
		}
	}

	return packet, nil
}

// NewPRUDPPacketsV1 reads all possible PRUDPv1 packets from the stream
func NewPRUDPPacketsV1(server *PRUDPServer, connection *PRUDPConnection, readStream *ByteStreamIn) ([]PRUDPPacketInterface, error) {
	packets := make([]PRUDPPacketInterface, 0)

	for readStream.Remaining() > 0 {
		packet, err := NewPRUDPPacketV1(server, connection, readStream)
		if err != nil {
			return packets, err
		}

		packets = append(packets, packet)
	}

	return packets, nil
}

func defaultPRUDPv1ConnectionSignature(packet *PRUDPPacketV1, addr net.Addr) ([]byte, error) {
	var ip net.IP
	var port int

	switch v := addr.(type) {
	case *net.UDPAddr:
		ip = v.IP.To4()
		port = v.Port
	default:
		return nil, fmt.Errorf("Unsupported network type: %T", addr)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))

	data := append(ip, portBytes...)
	hash := hmac.New(md5.New, packet.server.PRUDPv1ConnectionSignatureKey)
	hash.Write(data)

	return hash.Sum(nil), nil
}

func defaultPRUDPv1CalculateSignature(packet *PRUDPPacketV1, sessionKey, connectionSignature []byte) []byte {
	accessKeyBytes := []byte(packet.server.AccessKey)
	options := packet.encodeOptions()
	header := packet.encodeHeader()

	accessKeySum := sum[byte, uint32](accessKeyBytes)
	accessKeySumBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(accessKeySumBytes, accessKeySum)

	key := md5.Sum(accessKeyBytes)
	mac := hmac.New(md5.New, key[:])

	if packet.packetType == constants.ConnectPacket && packet.server.PRUDPV1Settings.LegacyConnectionSignature {
		connectionSignature = make([]byte, 0)
	}

	mac.Write(header[4:])
	mac.Write(sessionKey)
	mac.Write(accessKeySumBytes)
	mac.Write(connectionSignature)
	mac.Write(options)
	mac.Write(packet.payload)

	return mac.Sum(nil)
}
