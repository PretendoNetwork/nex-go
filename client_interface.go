package nex

import "net"

// ClientInterface defines all the methods a client should have regardless of server type
type ClientInterface interface {
	Server() ServerInterface
	Address() net.Addr
	PID() uint32
	SetPID(pid uint32)
}
