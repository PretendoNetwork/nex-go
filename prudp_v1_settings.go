package nex

// TODO - We can also breakout the decoding/encoding functions here too, but that would require getters and setters for all packet fields

// PRUDPV1Settings defines settings for how to handle aspects of PRUDPv1 packets
type PRUDPV1Settings struct {
	LegacyConnectionSignature bool
}

// NewPRUDPV1Settings returns a new PRUDPV1Settings
func NewPRUDPV1Settings() *PRUDPV1Settings {
	return &PRUDPV1Settings{
		LegacyConnectionSignature: false,
	}
}
