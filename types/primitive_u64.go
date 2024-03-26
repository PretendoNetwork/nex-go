package types

import "fmt"

// PrimitiveU64 is wrapper around a Go primitive uint64 with receiver methods to conform to RVType
type PrimitiveU64 struct {
	Value uint64
}

// WriteTo writes the uint64 to the given writable
func (u64 *PrimitiveU64) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt64LE(u64.Value)
}

// ExtractFrom extracts the uint64 from the given readable
func (u64 *PrimitiveU64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt64LE()
	if err != nil {
		return err
	}

	u64.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint64. Requires type assertion when used
func (u64 *PrimitiveU64) Copy() RVType {
	return NewPrimitiveU64(u64.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u64 *PrimitiveU64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU64); !ok {
		return false
	}

	return u64.Value == o.(*PrimitiveU64).Value
}

// String returns a string representation of the struct
func (u64 *PrimitiveU64) String() string {
	return fmt.Sprintf("%d", u64.Value)
}

// AND runs a bitwise AND operation on the PrimitiveU64 value. Consumes and returns a NEX primitive
func (u64 *PrimitiveU64) AND(other *PrimitiveU64) *PrimitiveU64 {
	return NewPrimitiveU64(u64.PAND(other.Value))
}

// PAND (Primitive AND) runs a bitwise AND operation on the PrimitiveU64 value. Consumes and returns a Go primitive
func (u64 *PrimitiveU64) PAND(value uint64) uint64 {
	return u64.Value & value
}

// OR runs a bitwise OR operation on the PrimitiveU64 value. Consumes and returns a NEX primitive
func (u64 *PrimitiveU64) OR(other *PrimitiveU64) *PrimitiveU64 {
	return NewPrimitiveU64(u64.POR(other.Value))
}

// POR (Primitive OR) runs a bitwise OR operation on the PrimitiveU64 value. Consumes and returns a Go primitive
func (u64 *PrimitiveU64) POR(value uint64) uint64 {
	return u64.Value | value
}

// XOR runs a bitwise XOR operation on the PrimitiveU64 value. Consumes and returns a NEX primitive
func (u64 *PrimitiveU64) XOR(other *PrimitiveU64) *PrimitiveU64 {
	return NewPrimitiveU64(u64.PXOR(other.Value))
}

// PXOR (Primitive XOR) runs a bitwise XOR operation on the PrimitiveU64 value. Consumes and returns a Go primitive
func (u64 *PrimitiveU64) PXOR(value uint64) uint64 {
	return u64.Value ^ value
}

// NOT runs a bitwise NOT operation on the PrimitiveU64 value. Returns a NEX primitive
func (u64 *PrimitiveU64) NOT() *PrimitiveU64 {
	return NewPrimitiveU64(u64.PNOT())
}

// PNOT (Primitive NOT) runs a bitwise NOT operation on the PrimitiveU64 value. Returns a Go primitive
func (u64 *PrimitiveU64) PNOT() uint64 {
	return ^u64.Value
}

// ANDNOT runs a bitwise ANDNOT operation on the PrimitiveU64 value. Consumes and returns a NEX primitive
func (u64 *PrimitiveU64) ANDNOT(other *PrimitiveU64) *PrimitiveU64 {
	return NewPrimitiveU64(u64.PANDNOT(other.Value))
}

// PANDNOT (Primitive AND-NOT) runs a bitwise AND-NOT operation on the PrimitiveU64 value. Consumes and returns a Go primitive
func (u64 *PrimitiveU64) PANDNOT(value uint64) uint64 {
	return u64.Value &^ value
}

// LShift runs a left shift operation on the PrimitiveU64 value. Consumes and returns a NEX primitive
func (u64 *PrimitiveU64) LShift(other *PrimitiveU64) *PrimitiveU64 {
	return NewPrimitiveU64(u64.PLShift(other.Value))
}

// PLShift (Primitive Left Shift) runs a left shift operation on the PrimitiveU64 value. Consumes and returns a Go primitive
func (u64 *PrimitiveU64) PLShift(value uint64) uint64 {
	return u64.Value &^ value
}

// RShift runs a right shift operation on the PrimitiveU64 value. Consumes and returns a NEX primitive
func (u64 *PrimitiveU64) RShift(other *PrimitiveU64) *PrimitiveU64 {
	return NewPrimitiveU64(u64.PRShift(other.Value))
}

// PRShift (Primitive Right Shift) runs a right shift operation on the PrimitiveU64 value. Consumes and returns a Go primitive
func (u64 *PrimitiveU64) PRShift(value uint64) uint64 {
	return u64.Value &^ value
}

// NewPrimitiveU64 returns a new PrimitiveU64
func NewPrimitiveU64(ui64 uint64) *PrimitiveU64 {
	return &PrimitiveU64{Value: ui64}
}
