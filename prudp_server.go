package nex

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/lxzan/gws"
)

// PRUDPServer represents a bare-bones PRUDP server
type PRUDPServer struct {
	udpSocket                     *net.UDPConn
	websocketServer               *WebSocketServer
	Endpoints                     *MutexMap[uint8, *PRUDPEndPoint]
	SupportedFunctions            uint32
	AccessKey                     string
	KerberosTicketVersion         int
	SessionKeyLength              int
	FragmentSize                  int
	PRUDPv1ConnectionSignatureKey []byte
	LibraryVersions               *LibraryVersions
	ByteStreamSettings            *ByteStreamSettings
	PRUDPV0Settings               *PRUDPV0Settings
	PRUDPV1Settings               *PRUDPV1Settings
	UseVerboseRMC                 bool
}

// BindPRUDPEndPoint binds a provided PRUDPEndPoint to the server
func (ps *PRUDPServer) BindPRUDPEndPoint(endpoint *PRUDPEndPoint) {
	if ps.Endpoints.Has(endpoint.StreamID) {
		logger.Warningf("Tried to bind already existing PRUDPEndPoint %d", endpoint.StreamID)
		return
	}

	endpoint.Server = ps
	ps.Endpoints.Set(endpoint.StreamID, endpoint)
}

// Listen is an alias of ListenUDP. Implemented to conform to the EndpointInterface
func (ps *PRUDPServer) Listen(port int) {
	ps.ListenUDP(port)
}

// ListenUDP starts a PRUDP server on a given port using a UDP server
func (ps *PRUDPServer) ListenUDP(port int) {
	ps.initPRUDPv1ConnectionSignatureKey()

	udpAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	socket, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		panic(err)
	}

	ps.udpSocket = socket

	quit := make(chan struct{})

	for i := 0; i < runtime.NumCPU(); i++ {
		go ps.listenDatagram(quit)
	}

	<-quit
}

func (ps *PRUDPServer) listenDatagram(quit chan struct{}) {
	var err error
	buffer := make([]byte, 64000)

	for err == nil {
		var read int
		var addr *net.UDPAddr

		read, addr, err = ps.udpSocket.ReadFromUDP(buffer)
		if err == nil {
			packetData := make([]byte, read)
			copy(packetData, buffer[:read])

			err = ps.handleSocketMessage(packetData, addr, nil)
		}
	}

	quit <- struct{}{}

	panic(err)
}

// ListenWebSocket starts a PRUDP server on a given port using a WebSocket server
func (ps *PRUDPServer) ListenWebSocket(port int) {
	ps.initPRUDPv1ConnectionSignatureKey()

	ps.websocketServer = &WebSocketServer{
		prudpServer: ps,
	}

	ps.websocketServer.listen(port)
}

// ListenWebSocketSecure starts a PRUDP server on a given port using a secure (TLS) WebSocket server
func (ps *PRUDPServer) ListenWebSocketSecure(port int, certFile, keyFile string) {
	ps.initPRUDPv1ConnectionSignatureKey()

	ps.websocketServer = &WebSocketServer{
		prudpServer: ps,
	}

	ps.websocketServer.listenSecure(port, certFile, keyFile)
}

func (ps *PRUDPServer) initPRUDPv1ConnectionSignatureKey() {
	// * Ensure the server has a key for PRUDPv1 connection signatures
	if len(ps.PRUDPv1ConnectionSignatureKey) != 16 {
		ps.PRUDPv1ConnectionSignatureKey = make([]byte, 16)
		_, err := rand.Read(ps.PRUDPv1ConnectionSignatureKey)
		if err != nil {
			panic(err)
		}
	}
}

func (ps *PRUDPServer) handleSocketMessage(packetData []byte, address net.Addr, webSocketConnection *gws.Conn) error {
	// * Check that the message is long enough for initial parsing
	if len(packetData) < 2 {
		return nil
	}

	readStream := NewByteStreamIn(packetData, ps.LibraryVersions, ps.ByteStreamSettings)

	var packets []PRUDPPacketInterface

	// * Support any packet type the client sends and respond
	// * with that same type. Also keep reading from the stream
	// * until no more data is left, to account for multiple
	// * packets being sent at once
	if ps.websocketServer != nil && packetData[0] == 0x80 {
		packets, _ = NewPRUDPPacketsLite(ps, nil, readStream)
	} else if bytes.Equal(packetData[:2], []byte{0xEA, 0xD0}) {
		packets, _ = NewPRUDPPacketsV1(ps, nil, readStream)
	} else {
		packets, _ = NewPRUDPPacketsV0(ps, nil, readStream)
	}

	for _, packet := range packets {
		go ps.processPacket(packet, address, webSocketConnection)
	}

	return nil
}

func (ps *PRUDPServer) processPacket(packet PRUDPPacketInterface, address net.Addr, webSocketConnection *gws.Conn) {
	if !ps.Endpoints.Has(packet.DestinationVirtualPortStreamID()) {
		logger.Warningf("Client %s trying to connect to unbound PRUDPEndPoint %d", address.String(), packet.DestinationVirtualPortStreamID())
		return
	}

	endpoint, ok := ps.Endpoints.Get(packet.DestinationVirtualPortStreamID())
	if !ok {
		logger.Warningf("Client %s trying to connect to unbound PRUDPEndPoint %d", address.String(), packet.DestinationVirtualPortStreamID())
		return
	}

	if packet.DestinationVirtualPortStreamType() != packet.SourceVirtualPortStreamType() {
		logger.Warningf("Client %s trying to use non matching destination and source stream types %d and %d", address.String(), packet.DestinationVirtualPortStreamType(), packet.SourceVirtualPortStreamType())
		return
	}

	if packet.DestinationVirtualPortStreamType() > constants.StreamTypeRelay {
		logger.Warningf("Client %s trying to use invalid to destination stream type %d", address.String(), packet.DestinationVirtualPortStreamType())
		return
	}

	if packet.SourceVirtualPortStreamType() > constants.StreamTypeRelay {
		logger.Warningf("Client %s trying to use invalid to source stream type %d", address.String(), packet.DestinationVirtualPortStreamType())
		return
	}

	sourcePortNumber := packet.SourceVirtualPortStreamID()
	invalidSourcePort := false

	// * PRUDPLite packets can use port numbers 0-31
	// * PRUDPv0 and PRUDPv1 can use port numbers 0-15
	if _, ok := packet.(*PRUDPPacketLite); ok {
		if sourcePortNumber > 31 {
			invalidSourcePort = true
		}
	} else if sourcePortNumber > 15 {
		invalidSourcePort = true
	}

	if invalidSourcePort {
		logger.Warningf("Client %s trying to use invalid to source port number %d. Port number too large", address.String(), sourcePortNumber)
		return
	}

	socket := NewSocketConnection(ps, address, webSocketConnection)
	endpoint.processPacket(packet, socket)
}

// Send sends the packet to the packets sender
func (ps *PRUDPServer) Send(packet PacketInterface) {
	if packet, ok := packet.(PRUDPPacketInterface); ok {
		data := packet.Payload()
		fragments := int(len(data) / ps.FragmentSize)

		var fragmentID uint8 = 1
		for i := 0; i <= fragments; i++ {
			if len(data) < ps.FragmentSize {
				packet.SetPayload(data)
				packet.setFragmentID(0)
			} else {
				packet.SetPayload(data[:ps.FragmentSize])
				packet.setFragmentID(fragmentID)

				data = data[ps.FragmentSize:]
				fragmentID++
			}

			ps.sendPacket(packet)

			// * This delay is here to prevent the server from overloading the client with too many packets.
			// * The 16ms (1/60th of a second) value is chosen based on testing with the friends server and is a good balance between
			// * Not being too slow and also not dropping any packets because we've overloaded the client. This may be because it
			// * roughly matches the framerate that most games target (60fps)
			if i < fragments {
				time.Sleep(16 * time.Millisecond)
			}
		}
	}
}

func (ps *PRUDPServer) sendPacket(packet PRUDPPacketInterface) {
	// * PRUDPServer.Send will send fragments as the same packet,
	// * just with different fields. In order to prevent modifying
	// * multiple packets at once, due to the same pointer being
	// * reused, we must make a copy of the packet being sent
	packetCopy := packet.Copy()
	connection := packetCopy.Sender().(*PRUDPConnection)

	if !packetCopy.HasFlag(constants.PacketFlagAck) && !packetCopy.HasFlag(constants.PacketFlagMultiAck) {
		if packetCopy.HasFlag(constants.PacketFlagReliable) {
			slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
			packetCopy.SetSequenceID(slidingWindow.NextOutgoingSequenceID())
		} else if packetCopy.Type() == constants.DataPacket {
			packetCopy.SetSequenceID(connection.outgoingUnreliableSequenceIDCounter.Next())
		} else if packetCopy.Type() == constants.PingPacket {
			packetCopy.SetSequenceID(connection.outgoingPingSequenceIDCounter.Next())
			connection.lastSentPingTime = time.Now()
		} else {
			packetCopy.SetSequenceID(0)
		}
	}

	packetCopy.SetSessionID(connection.ServerSessionID)

	if packetCopy.Type() == constants.DataPacket && !packetCopy.HasFlag(constants.PacketFlagAck) && !packetCopy.HasFlag(constants.PacketFlagMultiAck) {
		if packetCopy.HasFlag(constants.PacketFlagReliable) {
			slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
			payload := packetCopy.Payload()

			compressedPayload, err := slidingWindow.streamSettings.CompressionAlgorithm.Compress(payload)
			if err != nil {
				logger.Error(err.Error())
			}

			encryptedPayload, err := slidingWindow.streamSettings.EncryptionAlgorithm.Encrypt(compressedPayload)
			if err != nil {
				logger.Error(err.Error())
			}

			packetCopy.SetPayload(encryptedPayload)
		} else {
			// * PRUDPLite does not encrypt payloads, since they go over WSS
			if packetCopy.Version() != 2 {
				packetCopy.SetPayload(packetCopy.processUnreliableCrypto())
			}
		}
	}

	if ps.PRUDPV1Settings.LegacyConnectionSignature {
		packetCopy.setSignature(packetCopy.calculateSignature(connection.SessionKey, connection.Signature))
	} else {
		packetCopy.setSignature(packetCopy.calculateSignature(connection.SessionKey, connection.ServerConnectionSignature))
	}

	packetCopy.incrementSendCount()
	packetCopy.setSentAt(time.Now())

	if packetCopy.HasFlag(constants.PacketFlagReliable) && packetCopy.HasFlag(constants.PacketFlagNeedsAck) {
		slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
		slidingWindow.TimeoutManager.SchedulePacketTimeout(packetCopy)
	}

	ps.sendRaw(packetCopy.Sender().(*PRUDPConnection).Socket, packetCopy.Bytes())
}

// sendRaw will send the given socket the provided packet
func (ps *PRUDPServer) sendRaw(socket *SocketConnection, data []byte) {
	// TODO - Should this return the error too?

	var err error

	if address, ok := socket.Address.(*net.UDPAddr); ok && ps.udpSocket != nil {
		_, err = ps.udpSocket.WriteToUDP(data, address)
	} else if socket.WebSocketConnection != nil {
		err = socket.WebSocketConnection.WriteMessage(gws.OpcodeBinary, data)
	}

	if err != nil {
		logger.Error(err.Error())
	}
}

// SetFragmentSize sets the max size for a packets payload
func (ps *PRUDPServer) SetFragmentSize(fragmentSize int) {
	// TODO - Derive this value from the MTU
	// * From the wiki:
	// *
	// * The fragment size depends on the implementation.
	// * It is generally set to the MTU minus the packet overhead.
	// *
	// * In old NEX versions, which only support PRUDP v0, the MTU is
	// * hardcoded to 1000 and the maximum payload size seems to be 962 bytes.
	// *
	// * Later, the MTU was increased to 1364, and the maximum payload
	// * size is seems to be 1300 bytes, unless PRUDP v0 is used, in which case it’s 1264 bytes.
	ps.FragmentSize = fragmentSize
}

// NewPRUDPServer will return a new PRUDP server
func NewPRUDPServer() *PRUDPServer {
	return &PRUDPServer{
		Endpoints:          NewMutexMap[uint8, *PRUDPEndPoint](),
		SessionKeyLength:   32,
		FragmentSize:       1300,
		LibraryVersions:    NewLibraryVersions(),
		ByteStreamSettings: NewByteStreamSettings(),
		PRUDPV0Settings:    NewPRUDPV0Settings(),
		PRUDPV1Settings:    NewPRUDPV1Settings(),
	}
}
