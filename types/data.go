package types

import (
	"fmt"
	"strings"
)

// Data is the base class for many other structures. The structure itself has no fields
type Data struct {
	Structure
}

// WriteTo writes the Data to the given writable
func (e *Data) WriteTo(writable Writable) {
	e.WriteHeaderTo(writable, 0)
}

// ExtractFrom extracts the Data from the given readable
func (e *Data) ExtractFrom(readable Readable) error {
	if err := e.ExtractHeaderFrom(readable); err != nil {
		return fmt.Errorf("Failed to read Data header. %s", err.Error())
	}

	return nil
}

// Copy returns a pointer to a copy of the Data. Requires type assertion when used
func (e *Data) Copy() RVType {
	copied := NewData()
	copied.StructureVersion = e.StructureVersion

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (e *Data) Equals(o RVType) bool {
	if _, ok := o.(*Data); !ok {
		return false
	}

	return (*e).StructureVersion == (*o.(*Data)).StructureVersion
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
	b.WriteString(fmt.Sprintf("%sStructureVersion: %d\n", indentationValues, e.StructureVersion))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewData returns a new Data Structure
func NewData() *Data {
	return &Data{}
}
