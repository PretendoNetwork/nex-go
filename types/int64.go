package types

import "fmt"

// Int64 is a type alias for the Go basic type int64 for use as an RVType
type Int64 int64

// WriteTo writes the Int64 to the given writable
func (i64 Int64) WriteTo(writable Writable) {
	writable.WriteInt64LE(int64(i64))
}

// ExtractFrom extracts the Int64 value from the given readable
func (i64 *Int64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt64LE()
	if err != nil {
		return err
	}

	*i64 = Int64(value)
	return nil
}

// Copy returns a pointer to a copy of the Int64. Requires type assertion when used
func (i64 Int64) Copy() RVType {
	return NewInt64(int64(i64))
}

// Equals checks if the input is equal in value to the current instance
func (i64 Int64) Equals(o RVType) bool {
	other, ok := o.(Int64)
	if !ok {
		return false
	}

	return i64 == other
}

// CopyRef copies the current value of the Int64
// and returns a pointer to the new copy
func (i64 Int64) CopyRef() RVTypePtr {
	copied := NewInt64(int64(i64))
	return &copied
}

// Deref takes a pointer to the Int64
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (i64 *Int64) Deref() RVType {
	return *i64
}

// String returns a string representation of the Int64
func (i64 Int64) String() string {
	return fmt.Sprintf("%d", i64)
}

// NewInt64 returns a new Int64
func NewInt64(input int64) Int64 {
	i64 := Int64(input)
	return i64
}
