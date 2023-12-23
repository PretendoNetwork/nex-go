package types

import (
	"errors"
	"fmt"
)

// ResultRange class which holds information about how to make queries
type ResultRange struct {
	Structure
	Offset uint32 // TODO - Replace this with PrimitiveU32?
	Length uint32 // TODO - Replace this with PrimitiveU32?
}

// WriteTo writes the ResultRange to the given writable
func (rr *ResultRange) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	contentWritable.WritePrimitiveUInt32LE(rr.Offset)
	contentWritable.WritePrimitiveUInt32LE(rr.Length)

	content := contentWritable.Bytes()

	if writable.UseStructureHeader() {
		writable.WritePrimitiveUInt8(rr.StructureVersion())
		writable.WritePrimitiveUInt32LE(uint32(len(content)))
	}

	writable.Write(content)
}

// ExtractFrom extracts the ResultRange to the given readable
func (rr *ResultRange) ExtractFrom(readable Readable) error {
	if readable.UseStructureHeader() {
		version, err := readable.ReadPrimitiveUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read ResultRange version. %s", err.Error())
		}

		contentLength, err := readable.ReadPrimitiveUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read ResultRange content length. %s", err.Error())
		}

		if readable.Remaining() < uint64(contentLength) {
			return errors.New("ResultRange content length longer than data size")
		}

		rr.SetStructureVersion(version)
	}

	offset, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read ResultRange offset. %s", err.Error())
	}

	length, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read ResultRange length. %s", err.Error())
	}

	rr.Offset = offset
	rr.Length = length

	return nil
}

// Copy returns a new copied instance of ResultRange
func (rr *ResultRange) Copy() RVType {
	copied := NewResultRange()

	copied.structureVersion = rr.structureVersion
	copied.Offset = rr.Offset
	copied.Length = rr.Length

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (rr *ResultRange) Equals(o RVType) bool {
	if _, ok := o.(*ResultRange); !ok {
		return false
	}

	other := o.(*ResultRange)

	if rr.structureVersion != other.structureVersion {
		return false
	}

	if rr.Offset != other.Offset {
		return false
	}

	return rr.Length == other.Length
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewResultRange returns a new ResultRange
func NewResultRange() *ResultRange {
	return &ResultRange{}
}
