package nex

import (
	"strings"

	"github.com/superwhiskers/crunch"
)

// Stream is an extension of github.com/superwhiskers/crunch with NEX type support
type Stream struct {
	*crunch.Buffer
}

// ReadNEXStringNext reads and returns a NEX string type
func (stream *Stream) ReadNEXStringNext() string {
	length := stream.ReadU16LENext(1)[0]
	stringData := stream.ReadBytesNext(int64(length))
	str := string(stringData)

	return strings.TrimRight(str, "\x00")
}

// WriteNEXStringNext reads and returns a NEX string type
func (stream *Stream) WriteNEXStringNext(str string) {
	str = str + "\x00"

	stream.WriteU16LENext([]uint16{uint16(len(str))})
	stream.WriteBytesNext([]byte(str))
}

// WriteNEXBufferNext writes a NEX Buffer type
func (stream *Stream) WriteNEXBufferNext(data []byte) {
	stream.WriteU32LENext([]uint32{uint32(len(data))})
	stream.WriteBytesNext(data)
}

// ReadNEXBufferNext reads a NEX Buffer type
func (stream *Stream) ReadNEXBufferNext() []byte {
	length := stream.ReadU32LENext(1)[0]
	data := stream.ReadBytesNext(int64(length))

	return data
}

// WriteNEXBuffer writes a NEX Buffer type and does not move the position
func (stream *Stream) WriteNEXBuffer(data []byte) {
	stream.WriteU32LENext([]uint32{uint32(len(data))})
	stream.WriteBytes(stream.ByteOffset(), data)
}

// ReadNEXBufferNext reads a NEX qBuffer type
func (stream *Stream) ReadNEXQBufferNext() []byte {
	length := stream.ReadU16LENext(1)[0]
	data := stream.ReadBytesNext(int64(length))

	return data
}

// NewStream returns a new NEX stream
func NewStream(data ...[]byte) *Stream {
	var buff *crunch.Buffer

	if len(data) > 0 {
		buff = crunch.NewBuffer(data[0])
	} else {
		buff = crunch.NewBuffer()
	}

	return &Stream{
		Buffer: buff,
	}
}
