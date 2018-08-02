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

    server.On("Syn", func(client *net.UDPAddr, packet General.Packet) {
        fmt.Println("Handle SYN PRUDP Packet")
    })

    server.On("Connect", func(client *net.UDPAddr, packet General.Packet) {
        fmt.Println("Handle CONNECT PRUDP Packet")
    })

    server.On("Data", func(client *net.UDPAddr, packet General.Packet) {
        fmt.Println("Handle DATA PRUDP Packet")
    })

    server.On("Disconnect", func(client *net.UDPAddr, packet General.Packet) {
        fmt.Println("Handle DISCONNECT PRUDP Packet")
    })

    server.On("Ping", func(client *net.UDPAddr, packet General.Packet) {
        fmt.Println("Handle PING PRUDP Packet")
    })


    server.Listen(":60000")
}
```