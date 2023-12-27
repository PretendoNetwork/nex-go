package types

// PrimitiveU32 is a struct of uint32 with receiver methods to conform to RVType
type PrimitiveU32 struct {
	Value uint32
}

// WriteTo writes the uint32 to the given writable
func (u32 *PrimitiveU32) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(u32.Value)
}

// ExtractFrom extracts the uint32 from the given readable
func (u32 *PrimitiveU32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return err
	}

	u32.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint32. Requires type assertion when used
func (u32 *PrimitiveU32) Copy() RVType {
	return NewPrimitiveU32(u32.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u32 *PrimitiveU32) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU32); !ok {
		return false
	}

	return u32.Value == o.(*PrimitiveU32).Value
}

// NewPrimitiveU32 returns a new PrimitiveU32
func NewPrimitiveU32(ui32 uint32) *PrimitiveU32 {
	return &PrimitiveU32{Value: ui32}
}
