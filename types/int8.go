package types

import "fmt"

// Int8 is a type alias for the Go basic type int8 for use as an RVType
type Int8 int8

// WriteTo writes the Int8 to the given writable
func (s8 Int8) WriteTo(writable Writable) {
	writable.WritePrimitiveInt8(int8(s8))
}

// ExtractFrom extracts the Int8 value from the given readable
func (s8 *Int8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt8()
	if err != nil {
		return err
	}

	*s8 = Int8(value)
	return nil
}

// Copy returns a pointer to a copy of the Int8. Requires type assertion when used
func (s8 Int8) Copy() RVType {
	return &s8
}

// Equals checks if the input is equal in value to the current instance
func (s8 Int8) Equals(o RVType) bool {
	other, ok := o.(*Int8)
	if !ok {
		return false
	}
	return s8 == *other
}

// String returns a string representation of the Int8
func (s8 Int8) String() string {
	return fmt.Sprintf("%d", s8)
}

// NewInt8 returns a new Int8 pointer
func NewInt8(input int8) *Int8 {
	s8 := Int8(input)
	return &s8
}

