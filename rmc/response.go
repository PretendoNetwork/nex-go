package nex

import (
	"fmt"

	"bytes"
	"encoding/binary"
	"os"
)

// RMCResponse represents a RMC protocol response
// Size      : The size of the response (minus this value)
// ProtocolID: ID of the NEX protocol being used (not ORed)
// Success   : 1 or 0 based on if the response is a success or error
// Body      : The response
type RMCResponse struct {
	Size       uint32
	ProtocolID int
	Success    int
	Body       interface{}
}

// RMCSuccess represents a successful RMC payload
// CallID  : ID of this call (incrementing int)
// MethodID: ID of the method (ORed with 0x8000)
// Data    : The payload data
type RMCSuccess struct {
	CallID   uint32
	MethodID uint32
	Data     []byte
}

// RMCError represents a RMC error payload
// ErrorCode: The error code of the error
// CallID   : ID of this call (incrementing int)
type RMCError struct {
	ErrorCode uint32
	CallID    uint32
}

// SetSuccess sets the RMCResponse payload to an instance of RMCSuccess
func (Response *RMCResponse) SetSuccess(CallID uint32, MethodID uint32, Data []byte) uint32 {
	Response.Success = 1
	Response.Body = RMCSuccess{CallID, MethodID | 0x8000, Data}

	size := 10 + len(Data)

	return uint32(size)
}

// SetError sets the RMCResponse payload to an instance of RMCError
func (Response *RMCResponse) SetError(ErrorCode uint32, CallID uint32) uint32 {
	Response.Success = 0
	Response.Body = RMCError{ErrorCode, CallID}

	size := 10

	return uint32(size)
}

// Bytes converts a RMCResponse struct into a usable byte array
func (Response *RMCResponse) Bytes() []byte {
	data := bytes.NewBuffer(make([]byte, 0, Response.Size+1))

	binary.Write(data, binary.LittleEndian, uint32(Response.Size))
	binary.Write(data, binary.LittleEndian, byte(Response.ProtocolID))
	binary.Write(data, binary.LittleEndian, byte(Response.Success))

	if Response.Success == 1 {
		body := Response.Body.(RMCSuccess)

		binary.Write(data, binary.LittleEndian, uint32(body.CallID))
		binary.Write(data, binary.LittleEndian, uint32(body.MethodID))
		binary.Write(data, binary.LittleEndian, body.Data)
	} else if Response.Success == 0 {
		body := Response.Body.(RMCError)

		binary.Write(data, binary.LittleEndian, uint32(body.ErrorCode))
		binary.Write(data, binary.LittleEndian, uint32(body.CallID))
	} else {
		fmt.Println("Invalid RMC success type", Response.Success)
		os.Exit(1)
	}

	return data.Bytes()
}

/*
func main() {
	var response RMCResponse

	ProtocolID := 0x66  // Friends service (WiiU) NOT ORED! IN REQUEST THIS WILL BE 0xE6 (0x66 | 0x80)
	CallID := uint32(1) // dummy values. These would actually be gotten from a request
	//MethodID   := uint32(1) // dummy values. These would actually be gotten from a request

	response.ProtocolID = ProtocolID
	response.Size = response.SetError(0x8068000B, CallID)

	fmt.Println(response.Bytes())
}
*/
