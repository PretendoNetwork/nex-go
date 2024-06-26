package types

import "fmt"

// Int32 is a type alias for the Go basic type int32 for use as an RVType
type Int32 int32

// WriteTo writes the Int32 to the given writable
func (i32 Int32) WriteTo(writable Writable) {
	writable.WriteInt32LE(int32(i32))
}

// ExtractFrom extracts the Int32 value from the given readable
func (i32 *Int32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt32LE()
	if err != nil {
		return err
	}

	*i32 = Int32(value)
	return nil
}

// Copy returns a pointer to a copy of the Int32. Requires type assertion when used
func (i32 Int32) Copy() RVType {
	return NewInt32(int32(i32))
}

// Equals checks if the input is equal in value to the current instance
func (i32 Int32) Equals(o RVType) bool {
	other, ok := o.(Int32)
	if !ok {
		return false
	}

	return i32 == other
}

// String returns a string representation of the Int32
func (i32 Int32) String() string {
	return fmt.Sprintf("%d", i32)
}

// NewInt32 returns a new Int32
func NewInt32(input int32) Int32 {
	i32 := Int32(input)
	return i32
}
