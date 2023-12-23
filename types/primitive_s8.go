package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveS8 is a type alias of int8 with receiver methods to conform to RVType
type PrimitiveS8 int8 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the int8 to the given writable
func (s8 *PrimitiveS8) WriteTo(writable Writable) {
	writable.WritePrimitiveInt8(int8(*s8))
}

// ExtractFrom extracts the int8 to the given readable
func (s8 *PrimitiveS8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt8()
	if err != nil {
		return err
	}

	*s8 = PrimitiveS8(value)

	return nil
}

// Copy returns a pointer to a copy of the int8. Requires type assertion when used
func (s8 PrimitiveS8) Copy() RVType {
	return &s8
}

// Equals checks if the input is equal in value to the current instance
func (s8 *PrimitiveS8) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS8); !ok {
		return false
	}

	return *s8 == *o.(*PrimitiveS8)
}

// NewPrimitiveS8 returns a new PrimitiveS8
func NewPrimitiveS8(i8 int8) *PrimitiveS8 {
	s8 := PrimitiveS8(i8)

	return &s8
}
