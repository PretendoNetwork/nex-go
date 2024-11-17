package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

// Buffer is an implementation of rdv::Buffer.
// Type alias of []byte.
// Same as QBuffer but with a uint32 length field.
type Buffer []byte

// WriteTo writes the Buffer to the given writable
func (b Buffer) WriteTo(writable Writable) {
	length := len(b)

	writable.WriteUInt32LE(uint32(length))

	if length > 0 {
		writable.Write(b)
	}
}

// ExtractFrom extracts the Buffer from the given readable
func (b *Buffer) ExtractFrom(readable Readable) error {
	length, err := readable.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer length. %s", err.Error())
	}

	value, err := readable.Read(uint64(length))
	if err != nil {
		return fmt.Errorf("Failed to read NEX Buffer data. %s", err.Error())
	}

	*b = Buffer(value)
	return nil
}

// Copy returns a pointer to a copy of the Buffer. Requires type assertion when used
func (b Buffer) Copy() RVType {
	return NewBuffer(b)
}

// Equals checks if the input is equal in value to the current instance
func (b Buffer) Equals(o RVType) bool {
	if _, ok := o.(Buffer); !ok {
		return false
	}

	return bytes.Equal(b, o.(Buffer))
}

// CopyRef copies the current value of the Buffer
// and returns a pointer to the new copy
func (b Buffer) CopyRef() RVTypePtr {
	copied := NewBuffer(b)
	return &copied
}

// Deref takes a pointer to the Buffer
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (b *Buffer) Deref() RVType {
	return *b
}

// String returns a string representation of the struct
func (b Buffer) String() string {
	return hex.EncodeToString(b)
}

// NewBuffer returns a new Buffer
func NewBuffer(input []byte) Buffer {
	b := make(Buffer, len(input))
	copy(b, input)

	return b
}
