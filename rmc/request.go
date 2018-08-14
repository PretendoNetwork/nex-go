package rmc

import (
	"bytes"
	"encoding/binary"
)

// RMCRequest represents a RMC protocol request
// Size      : The size of the response (minus this value)
// ProtocolID: ID of the NEX protocol being used (ORed with 0x80)
// CallID    : ID of this call (incrementing int)
// MethodID  : ID of the method
// Parameters: Request parameters
type RMCRequest struct {
	Size       uint32
	ProtocolID int
	CallID     uint32
	MethodID   uint32
	Parameters []byte
}

// FromBytes converts a byte array (payload) to a workable RMCRequest
func (Request *RMCRequest) FromBytes(Data []byte) (RMCRequest, error) {
	buffer := bytes.NewReader(Data)
	ret := RMCRequest{}

	SizeBuffer := make([]byte, 4)
	ProtocolIDBuffer := make([]byte, 1)
	CallIDBuffer := make([]byte, 4)
	MethodIDBuffer := make([]byte, 4)

	_, err := buffer.Read(SizeBuffer)
	if err != nil {
		return ret, err
	}

	_, err = buffer.Read(ProtocolIDBuffer)
	if err != nil {
		return ret, err
	}

	_, err = buffer.Read(CallIDBuffer)
	if err != nil {
		return ret, err
	}

	_, err = buffer.Read(MethodIDBuffer)
	if err != nil {
		return ret, err
	}

	Size := binary.LittleEndian.Uint16(SizeBuffer)
	ProtocolID := int(ProtocolIDBuffer[0])
	CallID := binary.LittleEndian.Uint16(CallIDBuffer)
	MethodID := binary.LittleEndian.Uint16(MethodIDBuffer)
	Parameters := Data[Size-13:]

	ret.Size = uint32(Size)
	ret.ProtocolID = ProtocolID
	ret.CallID = uint32(CallID)
	ret.MethodID = uint32(MethodID)
	ret.Parameters = Parameters

	return ret, nil
}
