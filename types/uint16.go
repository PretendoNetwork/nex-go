package types

import "fmt"

// UInt16 is a type alias for the Go basic type uint16 for use as an RVType
type UInt16 uint16

// WriteTo writes the UInt16 to the given writable
func (u16 UInt16) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt16LE(uint16(u16))
}

// ExtractFrom extracts the UInt16 value from the given readable
func (u16 *UInt16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveUInt16LE()
	if err != nil {
		return err
	}

	*u16 = UInt16(value)
	return nil
}

// Copy returns a pointer to a copy of the UInt16. Requires type assertion when used
func (u16 UInt16) Copy() RVType {
	copy := u16
	return &copy
}

// Equals checks if the input is equal in value to the current instance
func (u16 UInt16) Equals(o RVType) bool {
	other, ok := o.(*UInt16)
	if !ok {
		return false
	}
	return u16 == *other
}

// String returns a string representation of the UInt16
func (u16 UInt16) String() string {
	return fmt.Sprintf("%d", u16)
}

// NewUInt16 returns a new UInt16 pointer
func NewUInt16(input uint16) *UInt16 {
	u16 := UInt16(input)
	return &u16
}
