package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveU64 is a type alias of uint64 with receiver methods to conform to RVType
type PrimitiveU64 uint64 // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the uint64 to the given writable
func (u64 *PrimitiveU64) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt64LE(uint64(*u64))
}

// ExtractFrom extracts the uint64 to the given readable
func (u64 *PrimitiveU64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt64LE()
	if err != nil {
		return err
	}

	*u64 = PrimitiveU64(value)

	return nil
}

// Copy returns a pointer to a copy of the uint64. Requires type assertion when used
func (u64 PrimitiveU64) Copy() RVType {
	return &u64
}

// Equals checks if the input is equal in value to the current instance
func (u64 *PrimitiveU64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU64); !ok {
		return false
	}

	return *u64 == *o.(*PrimitiveU64)
}

// NewPrimitiveU64 returns a new PrimitiveU64
func NewPrimitiveU64(ui64 uint64) *PrimitiveU64 {
	u64 := PrimitiveU64(ui64)

	return &u64
}
