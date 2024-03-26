package types

import "fmt"

// PrimitiveS8 is wrapper around a Go primitive int8 with receiver methods to conform to RVType
type PrimitiveS8 struct {
	Value int8
}

// WriteTo writes the int8 to the given writable
func (s8 *PrimitiveS8) WriteTo(writable Writable) {
	writable.WritePrimitiveInt8(s8.Value)
}

// ExtractFrom extracts the int8 from the given readable
func (s8 *PrimitiveS8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt8()
	if err != nil {
		return err
	}

	s8.Value = value

	return nil
}

// Copy returns a pointer to a copy of the int8. Requires type assertion when used
func (s8 *PrimitiveS8) Copy() RVType {
	return NewPrimitiveS8(s8.Value)
}

// Equals checks if the input is equal in value to the current instance
func (s8 *PrimitiveS8) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS8); !ok {
		return false
	}

	return s8.Value == o.(*PrimitiveS8).Value
}

// String returns a string representation of the struct
func (s8 *PrimitiveS8) String() string {
	return fmt.Sprintf("%d", s8.Value)
}

// AND runs a bitwise AND operation on the PrimitiveS8 value. Consumes and returns a NEX primitive
func (s8 *PrimitiveS8) AND(other *PrimitiveS8) *PrimitiveS8 {
	return NewPrimitiveS8(s8.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveS8 value. Consumes and returns a Go primitive
func (s8 *PrimitiveS8) PAND(value int8) int8 {
	return s8.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveS8 value. Consumes and returns a NEX primitive
func (s8 *PrimitiveS8) OR(other *PrimitiveS8) *PrimitiveS8 {
	return NewPrimitiveS8(s8.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveS8 value. Consumes and returns a Go primitive
func (s8 *PrimitiveS8) POR(value int8) int8 {
	return s8.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveS8 value. Consumes and returns a NEX primitive
func (s8 *PrimitiveS8) XOR(other *PrimitiveS8) *PrimitiveS8 {
	return NewPrimitiveS8(s8.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveS8 value. Consumes and returns a Go primitive
func (s8 *PrimitiveS8) PXOR(value int8) int8 {
	return s8.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveS8 value. Returns a NEX primitive
func (s8 *PrimitiveS8) NOT() *PrimitiveS8 {
	return NewPrimitiveS8(s8.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveS8 value. Returns a Go primitive
func (s8 *PrimitiveS8) PNOT() int8 {
	return ^s8.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveS8 value. Consumes and returns a NEX primitive
func (s8 *PrimitiveS8) ANDNOT(other *PrimitiveS8) *PrimitiveS8 {
	return NewPrimitiveS8(s8.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveS8 value. Consumes and returns a Go primitive
func (s8 *PrimitiveS8) PANDNOT(value int8) int8 {
	return s8.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveS8 value. Consumes and returns a NEX primitive
func (s8 *PrimitiveS8) LShift(other *PrimitiveS8) *PrimitiveS8 {
	return NewPrimitiveS8(s8.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveS8 value. Consumes and returns a Go primitive
func (s8 *PrimitiveS8) PLShift(value int8) int8 {
	return s8.Value &^ value
}

// RShift runs a right shift operation on the PrimitiveS8 value. Consumes and returns a NEX primitive
func (s8 *PrimitiveS8) RShift(other *PrimitiveS8) *PrimitiveS8 {
	return NewPrimitiveS8(s8.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveS8 value. Consumes and returns a Go primitive
func (s8 *PrimitiveS8) PRShift(value int8) int8 {
	return s8.Value &^ value
}

// NewPrimitiveS8 returns a new PrimitiveS8
func NewPrimitiveS8(i8 int8) *PrimitiveS8 {
	return &PrimitiveS8{Value: i8}
}
