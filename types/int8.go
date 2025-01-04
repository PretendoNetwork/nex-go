package types

import "fmt"

// Int8 is a type alias for the Go basic type int8 for use as an RVType
type Int8 int8

// WriteTo writes the Int8 to the given writable
func (i8 Int8) WriteTo(writable Writable) {
	writable.WriteInt8(int8(i8))
}

// ExtractFrom extracts the Int8 value from the given readable
func (i8 *Int8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt8()
	if err != nil {
		return err
	}

	*i8 = Int8(value)
	return nil
}

// Copy returns a pointer to a copy of the Int8. Requires type assertion when used
func (i8 Int8) Copy() RVType {
	return NewInt8(int8(i8))
}

// Equals checks if the input is equal in value to the current instance
func (i8 Int8) Equals(o RVType) bool {
	other, ok := o.(Int8)
	if !ok {
		return false
	}

	return i8 == other
}

// CopyRef copies the current value of the Int8
// and returns a pointer to the new copy
func (i8 Int8) CopyRef() RVTypePtr {
	copied := i8.Copy().(Int8)
	return &copied
}

// Deref takes a pointer to the Int8
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (i8 *Int8) Deref() RVType {
	return *i8
}

// String returns a string representation of the Int8
func (i8 Int8) String() string {
	return fmt.Sprintf("%d", i8)
}

// NewInt8 returns a new Int8
func NewInt8(input int8) Int8 {
	i8 := Int8(input)
	return i8
}
