package nex

const (
	FlagAck      uint16 = 0x1
	FlagReliable uint16 = 0x2
	FlagNeedsAck uint16 = 0x4
	FlagHasSize  uint16 = 0x8
	FlagMultiAck uint16 = 0x200
)
