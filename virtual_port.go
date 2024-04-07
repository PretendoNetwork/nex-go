package nex

import "github.com/PretendoNetwork/nex-go/v2/constants"

// VirtualPort in an implementation of rdv::VirtualPort.
// PRUDP will reuse a single physical socket connection for many virtual PRUDP connections.
// VirtualPorts are a byte which represents a stream for a virtual PRUDP connection.
// This byte is two 4-bit fields. The upper 4 bits are the stream type, the lower 4 bits
// are the stream ID. The client starts with stream ID 15, decrementing by one with each new
// virtual connection.
type VirtualPort byte

// SetStreamType sets the VirtualPort stream type
func (vp *VirtualPort) SetStreamType(streamType constants.StreamType) {
	*vp = VirtualPort((byte(*vp) & 0x0F) | (byte(streamType) << 4))
}

// StreamType returns the VirtualPort stream type
func (vp VirtualPort) StreamType() constants.StreamType {
	return constants.StreamType(vp >> 4)
}

// SetStreamID sets the VirtualPort stream ID
func (vp *VirtualPort) SetStreamID(streamID uint8) {
	*vp = VirtualPort((byte(*vp) & 0xF0) | (streamID & 0x0F))
}

// StreamID returns the VirtualPort stream ID
func (vp VirtualPort) StreamID() uint8 {
	return uint8(vp & 0xF)
}
