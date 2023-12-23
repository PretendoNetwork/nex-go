package types

// ClassVersionContainer contains version info for Structures used in verbose RMC messages
type ClassVersionContainer struct {
	Structure
	ClassVersions *Map[*String, *PrimitiveU16]
}

// WriteTo writes the ClassVersionContainer to the given writable
func (cvc *ClassVersionContainer) WriteTo(writable Writable) {
	cvc.ClassVersions.WriteTo(writable)
}

// ExtractFrom extracts the ClassVersionContainer to the given readable
func (cvc *ClassVersionContainer) ExtractFrom(readable Readable) error {
	cvc.ClassVersions = NewMap[*String, *PrimitiveU16]()
	cvc.ClassVersions.KeyType = NewString("")
	cvc.ClassVersions.ValueType = NewPrimitiveU16(0)

	return cvc.ClassVersions.ExtractFrom(readable)
}

// Copy returns a pointer to a copy of the ClassVersionContainer. Requires type assertion when used
func (cvc *ClassVersionContainer) Copy() RVType {
	copied := NewClassVersionContainer()
	copied.ClassVersions = cvc.ClassVersions.Copy().(*Map[*String, *PrimitiveU16])

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (cvc *ClassVersionContainer) Equals(o RVType) bool {
	if _, ok := o.(*ClassVersionContainer); !ok {
		return false
	}

	return cvc.ClassVersions.Equals(o)
}

// NewClassVersionContainer returns a new ClassVersionContainer
func NewClassVersionContainer() *ClassVersionContainer {
	return &ClassVersionContainer{}
}
