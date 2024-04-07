package nex

import "fmt"

// Error is a custom error type implementing the error interface.
type Error struct {
	ResultCode uint32          // * NEX result code. See result_codes.go for details
	Message    string          // * The error base message
	Packet     PacketInterface // * The packet which caused the error. May not always be present
}

// Error satisfies the error interface and prints the underlying error
func (e Error) Error() string {
	resultCode := e.ResultCode

	if int(resultCode)&errorMask != 0 {
		// * Result codes are stored without the MSB set
		resultCode = resultCode & ^uint32(errorMask)
	}

	return fmt.Sprintf("[%s] %s", ResultCodeToName(resultCode), e.Message)
}

// NewError returns a new NEX error with a RDV result code
func NewError(resultCode uint32, message string) *Error {
	if int(resultCode)&errorMask == 0 {
		// * Set the MSB to mark the result as an error
		resultCode = uint32(int(resultCode) | errorMask)
	}

	return &Error{
		ResultCode: resultCode,
		Message:    message,
	}
}
