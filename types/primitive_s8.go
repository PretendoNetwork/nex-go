package types

import "fmt"

// PrimitiveS8 is a struct of int8 with receiver methods to conform to RVType
type PrimitiveS8 struct {
	Value int8
}

// WriteTo writes the int8 to the given writable
func (s8 *PrimitiveS8) WriteTo(writable Writable) {
	writable.WritePrimitiveInt8(s8.Value)
}

// ExtractFrom extracts the int8 from the given readable
func (s8 *PrimitiveS8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt8()
	if err != nil {
		return err
	}

	s8.Value = value

	return nil
}

// Copy returns a pointer to a copy of the int8. Requires type assertion when used
func (s8 *PrimitiveS8) Copy() RVType {
	return NewPrimitiveS8(s8.Value)
}

// Equals checks if the input is equal in value to the current instance
func (s8 *PrimitiveS8) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS8); !ok {
		return false
	}

	return s8.Value == o.(*PrimitiveS8).Value
}

// String returns a string representation of the struct
func (s8 *PrimitiveS8) String() string {
	return fmt.Sprintf("%d", s8.Value)
}

// NewPrimitiveS8 returns a new PrimitiveS8
func NewPrimitiveS8(i8 int8) *PrimitiveS8 {
	return &PrimitiveS8{Value: i8}
}
