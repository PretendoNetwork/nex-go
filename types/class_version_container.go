package types

import (
	"fmt"
	"strings"
)

// ClassVersionContainer is an implementation of rdv::ClassVersionContainer.
// Contains version info for Structures used in verbose RMC messages.
type ClassVersionContainer struct {
	Structure
	ClassVersions Map[String, UInt16]
}

// WriteTo writes the ClassVersionContainer to the given writable
func (cvc ClassVersionContainer) WriteTo(writable Writable) {
	cvc.ClassVersions.WriteTo(writable)
}

// ExtractFrom extracts the ClassVersionContainer from the given readable
func (cvc *ClassVersionContainer) ExtractFrom(readable Readable) error {
	return cvc.ClassVersions.ExtractFrom(readable)
}

// Copy returns a pointer to a copy of the ClassVersionContainer. Requires type assertion when used
func (cvc ClassVersionContainer) Copy() RVType {
	copied := NewClassVersionContainer()
	copied.ClassVersions = cvc.ClassVersions.Copy().(Map[String, UInt16])

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (cvc ClassVersionContainer) Equals(o RVType) bool {
	if _, ok := o.(ClassVersionContainer); !ok {
		return false
	}

	return cvc.ClassVersions.Equals(o)
}

// String returns a string representation of the struct
func (cvc ClassVersionContainer) String() string {
	return cvc.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (cvc ClassVersionContainer) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("ClassVersionContainer{\n")
	b.WriteString(fmt.Sprintf("%sStructureVersion: %d,\n", indentationValues, cvc.StructureVersion))
	b.WriteString(fmt.Sprintf("%sClassVersions: %s\n", indentationValues, cvc.ClassVersions))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewClassVersionContainer returns a new ClassVersionContainer
func NewClassVersionContainer() ClassVersionContainer {
	cvc := ClassVersionContainer{
		ClassVersions: NewMap[String, UInt16](),
	}

	return cvc
}
