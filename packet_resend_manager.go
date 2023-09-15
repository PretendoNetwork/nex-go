package nex

import (
	"time"
)

// PendingPacket represents a packet which the server has sent but not received an ACK for
// it handles it's own retransmission on a per-packet timer
type PendingPacket struct {
	ticker        *time.Ticker
	quit          chan struct{}
	packet        PacketInterface
	iterations    *Counter
	maxIterations int
}

// BeginTimeoutTimer starts the pending packets timeout timer until it is either stopped or maxIterations is hit
func (p *PendingPacket) BeginTimeoutTimer() {
	go func() {
		for {
			select {
			case <-p.quit:
				//fmt.Println("Stopped")
				return
			case <-p.ticker.C:
				client := p.packet.Sender()
				server := client.Server()

				if int(p.iterations.Increment()) > p.maxIterations {
					// * Max iterations hit. Assume client is dead
					server.TimeoutKick(client)
					p.StopTimeoutTimer()
					return
				} else {
					// * Resend the packet
					server.SendRaw(client.Address(), p.packet.Bytes())
				}
			}
		}
	}()
}

// StopTimeoutTimer stops the packet retransmission timer
func (p *PendingPacket) StopTimeoutTimer() {
	close(p.quit)
	p.ticker.Stop()
}

// NewPendingPacket returns a new PendingPacket
func NewPendingPacket(packet PacketInterface, timeoutTime time.Duration, maxIterations int) *PendingPacket {
	p := &PendingPacket{
		ticker:        time.NewTicker(timeoutTime),
		quit:          make(chan struct{}),
		packet:        packet,
		iterations:    NewCounter(0),
		maxIterations: maxIterations,
	}

	return p
}

// PacketResendManager manages all the pending packets sent the client waiting to be ACKed
type PacketResendManager struct {
	pending       *MutexMap[uint16, *PendingPacket]
	timeoutTime   time.Duration
	maxIterations int
}

// Add creates a PendingPacket, adds it to the pool, and begins it's timeout timer
func (p *PacketResendManager) Add(packet PacketInterface) {
	cached := NewPendingPacket(packet, p.timeoutTime, p.maxIterations)
	p.pending.Set(packet.SequenceID(), cached)

	cached.BeginTimeoutTimer()
}

// Remove removes a packet from pool and stops it's timer
func (p *PacketResendManager) Remove(sequenceID uint16) {
	if cached, ok := p.pending.Get(sequenceID); ok {
		cached.StopTimeoutTimer()
		p.pending.Delete(sequenceID)
	}
}

// Clear removes all packets from pool and stops their timers
func (p *PacketResendManager) Clear() {
	p.pending.Clear(func(key uint16, value *PendingPacket) {
		value.StopTimeoutTimer()
	})
}

// NewPacketResendManager returns a new PacketResendManager
func NewPacketResendManager(timeoutTime time.Duration, maxIterations int) *PacketResendManager {
	return &PacketResendManager{
		pending:       NewMutexMap[uint16, *PendingPacket](),
		timeoutTime:   timeoutTime,
		maxIterations: maxIterations,
	}
}
