package types

// TODO - Should this have a "Value"-kind of method to get the original value?

// PrimitiveBool is a type alias of bool with receiver methods to conform to RVType
type PrimitiveBool bool // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the bool to the given writable
func (b *PrimitiveBool) WriteTo(writable Writable) {
	writable.WritePrimitiveBool(bool(*b))
}

// ExtractFrom extracts the bool to the given readable
func (b *PrimitiveBool) ExtractFrom(readable Readable) error {
	value, err := readable.ReadPrimitiveBool()
	if err != nil {
		return err
	}

	*b = PrimitiveBool(value)

	return nil
}

// Copy returns a pointer to a copy of the PrimitiveBool. Requires type assertion when used
func (b *PrimitiveBool) Copy() RVType {
	copied := PrimitiveBool(*b)

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (b *PrimitiveBool) Equals(o RVType) bool {
	if _, ok := o.(*PrimitiveBool); !ok {
		return false
	}

	return *b == *o.(*PrimitiveBool)
}

// NewPrimitiveBool returns a new PrimitiveBool
func NewPrimitiveBool(boolean bool) *PrimitiveBool {
	b := PrimitiveBool(boolean)

	return &b
}
