package types

// PrimitiveS64 is a struct of int64 with receiver methods to conform to RVType
type PrimitiveS64 struct {
	Value int64
}

// WriteTo writes the int64 to the given writable
func (s64 *PrimitiveS64) WriteTo(writable Writable) {
	writable.WritePrimitiveInt64LE(s64.Value)
}

// ExtractFrom extracts the int64 to the given readable
func (s64 *PrimitiveS64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveInt64LE()
	if err != nil {
		return err
	}

	s64.Value = value

	return nil
}

// Copy returns a pointer to a copy of the int64. Requires type assertion when used
func (s64 *PrimitiveS64) Copy() RVType {
	return NewPrimitiveS64(s64.Value)
}

// Equals checks if the input is equal in value to the current instance
func (s64 *PrimitiveS64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveS64); !ok {
		return false
	}

	return s64.Value == o.(*PrimitiveS64).Value
}

// NewPrimitiveS64 returns a new PrimitiveS64
func NewPrimitiveS64(i64 int64) *PrimitiveS64 {
	return &PrimitiveS64{Value: i64}
}
