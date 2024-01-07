package types

import "fmt"

// PrimitiveU64 is wrapper around a Go primitive uint64 with receiver methods to conform to RVType
type PrimitiveU64 struct {
	Value uint64
}

// WriteTo writes the uint64 to the given writable
func (u64 *PrimitiveU64) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt64LE(u64.Value)
}

// ExtractFrom extracts the uint64 from the given readable
func (u64 *PrimitiveU64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt64LE()
	if err != nil {
		return err
	}

	u64.Value = value

	return nil
}

// Copy returns a pointer to a copy of the uint64. Requires type assertion when used
func (u64 *PrimitiveU64) Copy() RVType {
	return NewPrimitiveU64(u64.Value)
}

// Equals checks if the input is equal in value to the current instance
func (u64 *PrimitiveU64) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveU64); !ok {
		return false
	}

	return u64.Value == o.(*PrimitiveU64).Value
}

// String returns a string representation of the struct
func (u64 *PrimitiveU64) String() string {
	return fmt.Sprintf("%d", u64.Value)
}

// NewPrimitiveU64 returns a new PrimitiveU64
func NewPrimitiveU64(ui64 uint64) *PrimitiveU64 {
	return &PrimitiveU64{Value: ui64}
}
