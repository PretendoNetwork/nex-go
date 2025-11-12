package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// Float is a type alias for the Go basic type float32 for use as an RVType
type Float float32

// WriteTo writes the Float to the given writable
func (f Float) WriteTo(writable Writable) {
	writable.WriteFloat32LE(float32(f))
}

// ExtractFrom extracts the Float value from the given readable
func (f *Float) ExtractFrom(readable Readable) error {
	value, err := readable.ReadFloat32LE()
	if err != nil {
		return err
	}

	*f = Float(value)
	return nil
}

// Copy returns a pointer to a copy of the Float. Requires type assertion when used
func (f Float) Copy() RVType {
	return NewFloat(float32(f))
}

// Equals checks if the input is equal in value to the current instance
func (f Float) Equals(o RVType) bool {
	other, ok := o.(Float)
	if !ok {
		return false
	}

	return f == other
}

// CopyRef copies the current value of the Float
// and returns a pointer to the new copy
func (f Float) CopyRef() RVTypePtr {
	copied := f.Copy().(Float)
	return &copied
}

// Deref takes a pointer to the Float
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (f *Float) Deref() RVType {
	return *f
}

// String returns a string representation of the Float
func (f Float) String() string {
	return fmt.Sprintf("%f", f)
}

// Scan implements sql.Scanner for database/sql
func (f *Float) Scan(value any) error {
	if value == nil {
		*f = Float(0.0)
		return nil
	}

	switch v := value.(type) {
	case float64:
		*f = Float(value.(float64))
	case string:
		parsed, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return fmt.Errorf("cannot parse string %q into Float: %w", v, err)
		}
		*f = Float(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Float", value)
	}

	return nil
}

// Value implements driver.Valuer for database/sql
func (f Float) Value() (driver.Value, error) {
	return float64(f), nil
}

// NewFloat returns a new Float
func NewFloat(input float32) Float {
	f := Float(input)
	return f
}
