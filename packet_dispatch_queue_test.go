package nex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReorderPackets(t *testing.T) {
	pdq := NewPacketDispatchQueue()

	packet1 := makePacket(2)
	packet2 := makePacket(3)
	packet3 := makePacket(4)

	pdq.Queue(packet2)
	pdq.Queue(packet3)
	pdq.Queue(packet1)

	result, ok := pdq.GetNextToDispatch()
	assert.True(t, ok)
	assert.Equal(t, packet1, result)
	pdq.Dispatched(packet1)

	result, ok = pdq.GetNextToDispatch()
	assert.True(t, ok)
	assert.Equal(t, packet2, result)
	pdq.Dispatched(packet2)

	result, ok = pdq.GetNextToDispatch()
	assert.True(t, ok)
	assert.Equal(t, packet3, result)
	pdq.Dispatched(packet3)

	result, ok = pdq.GetNextToDispatch()
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestCallingInLoop(t *testing.T) {
	pdq := NewPacketDispatchQueue()

	packet1 := makePacket(2)
	packet2 := makePacket(3)
	packet3 := makePacket(4)

	pdq.Queue(packet2)
	pdq.Queue(packet3)
	pdq.Queue(packet1)

	for nextPacket, ok := pdq.GetNextToDispatch(); ok; nextPacket, ok = pdq.GetNextToDispatch() {
		pdq.Dispatched(nextPacket)
	}

	assert.Equal(t, uint16(5), pdq.nextExpectedSequenceId.Value)
}

func makePacket(sequenceID uint16) PRUDPPacketInterface {
	packet, _ := NewPRUDPPacketV0(nil, nil, nil)
	packet.SetSequenceID(sequenceID)

	return packet
}
