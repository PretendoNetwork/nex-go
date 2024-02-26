package nex

// TODO - Should this be moved to the types module?

// StreamType is an implementation of the rdv::Stream::Type enum.
//
// StreamType is used to create VirtualPorts used in PRUDP virtual
// connections. Each stream may be one of these types, and each stream
// has it's own state.
type StreamType uint8

// EnumIndex returns the StreamType enum index as a uint8
func (st StreamType) EnumIndex() uint8 {
	return uint8(st)
}

const (
	// StreamTypeDO represents the DO PRUDP virtual connection stream type
	StreamTypeDO StreamType = iota + 1

	// StreamTypeRV represents the RV PRUDP virtual connection stream type
	StreamTypeRV

	// StreamTypeOldRVSec represents the OldRVSec PRUDP virtual connection stream type
	StreamTypeOldRVSec

	// StreamTypeSBMGMT represents the SBMGMT PRUDP virtual connection stream type
	StreamTypeSBMGMT

	// StreamTypeNAT represents the NAT PRUDP virtual connection stream type
	StreamTypeNAT

	// StreamTypeSessionDiscovery represents the SessionDiscovery PRUDP virtual connection stream type
	StreamTypeSessionDiscovery

	// StreamTypeNATEcho represents the NATEcho PRUDP virtual connection stream type
	StreamTypeNATEcho

	// StreamTypeRouting represents the Routing PRUDP virtual connection stream type
	StreamTypeRouting

	// StreamTypeGame represents the Game PRUDP virtual connection stream type
	StreamTypeGame

	// StreamTypeRVSecure represents the RVSecure PRUDP virtual connection stream type
	StreamTypeRVSecure

	// StreamTypeRelay represents the Relay PRUDP virtual connection stream type
	StreamTypeRelay
)
