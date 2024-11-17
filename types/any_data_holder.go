package types

import (
	"errors"
	"fmt"
	"strings"
)

// AnyDataHolderObjects holds a mapping of RVTypes that are accessible in a AnyDataHolder
var AnyDataHolderObjects = make(map[string]RVType)

// RegisterDataHolderType registers a RVType to be accessible in a AnyDataHolder
func RegisterDataHolderType(name string, rvType RVType) {
	AnyDataHolderObjects[name] = rvType
}

// AnyDataHolder is a class which can contain any Structure. The official type name and namespace is unknown.
// These Structures usually inherit from at least one other Structure. Typically this base class is the empty
// `Data` Structure, but this is not always the case. The contained Structures name & length are sent with the
// Structure body, so the receiver can properly decode it.
type AnyDataHolder struct {
	TypeName   String `json:"type_name" db:"type_name" bson:"type_name" xml:"TypeName"`
	Length1    UInt32 `json:"length1" db:"length1" bson:"length1" xml:"Length1"` // * Length of ObjectData + Length2
	Length2    UInt32 `json:"length2" db:"length2" bson:"length2" xml:"Length2"` // * Length of ObjectData
	ObjectData RVType `json:"object_data" db:"object_data" bson:"object_data" xml:"ObjectData"`
}

// WriteTo writes the AnyDataHolder to the given writable
func (adh AnyDataHolder) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	adh.ObjectData.WriteTo(contentWritable)

	objectData := contentWritable.Bytes()
	length1 := uint32(len(objectData) + 4)
	length2 := uint32(len(objectData))

	adh.TypeName.WriteTo(writable)
	writable.WriteUInt32LE(length1)
	writable.WriteUInt32LE(length2)
	writable.Write(objectData)
}

// ExtractFrom extracts the AnyDataHolder from the given readable
func (adh *AnyDataHolder) ExtractFrom(readable Readable) error {
	var err error

	err = adh.TypeName.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyDataHolder type name. %s", err.Error())
	}

	err = adh.Length1.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyDataHolder length 1. %s", err.Error())
	}

	err = adh.Length2.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read AnyDataHolder length 2. %s", err.Error())
	}

	typeName := string(adh.TypeName)

	if _, ok := AnyDataHolderObjects[typeName]; !ok {
		return fmt.Errorf("Unknown AnyDataHolder type: %s", typeName)
	}

	adh.ObjectData = AnyDataHolderObjects[typeName].Copy()

	ptr, ok := any(&adh.ObjectData).(RVTypePtr)
	if !ok {
		return errors.New("AnyDataHolder object data is not a valid RVType. Missing ExtractFrom pointer receiver")
	}

	if err := ptr.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read AnyDataHolder object data. %s", err.Error())
	}

	return nil
}

// Copy returns a new copied instance of AnyDataHolder
func (adh AnyDataHolder) Copy() RVType {
	copied := NewAnyDataHolder()

	copied.TypeName = adh.TypeName
	copied.Length1 = adh.Length1.Copy().(UInt32)
	copied.Length2 = adh.Length2.Copy().(UInt32)
	copied.ObjectData = adh.ObjectData.Copy()

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (adh AnyDataHolder) Equals(o RVType) bool {
	if _, ok := o.(AnyDataHolder); !ok {
		return false
	}

	other := o.(AnyDataHolder)

	if !adh.TypeName.Equals(other.TypeName) {
		return false
	}

	if !adh.Length1.Equals(other.Length1) {
		return false
	}

	if !adh.Length2.Equals(other.Length2) {
		return false
	}

	return adh.ObjectData.Equals(other.ObjectData)
}

// CopyRef copies the current value of the AnyDataHolder
// and returns a pointer to the new copy
func (adh AnyDataHolder) CopyRef() RVTypePtr {
	copied := NewAnyDataHolder()

	copied.TypeName = adh.TypeName
	copied.Length1 = adh.Length1.Copy().(UInt32)
	copied.Length2 = adh.Length2.Copy().(UInt32)
	copied.ObjectData = adh.ObjectData.Copy()

	return &copied
}

// Deref takes a pointer to the AnyDataHolder
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (adh *AnyDataHolder) Deref() RVType {
	return *adh
}

// String returns a string representation of the struct
func (adh AnyDataHolder) String() string {
	return adh.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (adh AnyDataHolder) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("AnyDataHolder{\n")
	b.WriteString(fmt.Sprintf("%sTypeName: %s,\n", indentationValues, adh.TypeName))
	b.WriteString(fmt.Sprintf("%sLength1: %s,\n", indentationValues, adh.Length1))
	b.WriteString(fmt.Sprintf("%sLength2: %s,\n", indentationValues, adh.Length2))
	b.WriteString(fmt.Sprintf("%sObjectData: %s\n", indentationValues, adh.ObjectData))

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewAnyDataHolder returns a new AnyDataHolder
func NewAnyDataHolder() AnyDataHolder {
	return AnyDataHolder{
		TypeName: NewString(""),
		Length1:  NewUInt32(0),
		Length2:  NewUInt32(0),
	}
}
