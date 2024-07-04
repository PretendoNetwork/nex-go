package nex

// SlidingWindow is an implementation of rdv::SlidingWindow.
// Currently this is a stub and does not reflect the interface and usage of rdv:SlidingWindow.
// In the original library this is used to manage sequencing of outgoing packets.
// each virtual connection stream only uses a single SlidingWindow, but starting
// in PRUDPv1 with NEX virtual connections may have multiple reliable substreams and thus multiple SlidingWindows.
type SlidingWindow struct {
	sequenceIDCounter *Counter[uint16]
	streamSettings    *StreamSettings
	TimeoutManager    *TimeoutManager
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
func NewSlidingWindow() *SlidingWindow {
	sw := &SlidingWindow{
		sequenceIDCounter: NewCounter[uint16](0),
		TimeoutManager:    NewTimeoutManager(),
	}

	return sw
}
