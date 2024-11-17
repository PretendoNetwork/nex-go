package types

import "fmt"

// UInt32 is a type alias for the Go basic type uint32 for use as an RVType
type UInt32 uint32

// WriteTo writes the UInt32 to the given writable
func (u32 UInt32) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(u32))
}

// ExtractFrom extracts the UInt32 value from the given readable
func (u32 *UInt32) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt32LE()
	if err != nil {
		return err
	}

	*u32 = UInt32(value)
	return nil
}

// Copy returns a pointer to a copy of the UInt32. Requires type assertion when used
func (u32 UInt32) Copy() RVType {
	return NewUInt32(uint32(u32))
}

// Equals checks if the input is equal in value to the current instance
func (u32 UInt32) Equals(o RVType) bool {
	other, ok := o.(UInt32)
	if !ok {
		return false
	}

	return u32 == other
}

// CopyRef copies the current value of the UInt32
// and returns a pointer to the new copy
func (u32 UInt32) CopyRef() RVTypePtr {
	copied := u32.Copy().(UInt32)
	return &copied
}

// Deref takes a pointer to the UInt32
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (u32 *UInt32) Deref() RVType {
	return *u32
}

// String returns a string representation of the UInt32
func (u32 UInt32) String() string {
	return fmt.Sprintf("%d", u32)
}

// NewUInt32 returns a new UInt32
func NewUInt32(input uint32) UInt32 {
	u32 := UInt32(input)
	return u32
}
