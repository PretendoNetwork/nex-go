package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveF64 is a type alias of float64 with receiver methods to conform to RVType
type PrimitiveF64 float64 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the float64 to the given writable
func (f64 *PrimitiveF64) WriteTo(writable Writable) {
	writable.WritePrimitiveFloat64LE(float64(*f64))
}

// ExtractFrom extracts the float64 to the given readable
func (f64 *PrimitiveF64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveFloat64LE()
	if err != nil {
		return err
	}

	*f64 = PrimitiveF64(value)

	return nil
}

// Copy returns a pointer to a copy of the float64. Requires type assertion when used
func (f64 PrimitiveF64) Copy() RVType {
	return &f64
}

// Equals checks if the input is equal in value to the current instance
func (f64 *PrimitiveF64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveF64); !ok {
		return false
	}

	return *f64 == *o.(*PrimitiveF64)
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewPrimitiveF64 returns a new PrimitiveF64
func NewPrimitiveF64() *PrimitiveF64 {
	var f64 PrimitiveF64
	return &f64
}
