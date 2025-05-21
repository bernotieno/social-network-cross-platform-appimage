package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader is a wrapper around gorilla/websocket.Upgrader
type Upgrader struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool
	upgrader        websocket.Upgrader
}

// Upgrade upgrades the HTTP server connection to the WebSocket protocol
func (u *Upgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	// Initialize the gorilla upgrader if not already done
	if u.upgrader.ReadBufferSize == 0 {
		u.upgrader = websocket.Upgrader{
			ReadBufferSize:  u.ReadBufferSize,
			WriteBufferSize: u.WriteBufferSize,
			CheckOrigin:     u.CheckOrigin,
		}
	}

	// Upgrade the connection
	return u.upgrader.Upgrade(w, r, responseHeader)
}
