package nex

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
)

func TestPacketIsSentAndAcknowledged(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	connectionClosed, closeConnection := createCloseFunction()

	var resendScheduler *ResendScheduler

	sentPackets := make([]PRUDPPacketInterface, 0)
	sendMux := sync.Mutex{}
	sendFn := func(packet PRUDPPacketInterface) {
		sendMux.Lock()
		defer sendMux.Unlock()
		sentPackets = append(sentPackets, packet)
		resendScheduler.AcknowledgePacket(packet.SequenceID())
	}

	resendScheduler = NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)
	packet, _ := NewPRUDPPacketV0(nil, nil, nil)

	results := make(chan PacketSendResult)
	ok := resendScheduler.AddPacket(packet, results)
	assert.True(t, ok)

	clock.Advance(time.Millisecond * time.Duration(settings.KeepAliveTimeout))

	result := <-results
	assert.True(t, result.IsSuccess)
	assert.Nil(t, result.Err)

	resendScheduler.Stop()

	assert.False(t, *connectionClosed)
	assert.Equal(t, 1, len(sentPackets))
	assert.Equal(t, packet, sentPackets[0])
}

func TestPacketIsCancelledAfterTooManyRetries(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	connectionClosed, closeConnection := createCloseFunction()

	var resendScheduler *ResendScheduler
	sentPackets := make([]PRUDPPacketInterface, 0)
	sendMux := sync.Mutex{}
	sendFn := func(packet PRUDPPacketInterface) {
		sendMux.Lock()
		defer sendMux.Unlock()
		sentPackets = append(sentPackets, packet)
	}

	resendScheduler = NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)
	packet, _ := NewPRUDPPacketV0(nil, nil, nil)

	before := clock.Now()

	results := make(chan PacketSendResult)
	ok := resendScheduler.AddPacket(packet, results)
	assert.True(t, ok)

	result := runClockUntilReturn(clock, results)

	after := clock.Now()

	assert.InDelta(t, time.Second*283, after.Sub(before), float64(time.Second))

	assert.False(t, result.IsSuccess)
	assert.Error(t, result.Err, "packet with id 0 wasn't acknowledged in time")
	assert.True(t, *connectionClosed)

	resendScheduler.Stop()

	assert.Equal(t, 20, len(sentPackets))
	for _, sentPacket := range sentPackets {
		assert.Equal(t, packet, sentPacket)
	}
}

func TestStop(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	closeConnection := func() {}
	sendFn := func(packet PRUDPPacketInterface) {}

	resendScheduler := NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)
	packet, _ := NewPRUDPPacketV0(nil, nil, nil)

	results := make(chan PacketSendResult)
	ok := resendScheduler.AddPacket(packet, results)
	assert.True(t, ok)

	resendScheduler.Stop()

	result := <-results
	assert.False(t, result.IsSuccess)
	assert.Error(t, result.Err, "scheduler was stopped before packet with id 0 was acknowledged")
}

func TestCancel(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	ctx, cancel := context.WithCancel(context.Background())
	closeConnection := func() {}
	sendFn := func(packet PRUDPPacketInterface) {}

	resendScheduler := NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(ctx, clock)
	packet, _ := NewPRUDPPacketV0(nil, nil, nil)

	results := make(chan PacketSendResult)
	ok := resendScheduler.AddPacket(packet, results)
	assert.True(t, ok)

	assert.True(t, resendScheduler.IsRunning())

	cancel()
	result := <-results
	assert.False(t, result.IsSuccess)
	assert.Error(t, result.Err, "scheduler was stopped before packet with id 0 was acknowledged")

	assert.False(t, resendScheduler.IsRunning())
}

func TestAcknowledgeMany(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	closeConnection := func() {}
	sendFn := func(packet PRUDPPacketInterface) {}

	resendScheduler := NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)

	packet1, _ := NewPRUDPPacketV0(nil, nil, nil)
	results1 := make(chan PacketSendResult, 1)
	ok := resendScheduler.AddPacket(packet1, results1)
	assert.True(t, ok)

	packet2, _ := NewPRUDPPacketV0(nil, nil, nil)
	packet2.SetSequenceID(1)
	results2 := make(chan PacketSendResult, 1)
	ok = resendScheduler.AddPacket(packet2, results2)
	assert.True(t, ok)

	packet3, _ := NewPRUDPPacketV0(nil, nil, nil)
	packet3.SetSequenceID(2)
	results3 := make(chan PacketSendResult, 1)
	ok = resendScheduler.AddPacket(packet3, results3)
	assert.True(t, ok)

	ok = resendScheduler.AcknowledgeMany([]uint16{0, 2})
	assert.True(t, ok)

	result1 := runClockUntilReturn(clock, results1)
	assert.True(t, result1.IsSuccess)

	result3 := runClockUntilReturn(clock, results3)
	assert.True(t, result3.IsSuccess)

	resendScheduler.Stop()

	result2 := runClockUntilReturn(clock, results2)
	assert.False(t, result2.IsSuccess)
}

func TestAcknowledgeUpTo(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	closeConnection := func() {}
	sendFn := func(packet PRUDPPacketInterface) {}

	resendScheduler := NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)

	packet1, _ := NewPRUDPPacketV0(nil, nil, nil)
	results1 := make(chan PacketSendResult, 1)
	ok := resendScheduler.AddPacket(packet1, results1)
	assert.True(t, ok)

	packet2, _ := NewPRUDPPacketV0(nil, nil, nil)
	packet2.SetSequenceID(1)
	results2 := make(chan PacketSendResult, 1)
	ok = resendScheduler.AddPacket(packet2, results2)
	assert.True(t, ok)

	packet3, _ := NewPRUDPPacketV0(nil, nil, nil)
	packet3.SetSequenceID(2)
	results3 := make(chan PacketSendResult, 1)
	ok = resendScheduler.AddPacket(packet3, results3)
	assert.True(t, ok)

	ok = resendScheduler.AcknowledgeUpTo(1)
	assert.True(t, ok)

	result1 := runClockUntilReturn(clock, results1)
	assert.True(t, result1.IsSuccess)

	result2 := runClockUntilReturn(clock, results2)
	assert.True(t, result2.IsSuccess)

	resendScheduler.Stop()

	result3 := runClockUntilReturn(clock, results3)
	assert.False(t, result3.IsSuccess)
}

func TestAddAndAcknowledgeFailAfterStop(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	closeConnection := func() {}
	sendFn := func(packet PRUDPPacketInterface) {}

	resendScheduler := NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)
	packet, _ := NewPRUDPPacketV0(nil, nil, nil)

	resendScheduler.Stop()

	results := make(chan PacketSendResult)
	ok := resendScheduler.AddPacket(packet, results)
	assert.False(t, ok)
	ok = resendScheduler.AcknowledgePacket(packet.SequenceID())
	assert.False(t, ok)
	ok = resendScheduler.AcknowledgeMany([]uint16{0})
	assert.False(t, ok)
	ok = resendScheduler.AcknowledgeUpTo(0)
	assert.False(t, ok)
}

func TestStopIsIdempotent(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	closeConnection := func() {}
	sendFn := func(packet PRUDPPacketInterface) {}

	resendScheduler := NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)
	resendScheduler.Stop()
	resendScheduler.Stop()
}

func TestIsRunning(t *testing.T) {
	clock := clockwork.NewFakeClock()
	settings := NewStreamSettings()
	closeConnection := func() {}
	sendFn := func(packet PRUDPPacketInterface) {}

	resendScheduler := NewResendScheduler(settings, closeConnection, sendFn)
	resendScheduler.Start(context.Background(), clock)

	assert.True(t, resendScheduler.IsRunning())
	resendScheduler.Stop()
	assert.False(t, resendScheduler.IsRunning())
}

func runClockUntilReturn[A any](clock clockwork.FakeClock, c <-chan A) A {
	for {
		select {
		case result := <-c:
			return result
		default:
			clock.Advance(time.Microsecond)
		}
	}
}

func createCloseFunction() (*bool, func()) {
	called := false
	closeFn := func() {
		called = true
	}
	return &called, closeFn
}
