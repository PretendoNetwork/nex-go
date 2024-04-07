package types

import "fmt"

// PrimitiveU16 is wrapper around a Go primitive uint16 with receiver methods to conform to RVType
type PrimitiveU16 struct {
	Value uint16
}

// WriteTo writes the uint16 to the given writable
func (u16 *PrimitiveU16) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt16LE(u16.Value)
}

// ExtractFrom extracts the uint16 from the given readable
func (u16 *PrimitiveU16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt16LE()
	if err != nil {
		return err
	}

	u16.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint16. Requires type assertion when used
func (u16 *PrimitiveU16) Copy() RVType {
	return NewPrimitiveU16(u16.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u16 *PrimitiveU16) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU16); !ok {
		return false
	}

	return u16.Value == o.(*PrimitiveU16).Value
}

// String returns a string representation of the struct
func (u16 *PrimitiveU16) String() string {
	return fmt.Sprintf("%d", u16.Value)
}

// AND runs a bitwise AND operation on the PrimitiveU16 value. Consumes and returns a NEX primitive
func (u16 *PrimitiveU16) AND(other *PrimitiveU16) *PrimitiveU16 {
	return NewPrimitiveU16(u16.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveU16 value. Consumes and returns a Go primitive
func (u16 *PrimitiveU16) PAND(value uint16) uint16 {
	return u16.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveU16 value. Consumes and returns a NEX primitive
func (u16 *PrimitiveU16) OR(other *PrimitiveU16) *PrimitiveU16 {
	return NewPrimitiveU16(u16.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveU16 value. Consumes and returns a Go primitive
func (u16 *PrimitiveU16) POR(value uint16) uint16 {
	return u16.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveU16 value. Consumes and returns a NEX primitive
func (u16 *PrimitiveU16) XOR(other *PrimitiveU16) *PrimitiveU16 {
	return NewPrimitiveU16(u16.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveU16 value. Consumes and returns a Go primitive
func (u16 *PrimitiveU16) PXOR(value uint16) uint16 {
	return u16.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveU16 value. Returns a NEX primitive
func (u16 *PrimitiveU16) NOT() *PrimitiveU16 {
	return NewPrimitiveU16(u16.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveU16 value. Returns a Go primitive
func (u16 *PrimitiveU16) PNOT() uint16 {
	return ^u16.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveU16 value. Consumes and returns a NEX primitive
func (u16 *PrimitiveU16) ANDNOT(other *PrimitiveU16) *PrimitiveU16 {
	return NewPrimitiveU16(u16.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveU16 value. Consumes and returns a Go primitive
func (u16 *PrimitiveU16) PANDNOT(value uint16) uint16 {
	return u16.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveU16 value. Consumes and returns a NEX primitive
func (u16 *PrimitiveU16) LShift(other *PrimitiveU16) *PrimitiveU16 {
	return NewPrimitiveU16(u16.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveU16 value. Consumes and returns a Go primitive
func (u16 *PrimitiveU16) PLShift(value uint16) uint16 {
	return u16.Value << value
}

// RShift runs a right shift operation on the PrimitiveU16 value. Consumes and returns a NEX primitive
func (u16 *PrimitiveU16) RShift(other *PrimitiveU16) *PrimitiveU16 {
	return NewPrimitiveU16(u16.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveU16 value. Consumes and returns a Go primitive
func (u16 *PrimitiveU16) PRShift(value uint16) uint16 {
	return u16.Value >> value
}

// NewPrimitiveU16 returns a new PrimitiveU16
func NewPrimitiveU16(ui16 uint16) *PrimitiveU16 {
	return &PrimitiveU16{Value: ui16}
}
