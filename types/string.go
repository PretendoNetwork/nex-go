package types

// TODO - Should this have a "Value"-kind of method to get the original value?

import (
	"errors"
	"fmt"
	"strings"
)

// String is a type alias of string with receiver methods to conform to RVType
type String string // TODO - Should we make this a struct instead of a type alias?

// WriteTo writes the String to the given writable
func (s *String) WriteTo(writable Writable) {
	str := *s + "\x00"
	strLength := len(str)

	if writable.StringLengthSize() == 4 {
		writable.WritePrimitiveUInt32LE(uint32(strLength))
	} else {
		writable.WritePrimitiveUInt16LE(uint16(strLength))
	}

	writable.Write([]byte(str))
}

// ExtractFrom extracts the String to the given readable
func (s *String) ExtractFrom(readable Readable) error {
	var length uint64
	var err error

	if readable.StringLengthSize() == 4 {
		l, e := readable.ReadPrimitiveUInt32LE()
		length = uint64(l)
		err = e
	} else {
		l, e := readable.ReadPrimitiveUInt16LE()
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
func (s *String) Copy() RVType {
	copied := String(*s)

	return &copied
}

// Equals checks if the input is equal in value to the current instance
func (s *String) Equals(o RVType) bool {
	if _, ok := o.(*String); !ok {
		return false
	}

	return *s == *o.(*String)
}

// NewString returns a new String
func NewString(str string) *String {
	s := String(str)

	return &s
}
