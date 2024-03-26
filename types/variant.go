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

// Variant is an implementation of rdv::Variant.
// This type can hold many other types, denoted by a type ID.
type Variant struct {
	TypeID *PrimitiveU8
	Type   RVType
}

// WriteTo writes the Variant to the given writable
func (v *Variant) WriteTo(writable Writable) {
	v.TypeID.WriteTo(writable)

	if v.Type != nil {
		v.Type.WriteTo(writable)
	}
}

// ExtractFrom extracts the Variant from the given readable
func (v *Variant) ExtractFrom(readable Readable) error {
	err := v.TypeID.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read Variant type ID. %s", err.Error())
	}

	// * Type ID of 0 is a "None" type. There is no data
	if v.TypeID.Value == 0 {
		return nil
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

	if v.Type != nil {
		copied.Type = v.Type.Copy()
	}

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

	if v.Type != nil {
		return v.Type.Equals(other.Type)
	}

	return true
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
	b.WriteString(fmt.Sprintf("%sTypeID: %s,\n", indentationValues, v.TypeID))

	if v.Type != nil {
		b.WriteString(fmt.Sprintf("%sType: %s\n", indentationValues, v.Type))
	} else {
		b.WriteString(fmt.Sprintf("%sType: None\n", indentationValues))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewVariant returns a new Variant
func NewVariant() *Variant {
	// * Type ID of 0 is a "None" type. There is no data
	return &Variant{
		TypeID: NewPrimitiveU8(0),
		Type:   nil,
	}
}
