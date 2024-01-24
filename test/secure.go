package main

import (
	"fmt"
	"net"
	"strconv"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-go/types"
)

var secureServer *nex.PRUDPServer

// * Took these structs out of the protocols lib for convenience

type principalPreference struct {
	types.Structure
	*types.Data
	ShowOnlinePresence  *types.PrimitiveBool
	ShowCurrentTitle    *types.PrimitiveBool
	BlockFriendRequests *types.PrimitiveBool
}

func (pp *principalPreference) WriteTo(writable types.Writable) {
	pp.ShowOnlinePresence.WriteTo(writable)
	pp.ShowCurrentTitle.WriteTo(writable)
	pp.BlockFriendRequests.WriteTo(writable)
}

type comment struct {
	types.Structure
	*types.Data
	Unknown     *types.PrimitiveU8
	Contents    *types.String
	LastChanged *types.DateTime
}

func (c *comment) WriteTo(writable types.Writable) {
	c.Unknown.WriteTo(writable)
	c.Contents.WriteTo(writable)
	c.LastChanged.WriteTo(writable)
}

func startSecureServer() {
	fmt.Println("Starting secure")

	secureServer = nex.NewPRUDPServer()

	endpoint := nex.NewPRUDPEndPoint(1)

	endpoint.AccountDetailsByPID = accountDetailsByPID
	endpoint.AccountDetailsByUsername = accountDetailsByUsername
	endpoint.ServerAccount = secureServerAccount

	endpoint.OnData(func(packet nex.PacketInterface) {
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

	secureServer.SetFragmentSize(962)
	secureServer.SetDefaultLibraryVersion(nex.NewLibraryVersion(1, 1, 0))
	secureServer.SessionKeyLength = 16
	secureServer.SetAccessKey("ridfebb9")
	secureServer.BindPRUDPEndPoint(endpoint)
	secureServer.Listen(60001)
}

func registerEx(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(secureServer)
	connection := packet.Sender().(*nex.PRUDPConnection)

	parameters := request.Parameters

	parametersStream := nex.NewByteStreamIn(parameters, secureServer)

	vecMyURLs := types.NewList[*types.StationURL]()
	vecMyURLs.Type = types.NewStationURL("")
	if err := vecMyURLs.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	hCustomData := types.NewAnyDataHolder()
	if err := hCustomData.ExtractFrom(parametersStream); err != nil {
		fmt.Println(err)
	}

	localStation, _ := vecMyURLs.Get(0)

	address := packet.Sender().Address().(*net.UDPAddr).IP.String()

	localStation.Fields["address"] = address
	localStation.Fields["port"] = strconv.Itoa(packet.Sender().Address().(*net.UDPAddr).Port)

	retval := types.NewQResultSuccess(0x00010001)
	localStationURL := types.NewString(localStation.EncodeToString())

	responseStream := nex.NewByteStreamOut(secureServer)

	retval.WriteTo(responseStream)
	responseStream.WritePrimitiveUInt32LE(connection.ID)
	localStationURL.WriteTo(responseStream)

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(connection, nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	responsePacket.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	responsePacket.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	responsePacket.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}

func updateAndGetAllInformation(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(secureServer)

	responseStream := nex.NewByteStreamOut(secureServer)

	(&principalPreference{
		ShowOnlinePresence:  types.NewPrimitiveBool(true),
		ShowCurrentTitle:    types.NewPrimitiveBool(true),
		BlockFriendRequests: types.NewPrimitiveBool(false),
	}).WriteTo(responseStream)
	(&comment{
		Unknown:     types.NewPrimitiveU8(0),
		Contents:    types.NewString("Rewrite Test"),
		LastChanged: types.NewDateTime(0),
	}).WriteTo(responseStream)
	responseStream.WritePrimitiveUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(friendList)
	responseStream.WritePrimitiveUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(friendRequestsOut)
	responseStream.WritePrimitiveUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(friendRequestsIn)
	responseStream.WritePrimitiveUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(blockList)
	responseStream.WritePrimitiveBool(false) // * Unknown
	responseStream.WritePrimitiveUInt32LE(0) // * Stubbed empty list. responseStream.WriteListStructure(notifications)
	responseStream.WritePrimitiveBool(false) // * Unknown

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(packet.Sender().(*nex.PRUDPConnection), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	responsePacket.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	responsePacket.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	responsePacket.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}

func checkSettingStatus(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(secureServer)

	responseStream := nex.NewByteStreamOut(secureServer)

	responseStream.WritePrimitiveUInt8(0) // * Unknown

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(packet.Sender().(*nex.PRUDPConnection), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	responsePacket.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	responsePacket.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	responsePacket.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}

func updatePresence(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(secureServer)

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID

	responsePacket, _ := nex.NewPRUDPPacketV0(packet.Sender().(*nex.PRUDPConnection), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(nex.FlagHasSize)
	responsePacket.AddFlag(nex.FlagReliable)
	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	responsePacket.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	responsePacket.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	responsePacket.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	secureServer.Send(responsePacket)
}
