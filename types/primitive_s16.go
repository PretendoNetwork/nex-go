package types

import "fmt"

// PrimitiveS16 is wrapper around a Go primitive int16 with receiver methods to conform to RVType
type PrimitiveS16 struct {
	Value int16
}

// WriteTo writes the int16 to the given writable
func (s16 *PrimitiveS16) WriteTo(writable Writable) {
	writable.WritePrimitiveInt16LE(s16.Value)
}

// ExtractFrom extracts the int16 from the given readable
func (s16 *PrimitiveS16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt16LE()
	if err != nil {
		return err
	}

	s16.Value = value

	return nil
}

// Copy returns a pointer to a copy of the int16. Requires type assertion when used
func (s16 *PrimitiveS16) Copy() RVType {
	return NewPrimitiveS16(s16.Value)
}

// Equals checks if the input is equal in value to the current instance
func (s16 *PrimitiveS16) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS16); !ok {
		return false
	}

	return s16.Value == o.(*PrimitiveS16).Value
}

// String returns a string representation of the struct
func (s16 *PrimitiveS16) String() string {
	return fmt.Sprintf("%d", s16.Value)
}

// AND runs a bitwise AND operation on the PrimitiveS16 value. Consumes and returns a NEX primitive
func (s16 *PrimitiveS16) AND(other *PrimitiveS16) *PrimitiveS16 {
	return NewPrimitiveS16(s16.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveS16 value. Consumes and returns a Go primitive
func (s16 *PrimitiveS16) PAND(value int16) int16 {
	return s16.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveS16 value. Consumes and returns a NEX primitive
func (s16 *PrimitiveS16) OR(other *PrimitiveS16) *PrimitiveS16 {
	return NewPrimitiveS16(s16.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveS16 value. Consumes and returns a Go primitive
func (s16 *PrimitiveS16) POR(value int16) int16 {
	return s16.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveS16 value. Consumes and returns a NEX primitive
func (s16 *PrimitiveS16) XOR(other *PrimitiveS16) *PrimitiveS16 {
	return NewPrimitiveS16(s16.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveS16 value. Consumes and returns a Go primitive
func (s16 *PrimitiveS16) PXOR(value int16) int16 {
	return s16.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveS16 value. Returns a NEX primitive
func (s16 *PrimitiveS16) NOT() *PrimitiveS16 {
	return NewPrimitiveS16(s16.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveS16 value. Returns a Go primitive
func (s16 *PrimitiveS16) PNOT() int16 {
	return ^s16.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveS16 value. Consumes and returns a NEX primitive
func (s16 *PrimitiveS16) ANDNOT(other *PrimitiveS16) *PrimitiveS16 {
	return NewPrimitiveS16(s16.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveS16 value. Consumes and returns a Go primitive
func (s16 *PrimitiveS16) PANDNOT(value int16) int16 {
	return s16.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveS16 value. Consumes and returns a NEX primitive
func (s16 *PrimitiveS16) LShift(other *PrimitiveS16) *PrimitiveS16 {
	return NewPrimitiveS16(s16.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveS16 value. Consumes and returns a Go primitive
func (s16 *PrimitiveS16) PLShift(value int16) int16 {
	return s16.Value &^ value
}

// RShift runs a right shift operation on the PrimitiveS16 value. Consumes and returns a NEX primitive
func (s16 *PrimitiveS16) RShift(other *PrimitiveS16) *PrimitiveS16 {
	return NewPrimitiveS16(s16.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveS16 value. Consumes and returns a Go primitive
func (s16 *PrimitiveS16) PRShift(value int16) int16 {
	return s16.Value &^ value
}

// NewPrimitiveS16 returns a new PrimitiveS16
func NewPrimitiveS16(i16 int16) *PrimitiveS16 {
	return &PrimitiveS16{Value: i16}
}
