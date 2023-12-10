package main

import (
	"fmt"
	"net"
	"strconv"

	"github.com/PretendoNetwork/nex-go"
)

var secureServer *nex.PRUDPServer

// * Took these structs out of the protocols lib for convenience

type principalPreference struct {
	nex.Structure
	*nex.Data
	ShowOnlinePresence  bool
	ShowCurrentTitle    bool
	BlockFriendRequests bool
}

func (pp *principalPreference) Bytes(stream *nex.StreamOut) []byte {
	stream.WriteBool(pp.ShowOnlinePresence)
	stream.WriteBool(pp.ShowCurrentTitle)
	stream.WriteBool(pp.BlockFriendRequests)

	return stream.Bytes()
}

type comment struct {
	nex.Structure
	*nex.Data
	Unknown     uint8
	Contents    string
	LastChanged *nex.DateTime
}

func (c *comment) Bytes(stream *nex.StreamOut) []byte {
	stream.WriteUInt8(c.Unknown)
	stream.WriteString(c.Contents)
	stream.WriteDateTime(c.LastChanged)

	return stream.Bytes()
}

func startSecureServer() {
	fmt.Println("Starting secure")

	secureServer = nex.NewPRUDPServer()

	secureServer.OnData(func(packet nex.PacketInterface) {
		if packet, ok := packet.(nex.PRUDPPacketInterface); ok {
			request := packet.RMCMessage()

			fmt.Println("[SECR]", request.ProtocolID, request.MethodID)

			if request.ProtocolID == 0xB { // * Secure Connection
				if request.MethodID == 0x4 {
					registerEx(packet)
				}
			}

			if request.ProtocolID == 0x66 { // * Friends (WiiU)
				if request.MethodID == 1 {
					updateAndGetAllInformation(packet)
				} else if request.MethodID == 19 {
					checkSettingStatus(packet)
				} else if request.MethodID == 13 {
					updatePresence(packet)
				} else {
					panic(fmt.Sprintf("Unknown method %d", request.MethodID))
				}
			}
		}
	})

	secureServer.SecureVirtualServerPorts = []uint8{1}
	//secureServer.PRUDPVersion = 1
	secureServer.SetFragmentSize(962)
	secureServer.SetDefaultLibraryVersion(nex.NewLibraryVersion(1, 1, 0))
	secureServer.SetKerberosPassword([]byte("password"))
	secureServer.SetKerberosKeySize(16)
	secureServer.SetAccessKey("ridfebb9")
	secureServer.Listen(60001)
}

func registerEx(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage()

	parameters := request.Parameters

	parametersStream := nex.NewStreamIn(parameters, authServer)

	vecMyURLs, err := parametersStream.ReadListStationURL()
	if err != nil {
		panic(err)
	}

	_, err = parametersStream.ReadDataHolder()
	if err != nil {
		fmt.Println(err)
	}

	localStation := vecMyURLs[0]

	address := packet.Sender().Address().(*net.UDPAddr).IP.String()

	localStation.Fields.Set("address", address)
	localStation.Fields.Set("port", strconv.Itoa(packet.Sender().Address().(*net.UDPAddr).Port))

	retval := nex.NewResultSuccess(0x00010001)
	localStationURL := localStation.EncodeToString()

	responseStream := nex.NewStreamOut(authServer)

	responseStream.WriteResult(retval)
	responseStream.WriteUInt32LE(secureServer.ConnectionIDCounter().Next())
	responseStream.WriteString(localStationURL)

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(packet.Sender().(*nex.PRUDPClient), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceStreamType(packet.DestinationStreamType())
	responsePacket.SetSourcePort(packet.DestinationPort())
	responsePacket.SetDestinationStreamType(packet.SourceStreamType())
	responsePacket.SetDestinationPort(packet.SourcePort())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}

func updateAndGetAllInformation(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage()

	responseStream := nex.NewStreamOut(authServer)

	responseStream.WriteStructure(&principalPreference{
		ShowOnlinePresence:  true,
		ShowCurrentTitle:    true,
		BlockFriendRequests: false,
	})
	responseStream.WriteStructure(&comment{
		Unknown:     0,
		Contents:    "Rewrite Test",
		LastChanged: nex.NewDateTime(0),
	})
	responseStream.WriteUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(friendList)
	responseStream.WriteUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(friendRequestsOut)
	responseStream.WriteUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(friendRequestsIn)
	responseStream.WriteUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(blockList)
	responseStream.WriteBool(false) // * Unknown
	responseStream.WriteUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(notifications)
	responseStream.WriteBool(false) // * Unknown

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(packet.Sender().(*nex.PRUDPClient), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceStreamType(packet.DestinationStreamType())
	responsePacket.SetSourcePort(packet.DestinationPort())
	responsePacket.SetDestinationStreamType(packet.SourceStreamType())
	responsePacket.SetDestinationPort(packet.SourcePort())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}

func checkSettingStatus(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage()

	responseStream := nex.NewStreamOut(authServer)

	responseStream.WriteUInt8(0) // * Unknown

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(packet.Sender().(*nex.PRUDPClient), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceStreamType(packet.DestinationStreamType())
	responsePacket.SetSourcePort(packet.DestinationPort())
	responsePacket.SetDestinationStreamType(packet.SourceStreamType())
	responsePacket.SetDestinationPort(packet.SourcePort())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}

func updatePresence(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage()

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID

	responsePacket, _ := nex.NewPRUDPPacketV0(packet.Sender().(*nex.PRUDPClient), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceStreamType(packet.DestinationStreamType())
	responsePacket.SetSourcePort(packet.DestinationPort())
	responsePacket.SetDestinationStreamType(packet.SourceStreamType())
	responsePacket.SetDestinationPort(packet.SourcePort())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}
