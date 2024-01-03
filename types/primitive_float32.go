package types

import "fmt"

// PrimitiveF32 is a struct of float32 with receiver methods to conform to RVType
type PrimitiveF32 struct {
	Value float32
}

// WriteTo writes the float32 to the given writable
func (f32 *PrimitiveF32) WriteTo(writable Writable) {
	writable.WritePrimitiveFloat32LE(f32.Value)
}

// ExtractFrom extracts the float32 from the given readable
func (f32 *PrimitiveF32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveFloat32LE()
	if err != nil {
		return err
	}

	f32.Value = value

	return nil
}

// Copy returns a pointer to a copy of the float32. Requires type assertion when used
func (f32 *PrimitiveF32) Copy() RVType {
	return NewPrimitiveF32(f32.Value)
}

// Equals checks if the input is equal in value to the current instance
func (f32 *PrimitiveF32) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveF32); !ok {
		return false
	}

	return f32.Value == o.(*PrimitiveF32).Value
}

// String returns a string representation of the struct
func (f32 *PrimitiveF32) String() string {
	return fmt.Sprintf("%f", f32.Value)
}

// NewPrimitiveF32 returns a new PrimitiveF32
func NewPrimitiveF32(float float32) *PrimitiveF32 {
	return &PrimitiveF32{Value: float}
}
