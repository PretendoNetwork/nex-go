package types

import (
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
	TypeName   *String
	Length1    *PrimitiveU32 // Length of ObjectData + Length2
	Length2    *PrimitiveU32 // Length of ObjectData
	ObjectData RVType
}

// WriteTo writes the AnyDataHolder to the given writable
func (adh *AnyDataHolder) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	adh.ObjectData.WriteTo(contentWritable)

	objectData := contentWritable.Bytes()
	length1 := uint32(len(objectData) + 4)
	length2 := uint32(len(objectData))

	adh.TypeName.WriteTo(writable)
	writable.WritePrimitiveUInt32LE(length1)
	writable.WritePrimitiveUInt32LE(length2)
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

	if _, ok := AnyDataHolderObjects[adh.TypeName.Value]; !ok {
		return fmt.Errorf("Unknown AnyDataHolder type: %s", adh.TypeName.Value)
	}

	adh.ObjectData = AnyDataHolderObjects[adh.TypeName.Value].Copy()

	if err := adh.ObjectData.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read AnyDataHolder object data. %s", err.Error())
	}

	return nil
}

// Copy returns a new copied instance of AnyDataHolder
func (adh *AnyDataHolder) Copy() RVType {
	copied := NewAnyDataHolder()

	copied.TypeName = adh.TypeName.Copy().(*String)
	copied.Length1 = adh.Length1.Copy().(*PrimitiveU32)
	copied.Length2 = adh.Length2.Copy().(*PrimitiveU32)
	copied.ObjectData = adh.ObjectData.Copy()

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (adh *AnyDataHolder) Equals(o RVType) bool {
	if _, ok := o.(*AnyDataHolder); !ok {
		return false
	}

	other := o.(*AnyDataHolder)

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

// String returns a string representation of the struct
func (adh *AnyDataHolder) String() string {
	return adh.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (adh *AnyDataHolder) FormatToString(indentationLevel int) string {
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
func NewAnyDataHolder() *AnyDataHolder {
	return &AnyDataHolder{
		TypeName: NewString(""),
		Length1:  NewPrimitiveU32(0),
		Length2:  NewPrimitiveU32(0),
	}
}
