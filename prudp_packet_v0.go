package nex

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"slices"
)

// PRUDPPacketV0 represents a PRUDPv0 packet
type PRUDPPacketV0 struct {
	PRUDPPacket
}

// Copy copies the packet into a new PRUDPPacketV0
//
// Retains the same PRUDPClient pointer
func (p *PRUDPPacketV0) Copy() PRUDPPacketInterface {
	copied, _ := NewPRUDPPacketV0(p.sender, nil)

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

	return copied
}

// Version returns the packets PRUDP version
func (p *PRUDPPacketV0) Version() int {
	return 0
}

func (p *PRUDPPacketV0) decode() error {
	// * Header is technically 11 bytes but checking for 12 includes the checksum
	if p.readStream.Remaining() < 12 {
		return errors.New("Failed to read PRUDPv0 header. Not have enough data")
	}

	server := p.sender.server
	start := p.readStream.ByteOffset()

	source, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 source. %s", err.Error())
	}

	destination, err := p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 destination. %s", err.Error())
	}

	p.sourceStreamType = source >> 4
	p.sourcePort = source & 0xF
	p.destinationStreamType = destination >> 4
	p.destinationPort = destination & 0xF

	if server.IsQuazalMode {
		typeAndFlags, err := p.readStream.ReadUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv0 type and flags. %s", err.Error())
		}

		p.flags = uint16(typeAndFlags >> 3)
		p.packetType = uint16(typeAndFlags & 7)
	} else {
		typeAndFlags, err := p.readStream.ReadUInt16LE()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv0 type and flags. %s", err.Error())
		}

		p.flags = typeAndFlags >> 4
		p.packetType = typeAndFlags & 0xF
	}

	if _, ok := validPacketTypes[p.packetType]; !ok {
		return errors.New("Invalid PRUDPv0 packet type")
	}

	p.sessionID, err = p.readStream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 session ID. %s", err.Error())
	}

	p.signature = p.readStream.ReadBytesNext(4)

	p.sequenceID, err = p.readStream.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 sequence ID. %s", err.Error())
	}

	if p.packetType == SynPacket || p.packetType == ConnectPacket {
		if p.readStream.Remaining() < 4 {
			return errors.New("Failed to read PRUDPv0 connection signature. Not have enough data")
		}

		p.connectionSignature = p.readStream.ReadBytesNext(4)
	}

	if p.packetType == DataPacket {
		if p.readStream.Remaining() < 1 {
			return errors.New("Failed to read PRUDPv0 fragment ID. Not have enough data")
		}

		p.fragmentID, err = p.readStream.ReadUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv0 fragment ID. %s", err.Error())
		}
	}

	var payloadSize uint16

	if p.HasFlag(FlagHasSize) {
		if p.readStream.Remaining() < 2 {
			return errors.New("Failed to read PRUDPv0 payload size. Not have enough data")
		}

		payloadSize, err = p.readStream.ReadUInt16LE()
		if err != nil {
			return fmt.Errorf("Failed to read PRUDPv0 payload size. %s", err.Error())
		}
	} else {
		// * Quazal used a 4 byte checksum. NEX uses 1 byte
		if server.IsQuazalMode {
			payloadSize = uint16(p.readStream.Remaining() - 4)
		} else {
			payloadSize = uint16(p.readStream.Remaining() - 1)
		}
	}

	if p.readStream.Remaining() < int(payloadSize) {
		return errors.New("Failed to read PRUDPv0 payload. Not have enough data")
	}

	p.payload = p.readStream.ReadBytesNext(int64(payloadSize))

	if server.IsQuazalMode && p.readStream.Remaining() < 4 {
		return errors.New("Failed to read PRUDPv0 checksum. Not have enough data")
	} else if p.readStream.Remaining() < 1 {
		return errors.New("Failed to read PRUDPv0 checksum. Not have enough data")
	}

	checksumData := p.readStream.Bytes()[start:p.readStream.ByteOffset()]

	var checksum uint32
	var checksumU8 uint8

	if server.IsQuazalMode {
		checksum, err = p.readStream.ReadUInt32LE()
	} else {
		checksumU8, err = p.readStream.ReadUInt8()
		checksum = uint32(checksumU8)
	}

	if err != nil {
		return fmt.Errorf("Failed to read PRUDPv0 checksum. %s", err.Error())
	}

	calculatedChecksum := p.calculateChecksum(checksumData)

	if checksum != calculatedChecksum {
		return errors.New("Invalid PRUDPv0 checksum")
	}

	return nil
}

// Bytes encodes a PRUDPv0 packet into a byte slice
func (p *PRUDPPacketV0) Bytes() []byte {
	server := p.sender.server
	stream := NewStreamOut(server)

	stream.WriteUInt8(p.sourcePort | (p.sourceStreamType << 4))
	stream.WriteUInt8(p.destinationPort | (p.destinationStreamType << 4))

	if server.IsQuazalMode {
		stream.WriteUInt8(uint8(p.packetType | (p.flags << 3)))
	} else {
		stream.WriteUInt16LE(p.packetType | (p.flags << 4))
	}

	stream.WriteUInt8(p.sessionID)
	stream.Grow(int64(len(p.signature)))
	stream.WriteBytesNext(p.signature)
	stream.WriteUInt16LE(p.sequenceID)

	if p.packetType == SynPacket || p.packetType == ConnectPacket {
		stream.Grow(int64(len(p.connectionSignature)))
		stream.WriteBytesNext(p.connectionSignature)
	}

	if p.packetType == DataPacket {
		stream.WriteUInt8(p.fragmentID)
	}

	if p.HasFlag(FlagHasSize) {
		stream.WriteUInt16LE(uint16(len(p.payload)))
	}

	if len(p.payload) > 0 {
		stream.Grow(int64(len(p.payload)))
		stream.WriteBytesNext(p.payload)
	}

	checksum := p.calculateChecksum(stream.Bytes())

	if server.IsQuazalMode {
		stream.WriteUInt32LE(checksum)
	} else {
		stream.WriteUInt8(uint8(checksum))
	}

	return stream.Bytes()
}

func (p *PRUDPPacketV0) calculateConnectionSignature(addr net.Addr) ([]byte, error) {
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
	hash := md5.Sum(data)
	signatureBytes := hash[:4]

	slices.Reverse(signatureBytes)

	return signatureBytes, nil
}

func (p *PRUDPPacketV0) calculateSignature(sessionKey, connectionSignature []byte) []byte {
	if p.packetType == DataPacket {
		return p.calculateDataSignature(sessionKey)
	}

	if p.packetType == DisconnectPacket && p.sender.server.accessKey != "ridfebb9" {
		return p.calculateDataSignature(sessionKey)
	}

	if len(connectionSignature) != 0 {
		return connectionSignature
	}

	return make([]byte, 4)
}

func (p *PRUDPPacketV0) calculateDataSignature(sessionKey []byte) []byte {
	server := p.sender.server
	data := p.payload

	if server.AccessKey() != "ridfebb9" {
		header := []byte{0, 0, p.fragmentID}
		binary.LittleEndian.PutUint16(header[:2], p.sequenceID)

		data = append(sessionKey, header...)
		data = append(data, p.payload...)
	}

	if len(data) > 0 {
		key := md5.Sum([]byte(server.AccessKey()))
		mac := hmac.New(md5.New, key[:])

		mac.Write(data)

		digest := mac.Sum(nil)

		return digest[:4]
	}

	return []byte{0x78, 0x56, 0x34, 0x12}
}

func (p *PRUDPPacketV0) calculateChecksum(data []byte) uint32 {
	server := p.sender.server
	checksum := sum[byte, uint32]([]byte(server.AccessKey()))

	if server.IsQuazalMode {
		padSize := (len(data) + 3) &^ 3
		data = append(data, make([]byte, padSize-len(data))...)
		words := make([]uint32, len(data)/4)

		for i := 0; i < len(data)/4; i++ {
			words[i] = binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
		}

		result := (checksum & 0xFF) + sum[uint32, uint32](words)

		return result & 0xFFFFFFFF
	} else {
		words := make([]uint32, len(data)/4)

		for i := 0; i < len(data)/4; i++ {
			words[i] = binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
		}

		temp := sum[uint32, uint32](words) & 0xFFFFFFFF

		checksum += sum[byte, uint32](data[len(data)&^3:])

		tempBytes := make([]byte, 4)

		binary.LittleEndian.PutUint32(tempBytes, temp)

		checksum += sum[byte, uint32](tempBytes)

		return checksum & 0xFF
	}
}

// NewPRUDPPacketV0 creates and returns a new PacketV0 using the provided Client and stream
func NewPRUDPPacketV0(client *PRUDPClient, readStream *StreamIn) (*PRUDPPacketV0, error) {
	packet := &PRUDPPacketV0{
		PRUDPPacket: PRUDPPacket{
			sender:     client,
			readStream: readStream,
		},
	}

	if readStream != nil {
		err := packet.decode()
		if err != nil {
			return nil, fmt.Errorf("Failed to decode PRUDPv0 packet. %s", err.Error())
		}
	}

	return packet, nil
}

// NewPRUDPPacketsV0 reads all possible PRUDPv0 packets from the stream
func NewPRUDPPacketsV0(client *PRUDPClient, readStream *StreamIn) ([]PRUDPPacketInterface, error) {
	packets := make([]PRUDPPacketInterface, 0)

	for readStream.Remaining() > 0 {
		packet, err := NewPRUDPPacketV0(client, readStream)
		if err != nil {
			return packets, err
		}

		packets = append(packets, packet)
	}

	return packets, nil
}
