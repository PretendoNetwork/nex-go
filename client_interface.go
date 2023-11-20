// Package nex provides a collection of utility structs, functions, and data types for making NEX/QRV servers
package nex

import "net"

// ClientInterface defines all the methods a client should have regardless of server type
type ClientInterface interface {
	Server() ServerInterface
	Address() net.Addr
	PID() *PID
	SetPID(pid *PID)
}
