package nex

import (
	"crypto/rc4"
	"time"
)

// ReliablePacketSubstreamManager represents a substream manager for reliable PRUDP packets
type ReliablePacketSubstreamManager struct {
	packetMap                 *MutexMap[uint16, PRUDPPacketInterface]
	incomingSequenceIDCounter *Counter[uint16]
	outgoingSequenceIDCounter *Counter[uint16]
	cipher                    *rc4.Cipher
	decipher                  *rc4.Cipher
	fragmentedPayload         []byte
	ResendScheduler           *ResendScheduler
}

// Update adds an incoming packet to the list of known packets and returns a list of packets to be processed in order
func (psm *ReliablePacketSubstreamManager) Update(packet PRUDPPacketInterface) []PRUDPPacketInterface {
	packets := make([]PRUDPPacketInterface, 0)

	if packet.SequenceID() >= psm.incomingSequenceIDCounter.Value && !psm.packetMap.Has(packet.SequenceID()) {
		psm.packetMap.Set(packet.SequenceID(), packet)

		for psm.packetMap.Has(psm.incomingSequenceIDCounter.Value) {
			storedPacket, _ := psm.packetMap.Get(psm.incomingSequenceIDCounter.Value)
			packets = append(packets, storedPacket)
			psm.packetMap.Delete(psm.incomingSequenceIDCounter.Value)
			psm.incomingSequenceIDCounter.Next()
		}
	}

	return packets
}

// SetCipherKey sets the reliable substreams RC4 cipher keys
func (psm *ReliablePacketSubstreamManager) SetCipherKey(key []byte) {
	cipher, _ := rc4.NewCipher(key)
	decipher, _ := rc4.NewCipher(key)

	psm.cipher = cipher
	psm.decipher = decipher
}

// NextOutgoingSequenceID sets the reliable substreams RC4 cipher keys
func (psm *ReliablePacketSubstreamManager) NextOutgoingSequenceID() uint16 {
	return psm.outgoingSequenceIDCounter.Next()
}

// Decrypt decrypts the provided data with the substreams decipher
func (psm *ReliablePacketSubstreamManager) Decrypt(data []byte) []byte {
	ciphered := make([]byte, len(data))

	psm.decipher.XORKeyStream(ciphered, data)

	return ciphered
}

// Encrypt encrypts the provided data with the substreams cipher
func (psm *ReliablePacketSubstreamManager) Encrypt(data []byte) []byte {
	ciphered := make([]byte, len(data))

	psm.cipher.XORKeyStream(ciphered, data)

	return ciphered
}

// AddFragment adds the given fragment to the substreams fragmented payload
// Returns the current fragmented payload
func (psm *ReliablePacketSubstreamManager) AddFragment(fragment []byte) []byte {
	psm.fragmentedPayload = append(psm.fragmentedPayload, fragment...)

	return psm.fragmentedPayload
}

// ResetFragmentedPayload resets the substreams fragmented payload
func (psm *ReliablePacketSubstreamManager) ResetFragmentedPayload() {
	psm.fragmentedPayload = make([]byte, 0)
}

// NewReliablePacketSubstreamManager initializes a new ReliablePacketSubstreamManager with a starting counter value.
func NewReliablePacketSubstreamManager(startingIncomingSequenceID, startingOutgoingSequenceID uint16) *ReliablePacketSubstreamManager {
	psm := &ReliablePacketSubstreamManager{
		packetMap:                 NewMutexMap[uint16, PRUDPPacketInterface](),
		incomingSequenceIDCounter: NewCounter[uint16](startingIncomingSequenceID),
		outgoingSequenceIDCounter: NewCounter[uint16](startingOutgoingSequenceID),
		ResendScheduler:           NewResendScheduler(5, time.Second, 0),
	}

	psm.SetCipherKey([]byte("CD&ML"))

	return psm
}
