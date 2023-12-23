package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveU32 is a type alias of uint32 with receiver methods to conform to RVType
type PrimitiveU32 uint32 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the uint32 to the given writable
func (u32 *PrimitiveU32) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(uint32(*u32))
}

// ExtractFrom extracts the uint32 to the given readable
func (u32 *PrimitiveU32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return err
	}

	*u32 = PrimitiveU32(value)

	return nil
}

// Copy returns a pointer to a copy of the uint32. Requires type assertion when used
func (u32 PrimitiveU32) Copy() RVType {
	return &u32
}

// Equals checks if the input is equal in value to the current instance
func (u32 *PrimitiveU32) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU32); !ok {
		return false
	}

	return *u32 == *o.(*PrimitiveU32)
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewPrimitiveU32 returns a new PrimitiveU32
func NewPrimitiveU32() *PrimitiveU32 {
	var u32 PrimitiveU32
	return &u32
}
