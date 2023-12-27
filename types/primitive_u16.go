package types

// PrimitiveU16 is a struct of uint16 with receiver methods to conform to RVType
type PrimitiveU16 struct {
	Value uint16
}

// WriteTo writes the uint16 to the given writable
func (u16 *PrimitiveU16) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt16LE(u16.Value)
}

// ExtractFrom extracts the uint16 to the given readable
func (u16 *PrimitiveU16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt16LE()
	if err != nil {
		return err
	}

	u16.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint16. Requires type assertion when used
func (u16 *PrimitiveU16) Copy() RVType {
	return NewPrimitiveU16(u16.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u16 *PrimitiveU16) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU16); !ok {
		return false
	}

	return u16.Value == o.(*PrimitiveU16).Value
}

// NewPrimitiveU16 returns a new PrimitiveU16
func NewPrimitiveU16(ui16 uint16) *PrimitiveU16 {
	return &PrimitiveU16{Value: ui16}
}
