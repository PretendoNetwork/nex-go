package nex

type PacketDispatchQueue struct {
	queue   map[uint16]PRUDPPacketInterface
	counter *Counter[uint16]
}

func (pdq *PacketDispatchQueue) Queue(packet PRUDPPacketInterface) {
	pdq.queue[packet.SequenceID()] = packet
}

func (pdq *PacketDispatchQueue) GetNextToDispatch() (PRUDPPacketInterface, bool) {

	if packet, ok := pdq.queue[pdq.counter.Value]; ok {
		return packet, true
	}

	return nil, false
}

func (pdq *PacketDispatchQueue) Dispatched(packet PRUDPPacketInterface) {
	pdq.counter.Next()
	delete(pdq.queue, packet.SequenceID())
}

func (pdq *PacketDispatchQueue) Purge() {
	for k := range pdq.queue {
		delete(pdq.queue, k)
	}
}

func NewPacketDispatchQueue() *PacketDispatchQueue {
	return &PacketDispatchQueue{
		queue:   make(map[uint16]PRUDPPacketInterface),
		counter: NewCounter[uint16](1),
	}
}
