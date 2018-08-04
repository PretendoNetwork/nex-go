package General

type Packet struct {
	Version        string
	Source         VPort
	Destination    VPort
	Type           int
	Flags          FLAGS
	SessionId     uint8
	PacketSig     uint32
	SequenceId    uint16
	ConnectionSig uint32
	FragmentId    uint8
	PayloadSize   uint16
	MultiAckVer  int
	Payload        []byte //hex
	Checksum       uint8
}

const (
	PRUDPV0   string = "0"
	PRUDPV1   string = "1"
	PRUDPLite string = "Lite"
)

type VPort struct {
	Type string
	Id   string
}

type FLAGS struct {
	ACK       bool
	RELIABLE  bool
	NEED_ACK  bool
	HAS_SIZE  bool
	MULTI_ACK bool
}

type FLAGInts struct {
	ACK       int
	RELIABLE  int
	NEED_ACK  int
	HAS_SIZE  int
	MULTI_ACK int
}

func (f FLAGInts) SetupFlags() FLAGInts {
	f.ACK = 0x001
	f.RELIABLE = 0x002
	f.NEED_ACK = 0x004
	f.HAS_SIZE = 0x008
	f.MULTI_ACK = 0x200
	return f
}
