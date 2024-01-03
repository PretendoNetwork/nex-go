package types

import (
	"bytes"
	"fmt"
)

// Buffer is a struct of []byte with receiver methods to conform to RVType
type Buffer struct {
	Value []byte
}

// WriteTo writes the []byte to the given writable
func (b *Buffer) WriteTo(writable Writable) {
	length := len(b.Value)

	writable.WritePrimitiveUInt32LE(uint32(length))

	if length > 0 {
		writable.Write(b.Value)
	}
}

// ExtractFrom extracts the Buffer from the given readable
func (b *Buffer) ExtractFrom(readable Readable) error {
	length, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer length. %s", err.Error())
	}

	value, err := readable.Read(uint64(length))
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer data. %s", err.Error())
	}

	b.Value = value

	return nil
}

// Copy returns a pointer to a copy of the Buffer. Requires type assertion when used
func (b *Buffer) Copy() RVType {
	return NewBuffer(b.Value)
}

// Equals checks if the input is equal in value to the current instance
func (b *Buffer) Equals(o RVType) bool {
	if _, ok := o.(*Buffer); !ok {
		return false
	}

	return bytes.Equal(b.Value, o.(*Buffer).Value)
}

// String returns a string representation of the struct
func (b *Buffer) String() string {
	return fmt.Sprintf("%x", b.Value)
}

// NewBuffer returns a new Buffer
func NewBuffer(data []byte) *Buffer {
	return &Buffer{Value: data}
}
