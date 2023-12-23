package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveS64 is a type alias of int64 with receiver methods to conform to RVType
type PrimitiveS64 int64 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the int64 to the given writable
func (s64 *PrimitiveS64) WriteTo(writable Writable) {
	writable.WritePrimitiveInt64LE(int64(*s64))
}

// ExtractFrom extracts the int64 to the given readable
func (s64 *PrimitiveS64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt64LE()
	if err != nil {
		return err
	}

	*s64 = PrimitiveS64(value)

	return nil
}

// Copy returns a pointer to a copy of the int64. Requires type assertion when used
func (s64 PrimitiveS64) Copy() RVType {
	return &s64
}

// Equals checks if the input is equal in value to the current instance
func (s64 *PrimitiveS64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS64); !ok {
		return false
	}

	return *s64 == *o.(*PrimitiveS64)
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewPrimitiveS64 returns a new PrimitiveS64
func NewPrimitiveS64() *PrimitiveS64 {
	var s64 PrimitiveS64
	return &s64
}
