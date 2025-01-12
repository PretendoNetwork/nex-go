package types

import "fmt"

// UInt8 is a type alias for the Go basic type uint8 for use as an RVType
type UInt8 uint8

// WriteTo writes the UInt8 to the given writable
func (u8 UInt8) WriteTo(writable Writable) {
	writable.WriteUInt8(uint8(u8))
}

// ExtractFrom extracts the UInt8 value from the given readable
func (u8 *UInt8) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt8()
	if err != nil {
		return err
	}

	*u8 = UInt8(value)
	return nil
}

// Copy returns a pointer to a copy of the UInt8. Requires type assertion when used
func (u8 UInt8) Copy() RVType {
	return NewUInt8(uint8(u8))
}

// Equals checks if the input is equal in value to the current instance
func (u8 UInt8) Equals(o RVType) bool {
	other, ok := o.(UInt8)
	if !ok {
		return false
	}

	return u8 == other
}

// CopyRef copies the current value of the UInt8
// and returns a pointer to the new copy
func (u8 UInt8) CopyRef() RVTypePtr {
	copied := u8.Copy().(UInt8)
	return &copied
}

// Deref takes a pointer to the UInt8
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (u8 *UInt8) Deref() RVType {
	return *u8
}

// String returns a string representation of the UInt8
func (u8 UInt8) String() string {
	return fmt.Sprintf("%d", u8)
}

// NewUInt8 returns a new UInt8
func NewUInt8(input uint8) UInt8 {
	u8 := UInt8(input)
	return u8
}
