package nex

import (
	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamOut is an abstraction of github.com/superwhiskers/crunch with nex type support
type StreamOut struct {
	*crunch.Buffer
	server *Server
}

// WriteStringNext reads and returns a nex string type
func (stream *StreamOut) WriteStringNext(str string) {
	str = str + "\x00"

	stream.WriteU16LENext([]uint16{uint16(len(str))})
	stream.WriteBytesNext([]byte(str))
}

// WriteBufferNext writes a nex Buffer type
func (stream *StreamOut) WriteBufferNext(data []byte) {
	stream.WriteU32LENext([]uint32{uint32(len(data))})
	stream.WriteBytesNext(data)
}

// WriteBuffer writes a nex Buffer type and does not move the position
func (stream *StreamOut) WriteBuffer(data []byte) {
	stream.WriteU32LENext([]uint32{uint32(len(data))})
	stream.WriteBytes(stream.ByteOffset(), data)
}

// WriteStructureNext writes a nex Structure type
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

// NewStreamOut returns a new nex output stream
func NewStreamOut(server *Server) *StreamOut {
	buff := crunch.NewBuffer()

	return &StreamOut{
		Buffer: buff,
		server: server,
	}
}
