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
- [x] Quazal compatibility mode/settings
- [x] [HPP servers](https://nintendo-wiki.pretendo.network/docs/hpp) (NEX over HTTP)
- [x] [PRUDP servers](https://nintendo-wiki.pretendo.network/docs/prudp)
  - [x] UDP transport
  - [x] WebSocket transport (Experimental, largely untested)
  - [x] PRUDPv0 packets
  - [x] PRUDPv1 packets
  - [x] PRUDPLite packets
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

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
)

func main() {
	// Skeleton of a WiiU/3DS Friends server running on PRUDPv0 with a single endpoint

	authServer := nex.NewPRUDPServer() // The main PRUDP server
	endpoint := nex.NewPRUDPEndPoint(1) // A PRUDP endpoint for PRUDP connections to connect to. Bound to StreamID 1
	endpoint.ServerAccount = nex.NewAccount(types.NewPID(1), "Quazal Authentication", "password"))
	endpoint.AccountDetailsByPID = accountDetailsByPID
	endpoint.AccountDetailsByUsername = accountDetailsByUsername

	// Setup event handlers for the endpoint
	endpoint.OnData(func(packet nex.PacketInterface) {
		if packet, ok := packet.(nex.PRUDPPacketInterface); ok {
			request := packet.RMCMessage()

			fmt.Println("[AUTH]", request.ProtocolID, request.MethodID)

			if request.ProtocolID == 0xA { // TicketGrantingProtocol
				if request.MethodID == 0x1 { // TicketGrantingProtocol::Login
					handleLogin(packet)
				}

				if request.MethodID == 0x3 { // TicketGrantingProtocol::RequestTicket
					handleRequestTicket(packet)
				}
			}
		}
	})

	// Bind the endpoint to the server and configure it's settings
	authServer.BindPRUDPEndPoint(endpoint)
	authServer.SetFragmentSize(962)
	authServer.LibraryVersions.SetDefault(nex.NewLibraryVersion(1, 1, 0))
	authServer.SessionKeyLength = 16
	authServer.AccessKey = "ridfebb9"
	authServer.Listen(60000)
}
```
