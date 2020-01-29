package nex

import (
	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamOut is an extension of github.com/superwhiskers/crunch with NEX type support
type StreamOut struct {
	*crunch.Buffer
	server *Server
}

// WriteStringNext reads and returns a NEX string type
func (stream *StreamOut) WriteStringNext(str string) {
	str = str + "\x00"

	stream.WriteU16LENext([]uint16{uint16(len(str))})
	stream.WriteBytesNext([]byte(str))
}

// WriteBufferNext writes a NEX Buffer type
func (stream *StreamOut) WriteBufferNext(data []byte) {
	stream.WriteU32LENext([]uint32{uint32(len(data))})
	stream.WriteBytesNext(data)
}

// WriteBuffer writes a NEX Buffer type and does not move the position
func (stream *StreamOut) WriteBuffer(data []byte) {
	stream.WriteU32LENext([]uint32{uint32(len(data))})
	stream.WriteBytes(stream.ByteOffset(), data)
}

func (stream *StreamOut) WriteStructureNext(structure StructureInterface) {
	content := structure.Bytes(NewStreamOut(stream.server))

	if stream.server.GetNexMinorVersion() >= 3 {
		stream.Grow(5)
		stream.WriteByteNext(1) // version
		stream.WriteU32LENext([]uint32{uint32(len(content))})
	}

	stream.Grow(int64(len(content)))
	stream.WriteBytesNext(content)
}

// NewStreamOut returns a new NEX output stream
func NewStreamOut(server *Server) *StreamOut {
	buff := crunch.NewBuffer()

	return &StreamOut{
		Buffer: buff,
		server: server,
	}
}
