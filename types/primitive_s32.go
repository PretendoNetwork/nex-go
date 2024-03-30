package types

import "fmt"

// PrimitiveS32 is wrapper around a Go primitive int32 with receiver methods to conform to RVType
type PrimitiveS32 struct {
	Value int32
}

// WriteTo writes the int32 to the given writable
func (s32 *PrimitiveS32) WriteTo(writable Writable) {
	writable.WritePrimitiveInt32LE(s32.Value)
}

// ExtractFrom extracts the int32 from the given readable
func (s32 *PrimitiveS32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt32LE()
	if err != nil {
		return err
	}

	s32.Value = value

	return nil
}

// Copy returns a pointer to a copy of the int32. Requires type assertion when used
func (s32 *PrimitiveS32) Copy() RVType {
	return NewPrimitiveS32(s32.Value)
}

// Equals checks if the input is equal in value to the current instance
func (s32 *PrimitiveS32) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS32); !ok {
		return false
	}

	return s32.Value == o.(*PrimitiveS32).Value
}

// String returns a string representation of the struct
func (s32 *PrimitiveS32) String() string {
	return fmt.Sprintf("%d", s32.Value)
}

// AND runs a bitwise AND operation on the PrimitiveS32 value. Consumes and returns a NEX primitive
func (s32 *PrimitiveS32) AND(other *PrimitiveS32) *PrimitiveS32 {
	return NewPrimitiveS32(s32.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveS32 value. Consumes and returns a Go primitive
func (s32 *PrimitiveS32) PAND(value int32) int32 {
	return s32.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveS32 value. Consumes and returns a NEX primitive
func (s32 *PrimitiveS32) OR(other *PrimitiveS32) *PrimitiveS32 {
	return NewPrimitiveS32(s32.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveS32 value. Consumes and returns a Go primitive
func (s32 *PrimitiveS32) POR(value int32) int32 {
	return s32.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveS32 value. Consumes and returns a NEX primitive
func (s32 *PrimitiveS32) XOR(other *PrimitiveS32) *PrimitiveS32 {
	return NewPrimitiveS32(s32.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveS32 value. Consumes and returns a Go primitive
func (s32 *PrimitiveS32) PXOR(value int32) int32 {
	return s32.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveS32 value. Returns a NEX primitive
func (s32 *PrimitiveS32) NOT() *PrimitiveS32 {
	return NewPrimitiveS32(s32.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveS32 value. Returns a Go primitive
func (s32 *PrimitiveS32) PNOT() int32 {
	return ^s32.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveS32 value. Consumes and returns a NEX primitive
func (s32 *PrimitiveS32) ANDNOT(other *PrimitiveS32) *PrimitiveS32 {
	return NewPrimitiveS32(s32.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveS32 value. Consumes and returns a Go primitive
func (s32 *PrimitiveS32) PANDNOT(value int32) int32 {
	return s32.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveS32 value. Consumes and returns a NEX primitive
func (s32 *PrimitiveS32) LShift(other *PrimitiveS32) *PrimitiveS32 {
	return NewPrimitiveS32(s32.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveS32 value. Consumes and returns a Go primitive
func (s32 *PrimitiveS32) PLShift(value int32) int32 {
	return s32.Value << value
}

// RShift runs a right shift operation on the PrimitiveS32 value. Consumes and returns a NEX primitive
func (s32 *PrimitiveS32) RShift(other *PrimitiveS32) *PrimitiveS32 {
	return NewPrimitiveS32(s32.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveS32 value. Consumes and returns a Go primitive
func (s32 *PrimitiveS32) PRShift(value int32) int32 {
	return s32.Value >> value
}

// NewPrimitiveS32 returns a new PrimitiveS32
func NewPrimitiveS32(i32 int32) *PrimitiveS32 {
	return &PrimitiveS32{Value: i32}
}
