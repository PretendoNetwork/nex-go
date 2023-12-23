package types

// Writable represents a struct that types can write to
type Writable interface {
	StringLengthSize() int
	PIDSize() int
	UseStructureHeader() bool
	CopyNew() Writable
	Write(data []byte)
	WritePrimitiveUInt8(value uint8)
	WritePrimitiveUInt16LE(value uint16)
	WritePrimitiveUInt32LE(value uint32)
	WritePrimitiveUInt64LE(value uint64)
	WritePrimitiveInt8(value int8)
	WritePrimitiveInt16LE(value int16)
	WritePrimitiveInt32LE(value int32)
	WritePrimitiveInt64LE(value int64)
	WritePrimitiveFloat32LE(value float32)
	WritePrimitiveFloat64LE(value float64)
	WritePrimitiveBool(value bool)
	Bytes() []byte
}
