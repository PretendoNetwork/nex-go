package nex

import "sync"

// PacketManager implements an API for pushing/popping packets in the correct order
type PacketManager struct {
	currentSequenceID *Counter
	packets           []PacketInterface
	*sync.RWMutex
}

// Next gets the next packet in the sequence. Returns nil if the next packet has not been sent yet
func (p *PacketManager) Next() PacketInterface {
	p.Lock()
	defer p.Unlock()

	var packet PacketInterface

	for i := 0; i < len(p.packets); i++ {
		if p.currentSequenceID.Value() == uint32(p.packets[i].SequenceID()) {
			packet = p.packets[i]
			p.RemoveByIndex(i)
			p.currentSequenceID.Increment()
			break
		}
	}

	return packet
}

// Push adds a packet to the pool to choose from in Next
func (p *PacketManager) Push(packet PacketInterface) {
	p.Lock()
	defer p.Unlock()

	p.packets = append(p.packets, packet)
}

// RemoveByIndex removes a packet from the pool using it's index in the slice
func (p *PacketManager) RemoveByIndex(i int) {
	// * https://stackoverflow.com/a/37335777
	p.packets[i] = p.packets[len(p.packets)-1]
	p.packets = p.packets[:len(p.packets)-1]
}

// NewPacketManager returns a new PacketManager
func NewPacketManager() *PacketManager {
	return &PacketManager{
		RWMutex: &sync.RWMutex{},
		currentSequenceID: NewCounter(0),
		packets:           make([]PacketInterface, 0),
	}
}
