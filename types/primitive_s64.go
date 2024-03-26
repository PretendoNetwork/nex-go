package types

import "fmt"

// PrimitiveS64 is wrapper around a Go primitive int64 with receiver methods to conform to RVType
type PrimitiveS64 struct {
	Value int64
}

// WriteTo writes the int64 to the given writable
func (s64 *PrimitiveS64) WriteTo(writable Writable) {
	writable.WritePrimitiveInt64LE(s64.Value)
}

// ExtractFrom extracts the int64 from the given readable
func (s64 *PrimitiveS64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt64LE()
	if err != nil {
		return err
	}

	s64.Value = value

	return nil
}

// Copy returns a pointer to a copy of the int64. Requires type assertion when used
func (s64 *PrimitiveS64) Copy() RVType {
	return NewPrimitiveS64(s64.Value)
}

// Equals checks if the input is equal in value to the current instance
func (s64 *PrimitiveS64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS64); !ok {
		return false
	}

	return s64.Value == o.(*PrimitiveS64).Value
}

// String returns a string representation of the struct
func (s64 *PrimitiveS64) String() string {
	return fmt.Sprintf("%d", s64.Value)
}

// AND runs a bitwise AND operation on the PrimitiveS64 value. Consumes and returns a NEX primitive
func (s64 *PrimitiveS64) AND(other *PrimitiveS64) *PrimitiveS64 {
	return NewPrimitiveS64(s64.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveS64 value. Consumes and returns a Go primitive
func (s64 *PrimitiveS64) PAND(value int64) int64 {
	return s64.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveS64 value. Consumes and returns a NEX primitive
func (s64 *PrimitiveS64) OR(other *PrimitiveS64) *PrimitiveS64 {
	return NewPrimitiveS64(s64.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveS64 value. Consumes and returns a Go primitive
func (s64 *PrimitiveS64) POR(value int64) int64 {
	return s64.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveS64 value. Consumes and returns a NEX primitive
func (s64 *PrimitiveS64) XOR(other *PrimitiveS64) *PrimitiveS64 {
	return NewPrimitiveS64(s64.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveS64 value. Consumes and returns a Go primitive
func (s64 *PrimitiveS64) PXOR(value int64) int64 {
	return s64.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveS64 value. Returns a NEX primitive
func (s64 *PrimitiveS64) NOT() *PrimitiveS64 {
	return NewPrimitiveS64(s64.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveS64 value. Returns a Go primitive
func (s64 *PrimitiveS64) PNOT() int64 {
	return ^s64.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveS64 value. Consumes and returns a NEX primitive
func (s64 *PrimitiveS64) ANDNOT(other *PrimitiveS64) *PrimitiveS64 {
	return NewPrimitiveS64(s64.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveS64 value. Consumes and returns a Go primitive
func (s64 *PrimitiveS64) PANDNOT(value int64) int64 {
	return s64.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveS64 value. Consumes and returns a NEX primitive
func (s64 *PrimitiveS64) LShift(other *PrimitiveS64) *PrimitiveS64 {
	return NewPrimitiveS64(s64.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveS64 value. Consumes and returns a Go primitive
func (s64 *PrimitiveS64) PLShift(value int64) int64 {
	return s64.Value &^ value
}

// RShift runs a right shift operation on the PrimitiveS64 value. Consumes and returns a NEX primitive
func (s64 *PrimitiveS64) RShift(other *PrimitiveS64) *PrimitiveS64 {
	return NewPrimitiveS64(s64.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveS64 value. Consumes and returns a Go primitive
func (s64 *PrimitiveS64) PRShift(value int64) int64 {
	return s64.Value &^ value
}

// NewPrimitiveS64 returns a new PrimitiveS64
func NewPrimitiveS64(i64 int64) *PrimitiveS64 {
	return &PrimitiveS64{Value: i64}
}
