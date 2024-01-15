// Package main implements a test server
package main

import (
	"fmt"
	"strconv"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-go/types"
)

var authServer *nex.PRUDPServer

func startAuthenticationServer() {
	fmt.Println("Starting auth")

	authServer = nex.NewPRUDPServer()

	endpoint := nex.NewPRUDPEndPoint(1)

	endpoint.OnData(func(packet nex.PacketInterface) {
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
	authServer.SetDefaultLibraryVersion(nex.NewLibraryVersion(1, 1, 0))
	authServer.SetKerberosPassword([]byte("password"))
	authServer.SetKerberosKeySize(16)
	authServer.SetAccessKey("ridfebb9")
	authServer.BindPRUDPEndPoint(endpoint)
	authServer.Listen(60000)
}

func login(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(authServer)

	parameters := request.Parameters

	parametersStream := nex.NewByteStreamIn(parameters, authServer)

	strUserName := types.NewString("")
	if err := strUserName.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	converted, err := strconv.Atoi(strUserName.Value)
	if err != nil {
		panic(err)
	}

	retval := types.NewResultSuccess(0x00010001)
	pidPrincipal := types.NewPID(uint64(converted))
	pbufResponse := types.NewBuffer(generateTicket(pidPrincipal, types.NewPID(2)))
	pConnectionData := types.NewRVConnectionData()
	strReturnMsg := types.NewString("Test Build")

	pConnectionData.StationURL = types.NewStationURL("prudps:/address=192.168.1.98;port=60001;CID=1;PID=2;sid=1;stream=10;type=2")
	pConnectionData.SpecialProtocols = types.NewList[*types.PrimitiveU8]()
	pConnectionData.SpecialProtocols.Type = types.NewPrimitiveU8(0)
	pConnectionData.StationURLSpecialProtocols = types.NewStationURL("")
	pConnectionData.Time = types.NewDateTime(0).Now()

	responseStream := nex.NewByteStreamOut(authServer)

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

	authServer.Send(responsePacket)
}

func requestTicket(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(authServer)

	parameters := request.Parameters

	parametersStream := nex.NewByteStreamIn(parameters, authServer)

	idSource := types.NewPID(0)
	if err := idSource.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	idTarget := types.NewPID(0)
	if err := idTarget.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	retval := types.NewResultSuccess(0x00010001)
	pbufResponse := types.NewBuffer(generateTicket(idSource, idTarget))

	responseStream := nex.NewByteStreamOut(authServer)

	retval.WriteTo(responseStream)
	pbufResponse.WriteTo(responseStream)

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

	authServer.Send(responsePacket)
}
