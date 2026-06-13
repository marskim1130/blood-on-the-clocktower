package ws

import (
	"fmt"
	"sync"
	"testing"
)

// TestCreateRoomConcurrentUniqueness verifies that concurrent CreateRoom
// calls never produce duplicate room IDs. This catches the TOCTOU bug
// where generateRoomID checks uniqueness under RLock but CreateRoom
// inserts under Lock in a separate critical section.
func TestCreateRoomConcurrentUniqueness(t *testing.T) {
	rm := NewRoomManager()
	const goroutines = 100

	var wg sync.WaitGroup
	rooms := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			room := rm.CreateRoom("player", 10)
			if room == nil {
				t.Errorf("expected room for goroutine %d, got nil", id)
				return
			}
			rooms <- room.id
		}(i)
	}

	wg.Wait()
	close(rooms)

	seen := make(map[string]bool)
	for id := range rooms {
		if seen[id] {
			t.Errorf("duplicate room ID: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != goroutines {
		t.Errorf("expected %d unique room IDs, got %d", goroutines, len(seen))
	}
}

// TestGenerateRoomIDExhaustionReturnsError verifies that room ID generation
// fails without panicking when the 6-digit ID space is exhausted.
func TestGenerateRoomIDExhaustionReturnsError(t *testing.T) {
	rm := NewRoomManager()

	rm.mu.Lock()
	for i := 0; i < 1000000; i++ {
		id := fmt.Sprintf("%06d", i)
		rm.rooms[id] = &Room{id: id, clients: make(map[string]*Client)}
	}

	defer rm.mu.Unlock()

	if _, err := rm.generateRoomIDUnlocked(); err == nil {
		t.Fatal("expected room ID space exhaustion error")
	}
}

func TestCreateRoomReturnsNilWhenRoomIDSpaceIsExhausted(t *testing.T) {
	rm := NewRoomManager()

	rm.mu.Lock()
	for i := 0; i < 1000000; i++ {
		id := fmt.Sprintf("%06d", i)
		rm.rooms[id] = &Room{id: id, clients: make(map[string]*Client)}
	}
	rm.mu.Unlock()

	if room := rm.CreateRoom("player", 10); room != nil {
		t.Fatalf("expected nil room when room ID space is exhausted, got %#v", room)
	}
}
