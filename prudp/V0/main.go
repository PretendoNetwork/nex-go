package V0

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"

	//Prod imports
	General "github.com/PretendoNetwork/nex-go/prudp/General"
	Checksum "github.com/PretendoNetwork/nex-go/prudp/V0/checksum"
	Convert "github.com/PretendoNetwork/nex-go/prudp/V0/simple_converter"

	/*//Development imports
	General "../General"
	Checksum "./checksum"
	Convert "./simple_converter"*/
)

//There's probably better ways to do this but meh
func NewPacket(data []byte) General.Packet {

	p := General.Packet{}

	//initialization
	flags := General.FLAGInts{}
	flags = flags.SetupFlags()
	p.Version = General.PRUDPV0

	//virtual ports
	p.Source.Type = string(Convert.ByteSliceToHexString(data[0:1])[0])
	p.Source.Id = string(Convert.ByteSliceToHexString(data[0:1])[1])
	p.Destination.Type = string(Convert.ByteSliceToHexString(data[1:2])[0])
	p.Destination.Id = string(Convert.ByteSliceToHexString(data[1:2])[1])

	//types and flags
	_type, err := strconv.ParseInt(hex.EncodeToString(data[2:3])[1:2], 16, 0)
	if err != nil {
	}
	p.Type = int(_type)
	_flags, errr := strconv.ParseInt(hex.EncodeToString(data[3:4])+string(hex.EncodeToString(data[2:3])[0])[0:], 16, 0)
	if errr != nil {
	}
	_flag := int(_flags)
	if _flag&flags.ACK == flags.ACK {
		p.Flags.ACK = true
	}
	if _flag&flags.RELIABLE == flags.RELIABLE {
		p.Flags.RELIABLE = true
	}
	if _flag&flags.NEED_ACK == flags.NEED_ACK {
		p.Flags.NEED_ACK = true
	}
	if _flag&flags.HAS_SIZE == flags.HAS_SIZE {
		p.Flags.HAS_SIZE = true
	}
	if _flag&flags.MULTI_ACK == flags.MULTI_ACK {
		p.Flags.MULTI_ACK = true
	}

	//session_id, packet_sig, & sequence_id
	p.SessionId = uint8(data[4])
	p.PacketSig = binary.LittleEndian.Uint32(data[5:9])
	p.SequenceId = binary.LittleEndian.Uint16(data[9:0xB])

	//packet-specifics
	pos := 0xB
	if p.Type == 0 || p.Type == 1 { //SYN or CONNECT
		p.ConnectionSig = binary.LittleEndian.Uint32(data[pos : pos+4])
		pos = pos + 4
	}
	if p.Type == 2 { //DATA
		p.FragmentId = uint8(data[pos])
		pos = pos + 1
	}
	if p.Flags.HAS_SIZE {
		p.PayloadSize = binary.LittleEndian.Uint16(data[pos : pos+2])
		pos = pos + 2
	}

	//payload & checksum
	p.Payload = data[pos : len(data)-1]
	p.Checksum = uint8(data[len(data)-1])

	return p
}

func BuildPacket(data General.Packet) []byte {

	packet := make([]byte, 64000)

	//virtual ports, type, and flags
	_vports := data.Source.Type + data.Source.Id + data.Destination.Type + data.Destination.Id
	_type := fmt.Sprintf("%x", data.Type)
	flagdata := 0
	if data.Flags.ACK {
		flagdata = flagdata + 1
	}
	if data.Flags.RELIABLE {
		flagdata = flagdata + 2
	}
	if data.Flags.NEED_ACK {
		flagdata = flagdata + 4
	}
	if data.Flags.HAS_SIZE {
		flagdata = flagdata + 8
	}
	if data.Flags.MULTI_ACK {
		flagdata = flagdata + 512
	}
	_flags := Convert.IntToHexString(flagdata)
	_flag_prepend := "00000"
	_flags = _flag_prepend + _flags
	copy(packet[0:4], Convert.HexStringToByteSlice(_vports + string(_flags[len(_flags)-1]) + _type + _flags[len(_flags)-3:len(_flags)-1]))

	//session_id, packet_sig, & sequence_id
	packet[4] = byte(data.SessionId)
	binary.LittleEndian.PutUint32(packet[5:], data.PacketSig)
	binary.LittleEndian.PutUint16(packet[9:], data.SequenceId)

	pos := 0xB
	//packet-specific data
	if data.Type == 0 || data.Type == 1 {
		binary.LittleEndian.PutUint32(packet[pos:], data.ConnectionSig)
		pos = pos + 4
	}
	if data.Type == 2 {
		packet[pos] = byte(data.FragmentId)
		pos = pos + 1
	}
	if data.Flags.HAS_SIZE {
		binary.LittleEndian.PutUint16(packet[pos:], data.PayloadSize)
		pos = pos + 2
	}

	//payload & checksum
	copy(packet[pos:], data.Payload)
	pos += len(data.Payload)
	packet[pos] = byte(Checksum.CalcChecksum("ridfebb9", packet))

	retpkt := make([]byte, pos + 1)
	copy(retpkt, packet[0:pos + 1])

	return retpkt
}
