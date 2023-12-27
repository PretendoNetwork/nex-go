package types

import (
	"bytes"
	"fmt"
)

// QBuffer is a struct of []byte with receiver methods to conform to RVType
type QBuffer struct {
	Value []byte
}

// WriteTo writes the []byte to the given writable
func (qb *QBuffer) WriteTo(writable Writable) {
	length := len(qb.Value)

	writable.WritePrimitiveUInt16LE(uint16(length))

	if length > 0 {
		writable.Write(qb.Value)
	}
}

// ExtractFrom extracts the QBuffer to the given readable
func (qb *QBuffer) ExtractFrom(readable Readable) error {
	length, err := readable.ReadPrimitiveUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read NEX qBuffer length. %s", err.Error())
	}

	data, err := readable.Read(uint64(length))
	if err != nil {
		return fmt.Errorf("Failed to read NEX qBuffer data. %s", err.Error())
	}

	qb.Value = data

	return nil
}

// Copy returns a pointer to a copy of the qBuffer. Requires type assertion when used
func (qb *QBuffer) Copy() RVType {
	return NewQBuffer(qb.Value)
}

// Equals checks if the input is equal in value to the current instance
func (qb *QBuffer) Equals(o RVType) bool {
	if _, ok := o.(*QBuffer); !ok {
		return false
	}

	return bytes.Equal(qb.Value, o.(*QBuffer).Value)
}

// NewQBuffer returns a new QBuffer
func NewQBuffer(data []byte) *QBuffer {
	return &QBuffer{Value: data}
}
