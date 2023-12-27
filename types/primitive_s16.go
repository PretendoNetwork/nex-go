package types

// PrimitiveS16 is a struct of int16 with receiver methods to conform to RVType
type PrimitiveS16 struct {
	Value int16
}

// WriteTo writes the int16 to the given writable
func (s16 *PrimitiveS16) WriteTo(writable Writable) {
	writable.WritePrimitiveInt16LE(s16.Value)
}

// ExtractFrom extracts the int16 to the given readable
func (s16 *PrimitiveS16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt16LE()
	if err != nil {
		return err
	}

	s16.Value = value

	return nil
}

// Copy returns a pointer to a copy of the int16. Requires type assertion when used
func (s16 *PrimitiveS16) Copy() RVType {
	return NewPrimitiveS16(s16.Value)
}

// Equals checks if the input is equal in value to the current instance
func (s16 *PrimitiveS16) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS16); !ok {
		return false
	}

	return s16.Value == o.(*PrimitiveS16).Value
}

// NewPrimitiveS16 returns a new PrimitiveS16
func NewPrimitiveS16(i16 int16) *PrimitiveS16 {
	return &PrimitiveS16{Value: i16}
}
