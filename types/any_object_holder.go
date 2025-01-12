package types

import (
	"fmt"
	"strings"
)

// HoldableObject defines a common interface for types which can be placed in AnyObjectHolder
type HoldableObject interface {
	RVType
	ObjectID() RVType // Returns the object identifier of the type
}

// AnyObjectHolderObjects holds a mapping of RVTypes that are accessible in a AnyDataHolder
var AnyObjectHolderObjects = make(map[RVType]HoldableObject)

// RegisterObjectHolderType registers a RVType to be accessible in a AnyDataHolder
func RegisterObjectHolderType(rvType HoldableObject) {
	AnyObjectHolderObjects[rvType.ObjectID()] = rvType
}

// AnyObjectHolder can hold a reference to any RVType which can be held
type AnyObjectHolder[T HoldableObject] struct {
	Object T
}

// WriteTo writes the AnyObjectHolder to the given writable
func (aoh AnyObjectHolder[T]) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	aoh.Object.WriteTo(contentWritable)

	objectBuffer := NewBuffer(contentWritable.Bytes())

	objectBufferLength := uint32(len(objectBuffer) + 4) // * Length of the Buffer

	aoh.Object.ObjectID().WriteTo(writable)
	writable.WriteUInt32LE(objectBufferLength)
	objectBuffer.WriteTo(writable)
}

// ExtractFrom extracts the AnyObjectHolder from the given readable
func (aoh *AnyObjectHolder[T]) ExtractFrom(readable Readable) error {
	var err error

	// TODO - This assumes the identifier is a String
	identifier := NewString("")
	err = identifier.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder identifier. %s", err.Error())
	}

	length := NewUInt32(0)
	err = length.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder length. %s", err.Error())
	}

	// * This is technically a Buffer, but we can't instantiate a new Readable from here so interpret it as a UInt32 and the object data
	bufferLength := NewUInt32(0)
	err = bufferLength.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder buffer length. %s", err.Error())
	}

	if _, ok := AnyObjectHolderObjects[identifier]; !ok {
		return fmt.Errorf("Unknown AnyObjectHolder identifier: %s", identifier)
	}

	ptr := AnyObjectHolderObjects[identifier].CopyRef()

	if err := ptr.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read AnyObjectHolder object. %s", err.Error())
	}

	var ok bool
	if aoh.Object, ok = ptr.Deref().(T); !ok {
		return fmt.Errorf("Input AnyObjectHolder object %s is invalid", identifier)
	}

	return nil
}

// Copy returns a new copied instance of AnyObjectHolder
func (aoh AnyObjectHolder[T]) Copy() RVType {
	copied := NewAnyObjectHolder[T]()

	copied.Object = aoh.Object.Copy().(T)

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (aoh AnyObjectHolder[T]) Equals(o RVType) bool {
	if _, ok := o.(AnyObjectHolder[T]); !ok {
		return false
	}

	other := o.(AnyObjectHolder[T])

	return aoh.Object.Equals(other.Object)
}

// CopyRef copies the current value of the AnyObjectHolder
// and returns a pointer to the new copy
func (aoh AnyObjectHolder[T]) CopyRef() RVTypePtr {
	copied := aoh.Copy().(AnyObjectHolder[T])
	return &copied
}

// Deref takes a pointer to the AnyObjectHolder
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (aoh *AnyObjectHolder[T]) Deref() RVType {
	return *aoh
}

// String returns a string representation of the struct
func (aoh AnyObjectHolder[T]) String() string {
	return aoh.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (aoh AnyObjectHolder[T]) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("AnyDataHolder{\n")
	b.WriteString(fmt.Sprintf("%sIdentifier: %s,\n", indentationValues, aoh.Object.ObjectID()))
	b.WriteString(fmt.Sprintf("%sObject: %s\n", indentationValues, aoh.Object))

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewAnyObjectHolder returns a new AnyObjectHolder
func NewAnyObjectHolder[T HoldableObject]() AnyObjectHolder[T] {
	return AnyObjectHolder[T]{}
}
