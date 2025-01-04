package nex

import (
	"errors"
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2/types"
)

// RMCMessage represents a message in the RMC (Remote Method Call) protocol
type RMCMessage struct {
	Endpoint         EndpointInterface
	IsRequest        bool                         // * Indicates if the message is a request message (true) or response message (false)
	IsSuccess        bool                         // * Indicates if the message is a success message (true) for a response message
	IsHPP            bool                         // * Indicates if the message is an HPP message
	ProtocolID       uint16                       // * Protocol ID of the message. Only present in "packed" variations
	ProtocolName     types.String                 // * Protocol name of the message. Only present in "verbose" variations
	CallID           uint32                       // * Call ID associated with the message
	MethodID         uint32                       // * Method ID in the requested protocol. Only present in "packed" variations
	MethodName       types.String                 // * Method name in the requested protocol. Only present in "verbose" variations
	ErrorCode        uint32                       // * Error code for a response message
	VersionContainer *types.ClassVersionContainer // * Contains version info for Structures in the request. Only present in "verbose" variations. Pointer to allow for nil checks
	Parameters       []byte                       // * Input for the method
	// TODO - Verbose messages suffix response method names with "*". Should we have a "HasResponsePointer" sort of field?
}

// Copy copies the message into a new RMCMessage
func (rmc *RMCMessage) Copy() *RMCMessage {
	copied := NewRMCMessage(rmc.Endpoint)

	copied.IsRequest = rmc.IsRequest
	copied.IsSuccess = rmc.IsSuccess
	copied.IsHPP = rmc.IsHPP
	copied.ProtocolID = rmc.ProtocolID
	copied.ProtocolName = rmc.ProtocolName
	copied.CallID = rmc.CallID
	copied.MethodID = rmc.MethodID
	copied.MethodName = rmc.MethodName
	copied.ErrorCode = rmc.ErrorCode

	if rmc.Parameters != nil {
		copied.Parameters = append([]byte(nil), rmc.Parameters...)
	}

	return copied
}

// FromBytes decodes an RMCMessage from the given byte slice.
func (rmc *RMCMessage) FromBytes(data []byte) error {
	if rmc.Endpoint.UseVerboseRMC() {
		return rmc.decodeVerbose(data)
	} else {
		return rmc.decodePacked(data)
	}
}

func (rmc *RMCMessage) decodePacked(data []byte) error {
	stream := NewByteStreamIn(data, rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	length, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message size. %s", err.Error())
	}

	if stream.Remaining() != uint64(length) {
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
			rmc.Parameters = stream.ReadRemaining()

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

func (rmc *RMCMessage) decodeVerbose(data []byte) error {
	stream := NewByteStreamIn(data, rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	length, err := stream.ReadUInt32LE()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message size. %s", err.Error())
	}

	if stream.Remaining() != uint64(length) {
		return errors.New("RMC Message has unexpected size")
	}

	rmc.ProtocolName = types.NewString("")
	if err := rmc.ProtocolName.ExtractFrom(stream); err != nil {
		return fmt.Errorf("Failed to read RMC Message protocol name. %s", err.Error())
	}

	rmc.IsRequest, err = stream.ReadBool()
	if err != nil {
		return fmt.Errorf("Failed to read RMC Message \"is request\" bool. %s", err.Error())
	}

	if rmc.IsRequest {
		rmc.CallID, err = stream.ReadUInt32LE()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) call ID. %s", err.Error())
		}

		rmc.MethodName = types.NewString("")
		if err := rmc.MethodName.ExtractFrom(stream); err != nil {
			return fmt.Errorf("Failed to read RMC Message (request) method name. %s", err.Error())
		}

		versionContainer := types.NewClassVersionContainer()
		if err := versionContainer.ExtractFrom(stream); err != nil {
			return fmt.Errorf("Failed to read RMC Message ClassVersionContainer. %s", err.Error())
		}

		rmc.VersionContainer = &versionContainer
		rmc.Parameters = stream.ReadRemaining()
	} else {
		rmc.IsSuccess, err = stream.ReadBool()
		if err != nil {
			return fmt.Errorf("Failed to read RMC Message (response) error check. %s", err.Error())
		}

		if rmc.IsSuccess {
			rmc.CallID, err = stream.ReadUInt32LE()
			if err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) call ID. %s", err.Error())
			}

			rmc.MethodName = types.NewString("")
			if err := rmc.MethodName.ExtractFrom(stream); err != nil {
				return fmt.Errorf("Failed to read RMC Message (response) method name. %s", err.Error())
			}

			rmc.Parameters = stream.ReadRemaining()

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
	if rmc.Endpoint.UseVerboseRMC() {
		return rmc.encodeVerbose()
	} else {
		return rmc.encodePacked()
	}
}

func (rmc *RMCMessage) encodePacked() []byte {
	stream := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	// * RMC requests have their protocol IDs ORed with 0x80
	var protocolIDFlag uint16 = 0x80
	if !rmc.IsRequest {
		protocolIDFlag = 0
	}

	// * HPP does not include the protocol ID on the response. We technically
	// * don't have to support converting HPP requests to bytes but we'll
	// * do it for accuracy.
	if !rmc.IsHPP || (rmc.IsHPP && rmc.IsRequest) {
		if rmc.ProtocolID < 0x80 {
			stream.WriteUInt8(uint8(rmc.ProtocolID | protocolIDFlag))
		} else {
			stream.WriteUInt8(uint8(0x7F | protocolIDFlag))
			stream.WriteUInt16LE(rmc.ProtocolID)
		}
	}

	if rmc.IsRequest {
		stream.WriteUInt32LE(rmc.CallID)
		stream.WriteUInt32LE(rmc.MethodID)

		if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
			stream.Grow(int64(len(rmc.Parameters)))
			stream.WriteBytesNext(rmc.Parameters)
		}
	} else {
		stream.WriteBool(rmc.IsSuccess)

		if rmc.IsSuccess {
			stream.WriteUInt32LE(rmc.CallID)
			stream.WriteUInt32LE(rmc.MethodID | 0x8000)

			if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
				stream.Grow(int64(len(rmc.Parameters)))
				stream.WriteBytesNext(rmc.Parameters)
			}
		} else {
			stream.WriteUInt32LE(uint32(rmc.ErrorCode))
			stream.WriteUInt32LE(rmc.CallID)
		}
	}

	serialized := stream.Bytes()

	message := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	message.WriteUInt32LE(uint32(len(serialized)))
	message.Grow(int64(len(serialized)))
	message.WriteBytesNext(serialized)

	return message.Bytes()
}

func (rmc *RMCMessage) encodeVerbose() []byte {
	stream := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	rmc.ProtocolName.WriteTo(stream)
	stream.WriteBool(rmc.IsRequest)

	if rmc.IsRequest {
		stream.WriteUInt32LE(rmc.CallID)
		rmc.MethodName.WriteTo(stream)

		if rmc.VersionContainer != nil {
			rmc.VersionContainer.WriteTo(stream)
		} else {
			// * Fail safe. This is always present even if no structures are used
			stream.WriteUInt32LE(0)
		}

		if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
			stream.Grow(int64(len(rmc.Parameters)))
			stream.WriteBytesNext(rmc.Parameters)
		}
	} else {
		stream.WriteBool(rmc.IsSuccess)

		if rmc.IsSuccess {
			stream.WriteUInt32LE(rmc.CallID)
			rmc.MethodName.WriteTo(stream)

			if rmc.Parameters != nil && len(rmc.Parameters) > 0 {
				stream.Grow(int64(len(rmc.Parameters)))
				stream.WriteBytesNext(rmc.Parameters)
			}
		} else {
			stream.WriteUInt32LE(uint32(rmc.ErrorCode))
			stream.WriteUInt32LE(rmc.CallID)
		}
	}

	serialized := stream.Bytes()

	message := NewByteStreamOut(rmc.Endpoint.LibraryVersions(), rmc.Endpoint.ByteStreamSettings())

	message.WriteUInt32LE(uint32(len(serialized)))
	message.Grow(int64(len(serialized)))
	message.WriteBytesNext(serialized)

	return message.Bytes()
}

// NewRMCMessage returns a new generic RMC Message
func NewRMCMessage(endpoint EndpointInterface) *RMCMessage {
	return &RMCMessage{
		Endpoint: endpoint,
	}
}

// NewRMCRequest returns a new blank RMCRequest
func NewRMCRequest(endpoint EndpointInterface) *RMCMessage {
	return &RMCMessage{
		Endpoint:  endpoint,
		IsRequest: true,
	}
}

// NewRMCSuccess returns a new RMC Message configured as a success response
func NewRMCSuccess(endpoint EndpointInterface, parameters []byte) *RMCMessage {
	message := NewRMCMessage(endpoint)
	message.IsRequest = false
	message.IsSuccess = true
	message.Parameters = parameters

	return message
}

// NewRMCError returns a new RMC Message configured as a error response
func NewRMCError(endpoint EndpointInterface, errorCode uint32) *RMCMessage {
	if int(errorCode)&errorMask == 0 {
		errorCode = uint32(int(errorCode) | errorMask)
	}

	message := NewRMCMessage(endpoint)
	message.IsRequest = false
	message.IsSuccess = false
	message.ErrorCode = errorCode

	return message
}
