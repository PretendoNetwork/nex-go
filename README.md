# Barebones PRUDP/NEX server library written in Go

[![GoDoc](https://godoc.org/github.com/PretendoNetwork/nex-go?status.svg)](https://godoc.org/github.com/PretendoNetwork/nex-go)

### Install

`go get github.com/PretendoNetwork/nex-go`

### Usage note

This module provides a barebones PRUDP server for use with titles using the Nintendo NEX library. It does not provide any support for titles using the original Rendez-Vous library developed by Quazal. This library only provides the low level packet data, as such it is recommended to use [NEX Protocols Go](https://github.com/PretendoNetwork/nex-protocols-go) to develop servers.

### Usage

```Golang
package main

import (
	"fmt"

	nex "github.com/PretendoNetwork/nex-go"
)

func main() {
	nexServer := nex.NewServer()
	nexServer.SetPrudpVersion(0)
	nexServer.SetSignatureVersion(1)
	nexServer.SetKerberosKeySize(16)
	nexServer.SetAccessKey("ridfebb9")

	nexServer.On("Data", func(packet *nex.PacketV0) {
		request := packet.RMCRequest()

		fmt.Println("==Friends - Auth==")
		fmt.Printf("Protocol ID: %#v\n", request.ProtocolID())
		fmt.Printf("Method ID: %#v\n", request.MethodID())
		fmt.Println("==================")
	})

	nexServer.Listen(":60000")
}
```