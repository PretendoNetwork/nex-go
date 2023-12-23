package types

import (
	"errors"
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

	if rvcd.structureVersion >= 1 {
		rvcd.Time.WriteTo(contentWritable)
	}

	content := contentWritable.Bytes()

	if writable.UseStructureHeader() {
		writable.WritePrimitiveUInt8(rvcd.StructureVersion())
		writable.WritePrimitiveUInt32LE(uint32(len(content)))
	}

	writable.Write(content)
}

// ExtractFrom extracts the RVConnectionData to the given readable
func (rvcd *RVConnectionData) ExtractFrom(readable Readable) error {
	if readable.UseStructureHeader() {
		version, err := readable.ReadPrimitiveUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read RVConnectionData version. %s", err.Error())
		}

		contentLength, err := readable.ReadPrimitiveUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read RVConnectionData content length. %s", err.Error())
		}

		if readable.Remaining() < uint64(contentLength) {
			return errors.New("RVConnectionData content length longer than data size")
		}

		rvcd.SetStructureVersion(version)
	}

	var stationURL *StationURL
	specialProtocols := NewList[*PrimitiveU8]()
	var stationURLSpecialProtocols *StationURL
	var time *DateTime

	specialProtocols.Type = NewPrimitiveU8()

	if err := stationURL.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read RVConnectionData StationURL. %s", err.Error())
	}

	if err := specialProtocols.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read SpecialProtocols StationURL. %s", err.Error())
	}

	if err := stationURLSpecialProtocols.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read StationURLSpecialProtocols StationURL. %s", err.Error())
	}

	if rvcd.structureVersion >= 1 {
		if err := time.ExtractFrom(readable); err != nil {
			return fmt.Errorf("Failed to read Time StationURL. %s", err.Error())
		}
	}

	rvcd.StationURL = stationURL
	rvcd.SpecialProtocols = specialProtocols
	rvcd.StationURLSpecialProtocols = stationURLSpecialProtocols
	rvcd.Time = time

	return nil
}

// Copy returns a new copied instance of RVConnectionData
func (rvcd *RVConnectionData) Copy() RVType {
	copied := NewRVConnectionData()

	copied.structureVersion = rvcd.structureVersion
	copied.StationURL = rvcd.StationURL.Copy().(*StationURL)
	copied.SpecialProtocols = rvcd.SpecialProtocols.Copy().(*List[*PrimitiveU8])
	copied.StationURLSpecialProtocols = rvcd.StationURLSpecialProtocols.Copy().(*StationURL)

	if rvcd.structureVersion >= 1 {
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

	if rvcd.structureVersion != other.structureVersion {
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

	if rvcd.structureVersion >= 1 {
		if !rvcd.Time.Equals(other.Time) {
			return false
		}
	}

	return true
}

// TODO - Should this take in a default value, or take in nothing and have a "SetFromData"-kind of method?
// NewRVConnectionData returns a new RVConnectionData
func NewRVConnectionData() *RVConnectionData {
	return &RVConnectionData{}
}
