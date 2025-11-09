package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// Int16 is a type alias for the Go basic type int16 for use as an RVType
type Int16 int16

// WriteTo writes the Int16 to the given writable
func (i16 Int16) WriteTo(writable Writable) {
	writable.WriteInt16LE(int16(i16))
}

// ExtractFrom extracts the Int16 value from the given readable
func (i16 *Int16) ExtractFrom(readable Readable) error {
	value, err := readable.ReadInt16LE()
	if err != nil {
		return err
	}

	*i16 = Int16(value)
	return nil
}

// Copy returns a pointer to a copy of the Int16. Requires type assertion when used
func (i16 Int16) Copy() RVType {
	return NewInt16(int16(i16))
}

// Equals checks if the input is equal in value to the current instance
func (i16 Int16) Equals(o RVType) bool {
	other, ok := o.(Int16)
	if !ok {
		return false
	}

	return i16 == other
}

// CopyRef copies the current value of the Int16
// and returns a pointer to the new copy
func (i16 Int16) CopyRef() RVTypePtr {
	copied := i16.Copy().(Int16)
	return &copied
}

// Deref takes a pointer to the Int16
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (i16 *Int16) Deref() RVType {
	return *i16
}

// String returns a string representation of the Int16
func (i16 Int16) String() string {
	return fmt.Sprintf("%d", i16)
}

// Scan implements sql.Scanner for database/sql
func (i16 *Int16) Scan(value any) error {
	if value == nil {
		*i16 = Int16(0)
		return nil
	}

	switch v := value.(type) {
	case int64:
		*i16 = Int16(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Int16: %w", v, err)
		}
		*i16 = Int16(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Int16", value)
	}

	return nil
}

// Value implements driver.Valuer for database/sql
func (i16 Int16) Value() (driver.Value, error) {
	return int64(i16), nil
}

// NewInt16 returns a new Int16
func NewInt16(input int16) Int16 {
	i16 := Int16(input)
	return i16
}
