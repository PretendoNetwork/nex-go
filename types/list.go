package types

import "errors"

// List represents a Quazal Rendez-Vous/NEX List type
type List[T RVType] struct {
	real []T
	Type T
}

// WriteTo writes the bool to the given writable
func (l *List[T]) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(uint32(len(l.real)))

	for _, v := range l.real {
		v.WriteTo(writable)
	}
}

// ExtractFrom extracts the bool to the given readable
func (l *List[T]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return err
	}

	slice := make([]T, 0, length)

	for i := 0; i < int(length); i++ {
		value := l.Type.Copy()
		if err := value.ExtractFrom(readable); err != nil {
			return err
		}

		slice = append(slice, value.(T))
	}

	l.real = slice

	return nil
}

// Copy returns a pointer to a copy of the List[T]. Requires type assertion when used
func (l *List[T]) Copy() RVType {
	copied := NewList[T]()
	copied.real = make([]T, len(l.real))
	copied.Type = l.Type.Copy().(T)

	for i, v := range l.real {
		copied.real[i] = v.Copy().(T)
	}

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (l *List[T]) Equals(o RVType) bool {
	if _, ok := o.(*List[T]); !ok {
		return false
	}

	other := o.(*List[T])

	if len(l.real) != len(other.real) {
		return false
	}

	for i := 0; i < len(l.real); i++ {
		if !l.real[i].Equals(other.real[i]) {
			return false
		}
	}

	return true
}

// Append appends an element to the List internal slice
func (l *List[T]) Append(value T) {
	l.real = append(l.real, value)
}

// Get returns an element at the given index. Returns an error if the index is OOB
func (l *List[T]) Get(index int) (T, error) {
	if index < 0 || index >= len(l.real) {
		return l.Type.Copy().(T), errors.New("Index out of bounds")
	}

	return l.real[index], nil
}

// SetFromData sets the List's internal slice to the input data
func (l *List[T]) SetFromData(data []T) {
	l.real = data
}

// NewList returns a new List of the provided type
func NewList[T RVType]() *List[T] {
	return &List[T]{real: make([]T, 0)}
}
