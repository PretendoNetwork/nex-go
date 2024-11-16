package types

import (
	"fmt"
)

// List is an implementation of rdv::qList.
// This data type holds an array of other types.
//
// Unlike Buffer and qBuffer, which use the same data type with differing size field lengths,
// there does not seem to be an official rdv::List type
type List[T RVType] []T

// WriteTo writes the List to the given writable
func (l List[T]) WriteTo(writable Writable) {
	writable.WriteUInt32LE(uint32(len(l)))

	for _, v := range l {
		v.WriteTo(writable)
	}
}

func (l List[T]) extractType(t any, readable Readable) error {
	// * This just makes List.ExtractFrom() a bit cleaner
	// * since it doesn't have to type check
	if ptr, ok := t.(RVTypePtr); ok {
		return ptr.ExtractFrom(readable)
	}

	// * Maybe support other types..?

	return fmt.Errorf("Unsupported List type %T", t)
}

// ExtractFrom extracts the List from the given readable
func (l *List[T]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt32LE()
	if err != nil {
		return err
	}

	slice := make([]T, 0, length)

	for i := 0; i < int(length); i++ {
		var value T
		if err := l.extractType(&value, readable); err != nil {
			return err
		}

		slice = append(slice, value)
	}

	*l = slice

	return nil
}

// Copy returns a pointer to a copy of the List. Requires type assertion when used
func (l List[T]) Copy() RVType {
	copied := make(List[T], 0)

	for _, v := range l {
		copied = append(copied, v.Copy().(T))
	}

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (l List[T]) Equals(o RVType) bool {
	if _, ok := o.(List[T]); !ok {
		return false
	}

	other := o.(List[T])

	if len(l) != len(other) {
		return false
	}

	for i := 0; i < len(l); i++ {
		if !l[i].Equals(other[i]) {
			return false
		}
	}

	return true
}

// CopyRef copies the current value of the List
// and returns a pointer to the new copy
func (l List[T]) CopyRef() RVTypePtr {
	return &l
}

// Deref takes a pointer to the List
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (l *List[T]) Deref() RVType {
	return *l
}

// Contains checks if the provided value exists in the List
func (l List[T]) Contains(checkValue T) bool {
	for _, v := range l {
		if v.Equals(checkValue) {
			return true
		}
	}

	return false
}

// String returns a string representation of the struct
func (l List[T]) String() string {
	return fmt.Sprintf("%v", ([]T)(l))
}

// NewList returns a new List of the provided type
func NewList[T RVType]() List[T] {
	return make(List[T], 0)
}
