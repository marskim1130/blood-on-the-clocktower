package ws

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type fakeConnection struct {
	mu         sync.Mutex
	messages   []any
	closed     bool
	incomingCh chan []byte
	closeCh    chan struct{}
}

func newFakeConnection() *fakeConnection {
	return &fakeConnection{closeCh: make(chan struct{})}
}

func (f *fakeConnection) SendJSON(v any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, v)
	return nil
}

func (f *fakeConnection) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	close(f.closeCh)
	return nil
}

func (f *fakeConnection) Messages() []any {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]any, len(f.messages))
	copy(result, f.messages)
	return result
}

func (f *fakeConnection) IsClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

func (f *fakeConnection) ClearMessages() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = nil
}

func (f *fakeConnection) ReadMessage() (int, []byte, error) {
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
