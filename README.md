# Barebones PRUDP/NEX server library written in Go

[![GoDoc](https://godoc.org/github.com/PretendoNetwork/nex-go?status.svg)](https://godoc.org/github.com/PretendoNetwork/nex-go)

### Install

`go get github.com/PretendoNetwork/nex-go`

### Usage note

While this package can be used stand-alone, it only provides the bare minimum for a PRUDP/NEX server. It does not support any NEX protocols. To make proper NEX servers, see [NEX Protocols Go](https://github.com/PretendoNetwork/nex-protocols-go)

### Usage

```Golang
package main

import (
	"github.com/PretendoNetwork/nex-go"
)

func main() {
	nexServer := nex.NewServer()

	nexServer.SetPrudpVersion(0)
	nexServer.SetSignatureVersion(1)
	nexServer.SetKerberosKeySize(16)
	nexServer.SetAccessKey("ridfebb9")

	nexServer.On("Data", func(packet *nex.PacketV0) {
		// Handle data packet
	})

	nexServer.Listen("192.168.0.28:60000")
}
```