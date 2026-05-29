package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Connection is the seam between transport and room/game logic.
// Two adapters: wsConn (production) and FakeConnection (tests).
type Connection interface {
	SendJSON(v any) error
	Close() error
}

// wsConn wraps *websocket.Conn for production use.
type wsConn struct {
	conn *websocket.Conn
}

func newWSConn(conn *websocket.Conn) Connection {
	return &wsConn{conn: conn}
}

func (w *wsConn) SendJSON(v any) error {
	return w.conn.WriteJSON(v)
}

func (w *wsConn) Close() error {
	return w.conn.Close()
}

// Client represents a connected player.
// PlayerID-keyed map supports reconnection.
type Client struct {
	Conn     Connection
	PlayerID string
}

// FakeConnection records messages for tests.
type FakeConnection struct {
	mu       sync.Mutex
	messages []any
	closed   bool
}

func NewFakeConnection() *FakeConnection {
	return &FakeConnection{}
}

func (f *FakeConnection) SendJSON(v any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, v)
	return nil
}

func (f *FakeConnection) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *FakeConnection) Messages() []any {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]any, len(f.messages))
	copy(result, f.messages)
	return result
}

func (f *FakeConnection) IsClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}
