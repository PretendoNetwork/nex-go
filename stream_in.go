package nex

import (
	"errors"

	crunch "github.com/superwhiskers/crunch/v3"
)

// StreamIn is an input stream abstraction of github.com/superwhiskers/crunch/v3 with nex type support
type StreamIn struct {
	*crunch.Buffer
	Server ServerInterface
}

// StringLengthSize returns the expected size of String length fields
func (s *StreamIn) StringLengthSize() int {
	size := 2

	if s.Server != nil {
		size = s.Server.StringLengthSize()
	}

	return size
}

// PIDSize returns the size of PID types
func (s *StreamIn) PIDSize() int {
	size := 4

	if s.Server != nil && s.Server.LibraryVersion().GreaterOrEqual("4.0.0") {
		size = 8
	}

	return size
}

// UseStructureHeader determines if Structure headers should be used
func (s *StreamIn) UseStructureHeader() bool {
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

// Remaining returns the amount of data left to be read in the buffer
func (s *StreamIn) Remaining() uint64 {
	return uint64(len(s.Bytes()[s.ByteOffset():]))
}

// ReadRemaining reads all the data left to be read in the buffer
func (s *StreamIn) ReadRemaining() []byte {
	// * Can safely ignore this error, since s.Remaining() will never be less than itself
	remaining, _ := s.Read(uint64(s.Remaining()))

	return remaining
}

// Read reads the specified number of bytes. Returns an error if OOB
func (s *StreamIn) Read(length uint64) ([]byte, error) {
	if s.Remaining() < length {
		return []byte{}, errors.New("Read is OOB")
	}

	return s.ReadBytesNext(int64(length)), nil
}

// ReadPrimitiveUInt8 reads a uint8
func (s *StreamIn) ReadPrimitiveUInt8() (uint8, error) {
	if s.Remaining() < 1 {
		return 0, errors.New("Not enough data to read uint8")
	}

	return uint8(s.ReadByteNext()), nil
}

// ReadPrimitiveUInt16LE reads a Little-Endian encoded uint16
func (s *StreamIn) ReadPrimitiveUInt16LE() (uint16, error) {
	if s.Remaining() < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return s.ReadU16LENext(1)[0], nil
}

// ReadPrimitiveUInt32LE reads a Little-Endian encoded uint32
func (s *StreamIn) ReadPrimitiveUInt32LE() (uint32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return s.ReadU32LENext(1)[0], nil
}

// ReadPrimitiveUInt64LE reads a Little-Endian encoded uint64
func (s *StreamIn) ReadPrimitiveUInt64LE() (uint64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return s.ReadU64LENext(1)[0], nil
}

// ReadPrimitiveInt8 reads a uint8
func (s *StreamIn) ReadPrimitiveInt8() (int8, error) {
	if s.Remaining() < 1 {
		return 0, errors.New("Not enough data to read int8")
	}

	return int8(s.ReadByteNext()), nil
}

// ReadPrimitiveInt16LE reads a Little-Endian encoded int16
func (s *StreamIn) ReadPrimitiveInt16LE() (int16, error) {
	if s.Remaining() < 2 {
		return 0, errors.New("Not enough data to read int16")
	}

	return int16(s.ReadU16LENext(1)[0]), nil
}

// ReadPrimitiveInt32LE reads a Little-Endian encoded int32
func (s *StreamIn) ReadPrimitiveInt32LE() (int32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(s.ReadU32LENext(1)[0]), nil
}

// ReadPrimitiveInt64LE reads a Little-Endian encoded int64
func (s *StreamIn) ReadPrimitiveInt64LE() (int64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(s.ReadU64LENext(1)[0]), nil
}

// ReadPrimitiveFloat32LE reads a Little-Endian encoded float32
func (s *StreamIn) ReadPrimitiveFloat32LE() (float32, error) {
	if s.Remaining() < 4 {
		return 0, errors.New("Not enough data to read float32")
	}

	return s.ReadF32LENext(1)[0], nil
}

// ReadPrimitiveFloat64LE reads a Little-Endian encoded float64
func (s *StreamIn) ReadPrimitiveFloat64LE() (float64, error) {
	if s.Remaining() < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return s.ReadF64LENext(1)[0], nil
}

// ReadPrimitiveBool reads a bool
func (s *StreamIn) ReadPrimitiveBool() (bool, error) {
	if s.Remaining() < 1 {
		return false, errors.New("Not enough data to read bool")
	}

	return s.ReadByteNext() == 1, nil
}

// NewStreamIn returns a new NEX input stream
func NewStreamIn(data []byte, server ServerInterface) *StreamIn {
	return &StreamIn{
		Buffer: crunch.NewBuffer(data),
		Server: server,
	}
}
