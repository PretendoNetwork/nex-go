package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveF32 is a type alias of float32 with receiver methods to conform to RVType
type PrimitiveF32 float32 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the float32 to the given writable
func (f32 *PrimitiveF32) WriteTo(writable Writable) {
	writable.WritePrimitiveFloat32LE(float32(*f32))
}

// ExtractFrom extracts the float32 to the given readable
func (f32 *PrimitiveF32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveFloat32LE()
	if err != nil {
		return err
	}

	*f32 = PrimitiveF32(value)

	return nil
}

// Copy returns a pointer to a copy of the float32. Requires type assertion when used
func (f32 *PrimitiveF32) Copy() RVType {
	copied := PrimitiveF32(*f32)

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (f32 *PrimitiveF32) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveF32); !ok {
		return false
	}

	return *f32 == *o.(*PrimitiveF32)
}

// NewPrimitiveF32 returns a new PrimitiveF32
func NewPrimitiveF32(float float32) *PrimitiveF32 {
	f32 := PrimitiveF32(float)

	return &f32
}
