# NEX Go

[![GoDoc](https://godoc.org/github.com/PretendoNetwork/nex-go?status.svg)](https://godoc.org/github.com/PretendoNetwork/nex-go)

### Overview
NEX is the networking library used by all 1st party, and many 3rd party, games on the Nintendo Wii U, 3DS, and Switch which have online features. The NEX library has many different parts, ranging from low level packet transport to higher level service implementations

This library implements the lowest level parts of NEX, the transport protocols. For other parts of the NEX stack, see the below libraries. For detailed information on NEX as a whole, see our wiki docs https://nintendo-wiki.pretendo.network/docs/nex

### Install

```
go get github.com/PretendoNetwork/nex-go
```

### Other NEX libraries
- [nex-protocols-go](https://github.com/PretendoNetwork/nex-protocols-go) - NEX protocol definitions
- [nex-protocols-common-go](https://github.com/PretendoNetwork/nex-protocols-common-go) - Implementations of common NEX protocols which can be reused on many servers

### Quazal Rendez-Vous
Nintendo did not make NEX from scratch. NEX is largely based on an existing library called Rendez-Vous (QRV), made by Canadian software company Quazal. Quazal licensed Rendez-Vous out to many other companies, and was eventually bought out by Ubisoft. Because of this, QRV is seen in many many other games on all major platforms, especially Ubisoft

Nintendo modified Rendez-Vous somewhat heavily, simplifying the library/transport protocol quite a bit, and adding several custom services

While the main goal of this library is to support games which use the NEX variant of Rendez-Vous made by Nintendo, we also aim to be compatible with games using the original Rendez-Vous library. Due to the extensible nature of Rendez-Vous, many games may feature customizations much like NEX and have non-standard features/behavior. We do our best to support these cases, but there may be times where supporting all variations becomes untenable. In those cases, a fork of these libraries should be made instead if they require heavy modifications

### Supported features
- [x] [HPP servers](https://nintendo-wiki.pretendo.network/docs/hpp) (NEX over HTTP)
- [ ] [PRUDP servers](https://nintendo-wiki.pretendo.network/docs/prudp)
  - [x] UDP transport
  - [ ] WebSocket transport
  - [x] PRUDPv0 packets
  - [x] PRUDPv1 packets
  - [ ] PRUDPLite packets
- [x] Fragmented packet payloads
- [x] Packet retransmission
- [x] Reliable packets
- [x] Unreliable packets
- [x] [Virtual ports](https://nintendo-wiki.pretendo.network/docs/prudp#virtual-ports)
- [x] Packet compression
- [x] [RMC](https://nintendo-wiki.pretendo.network/docs/rmc)
  - [x] Request messages
  - [x] Response messages
  - [x] "Packed" encoded messages
  - [x] "Packed" (extended) encoded messages
  - [x] "Verbose" encoded messages
- [x] [Kerberos authentication](https://nintendo-wiki.pretendo.network/docs/nex/kerberos)

### Example

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
