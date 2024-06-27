package nex

import (
	"context"
	"time"
)

// TODO - REMOVE THIS ENTIRELY AND REPLACE IT WITH AN IMPLEMENTATION OF rdv::Timeout AND rdv::TimeoutManager AND USE MORE STREAM SETTINGS!

// ResendScheduler manages the resending of reliable PRUDP packets
type ResendScheduler struct {
	ctx            context.Context
	cancel         context.CancelFunc
	packets        *MutexMap[uint16, PRUDPPacketInterface]
	streamSettings *StreamSettings
}

// AddPacket adds a packet to the scheduler and begins it's timer
func (rs *ResendScheduler) AddPacket(packet PRUDPPacketInterface) {
	endpoint := packet.Sender().Endpoint().(*PRUDPEndPoint)

	rto := endpoint.ComputeRetransmitTimeout(packet)
	ctx, cancel := context.WithTimeout(rs.ctx, rto)

	timeout := NewTimeout()
	timeout.SetRTO(rto)
	timeout.ctx = ctx
	timeout.cancel = cancel
	packet.setTimeout(timeout)

	rs.packets.Set(packet.SequenceID(), packet)
	go rs.start(packet)
}

// AcknowledgePacket marks a pending packet as acknowledged. It will be ignored at the next resend attempt
func (rs *ResendScheduler) AcknowledgePacket(sequenceID uint16) {
	if packet, ok := rs.packets.Get(sequenceID); ok {
		// * Acknowledge the packet
		rs.packets.Delete(sequenceID)

		// * Update the RTT on the connection if the packet hasn't been resent
		if packet.SendCount() <= rs.streamSettings.RTTRetransmit {
			rttm := time.Since(packet.SentAt())
			packet.Sender().(*PRUDPConnection).rtt.Adjust(rttm)
		}
	}
}

func (rs *ResendScheduler) start(packet PRUDPPacketInterface) {
	<-packet.getTimeout().ctx.Done()

	connection := packet.Sender().(*PRUDPConnection)
	connection.Lock()
	defer connection.Unlock()

	// * If the connection is closed stop trying to resend
	if connection.ConnectionState != StateConnected {
		return
	}

	if rs.packets.Has(packet.SequenceID()) {
		// * This is `<` instead of `<=` for accuracy with observed behavior, even though we're comparing send count vs _resend_ max
		if packet.SendCount() < rs.streamSettings.MaxPacketRetransmissions {
			endpoint := packet.Sender().Endpoint().(*PRUDPEndPoint)

			packet.incrementSendCount()
			packet.setSentAt(time.Now())
			rto := endpoint.ComputeRetransmitTimeout(packet)

			ctx, cancel := context.WithTimeout(rs.ctx, rto)
			timeout := packet.getTimeout()
			timeout.timeout = rto
			timeout.ctx = ctx
			timeout.cancel = cancel

			// * Schedule the packet to be resent
			go rs.start(packet)

			// * Resend the packet to the connection
			server := connection.endpoint.Server
			data := packet.Bytes()
			server.sendRaw(connection.Socket, data)
		} else {
			// * Packet has been retried too many times, consider the connection dead
			connection.cleanup()
		}
	}
}

// Stop kills the resend scheduler and stops all pending packets
func (rs *ResendScheduler) Stop() {
	rs.cancel()
	rs.packets.Clear(func(key uint16, value PRUDPPacketInterface) {})
}

// NewResendScheduler creates a new ResendScheduler
func NewResendScheduler() *ResendScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &ResendScheduler{
		ctx:            ctx,
		cancel:         cancel,
		packets:        NewMutexMap[uint16, PRUDPPacketInterface](),
		streamSettings: NewStreamSettings(),
	}
}
