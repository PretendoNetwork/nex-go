package nex

import (
	"context"
	"time"
)

// TimeoutManager is an implementation of rdv::TimeoutManager and manages the resending of reliable PRUDP packets
type TimeoutManager struct {
	ctx            context.Context
	cancel         context.CancelFunc
	packets        *MutexMap[uint16, PRUDPPacketInterface]
	streamSettings *StreamSettings
}

// SchedulePacketTimeout adds a packet to the scheduler and begins it's timer
func (tm *TimeoutManager) SchedulePacketTimeout(packet PRUDPPacketInterface) {
	endpoint := packet.Sender().Endpoint().(*PRUDPEndPoint)

	rto := endpoint.ComputeRetransmitTimeout(packet)
	ctx, cancel := context.WithTimeout(tm.ctx, rto)

	timeout := NewTimeout()
	timeout.SetRTO(rto)
	timeout.ctx = ctx
	timeout.cancel = cancel
	packet.setTimeout(timeout)

	tm.packets.Set(packet.SequenceID(), packet)
	go tm.start(packet)
}

// AcknowledgePacket marks a pending packet as acknowledged. It will be ignored at the next resend attempt
func (tm *TimeoutManager) AcknowledgePacket(sequenceID uint16) {
	// * Acknowledge the packet
	tm.packets.RunAndDelete(sequenceID, func(_ uint16, packet PRUDPPacketInterface) {
		// * Update the RTT on the connection if the packet hasn't been resent
		if packet.SendCount() >= tm.streamSettings.RTTRetransmit {
			rttm := time.Since(packet.SentAt())
			packet.Sender().(*PRUDPConnection).rtt.Adjust(rttm)
		}
	})
}

func (tm *TimeoutManager) start(packet PRUDPPacketInterface) {
	<-packet.getTimeout().ctx.Done()

	connection := packet.Sender().(*PRUDPConnection)

	// * If the connection is closed stop trying to resend
	if connection.ConnectionState != StateConnected {
		return
	}

	if tm.packets.Has(packet.SequenceID()) {
		endpoint := packet.Sender().Endpoint().(*PRUDPEndPoint)

		// * This is `<` instead of `<=` for accuracy with observed behavior, even though we're comparing send count vs _resend_ max
		if packet.SendCount() < tm.streamSettings.MaxPacketRetransmissions {
			packet.incrementSendCount()
			packet.setSentAt(time.Now())
			rto := endpoint.ComputeRetransmitTimeout(packet)

			ctx, cancel := context.WithTimeout(tm.ctx, rto)
			timeout := packet.getTimeout()
			timeout.timeout = rto
			timeout.ctx = ctx
			timeout.cancel = cancel

			// * Schedule the packet to be resent
			go tm.start(packet)

			// * Resend the packet to the connection
			server := connection.endpoint.Server
			data := packet.Bytes()
			server.sendRaw(connection.Socket, data)
		} else {
			// * Packet has been retried too many times, consider the connection dead
			endpoint.cleanupConnection(connection)
		}
	}
}

// Stop kills the resend scheduler and stops all pending packets
func (tm *TimeoutManager) Stop() {
	tm.cancel()
	tm.packets.Clear(func(key uint16, value PRUDPPacketInterface) {})
}

// NewTimeoutManager creates a new TimeoutManager
func NewTimeoutManager() *TimeoutManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TimeoutManager{
		ctx:            ctx,
		cancel:         cancel,
		packets:        NewMutexMap[uint16, PRUDPPacketInterface](),
		streamSettings: NewStreamSettings(),
	}
}
