package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveU16 is a type alias of uint16 with receiver methods to conform to RVType
type PrimitiveU16 uint16 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the uint16 to the given writable
func (u16 *PrimitiveU16) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt16LE(uint16(*u16))
}

// ExtractFrom extracts the uint16 to the given readable
func (u16 *PrimitiveU16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt16LE()
	if err != nil {
		return err
	}

	*u16 = PrimitiveU16(value)

	return nil
}

// Copy returns a pointer to a copy of the uint16. Requires type assertion when used
func (u16 PrimitiveU16) Copy() RVType {
	return &u16
}

// Equals checks if the input is equal in value to the current instance
func (u16 *PrimitiveU16) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU16); !ok {
		return false
	}

	return *u16 == *o.(*PrimitiveU16)
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewPrimitiveU16 returns a new PrimitiveU16
func NewPrimitiveU16() *PrimitiveU16 {
	var u16 PrimitiveU16
	return &u16
}
