package main

import (
	"fmt"

	"github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-go/types"
)

var hppServer *nex.HPPServer

// * Took these structs out of the protocols lib for convenience

type dataStoreGetNotificationURLParam struct {
	types.Structure
	PreviousURL *types.String
}

func (d *dataStoreGetNotificationURLParam) ExtractFrom(readable types.Readable) error {
	var err error

	if err = d.ExtractHeaderFrom(readable); err != nil {
		return fmt.Errorf("Failed to extract DataStoreGetNotificationURLParam header. %s", err.Error())
	}

	err = d.PreviousURL.ExtractFrom(readable)
	if err != nil {
		return fmt.Errorf("Failed to extract DataStoreGetNotificationURLParam.PreviousURL. %s", err.Error())
	}

	return nil
}

type dataStoreReqGetNotificationURLInfo struct {
	types.Structure
	URL        *types.String
	Key        *types.String
	Query      *types.String
	RootCACert *types.Buffer
}

func (d *dataStoreReqGetNotificationURLInfo) WriteTo(writable types.Writable) {
	contentWritable := writable.CopyNew()

	d.URL.WriteTo(contentWritable)
	d.Key.WriteTo(contentWritable)
	d.Query.WriteTo(contentWritable)
	d.RootCACert.WriteTo(contentWritable)

	content := contentWritable.Bytes()

	d.WriteHeaderTo(writable, uint32(len(content)))

	writable.Write(content)
}

func passwordFromPID(pid *types.PID) (string, uint32) {
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

	hppServer.Listen(12345)
}

func getNotificationURL(packet *nex.HPPPacket) {
	request := packet.RMCMessage()
	response := nex.NewRMCMessage(hppServer)

	parameters := request.Parameters

	parametersStream := nex.NewByteStreamIn(parameters, hppServer)

	param := &dataStoreGetNotificationURLParam{}
	param.PreviousURL = types.NewString("")
	err := param.ExtractFrom(parametersStream)
	if err != nil {
		fmt.Println("[HPP]", err)
		return
	}

	fmt.Println("[HPP]", param.PreviousURL)

	responseStream := nex.NewByteStreamOut(hppServer)

	info := &dataStoreReqGetNotificationURLInfo{}
	info.URL = types.NewString("https://example.com")
	info.Key = types.NewString("whatever/key")
	info.Query = types.NewString("?pretendo=1")
	info.RootCACert = types.NewBuffer(nil)

	info.WriteTo(responseStream)

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
