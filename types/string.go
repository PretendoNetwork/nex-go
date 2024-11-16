package types

import (
	"errors"
	"fmt"
	"strings"
)

// String is an implementation of rdv::String.
// Type alias of string
type String string

// WriteTo writes the String to the given writable
func (s String) WriteTo(writable Writable) {
	s = s + "\x00"
	strLength := len(s)

	if writable.StringLengthSize() == 4 {
		writable.WriteUInt32LE(uint32(strLength))
	} else {
		writable.WriteUInt16LE(uint16(strLength))
	}

	writable.Write([]byte(s))
}

// ExtractFrom extracts the String from the given readable
func (s *String) ExtractFrom(readable Readable) error {
	var length uint64
	var err error

	if readable.StringLengthSize() == 4 {
		l, e := readable.ReadUInt32LE()
		length = uint64(l)
		err = e
	} else {
		l, e := readable.ReadUInt16LE()
		length = uint64(l)
		err = e
	}

	if err != nil {
		return fmt.Errorf("Failed to read NEX string length. %s", err.Error())
	}

	if readable.Remaining() < length {
		return errors.New("NEX string length longer than data size")
	}

	stringData, err := readable.Read(length)
	if err != nil {
		return fmt.Errorf("Failed to read NEX string length. %s", err.Error())
	}

	str := strings.TrimRight(string(stringData), "\x00")

	*s = String(str)
	return nil
}

// Copy returns a pointer to a copy of the String. Requires type assertion when used
func (s String) Copy() RVType {
	return NewString(string(s))
}

// Equals checks if the input is equal in value to the current instance
func (s String) Equals(o RVType) bool {
	if _, ok := o.(String); !ok {
		return false
	}

	return s == o.(String)
}

// CopyRef copies the current value of the String
// and returns a pointer to the new copy
func (s String) CopyRef() RVTypePtr {
	return &s
}

// Deref takes a pointer to the String
// and dereferences it to the raw value.
// Only useful when working with an instance of RVTypePtr
func (s *String) Deref() RVType {
	return *s
}

// String returns a string representation of the struct
func (s String) String() string {
	return fmt.Sprintf("%q", string(s))
}

// NewString returns a new String
func NewString(input string) String {
	s := String(input)
	return s
}
