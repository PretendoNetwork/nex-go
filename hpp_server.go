package nex

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/PretendoNetwork/nex-go/v2/types"
)

// HPPServer represents a bare-bones HPP server
type HPPServer struct {
	server                   *http.Server
	accessKey                string
	libraryVersions          *LibraryVersions
	dataHandlers             []func(packet PacketInterface)
	errorEventHandlers       []func(err *Error)
	byteStreamSettings       *ByteStreamSettings
	AccountDetailsByPID      func(pid types.PID) (*Account, *Error)
	AccountDetailsByUsername func(username string) (*Account, *Error)
	useVerboseRMC            bool
}

// RegisterServiceProtocol registers a NEX service with the HPP server
func (s *HPPServer) RegisterServiceProtocol(protocol ServiceProtocol) {
	protocol.SetEndpoint(s)
	s.OnData(protocol.HandlePacket)
}

// OnData adds an event handler which is fired when a new HPP request is received
func (s *HPPServer) OnData(handler func(packet PacketInterface)) {
	s.dataHandlers = append(s.dataHandlers, handler)
}

// EmitError calls all the endpoints error event handlers with the provided error
func (s *HPPServer) EmitError(err *Error) {
	for _, handler := range s.errorEventHandlers {
		go handler(err)
	}
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
	client.SetPID(types.NewPID(uint64(pid)))

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
		errorResponse := NewRMCError(s, ResultCodes.PythonCore.ValidationError)
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
	s.server.Addr = fmt.Sprintf(":%d", port)

	err := s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// ListenSecure starts a HPP server on a given port using a secure (TLS) server
func (s *HPPServer) ListenSecure(port int, certFile, keyFile string) {
	s.server.Addr = fmt.Sprintf(":%d", port)

	err := s.server.ListenAndServeTLS(certFile, keyFile)
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

// LibraryVersions returns the versions that the server has
func (s *HPPServer) LibraryVersions() *LibraryVersions {
	return s.libraryVersions
}

// AccessKey returns the servers sandbox access key
func (s *HPPServer) AccessKey() string {
	return s.accessKey
}

// SetAccessKey sets the servers sandbox access key
func (s *HPPServer) SetAccessKey(accessKey string) {
	s.accessKey = accessKey
}

// ByteStreamSettings returns the settings to be used for ByteStreams
func (s *HPPServer) ByteStreamSettings() *ByteStreamSettings {
	return s.byteStreamSettings
}

// SetByteStreamSettings sets the settings to be used for ByteStreams
func (s *HPPServer) SetByteStreamSettings(byteStreamSettings *ByteStreamSettings) {
	s.byteStreamSettings = byteStreamSettings
}

// UseVerboseRMC checks whether or not the endpoint uses verbose RMC
func (s *HPPServer) UseVerboseRMC() bool {
	return s.useVerboseRMC
}

// EnableVerboseRMC enable or disables the use of verbose RMC
func (s *HPPServer) EnableVerboseRMC(enable bool) {
	s.useVerboseRMC = enable
}

// NewHPPServer returns a new HPP server
func NewHPPServer() *HPPServer {
	s := &HPPServer{
		dataHandlers:       make([]func(packet PacketInterface), 0),
		errorEventHandlers: make([]func(err *Error), 0),
		libraryVersions:    NewLibraryVersions(),
		byteStreamSettings: NewByteStreamSettings(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hpp/", s.handleRequest)

	httpServer := &http.Server{
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS11, // * The 3DS and Wii U only support up to TLS 1.1 natively
		},
	}

	s.server = httpServer

	return s
}
