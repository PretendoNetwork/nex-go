package nex

// SequenceIDManager implements an API for managing the sequence IDs of different packet streams on a client
type SequenceIDManager struct {
	reliableCounter *Counter // TODO - NEX only uses one reliable stream, but Rendezvous supports many. This needs to be a slice!
	pingCounter     *Counter
	// TODO - Unreliable packets for Rendezvous
}

// Next gets the next sequence ID for the packet. Returns 0 for an unsupported packet
func (s *SequenceIDManager) Next(packet PacketInterface) uint32 {
	if packet.HasFlag(FlagReliable) {
		return s.reliableCounter.Increment()
	}

	if packet.Type() == PingPacket {
		return s.pingCounter.Increment()
	}

	return 0
}

// NewSequenceIDManager returns a new SequenceIDManager
func NewSequenceIDManager() *SequenceIDManager {
	return &SequenceIDManager{
		reliableCounter: NewCounter(0),
		pingCounter:     NewCounter(0),
	}
}
