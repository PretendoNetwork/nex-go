package types

import (
	"errors"
	"fmt"
	"strings"
)

// NullData is a Structure with no fields
type NullData struct {
	Structure
}

// WriteTo writes the NullData to the given writable
func (nd *NullData) WriteTo(writable Writable) {
	if writable.UseStructureHeader() {
		writable.WritePrimitiveUInt8(nd.StructureVersion())
		writable.WritePrimitiveUInt32LE(0)
	}
}

// ExtractFrom extracts the NullData to the given readable
func (nd *NullData) ExtractFrom(readable Readable) error {
	if readable.UseStructureHeader() {
		version, err := readable.ReadPrimitiveUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read NullData version. %s", err.Error())
		}

		contentLength, err := readable.ReadPrimitiveUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read NullData content length. %s", err.Error())
		}

		if readable.Remaining() < uint64(contentLength) {
			return errors.New("NullData content length longer than data size")
		}

		nd.SetStructureVersion(version)
	}

	return nil
}

// Copy returns a pointer to a copy of the NullData. Requires type assertion when used
func (nd NullData) Copy() RVType {
	copied := NewNullData()
	copied.structureVersion = nd.structureVersion

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (nd *NullData) Equals(o RVType) bool {
	if _, ok := o.(*NullData); !ok {
		return false
	}

	return (*nd).structureVersion == (*o.(*NullData)).structureVersion
}

// String returns a string representation of the struct
func (nd *NullData) String() string {
	return nd.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (nd *NullData) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("NullData{\n")
	b.WriteString(fmt.Sprintf("%sstructureVersion: %d\n", indentationValues, nd.structureVersion))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewNullData returns a new NullData Structure
func NewNullData() *NullData {
	return &NullData{}
}
