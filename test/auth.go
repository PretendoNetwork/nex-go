package main

import (
	"fmt"
	"strconv"

	"github.com/PretendoNetwork/nex-go"
)

var authServer *nex.PRUDPServer

func startAuthenticationServer() {
	fmt.Println("Starting auth")

	authServer = nex.NewPRUDPServer()

	authServer.OnReliableData(func(packet nex.PacketInterface) {
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
	response := nex.NewRMCMessage()

	parameters := request.Parameters

	parametersStream := nex.NewStreamIn(parameters, authServer)

	strUserName, err := parametersStream.ReadString()
	if err != nil {
		panic(err)
	}

	converted, err := strconv.Atoi(strUserName)
	if err != nil {
		panic(err)
	}

	retval := nex.NewResultSuccess(0x00010001)
	pidPrincipal := uint32(converted)
	pbufResponse := generateTicket(pidPrincipal, 2)
	pConnectionData := nex.NewRVConnectionData()
	strReturnMsg := "Test Build"

	pConnectionData.SetStationURL("prudps:/address=192.168.1.98;port=60001;CID=1;PID=2;sid=1;stream=10;type=2")
	pConnectionData.SetSpecialProtocols([]byte{})
	pConnectionData.SetStationURLSpecialProtocols("")
	serverTime := nex.NewDateTime(0)
	pConnectionData.SetTime(nex.NewDateTime(serverTime.UTC()))

	responseStream := nex.NewStreamOut(authServer)

	responseStream.WriteResult(retval)
	responseStream.WriteUInt32LE(pidPrincipal)
	responseStream.WriteBuffer(pbufResponse)
	responseStream.WriteStructure(pConnectionData)
	responseStream.WriteString(strReturnMsg)

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
	response := nex.NewRMCMessage()

	parameters := request.Parameters

	parametersStream := nex.NewStreamIn(parameters, authServer)

	idSource, err := parametersStream.ReadUInt32LE()
	if err != nil {
		panic(err)
	}

	idTarget, err := parametersStream.ReadUInt32LE()
	if err != nil {
		panic(err)
	}

	retval := nex.NewResultSuccess(0x00010001)
	pbufResponse := generateTicket(idSource, idTarget)

	responseStream := nex.NewStreamOut(authServer)

	responseStream.WriteResult(retval)
	responseStream.WriteBuffer(pbufResponse)

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
