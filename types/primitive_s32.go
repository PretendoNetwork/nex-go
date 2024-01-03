package types

import "fmt"

// PrimitiveS32 is a struct of int32 with receiver methods to conform to RVType
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

// NewPrimitiveS32 returns a new PrimitiveS32
func NewPrimitiveS32(i32 int32) *PrimitiveS32 {
	return &PrimitiveS32{Value: i32}
}
