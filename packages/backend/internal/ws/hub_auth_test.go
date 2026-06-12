package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// TestSetStorytellerRejectsForgedPlayerID verifies that a non-creator
// connection cannot bypass the creator check by forging msg.PlayerID.
//
// Attack scenario: attacker sends SET_STORYTELLER with PlayerID="creator",
// but the connection actually belongs to "attacker". The server should
// derive identity from the connection, not trust the message payload.
func TestSetStorytellerRejectsForgedPlayerID(t *testing.T) {
	h := NewHub()

	// Creator creates a room
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	creatorMsgs := creatorConn.Messages()
	if len(creatorMsgs) == 0 {
		t.Fatal("expected ROOM_STATE from CREATE_ROOM")
	}
	roomID := creatorMsgs[0].(ServerMessage).RoomID

	// Attacker joins the room
	attackerConn := NewFakeConnection()
	h.handleMessage(attackerConn, ClientMessage{
		Type:       "JOIN_ROOM",
		RoomID:     roomID,
		PlayerID:   "attacker",
		PlayerName: "Attacker",
	})

	// Clear attacker's messages
	attackerConn.ClearMessages()

	// Attacker sends SET_STORYTELLER with FORGED PlayerID = "creator"
	h.handleMessage(attackerConn, ClientMessage{
		Type:           "SET_STORYTELLER",
		PlayerID:       "creator", // forged!
		TargetPlayerID: "attacker",
	})

	// Should be rejected — the connection belongs to "attacker", not "creator"
	msgs := attackerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response for forged PlayerID")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok {
		t.Fatalf("expected ServerMessage, got %T", msgs[0])
	}
	if errMsg.Type != "ERROR" {
		t.Errorf("expected ERROR response, got %s", errMsg.Type)
	}
	if errMsg.Error != "only room creator can set storyteller" {
		t.Errorf("expected 'only room creator can set storyteller', got %q", errMsg.Error)
	}
}

// TestAssignCharactersRejectsForgedPlayerID verifies that a non-storyteller
// connection cannot assign characters by forging msg.PlayerID.
func TestAssignCharactersRejectsForgedPlayerID(t *testing.T) {
	h := NewHub()

	// Creator creates a room and sets storyteller
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	creatorMsgs := creatorConn.Messages()
	if len(creatorMsgs) == 0 {
		t.Fatal("expected ROOM_STATE from CREATE_ROOM")
	}
	roomID := creatorMsgs[0].(ServerMessage).RoomID

	// Set creator as storyteller
	h.handleMessage(creatorConn, ClientMessage{
		Type:           "SET_STORYTELLER",
		PlayerID:       "creator",
		TargetPlayerID: "creator",
	})

	// Add another player
	otherConn := NewFakeConnection()
	h.handleMessage(otherConn, ClientMessage{
		Type:       "JOIN_ROOM",
		RoomID:     roomID,
		PlayerID:   "other",
		PlayerName: "Other",
	})

	// Clear messages
	otherConn.ClearMessages()

	// Other player sends ASSIGN_CHARACTERS with FORGED PlayerID = "creator"
	h.handleMessage(otherConn, ClientMessage{
		Type:     "ASSIGN_CHARACTERS",
		PlayerID: "creator", // forged!
		Assignments: map[string]string{
			"other": "imp",
		},
	})

	// Should be rejected — the connection belongs to "other", not the storyteller "creator"
	msgs := otherConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response for forged PlayerID")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok {
		t.Fatalf("expected ServerMessage, got %T", msgs[0])
	}
	if errMsg.Type != "ERROR" {
		t.Errorf("expected ERROR response, got %s", errMsg.Type)
	}
	if errMsg.Error != "only storyteller can assign characters" {
		t.Errorf("expected 'only storyteller can assign characters', got %q", errMsg.Error)
	}
}

// TestLeaveRoomIgnoresForgedPlayerID verifies that a player cannot force
// another player to leave by forging msg.PlayerID. The server must derive
// the player identity from the connection, not trust the message payload.
//
// Attack scenario: attacker sends LEAVE_ROOM with PlayerID="victim",
// but the connection actually belongs to "attacker". The server should
// use the connection-derived identity, so the attacker leaves themselves,
// not the victim.
func TestLeaveRoomIgnoresForgedPlayerID(t *testing.T) {
	h := NewHub()

	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID

	// Victim joins
	victimConn := NewFakeConnection()
	h.handleMessage(victimConn, ClientMessage{
		Type:       "JOIN_ROOM",
		RoomID:     roomID,
		PlayerID:   "victim",
		PlayerName: "Victim",
	})

	// Attacker joins
	attackerConn := NewFakeConnection()
	h.handleMessage(attackerConn, ClientMessage{
		Type:       "JOIN_ROOM",
		RoomID:     roomID,
		PlayerID:   "attacker",
		PlayerName: "Attacker",
	})

	attackerConn.ClearMessages()

	// Attacker sends LEAVE_ROOM with forged PlayerID = "victim"
	h.handleMessage(attackerConn, ClientMessage{
		Type:     "LEAVE_ROOM",
		PlayerID: "victim", // forged! connection actually belongs to "attacker"
	})

	// Victim must still be in the room
	clients := h.rm.GetClientsByRoom(roomID)
	if _, exists := clients["victim"]; !exists {
		t.Error("victim should still be in room — server must not trust msg.PlayerID")
	}
	// Attacker should have been removed (server used connection-derived identity)
	if _, exists := clients["attacker"]; exists {
		t.Error("attacker should have been removed — server used connection-derived identity")
	}

	// Verify connToRoom is cleaned up for the attacker
	h.mu.RLock()
	_, connExists := h.connToRoom[attackerConn]
	h.mu.RUnlock()
	if connExists {
		t.Error("attackerConn should be removed from connToRoom after leave")
	}
}

func TestKickPlayerRejectsForgedCreatorID(t *testing.T) {
	h := NewHub()

	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})
	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID

	victimConn := NewFakeConnection()
	h.handleMessage(victimConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "victim",
		PlayerName: "Victim",
	})

	attackerConn := NewFakeConnection()
	h.handleMessage(attackerConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "attacker",
		PlayerName: "Attacker",
	})
	attackerConn.ClearMessages()

	h.handleMessage(attackerConn, ClientMessage{
		Type:           MsgKickPlayer,
		PlayerID:       "creator",
		TargetPlayerID: "victim",
	})

	msgs := attackerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response for forged creator kick")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "only room creator can kick players" {
		t.Fatalf("expected creator-only kick error, got %q", errMsg.Error)
	}
	if _, exists := h.rm.GetClientsByRoom(roomID)["victim"]; !exists {
		t.Fatal("victim should remain in room after forged kick attempt")
	}
	if len(victimConn.Messages()) == 0 {
		t.Fatal("expected victim to still have prior join messages")
	}
}

// TestSetStorytellerRejectsUnknownConnection verifies that a connection
// not yet associated with any room gets a clear error when trying to
// set storyteller, not a misleading "no game session" message.
func TestSetStorytellerRejectsUnknownConnection(t *testing.T) {
	h := NewHub()
	conn := NewFakeConnection()

	h.handleMessage(conn, ClientMessage{
		Type:           "SET_STORYTELLER",
		PlayerID:       "anyone",
		TargetPlayerID: "anyone",
	})

	msgs := conn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error message")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok {
		t.Fatalf("expected ServerMessage, got %T", msgs[0])
	}
	if errMsg.Type != "ERROR" {
		t.Errorf("expected ERROR, got %s", errMsg.Type)
	}
	if errMsg.Error != "not in any room" {
		t.Errorf("expected 'not in any room', got %q", errMsg.Error)
	}
}

// TestSetStorytellerUsesConnectionIdentity verifies that the SenderID
// passed to GameSession.Apply for SET_STORYTELLER is derived from the
// connection, not taken from msg.PlayerID.
func TestSetStorytellerUsesConnectionIdentity(t *testing.T) {
	h := NewHub()

	// Create room as "creator"
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID

	// Creator sets themselves as storyteller (uses correct PlayerID)
	h.handleMessage(creatorConn, ClientMessage{
		Type:           "SET_STORYTELLER",
		PlayerID:       "creator",
		TargetPlayerID: "creator",
	})

	// Verify storyteller is set to "creator", not something from msg
	gs := h.sessions[roomID]
	if gs == nil {
		t.Fatal("expected game session")
	}
	if gs.StorytellerID() != "creator" {
		t.Errorf("expected storyteller to be 'creator', got %q", gs.StorytellerID())
	}
}

// TestAssignCharactersRejectsUnknownConnection verifies that an unjoined
// connection gets a clear error instead of the generic no-session path.
func TestAssignCharactersRejectsUnknownConnection(t *testing.T) {
	h := NewHub()
	conn := NewFakeConnection()

	h.handleMessage(conn, ClientMessage{
		Type: "ASSIGN_CHARACTERS",
		Assignments: map[string]string{
			"p1": "imp",
		},
	})

	msgs := conn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error message")
	}
	errMsg := msgs[0].(ServerMessage)
	if errMsg.Type != "ERROR" || errMsg.Error != "not in any room" {
		t.Errorf("expected ERROR not in any room, got %+v", errMsg)
	}
}

// TestSubmitEventRejectsUnknownConnection verifies that an unjoined
// connection gets a clear error instead of silently dropping the event.
func TestSubmitEventRejectsUnknownConnection(t *testing.T) {
	h := NewHub()
	conn := NewFakeConnection()

	h.handleMessage(conn, ClientMessage{
		Type: "SUBMIT_EVENT",
		Event: &game.GameEvent{
			PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay},
		},
	})

	msgs := conn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error message")
	}
	errMsg := msgs[0].(ServerMessage)
	if errMsg.Type != "ERROR" || errMsg.Error != "not in any room" {
		t.Errorf("expected ERROR not in any room, got %+v", errMsg)
	}
}
