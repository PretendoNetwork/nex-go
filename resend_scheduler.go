package nex

import (
	"context"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
)

type resendSchedulerState uint8

type SendPacketFn func(packet PRUDPPacketInterface)

const (
	NotStarted resendSchedulerState = iota
	Running
	Stopped
)

type pendingPacket struct {
	resultChannel chan<- PacketSendResult
	cancel        context.CancelFunc
	packet        PRUDPPacketInterface
}

func newPendingPacket(packet PRUDPPacketInterface) *pendingPacket {
	return &pendingPacket{
		packet: packet,
	}
}

func (pp *pendingPacket) signalSuccess() {
	if pp.resultChannel != nil {
		pp.resultChannel <- PacketSendResult{IsSuccess: true}
	}
}

func (pp *pendingPacket) signalFailure(err error) {
	if pp.resultChannel != nil {
		pp.resultChannel <- PacketSendResult{Err: err}
	}
}

// ResendScheduler manages the submission of reliable packets.
type ResendScheduler struct {
	pendingPackets  *MutexMap[uint16, *pendingPacket]
	incomingPackets chan *pendingPacket
	settings        *StreamSettings
	closeConnection context.CancelFunc
	cancel          context.CancelFunc
	sendPacket      SendPacketFn //! This should be implemented by some other interface
	state           resendSchedulerState
}

// PacketSendResult is a struct representing the final state of a
// sent packet.
type PacketSendResult struct {
	// Err is populated with an Error describing the failure if IsSuccess is false.
	Err error
	// IsSuccess indicates whether the packet was successfully acknowledged or not.
	IsSuccess bool
}

// AddPacket queues a packet to be sent. Each added packet will be sent
// repeatedly with an increasing interval between retries. AddPacket returns
// a bool which signals whether the given packet was accepted for processing.
//
// Add packet optionally takes a resultChannel parameter which will report
// the final state of any packet that gets accepted.
//
// Terminal states of a packet are
//  1. The packet has been acknowledged
//  2. The maximum number of retries is reached
//  3. The ResendScheduler is stopped
func (rs *ResendScheduler) AddPacket(packet PRUDPPacketInterface, resultChannel chan<- PacketSendResult) bool {
	if rs.state == Stopped {
		return false
	}

	pendingPacket := newPendingPacket(packet)
	pendingPacket.resultChannel = resultChannel
	rs.incomingPackets <- pendingPacket
	return true
}

// AcknowledgePacket acknowledges a packet by sequenceID, removing it from
// the collection of packets to send. It returns a bool which reports whether
// there was a matching queued packet.
func (rs *ResendScheduler) AcknowledgePacket(sequenceID uint16) bool {
	pendingPacket, ok := rs.pendingPackets.Get(sequenceID)

	if ok {
		pendingPacket.cancel()
		rs.pendingPackets.Delete(sequenceID)
		pendingPacket.signalSuccess()
	}

	return ok
}

// AcknowledgeUpTo acknowledges all pending packets with sequenceIDs up to
// and including the given sequenceID.
//
// AcknowledgeUpTo returns a bool which indicates whether the scheduler
// was running.
func (rs *ResendScheduler) AcknowledgeUpTo(sequenceID uint16) bool {
	if rs.state == Stopped {
		return false
	}

	sequenceIDs := make([]uint16, 0)

	rs.pendingPackets.Each(func(id uint16, _ *pendingPacket) bool {
		if id <= sequenceID {
			sequenceIDs = append(sequenceIDs, id)
		}
		return false
	})

	rs.AcknowledgeMany(sequenceIDs)

	return true
}

// AcknowledgeMany acknowledges all pending packets with sequenceIDs in the
// given slice.
//
// AcknowledgeMany returns a bool which indicates whether the scheduler
// was running.
func (rs *ResendScheduler) AcknowledgeMany(sequenceIDs []uint16) bool {
	if rs.state == Stopped {
		return false
	}

	for _, sequenceID := range sequenceIDs {
		rs.AcknowledgePacket(sequenceID)
	}

	return true
}

// IsRunning reports whether the ResendScheduler is still able to accept packets.
func (rs *ResendScheduler) IsRunning() bool {
	return rs.state == Running
}

// Start starts the internal goroutines of the ResendScheduler which attempts to send
// queued packets.
func (rs *ResendScheduler) Start(ctx context.Context, clock clockwork.Clock) {
	if rs.state == NotStarted {
		newCtx, cancel := context.WithCancel(ctx)
		rs.cancel = cancel
		rs.state = Running
		go rs.run(newCtx, clock)
	}
}

// Stop stops the ResendScheduler preventing new packets being added
// and cancelling all internal goroutines.
func (rs *ResendScheduler) Stop() {
	if rs.state == Running {
		rs.state = Stopped
		rs.cancel()
	}
}

// NewResendScheduler creates a new ResendScheduler
func NewResendScheduler(
	settings *StreamSettings,
	closeConnection context.CancelFunc,
	sendPacket SendPacketFn,
) *ResendScheduler {
	scheduler := &ResendScheduler{
		pendingPackets:  NewMutexMap[uint16, *pendingPacket](),
		incomingPackets: make(chan *pendingPacket),
		settings:        settings,
		state:           NotStarted,
		closeConnection: closeConnection,
		sendPacket:      sendPacket,
	}

	return scheduler
}

func (rs *ResendScheduler) run(ctx context.Context, clock clockwork.Clock) {
	for {
		select {
		case <-ctx.Done():
			rs.state = Stopped
			rs.pendingPackets.Clear(func(id uint16, pending *pendingPacket) {
				err := fmt.Errorf("scheduler was stopped before packet with id %v was acknowledged", id)
				pending.signalFailure(err)
			})
			return
		case pending, ok := <-rs.incomingPackets:
			if ok {
				newCtx, cancel := context.WithCancel(ctx)
				pending.cancel = cancel
				rs.pendingPackets.Set(pending.packet.SequenceID(), pending)
				go rs.startSubmissions(newCtx, clock, pending)
			}
		}
	}
}

func (rs *ResendScheduler) startSubmissions(ctx context.Context, clock clockwork.Clock, pending *pendingPacket) {
	attempts := uint32(0)
	interval := time.Duration(rs.settings.KeepAliveTimeout) * time.Millisecond
	timer := clock.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.Chan():
			if attempts >= rs.settings.MaxPacketRetransmissions {
				err := fmt.Errorf("packet with id %v wasn't acknowledged in time", pending.packet.SequenceID())
				rs.closeConnection()
				pending.signalFailure(err)
				return
			}

			rs.sendPacket(pending.packet)
			attempts++

			retransmitTimeoutMultiplier := float32(rs.settings.ExtraRestransmitTimeoutTrigger)
			if attempts < rs.settings.ExtraRestransmitTimeoutTrigger {
				retransmitTimeoutMultiplier = rs.settings.RetransmitTimeoutMultiplier
			}
			interval += time.Duration(uint32(float32(rs.settings.KeepAliveTimeout)*retransmitTimeoutMultiplier)) * time.Millisecond

			timer.Reset(interval)
		}
	}
}
