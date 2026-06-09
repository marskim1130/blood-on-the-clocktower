package ws

import (
	"sync"
	"testing"
	"time"
)

// TestFakeConnectionConcurrentSendJSON verifies that FakeConnection
// handles concurrent SendJSON calls without data corruption.
// This test establishes the baseline — FakeConnection already has a mutex.
func TestFakeConnectionConcurrentSendJSON(t *testing.T) {
	conn := NewFakeConnection()
	const goroutines = 50
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				err := conn.SendJSON(map[string]int{"goroutine": id, "msg": j})
				if err != nil {
					t.Errorf("SendJSON failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	msgs := conn.Messages()
	expected := goroutines * messagesPerGoroutine
	if len(msgs) != expected {
		t.Errorf("expected %d messages, got %d", expected, len(msgs))
	}
}

// TestFakeConnectionReadMessageClosePropagation verifies that Close()
// properly unblocks ReadMessage(), both when incomingCh is set and
// when it is nil (the blocking-without-channel path).
func TestFakeConnectionReadMessageClosePropagation(t *testing.T) {
	t.Run("without incomingCh", func(t *testing.T) {
		conn := NewFakeConnection()

		done := make(chan error, 1)
		go func() {
			_, _, err := conn.ReadMessage()
			done <- err
		}()

		// Give the goroutine time to start blocking
		time.Sleep(20 * time.Millisecond)

		conn.Close()

		select {
		case err := <-done:
			if err == nil {
				t.Error("expected error from ReadMessage after Close, got nil")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("ReadMessage did not return within 2s after Close — goroutine leaked")
		}
	})

	t.Run("with incomingCh", func(t *testing.T) {
		conn := NewFakeConnection()
		conn.incomingCh = make(chan []byte)

		done := make(chan error, 1)
		go func() {
			_, _, err := conn.ReadMessage()
			done <- err
		}()

		// Give the goroutine time to start blocking on the channel
		time.Sleep(20 * time.Millisecond)

		conn.Close()

		select {
		case err := <-done:
			if err == nil {
				t.Error("expected error from ReadMessage after Close, got nil")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("ReadMessage did not return within 2s after Close — Close did not propagate")
		}
	})
}

// TestFakeConnectionReadMessageReturnsData verifies the happy path:
// data written to incomingCh is returned correctly.
func TestFakeConnectionReadMessageReturnsData(t *testing.T) {
	conn := NewFakeConnection()
	conn.incomingCh = make(chan []byte, 1)

	testData := []byte(`{"type":"TEST"}`)
	conn.incomingCh <- testData

	msgType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("expected %q, got %q", testData, data)
	}
	if msgType == 0 {
		t.Error("expected non-zero message type")
	}
}
