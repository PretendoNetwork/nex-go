package types

import "fmt"

// PrimitiveU32 is wrapper around a Go primitive uint32 with receiver methods to conform to RVType
type PrimitiveU32 struct {
	Value uint32
}

// WriteTo writes the uint32 to the given writable
func (u32 *PrimitiveU32) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(u32.Value)
}

// ExtractFrom extracts the uint32 from the given readable
func (u32 *PrimitiveU32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return err
	}

	u32.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint32. Requires type assertion when used
func (u32 *PrimitiveU32) Copy() RVType {
	return NewPrimitiveU32(u32.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u32 *PrimitiveU32) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU32); !ok {
		return false
	}

	return u32.Value == o.(*PrimitiveU32).Value
}

// String returns a string representation of the struct
func (u32 *PrimitiveU32) String() string {
	return fmt.Sprintf("%d", u32.Value)
}

// AND runs a bitwise AND operation on the PrimitiveU32 value. Consumes and returns a NEX primitive
func (u32 *PrimitiveU32) AND(other *PrimitiveU32) *PrimitiveU32 {
	return NewPrimitiveU32(u32.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveU32 value. Consumes and returns a Go primitive
func (u32 *PrimitiveU32) PAND(value uint32) uint32 {
	return u32.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveU32 value. Consumes and returns a NEX primitive
func (u32 *PrimitiveU32) OR(other *PrimitiveU32) *PrimitiveU32 {
	return NewPrimitiveU32(u32.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveU32 value. Consumes and returns a Go primitive
func (u32 *PrimitiveU32) POR(value uint32) uint32 {
	return u32.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveU32 value. Consumes and returns a NEX primitive
func (u32 *PrimitiveU32) XOR(other *PrimitiveU32) *PrimitiveU32 {
	return NewPrimitiveU32(u32.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveU32 value. Consumes and returns a Go primitive
func (u32 *PrimitiveU32) PXOR(value uint32) uint32 {
	return u32.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveU32 value. Returns a NEX primitive
func (u32 *PrimitiveU32) NOT() *PrimitiveU32 {
	return NewPrimitiveU32(u32.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveU32 value. Returns a Go primitive
func (u32 *PrimitiveU32) PNOT() uint32 {
	return ^u32.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveU32 value. Consumes and returns a NEX primitive
func (u32 *PrimitiveU32) ANDNOT(other *PrimitiveU32) *PrimitiveU32 {
	return NewPrimitiveU32(u32.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveU32 value. Consumes and returns a Go primitive
func (u32 *PrimitiveU32) PANDNOT(value uint32) uint32 {
	return u32.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveU32 value. Consumes and returns a NEX primitive
func (u32 *PrimitiveU32) LShift(other *PrimitiveU32) *PrimitiveU32 {
	return NewPrimitiveU32(u32.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveU32 value. Consumes and returns a Go primitive
func (u32 *PrimitiveU32) PLShift(value uint32) uint32 {
	return u32.Value << value
}

// RShift runs a right shift operation on the PrimitiveU32 value. Consumes and returns a NEX primitive
func (u32 *PrimitiveU32) RShift(other *PrimitiveU32) *PrimitiveU32 {
	return NewPrimitiveU32(u32.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveU32 value. Consumes and returns a Go primitive
func (u32 *PrimitiveU32) PRShift(value uint32) uint32 {
	return u32.Value >> value
}

// NewPrimitiveU32 returns a new PrimitiveU32
func NewPrimitiveU32(ui32 uint32) *PrimitiveU32 {
	return &PrimitiveU32{Value: ui32}
}
