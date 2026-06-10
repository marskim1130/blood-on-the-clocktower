package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestCharacterAssignmentsArePrivateToPlayerAndStoryteller(t *testing.T) {
	h := NewHub()

	storytellerConn := NewFakeConnection()
	h.handleMessage(storytellerConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "storyteller",
		PlayerName: "Storyteller",
		MaxPlayers: 5,
	})
	roomID := storytellerConn.Messages()[0].(ServerMessage).RoomID

	playerConns := map[string]*FakeConnection{}
	for _, playerID := range []string{"p1", "p2", "p3", "p4", "p5"} {
		conn := NewFakeConnection()
		playerConns[playerID] = conn
		h.handleMessage(conn, ClientMessage{
			Type:       "JOIN_ROOM",
			RoomID:     roomID,
			PlayerID:   playerID,
			PlayerName: playerID,
		})
	}

	h.handleMessage(storytellerConn, ClientMessage{
		Type:           "SET_STORYTELLER",
		TargetPlayerID: "storyteller",
	})

	storytellerConn.ClearMessages()
	for _, conn := range playerConns {
		conn.ClearMessages()
	}

	h.handleMessage(storytellerConn, ClientMessage{
		Type: "ASSIGN_CHARACTERS",
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "librarian",
			"p3": "investigator",
			"p4": "poisoner",
			"p5": "imp",
		},
	})

	storytellerMessages := storytellerConn.Messages()
	storytellerAssignments := characterAssignmentMessages(storytellerMessages)
	if len(storytellerAssignments) != 5 {
		t.Fatalf("expected storyteller to receive 5 character assignments, got %d: %#v", len(storytellerAssignments), storytellerMessages)
	}
	storytellerState := lastRoomState(t, storytellerMessages)
	assertVisibleCharacters(t, storytellerState, map[string]bool{
		"p1": true,
		"p2": true,
		"p3": true,
		"p4": true,
		"p5": true,
	})

	for playerID, conn := range playerConns {
		messages := conn.Messages()
		assignments := characterAssignmentMessages(messages)
		if len(assignments) != 1 {
			t.Fatalf("expected %s to receive exactly 1 assignment, got %d: %#v", playerID, len(assignments), messages)
		}
		if assignments[0].PlayerID != playerID {
			t.Fatalf("expected %s to receive their own assignment, got %s", playerID, assignments[0].PlayerID)
		}

		state := lastRoomState(t, messages)
		expectedVisibility := map[string]bool{
			"p1": false,
			"p2": false,
			"p3": false,
			"p4": false,
			"p5": false,
		}
		expectedVisibility[playerID] = true
		assertVisibleCharacters(t, state, expectedVisibility)
	}
}

func TestCharacterAssignmentDoesNotLeakToSamePlayerIDInAnotherRoom(t *testing.T) {
	h := NewHub()

	currentRoom := h.rm.CreateRoom("storyteller", 5)
	otherRoom := h.rm.CreateRoom("other-creator", 5)
	storytellerConn := NewFakeConnection()
	otherRoomSamePlayerConn := NewFakeConnection()

	currentRoom.mu.Lock()
	currentRoom.addClient(storytellerConn, "storyteller")
	currentRoom.mu.Unlock()

	otherRoom.mu.Lock()
	otherRoom.addClient(otherRoomSamePlayerConn, "p1")
	otherRoom.mu.Unlock()

	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true},
		{ID: "p2", Name: "P2", IsAlive: true},
		{ID: "p3", Name: "P3", IsAlive: true},
		{ID: "p4", Name: "P4", IsAlive: true},
		{ID: "p5", Name: "P5", IsAlive: true},
	})

	h.mu.Lock()
	h.sessions[currentRoom.id] = gs
	h.connToRoom[storytellerConn] = currentRoom.id
	h.mu.Unlock()

	h.handleMessage(storytellerConn, ClientMessage{
		Type:           "SET_STORYTELLER",
		TargetPlayerID: "storyteller",
	})
	storytellerConn.ClearMessages()
	otherRoomSamePlayerConn.ClearMessages()

	h.handleMessage(storytellerConn, ClientMessage{
		Type: "ASSIGN_CHARACTERS",
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "librarian",
			"p3": "investigator",
			"p4": "poisoner",
			"p5": "imp",
		},
	})

	if assignments := characterAssignmentMessages(otherRoomSamePlayerConn.Messages()); len(assignments) != 0 {
		t.Fatalf("expected same playerID in another room to receive no character assignments, got %#v", assignments)
	}
}

func characterAssignmentMessages(messages []any) []game.CharacterAssigned {
	assignments := []game.CharacterAssigned{}
	for _, raw := range messages {
		msg, ok := raw.(ServerMessage)
		if !ok || msg.Event == nil || msg.Event.CharacterAssigned == nil {
			continue
		}
		assignments = append(assignments, *msg.Event.CharacterAssigned)
	}
	return assignments
}

func lastRoomState(t *testing.T, messages []any) *RoomState {
	t.Helper()
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(ServerMessage)
		if !ok || msg.Type != "ROOM_STATE" {
			continue
		}
		if msg.State == nil {
			t.Fatal("expected ROOM_STATE with state")
		}
		return msg.State
	}
	t.Fatalf("expected ROOM_STATE in messages: %#v", messages)
	return nil
}

func assertVisibleCharacters(t *testing.T, state *RoomState, expected map[string]bool) {
	t.Helper()
	if len(state.Players) != len(expected) {
		t.Fatalf("expected %d players in state, got %d", len(expected), len(state.Players))
	}
	for _, player := range state.Players {
		expectVisible, ok := expected[player.ID]
		if !ok {
			t.Fatalf("unexpected player %s in state", player.ID)
		}
		if expectVisible && player.Character == nil {
			t.Errorf("expected %s character to be visible", player.ID)
		}
		if !expectVisible && player.Character != nil {
			t.Errorf("expected %s character to be hidden, got %+v", player.ID, player.Character)
		}
	}
}
