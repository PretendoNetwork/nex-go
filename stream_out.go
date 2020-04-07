package nex

import crunch "github.com/superwhiskers/crunch/v3"

// StreamOut is an abstraction of github.com/superwhiskers/crunch with nex type support
type StreamOut struct {
	*crunch.Buffer
	server *Server
}

// NewStreamOut returns a new nex output stream
func NewStreamOut(server *Server) *StreamOut {
	return &StreamOut{
		Buffer: crunch.NewBuffer(),
		server: server,
	}
}

// WriteUInt8 writes a uint8
func (stream *StreamOut) WriteUInt8(u8 uint8) {
	stream.Grow(1)
	stream.WriteByteNext(u8)
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
	strLen := len(str) + 1

	stream.Grow(int64(strLen) + 2) // account for the additional size of the string length
	stream.WriteU16LENext([]uint16{uint16(strLen)})
	stream.WriteBytesNext([]byte(str))
	stream.WriteByteNext(0x00)
}

// WriteBuffer writes a NEX Buffer type
func (stream *StreamOut) WriteBuffer(data []byte) {
	dataLength := len(data)

	stream.Grow(int64(dataLength) + 4) // account for bytebuf length
	stream.WriteU32LENext([]uint32{uint32(dataLength)})
	stream.WriteBytesNext(data)
}

// WriteStructure writes a nex Structure type
func (stream *StreamOut) WriteStructure(structure StructureInterface) {
	content := structure.Bytes(NewStreamOut(stream.server))

	if stream.server.GetNexMinorVersion() >= 3 {
		stream.Grow(5)
		stream.WriteByteNext(1) // version
		stream.WriteU32LENext([]uint32{uint32(len(content))})
	}

	stream.Grow(int64(len(content)))
	stream.WriteBytesNext(content)
}

// WriteListStructure writes a list of Structure types
func (stream *StreamOut) WriteListStructure(structures []StructureInterface) {
	stream.Grow(4)
	stream.WriteU32LENext([]uint32{uint32(len(structures))})
	for i := 0; i < len(structures); i++ {
		stream.WriteStructure(structures[i])
	}
}

// WriteListUInt8 writes a list of uint8 types
func (stream *StreamOut) WriteListUInt8(list []uint8) {
	stream.Grow(int64(len(list) + 4))

	stream.WriteU32LENext([]uint32{uint32(len(list))})
	stream.WriteBytesNext(list)
}

// WriteListUInt16LE writes a list of uint16 types
func (stream *StreamOut) WriteListUInt16LE(list []uint16) {
	stream.Grow(int64((len(list) * 2) + 4))

	stream.WriteU32LENext([]uint32{uint32(len(list))})
	stream.WriteU16LENext(list)
}

// WriteListUInt32LE writes a list of uint32 types
func (stream *StreamOut) WriteListUInt32LE(list []uint32) {
	stream.Grow(int64((len(list) * 4) + 4))

	stream.WriteU32LENext([]uint32{uint32(len(list))})
	stream.WriteU32LENext(list)
}

// WriteListUInt64LE writes a list of uint64 types
func (stream *StreamOut) WriteListUInt64LE(list []uint64) {
	stream.Grow(int64((len(list) * 8) + 4))

	stream.WriteU32LENext([]uint32{uint32(len(list))})
	stream.WriteU64LENext(list)
}
