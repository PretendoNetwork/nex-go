package types

import "fmt"

// UInt64 is a type alias for the Go basic type uint64 for use as an RVType
type UInt64 uint64

// WriteTo writes the UInt64 to the given writable
func (u64 UInt64) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt64LE(uint64(u64))
}

// ExtractFrom extracts the UInt64 value from the given readable
func (u64 *UInt64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt64LE()
	if err != nil {
		return err
	}

	*u64 = UInt64(value)
	return nil
}

// Copy returns a pointer to a copy of the UInt64. Requires type assertion when used
func (u64 UInt64) Copy() RVType {
	return &u64
}

// Equals checks if the input is equal in value to the current instance
func (u64 UInt64) Equals(o RVType) bool {
	other, ok := o.(*UInt64)
	if !ok {
		return false
	}
	return u64 == *other
}

// String returns a string representation of the UInt64
func (u64 UInt64) String() string {
	return fmt.Sprintf("%d", u64)
}

// NewUInt64 returns a new UInt64 pointer
func NewUInt64(input uint64) *UInt64 {
	u64 := UInt64(input)
	return &u64
}

