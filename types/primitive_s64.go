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
func (s64 *PrimitiveS64) Copy() RVType {
	copied := PrimitiveS64(*s64)

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (s64 *PrimitiveS64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS64); !ok {
		return false
	}

	return *s64 == *o.(*PrimitiveS64)
}

// NewPrimitiveS64 returns a new PrimitiveS64
func NewPrimitiveS64(i64 int64) *PrimitiveS64 {
	s64 := PrimitiveS64(i64)

	return &s64
}
