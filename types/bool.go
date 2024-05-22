package types

import "fmt"

// Bool is a type alias for the Go basic type bool for use as an RVType
type Bool bool

// WriteTo writes the Bool to the given writable
func (b Bool) WriteTo(writable Writable) {
	writable.WritePrimitiveBool(bool(b))
}

// ExtractFrom extracts the Bool value from the given readable
func (b *Bool) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveBool()
	if err != nil {
		return err
	}

	*b = Bool(value)
	return nil
}

// Copy returns a pointer to a copy of the Bool. Requires type assertion when used
func (b Bool) Copy() RVType {
	copy := b
	return &copy
}

// Equals checks if the input is equal in value to the current instance
func (b Bool) Equals(o RVType) bool {
	other, ok := o.(*Bool)
	if !ok {
		return false
	}
	return b == *other
}

// String returns a string representation of the Bool
func (b Bool) String() string {
	return fmt.Sprintf("%t", b)
}

// NewBool returns a new Bool pointer
func NewBool(input bool) *Bool {
	b := Bool(input)
	return &b
}
