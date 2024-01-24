package nex

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/lxzan/gws"
)

// PRUDPServer represents a bare-bones PRUDP server
type PRUDPServer struct {
	udpSocket                     *net.UDPConn
	websocketServer               *WebSocketServer
	Endpoints                     *MutexMap[uint8, *PRUDPEndPoint]
	Connections                   *MutexMap[string, *SocketConnection]
	SupportedFunctions            uint32
	accessKey                     string
	KerberosTicketVersion         int
	SessionKeyLength              int
	FragmentSize                  int
	version                       *LibraryVersion
	datastoreProtocolVersion      *LibraryVersion
	matchMakingProtocolVersion    *LibraryVersion
	rankingProtocolVersion        *LibraryVersion
	ranking2ProtocolVersion       *LibraryVersion
	messagingProtocolVersion      *LibraryVersion
	utilityProtocolVersion        *LibraryVersion
	natTraversalProtocolVersion   *LibraryVersion
	pingTimeout                   time.Duration
	PRUDPv1ConnectionSignatureKey []byte
	byteStreamSettings            *ByteStreamSettings
	PRUDPV0Settings               *PRUDPV0Settings
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

// Listen is an alias of ListenUDP. Implemented to conform to the ServerInterface
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

	for err == nil {
		buffer := make([]byte, 64000)
		var read int
		var addr *net.UDPAddr

		read, addr, err = ps.udpSocket.ReadFromUDP(buffer)
		packetData := buffer[:read]

		err = ps.handleSocketMessage(packetData, addr, nil)
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
	readStream := NewByteStreamIn(packetData, ps)

	var packets []PRUDPPacketInterface

	// * Support any packet type the client sends and respond
	// * with that same type. Also keep reading from the stream
	// * until no more data is left, to account for multiple
	// * packets being sent at once
	if ps.websocketServer != nil && packetData[0] == 0x80 {
		packets, _ = NewPRUDPPacketsLite(nil, readStream)
	} else if bytes.Equal(packetData[:2], []byte{0xEA, 0xD0}) {
		packets, _ = NewPRUDPPacketsV1(nil, readStream)
	} else {
		packets, _ = NewPRUDPPacketsV0(nil, readStream)
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

	if packet.DestinationVirtualPortStreamType() > StreamTypeRelay {
		logger.Warningf("Client %s trying to use invalid to destination stream type %d", address.String(), packet.DestinationVirtualPortStreamType())
		return
	}

	if packet.SourceVirtualPortStreamType() > StreamTypeRelay {
		logger.Warningf("Client %s trying to use invalid to source stream type %d", address.String(), packet.DestinationVirtualPortStreamType())
		return
	}

	sourcePortNumber := packet.SourceVirtualPortStreamID()
	invalidSourcePort := false

	// * PRUDPLite packets can use port numbers 0-31
	// * PRUDPv0 and PRUDPv1 can use port numbers 0-15
	if _, ok := packet.(*PRUDPPacketLite); ok && sourcePortNumber > 31 {
		invalidSourcePort = true
	} else if sourcePortNumber > 15 {
		invalidSourcePort = true
	}

	if invalidSourcePort {
		logger.Warningf("Client %s trying to use invalid to source port number %d. Port number too large", address.String(), sourcePortNumber)
		return
	}

	discriminator := address.String()
	socket, ok := ps.Connections.Get(discriminator)
	if !ok {
		socket = NewSocketConnection(ps, address, webSocketConnection)
		ps.Connections.Set(discriminator, socket)
	}

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

	if !packetCopy.HasFlag(FlagAck) && !packetCopy.HasFlag(FlagMultiAck) {
		if packetCopy.HasFlag(FlagReliable) {
			slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
			packetCopy.SetSequenceID(slidingWindow.NextOutgoingSequenceID())
		} else if packetCopy.Type() == DataPacket {
			packetCopy.SetSequenceID(connection.outgoingUnreliableSequenceIDCounter.Next())
		} else if packetCopy.Type() == PingPacket {
			packetCopy.SetSequenceID(connection.outgoingPingSequenceIDCounter.Next())
		} else {
			packetCopy.SetSequenceID(0)
		}
	}

	packetCopy.SetSessionID(connection.ServerSessionID)

	if packetCopy.Type() == DataPacket && !packetCopy.HasFlag(FlagAck) && !packetCopy.HasFlag(FlagMultiAck) {
		if packetCopy.HasFlag(FlagReliable) {
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

	packetCopy.setSignature(packetCopy.calculateSignature(connection.SessionKey, connection.ServerConnectionSignature))

	if packetCopy.HasFlag(FlagReliable) && packetCopy.HasFlag(FlagNeedsAck) {
		slidingWindow := connection.SlidingWindow(packetCopy.SubstreamID())
		slidingWindow.ResendScheduler.AddPacket(packetCopy)
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

// AccessKey returns the servers sandbox access key
func (ps *PRUDPServer) AccessKey() string {
	return ps.accessKey
}

// SetAccessKey sets the servers sandbox access key
func (ps *PRUDPServer) SetAccessKey(accessKey string) {
	ps.accessKey = accessKey
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

// LibraryVersion returns the server NEX version
func (ps *PRUDPServer) LibraryVersion() *LibraryVersion {
	return ps.version
}

// SetDefaultLibraryVersion sets the default NEX protocol versions
func (ps *PRUDPServer) SetDefaultLibraryVersion(version *LibraryVersion) {
	ps.version = version
	ps.datastoreProtocolVersion = version.Copy()
	ps.matchMakingProtocolVersion = version.Copy()
	ps.rankingProtocolVersion = version.Copy()
	ps.ranking2ProtocolVersion = version.Copy()
	ps.messagingProtocolVersion = version.Copy()
	ps.utilityProtocolVersion = version.Copy()
	ps.natTraversalProtocolVersion = version.Copy()
}

// DataStoreProtocolVersion returns the servers DataStore protocol version
func (ps *PRUDPServer) DataStoreProtocolVersion() *LibraryVersion {
	return ps.datastoreProtocolVersion
}

// SetDataStoreProtocolVersion sets the servers DataStore protocol version
func (ps *PRUDPServer) SetDataStoreProtocolVersion(version *LibraryVersion) {
	ps.datastoreProtocolVersion = version
}

// MatchMakingProtocolVersion returns the servers MatchMaking protocol version
func (ps *PRUDPServer) MatchMakingProtocolVersion() *LibraryVersion {
	return ps.matchMakingProtocolVersion
}

// SetMatchMakingProtocolVersion sets the servers MatchMaking protocol version
func (ps *PRUDPServer) SetMatchMakingProtocolVersion(version *LibraryVersion) {
	ps.matchMakingProtocolVersion = version
}

// RankingProtocolVersion returns the servers Ranking protocol version
func (ps *PRUDPServer) RankingProtocolVersion() *LibraryVersion {
	return ps.rankingProtocolVersion
}

// SetRankingProtocolVersion sets the servers Ranking protocol version
func (ps *PRUDPServer) SetRankingProtocolVersion(version *LibraryVersion) {
	ps.rankingProtocolVersion = version
}

// Ranking2ProtocolVersion returns the servers Ranking2 protocol version
func (ps *PRUDPServer) Ranking2ProtocolVersion() *LibraryVersion {
	return ps.ranking2ProtocolVersion
}

// SetRanking2ProtocolVersion sets the servers Ranking2 protocol version
func (ps *PRUDPServer) SetRanking2ProtocolVersion(version *LibraryVersion) {
	ps.ranking2ProtocolVersion = version
}

// MessagingProtocolVersion returns the servers Messaging protocol version
func (ps *PRUDPServer) MessagingProtocolVersion() *LibraryVersion {
	return ps.messagingProtocolVersion
}

// SetMessagingProtocolVersion sets the servers Messaging protocol version
func (ps *PRUDPServer) SetMessagingProtocolVersion(version *LibraryVersion) {
	ps.messagingProtocolVersion = version
}

// UtilityProtocolVersion returns the servers Utility protocol version
func (ps *PRUDPServer) UtilityProtocolVersion() *LibraryVersion {
	return ps.utilityProtocolVersion
}

// SetUtilityProtocolVersion sets the servers Utility protocol version
func (ps *PRUDPServer) SetUtilityProtocolVersion(version *LibraryVersion) {
	ps.utilityProtocolVersion = version
}

// SetNATTraversalProtocolVersion sets the servers NAT Traversal protocol version
func (ps *PRUDPServer) SetNATTraversalProtocolVersion(version *LibraryVersion) {
	ps.natTraversalProtocolVersion = version
}

// NATTraversalProtocolVersion returns the servers NAT Traversal protocol version
func (ps *PRUDPServer) NATTraversalProtocolVersion() *LibraryVersion {
	return ps.natTraversalProtocolVersion
}

// ByteStreamSettings returns the settings to be used for ByteStreams
func (ps *PRUDPServer) ByteStreamSettings() *ByteStreamSettings {
	return ps.byteStreamSettings
}

// SetByteStreamSettings sets the settings to be used for ByteStreams
func (ps *PRUDPServer) SetByteStreamSettings(byteStreamSettings *ByteStreamSettings) {
	ps.byteStreamSettings = byteStreamSettings
}

// NewPRUDPServer will return a new PRUDP server
func NewPRUDPServer() *PRUDPServer {
	return &PRUDPServer{
		Endpoints:          NewMutexMap[uint8, *PRUDPEndPoint](),
		Connections:        NewMutexMap[string, *SocketConnection](),
		SessionKeyLength:   32,
		FragmentSize:       1300,
		pingTimeout:        time.Second * 15,
		byteStreamSettings: NewByteStreamSettings(),
		PRUDPV0Settings:    NewPRUDPV0Settings(),
	}
}
