package types

import (
	"fmt"
	"strings"
)

// Data is an implementation of rdv::Data.
// This structure has no data, and instead acts as the base class for many other structures.
type Data struct {
	Structure
}

// ObjectID returns the object identifier of the type
func (d Data) ObjectID() RVType {
	return d.DataObjectID()
}

// DataObjectID returns the object identifier of the type embedding Data
func (d Data) DataObjectID() RVType {
	return NewString("Data")
}

// WriteTo writes the Data to the given writable
func (d Data) WriteTo(writable Writable) {
	d.WriteHeaderTo(writable, 0)
}

// ExtractFrom extracts the Data from the given readable
func (d *Data) ExtractFrom(readable Readable) error {
	if err := d.ExtractHeaderFrom(readable); err != nil {
		return fmt.Errorf("Failed to read Data header. %s", err.Error())
	}

	return nil
}

// Copy returns a pointer to a copy of the Data. Requires type assertion when used
func (d Data) Copy() RVType {
	copied := NewData()
	copied.StructureVersion = d.StructureVersion

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (d Data) Equals(o RVType) bool {
	if _, ok := o.(Data); !ok {
		return false
	}

	other := o.(Data)

	return d.StructureVersion == other.StructureVersion
}

// CopyRef copies the current value of the Data
// and returns a pointer to the new copy
func (d Data) CopyRef() RVTypePtr {
	copied := d.Copy().(Data)
	return &copied
}

// Deref takes a pointer to the Data
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (d *Data) Deref() RVType {
	return *d
}

// String returns a string representation of the struct
func (d Data) String() string {
	return d.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (d Data) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Data{\n")
	b.WriteString(fmt.Sprintf("%sStructureVersion: %d\n", indentationValues, d.StructureVersion))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewData returns a new Data Structure
func NewData() Data {
	return Data{}
}
