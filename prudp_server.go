package nex

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/PretendoNetwork/nex-go/types"
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
	kerberosPassword              []byte
	kerberosTicketVersion         int
	kerberosKeySize               int
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
	passwordFromPIDHandler        func(pid *types.PID) (string, uint32)
	PRUDPv1ConnectionSignatureKey []byte
	byteStreamSettings            *ByteStreamSettings
	PRUDPV0Settings               *PRUDPV0Settings
}

// BindPRUDPEndPoint binds a provided PRUDPEndPoint to the server
func (s *PRUDPServer) BindPRUDPEndPoint(endpoint *PRUDPEndPoint) {
	if s.Endpoints.Has(endpoint.StreamID) {
		logger.Warningf("Tried to bind already existing PRUDPEndPoint %d", endpoint.StreamID)
		return
	}

	endpoint.Server = s
	s.Endpoints.Set(endpoint.StreamID, endpoint)
}

// Listen is an alias of ListenUDP. Implemented to conform to the ServerInterface
func (s *PRUDPServer) Listen(port int) {
	s.ListenUDP(port)
}

// ListenUDP starts a PRUDP server on a given port using a UDP server
func (s *PRUDPServer) ListenUDP(port int) {
	s.initPRUDPv1ConnectionSignatureKey()

	udpAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	socket, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		panic(err)
	}

	s.udpSocket = socket

	quit := make(chan struct{})

	for i := 0; i < runtime.NumCPU(); i++ {
		go s.listenDatagram(quit)
	}

	<-quit
}

func (s *PRUDPServer) listenDatagram(quit chan struct{}) {
	var err error

	for err == nil {
		buffer := make([]byte, 64000)
		var read int
		var addr *net.UDPAddr

		read, addr, err = s.udpSocket.ReadFromUDP(buffer)
		packetData := buffer[:read]

		err = s.handleSocketMessage(packetData, addr, nil)
	}

	quit <- struct{}{}

	panic(err)
}

// ListenWebSocket starts a PRUDP server on a given port using a WebSocket server
func (s *PRUDPServer) ListenWebSocket(port int) {
	s.initPRUDPv1ConnectionSignatureKey()
	//s.initVirtualPorts()

	s.websocketServer = &WebSocketServer{
		prudpServer: s,
	}

	s.websocketServer.listen(port)
}

// ListenWebSocketSecure starts a PRUDP server on a given port using a secure (TLS) WebSocket server
func (s *PRUDPServer) ListenWebSocketSecure(port int, certFile, keyFile string) {
	s.initPRUDPv1ConnectionSignatureKey()
	//s.initVirtualPorts()

	s.websocketServer = &WebSocketServer{
		prudpServer: s,
	}

	s.websocketServer.listenSecure(port, certFile, keyFile)
}

func (s *PRUDPServer) initPRUDPv1ConnectionSignatureKey() {
	// * Ensure the server has a key for PRUDPv1 connection signatures
	if len(s.PRUDPv1ConnectionSignatureKey) != 16 {
		s.PRUDPv1ConnectionSignatureKey = make([]byte, 16)
		_, err := rand.Read(s.PRUDPv1ConnectionSignatureKey)
		if err != nil {
			panic(err)
		}
	}
}

func (s *PRUDPServer) handleSocketMessage(packetData []byte, address net.Addr, webSocketConnection *gws.Conn) error {
	readStream := NewByteStreamIn(packetData, s)

	var packets []PRUDPPacketInterface

	// * Support any packet type the client sends and respond
	// * with that same type. Also keep reading from the stream
	// * until no more data is left, to account for multiple
	// * packets being sent at once
	if s.websocketServer != nil && packetData[0] == 0x80 {
		packets, _ = NewPRUDPPacketsLite(nil, readStream)
	} else if bytes.Equal(packetData[:2], []byte{0xEA, 0xD0}) {
		packets, _ = NewPRUDPPacketsV1(nil, readStream)
	} else {
		packets, _ = NewPRUDPPacketsV0(nil, readStream)
	}

	for _, packet := range packets {
		go s.processPacket(packet, address, webSocketConnection)
	}

	return nil
}

func (s *PRUDPServer) processPacket(packet PRUDPPacketInterface, address net.Addr, webSocketConnection *gws.Conn) {
	if !s.Endpoints.Has(packet.DestinationVirtualPortStreamID()) {
		logger.Warningf("Client %s trying to connect to unbound PRUDPEndPoint %d", address.String(), packet.DestinationVirtualPortStreamID())
		return
	}

	endpoint, ok := s.Endpoints.Get(packet.DestinationVirtualPortStreamID())
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
	socket, ok := s.Connections.Get(discriminator)
	if !ok {
		socket = NewSocketConnection(s, address, webSocketConnection)
		s.Connections.Set(discriminator, socket)
	}

	endpoint.processPacket(packet, socket)
}

// Send sends the packet to the packets sender
func (s *PRUDPServer) Send(packet PacketInterface) {
	if packet, ok := packet.(PRUDPPacketInterface); ok {
		data := packet.Payload()
		fragments := int(len(data) / s.FragmentSize)

		var fragmentID uint8 = 1
		for i := 0; i <= fragments; i++ {
			if len(data) < s.FragmentSize {
				packet.SetPayload(data)
				packet.setFragmentID(0)
			} else {
				packet.SetPayload(data[:s.FragmentSize])
				packet.setFragmentID(fragmentID)

				data = data[s.FragmentSize:]
				fragmentID++
			}

			s.sendPacket(packet)
		}
	}
}

func (s *PRUDPServer) sendPacket(packet PRUDPPacketInterface) {
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

			// * According to other Quazal server implementations,
			// * the RC4 stream is always reset to the default key
			// * regardless if the client is connecting to a secure
			// * server (prudps) or not
			if packet.Version() == 0 && s.PRUDPV0Settings.IsQuazalMode {
				slidingWindow.SetCipherKey([]byte("CD&ML"))
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

	s.sendRaw(packetCopy.Sender().(*PRUDPConnection).Socket, packetCopy.Bytes())
}

// sendRaw will send the given socket the provided packet
func (s *PRUDPServer) sendRaw(socket *SocketConnection, data []byte) {
	// TODO - Should this return the error too?

	var err error

	if address, ok := socket.Address.(*net.UDPAddr); ok && s.udpSocket != nil {
		_, err = s.udpSocket.WriteToUDP(data, address)
	} else if socket.WebSocketConnection != nil {
		err = socket.WebSocketConnection.WriteMessage(gws.OpcodeBinary, data)
	}

	if err != nil {
		logger.Error(err.Error())
	}
}

// AccessKey returns the servers sandbox access key
func (s *PRUDPServer) AccessKey() string {
	return s.accessKey
}

// SetAccessKey sets the servers sandbox access key
func (s *PRUDPServer) SetAccessKey(accessKey string) {
	s.accessKey = accessKey
}

// KerberosPassword returns the server kerberos password
func (s *PRUDPServer) KerberosPassword() []byte {
	return s.kerberosPassword
}

// SetKerberosPassword sets the server kerberos password
func (s *PRUDPServer) SetKerberosPassword(kerberosPassword []byte) {
	s.kerberosPassword = kerberosPassword
}

// SetFragmentSize sets the max size for a packets payload
func (s *PRUDPServer) SetFragmentSize(fragmentSize int) {
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
	s.FragmentSize = fragmentSize
}

// SetKerberosTicketVersion sets the version used when handling kerberos tickets
func (s *PRUDPServer) SetKerberosTicketVersion(kerberosTicketVersion int) {
	s.kerberosTicketVersion = kerberosTicketVersion
}

// KerberosKeySize gets the size for the kerberos session key
func (s *PRUDPServer) KerberosKeySize() int {
	return s.kerberosKeySize
}

// SetKerberosKeySize sets the size for the kerberos session key
func (s *PRUDPServer) SetKerberosKeySize(kerberosKeySize int) {
	s.kerberosKeySize = kerberosKeySize
}

// LibraryVersion returns the server NEX version
func (s *PRUDPServer) LibraryVersion() *LibraryVersion {
	return s.version
}

// SetDefaultLibraryVersion sets the default NEX protocol versions
func (s *PRUDPServer) SetDefaultLibraryVersion(version *LibraryVersion) {
	s.version = version
	s.datastoreProtocolVersion = version.Copy()
	s.matchMakingProtocolVersion = version.Copy()
	s.rankingProtocolVersion = version.Copy()
	s.ranking2ProtocolVersion = version.Copy()
	s.messagingProtocolVersion = version.Copy()
	s.utilityProtocolVersion = version.Copy()
	s.natTraversalProtocolVersion = version.Copy()
}

// DataStoreProtocolVersion returns the servers DataStore protocol version
func (s *PRUDPServer) DataStoreProtocolVersion() *LibraryVersion {
	return s.datastoreProtocolVersion
}

// SetDataStoreProtocolVersion sets the servers DataStore protocol version
func (s *PRUDPServer) SetDataStoreProtocolVersion(version *LibraryVersion) {
	s.datastoreProtocolVersion = version
}

// MatchMakingProtocolVersion returns the servers MatchMaking protocol version
func (s *PRUDPServer) MatchMakingProtocolVersion() *LibraryVersion {
	return s.matchMakingProtocolVersion
}

// SetMatchMakingProtocolVersion sets the servers MatchMaking protocol version
func (s *PRUDPServer) SetMatchMakingProtocolVersion(version *LibraryVersion) {
	s.matchMakingProtocolVersion = version
}

// RankingProtocolVersion returns the servers Ranking protocol version
func (s *PRUDPServer) RankingProtocolVersion() *LibraryVersion {
	return s.rankingProtocolVersion
}

// SetRankingProtocolVersion sets the servers Ranking protocol version
func (s *PRUDPServer) SetRankingProtocolVersion(version *LibraryVersion) {
	s.rankingProtocolVersion = version
}

// Ranking2ProtocolVersion returns the servers Ranking2 protocol version
func (s *PRUDPServer) Ranking2ProtocolVersion() *LibraryVersion {
	return s.ranking2ProtocolVersion
}

// SetRanking2ProtocolVersion sets the servers Ranking2 protocol version
func (s *PRUDPServer) SetRanking2ProtocolVersion(version *LibraryVersion) {
	s.ranking2ProtocolVersion = version
}

// MessagingProtocolVersion returns the servers Messaging protocol version
func (s *PRUDPServer) MessagingProtocolVersion() *LibraryVersion {
	return s.messagingProtocolVersion
}

// SetMessagingProtocolVersion sets the servers Messaging protocol version
func (s *PRUDPServer) SetMessagingProtocolVersion(version *LibraryVersion) {
	s.messagingProtocolVersion = version
}

// UtilityProtocolVersion returns the servers Utility protocol version
func (s *PRUDPServer) UtilityProtocolVersion() *LibraryVersion {
	return s.utilityProtocolVersion
}

// SetUtilityProtocolVersion sets the servers Utility protocol version
func (s *PRUDPServer) SetUtilityProtocolVersion(version *LibraryVersion) {
	s.utilityProtocolVersion = version
}

// SetNATTraversalProtocolVersion sets the servers NAT Traversal protocol version
func (s *PRUDPServer) SetNATTraversalProtocolVersion(version *LibraryVersion) {
	s.natTraversalProtocolVersion = version
}

// NATTraversalProtocolVersion returns the servers NAT Traversal protocol version
func (s *PRUDPServer) NATTraversalProtocolVersion() *LibraryVersion {
	return s.natTraversalProtocolVersion
}

// PasswordFromPID calls the function set with SetPasswordFromPIDFunction and returns the result
func (s *PRUDPServer) PasswordFromPID(pid *types.PID) (string, uint32) {
	if s.passwordFromPIDHandler == nil {
		logger.Errorf("Missing PasswordFromPID handler. Set with SetPasswordFromPIDFunction")
		return "", Errors.Core.NotImplemented
	}

	return s.passwordFromPIDHandler(pid)
}

// SetPasswordFromPIDFunction sets the function for the auth server to get a NEX password using the PID
func (s *PRUDPServer) SetPasswordFromPIDFunction(handler func(pid *types.PID) (string, uint32)) {
	s.passwordFromPIDHandler = handler
}

// ByteStreamSettings returns the settings to be used for ByteStreams
func (s *PRUDPServer) ByteStreamSettings() *ByteStreamSettings {
	return s.byteStreamSettings
}

// SetByteStreamSettings sets the settings to be used for ByteStreams
func (s *PRUDPServer) SetByteStreamSettings(byteStreamSettings *ByteStreamSettings) {
	s.byteStreamSettings = byteStreamSettings
}

// NewPRUDPServer will return a new PRUDP server
func NewPRUDPServer() *PRUDPServer {
	return &PRUDPServer{
		Endpoints:          NewMutexMap[uint8, *PRUDPEndPoint](),
		Connections:        NewMutexMap[string, *SocketConnection](),
		kerberosKeySize:    32,
		FragmentSize:       1300,
		pingTimeout:        time.Second * 15,
		byteStreamSettings: NewByteStreamSettings(),
		PRUDPV0Settings:    NewPRUDPV0Settings(),
	}
}
