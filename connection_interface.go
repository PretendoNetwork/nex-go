// Package nex provides a collection of utility structs, functions, and data types for making NEX/QRV servers
package nex

import (
	"net"

	"github.com/PretendoNetwork/nex-go/v2/types"
)

// ConnectionInterface defines all the methods a connection should have regardless of server type
type ConnectionInterface interface {
	Endpoint() EndpointInterface
	Address() net.Addr
	PID() types.PID
	SetPID(pid types.PID)
}
