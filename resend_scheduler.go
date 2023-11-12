package nex

import (
	"fmt"
	"time"
)

// PendingPacket represends a packet scheduled to be resent
type PendingPacket struct {
	packet         PRUDPPacketInterface
	lastSendTime   time.Time
	resendCount    int
	isAcknowledged bool
	interval       time.Duration
	ticker         *time.Ticker
	rs             *ResendScheduler
}

func (pi *PendingPacket) startResendTimer() {
	pi.lastSendTime = time.Now()
	pi.ticker = time.NewTicker(pi.interval)

	for range pi.ticker.C {
		if pi.isAcknowledged {
			pi.ticker.Stop()
			pi.rs.packets.Delete(pi.packet.SequenceID())
		} else {
			pi.rs.resendPacket(pi)
		}
	}
}

// ResendScheduler manages the resending of reliable PRUDP packets
type ResendScheduler struct {
	packets        *MutexMap[uint16, *PendingPacket]
	MaxResendCount int
	Interval       time.Duration
	Increase       time.Duration
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
			pendingPacket.ticker.Stop()
			rs.packets.Delete(sequenceID)
		}
	}
}

// AddPacket adds a packet to the scheduler and begins it's timer
func (rs *ResendScheduler) AddPacket(packet PRUDPPacketInterface) {
	pendingPacket := &PendingPacket{
		packet:   packet,
		rs:       rs,
		interval: rs.Interval,
	}

	if packet.Sender().Server().(*PRUDPServer).IsSecureServer {
		fmt.Println("[SECR] Adding packet", packet.SequenceID(), "to resend queue")
	} else {
		fmt.Println("[AUTH] Adding packet", packet.SequenceID(), "to resend queue")
	}

	rs.packets.Set(packet.SequenceID(), pendingPacket)

	go pendingPacket.startResendTimer()
}

// AcknowledgePacket marks a pending packet as acknowledged. It will be ignored at the next resend attempt
func (rs *ResendScheduler) AcknowledgePacket(sequenceID uint16) {
	if pendingPacket, ok := rs.packets.Get(sequenceID); ok {
		if pendingPacket.packet.Sender().Server().(*PRUDPServer).IsSecureServer {
			fmt.Println("[SECR] Acknowledged", sequenceID)
		} else {
			fmt.Println("[AUTH] Acknowledged", sequenceID)
		}

		pendingPacket.isAcknowledged = true
	}
}

func (rs *ResendScheduler) resendPacket(pendingPacket *PendingPacket) {
	if pendingPacket.isAcknowledged {
		// * Prevent a race condition where resendPacket may be called
		// * at the same time a packet is acknowledged
		return
	}

	packet := pendingPacket.packet
	client := packet.Sender().(*PRUDPClient)

	if pendingPacket.resendCount >= rs.MaxResendCount {
		// * The maximum resend count has been reached, consider the client dead.
		if pendingPacket.packet.Sender().Server().(*PRUDPServer).IsSecureServer {
			fmt.Println("[SECR] Max resends hit for", pendingPacket.packet.SequenceID())
		} else {
			fmt.Println("[AUTH] Max resends hit for", pendingPacket.packet.SequenceID())
		}

		pendingPacket.ticker.Stop()
		rs.packets.Delete(packet.SequenceID())
		client.cleanup() // * "removed" event is dispatched here
		client.server.clients.Delete(client.address.String())
		return
	}

	if time.Since(pendingPacket.lastSendTime) >= rs.Interval {
		if pendingPacket.packet.Sender().Server().(*PRUDPServer).IsSecureServer {
			fmt.Println("[SECR] Resending packet", pendingPacket.packet.SequenceID())
		} else {
			fmt.Println("[AUTH] Resending packet", pendingPacket.packet.SequenceID())
		}

		// * Resend the packet to the client
		server := client.server
		data := packet.Bytes()
		server.sendRaw(client.Address(), data)

		pendingPacket.interval += rs.Increase
		pendingPacket.ticker.Reset(pendingPacket.interval)
		pendingPacket.resendCount++
		pendingPacket.lastSendTime = time.Now()
	}
}

// NewResendScheduler creates a new ResendScheduler with the provided max resend count and interval and increase durations
//
// If increase is non-zero then every resend will have it's duration increased by that amount. For example an interval of
// 1 second and an increase of 5 seconds. The 1st resend happens after 1 second, the 2nd will take place 6 seconds
// after the 1st, and the 3rd will take place 11 seconds after the 2nd
func NewResendScheduler(maxResendCount int, interval, increase time.Duration) *ResendScheduler {
	return &ResendScheduler{
		packets:        NewMutexMap[uint16, *PendingPacket](),
		MaxResendCount: maxResendCount,
		Interval:       interval,
		Increase:       increase,
	}
}
