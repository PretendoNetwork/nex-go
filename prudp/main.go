package PRUDPLib

import (
	"encoding/hex"
	"fmt"

	//Prod imports
	V0 "github.com/PretendoNetwork/nex-go/prudp/V0"
	//V1 "github.com/PretendoNetwork/nex-go/prudp/V1"
	//Lite "github.com/PretendoNetwork/nex-go/prudp/Lite"
	General "github.com/PretendoNetwork/nex-go/prudp/General"

	/*//Development imports
	V0 "./V0"
	//V1 "./V1"
	//Lite "./Lite"
	General "./General"*/
)

var debugLog bool = false

func NewPacket () General.Packet {
	return General.Packet{}
}

func FromBytes(data []byte) (General.Packet, int) {
	//detect type of packet

	pack := NewPacket()

	if hex.EncodeToString(data[0:2]) == "ead0" { //V1
		//fmt.Println("PRUDPV1 detected")
		return pack, 1
	} else if hex.EncodeToString(data[0:1]) == "80" { //Lite
		//fmt.Println("PRUDPLite detected")
		return pack, 1
	} else if string(hex.EncodeToString(data[0:1])[0]) == "a" { //V0
		//fmt.Println("PRUDPV0 detected")
		pack = V0.NewPacket(data)
		if debugLog {
			printData(pack, data)
		}
		return pack, 0
	} else { //ERR
		//fmt.Println("Error occurred, packet invalid.")
		return pack, 1
	}
}

func Bytes(data General.Packet) ([]byte, int) {
	//pick type
	switch data.Version {
	case "0":
		ret := V0.BuildPacket(data)
		if debugLog {
			printData(data, ret)
		}
		return ret, 0
	case "1":
		return []byte{0, 0, 0, 0}, 1
	case "Lite":
		return []byte{0, 0, 0, 0}, 1
	default:
		return []byte{0, 0, 0, 0}, 1
	}
}

func printData(packet General.Packet, data []byte) {
	flagstr := ""
	if packet.Flags.ACK {
		flagstr = flagstr + "ACK "
	}
	if packet.Flags.RELIABLE {
		flagstr = flagstr + "REL "
	}
	if packet.Flags.NEED_ACK {
		flagstr = flagstr + "NACK "
	}
	if packet.Flags.HAS_SIZE {
		flagstr = flagstr + "SIZ "
	}
	if packet.Flags.MULTI_ACK {
		flagstr = flagstr + "MACK "
	}
	typestr := ""
	switch packet.Type {
	case 0:
		typestr = "SYN"
	case 1:
		typestr = "CONNECT"
	case 2:
		typestr = "DATA"
	case 3:
		typestr = "DISCONNECT"
	case 4:
		typestr = "PING"
	}
	fmt.Println(" " + typestr + " | " + flagstr + "| Version " + packet.Version + " | " + hex.EncodeToString(data))
}

func ToggleDebugLogging () {
	if debugLog == true {
		fmt.Println("Debug logging is now disabled.")
		debugLog = false
	}
	if debugLog == false {
		fmt.Println("Debug logging is now enabled.")
		debugLog = true
	}
}