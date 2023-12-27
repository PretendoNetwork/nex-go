package types

import (
	"fmt"
)

// AnyDataHolderObjects holds a mapping of RVTypes that are accessible in a AnyDataHolder
var AnyDataHolderObjects = make(map[string]RVType)

// RegisterDataHolderType registers a RVType to be accessible in a AnyDataHolder
func RegisterDataHolderType(name string, rvType RVType) {
	AnyDataHolderObjects[name] = rvType
}

// AnyDataHolder is a class which can contain any Structure. These Structures usually inherit from at least one
// other Structure. Typically this base class is the empty `Data` Structure, but this is not always the case.
// The contained Structures name & length are sent with the Structure body, so the receiver can properly decode it
type AnyDataHolder struct {
	TypeName   string
	Length1    uint32
	Length2    uint32
	ObjectData RVType
}

// WriteTo writes the AnyDataholder to the given writable
func (adh *AnyDataHolder) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	adh.ObjectData.WriteTo(contentWritable)

	objectData := contentWritable.Bytes()
	typeName := NewString(adh.TypeName)
	length1 := uint32(len(objectData) + 4)
	length2 := uint32(len(objectData))

	typeName.WriteTo(writable)
	writable.WritePrimitiveUInt32LE(length1)
	writable.WritePrimitiveUInt32LE(length2)
	writable.Write(objectData)
}

// ExtractFrom extracts the AnyDataholder to the given readable
func (adh *AnyDataHolder) ExtractFrom(readable Readable) error {
	typeName := NewString("")

	err := typeName.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read DanyDataHolder type name. %s", err.Error())
	}

	length1, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read DanyDataHolder length 1. %s", err.Error())
	}

	length2, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read DanyDataHolder length 2. %s", err.Error())
	}

	if _, ok := AnyDataHolderObjects[typeName.Value]; !ok {
		return fmt.Errorf("Unknown AnyDataHolder type: %s", typeName.Value)
	}

	adh.ObjectData = AnyDataHolderObjects[typeName.Value].Copy()

	if err := adh.ObjectData.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read DanyDataHolder object data. %s", err.Error())
	}

	adh.TypeName = typeName.Value
	adh.Length1 = length1
	adh.Length2 = length2

	return nil
}

// Copy returns a new copied instance of DataHolder
func (adh *AnyDataHolder) Copy() RVType {
	copied := NewAnyDataHolder()

	copied.TypeName = adh.TypeName
	copied.Length1 = adh.Length1
	copied.Length2 = adh.Length2
	copied.ObjectData = adh.ObjectData.Copy()

	return copied
}

// Equals checks if the passed Structure contains the same data as the current instance
func (adh *AnyDataHolder) Equals(o RVType) bool {
	if _, ok := o.(*AnyDataHolder); !ok {
		return false
	}

	other := o.(*AnyDataHolder)

	if adh.TypeName != other.TypeName {
		return false
	}

	if adh.Length1 != other.Length1 {
		return false
	}

	if adh.Length2 != other.Length2 {
		return false
	}

	return adh.ObjectData.Equals(other.ObjectData)
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewAnyDataHolder returns a new AnyDataHolder
func NewAnyDataHolder() *AnyDataHolder {
	return &AnyDataHolder{}
}
