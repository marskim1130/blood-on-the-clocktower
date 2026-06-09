package ws

import (
	"sync"
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// TestConnToRoomRaceSetStoryteller verifies that handleSetStoryteller
// can read h.connToRoom concurrently with writes from other goroutines
// without triggering a Go map concurrent read/write panic.
//
// Run with: go test -race -run TestConnToRoomRaceSetStoryteller
func TestConnToRoomRaceSetStoryteller(t *testing.T) {
	h := NewHub()

	// Create a room with a creator
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	// Get the room ID from the creator's response
	creatorMsgs := creatorConn.Messages()
	if len(creatorMsgs) == 0 {
		t.Fatal("expected ROOM_STATE message from CREATE_ROOM")
	}
	roomID := creatorMsgs[0].(ServerMessage).RoomID

	// Add more players so the room has concurrent joiners
	const numJoiners = 10
	var wg sync.WaitGroup

	// Goroutine group 1: Concurrent joiners (write to connToRoom under h.mu.Lock)
	for i := 0; i < numJoiners; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := NewFakeConnection()
			h.handleMessage(conn, ClientMessage{
				Type:     "JOIN_ROOM",
				RoomID:   roomID,
				PlayerID: "player" + string(rune('A'+i)),
				PlayerName: "Player",
			})
		}(i)
	}

	// Goroutine group 2: Concurrent storyteller attempts (read connToRoom WITHOUT lock)
	// This is the race: these reads race with the joiner writes above.
	for i := 0; i < numJoiners; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.handleMessage(creatorConn, ClientMessage{
				Type:           "SET_STORYTELLER",
				PlayerID:       "creator",
				TargetPlayerID: "creator",
			})
		}()
	}

	wg.Wait()
	// With -race flag, this test will FAIL if connToRoom reads are unprotected.
	// After the fix, it should PASS.
}

// TestConnToRoomRaceAssignCharacters verifies that handleAssignCharacters
// can read h.connToRoom concurrently with writes from other goroutines.
func TestConnToRoomRaceAssignCharacters(t *testing.T) {
	h := NewHub()

	// Create a room with creator as storyteller
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	creatorMsgs := creatorConn.Messages()
	if len(creatorMsgs) == 0 {
		t.Fatal("expected ROOM_STATE message from CREATE_ROOM")
	}
	roomID := creatorMsgs[0].(ServerMessage).RoomID

	// Set storyteller
	h.handleMessage(creatorConn, ClientMessage{
		Type:           "SET_STORYTELLER",
		PlayerID:       "creator",
		TargetPlayerID: "creator",
	})

	// Add players
	const numPlayers = 5
	conns := make([]*FakeConnection, numPlayers)
	for i := 0; i < numPlayers; i++ {
		conns[i] = NewFakeConnection()
		h.handleMessage(conns[i], ClientMessage{
			Type:     "JOIN_ROOM",
			RoomID:   roomID,
			PlayerID: "p" + string(rune('1'+i)),
			PlayerName: "Player",
		})
	}

	var wg sync.WaitGroup

	// Goroutine group 1: More joiners (write to connToRoom)
	for i := numPlayers; i < numPlayers+5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := NewFakeConnection()
			h.handleMessage(conn, ClientMessage{
				Type:     "JOIN_ROOM",
				RoomID:   roomID,
				PlayerID: "extra" + string(rune('A'+i)),
				PlayerName: "Extra",
			})
		}(i)
	}

	// Goroutine group 2: Assign characters (read connToRoom WITHOUT lock)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.handleMessage(creatorConn, ClientMessage{
				Type:     "ASSIGN_CHARACTERS",
				PlayerID: "creator",
				Assignments: map[string]string{
					"p1": "washerwoman",
					"p2": "librarian",
					"p3": "investigator",
					"p4": "imp",
					"p5": "butler",
				},
			})
		}()
	}

	wg.Wait()
}

// TestConnToRoomRaceSubmitEvent verifies that handleSubmitEvent
// can read h.connToRoom concurrently with writes from other goroutines.
func TestConnToRoomRaceSubmitEvent(t *testing.T) {
	h := NewHub()

	// Create a room
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	creatorMsgs := creatorConn.Messages()
	if len(creatorMsgs) == 0 {
		t.Fatal("expected ROOM_STATE message from CREATE_ROOM")
	}
	roomID := creatorMsgs[0].(ServerMessage).RoomID

	var wg sync.WaitGroup

	// Goroutine group 1: Concurrent joiners (write to connToRoom)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := NewFakeConnection()
			h.handleMessage(conn, ClientMessage{
				Type:     "JOIN_ROOM",
				RoomID:   roomID,
				PlayerID: "joiner" + string(rune('A'+i)),
				PlayerName: "Joiner",
			})
		}(i)
	}

	// Goroutine group 2: Submit events (read connToRoom WITHOUT lock)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.handleMessage(creatorConn, ClientMessage{
				Type:     "SUBMIT_EVENT",
				PlayerID: "creator",
				Event: &game.GameEvent{
					PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay},
				},
			})
		}()
	}

	wg.Wait()
}
