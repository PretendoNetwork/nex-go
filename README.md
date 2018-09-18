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

func main() {

	// Bare-bones example of the structure of the Friends service AUTH server for the WiiU/3DS
	Server := NEX.NewServer(NEX.Settings{
		PrudpVersion:            0,
		PrudpV0SignatureVersion: 1,
		PrudpV0FlagsVersion:     1,
		PrudpV0ChecksumVersion:  1,
		AccessKey:               "ridfebb9",
	})

	Server.On("Packet", func(Client *NEX.Client, Packet *NEX.Packet) {
		fmt.Println("Packet event")
	})

	Server.On("Syn", func(Client *NEX.Client, Packet *NEX.Packet) {
		if Packet.HasFlag(NEX.Flags["NeedAck"]) {
			Server.Acknowledge(Packet)
		}
	})

	Server.On("Connect", func(Client *NEX.Client, Packet *NEX.Packet) {
		if Packet.HasFlag(NEX.Flags["NeedAck"]) {
			Server.Acknowledge(Packet)
		}
	})

	Server.On("Data", func(Client *NEX.Client, Packet *NEX.Packet) {
		response := NEX.NewRMCResponse(0x0A, uint32(Packet.SequenceID))
		response.SetError(uint32(0x8068000B))

		ResponsePacket := NEX.NewPacket(Client)

		ResponsePacket.SetVersion(0)
		ResponsePacket.SetSource(Packet.Destination)
		ResponsePacket.SetDestination(Packet.Source)
		ResponsePacket.SetType(NEX.Types["Data"])
		ResponsePacket.SetPayload(response.Bytes())

		Server.Send(Client, &ResponsePacket)
	})

	Server.On("Disconnect", func(Client *NEX.Client, Packet *NEX.Packet) {
		fmt.Println("Disconnect event")
	})

	Server.On("Ping", func(Client *NEX.Client, Packet *NEX.Packet) {
		fmt.Println("Ping event")
	})

	Server.Listen(":60000")
}
```