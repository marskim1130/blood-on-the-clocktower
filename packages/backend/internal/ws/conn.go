package ws

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Connection is the seam between transport and room/game logic.
// Two adapters: wsConn (production) and FakeConnection (tests).
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

// Client represents a connected player.
// PlayerID-keyed map supports reconnection.
type Client struct {
	Conn     Connection
	PlayerID string
}

// FakeConnection records messages for tests.
// closeCh is closed on Close() to unblock ReadMessage callers.
// Set incomingCh to simulate inbound messages for ReadMessage;
// even with incomingCh set, Close() will unblock a blocking read.
type FakeConnection struct {
	mu         sync.Mutex
	messages   []any
	closed     bool
	incomingCh chan []byte   // if set, ReadMessage reads from here
	closeCh    chan struct{} // closed when connection is closed
}

func NewFakeConnection() *FakeConnection {
	return &FakeConnection{
		closeCh: make(chan struct{}),
	}
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
	if f.closed {
		return nil
	}
	f.closed = true
	close(f.closeCh)
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

func (f *FakeConnection) ClearMessages() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = nil
}

// ReadMessage reads from incomingCh if set, otherwise blocks until closed.
func (f *FakeConnection) ReadMessage() (int, []byte, error) {
	f.mu.Lock()
	ch := f.incomingCh
	closeCh := f.closeCh
	f.mu.Unlock()

	if ch != nil {
		select {
		case data, ok := <-ch:
			if !ok {
				return 0, nil, fmt.Errorf("connection closed")
			}
			return websocket.TextMessage, data, nil
		case <-closeCh:
			return 0, nil, fmt.Errorf("connection closed")
		}
	}
	<-closeCh
	return 0, nil, fmt.Errorf("connection closed")
}
