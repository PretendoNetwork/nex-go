package types

import (
	"fmt"
	"strings"
)

// VariantTypes holds a mapping of RVTypes that are accessible in a Variant
var VariantTypes = make(map[uint8]RVType)

// RegisterVariantType registers a RVType to be accessible in a Variant
func RegisterVariantType(id uint8, rvType RVType) {
	VariantTypes[id] = rvType
}

// Variant is a type which can old many other types
type Variant struct {
	TypeID *PrimitiveU8
	Type   RVType
}

// WriteTo writes the Variant to the given writable
func (v *Variant) WriteTo(writable Writable) {
	v.TypeID.WriteTo(writable)
	v.Type.WriteTo(writable)
}

// ExtractFrom extracts the Variant from the given readable
func (v *Variant) ExtractFrom(readable Readable) error {
	err := v.TypeID.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read Variant type ID. %s", err.Error())
	}

	if _, ok := VariantTypes[v.TypeID.Value]; !ok {
		return fmt.Errorf("Invalid Variant type ID %d", v.TypeID)
	}

	v.Type = VariantTypes[v.TypeID.Value].Copy()

	return v.Type.ExtractFrom(readable)
}

// Copy returns a pointer to a copy of the Variant. Requires type assertion when used
func (v *Variant) Copy() RVType {
	copied := NewVariant()

	copied.TypeID = v.TypeID.Copy().(*PrimitiveU8)
	copied.Type = v.Type.Copy()

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (v *Variant) Equals(o RVType) bool {
	if _, ok := o.(*Variant); !ok {
		return false
	}

	other := o.(*Variant)

	if !v.TypeID.Equals(other.TypeID) {
		return false
	}

	return v.Type.Equals(other.Type)
}

// String returns a string representation of the struct
func (v *Variant) String() string {
	return v.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (v *Variant) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Variant{\n")
	b.WriteString(fmt.Sprintf("%TypeID: %s,\n", indentationValues, v.TypeID))
	b.WriteString(fmt.Sprintf("%Type: %s\n", indentationValues, v.Type))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewVariant returns a new Variant
func NewVariant() *Variant {
	return &Variant{
		TypeID: NewPrimitiveU8(0),
	}
}
