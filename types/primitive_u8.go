package types


// PrimitiveU8 is a struct of uint8 with receiver methods to conform to RVType
type PrimitiveU8 struct {
	Value uint8
}

// WriteTo writes the uint8 to the given writable
func (u8 *PrimitiveU8) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt8(u8.Value)
}

// ExtractFrom extracts the uint8 to the given readable
func (u8 *PrimitiveU8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt8()
	if err != nil {
		return err
	}

	u8.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint8. Requires type assertion when used
func (u8 *PrimitiveU8) Copy() RVType {
	return NewPrimitiveU8(u8.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u8 *PrimitiveU8) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU8); !ok {
		return false
	}

	return u8.Value == o.(*PrimitiveU8).Value
}

// NewPrimitiveU8 returns a new PrimitiveU8
func NewPrimitiveU8(ui8 uint8) *PrimitiveU8 {
	return &PrimitiveU8{Value: ui8}
}
