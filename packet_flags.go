package nex

const (
	// FlagAck is the ID for the PRUDP Ack Flag
	FlagAck      uint16 = 0x1

	// FlagReliable is the ID for the PRUDP Reliable Flag
	FlagReliable uint16 = 0x2

	// FlagNeedsAck is the ID for the PRUDP NeedsAck Flag
	FlagNeedsAck uint16 = 0x4

	// FlagHasSize is the ID for the PRUDP HasSize Flag
	FlagHasSize  uint16 = 0x8

	// FlagMultiAck is the ID for the PRUDP MultiAck Flag
	FlagMultiAck uint16 = 0x200
)