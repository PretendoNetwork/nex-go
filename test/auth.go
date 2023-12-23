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

	authServer.OnData(func(packet nex.PacketInterface) {
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
	//authServer.PRUDPVersion = 1
	authServer.SetDefaultLibraryVersion(nex.NewLibraryVersion(1, 1, 0))
	authServer.SetKerberosPassword([]byte("password"))
	authServer.SetKerberosKeySize(16)
	authServer.SetAccessKey("ridfebb9")
	authServer.Listen(60000)
}

func login(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(authServer)

	parameters := request.Parameters

	parametersStream := nex.NewStreamIn(parameters, authServer)

	strUserName := types.NewString()
	if err := strUserName.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	converted, err := strconv.Atoi(string(*strUserName))
	if err != nil {
		panic(err)
	}

	retval := types.NewResultSuccess(0x00010001)
	pidPrincipal := types.NewPID(uint32(converted))
	pbufResponse := types.Buffer(generateTicket(pidPrincipal, types.NewPID[uint32](2)))
	pConnectionData := types.NewRVConnectionData()
	strReturnMsg := types.String("Test Build")

	pConnectionData.StationURL = types.NewStationURL("prudps:/address=192.168.1.98;port=60001;CID=1;PID=2;sid=1;stream=10;type=2")
	pConnectionData.SpecialProtocols = types.NewList[*types.PrimitiveU8]()
	pConnectionData.SpecialProtocols.Type = types.NewPrimitiveU8()
	pConnectionData.StationURLSpecialProtocols = types.NewStationURL("")
	pConnectionData.Time = types.NewDateTime(0).Now()

	responseStream := nex.NewStreamOut(authServer)

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

	authServer.Send(responsePacket)
}

func requestTicket(packet nex.PRUDPPacketInterface) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(authServer)

	parameters := request.Parameters

	parametersStream := nex.NewStreamIn(parameters, authServer)

	idSource := types.NewPID[uint64](0)
	if err := idSource.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	idTarget := types.NewPID[uint64](0)
	if err := idTarget.ExtractFrom(parametersStream); err != nil {
		panic(err)
	}

	retval := types.NewResultSuccess(0x00010001)
	pbufResponse := types.Buffer(generateTicket(idSource, idTarget))

	responseStream := nex.NewStreamOut(authServer)

	retval.WriteTo(responseStream)
	pbufResponse.WriteTo(responseStream)

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

	authServer.Send(responsePacket)
}
