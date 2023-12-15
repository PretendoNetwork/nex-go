package main

import (
	"fmt"

	"github.com/PretendoNetwork/nex-go"
)

var hppServer *nex.HPPServer

// * Took these structs out of the protocols lib for convenience

type dataStoreGetNotificationURLParam struct {
	nex.Structure
	PreviousURL string
}

func (d *dataStoreGetNotificationURLParam) ExtractFromStream(stream *nex.StreamIn) error {
	var err error

	d.PreviousURL, err = stream.ReadString()
	if err != nil {
		return fmt.Errorf("Failed to extract DataStoreGetNotificationURLParam.PreviousURL. %s", err.Error())
	}

	return nil
}

type dataStoreReqGetNotificationURLInfo struct {
	nex.Structure
	URL        string
	Key        string
	Query      string
	RootCACert []byte
}

func (d *dataStoreReqGetNotificationURLInfo) Bytes(stream *nex.StreamOut) []byte {
	stream.WriteString(d.URL)
	stream.WriteString(d.Key)
	stream.WriteString(d.Query)
	stream.WriteBuffer(d.RootCACert)

	return stream.Bytes()
}

func passwordFromPID(pid *nex.PID) (string, uint32) {
	return "notmypassword", 0
}

func startHPPServer() {
	fmt.Println("Starting HPP")

	hppServer = nex.NewHPPServer()

	hppServer.OnData(func(packet nex.PacketInterface) {
		if packet, ok := packet.(*nex.HPPPacket); ok {
			request := packet.RMCMessage()

			fmt.Println("[HPP]", request.ProtocolID, request.MethodID)

			if request.ProtocolID == 0x73 { // * DataStore
				if request.MethodID == 0xD {
					getNotificationURL(packet)
				}
			}
		}
	})

	hppServer.SetDefaultLibraryVersion(nex.NewLibraryVersion(2, 4, 1))
	hppServer.SetAccessKey("76f26496")
	hppServer.SetPasswordFromPIDFunction(passwordFromPID)

	hppServer.Listen(8085)
}

func getNotificationURL(packet *nex.HPPPacket) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(hppServer)

	parameters := request.Parameters

	parametersStream := nex.NewStreamIn(parameters, hppServer)

	param, err := nex.StreamReadStructure(parametersStream, &dataStoreGetNotificationURLParam{})
	if err != nil {
		fmt.Println("[HPP]", err)
		return
	}

	fmt.Println("[HPP]", param.PreviousURL)

	responseStream := nex.NewStreamOut(hppServer)

	info := &dataStoreReqGetNotificationURLInfo{}
	info.URL = "https://example.com"
	info.Key = "whatever/key"
	info.Query = "?pretendo=1"

	responseStream.WriteStructure(info)

	response.IsSuccess = true
	response.IsRequest = false
	response.ErrorCode = 0x00010001
	response.ProtocolID = request.ProtocolID
	response.CallID = request.CallID
	response.MethodID = request.MethodID
	response.Parameters = responseStream.Bytes()

	// * We replace the RMC message so that it can be delivered back
	packet.SetRMCMessage(response)

	hppServer.Send(packet)
}
