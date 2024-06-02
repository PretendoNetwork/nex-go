package types

import (
	"fmt"
	"reflect"
)

// List is an implementation of rdv::qList.
// This data type holds an array of other types.
//
// Unlike Buffer and qBuffer, which use the same data type with differing size field lengths,
// there does not seem to be an official rdv::List type
type List[T RVType] []T

// WriteTo writes the List to the given writable
func (l List[T]) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(uint32(len(l)))

	for _, v := range l {
		v.WriteTo(writable)
	}
}

func (l List[T]) newType() T {
	var t T
	tType := reflect.TypeOf(t).Elem()
	return reflect.New(tType).Interface().(T)
}

// ExtractFrom extracts the List from the given readable
func (l *List[T]) ExtractFrom(readable Readable) error {
	length, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return err
	}

	slice := make([]T, 0, length)

	for i := 0; i < int(length); i++ {
		value := l.newType()
		if err := value.ExtractFrom(readable); err != nil {
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

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (l List[T]) Equals(o RVType) bool {
	if _, ok := o.(*List[T]); !ok {
		return false
	}

	other := *o.(*List[T])

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
func NewList[T RVType]() *List[T] {
	l := make(List[T], 0)
	return &l
}
