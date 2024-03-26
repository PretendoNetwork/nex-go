package types

import (
	"errors"
	"fmt"
)

// Structure represents a Quazal Rendez-Vous/NEX Structure (custom class) base struct.
type Structure struct {
	StructureVersion uint8
}

// ExtractHeaderFrom extracts the structure header from the given readable
func (s *Structure) ExtractHeaderFrom(readable Readable) error {
	if readable.UseStructureHeader() {
		version, err := readable.ReadPrimitiveUInt8()
		if err != nil {
			return fmt.Errorf("Failed to read Structure version. %s", err.Error())
		}

		contentLength, err := readable.ReadPrimitiveUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read Structure content length. %s", err.Error())
		}

		if readable.Remaining() < uint64(contentLength) {
			return errors.New("Structure content length longer than data size")
		}

		s.StructureVersion = version
	}

	return nil
}

// WriteHeaderTo writes the structure header to the given writable
func (s *Structure) WriteHeaderTo(writable Writable, contentLength uint32) {
	if writable.UseStructureHeader() {
		writable.WritePrimitiveUInt8(s.StructureVersion)
		writable.WritePrimitiveUInt32LE(contentLength)
	}
}
