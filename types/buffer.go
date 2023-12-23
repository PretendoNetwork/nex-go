package types

import (
	"bytes"
	"fmt"
)

// TODO - Should this have a "Value"-kind of method to get the original value?

// Buffer is a type alias of []byte with receiver methods to conform to RVType
type Buffer []byte // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the []byte to the given writable
func (b *Buffer) WriteTo(writable Writable) {
	data := *b
	length := len(data)

	writable.WritePrimitiveUInt32LE(uint32(length))

	if length > 0 {
		writable.Write([]byte(data))
	}
}

// ExtractFrom extracts the Buffer to the given readable
func (b *Buffer) ExtractFrom(readable Readable) error {
	length, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer length. %s", err.Error())
	}

	data, err := readable.Read(uint64(length))
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer data. %s", err.Error())
	}

	*b = Buffer(data)

	return nil
}

// Copy returns a pointer to a copy of the Buffer. Requires type assertion when used
func (b *Buffer) Copy() RVType {
	copied := Buffer(*b)

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (b *Buffer) Equals(o RVType) bool {
	if _, ok := o.(*Buffer); !ok {
		return false
	}

	return bytes.Equal([]byte(*b), []byte(*o.(*Buffer)))
}

// NewBuffer returns a new Buffer
func NewBuffer(data []byte) *Buffer {
	var b Buffer = data

	return &b
}
