# NEX related libraries to make NEX servers in Go

[![GoDoc](https://godoc.org/github.com/PretendoNetwork/nex-go?status.svg)](https://godoc.org/github.com/PretendoNetwork/nex-go)

## How to install and use

### Install

`go get https://github.com/PretendoNetwork/nex-go`

### Usage

```Golang
package main

import (
	"fmt"

	NEX "https://github.com/PretendoNetwork/nex-go"
)

// Bare-bones example of the structure of the Friends service AUTH server for the WiiU/3DS

/*
	THIS IS VERY ROUGH EXAMPLE, USING TECHNIQUES THAT SHOULD NOT BE USED
	NO REAL AUTHENTICATION OR CHECKS ARE GOING ON, THIS EXAMPLE JUST
	BRUTE-FORCES THE USER THROUGH AUTHENTICATION.
*/

func main() {

	settings := NEX.NewSettings()

	settings.PrudpVersion = 0
	settings.PrudpV0SignatureVersion = 1
	settings.KerberosKeySize = 16
	settings.AccessKey = "ridfebb9"

	fmt.Println("STARTING FRIENDS AUTH SERVER")

	Server := NEX.NewServer(settings)

	Server.On("Packet", func(Client *NEX.Client, Packet *NEX.Packet) {
		// Acknowledge every packet if it needs it. If it makes it this far, it's good
		if Packet.HasFlag(NEX.Flags["NeedAck"]) {
			Server.Acknowledge(Packet)
		}
	})

	Server.On("Data", func(Client *NEX.Client, Packet *NEX.Packet) {
		stream := NEX.NewInputStream(Packet.RMCRequest.Parameters)
		response := NEX.NewRMCResponse(0x0A, Packet.RMCRequest.Header.CallID)
		responseStream := NEX.NewOutputStream()

		ResponsePacket := NEX.NewPacket(Client)

		ResponsePacket.SetVersion(0)
		ResponsePacket.SetSource(Packet.Destination)
		ResponsePacket.SetDestination(Packet.Source)
		ResponsePacket.SetType(NEX.Types["Data"])

		pid := 1234567890          // account PID/NEX username (dummy)
		password := "nex_password" // account NEX password (dummy)

		key := []byte(password)

		for i := 0; i < 65000+pid%1024; i++ {
			key = NEX.MD5Hash(key)
		}

		Kerberos := NEX.NewKerberos(string(key))

		// Checking the Method ID in this way is generally bad.
		// A better way would be to track each protocol and have a reference to it's methods stored by their ID,
		// and dynamically getting the handler
		if Packet.RMCRequest.Header.MethodID == 2 { // 0x0A::0x02 (LoginEx)
			_ = stream.String()     // username
			_ = stream.DataHolder() // login data

			Kerberos := NEX.NewKerberos(string(key))
			str := "test string test" // dummy data

			kerberosData := Kerberos.Encrypt([]byte(str))

			JSONBuffer := []byte(`{
				"stream":  "10",
				"type": "2",
				"PID": "2",
				"port": "60001",
				"address": "127.0.0.1",
				"sid": "1",
				"CID": "1"
			}`)

			var JSON map[string]string
			_ = json.Unmarshal(JSONBuffer, &JSON)

			station := NEX.NewStationURL("prudps", JSON)

			name := "branch:origin/feature/45925_FixAutoReconnect build:3_10_11_2006_0" // official server

			responseStream.UInt32LE(uint32(0x00010001)) // success
			responseStream.UInt32LE(uint32(pid))
			responseStream.Buffer(kerberosData)

			responseStream.String(station)     // Station URL (normal)
			responseStream.UInt32LE(uint32(0)) // Special protocols (unused)
			responseStream.String("")          // Station URL (special) (unused)

			responseStream.String(name)

			response.SetSuccess(uint32(2), responseStream.Bytes())

			ResponsePacket.SetPayload(response.Bytes())
		} else if Packet.RMCRequest.Header.MethodID == 3 {  // 0x0A::0x03 (RequestTicket)
			str := "test string test test string tes" // dummy data

			_ = stream.UInt32LE() // User PID
			_ = stream.UInt32LE() // Server PID?

			responseStream.UInt32LE(uint32(0x00010001)) // success

			kerberosStream := NEX.NewOutputStream()

			kerberosStream.Write([]byte("I like chickens.")) // Kerberos key (dummy)
			kerberosStream.UInt32LE(uint32(0xFFFFFFFF))      // Unknown PID
			kerberosStream.Buffer([]byte(str))               // Kerberos data

			kerberosData := Kerberos.Encrypt(kerberosStream.Bytes())

			responseStream.Buffer(kerberosData)

			response.SetSuccess(uint32(3), responseStream.Bytes())

		} else {
			panic("INVALID PROTOCOL METHOD ID " + string(Packet.RMCRequest.Header.MethodID))
		}

		ResponsePacket.SetPayload(response.Bytes())
		Server.Send(Client, &ResponsePacket)
	})

	Server.Listen(":60000")
}
```