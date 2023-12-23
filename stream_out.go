package nex

import (
	"github.com/PretendoNetwork/nex-go/types"
	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamOut is an abstraction of github.com/superwhiskers/crunch with nex type support
type StreamOut struct {
	*crunch.Buffer
	Server ServerInterface
}

// StringLengthSize returns the expected size of String length fields
func (s *StreamOut) StringLengthSize() int {
	size := 2

	if s.Server != nil {
		size = s.Server.StringLengthSize()
	}

	return size
}

// PIDSize returns the size of PID types
func (s *StreamOut) PIDSize() int {
	size := 4

	if s.Server != nil && s.Server.LibraryVersion().GreaterOrEqual("4.0.0") {
		size = 8
	}

	return size
}

// UseStructureHeader determines if Structure headers should be used
func (s *StreamOut) UseStructureHeader() bool {
	useStructureHeader := false

	if s.Server != nil {
		switch server := s.Server.(type) {
		case *PRUDPServer: // * Support QRV versions
			useStructureHeader = server.PRUDPMinorVersion >= 3
		default:
			useStructureHeader = server.LibraryVersion().GreaterOrEqual("3.5.0")
		}
	}

	return useStructureHeader
}

// CopyNew returns a copy of the StreamOut but with a blank internal buffer. Returns as types.Writable
func (s *StreamOut) CopyNew() types.Writable {
	return NewStreamOut(s.Server)
}

// Writes the input data to the end of the StreamOut
func (s *StreamOut) Write(data []byte) {
	s.Grow(int64(len(data)))
	s.WriteBytesNext(data)
}

// WritePrimitiveUInt8 writes a uint8
func (s *StreamOut) WritePrimitiveUInt8(u8 uint8) {
	s.Grow(1)
	s.WriteByteNext(byte(u8))
}

// WritePrimitiveUInt16LE writes a uint16 as LE
func (s *StreamOut) WritePrimitiveUInt16LE(u16 uint16) {
	s.Grow(2)
	s.WriteU16LENext([]uint16{u16})
}

// WritePrimitiveUInt32LE writes a uint32 as LE
func (s *StreamOut) WritePrimitiveUInt32LE(u32 uint32) {
	s.Grow(4)
	s.WriteU32LENext([]uint32{u32})
}

// WritePrimitiveUInt64LE writes a uint64 as LE
func (s *StreamOut) WritePrimitiveUInt64LE(u64 uint64) {
	s.Grow(8)
	s.WriteU64LENext([]uint64{u64})
}

// WritePrimitiveInt8 writes a int8
func (s *StreamOut) WritePrimitiveInt8(s8 int8) {
	s.Grow(1)
	s.WriteByteNext(byte(s8))
}

// WritePrimitiveInt16LE writes a uint16 as LE
func (s *StreamOut) WritePrimitiveInt16LE(s16 int16) {
	s.Grow(2)
	s.WriteU16LENext([]uint16{uint16(s16)})
}

// WritePrimitiveInt32LE writes a int32 as LE
func (s *StreamOut) WritePrimitiveInt32LE(s32 int32) {
	s.Grow(4)
	s.WriteU32LENext([]uint32{uint32(s32)})
}

// WritePrimitiveInt64LE writes a int64 as LE
func (s *StreamOut) WritePrimitiveInt64LE(s64 int64) {
	s.Grow(8)
	s.WriteU64LENext([]uint64{uint64(s64)})
}

// WritePrimitiveFloat32LE writes a float32 as LE
func (s *StreamOut) WritePrimitiveFloat32LE(f32 float32) {
	s.Grow(4)
	s.WriteF32LENext([]float32{f32})
}

// WritePrimitiveFloat64LE writes a float64 as LE
func (s *StreamOut) WritePrimitiveFloat64LE(f64 float64) {
	s.Grow(8)
	s.WriteF64LENext([]float64{f64})
}

// WritePrimitiveBool writes a bool
func (s *StreamOut) WritePrimitiveBool(b bool) {
	var bVar uint8
	if b {
		bVar = 1
	}

	s.Grow(1)
	s.WriteByteNext(byte(bVar))
}

// NewStreamOut returns a new nex output stream
func NewStreamOut(server ServerInterface) *StreamOut {
	return &StreamOut{
		Buffer: crunch.NewBuffer(),
		Server: server,
	}
}
