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
	TypeName   *String
	Length1    *PrimitiveU32
	Length2    *PrimitiveU32
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

// Copy returns a new copied instance of DataHolder
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

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewAnyDataHolder returns a new AnyDataHolder
func NewAnyDataHolder() *AnyDataHolder {
	return &AnyDataHolder{
		TypeName: NewString(""),
		Length1: NewPrimitiveU32(0),
		Length2: NewPrimitiveU32(0),
	}
}