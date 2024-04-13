package nex

import (
	"fmt"
	"time"
)

// TODO - REMOVE THIS ENTIRELY AND REPLACE IT WITH AN IMPLEMENTATION OF rdv::Timeout AND rdv::TimeoutManager AND USE MORE STREAM SETTINGS!

// PendingPacket represends a packet scheduled to be resent
type PendingPacket struct {
	packet         PRUDPPacketInterface
	lastSendTime   time.Time
	resendCount    uint32
	isAcknowledged bool
	interval       time.Duration
	ticker         *time.Ticker
	rs             *ResendScheduler
}

func (pi *PendingPacket) startResendTimer() {
	pi.lastSendTime = time.Now()
	pi.ticker = time.NewTicker(pi.interval)

	for range pi.ticker.C {
		finished := false

		if pi.isAcknowledged {
			pi.ticker.Stop()
			pi.rs.packets.Delete(pi.packet.SequenceID())
			finished = true
		} else {
			finished = pi.rs.resendPacket(pi)
		}

		if finished {
			return
		}
	}
}

// ResendScheduler manages the resending of reliable PRUDP packets
type ResendScheduler struct {
	packets *MutexMap[uint16, *PendingPacket]
}

// Stop kills the resend scheduler and stops all pending packets
func (rs *ResendScheduler) Stop() {
	stillPending := make([]uint16, rs.packets.Size())

	rs.packets.Each(func(sequenceID uint16, packet *PendingPacket) bool {
		if !packet.isAcknowledged {
			stillPending = append(stillPending, sequenceID)
		}

		return false
	})

	for _, sequenceID := range stillPending {
		if pendingPacket, ok := rs.packets.Get(sequenceID); ok {
			pendingPacket.isAcknowledged = true // * Prevent an edge case where the ticker is already being processed

			if pendingPacket.ticker != nil {
				// * This should never happen, but popped up in CTGP-7 testing?
				// * Did the GC clear this before we called it?
				pendingPacket.ticker.Stop()
			}

			rs.packets.Delete(sequenceID)
		}
	}
}

// AddPacket adds a packet to the scheduler and begins it's timer
func (rs *ResendScheduler) AddPacket(packet PRUDPPacketInterface) {
	connection := packet.Sender().(*PRUDPConnection)
	slidingWindow := connection.SlidingWindow(packet.SubstreamID())

	pendingPacket := &PendingPacket{
		packet: packet,
		rs:     rs,
		// TODO: This may not be accurate, needs more research
		interval: time.Duration(slidingWindow.streamSettings.KeepAliveTimeout) * time.Millisecond,
	}

	rs.packets.Set(packet.SequenceID(), pendingPacket)

	go pendingPacket.startResendTimer()
}

// AcknowledgePacket marks a pending packet as acknowledged. It will be ignored at the next resend attempt
func (rs *ResendScheduler) AcknowledgePacket(sequenceID uint16) {
	if pendingPacket, ok := rs.packets.Get(sequenceID); ok {
		pendingPacket.isAcknowledged = true
	}
}

func (rs *ResendScheduler) resendPacket(pendingPacket *PendingPacket) bool {
	if pendingPacket.isAcknowledged {
		// * Prevent a race condition where resendPacket may be called
		// * at the same time a packet is acknowledged
		return false
	}

	packet := pendingPacket.packet
	connection := packet.Sender().(*PRUDPConnection)
	slidingWindow := connection.SlidingWindow(packet.SubstreamID())

	if pendingPacket.resendCount >= slidingWindow.streamSettings.MaxPacketRetransmissions {
		// * The maximum resend count has been reached, consider the connection dead.
		pendingPacket.ticker.Stop()
		rs.packets.Delete(packet.SequenceID())
		connection.cleanup() // * "removed" event is dispatched here

		streamType := packet.SourceVirtualPortStreamType()
		streamID := packet.SourceVirtualPortStreamID()
		discriminator := fmt.Sprintf("%s-%d-%d", packet.Sender().Address().String(), streamType, streamID)

		connection.endpoint.Connections.Delete(discriminator)

		return true
	}

	// TODO: This may not be accurate, needs more research
	if time.Since(pendingPacket.lastSendTime) >= time.Duration(slidingWindow.streamSettings.KeepAliveTimeout)*time.Millisecond {
		// * Resend the packet to the connection
		server := connection.endpoint.Server
		data := packet.Bytes()
		server.sendRaw(connection.Socket, data)

		pendingPacket.resendCount++

		var retransmitTimeoutMultiplier float32
		if pendingPacket.resendCount < slidingWindow.streamSettings.ExtraRestransmitTimeoutTrigger {
			retransmitTimeoutMultiplier = slidingWindow.streamSettings.RetransmitTimeoutMultiplier
		} else {
			retransmitTimeoutMultiplier = slidingWindow.streamSettings.ExtraRetransmitTimeoutMultiplier
		}
		pendingPacket.interval += time.Duration(uint32(float32(slidingWindow.streamSettings.KeepAliveTimeout)*retransmitTimeoutMultiplier)) * time.Millisecond

		pendingPacket.ticker.Reset(pendingPacket.interval)
		pendingPacket.lastSendTime = time.Now()
	}

	return false
}

// NewResendScheduler creates a new ResendScheduler
func NewResendScheduler() *ResendScheduler {
	return &ResendScheduler{
		packets: NewMutexMap[uint16, *PendingPacket](),
	}
}
