package types

// TODO - Should this have a "Value"-kind of method to get the original value?

import (
	"bytes"
	"fmt"
)

// QBuffer is a type alias of []byte with receiver methods to conform to RVType
type QBuffer []byte // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the []byte to the given writable
func (qb *QBuffer) WriteTo(writable Writable) {
	data := *qb
	length := len(data)

	writable.WritePrimitiveUInt16LE(uint16(length))

	if length > 0 {
		writable.Write([]byte(data))
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

	*qb = QBuffer(data)

	return nil
}

// Copy returns a pointer to a copy of the qBuffer. Requires type assertion when used
func (qb QBuffer) Copy() RVType {
	return &qb
}

// Equals checks if the input is equal in value to the current instance
func (qb *QBuffer) Equals(o RVType) bool {
	if _, ok := o.(*QBuffer); !ok {
		return false
	}

	return bytes.Equal([]byte(*qb), []byte(*o.(*Buffer)))
}

// NewQBuffer returns a new QBuffer
func NewQBuffer(data []byte) *QBuffer {
	var qb QBuffer = data

	return &qb
}
