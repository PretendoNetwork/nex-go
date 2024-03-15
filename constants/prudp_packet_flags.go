package constants

const (
	// PacketFlagAck is the ID for the PRUDP Ack Flag
	PacketFlagAck uint16 = 0x1

	// PacketFlagReliable is the ID for the PRUDP Reliable Flag
	PacketFlagReliable uint16 = 0x2

	// PacketFlagNeedsAck is the ID for the PRUDP NeedsAck Flag
	PacketFlagNeedsAck uint16 = 0x4

	// PacketFlagHasSize is the ID for the PRUDP HasSize Flag
	PacketFlagHasSize uint16 = 0x8

	// PacketFlagMultiAck is the ID for the PRUDP MultiAck Flag
	PacketFlagMultiAck uint16 = 0x200
)
