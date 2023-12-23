package types

// Readable represents a struct that types can read from
type Readable interface {
	StringLengthSize() int
	PIDSize() int
	UseStructureHeader() bool
	Remaining() uint64
	ReadRemaining() []byte
	Read(length uint64) ([]byte, error)
	ReadPrimitiveUInt8() (uint8, error)
	ReadPrimitiveUInt16LE() (uint16, error)
	ReadPrimitiveUInt32LE() (uint32, error)
	ReadPrimitiveUInt64LE() (uint64, error)
	ReadPrimitiveInt8() (int8, error)
	ReadPrimitiveInt16LE() (int16, error)
	ReadPrimitiveInt32LE() (int32, error)
	ReadPrimitiveInt64LE() (int64, error)
	ReadPrimitiveFloat32LE() (float32, error)
	ReadPrimitiveFloat64LE() (float64, error)
	ReadPrimitiveBool() (bool, error)
}
