package nex

// SlidingWindow is an implementation of rdv::SlidingWindow.
// SlidingWindow reorders pending reliable packets to ensure they are handled in the expected order.
// In the original library each virtual connection stream only uses a single SlidingWindow, but starting
// in PRUDPv1 with NEX virtual connections may have multiple reliable substreams and thus multiple SlidingWindows.
type SlidingWindow struct {
	pendingPackets            *MutexMap[uint16, PRUDPPacketInterface]
	incomingSequenceIDCounter *Counter[uint16]
	outgoingSequenceIDCounter *Counter[uint16]
	streamSettings            *StreamSettings
	fragmentedPayload         []byte
	ResendScheduler           *ResendScheduler
}

// Update adds an incoming packet to the list of known packets and returns a list of packets to be processed in order
func (sw *SlidingWindow) Update(packet PRUDPPacketInterface) []PRUDPPacketInterface {
	packets := make([]PRUDPPacketInterface, 0)

	if packet.SequenceID() >= sw.incomingSequenceIDCounter.Value && !sw.pendingPackets.Has(packet.SequenceID()) {
		sw.pendingPackets.Set(packet.SequenceID(), packet)

		for sw.pendingPackets.Has(sw.incomingSequenceIDCounter.Value) {
			storedPacket, _ := sw.pendingPackets.Get(sw.incomingSequenceIDCounter.Value)
			packets = append(packets, storedPacket)
			sw.pendingPackets.Delete(sw.incomingSequenceIDCounter.Value)
			sw.incomingSequenceIDCounter.Next()
		}
	}

	return packets
}

// SetCipherKey sets the reliable substreams RC4 cipher keys
func (sw *SlidingWindow) SetCipherKey(key []byte) {
	sw.streamSettings.EncryptionAlgorithm.SetKey(key)
}

// NextOutgoingSequenceID sets the reliable substreams RC4 cipher keys
func (sw *SlidingWindow) NextOutgoingSequenceID() uint16 {
	return sw.outgoingSequenceIDCounter.Next()
}

// Decrypt decrypts the provided data with the substreams decipher
func (sw *SlidingWindow) Decrypt(data []byte) ([]byte, error) {
	return sw.streamSettings.EncryptionAlgorithm.Decrypt(data)
}

// Encrypt encrypts the provided data with the substreams cipher
func (sw *SlidingWindow) Encrypt(data []byte) ([]byte, error) {
	return sw.streamSettings.EncryptionAlgorithm.Encrypt(data)
}

// AddFragment adds the given fragment to the substreams fragmented payload
// Returns the current fragmented payload
func (sw *SlidingWindow) AddFragment(fragment []byte) []byte {
	sw.fragmentedPayload = append(sw.fragmentedPayload, fragment...)

	return sw.fragmentedPayload
}

// ResetFragmentedPayload resets the substreams fragmented payload
func (sw *SlidingWindow) ResetFragmentedPayload() {
	sw.fragmentedPayload = make([]byte, 0)
}

// NewSlidingWindow initializes a new SlidingWindow with a starting counter value.
func NewSlidingWindow() *SlidingWindow {
	sw := &SlidingWindow{
		pendingPackets:            NewMutexMap[uint16, PRUDPPacketInterface](),
		incomingSequenceIDCounter: NewCounter[uint16](0),
		outgoingSequenceIDCounter: NewCounter[uint16](0),
		ResendScheduler:           NewResendScheduler(),
	}

	return sw
}
