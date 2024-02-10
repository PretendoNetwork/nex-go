// Package nex provides a collection of utility structs, functions, and data types for making NEX/QRV servers
package nex

import (
	"net"

	"github.com/PretendoNetwork/nex-go/types"
)

// ClientInterface defines all the methods a client should have regardless of server type
type ClientInterface interface {
	Endpoint() EndpointInterface
	Address() net.Addr
	PID() *types.PID
	SetPID(pid *types.PID)
}
