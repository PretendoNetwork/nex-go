package nex

import (
	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamOut is an abstraction of github.com/superwhiskers/crunch with nex type support
type StreamOut struct {
	*crunch.Buffer
	server *Server
}

// WriteUInt8 writes a uint8
func (stream *StreamOut) WriteUInt8(u8 uint8) {
	stream.Grow(1)
	stream.WriteByteNext(byte(u8))
}

// WriteUInt16LE writes a uint16 as LE
func (stream *StreamOut) WriteUInt16LE(u16 uint16) {
	stream.Grow(2)
	stream.WriteU16LENext([]uint16{u16})
}

// WriteUInt32LE writes a uint32 as LE
func (stream *StreamOut) WriteUInt32LE(u32 uint32) {
	stream.Grow(4)
	stream.WriteU32LENext([]uint32{u32})
}

// WriteUInt64LE writes a uint64 as LE
func (stream *StreamOut) WriteUInt64LE(u64 uint64) {
	stream.Grow(8)
	stream.WriteU64LENext([]uint64{u64})
}

// WriteString writes a NEX string type
func (stream *StreamOut) WriteString(str string) {
	str = str + "\x00"
	strLength := len(str)

	stream.Grow(int64(strLength))
	stream.WriteUInt16LE(uint16(strLength))
	stream.WriteBytesNext([]byte(str))
}

// WriteBuffer writes a NEX Buffer type
func (stream *StreamOut) WriteBuffer(data []byte) {
	dataLength := len(data)

	stream.WriteUInt32LE(uint32(dataLength))
	stream.Grow(int64(dataLength))
	stream.WriteBytesNext(data)
}

// WriteStructure writes a nex Structure type
func (stream *StreamOut) WriteStructure(structure StructureInterface) {
	content := structure.Bytes(NewStreamOut(stream.server))

	if stream.server.GetNexMinorVersion() >= 3 {
		stream.WriteUInt8(1) // version
		stream.WriteUInt32LE(uint32(len(content)))
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
