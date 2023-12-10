# NEX Go
## Barebones PRUDP/NEX server library written in Go

[![GoDoc](https://godoc.org/github.com/PretendoNetwork/nex-go?status.svg)](https://godoc.org/github.com/PretendoNetwork/nex-go)

### Other NEX libraries
[nex-protocols-go](https://github.com/PretendoNetwork/nex-protocols-go) - NEX protocol definitions

[nex-protocols-common-go](https://github.com/PretendoNetwork/nex-protocols-common-go) - NEX protocols used by many games with premade handlers and a high level API

### Install

`go get github.com/PretendoNetwork/nex-go`

### Usage note

This module provides a barebones PRUDP server for use with titles using the Nintendo NEX library. It also provides support for titles using the original Rendez-Vous library developed by Quazal. This library only provides the low level packet data, as such it is recommended to use [NEX Protocols Go](https://github.com/PretendoNetwork/nex-protocols-go) to develop servers.

### Usage

```go
package main

import (
	"fmt"

	nex "github.com/PretendoNetwork/nex-go"
)

func main() {
	nexServer := nex.NewPRUDPServer()
	nexServer.PRUDPVersion = 0
	nexServer.SetFragmentSize(962)
	nexServer.SetDefaultLibraryVersion(nex.NewLibraryVersion(1, 1, 0))
	nexServer.SetKerberosPassword([]byte("password"))
	nexServer.SetKerberosKeySize(16)
	nexServer.SetAccessKey("ridfebb9")

	nexServer.OnData(func(packet nex.PacketInterface) {
		request := packet.RMCMessage()

		fmt.Println("==Friends - Auth==")
		fmt.Printf("Protocol ID: %#v\n", request.ProtocolID)
		fmt.Printf("Method ID: %#v\n", request.MethodID)
		fmt.Println("==================")
	})

	nexServer.Listen(60000)
}
```
