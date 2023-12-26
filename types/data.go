package types

import (
	"errors"
	"fmt"
	"strings"
)

// Data is the base class for many other structures. The structure itself has no fields
type Data struct {
	Structure
}

// WriteTo writes the Data to the given writable
func (e *Data) WriteTo(writable Writable) {
	if writable.UseStructureHeader() {
		writable.WritePrimitiveUInt8(e.StructureVersion())
		writable.WritePrimitiveUInt32LE(0)
	}
}

// ExtractFrom extracts the Data to the given readable
func (e *Data) ExtractFrom(readable Readable) error {
	if readable.UseStructureHeader() {
		version, err := readable.ReadPrimitiveUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read Data version. %s", err.Error())
		}

		contentLength, err := readable.ReadPrimitiveUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read Data content length. %s", err.Error())
		}

		if readable.Remaining() < uint64(contentLength) {
			return errors.New("Data content length longer than data size")
		}

		e.SetStructureVersion(version)
	}

	return nil
}

// Copy returns a pointer to a copy of the Data. Requires type assertion when used
func (e *Data) Copy() RVType {
	copied := NewData()
	copied.structureVersion = e.structureVersion

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (e *Data) Equals(o RVType) bool {
	if _, ok := o.(*Data); !ok {
		return false
	}

	return (*e).structureVersion == (*o.(*Data)).structureVersion
}

// String returns a string representation of the struct
func (e *Data) String() string {
	return e.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (e *Data) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Data{\n")
	b.WriteString(fmt.Sprintf("%sstructureVersion: %d\n", indentationValues, e.structureVersion))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewData returns a new Data Structure
func NewData() *Data {
	return &Data{}
}
