package types

import "fmt"

// Int32 is a type alias for the Go basic type int32 for use as an RVType
type Int32 int32

// WriteTo writes the Int32 to the given writable
func (s32 Int32) WriteTo(writable Writable) {
	writable.WritePrimitiveInt32LE(int32(s32))
}

// ExtractFrom extracts the Int32 value from the given readable
func (s32 *Int32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt32LE()
	if err != nil {
		return err
	}

	*s32 = Int32(value)
	return nil
}

// Copy returns a pointer to a copy of the Int32. Requires type assertion when used
func (s32 Int32) Copy() RVType {
	return &s32
}

// Equals checks if the input is equal in value to the current instance
func (s32 Int32) Equals(o RVType) bool {
	other, ok := o.(*Int32)
	if !ok {
		return false
	}
	return s32 == *other
}

// String returns a string representation of the Int32
func (s32 Int32) String() string {
	return fmt.Sprintf("%d", s32)
}

// NewInt32 returns a new Int32 pointer
func NewInt32(input int32) *Int32 {
	s32 := Int32(input)
	return &s32
}

