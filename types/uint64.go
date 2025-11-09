package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// UInt64 is a type alias for the Go basic type uint64 for use as an RVType
type UInt64 uint64

// WriteTo writes the UInt64 to the given writable
func (u64 UInt64) WriteTo(writable Writable) {
	writable.WriteUInt64LE(uint64(u64))
}

// ExtractFrom extracts the UInt64 value from the given readable
func (u64 *UInt64) ExtractFrom(readable Readable) error {
	value, err := readable.ReadUInt64LE()
	if err != nil {
		return err
	}

	*u64 = UInt64(value)
	return nil
}

// Copy returns a pointer to a copy of the UInt64. Requires type assertion when used
func (u64 UInt64) Copy() RVType {
	return NewUInt64(uint64(u64))
}

// Equals checks if the input is equal in value to the current instance
func (u64 UInt64) Equals(o RVType) bool {
	other, ok := o.(UInt64)
	if !ok {
		return false
	}

	return u64 == other
}

// CopyRef copies the current value of the UInt64
// and returns a pointer to the new copy
func (u64 UInt64) CopyRef() RVTypePtr {
	copied := u64.Copy().(UInt64)
	return &copied
}

// Deref takes a pointer to the UInt64
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (u64 *UInt64) Deref() RVType {
	return *u64
}

// String returns a string representation of the UInt64
func (u64 UInt64) String() string {
	return fmt.Sprintf("%d", u64)
}

// Scan implements sql.Scanner for database/sql
func (u64 *UInt64) Scan(value any) error {
	if value == nil {
		*u64 = UInt64(0)
		return nil
	}

	switch v := value.(type) {
	case int64: // * Postgres might store/return the data as an int64
		*u64 = UInt64(uint64(v))
	case []byte, string: // * Otherwise, it's stored/returned as a string
		var str string
		if b, ok := v.([]byte); ok {
			str = string(b)
		} else {
			str = v.(string)
		}

		parsed, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into UInt64: %w", v, err)
		}
		*u64 = UInt64(parsed)
	default:
		return fmt.Errorf("cannot scan %T into UInt64", value)
	}

	return nil
}

// Value implements driver.Valuer for database/sql
func (u64 UInt64) Value() (driver.Value, error) {
	return fmt.Sprintf("%d", uint64(u64)), nil
}

// NewUInt64 returns a new UInt64
func NewUInt64(input uint64) UInt64 {
	u64 := UInt64(input)
	return u64
}
