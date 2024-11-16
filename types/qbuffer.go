package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

// QBuffer is an implementation of rdv::qBuffer.
// Type alias of []byte.
// Same as Buffer but with a uint16 length field.
type QBuffer []byte

// WriteTo writes the []byte to the given writable
func (qb QBuffer) WriteTo(writable Writable) {
	length := len(qb)

	writable.WriteUInt16LE(uint16(length))

	if length > 0 {
		writable.Write(qb)
	}
}

// ExtractFrom extracts the QBuffer from the given readable
func (qb *QBuffer) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt16LE()
	if err != nil {
		return fmt.Errorf("Failed to read NEX qBuffer length. %s", err.Error())
	}

	data, err := readable.Read(uint64(length))
	if err != nil {
		return fmt.Errorf("Failed to read NEX qBuffer data. %s", err.Error())
	}

	*qb = data
	return nil
}

// Copy returns a pointer to a copy of the qBuffer. Requires type assertion when used
func (qb QBuffer) Copy() RVType {
	return NewQBuffer(qb)
}

// Equals checks if the input is equal in value to the current instance
func (qb QBuffer) Equals(o RVType) bool {
	if _, ok := o.(QBuffer); !ok {
		return false
	}

	return bytes.Equal(qb, o.(QBuffer))
}

// CopyRef copies the current value of the QBuffer
// and returns a pointer to the new copy
func (qb QBuffer) CopyRef() RVTypePtr {
	return &qb
}

// Deref takes a pointer to the QBuffer
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (qb *QBuffer) Deref() RVType {
	return *qb
}

// String returns a string representation of the struct
func (qb QBuffer) String() string {
	return hex.EncodeToString(qb)
}

// NewQBuffer returns a new QBuffer
func NewQBuffer(input []byte) QBuffer {
	qb := make(QBuffer, len(input))
	copy(qb, input)

	return qb
}
