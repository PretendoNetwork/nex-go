package types

import (
	"fmt"
	"strings"
)

var errorMask = 1 << 31

// Result is sent in methods which query large objects
type Result struct {
	Code uint32 // TODO - Replace this with PrimitiveU32?
}

// WriteTo writes the Result to the given writable
func (r *Result) WriteTo(writable Writable) {
	writable.WritePrimitiveUInt32LE(r.Code)
}

// ExtractFrom extracts the Result to the given readable
func (r *Result) ExtractFrom(readable Readable) error {
	code, err := readable.ReadPrimitiveUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read Result code. %s", err.Error())
	}

	r.Code = code

	return nil
}

// Copy returns a pointer to a copy of the Result. Requires type assertion when used
func (r *Result) Copy() RVType {
	return NewResult(r.Code)
}

// Equals checks if the input is equal in value to the current instance
func (r *Result) Equals(o RVType) bool {
	if _, ok := o.(*Result); !ok {
		return false
	}

	return r.Code == o.(*Result).Code
}

// IsSuccess returns true if the Result is a success
func (r *Result) IsSuccess() bool {
	return int(r.Code)&errorMask == 0
}

// IsError returns true if the Result is a error
func (r *Result) IsError() bool {
	return int(r.Code)&errorMask != 0
}

// String returns a string representation of the struct
func (r *Result) String() string {
	return r.FormatToString(0)
}

// FormatToString pretty-prints the struct data using the provided indentation level
func (r *Result) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("Result{\n")

	if r.IsSuccess() {
		b.WriteString(fmt.Sprintf("%scode: %d (success)\n", indentationValues, r.Code))
	} else {
		b.WriteString(fmt.Sprintf("%scode: %d (error)\n", indentationValues, r.Code))
	}

	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}

// NewResult returns a new Result
func NewResult(code uint32) *Result {
	return &Result{code}
}

// NewResultSuccess returns a new Result set as a success
func NewResultSuccess(code uint32) *Result {
	return NewResult(uint32(int(code) & ^errorMask))
}

// NewResultError returns a new Result set as an error
func NewResultError(code uint32) *Result {
	return NewResult(uint32(int(code) | errorMask))
}
