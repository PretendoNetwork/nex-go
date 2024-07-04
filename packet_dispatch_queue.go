package nex

// PacketDispatchQueue is an implementation of rdv::PacketDispatchQueue.
// PacketDispatchQueue is used to sequence incoming packets.
// In the original library each virtual connection stream only uses a single PacketDispatchQueue, but starting
// in PRUDPv1 NEX virtual connections may have multiple reliable substreams and thus multiple PacketDispatchQueues.
type PacketDispatchQueue struct {
	queue                  map[uint16]PRUDPPacketInterface
	nextExpectedSequenceId *Counter[uint16]
}

// Queue adds a packet to the queue to be dispatched
func (pdq *PacketDispatchQueue) Queue(packet PRUDPPacketInterface) {
	pdq.queue[packet.SequenceID()] = packet
}

// GetNextToDispatch returns the next packet to be dispatched, nil if there are no packets
// and a boolean indicating whether anything was returned.
func (pdq *PacketDispatchQueue) GetNextToDispatch() (PRUDPPacketInterface, bool) {
	if packet, ok := pdq.queue[pdq.nextExpectedSequenceId.Value]; ok {
		return packet, true
	}

	return nil, false
}

// Dispatched removes a packet from the queue to be dispatched.
func (pdq *PacketDispatchQueue) Dispatched(packet PRUDPPacketInterface) {
	pdq.nextExpectedSequenceId.Next()
	delete(pdq.queue, packet.SequenceID())
}

// Purge clears the queue of all pending packets.
func (pdq *PacketDispatchQueue) Purge() {
	clear(pdq.queue)
}

// NewPacketDispatchQueue initializes a new PacketDispatchQueue with a starting counter value.
func NewPacketDispatchQueue() *PacketDispatchQueue {
	return &PacketDispatchQueue{
		queue:                  make(map[uint16]PRUDPPacketInterface),
		nextExpectedSequenceId: NewCounter[uint16](2), // * First DATA packet from a client will always be 2 as the CONNECT packet is assigned 1
	}
}
