package types

import "fmt"

// PrimitiveU8 is wrapper around a Go primitive uint8 with receiver methods to conform to RVType
type PrimitiveU8 struct {
	Value uint8
}

// WriteTo writes the uint8 to the given writable
func (u8 *PrimitiveU8) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt8(u8.Value)
}

// ExtractFrom extracts the uint8 from the given readable
func (u8 *PrimitiveU8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt8()
	if err != nil {
		return err
	}

	u8.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint8. Requires type assertion when used
func (u8 *PrimitiveU8) Copy() RVType {
	return NewPrimitiveU8(u8.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u8 *PrimitiveU8) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU8); !ok {
		return false
	}

	return u8.Value == o.(*PrimitiveU8).Value
}

// String returns a string representation of the struct
func (u8 *PrimitiveU8) String() string {
	return fmt.Sprintf("%d", u8.Value)
}

// AND runs a bitwise AND operation on the PrimitiveU8 value. Consumes and returns a NEX primitive
func (u8 *PrimitiveU8) AND(other *PrimitiveU8) *PrimitiveU8 {
	return NewPrimitiveU8(u8.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveU8 value. Consumes and returns a Go primitive
func (u8 *PrimitiveU8) PAND(value uint8) uint8 {
	return u8.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveU8 value. Consumes and returns a NEX primitive
func (u8 *PrimitiveU8) OR(other *PrimitiveU8) *PrimitiveU8 {
	return NewPrimitiveU8(u8.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveU8 value. Consumes and returns a Go primitive
func (u8 *PrimitiveU8) POR(value uint8) uint8 {
	return u8.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveU8 value. Consumes and returns a NEX primitive
func (u8 *PrimitiveU8) XOR(other *PrimitiveU8) *PrimitiveU8 {
	return NewPrimitiveU8(u8.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveU8 value. Consumes and returns a Go primitive
func (u8 *PrimitiveU8) PXOR(value uint8) uint8 {
	return u8.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveU8 value. Returns a NEX primitive
func (u8 *PrimitiveU8) NOT() *PrimitiveU8 {
	return NewPrimitiveU8(u8.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveU8 value. Returns a Go primitive
func (u8 *PrimitiveU8) PNOT() uint8 {
	return ^u8.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveU8 value. Consumes and returns a NEX primitive
func (u8 *PrimitiveU8) ANDNOT(other *PrimitiveU8) *PrimitiveU8 {
	return NewPrimitiveU8(u8.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveU8 value. Consumes and returns a Go primitive
func (u8 *PrimitiveU8) PANDNOT(value uint8) uint8 {
	return u8.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveU8 value. Consumes and returns a NEX primitive
func (u8 *PrimitiveU8) LShift(other *PrimitiveU8) *PrimitiveU8 {
	return NewPrimitiveU8(u8.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveU8 value. Consumes and returns a Go primitive
func (u8 *PrimitiveU8) PLShift(value uint8) uint8 {
	return u8.Value &^ value
}

// RShift runs a right shift operation on the PrimitiveU8 value. Consumes and returns a NEX primitive
func (u8 *PrimitiveU8) RShift(other *PrimitiveU8) *PrimitiveU8 {
	return NewPrimitiveU8(u8.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveU8 value. Consumes and returns a Go primitive
func (u8 *PrimitiveU8) PRShift(value uint8) uint8 {
	return u8.Value &^ value
}

// NewPrimitiveU8 returns a new PrimitiveU8
func NewPrimitiveU8(ui8 uint8) *PrimitiveU8 {
	return &PrimitiveU8{Value: ui8}
}
