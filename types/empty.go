package types

import (
	"errors"
	"fmt"
	"strings"
)

// Empty is a Structure with no fields
type Empty struct {
	Structure
}

// WriteTo writes the Empty to the given writable
func (e *Empty) WriteTo(writable Writable) {
	if writable.UseStructureHeader() {
		writable.WritePrimitiveUInt8(e.StructureVersion())
		writable.WritePrimitiveUInt32LE(0)
	}
}

// ExtractFrom extracts the Empty to the given readable
func (e *Empty) ExtractFrom(readable Readable) error {
	if readable.UseStructureHeader() {
		version, err := readable.ReadPrimitiveUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read Empty version. %s", err.Error())
		}

		contentLength, err := readable.ReadPrimitiveUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read Empty content length. %s", err.Error())
		}

		if readable.Remaining() < uint64(contentLength) {
			return errors.New("Empty content length longer than data size")
		}

		e.SetStructureVersion(version)
	}

	return nil
}

// Copy returns a pointer to a copy of the Empty. Requires type assertion when used
func (e *Empty) Copy() RVType {
	copied := NewEmpty()
	copied.structureVersion = e.structureVersion

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (e *Empty) Equals(o RVType) bool {
	if _, ok := o.(*Empty); !ok {
		return false
	}

	return (*e).structureVersion == (*o.(*Empty)).structureVersion
}

// String returns a string representation of the struct
func (e *Empty) String() string {
	return e.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (e *Empty) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Empty{\n")
	b.WriteString(fmt.Sprintf("%sstructureVersion: %d\n", indentationValues, e.structureVersion))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewEmpty returns a new Empty Structure
func NewEmpty() *Empty {
	return &Empty{}
}
