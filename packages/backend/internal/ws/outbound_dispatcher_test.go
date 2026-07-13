package ws

import (
	"sync"
	"testing"
	"time"
)

type blockingSendConnection struct {
	*FakeConnection
	release chan struct{}
	once    sync.Once
}

func newBlockingSendConnection() *blockingSendConnection {
	return &blockingSendConnection{FakeConnection: NewFakeConnection(), release: make(chan struct{})}
}

func (c *blockingSendConnection) SendJSON(value any) error {
	<-c.release
	return c.FakeConnection.SendJSON(value)
}

func (c *blockingSendConnection) unblock() { c.once.Do(func() { close(c.release) }) }

func TestOutboundDispatcherPreservesEnqueueOrder(t *testing.T) {
	connection := NewFakeConnection()
	dispatcher := newOutboundDispatcher(nil)
	for revision := 1; revision <= 20; revision++ {
		dispatcher.send(connection, revision)
	}
	deadline := time.Now().Add(time.Second)
	for len(connection.Messages()) < 20 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	messages := connection.Messages()
	if len(messages) != 20 {
		t.Fatalf("received %d messages", len(messages))
	}
	for index, message := range messages {
		if message != index+1 {
			t.Fatalf("message %d = %v", index, message)
		}
	}
}

func TestSlowConnectionDoesNotBlockAnotherConnection(t *testing.T) {
	slow := newBlockingSendConnection()
	fast := NewFakeConnection()
	dispatcher := newOutboundDispatcher(nil)
	dispatcher.send(slow, "slow")
	dispatcher.send(fast, "fast")
	deadline := time.Now().Add(time.Second)
	for len(fast.Messages()) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(fast.Messages()) != 1 {
		t.Fatal("slow connection blocked fast connection")
	}
	slow.unblock()
}

func TestOutboundDispatcherClosesSlowConsumerWhenQueueIsFull(t *testing.T) {
	slow := newBlockingSendConnection()
	failed := make(chan struct{}, 1)
	dispatcher := newOutboundDispatcher(func(Connection) { failed <- struct{}{} })
	dispatcher.send(slow, "blocked")
	for index := 0; index < outboundQueueCapacity; index++ {
		dispatcher.send(slow, index)
	}
	if dispatcher.send(slow, "overflow") {
		t.Fatal("overflow message should be rejected")
	}
	select {
	case <-failed:
	case <-time.After(time.Second):
		t.Fatal("slow consumer was not closed")
	}
	slow.unblock()
}
