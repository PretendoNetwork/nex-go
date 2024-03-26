package nex

import (
	"github.com/PretendoNetwork/nex-go/compression"
	"github.com/PretendoNetwork/nex-go/encryption"
)

// StreamSettings is an implementation of rdv::StreamSettings.
// StreamSettings holds the state and settings for a PRUDP virtual connection stream.
// Each virtual connection is composed of a virtual port and stream type.
// In the original library this would be tied to a rdv::Stream class, but here it is not.
// The original library has more settings which are not present here as their use is unknown.
// Not all values are used at this time, and only exist to future-proof for a later time.
type StreamSettings struct {
	ExtraRestransmitTimeoutTrigger   uint32                // * The number of times a packet can be retransmitted before ExtraRetransmitTimeoutMultiplier is used
	MaxPacketRetransmissions         uint32                // * The number of times a packet can be retransmitted before the timeout time is checked
	KeepAliveTimeout                 uint32                // * Presumably the time a packet can be alive for without acknowledgement? Milliseconds?
	ChecksumBase                     uint32                // * Unused. The base value for PRUDPv0 checksum calculations
	FaultDetectionEnabled            bool                  // * Unused. Presumably used to detect PIA faults?
	InitialRTT                       uint32                // * Unused. The connections initial RTT
	EncryptionAlgorithm              encryption.Algorithm  // * The encryption algorithm used for packet payloads
	ExtraRetransmitTimeoutMultiplier float32               // * Used as part of the RTO calculations when retransmitting a packet. Only used if ExtraRestransmitTimeoutTrigger has been reached
	WindowSize                       uint32                // * Unused. The max number of (reliable?) packets allowed in a SlidingWindow
	CompressionAlgorithm             compression.Algorithm // * The compression algorithm used for packet payloads
	RTTRetransmit                    uint32                // * Unused. Unknown use
	RetransmitTimeoutMultiplier      float32               // * Used as part of the RTO calculations when retransmitting a packet. Only used if ExtraRestransmitTimeoutTrigger has not been reached
	MaxSilenceTime                   uint32                // * Presumably the time a connection can go without any packets from the other side? Milliseconds?
}

// Copy returns a new copy of the settings
func (ss *StreamSettings) Copy() *StreamSettings {
	copied := NewStreamSettings()

	copied.ExtraRestransmitTimeoutTrigger = ss.ExtraRestransmitTimeoutTrigger
	copied.MaxPacketRetransmissions = ss.MaxPacketRetransmissions
	copied.KeepAliveTimeout = ss.KeepAliveTimeout
	copied.ChecksumBase = ss.ChecksumBase
	copied.FaultDetectionEnabled = ss.FaultDetectionEnabled
	copied.InitialRTT = ss.InitialRTT
	copied.EncryptionAlgorithm = ss.EncryptionAlgorithm.Copy()
	copied.ExtraRetransmitTimeoutMultiplier = ss.ExtraRetransmitTimeoutMultiplier
	copied.WindowSize = ss.WindowSize
	copied.CompressionAlgorithm = ss.CompressionAlgorithm.Copy()
	copied.RTTRetransmit = ss.RTTRetransmit
	copied.RetransmitTimeoutMultiplier = ss.RetransmitTimeoutMultiplier
	copied.MaxSilenceTime = ss.MaxSilenceTime

	return copied
}

// NewStreamSettings returns a new instance of StreamSettings with default params
func NewStreamSettings() *StreamSettings {
	// * Default values based on WATCH_DOGS. Not all values are used currently, and only
	// * exist to mimic what is seen in that game. Many are planned for future use.
	return &StreamSettings{
		ExtraRestransmitTimeoutTrigger:   0x32,
		MaxPacketRetransmissions:         0x14,
		KeepAliveTimeout:                 1000,
		ChecksumBase:                     0,
		FaultDetectionEnabled:            true,
		InitialRTT:                       0xFA,
		EncryptionAlgorithm:              encryption.NewRC4Encryption(),
		ExtraRetransmitTimeoutMultiplier: 1.0,
		WindowSize:                       8,
		CompressionAlgorithm:             compression.NewDummyCompression(),
		RTTRetransmit:                    0x32,
		RetransmitTimeoutMultiplier:      1.25,
		MaxSilenceTime:                   5000,
	}
}
