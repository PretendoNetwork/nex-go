package types

import (
	"fmt"
)

// ResultRange class which holds information about how to make queries
type ResultRange struct {
	Structure
	Offset *PrimitiveU32
	Length *PrimitiveU32
}

// WriteTo writes the ResultRange to the given writable
func (rr *ResultRange) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	rr.Offset.WriteTo(contentWritable)
	rr.Length.WriteTo(contentWritable)

	content := contentWritable.Bytes()

	rr.WriteHeaderTo(writable, uint32(len(content)))

	writable.Write(content)
}

// ExtractFrom extracts the ResultRange from the given readable
func (rr *ResultRange) ExtractFrom(readable Readable) error {
	var err error

	if err = rr.ExtractHeaderFrom(readable); err != nil {
		return fmt.Errorf("Failed to read ResultRange header. %s", err.Error())
	}

	err = rr.Offset.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read ResultRange.Offset. %s", err.Error())
	}

	err = rr.Length.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read ResultRange.Length. %s", err.Error())
	}

	return nil
}

// Copy returns a new copied instance of ResultRange
func (rr *ResultRange) Copy() RVType {
	copied := NewResultRange()

	copied.StructureVersion = rr.StructureVersion
	copied.Offset = rr.Offset.Copy().(*PrimitiveU32)
	copied.Length = rr.Length.Copy().(*PrimitiveU32)

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (rr *ResultRange) Equals(o RVType) bool {
	if _, ok := o.(*ResultRange); !ok {
		return false
	}

	other := o.(*ResultRange)

	if rr.StructureVersion != other.StructureVersion {
		return false
	}

	if !rr.Offset.Equals(other.Offset) {
		return false
	}

	return rr.Length.Equals(other.Length)
}

// NewResultRange returns a new ResultRange
func NewResultRange() *ResultRange {
	return &ResultRange{
		Offset: NewPrimitiveU32(0),
		Length: NewPrimitiveU32(0),
	}
}
