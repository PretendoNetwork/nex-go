package types

import (
	"errors"
	"fmt"
	"strings"
)

// VariantTypes holds a mapping of RVTypes that are accessible in a Variant
var VariantTypes = make(map[UInt8]RVType)

// RegisterVariantType registers a RVType to be accessible in a Variant
func RegisterVariantType(id UInt8, rvType RVType) {
	VariantTypes[id] = rvType
}

// Variant is an implementation of rdv::Variant.
// This type can hold many other types, denoted by a type ID.
type Variant struct {
	TypeID UInt8
	Type   RVType
}

// WriteTo writes the Variant to the given writable
func (v Variant) WriteTo(writable Writable) {
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

	typeID := v.TypeID

	// * Type ID of 0 is a "None" type. There is no data
	if typeID == 0 {
		return nil
	}

	if _, ok := VariantTypes[typeID]; !ok {
		return fmt.Errorf("Invalid Variant type ID %d", typeID)
	}

	v.Type = VariantTypes[typeID].Copy()

	ptr, ok := any(&v.Type).(RVTypePtr)
	if !ok {
		return errors.New("Variant type data is not a valid RVType. Missing ExtractFrom pointer receiver")
	}

	if err := ptr.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read Variant type data. %s", err.Error())
	}

	return nil
}

// Copy returns a pointer to a copy of the Variant. Requires type assertion when used
func (v Variant) Copy() RVType {
	copied := NewVariant()

	copied.TypeID = v.TypeID.Copy().(UInt8)

	if v.Type != nil {
		copied.Type = v.Type.Copy()
	}

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (v Variant) Equals(o RVType) bool {
	if _, ok := o.(Variant); !ok {
		return false
	}

	other := o.(Variant)

	if !v.TypeID.Equals(other.TypeID) {
		return false
	}

	if v.Type != nil {
		return v.Type.Equals(other.Type)
	}

	return true
}

// String returns a string representation of the struct
func (v Variant) String() string {
	return v.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (v Variant) FormatToString(indentationLevel int) string {
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
func NewVariant() Variant {
	// * Type ID of 0 is a "None" type. There is no data
	return Variant{
		TypeID: NewUInt8(0),
		Type:   nil,
	}
}
