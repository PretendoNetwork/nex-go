package nex

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

// HPPServer represents a bare-bones HPP server
type HPPServer struct {
	accessKey                   string
	version                     *LibraryVersion
	datastoreProtocolVersion    *LibraryVersion
	matchMakingProtocolVersion  *LibraryVersion
	rankingProtocolVersion      *LibraryVersion
	ranking2ProtocolVersion     *LibraryVersion
	messagingProtocolVersion    *LibraryVersion
	utilityProtocolVersion      *LibraryVersion
	natTraversalProtocolVersion *LibraryVersion
	dataHandlers                []func(packet PacketInterface)
	passwordFromPIDHandler      func(pid *PID) (string, uint32)
}

// OnData adds an event handler which is fired when a new HPP request is received
func (s *HPPServer) OnData(handler func(packet PacketInterface)) {
	s.dataHandlers = append(s.dataHandlers, handler)
}

func (s *HPPServer) handleRequest(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pidValue := req.Header.Get("pid")
	if pidValue == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// * The server checks that the header exists, but doesn't verify the value
	token := req.Header.Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessKeySignature := req.Header.Get("signature1")
	if accessKeySignature == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passwordSignature := req.Header.Get("signature2")
	if passwordSignature == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pid, err := strconv.Atoi(pidValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rmcRequestString := req.FormValue("file")

	rmcRequestBytes := []byte(rmcRequestString)

	tcpAddr, err := net.ResolveTCPAddr("tcp", req.RemoteAddr)
	if err != nil {
		// * Should never happen?
		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := NewHPPClient(tcpAddr, s)
	client.SetPID(NewPID(uint32(pid)))

	hppPacket, err := NewHPPPacket(client, rmcRequestBytes)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hppPacket.validateAccessKeySignature(accessKeySignature)
	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hppPacket.validatePasswordSignature(passwordSignature)
	if err != nil {
		logger.Error(err.Error())

		rmcMessage := hppPacket.RMCMessage()

		// HPP returns PythonCore::ValidationError if password is missing or invalid
		errorResponse := NewRMCError(Errors.PythonCore.ValidationError)
		errorResponse.CallID = rmcMessage.CallID
		errorResponse.IsHPP = true

		_, err = w.Write(errorResponse.Bytes())
		if err != nil {
			logger.Error(err.Error())
		}

		return
	}

	for _, dataHandler := range s.dataHandlers {
		go dataHandler(hppPacket)
	}

	<-hppPacket.processed

	if len(hppPacket.payload) > 0 {
		_, err = w.Write(hppPacket.payload)
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

// Listen starts a HPP server on a given port
func (s *HPPServer) Listen(port int) {
	http.HandleFunc("/hpp/", s.handleRequest)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

// Send sends the packet to the packets sender
func (s *HPPServer) Send(packet PacketInterface) {
	if packet, ok := packet.(*HPPPacket); ok {
		packet.message.IsHPP = true
		packet.payload = packet.message.Bytes()

		packet.processed <- true
	}
}

// AccessKey returns the servers sandbox access key
func (s *HPPServer) AccessKey() string {
	return s.accessKey
}

// SetAccessKey sets the servers sandbox access key
func (s *HPPServer) SetAccessKey(accessKey string) {
	s.accessKey = accessKey
}

// LibraryVersion returns the server NEX version
func (s *HPPServer) LibraryVersion() *LibraryVersion {
	return s.version
}

// SetDefaultLibraryVersion sets the default NEX protocol versions
func (s *HPPServer) SetDefaultLibraryVersion(version *LibraryVersion) {
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
func (s *HPPServer) DataStoreProtocolVersion() *LibraryVersion {
	return s.datastoreProtocolVersion
}

// SetDataStoreProtocolVersion sets the servers DataStore protocol version
func (s *HPPServer) SetDataStoreProtocolVersion(version *LibraryVersion) {
	s.datastoreProtocolVersion = version
}

// MatchMakingProtocolVersion returns the servers MatchMaking protocol version
func (s *HPPServer) MatchMakingProtocolVersion() *LibraryVersion {
	return s.matchMakingProtocolVersion
}

// SetMatchMakingProtocolVersion sets the servers MatchMaking protocol version
func (s *HPPServer) SetMatchMakingProtocolVersion(version *LibraryVersion) {
	s.matchMakingProtocolVersion = version
}

// RankingProtocolVersion returns the servers Ranking protocol version
func (s *HPPServer) RankingProtocolVersion() *LibraryVersion {
	return s.rankingProtocolVersion
}

// SetRankingProtocolVersion sets the servers Ranking protocol version
func (s *HPPServer) SetRankingProtocolVersion(version *LibraryVersion) {
	s.rankingProtocolVersion = version
}

// Ranking2ProtocolVersion returns the servers Ranking2 protocol version
func (s *HPPServer) Ranking2ProtocolVersion() *LibraryVersion {
	return s.ranking2ProtocolVersion
}

// SetRanking2ProtocolVersion sets the servers Ranking2 protocol version
func (s *HPPServer) SetRanking2ProtocolVersion(version *LibraryVersion) {
	s.ranking2ProtocolVersion = version
}

// MessagingProtocolVersion returns the servers Messaging protocol version
func (s *HPPServer) MessagingProtocolVersion() *LibraryVersion {
	return s.messagingProtocolVersion
}

// SetMessagingProtocolVersion sets the servers Messaging protocol version
func (s *HPPServer) SetMessagingProtocolVersion(version *LibraryVersion) {
	s.messagingProtocolVersion = version
}

// UtilityProtocolVersion returns the servers Utility protocol version
func (s *HPPServer) UtilityProtocolVersion() *LibraryVersion {
	return s.utilityProtocolVersion
}

// SetUtilityProtocolVersion sets the servers Utility protocol version
func (s *HPPServer) SetUtilityProtocolVersion(version *LibraryVersion) {
	s.utilityProtocolVersion = version
}

// SetNATTraversalProtocolVersion sets the servers NAT Traversal protocol version
func (s *HPPServer) SetNATTraversalProtocolVersion(version *LibraryVersion) {
	s.natTraversalProtocolVersion = version
}

// NATTraversalProtocolVersion returns the servers NAT Traversal protocol version
func (s *HPPServer) NATTraversalProtocolVersion() *LibraryVersion {
	return s.natTraversalProtocolVersion
}

// PasswordFromPID calls the function set with SetPasswordFromPIDFunction and returns the result
func (s *HPPServer) PasswordFromPID(pid *PID) (string, uint32) {
	if s.passwordFromPIDHandler == nil {
		logger.Errorf("Missing PasswordFromPID handler. Set with SetPasswordFromPIDFunction")
		return "", Errors.Core.InvalidHandle
	}

	return s.passwordFromPIDHandler(pid)
}

// SetPasswordFromPIDFunction sets the function for HPP to get a NEX password using the PID
func (s *HPPServer) SetPasswordFromPIDFunction(handler func(pid *PID) (string, uint32)) {
	s.passwordFromPIDHandler = handler
}

// NewHPPServer returns a new HPP server
func NewHPPServer() *HPPServer {
	return &HPPServer{
		dataHandlers: make([]func(packet PacketInterface), 0),
	}
}
