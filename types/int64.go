package types

import "fmt"

// Int64 is a type alias for the Go basic type int64 for use as an RVType
type Int64 int64

// WriteTo writes the Int64 to the given writable
func (s64 Int64) WriteTo(writable Writable) {
	writable.WritePrimitiveInt64LE(int64(s64))
}

// ExtractFrom extracts the Int64 value from the given readable
func (s64 *Int64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt64LE()
	if err != nil {
		return err
	}

	*s64 = Int64(value)
	return nil
}

// Copy returns a pointer to a copy of the Int64. Requires type assertion when used
func (s64 Int64) Copy() RVType {
	copy := s64
	return &copy
}

// Equals checks if the input is equal in value to the current instance
func (s64 Int64) Equals(o RVType) bool {
	other, ok := o.(*Int64)
	if !ok {
		return false
	}
	return s64 == *other
}

// String returns a string representation of the Int64
func (s64 Int64) String() string {
	return fmt.Sprintf("%d", s64)
}

// NewInt64 returns a new Int64 pointer
func NewInt64(input int64) *Int64 {
	s64 := Int64(input)
	return &s64
}
