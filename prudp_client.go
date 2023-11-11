package nex

import (
	"net"
	"time"
)

// PRUDPClient represents a single PRUDP client
type PRUDPClient struct {
	address                             *net.UDPAddr
	server                              *PRUDPServer
	pid                                 uint32
	clientConnectionSignature           []byte
	serverConnectionSignature           []byte
	sessionKey                          []byte
	reliableSubstreams                  []*ReliablePacketSubstreamManager
	outgoingUnreliableSequenceIDCounter *Counter[uint16]
	outgoingPingSequenceIDCounter       *Counter[uint16]
	heartbeatTimer                      *time.Timer
	pingKickTimer                       *time.Timer
	SourceStreamType                    uint8
	SourcePort                          uint8
	DestinationStreamType               uint8
	DestinationPort                     uint8
	minorVersion                        uint32 // * Not currently used for anything, but maybe useful later?
	supportedFunctions                  uint32 // * Not currently used for anything, but maybe useful later?
}

// Reset sets the client back to it's default state
func (c *PRUDPClient) reset() {
	for _, substream := range c.reliableSubstreams {
		substream.ResendScheduler.Stop()
	}

	c.clientConnectionSignature = make([]byte, 0)
	c.serverConnectionSignature = make([]byte, 0)
	c.sessionKey = make([]byte, 0)
	c.reliableSubstreams = make([]*ReliablePacketSubstreamManager, 0)
	c.outgoingUnreliableSequenceIDCounter = NewCounter[uint16](0)
	c.outgoingPingSequenceIDCounter = NewCounter[uint16](0)
	c.SourceStreamType = 0
	c.SourcePort = 0
	c.DestinationStreamType = 0
	c.DestinationPort = 0
}

// Cleanup cleans up any resources the client may be using
//
// This is similar to Client.Reset(), with the key difference
// being that Cleanup does not care about the state the client
// is currently in, or will be in, after execution. It only
// frees resources that are not easily garbage collected
func (c *PRUDPClient) cleanup() {
	for _, substream := range c.reliableSubstreams {
		substream.ResendScheduler.Stop()
	}

	c.reliableSubstreams = make([]*ReliablePacketSubstreamManager, 0)
	c.stopHeartbeatTimers()
}

// Server returns the server the client is connecting to
func (c *PRUDPClient) Server() ServerInterface {
	return c.server
}

// Address returns the clients address as a net.Addr
func (c *PRUDPClient) Address() net.Addr {
	return c.address
}

// PID returns the clients NEX PID
func (c *PRUDPClient) PID() uint32 {
	return c.pid
}

// SetPID sets the clients NEX PID
func (c *PRUDPClient) SetPID(pid uint32) {
	c.pid = pid
}

// SetSessionKey sets the clients session key used for reliable RC4 ciphers
func (c *PRUDPClient) setSessionKey(sessionKey []byte) {
	c.sessionKey = sessionKey

	c.reliableSubstreams[0].SetCipherKey(sessionKey)

	// * Only the first substream uses the session key directly.
	// * All other substreams modify the key before it so that
	// * all substreams have a unique cipher key
	for _, substream := range c.reliableSubstreams[1:] {
		modifier := len(sessionKey)/2 + 1

		// * Create a new slice to avoid modifying past keys
		sessionKey = append(make([]byte, 0), sessionKey...)

		// * Only the first half of the key is modified
		for i := 0; i < len(sessionKey)/2; i++ {
			sessionKey[i] = (sessionKey[i] + byte(modifier-i)) & 0xFF
		}

		substream.SetCipherKey(sessionKey)
	}
}

// ReliableSubstream returns the clients reliable substream ID
func (c *PRUDPClient) reliableSubstream(substreamID uint8) *ReliablePacketSubstreamManager {
	return c.reliableSubstreams[substreamID]
}

// CreateReliableSubstreams creates the list of substreams used for reliable PRUDP packets
func (c *PRUDPClient) createReliableSubstreams(maxSubstreamID uint8) {
	substreams := maxSubstreamID + 1

	c.reliableSubstreams = make([]*ReliablePacketSubstreamManager, substreams)

	for i := 0; i < len(c.reliableSubstreams); i++ {
		// * First DATA packet from the client has sequence ID 2
		// * First DATA packet from the server has sequence ID 1 (starts counter at 0 and is incremeneted)
		c.reliableSubstreams[i] = NewReliablePacketSubstreamManager(2, 0)
	}
}

func (c *PRUDPClient) nextOutgoingUnreliableSequenceID() uint16 {
	return c.outgoingUnreliableSequenceIDCounter.Next()
}

func (c *PRUDPClient) nextOutgoingPingSequenceID() uint16 {
	return c.outgoingPingSequenceIDCounter.Next()
}

func (c *PRUDPClient) resetHeartbeat() {
	if c.pingKickTimer != nil {
		c.pingKickTimer.Stop()
	}

	if c.heartbeatTimer != nil {
		c.heartbeatTimer.Reset(c.server.pingTimeout)
	}
}

func (c *PRUDPClient) startHeartbeat() {
	server := c.server

	// * Every time a packet is sent, client.resetHeartbeat()
	// * is called which resets this timer. If this function
	// * ever executes, it means we haven't seen the client
	// * in the expected time frame. If this happens, send
	// * the client a PING packet to try and kick start the
	// * heartbeat again
	c.heartbeatTimer = time.AfterFunc(server.pingTimeout, func() {
		server.sendPing(c)

		// * If the heartbeat still did not restart, assume the
		// * client is dead and clean up
		c.pingKickTimer = time.AfterFunc(server.pingTimeout, func() {
			c.cleanup()
			c.server.clients.Delete(c.address.String())
		})
	})
}

func (c *PRUDPClient) stopHeartbeatTimers() {
	if c.pingKickTimer != nil {
		c.pingKickTimer.Stop()
	}

	if c.heartbeatTimer != nil {
		c.heartbeatTimer.Stop()
	}
}

// NewPRUDPClient creates and returns a new Client using the provided UDP address and server
func NewPRUDPClient(address *net.UDPAddr, server *PRUDPServer) *PRUDPClient {
	return &PRUDPClient{
		address:                       address,
		server:                        server,
		outgoingPingSequenceIDCounter: NewCounter[uint16](0),
	}
}
