package types

import "fmt"

// Int16 is a type alias for the Go basic type int16 for use as an RVType
type Int16 int16

// WriteTo writes the Int16 to the given writable
func (s16 Int16) WriteTo(writable Writable) {
	writable.WritePrimitiveInt16LE(int16(s16))
}

// ExtractFrom extracts the Int16 value from the given readable
func (s16 *Int16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt16LE()
	if err != nil {
		return err
	}

	*s16 = Int16(value)
	return nil
}

// Copy returns a pointer to a copy of the Int16. Requires type assertion when used
func (s16 Int16) Copy() RVType {
	copy := s16
	return &copy
}

// Equals checks if the input is equal in value to the current instance
func (s16 Int16) Equals(o RVType) bool {
	other, ok := o.(*Int16)
	if !ok {
		return false
	}
	return s16 == *other
}

// String returns a string representation of the Int16
func (s16 Int16) String() string {
	return fmt.Sprintf("%d", s16)
}

// NewInt16 returns a new Int16 pointer
func NewInt16(input int16) *Int16 {
	s16 := Int16(input)
	return &s16
}
