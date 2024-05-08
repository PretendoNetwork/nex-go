package types

import (
	"errors"
	"fmt"
)

// List is an implementation of rdv::qList.
// This data type holds an array of other types.
//
// Unlike Buffer and qBuffer, which use the same data type with differing size field lengths,
// there does not seem to be an official rdv::List type
type List[T RVType] struct {
	real []T
	Type T
}

// WriteTo writes the List to the given writable
func (l *List[T]) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(uint32(len(l.real)))

	for _, v := range l.real {
		v.WriteTo(writable)
	}
}

// ExtractFrom extracts the List from the given readable
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

// Copy returns a pointer to a copy of the List. Requires type assertion when used
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

// Slice returns the real underlying slice for the List
func (l *List[T]) Slice() []T {
	return l.real
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

// SetIndex sets a value in the List at the given index
func (l *List[T]) SetIndex(index int, value T) error {
	if index < 0 || index >= len(l.real) {
		return errors.New("Index out of bounds")
	}

	l.real[index] = value

	return nil
}

// DeleteIndex deletes an element at the given index. Returns an error if the index is OOB
func (l *List[T]) DeleteIndex(index int) error {
	if index < 0 || index >= len(l.real) {
		return errors.New("Index out of bounds")
	}

	l.real = append(l.real[:index], l.real[index+1:]...)

	return nil
}

// Remove removes the first occurance of the input from the List. Returns an error if the index is OOB
func (l *List[T]) Remove(check T) {
	for i, value := range l.real {
		if value.Equals(check) {
			l.DeleteIndex(i)
			return
		}
	}
}

// SetFromData sets the List's internal slice to the input data
func (l *List[T]) SetFromData(data []T) {
	l.real = data
}

// Length returns the number of elements in the List
func (l *List[T]) Length() int {
	return len(l.real)
}

// Each runs a callback function for every element in the List
// The List should not be modified inside the callback function
// Returns true if the loop was terminated early
func (l *List[T]) Each(callback func(i int, value T) bool) bool {
	for i, value := range l.real {
		if callback(i, value) {
			return true
		}
	}

	return false
}

// Contains checks if the provided value exists in the List
func (l *List[T]) Contains(checkValue T) bool {
	contains := false

	l.Each(func(_ int, value T) bool {
		if value.Equals(checkValue) {
			contains = true

			return true
		}

		return false
	})

	return contains
}

// String returns a string representation of the struct
func (l *List[T]) String() string {
	return fmt.Sprintf("%v", l.real)
}

// NewList returns a new List of the provided type
func NewList[T RVType]() *List[T] {
	return &List[T]{real: make([]T, 0)}
}

// TransformList applies closure f to each element of the List, returning the result as a slice
func TransformList[T RVType, R any](l *List[T], f func(T) R) []R {
	result := make([]R, l.Length())

	for i, v := range l.real {
		result[i] = f(v)
	}

	return result
}

// NewListTransformed applies closure f to each element of a slice, creating a List from the results
func NewListTransformed[T RVType, R any](l []R, f func(R) T) *List[T] {
	result := make([]T, len(l))

	for i, v := range l {
		result[i] = f(v)
	}

	nexlist := NewList[T]()
	nexlist.SetFromData(result)
	return nexlist
}