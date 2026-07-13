package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Connection is the seam between WebSocket transport and ordered delivery.
type Connection interface {
	SendJSON(v any) error
	Close() error
	// ReadMessage reads the next message from the connection.
	// Returns (messageType, payload, error).
	ReadMessage() (int, []byte, error)
}

// wsConn wraps *websocket.Conn for production use.
// mu protects concurrent writes — gorilla/websocket permits exactly one concurrent writer per connection.
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func newWSConn(conn *websocket.Conn) Connection {
	return &wsConn{conn: conn}
}

func (w *wsConn) SendJSON(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(v)
}

func (w *wsConn) Close() error {
	return w.conn.Close()
}

func (w *wsConn) ReadMessage() (int, []byte, error) {
	return w.conn.ReadMessage()
}
