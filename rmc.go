package nex

import "errors"

// RMCRequest represets a RMC request
type RMCRequest struct {
	protocolID uint8
	customID   uint16
	callID     uint32
	methodID   uint32
	parameters []byte
}

// ProtocolID sets the RMC request protocolID
func (request *RMCRequest) ProtocolID() uint8 {
	return request.protocolID
}

// CustomID returns the RMC request custom ID
func (request *RMCRequest) CustomID() uint16 {
	return request.customID
}

// CallID sets the RMC request callID
func (request *RMCRequest) CallID() uint32 {
	return request.callID
}

// MethodID sets the RMC request methodID
func (request *RMCRequest) MethodID() uint32 {
	return request.methodID
}

// Parameters sets the RMC request parameters
func (request *RMCRequest) Parameters() []byte {
	return request.parameters
}

// SetCustomID sets the RMC request custom ID
func (request *RMCRequest) SetCustomID(customID uint16) {
	request.customID = customID
}

// SetProtocolID sets the RMC request protocol ID
func (request *RMCRequest) SetProtocolID(protocolID uint8) {
	request.protocolID = protocolID
}

// SetCallID sets the RMC request call ID
func (request *RMCRequest) SetCallID(callID uint32) {
	request.callID = callID
}

// SetMethodID sets the RMC request method ID
func (request *RMCRequest) SetMethodID(methodID uint32) {
	request.methodID = methodID
}

// SetParameters sets the RMC request parameters
func (request *RMCRequest) SetParameters(parameters []byte) {
	request.parameters = parameters
}

// NewRMCRequest returns a new blank RMCRequest
func NewRMCRequest() RMCRequest {
	return RMCRequest{}
}

// FromBytes converts a byte slice into a RMCRequest
func (request *RMCRequest) FromBytes(data []byte) error {
	if len(data) < 13 {
		return errors.New("[RMC] Data size less than minimum")
	}

	stream := NewStreamIn(data, nil)

	size := int(stream.ReadUInt32LE())

	if size != (len(data) - 4) {
		return errors.New("[RMC] Data size does not match")
	}

	protocolID := stream.ReadUInt8() ^ 0x80
	if protocolID == 0x7f {
		request.customID = stream.ReadUInt16LE()
	}
	callID := stream.ReadUInt32LE()
	methodID := stream.ReadUInt32LE()
	parameters := data[stream.ByteOffset():]

	request.protocolID = protocolID
	request.callID = callID
	request.methodID = methodID
	request.parameters = parameters

	return nil
}

// Bytes converts a RMCRequest struct into a usable byte array
func (request *RMCRequest) Bytes() []byte {
	body := NewStreamOut(nil)

	body.WriteUInt8(request.protocolID | 0x80)
	if request.protocolID == 0x7f {
		body.WriteUInt16LE(request.customID)
	}

	body.WriteUInt32LE(request.callID)
	body.WriteUInt32LE(request.methodID)

	if request.parameters != nil && len(request.parameters) > 0 {
		body.Grow(int64(len(request.parameters)))
		body.WriteBytesNext(request.parameters)
	}

	data := NewStreamOut(nil)

	data.WriteBuffer(body.Bytes())

	return data.Bytes()
}

// RMCResponse represents a RMC response
type RMCResponse struct {
	protocolID uint8
	customID   uint16
	success    uint8
	callID     uint32
	methodID   uint32
	data       []byte
	errorCode  uint32
}

// CustomID returns the RMC response customID
func (response *RMCResponse) CustomID() uint16 {
	return response.customID
}

// SetCustomID sets the RMC response customID
func (response *RMCResponse) SetCustomID(customID uint16) {
	response.customID = customID
}

// SetSuccess sets the RMCResponse payload to an instance of RMCSuccess
func (response *RMCResponse) SetSuccess(methodID uint32, data []byte) {
	response.success = 1
	response.methodID = methodID
	response.data = data
}

// SetError sets the RMCResponse payload to an instance of RMCError
func (response *RMCResponse) SetError(errorCode uint32) {
	if int(errorCode)&errorMask == 0 {
		errorCode = uint32(int(errorCode) | errorMask)
	}

	response.success = 0
	response.errorCode = errorCode
}

// Bytes converts a RMCResponse struct into a usable byte array
func (response *RMCResponse) Bytes() []byte {
	body := NewStreamOut(nil)

	if response.protocolID > 0 {
		body.WriteUInt8(response.protocolID)
		if response.protocolID == 0x7f {
			body.WriteUInt16LE(response.customID)
		}
	}
	body.WriteUInt8(response.success)

	if response.success == 1 {
		body.WriteUInt32LE(response.callID)
		body.WriteUInt32LE(response.methodID | 0x8000)

		if response.data != nil && len(response.data) > 0 {
			body.Grow(int64(len(response.data)))
			body.WriteBytesNext(response.data)
		}
	} else {
		body.WriteUInt32LE(response.errorCode)
		body.WriteUInt32LE(response.callID)
	}

	data := NewStreamOut(nil)

	data.WriteBuffer(body.Bytes())

	return data.Bytes()
}

// NewRMCResponse returns a new RMCResponse
func NewRMCResponse(protocolID uint8, callID uint32) RMCResponse {
	response := RMCResponse{
		protocolID: protocolID,
		callID:     callID,
	}

	return response
}
