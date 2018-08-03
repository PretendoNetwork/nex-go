# NEX related libraries to make NEX servers in Go

[![GoDoc](https://godoc.org/github.com/PretendoNetwork/nex-go?status.svg)](https://godoc.org/github.com/PretendoNetwork/nex-go)

## How to install and use

### Install

Currently nex-go makes use of `prudplib`, a separate library used for handling PRUDP related functionality. This will soon be moved into the nex-go library

1. `go get https://github.com/PretendoNetwork/nex-go`
2. `go get https://github.com/PretendoNetwork/prudplib`

### Usage

```Golang
package main

import (
    "fmt"
    "net"

    NEXServer "github.com/PretendoNetwork/nex-go/server"
    "github.com/PretendoNetwork/prudplib/General"
)

func main() {
    server := NEXServer.NewServer()

    server.On("Syn", func(client NEXServer.Client, packet General.Packet) {
        // handle packet
        // build response
        server.Send(client, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
    })

    server.Listen(":60000")
}
```