package nex

import (
	"errors"

	crunch "github.com/superwhiskers/crunch/v3"
)

// ByteStreamIn is an input stream abstraction of github.com/superwhiskers/crunch/v3 with nex type support
type ByteStreamIn struct {
	*crunch.Buffer
	LibraryVersions *LibraryVersions
	Settings        *ByteStreamSettings
}

// StringLengthSize returns the expected size of String length fields
func (bsi *ByteStreamIn) StringLengthSize() int {
	size := 2

	if bsi.Settings != nil {
		size = bsi.Settings.StringLengthSize
	}

	return size
}

// PIDSize returns the size of PID types
func (bsi *ByteStreamIn) PIDSize() int {
	size := 4

	if bsi.Settings != nil {
		size = bsi.Settings.PIDSize
	}

	return size
}

// UseStructureHeader determines if Structure headers should be used
func (bsi *ByteStreamIn) UseStructureHeader() bool {
	useStructureHeader := false

	if bsi.Settings != nil {
		useStructureHeader = bsi.Settings.UseStructureHeader
	}

	return useStructureHeader
}

// Remaining returns the amount of data left to be read in the buffer
func (bsi *ByteStreamIn) Remaining() uint64 {
	return uint64(len(bsi.Bytes()[bsi.ByteOffset():]))
}

// ReadRemaining reads all the data left to be read in the buffer
func (bsi *ByteStreamIn) ReadRemaining() []byte {
	// * Can safely ignore this error, since bsi.Remaining() will never be less than itself
	remaining, _ := bsi.Read(uint64(bsi.Remaining()))

	return remaining
}

// Read reads the specified number of bytes. Returns an error if OOB
func (bsi *ByteStreamIn) Read(length uint64) ([]byte, error) {
	if bsi.Remaining() < length {
		return []byte{}, errors.New("Read is OOB")
	}

	return bsi.ReadBytesNext(int64(length)), nil
}

// ReadUInt8 reads a uint8
func (bsi *ByteStreamIn) ReadUInt8() (uint8, error) {
	if bsi.Remaining() < 1 {
		return 0, errors.New("Not enough data to read uint8")
	}

	return uint8(bsi.ReadByteNext()), nil
}

// ReadUInt16LE reads a Little-Endian encoded uint16
func (bsi *ByteStreamIn) ReadUInt16LE() (uint16, error) {
	if bsi.Remaining() < 2 {
		return 0, errors.New("Not enough data to read uint16")
	}

	return bsi.ReadU16LENext(1)[0], nil
}

// ReadUInt32LE reads a Little-Endian encoded uint32
func (bsi *ByteStreamIn) ReadUInt32LE() (uint32, error) {
	if bsi.Remaining() < 4 {
		return 0, errors.New("Not enough data to read uint32")
	}

	return bsi.ReadU32LENext(1)[0], nil
}

// ReadUInt64LE reads a Little-Endian encoded uint64
func (bsi *ByteStreamIn) ReadUInt64LE() (uint64, error) {
	if bsi.Remaining() < 8 {
		return 0, errors.New("Not enough data to read uint64")
	}

	return bsi.ReadU64LENext(1)[0], nil
}

// ReadInt8 reads a uint8
func (bsi *ByteStreamIn) ReadInt8() (int8, error) {
	if bsi.Remaining() < 1 {
		return 0, errors.New("Not enough data to read int8")
	}

	return int8(bsi.ReadByteNext()), nil
}

// ReadInt16LE reads a Little-Endian encoded int16
func (bsi *ByteStreamIn) ReadInt16LE() (int16, error) {
	if bsi.Remaining() < 2 {
		return 0, errors.New("Not enough data to read int16")
	}

	return int16(bsi.ReadU16LENext(1)[0]), nil
}

// ReadInt32LE reads a Little-Endian encoded int32
func (bsi *ByteStreamIn) ReadInt32LE() (int32, error) {
	if bsi.Remaining() < 4 {
		return 0, errors.New("Not enough data to read int32")
	}

	return int32(bsi.ReadU32LENext(1)[0]), nil
}

// ReadInt64LE reads a Little-Endian encoded int64
func (bsi *ByteStreamIn) ReadInt64LE() (int64, error) {
	if bsi.Remaining() < 8 {
		return 0, errors.New("Not enough data to read int64")
	}

	return int64(bsi.ReadU64LENext(1)[0]), nil
}

// ReadFloat32LE reads a Little-Endian encoded float32
func (bsi *ByteStreamIn) ReadFloat32LE() (float32, error) {
	if bsi.Remaining() < 4 {
		return 0, errors.New("Not enough data to read float32")
	}

	return bsi.ReadF32LENext(1)[0], nil
}

// ReadFloat64LE reads a Little-Endian encoded float64
func (bsi *ByteStreamIn) ReadFloat64LE() (float64, error) {
	if bsi.Remaining() < 8 {
		return 0, errors.New("Not enough data to read float64")
	}

	return bsi.ReadF64LENext(1)[0], nil
}

// ReadBool reads a bool
func (bsi *ByteStreamIn) ReadBool() (bool, error) {
	if bsi.Remaining() < 1 {
		return false, errors.New("Not enough data to read bool")
	}

	return bsi.ReadByteNext() == 1, nil
}

// NewByteStreamIn returns a new NEX input byte stream
func NewByteStreamIn(data []byte, libraryVersions *LibraryVersions, settings *ByteStreamSettings) *ByteStreamIn {
	return &ByteStreamIn{
		Buffer:          crunch.NewBuffer(data),
		LibraryVersions: libraryVersions,
		Settings:        settings,
	}
}
