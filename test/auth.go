// Package main implements a test server
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
)

var authServer *nex.PRUDPServer
var authEndpoint *nex.PRUDPEndPoint

func startAuthenticationServer() {
	fmt.Println("Starting auth")

	authServer = nex.NewPRUDPServer()

	authEndpoint = nex.NewPRUDPEndPoint(1)

	authServer.EnableMetrics("0.0.0.0:9090")

	authEndpoint.AccountDetailsByPID = accountDetailsByPID
	authEndpoint.AccountDetailsByUsername = accountDetailsByUsername
	authEndpoint.ServerAccount = authenticationServerAccount

	authEndpoint.OnData(func(packet nex.PacketInterface) {
		if packet, ok := packet.(nex.PRUDPPacketInterface); ok {
			request := packet.RMCMessage()

			fmt.Println("[AUTH]", request.ProtocolID, request.MethodID)

			if request.ProtocolID == 0xA { // * Ticket Granting
				if request.MethodID == 0x1 {
					login(packet)
				}

				if request.MethodID == 0x3 {
					requestTicket(packet)
				}
			}
		}
	})

	authServer.SetFragmentSize(962)
	authServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(1, 1, 0))
	authServer.SessionKeyLength = 16
	authServer.AccessKey = "ridfebb9"
	authServer.BindPRUDPEndPoint(authEndpoint)
	authServer.Listen(60000)
}

func login(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(authEndpoint)

	parameters := request.Parameters

	parametersStream := nex.NewByteStreamIn(parameters, authEndpoint.LibraryVersions(), authEndpoint.ByteStreamSettings())

	strUserName := types.NewString("")
	if err := strUserName.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	sourceAccount, _ := accountDetailsByUsername(string(strUserName))
	targetAccount, _ := accountDetailsByUsername(secureServerAccount.Username)

	retval := types.NewQResultSuccess(0x00010001)
	pidPrincipal := sourceAccount.PID
	pbufResponse := types.NewBuffer(generateTicket(sourceAccount, targetAccount, authServer.SessionKeyLength))
	pConnectionData := types.NewRVConnectionData()
	strReturnMsg := types.NewString("Test Build")

	pConnectionData.StationURL = types.NewStationURL("prudps:/address=192.168.1.98;port=60001;CID=1;PID=2;sid=1;stream=10;type=2")
	pConnectionData.SpecialProtocols = types.NewList[types.UInt8]()
	pConnectionData.StationURLSpecialProtocols = types.NewStationURL("")
	pConnectionData.Time = types.NewDateTime(0).Now()

	responseStream := nex.NewByteStreamOut(authEndpoint.LibraryVersions(), authEndpoint.ByteStreamSettings())

	retval.WriteTo(responseStream)
	pidPrincipal.WriteTo(responseStream)
	pbufResponse.WriteTo(responseStream)
	pConnectionData.WriteTo(responseStream)
	strReturnMsg.WriteTo(responseStream)

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(authServer, packet.Sender().(*nex.PRUDPConnection), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(constants.PacketFlagHasSize)
	responsePacket.AddFlag(constants.PacketFlagReliable)
	responsePacket.AddFlag(constants.PacketFlagNeedsAck)
	responsePacket.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	responsePacket.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	responsePacket.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	responsePacket.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	fmt.Println(hex.EncodeToString(responsePacket.Payload()))

	authServer.Send(responsePacket)
}

func requestTicket(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(authEndpoint)

	parameters := request.Parameters

	parametersStream := nex.NewByteStreamIn(parameters, authEndpoint.LibraryVersions(), authEndpoint.ByteStreamSettings())

	idSource := types.NewPID(0)
	if err := idSource.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	idTarget := types.NewPID(0)
	if err := idTarget.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	sourceAccount, _ := accountDetailsByPID(idSource)
	targetAccount, _ := accountDetailsByPID(idTarget)

	retval := types.NewQResultSuccess(0x00010001)
	pbufResponse := types.NewBuffer(generateTicket(sourceAccount, targetAccount, authServer.SessionKeyLength))

	responseStream := nex.NewByteStreamOut(authEndpoint.LibraryVersions(), authEndpoint.ByteStreamSettings())

	retval.WriteTo(responseStream)
	pbufResponse.WriteTo(responseStream)

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	responsePacket, _ := nex.NewPRUDPPacketV0(authServer, packet.Sender().(*nex.PRUDPConnection), nil)

	responsePacket.SetType(packet.Type())
	responsePacket.AddFlag(constants.PacketFlagHasSize)
	responsePacket.AddFlag(constants.PacketFlagReliable)
	responsePacket.AddFlag(constants.PacketFlagNeedsAck)
	responsePacket.SetSourceVirtualPortStreamType(packet.DestinationVirtualPortStreamType())
	responsePacket.SetSourceVirtualPortStreamID(packet.DestinationVirtualPortStreamID())
	responsePacket.SetDestinationVirtualPortStreamType(packet.SourceVirtualPortStreamType())
	responsePacket.SetDestinationVirtualPortStreamID(packet.SourceVirtualPortStreamID())
	responsePacket.SetSubstreamID(packet.SubstreamID())
	responsePacket.SetPayload(response.Bytes())

	fmt.Println(hex.EncodeToString(responsePacket.Payload()))

	authServer.Send(responsePacket)
}
