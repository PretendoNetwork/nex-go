package V1

import (
	"encoding/hex"

	//Prod imports
	General "github.com/PretendoNetwork/prudplib/General"
	//Development imports
	//General "../General"
)

//There's probably better ways to do this but meh
func NewPacket(data []byte) General.Packet {

	p := General.Packet{}

	//initialization
	flags := General.FLAGInts{}
	flags = flags.SetupFlags()
	p.Version = "1"

	//packet header - payload size & virtual ports
	p.Payload_size = hex.EncodeToString(data[2:4])
	p.Source.Type = string(hex.EncodeToString(data[4:5])[0])
	p.Source.Id = string(hex.EncodeToString(data[4:5])[1])
	p.Destination.Type = string(hex.EncodeToString(data[5:6])[0])
	p.Destination.Id = string(hex.EncodeToString(data[5:6])[1])

	//types and flags
	p.Type = string(hex.EncodeToString(data[6:7])[1])
	p.Flags = hex.EncodeToString(data[7:8]) + string(hex.EncodeToString(data[6:7])[0])

	//session_id, multi-ack version, & sequence_id
	p.Session_id = hex.EncodeToString(data[8:9])
	p.Multi_Ack_Ver = hex.EncodeToString(data[9:0xA])
	p.Sequence_id = hex.EncodeToString(data[0xA:0xC])

	//packet_sig, packet-specific data, and payload
	p.Packet_sig = hex.EncodeToString(data[0xC:0x1C])
	option_id := int(data[0x1C])
	option_size := int(data[0x1D])
	switch option_id {
	case 0:
		//supported function (purpose?)
	case 1:
		p.Connection_sig = hex.EncodeToString(data[0x1E : 0x1E+option_size])
	case 2:
		p.Fragment_id = hex.EncodeToString(data[0x1E : 0x1E+option_size])
	case 3:
		//unknown (random int)
	case 4:
		//unknown (0)
	default:
		//error
	}
	p.Payload = hex.EncodeToString(data[0x1E+option_size : len(data)-1])

	return p
}
