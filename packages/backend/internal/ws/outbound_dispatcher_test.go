package ws

import (
	"sync"
	"testing"
	"time"
)

type blockingSendConnection struct {
	*fakeConnection
	started     chan struct{}
	release     chan struct{}
	startedOnce sync.Once
	releaseOnce sync.Once
}

func newBlockingSendConnection() *blockingSendConnection {
	return &blockingSendConnection{
		fakeConnection: newFakeConnection(),
		started:        make(chan struct{}),
		release:        make(chan struct{}),
	}
}

func (c *blockingSendConnection) SendJSON(value any) error {
	c.startedOnce.Do(func() { close(c.started) })
	<-c.release
	return c.fakeConnection.SendJSON(value)
}

func (c *blockingSendConnection) waitUntilBlocked(t *testing.T) {
	t.Helper()
	select {
	case <-c.started:
	case <-time.After(time.Second):
		t.Fatal("sender did not start")
	}
}

func (c *blockingSendConnection) unblock() {
	c.releaseOnce.Do(func() { close(c.release) })
}

func TestOutboundDispatcherPreservesEnqueueOrder(t *testing.T) {
	connection := newFakeConnection()
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
	fast := newFakeConnection()
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
	slow.waitUntilBlocked(t)
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
