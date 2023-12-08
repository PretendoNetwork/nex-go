package nex

import (
	"crypto/md5"
	"net"
	"time"
)

// PRUDPClient represents a single PRUDP client
type PRUDPClient struct {
	address                             *net.UDPAddr
	server                              *PRUDPServer
	pid                                 *PID
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
	ConnectionID                        uint32
	StationURLs                         []*StationURL
	unreliableBaseKey                   []byte
}

// reset sets the client back to it's default state
func (c *PRUDPClient) reset() {
	for _, substream := range c.reliableSubstreams {
		substream.ResendScheduler.Stop()
	}

	c.clientConnectionSignature = make([]byte, 0)
	c.serverConnectionSignature = make([]byte, 0)
	c.sessionKey = make([]byte, 0)
	c.reliableSubstreams = make([]*ReliablePacketSubstreamManager, 0)
	c.outgoingUnreliableSequenceIDCounter = NewCounter[uint16](1)
	c.outgoingPingSequenceIDCounter = NewCounter[uint16](0)
	c.SourceStreamType = 0
	c.SourcePort = 0
	c.DestinationStreamType = 0
	c.DestinationPort = 0
}

// cleanup cleans up any resources the client may be using
//
// This is similar to Client.reset(), with the key difference
// being that cleanup does not care about the state the client
// is currently in, or will be in, after execution. It only
// frees resources that are not easily garbage collected
func (c *PRUDPClient) cleanup() {
	for _, substream := range c.reliableSubstreams {
		substream.ResendScheduler.Stop()
	}

	c.reliableSubstreams = make([]*ReliablePacketSubstreamManager, 0)
	c.stopHeartbeatTimers()

	c.server.emitRemoved(c)
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
func (c *PRUDPClient) PID() *PID {
	return c.pid
}

// SetPID sets the clients NEX PID
func (c *PRUDPClient) SetPID(pid *PID) {
	c.pid = pid
}

// setSessionKey sets the clients session key used for reliable RC4 ciphers
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

	// * Init the base key used for unreliable DATA packets.
	// *
	// * Since unreliable DATA packets can come in out of
	// * order, each packet uses a dedicated RC4 stream. The
	// * key of each RC4 stream is made up by using this base
	// * key, modified using the packets sequence/session IDs
	unreliableBaseKeyPart1 := md5.Sum(append(sessionKey, []byte{0x18, 0xD8, 0x23, 0x34, 0x37, 0xE4, 0xE3, 0xFE}...))
	unreliableBaseKeyPart2 := md5.Sum(append(sessionKey, []byte{0x23, 0x3E, 0x60, 0x01, 0x23, 0xCD, 0xAB, 0x80}...))

	c.unreliableBaseKey = append(unreliableBaseKeyPart1[:], unreliableBaseKeyPart2[:]...)
}

// reliableSubstream returns the clients reliable substream ID
func (c *PRUDPClient) reliableSubstream(substreamID uint8) *ReliablePacketSubstreamManager {
	// * Fail-safe. The client may not always have
	// * the correct number of substreams. See the
	// * comment in handleSocketMessage of PRUDPServer
	// * for more details
	if int(substreamID) >= len(c.reliableSubstreams) {
		return c.reliableSubstreams[0]
	} else {
		return c.reliableSubstreams[substreamID]
	}
}

// createReliableSubstreams creates the list of substreams used for reliable PRUDP packets
func (c *PRUDPClient) createReliableSubstreams(maxSubstreamID uint8) {
	// * Kill any existing substreams
	for _, substream := range c.reliableSubstreams {
		substream.ResendScheduler.Stop()
	}

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
			c.cleanup() // * "removed" event is dispatched here
			virtualStream := c.server.virtualConnectionManager.Get(c.DestinationPort, c.DestinationStreamType)
			virtualStream.clients.Delete(c.address.String())
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
		pid:                           NewPID[uint32](0),
		unreliableBaseKey:             make([]byte, 0x20),
	}
}
