package nex

import (
	"errors"

	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamIn is an input stream abstraction of github.com/superwhiskers/crunch with nex type support
type StreamIn struct {
	*crunch.Buffer
	server *Server
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server *Server) *StreamIn {
	return &StreamIn{
		Buffer: crunch.NewBuffer(data),
		server: server,
	}
}

// ReadUInt8 reads a uint8
func (stream *StreamIn) ReadUInt8() uint8 {
	return stream.ReadByteNext()
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
func (stream *StreamIn) ReadString() (data string, err error) {
	length := stream.ReadU16LENext(1)[0]

	if (stream.ByteCapacity() - stream.ByteOffset()) < int64(length) {
		err = errors.New("[StreamIn] Nex string length longer than data size")
	}

	data = string(stream.ReadBytesNext(int64(length))[0:length])
	return
}

// ReadBuffer reads a nex Buffer type
func (stream *StreamIn) ReadBuffer() (data []byte, err error) {
	length := stream.ReadU32LENext(1)[0]

	if (stream.ByteCapacity() - stream.ByteOffset()) < int64(length) {
		err = errors.New("[StreamIn] Nex buffer length longer than data size")
		return
	}

	data = stream.ReadBytesNext(int64(length))
	return
}

// ReadQBuffer reads a nex qBuffer type
func (stream *StreamIn) ReadQBuffer() (data []byte, err error) {
	length := stream.ReadU16LENext(1)[0]

	if (stream.ByteCapacity() - stream.ByteOffset()) < int64(length) {
		err = errors.New("[StreamIn] Nex qBuffer length longer than data size")
	}

	data = stream.ReadBytesNext(int64(length))
	return
}

// ReadStructure reads a nex Structure type
func (stream *StreamIn) ReadStructure(structure StructureInterface) (data StructureInterface, err error) {
	hierarchy := structure.GetHierarchy()
	data = structure // for some reason, go won't let you pipe data like that

	for _, class := range hierarchy {
		_, err = stream.ReadStructure(class)

		if err != nil {
			err = errors.New("[ReadStructure] " + err.Error())

		}
	}

	if stream.server.GetNexMinorVersion() >= 3 {
		// skip the new struct header as we don't really need the data there
		/* superwhiskers: this was here before but i replaced it with a seek
		_ = stream.ReadUInt8()    // structure header version
		_ = stream.ReadUInt32LE() // structure content length
		*/
		stream.SeekByte(5, true)
	}

	err = data.ExtractFromStream(stream)
	if err != nil {
		err = errors.New("[ReadStructure] " + err.Error())
	}

	return
}

// ReadListUInt8 reads a list of uint8 types
func (stream *StreamIn) ReadListUInt8() []uint8 {
	return stream.ReadBytesNext(int64(stream.ReadU32LENext(1)[0]))
}

// ReadListUInt16LE reads a list of uint16 types
func (stream *StreamIn) ReadListUInt16LE() []uint16 {
	return stream.ReadU16LENext(int64(stream.ReadU32LENext(1)[0]))
}

// ReadListUInt32LE reads a list of uint32 types
func (stream *StreamIn) ReadListUInt32LE() []uint32 {
	return stream.ReadU32LENext(int64(stream.ReadU32LENext(1)[0]))
}

// ReadListUInt64LE reads a list of uint64 types
func (stream *StreamIn) ReadListUInt64LE() []uint64 {
	return stream.ReadU64LENext(int64(stream.ReadU32LENext(1)[0]))
}
