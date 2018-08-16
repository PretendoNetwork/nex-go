package prudp

var Types = make(map[string]int, 5)
var Flags = make(map[string]int, 5)

func init() {
	Types["Syn"] = 0
	Types["Connect"] = 1
	Types["Data"] = 2
	Types["Disconnect"] = 3
	Types["Ping"] = 4

	Flags["Ack"] = 0x001
	Flags["Reliable"] = 0x002
	Flags["NeedAck"] = 0x004
	Flags["HasSize"] = 0x008
	Flags["MultiAck"] = 0x200
}
