package types

// Readable represents a struct that types can read from
type Readable interface {
	StringLengthSize() int              // Returns the size of the length field for rdv::String types. Only 2 and 4 are valid
	PIDSize() int                       // Returns the size of the length fields for nn::nex::PID types. Only 4 and 8 are valid
	UseStructureHeader() bool           // Returns whether or not Structure types should use a header
	Remaining() uint64                  // Returns the number of bytes left unread in the buffer
	ReadRemaining() []byte              // Reads the remaining data from the buffer
	Read(length uint64) ([]byte, error) // Reads up to length bytes of data from the buffer. Returns an error if the read failed, such as if there was not enough data to read
	ReadUInt8() (uint8, error)          // Reads a primitive Go uint8. Returns an error if the read failed, such as if there was not enough data to read
	ReadUInt16LE() (uint16, error)      // Reads a primitive Go uint16. Returns an error if the read failed, such as if there was not enough data to read
	ReadUInt32LE() (uint32, error)      // Reads a primitive Go uint32. Returns an error if the read failed, such as if there was not enough data to read
	ReadUInt64LE() (uint64, error)      // Reads a primitive Go uint64. Returns an error if the read failed, such as if there was not enough data to read
	ReadInt8() (int8, error)            // Reads a primitive Go int8. Returns an error if the read failed, such as if there was not enough data to read
	ReadInt16LE() (int16, error)        // Reads a primitive Go int16. Returns an error if the read failed, such as if there was not enough data to read
	ReadInt32LE() (int32, error)        // Reads a primitive Go int32. Returns an error if the read failed, such as if there was not enough data to read
	ReadInt64LE() (int64, error)        // Reads a primitive Go int64. Returns an error if the read failed, such as if there was not enough data to read
	ReadFloat32LE() (float32, error)    // Reads a primitive Go float32. Returns an error if the read failed, such as if there was not enough data to read
	ReadFloat64LE() (float64, error)    // Reads a primitive Go float64. Returns an error if the read failed, such as if there was not enough data to read
	ReadBool() (bool, error)            // Reads a primitive Go bool. Returns an error if the read failed, such as if there was not enough data to read
}
