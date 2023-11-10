package nex

import (
	"errors"
	"fmt"
)

// RMCMessage represents a message in the RMC (Remote Method Call) protocol
type RMCMessage struct {
	IsRequest  bool   // * Indicates if the message is a request message (true) or response message (false)
	IsSuccess  bool   // * Indicates if the message is a success message (true) for a response message
	ProtocolID uint16 // * Protocol ID of the message
	CallID     uint32 // * Call ID associated with the message
	MethodID   uint32 // * Method ID in the requested protocol
	ErrorCode  uint32 // * Error code for a response message
	Parameters []byte // * Input for the method
}

// FromBytes decodes an RMCMessage from the given byte slice.
func (rmc *RMCMessage) FromBytes(data []byte) error {
	stream := NewStreamIn(data, nil)

	length, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message size. %s", err.Error())
	}

	if stream.Remaining() != int(length) {
		return errors.New("RMC Message has unexpected size")
	}

	protocolID, err := stream.ReadUInt8()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message protocol ID. %s", err.Error())
	}

	rmc.ProtocolID = uint16(protocolID & ^byte(0x80))

	if rmc.ProtocolID == 0x7F {
		rmc.ProtocolID, err = stream.ReadUInt16LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message extended protocol ID. %s", err.Error())
		}
	}

	if protocolID&0x80 != 0 {
		rmc.IsRequest = true
		rmc.CallID, err = stream.ReadUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) call ID. %s", err.Error())
		}

		rmc.MethodID, err = stream.ReadUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) method ID. %s", err.Error())
		}

		rmc.Parameters = stream.ReadRemaining()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) parameters. %s", err.Error())
		}
	} else {
		rmc.IsRequest = false
		rmc.IsSuccess, err = stream.ReadBool()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (response) error check. %s", err.Error())
		}

		if rmc.IsSuccess {
			rmc.CallID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) call ID. %s", err.Error())
			}

			rmc.MethodID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) method ID. %s", err.Error())
			}

			rmc.MethodID = rmc.MethodID & ^uint32(0x8000)
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) method ID. %s", err.Error())
			}

			rmc.Parameters = stream.ReadRemaining()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) parameters. %s", err.Error())
			}

		} else {
			rmc.ErrorCode, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) error code. %s", err.Error())
			}

			rmc.CallID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) call ID. %s", err.Error())
			}

		}
	}

	return nil
}

// Bytes serializes the RMCMessage to a byte slice.
func (rmc *RMCMessage) Bytes() []byte {
	stream := NewStreamOut(nil)

	// * RMC requests have their protocol IDs ORed with 0x80
	var protocolIDFlag uint16 = 0x80
	if !rmc.IsRequest {
		protocolIDFlag = 0
	}

	if rmc.ProtocolID < 0x80 {
		stream.WriteUInt8(uint8(rmc.ProtocolID | protocolIDFlag))
	} else {
		stream.WriteUInt8(uint8(0x7F | protocolIDFlag))
		stream.WriteUInt16LE(rmc.ProtocolID)
	}

	if rmc.IsRequest {
		stream.WriteUInt32LE(rmc.CallID)
		stream.WriteUInt32LE(rmc.MethodID)
		stream.Grow(int64(len(rmc.Parameters)))
		stream.WriteBytesNext(rmc.Parameters)
	} else {
		if rmc.IsSuccess {
			stream.WriteBool(true)
			stream.WriteUInt32LE(rmc.CallID)
			stream.WriteUInt32LE(rmc.MethodID | 0x8000)
			stream.Grow(int64(len(rmc.Parameters)))
			stream.WriteBytesNext(rmc.Parameters)
		} else {
			stream.WriteBool(false)
			stream.WriteUInt32LE(uint32(rmc.ErrorCode))
			stream.WriteUInt32LE(rmc.CallID)
		}
	}

	serialized := stream.Bytes()

	message := NewStreamOut(nil)

	message.WriteUInt32LE(uint32(len(serialized)))
	message.Grow(int64(len(serialized)))
	message.WriteBytesNext(serialized)

	return message.Bytes()
}

// NewRMCMessage returns a new generic RMC Message
func NewRMCMessage() *RMCMessage {
	return &RMCMessage{}
}

// NewRMCRequest returns a new blank RMCRequest
func NewRMCRequest() RMCMessage {
	return RMCMessage{IsRequest: true}
}

// NewRMCSuccess returns a new RMC Message configured as a success response
func NewRMCSuccess(parameters []byte) *RMCMessage {
	message := NewRMCMessage()
	message.IsRequest = false
	message.IsSuccess = true
	message.Parameters = parameters

	return message
}

// NewRMCError returns a new RMC Message configured as a error response
func NewRMCError(errorCode uint32) *RMCMessage {
	if int(errorCode)&errorMask == 0 {
		errorCode = uint32(int(errorCode) | errorMask)
	}

	message := NewRMCMessage()
	message.IsRequest = false
	message.IsSuccess = false
	message.ErrorCode = errorCode

	return message
}
