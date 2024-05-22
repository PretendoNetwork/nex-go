package types

import "fmt"

// Double is a type alias for the Go basic type float64 for use as an RVType
type Double float64

// WriteTo writes the Double to the given writable
func (d Double) WriteTo(writable Writable) {
	writable.WritePrimitiveFloat64LE(float64(d))
}

// ExtractFrom extracts the Double value from the given readable
func (d *Double) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveFloat64LE()
	if err != nil {
		return err
	}

	*d = Double(value)
	return nil
}

// Copy returns a pointer to a copy of the Double. Requires type assertion when used
func (d Double) Copy() RVType {
	copy := d
	return &copy
}

// Equals checks if the input is equal in value to the current instance
func (d Double) Equals(o RVType) bool {
	other, ok := o.(*Double)
	if !ok {
		return false
	}
	return d == *other
}

// String returns a string representation of the Double
func (d Double) String() string {
	return fmt.Sprintf("%f", d)
}

// NewDouble returns a new Double pointer
func NewDouble(input float64) *Double {
	d := Double(input)
	return &d
}
