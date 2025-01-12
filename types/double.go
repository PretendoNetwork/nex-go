package types

import "fmt"

// Double is a type alias for the Go basic type float64 for use as an RVType
type Double float64

// WriteTo writes the Double to the given writable
func (d Double) WriteTo(writable Writable) {
	writable.WriteFloat64LE(float64(d))
}

// ExtractFrom extracts the Double value from the given readable
func (d *Double) ExtractFrom(readable Readable) error {
	value, err := readable.ReadFloat64LE()
	if err != nil {
		return err
	}

	*d = Double(value)
	return nil
}

// Copy returns a pointer to a copy of the Double. Requires type assertion when used
func (d Double) Copy() RVType {
	return NewDouble(float64(d))
}

// Equals checks if the input is equal in value to the current instance
func (d Double) Equals(o RVType) bool {
	other, ok := o.(Double)
	if !ok {
		return false
	}

	return d == other
}

// CopyRef copies the current value of the Double
// and returns a pointer to the new copy
func (d Double) CopyRef() RVTypePtr {
	copied := d.Copy().(Double)
	return &copied
}

// Deref takes a pointer to the Double
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (d *Double) Deref() RVType {
	return *d
}

// String returns a string representation of the Double
func (d Double) String() string {
	return fmt.Sprintf("%f", d)
}

// NewDouble returns a new Double
func NewDouble(input float64) Double {
	d := Double(input)
	return d
}
