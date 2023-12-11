package nex

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketServer wraps a  WebSocket server to create an easier API to consume
type WebSocketServer struct {
	mux                 *http.ServeMux
	upgrader            websocket.Upgrader
	handleSocketMessage func(packetData []byte, address net.Addr, webSocketConnection *websocket.Conn) error
}

func (ws *WebSocketServer) handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	conn.RemoteAddr()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			logger.Error(err.Error())
			return
		}

		ws.handleSocketMessage(data, conn.RemoteAddr(), conn)
	}
}

func (ws *WebSocketServer) initMux() {
	ws.mux = http.NewServeMux()
	ws.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := ws.upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		defer conn.Close()

		ws.handleConnection(conn)
	})
}

func (ws *WebSocketServer) listen(port int) {
	ws.initMux()

	http.ListenAndServe(fmt.Sprintf(":%d", port), ws.mux)
}

func (ws *WebSocketServer) listenSecure(port int, certFile, keyFile string) {
	ws.initMux()

	http.ListenAndServeTLS(fmt.Sprintf(":%d", port), certFile, keyFile, ws.mux)
}
