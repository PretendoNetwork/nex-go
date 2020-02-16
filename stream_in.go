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

// ReadStringNext reads and returns a nex string type
func (stream *StreamIn) ReadStringNext() (string, error) {
	length := stream.ReadUInt16LE()

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return "", errors.New("[StreamIn] Nex string length longer than data size")
	}

	stringData := stream.ReadBytesNext(int64(length))
	str := string(stringData)

	return strings.TrimRight(str, "\x00"), nil
}

// ReadBufferNext reads a nex Buffer type
func (stream *StreamIn) ReadBufferNext() ([]byte, error) {
	length := stream.ReadUInt32LE()

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return []byte{}, errors.New("[StreamIn] Nex buffer length longer than data size")
	}

	data := stream.ReadBytesNext(int64(length))

	return data, nil
}

// ReadQBufferNext reads a nex qBuffer type
func (stream *StreamIn) ReadQBufferNext() ([]byte, error) {
	length := stream.ReadUInt16LE()

	if len(stream.Bytes()[stream.ByteOffset():]) < int(length) {
		return []byte{}, errors.New("[StreamIn] Nex qBuffer length longer than data size")
	}

	data := stream.ReadBytesNext(int64(length))

	return data, nil
}

// ReadStructureNext reads a nex Structure type
func (stream *StreamIn) ReadStructureNext(structure StructureInterface) (StructureInterface, error) {
	hierarchy := structure.GetHierarchy()

	for _, class := range hierarchy {
		_, err := stream.ReadStructureNext(class)

		if err != nil {
			return structure, errors.New("[ReadStructureNext] " + err.Error())
		}
	}

	if stream.server.GetNexMinorVersion() >= 3 {
		// skip the new struct header as we don't really need the data there
		_ = stream.ReadUInt8()    // structure header version
		_ = stream.ReadUInt32LE() // structure content length
	}

	err := structure.ExtractFromStream(stream)

	if err != nil {
		return structure, errors.New("[ReadStructureNext] " + err.Error())
	}

	return structure, nil
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server *Server) *StreamIn {
	buff := crunch.NewBuffer(data)

	return &StreamIn{
		Buffer: buff,
		server: server,
	}
}
