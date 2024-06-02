package nex

import (
	"context"

	"github.com/jonboulle/clockwork"
)

// SlidingWindow is an implementation of rdv::SlidingWindow.
// Currently this is a stub and does not reflect the interface and usage of rdv:SlidingWindow.
// In the original library this is used to manage sequencing of outgoing packets.
// each virtual connection stream only uses a single SlidingWindow, but starting
// in PRUDPv1 with NEX virtual connections may have multiple reliable substreams and thus multiple SlidingWindows.
type SlidingWindow struct {
	sequenceIDCounter *Counter[uint16]
	streamSettings    *StreamSettings
	ResendScheduler   *ResendScheduler
}

// SetCipherKey sets the reliable substreams RC4 cipher keys
func (sw *SlidingWindow) SetCipherKey(key []byte) {
	sw.streamSettings.EncryptionAlgorithm.SetKey(key)
}

// NextOutgoingSequenceID sets the reliable substreams RC4 cipher keys
func (sw *SlidingWindow) NextOutgoingSequenceID() uint16 {
	return sw.sequenceIDCounter.Next()
}

// Decrypt decrypts the provided data with the substreams decipher
func (sw *SlidingWindow) Decrypt(data []byte) ([]byte, error) {
	return sw.streamSettings.EncryptionAlgorithm.Decrypt(data)
}

// Encrypt encrypts the provided data with the substreams cipher
func (sw *SlidingWindow) Encrypt(data []byte) ([]byte, error) {
	return sw.streamSettings.EncryptionAlgorithm.Encrypt(data)
}

// NewSlidingWindow initializes a new SlidingWindow with a starting counter value.
func NewSlidingWindow(connection *PRUDPConnection) *SlidingWindow {

	// TODO: Make PRUDPConnection use Context, this should just be the CancelFn for that Context
	//! There is currently a minor bug in both the original implementation and this implementation
	//! where `PRUDPConnection.cleanup` is called multiple times for most multifragment payloads
	//! this triggers multiple `connectionEnded` events
	//!
	//! This would be fixed by moving to context model as cancel calls are idempotent
	closeConnection := func() {
		connection.cleanup()
		connection.endpoint.deleteConnectionByID(connection.ID)
	}

	// TODO: This should belong on an interface, possibly the ConnectionInterface but I couldn't find
	// TODO: a good place to put it, and this is only a POC
	sendPacket := func(packet PRUDPPacketInterface) {
		connection := packet.Sender().(*PRUDPConnection)
		server := connection.endpoint.Server
		data := packet.Bytes()
		server.sendRaw(connection.Socket, data)
	}

	resendScheduler := NewResendScheduler(
		connection.StreamSettings.Copy(),
		closeConnection,
		sendPacket,
	)

	resendScheduler.Start(context.Background(), clockwork.NewRealClock())

	sw := &SlidingWindow{
		sequenceIDCounter: NewCounter[uint16](0),
		ResendScheduler:   resendScheduler,
	}

	return sw
}
