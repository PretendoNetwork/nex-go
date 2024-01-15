package nex

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lxzan/gws"
)

const (
	pingInterval = 5 * time.Second
	pingWait     = 10 * time.Second
)

type wsEventHandler struct {
	prudpServer *PRUDPServer
}

func (wseh *wsEventHandler) OnOpen(socket *gws.Conn) {
	_ = socket.SetDeadline(time.Now().Add(pingInterval + pingWait))
}

func (wseh *wsEventHandler) OnClose(wsConn *gws.Conn, err error) {
	connections := make([]*PRUDPConnection, 0)

	socket, ok := wseh.prudpServer.Connections.Get(wsConn.RemoteAddr().String())
	if !ok {
		// TODO - Error?
		return
	}

	socket.Connections.Each(func(_ uint8, connection *PRUDPConnection) bool {
		connections = append(connections, connection)
		return false
	})

	// * We cannot modify a MutexMap while looping over it
	// * since the mutex is locked. We first need to grab
	// * the entries we want to delete, and then loop over
	// * them here to actually clean them up
	for _, connection := range connections {
		connection.cleanup() // * "removed" event is dispatched here
	}
}

func (wseh *wsEventHandler) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.SetDeadline(time.Now().Add(pingInterval + pingWait))
	_ = socket.WritePong(nil)
}

func (wseh *wsEventHandler) OnPong(socket *gws.Conn, payload []byte) {}

func (wseh *wsEventHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	// * Create a COPY of the underlying *bytes.Buffer bytes.
	// * If this is not done, then the byte slice sometimes
	// * gets modified in unexpected places
	packetData := append([]byte(nil), message.Bytes()...)
	err := wseh.prudpServer.handleSocketMessage(packetData, socket.RemoteAddr(), socket)
	if err != nil {
		logger.Error(err.Error())
	}
}

// WebSocketServer wraps a WebSocket server to create an easier API to consume
type WebSocketServer struct {
	mux         *http.ServeMux
	upgrader    *gws.Upgrader
	prudpServer *PRUDPServer
}

func (ws *WebSocketServer) init() {
	ws.upgrader = gws.NewUpgrader(&wsEventHandler{
		prudpServer: ws.prudpServer,
	}, &gws.ServerOption{
		ReadAsyncEnabled: true,         // * Parallel message processing
		Recovery:         gws.Recovery, // * Exception recovery
		ReadBufferSize:   64000,
		WriteBufferSize:  64000,
	})

	ws.mux = http.NewServeMux()
	ws.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		socket, err := ws.upgrader.Upgrade(w, r)
		if err != nil {
			return
		}

		go func() {
			socket.ReadLoop() // * Blocking prevents the context from being GC
		}()
	})
}

func (ws *WebSocketServer) listen(port int) {
	ws.init()

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), ws.mux)
	if err != nil {
		panic(err)
	}
}

func (ws *WebSocketServer) listenSecure(port int, certFile, keyFile string) {
	ws.init()

	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, ws.mux)
	if err != nil {
		panic(err)
	}
}
