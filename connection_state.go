package nex

// ConnectionState is an implementation of the nn::nex::EndPoint::_ConnectionState enum.
//
// The state represents a PRUDP clients connection state. The original Rendez-Vous
// library supports  states 0-6, though NEX only supports 0-4. The remaining 2 are
// unknown
type ConnectionState uint8

const (
	// StateNotConnected indicates the client has not established a full PRUDP connection
	StateNotConnected ConnectionState = iota

	// StateConnecting indicates the client is attempting to establish a PRUDP connection
	StateConnecting

	// StateConnected indicates the client has established a full PRUDP connection
	StateConnected

	// StateDisconnecting indicates the client is disconnecting from a PRUDP connection. Currently unused
	StateDisconnecting

	// StateFaulty indicates the client connection is faulty. Currently unused
	StateFaulty
)
