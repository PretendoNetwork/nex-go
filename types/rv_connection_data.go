package types

import (
	"fmt"
)

// RVConnectionData is a class which holds data about a Rendez-Vous connection
type RVConnectionData struct {
	Structure
	StationURL                 *StationURL
	SpecialProtocols           *List[*PrimitiveU8]
	StationURLSpecialProtocols *StationURL
	Time                       *DateTime
}

// WriteTo writes the RVConnectionData to the given writable
func (rvcd *RVConnectionData) WriteTo(writable Writable) {
	contentWritable := writable.CopyNew()

	rvcd.StationURL.WriteTo(contentWritable)
	rvcd.SpecialProtocols.WriteTo(contentWritable)
	rvcd.StationURLSpecialProtocols.WriteTo(contentWritable)

	if rvcd.StructureVersion >= 1 {
		rvcd.Time.WriteTo(contentWritable)
	}

	content := contentWritable.Bytes()

	rvcd.WriteHeaderTo(writable, uint32(len(content)))

	writable.Write(content)
}

// ExtractFrom extracts the RVConnectionData to the given readable
func (rvcd *RVConnectionData) ExtractFrom(readable Readable) error {
	var err error
	if err = rvcd.ExtractHeaderFrom(readable); err != nil {
		return fmt.Errorf("Failed to read RVConnectionData header. %s", err.Error())
	}

	err = rvcd.StationURL.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read RVConnectionData.StationURL. %s", err.Error())
	}

	err = rvcd.SpecialProtocols.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read RVConnectionData.SpecialProtocols. %s", err.Error())
	}

	err = rvcd.StationURLSpecialProtocols.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to read RVConnectionData.StationURLSpecialProtocols. %s", err.Error())
	}

	if rvcd.StructureVersion >= 1 {
		err := rvcd.Time.ExtractFrom(readable)
		if err != nil {
			return fmt.Errorf("Failed to read RVConnectionData.Time. %s", err.Error())
		}
	}

	return nil
}

// Copy returns a new copied instance of RVConnectionData
func (rvcd *RVConnectionData) Copy() RVType {
	copied := NewRVConnectionData()

	copied.StructureVersion = rvcd.StructureVersion
	copied.StationURL = rvcd.StationURL.Copy().(*StationURL)
	copied.SpecialProtocols = rvcd.SpecialProtocols.Copy().(*List[*PrimitiveU8])
	copied.StationURLSpecialProtocols = rvcd.StationURLSpecialProtocols.Copy().(*StationURL)

	if rvcd.StructureVersion >= 1 {
		copied.Time = rvcd.Time.Copy().(*DateTime)
	}

	return copied
}

// Equals checks if the input is equal in value to the current instance
func (rvcd *RVConnectionData) Equals(o RVType) bool {
	if _, ok := o.(*RVConnectionData); !ok {
		return false
	}

	other := o.(*RVConnectionData)

	if rvcd.StructureVersion != other.StructureVersion {
		return false
	}

	if !rvcd.StationURL.Equals(other.StationURL) {
		return false
	}

	if !rvcd.SpecialProtocols.Equals(other.SpecialProtocols) {
		return false
	}

	if !rvcd.StationURLSpecialProtocols.Equals(other.StationURLSpecialProtocols) {
		return false
	}

	if rvcd.StructureVersion >= 1 {
		if !rvcd.Time.Equals(other.Time) {
			return false
		}
	}

	return true
}

// NewRVConnectionData returns a new RVConnectionData
func NewRVConnectionData() *RVConnectionData {
	rvcd := &RVConnectionData{
		StationURL: NewStationURL(""),
		SpecialProtocols: NewList[*PrimitiveU8](),
		StationURLSpecialProtocols: NewStationURL(""),
		Time: NewDateTime(0),
	}

	rvcd.SpecialProtocols.Type = NewPrimitiveU8(0)

	return rvcd
}
