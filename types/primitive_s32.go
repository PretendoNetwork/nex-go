package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveS32 is a type alias of int32 with receiver methods to conform to RVType
type PrimitiveS32 int32 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the int32 to the given writable
func (s32 *PrimitiveS32) WriteTo(writable Writable) {
	writable.WritePrimitiveInt32LE(int32(*s32))
}

// ExtractFrom extracts the int32 to the given readable
func (s32 *PrimitiveS32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt32LE()
	if err != nil {
		return err
	}

	*s32 = PrimitiveS32(value)

	return nil
}

// Copy returns a pointer to a copy of the int32. Requires type assertion when used
func (s32 *PrimitiveS32) Copy() RVType {
	copied := PrimitiveS32(*s32)

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (s32 *PrimitiveS32) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS32); !ok {
		return false
	}

	return *s32 == *o.(*PrimitiveS32)
}

// NewPrimitiveS32 returns a new PrimitiveS32
func NewPrimitiveS32(i32 int32) *PrimitiveS32 {
	s32 := PrimitiveS32(i32)

	return &s32
}
