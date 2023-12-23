package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveS16 is a type alias of int16 with receiver methods to conform to RVType
type PrimitiveS16 int16 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the int16 to the given writable
func (s16 *PrimitiveS16) WriteTo(writable Writable) {
	writable.WritePrimitiveInt16LE(int16(*s16))
}

// ExtractFrom extracts the int16 to the given readable
func (s16 *PrimitiveS16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt16LE()
	if err != nil {
		return err
	}

	*s16 = PrimitiveS16(value)

	return nil
}

// Copy returns a pointer to a copy of the int16. Requires type assertion when used
func (s16 *PrimitiveS16) Copy() RVType {
	copied := PrimitiveS16(*s16)

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (s16 *PrimitiveS16) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS16); !ok {
		return false
	}

	return *s16 == *o.(*PrimitiveS16)
}

// NewPrimitiveS16 returns a new PrimitiveS16
func NewPrimitiveS16(i16 int16) *PrimitiveS16 {
	s16 := PrimitiveS16(i16)

	return &s16
}
