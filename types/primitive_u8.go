package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveU8 is a type alias of uint8 with receiver methods to conform to RVType
type PrimitiveU8 uint8 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the uint8 to the given writable
func (u8 *PrimitiveU8) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt8(uint8(*u8))
}

// ExtractFrom extracts the uint8 to the given readable
func (u8 *PrimitiveU8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt8()
	if err != nil {
		return err
	}

	*u8 = PrimitiveU8(value)

	return nil
}

// Copy returns a pointer to a copy of the uint8. Requires type assertion when used
func (u8 PrimitiveU8) Copy() RVType {
	return &u8
}

// Equals checks if the input is equal in value to the current instance
func (u8 *PrimitiveU8) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU8); !ok {
		return false
	}

	return *u8 == *o.(*PrimitiveU8)
}

// NewPrimitiveU8 returns a new PrimitiveU8
func NewPrimitiveU8(ui8 uint8) *PrimitiveU8 {
	u8 := PrimitiveU8(ui8)

	return &u8
}
