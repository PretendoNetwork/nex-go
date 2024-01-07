package types

import "fmt"

// PrimitiveF64 is wrapper around a Go primitive float64 with receiver methods to conform to RVType
type PrimitiveF64 struct {
	Value float64
}

// WriteTo writes the float64 to the given writable
func (f64 *PrimitiveF64) WriteTo(writable Writable) {
	writable.WritePrimitiveFloat64LE(f64.Value)
}

// ExtractFrom extracts the float64 from the given readable
func (f64 *PrimitiveF64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveFloat64LE()
	if err != nil {
		return err
	}

	f64.Value = value

	return nil
}

// Copy returns a pointer to a copy of the float64. Requires type assertion when used
func (f64 *PrimitiveF64) Copy() RVType {
	return NewPrimitiveF64(f64.Value)
}

// Equals checks if the input is equal in value to the current instance
func (f64 *PrimitiveF64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveF64); !ok {
		return false
	}

	return *f64 == *o.(*PrimitiveF64)
}

// String returns a string representation of the struct
func (f64 *PrimitiveF64) String() string {
	return fmt.Sprintf("%f", f64.Value)
}

// NewPrimitiveF64 returns a new PrimitiveF64
func NewPrimitiveF64(float float64) *PrimitiveF64 {
	return &PrimitiveF64{Value: float}
}
