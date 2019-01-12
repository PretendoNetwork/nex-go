package nex

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// RMCRequest represents a RMC protocol request
type RMCRequest struct {
	Size       uint32
	ProtocolID uint8
	CallID     uint32
	MethodID   uint32
	Parameters []byte
}

// NewRMCRequest returns a new RMCRequest
func NewRMCRequest(Data []byte) RMCRequest {
	stream := NewInputStream(Data)

	return RMCRequest{
		Size:       stream.UInt32LE(),
		ProtocolID: stream.UInt8(),
		CallID:     stream.UInt32LE(),
		MethodID:   stream.UInt32LE(),
		Parameters: Data[13:],
	}
}

// RMCResponse represents a RMC protocol response
type RMCResponse struct {
	Size       uint32
	ProtocolID int
	Success    int
	Body       interface{}
	CallID     uint32
}

// RMCSuccess represents a successful RMC payload
type RMCSuccess struct {
	MethodID uint32
	Data     []byte
}

// RMCError represents a RMC error payload
type RMCError struct {
	ErrorCode uint32
}

// NewRMCResponse returns a new RMCResponse
func NewRMCResponse(ProtocolID int, CallID uint32) RMCResponse {
	return RMCResponse{
		ProtocolID: ProtocolID,
		CallID:     CallID,
	}
}

// SetSuccess sets the RMCResponse payload to an instance of RMCSuccess
func (Response *RMCResponse) SetSuccess(MethodID uint32, Data []byte) {
	Response.Success = 1
	Response.Body = RMCSuccess{MethodID | 0x8000, Data}

	Response.Size = uint32(10 + len(Data))
}

// SetError sets the RMCResponse payload to an instance of RMCError
func (Response *RMCResponse) SetError(ErrorCode uint32) {
	Response.Success = 0
	Response.Body = RMCError{ErrorCode}

	Response.Size = 10
}

// Bytes converts a RMCResponse struct into a usable byte array
func (Response *RMCResponse) Bytes() []byte {
	data := bytes.NewBuffer(make([]byte, 0, Response.Size+1))

	binary.Write(data, binary.LittleEndian, uint32(Response.Size))
	binary.Write(data, binary.LittleEndian, byte(Response.ProtocolID))
	binary.Write(data, binary.LittleEndian, byte(Response.Success))

	if Response.Success == 1 {
		body := Response.Body.(RMCSuccess)

		binary.Write(data, binary.LittleEndian, uint32(Response.CallID))
		binary.Write(data, binary.LittleEndian, uint32(body.MethodID))
		binary.Write(data, binary.LittleEndian, body.Data)
	} else if Response.Success == 0 {
		body := Response.Body.(RMCError)

		binary.Write(data, binary.LittleEndian, uint32(body.ErrorCode))
		binary.Write(data, binary.LittleEndian, uint32(Response.CallID))
	} else {
		fmt.Println("Invalid RMC success type", Response.Success)
		os.Exit(1)
	}

	return data.Bytes()
}
