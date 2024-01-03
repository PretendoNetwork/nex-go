package types

import "fmt"

// PrimitiveBool is a struct of bool with receiver methods to conform to RVType
type PrimitiveBool struct {
	Value bool
}

// WriteTo writes the bool to the given writable
func (b *PrimitiveBool) WriteTo(writable Writable) {
	writable.WritePrimitiveBool(b.Value)
}

// ExtractFrom extracts the bool from the given readable
func (b *PrimitiveBool) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveBool()
	if err != nil {
		return err
	}

	b.Value = value

	return nil
}

// Copy returns a pointer to a copy of the PrimitiveBool. Requires type assertion when used
func (b *PrimitiveBool) Copy() RVType {
	return NewPrimitiveBool(b.Value)
}

// Equals checks if the input is equal in value to the current instance
func (b *PrimitiveBool) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveBool); !ok {
		return false
	}

	return b.Value == o.(*PrimitiveBool).Value
}

// String returns a string representation of the struct
func (b *PrimitiveBool) String() string {
	return fmt.Sprintf("%t", b.Value)
}

// NewPrimitiveBool returns a new PrimitiveBool
func NewPrimitiveBool(boolean bool) *PrimitiveBool {
	return &PrimitiveBool{Value: boolean}
}
