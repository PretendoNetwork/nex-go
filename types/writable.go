package types

// Writable represents a struct that types can write to
type Writable interface {
	StringLengthSize() int                 // Returns the size of the length field for rdv::String types. Only 2 and 4 are valid
	PIDSize() int                          // Returns the size of the length fields for nn::nex::PID types. Only 4 and 8 are valid
	UseStructureHeader() bool              // Returns whether or not Structure types should use a header
	CopyNew() Writable                     // Returns a new Writable with the same settings, but an empty buffer
	Write(data []byte)                     // Writes the provided data to the buffer
	WritePrimitiveUInt8(value uint8)       // Writes a primitive Go uint8
	WritePrimitiveUInt16LE(value uint16)   // Writes a primitive Go uint16
	WritePrimitiveUInt32LE(value uint32)   // Writes a primitive Go uint32
	WritePrimitiveUInt64LE(value uint64)   // Writes a primitive Go uint64
	WritePrimitiveInt8(value int8)         // Writes a primitive Go int8
	WritePrimitiveInt16LE(value int16)     // Writes a primitive Go int16
	WritePrimitiveInt32LE(value int32)     // Writes a primitive Go int32
	WritePrimitiveInt64LE(value int64)     // Writes a primitive Go int64
	WritePrimitiveFloat32LE(value float32) // Writes a primitive Go float32
	WritePrimitiveFloat64LE(value float64) // Writes a primitive Go float64
	WritePrimitiveBool(value bool)         // Writes a primitive Go bool
	Bytes() []byte                         // Returns the data written t othe buffer
}
