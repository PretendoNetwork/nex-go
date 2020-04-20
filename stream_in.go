package nex

import (
	"errors"
	"strings"

	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamIn is an input stream abstraction of github.com/superwhiskers/crunch with nex type support
type StreamIn struct {
	*crunch.Buffer
	server *Server
}

// ReadUInt8 reads a uint8
func (stream *StreamIn) ReadUInt8() uint8 {
	return uint8(stream.ReadByteNext())
}

// ReadUInt16LE reads a uint16
func (stream *StreamIn) ReadUInt16LE() uint16 {
	return stream.ReadU16LENext(1)[0]
}

// ReadUInt32LE reads a uint32
func (stream *StreamIn) ReadUInt32LE() uint32 {
	return stream.ReadU32LENext(1)[0]
}

// ReadUInt64LE reads a uint64
func (stream *StreamIn) ReadUInt64LE() uint64 {
	return stream.ReadU64LENext(1)[0]
}

// ReadString reads and returns a nex string type
func (stream *StreamIn) ReadString() (string, error) {
	length := stream.ReadUInt16LE()

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return "", errors.New("[StreamIn] Nex string length longer than data size")
	}

	stringData := stream.ReadBytesNext(int64(length))
	str := string(stringData)

	return strings.TrimRight(str, "\x00"), nil
}

// ReadBuffer reads a nex Buffer type
func (stream *StreamIn) ReadBuffer() ([]byte, error) {
	length := stream.ReadUInt32LE()

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return []byte{}, errors.New("[StreamIn] Nex buffer length longer than data size")
	}

	data := stream.ReadBytesNext(int64(length))

	return data, nil
}

// ReadQBuffer reads a nex qBuffer type
func (stream *StreamIn) ReadQBuffer() ([]byte, error) {
	length := stream.ReadUInt16LE()

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return []byte{}, errors.New("[StreamIn] Nex qBuffer length longer than data size")
	}

	data := stream.ReadBytesNext(int64(length))

	return data, nil
}

// ReadStructure reads a nex Structure type
func (stream *StreamIn) ReadStructure(structure StructureInterface) (StructureInterface, error) {
	hierarchy := structure.Hierarchy()

	for _, class := range hierarchy {
		_, err := stream.ReadStructure(class)

		if err != nil {
			return structure, errors.New("[ReadStructure] " + err.Error())
		}
	}

	if stream.server.NexVersion() >= 3 {
		// skip the new struct header as we don't really need the data there
		_ = stream.ReadUInt8()    // structure header version
		_ = stream.ReadUInt32LE() // structure content length
	}

	err := structure.ExtractFromStream(stream)

	if err != nil {
		return structure, errors.New("[ReadStructure] " + err.Error())
	}

	return structure, nil
}

// ReadListUInt8 reads a list of uint8 types
func (stream *StreamIn) ReadListUInt8() []uint8 {
	length := stream.ReadUInt32LE()
	list := make([]uint8, 0, length)

	for i := 0; i < int(length); i++ {
		value := stream.ReadUInt8()
		list = append(list, value)
	}

	return list
}

// ReadListUInt16LE reads a list of uint16 types
func (stream *StreamIn) ReadListUInt16LE() []uint16 {
	length := stream.ReadUInt32LE()
	list := make([]uint16, 0, length)

	for i := 0; i < int(length); i++ {
		value := stream.ReadUInt16LE()
		list = append(list, value)
	}

	return list
}

// ReadListUInt32LE reads a list of uint32 types
func (stream *StreamIn) ReadListUInt32LE() []uint32 {
	length := stream.ReadUInt32LE()
	list := make([]uint32, 0, length)

	for i := 0; i < int(length); i++ {
		value := stream.ReadUInt32LE()
		list = append(list, value)
	}

	return list
}

// ReadListUInt64LE reads a list of uint64 types
func (stream *StreamIn) ReadListUInt64LE() []uint64 {
	length := stream.ReadUInt32LE()
	list := make([]uint64, 0, length)

	for i := 0; i < int(length); i++ {
		value := stream.ReadUInt64LE()
		list = append(list, value)
	}

	return list
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server *Server) *StreamIn {
	return &StreamIn{
		Buffer: crunch.NewBuffer(data),
		server: server,
	}
}
