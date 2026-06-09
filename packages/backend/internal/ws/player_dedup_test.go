package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// TestAddPlayerDedup verifies that adding the same playerID twice
// does not create duplicate entries in the players slice.
func TestAddPlayerDedup(t *testing.T) {
	gs := NewGameSession()

	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice Updated", IsAlive: true})

	players := gs.Players()
	if len(players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(players))
	}
	if players[0].Name != "Alice Updated" {
		t.Errorf("expected name 'Alice Updated', got %q", players[0].Name)
	}
}

// TestHubJoinRoomDuplicatePlayerID verifies that joining a room twice
// with the same playerID does not create duplicate game session entries.
func TestHubJoinRoomDuplicatePlayerID(t *testing.T) {
	h := NewHub()

	// Create room
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID

	// Join once
	conn1 := NewFakeConnection()
	h.handleMessage(conn1, ClientMessage{
		Type:       "JOIN_ROOM",
		RoomID:     roomID,
		PlayerID:   "p1",
		PlayerName: "Alice",
	})

	// Join again with same playerID (reconnect scenario)
	conn2 := NewFakeConnection()
	h.handleMessage(conn2, ClientMessage{
		Type:       "JOIN_ROOM",
		RoomID:     roomID,
		PlayerID:   "p1",
		PlayerName: "Alice V2",
	})

	// GameSession should have exactly 1 entry for p1
	gs := h.sessions[roomID]
	if gs == nil {
		t.Fatal("expected game session to exist")
	}
	players := gs.Players()
	count := 0
	for _, p := range players {
		if p.ID == "p1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 entry for p1 in game session, got %d (total players: %d)", count, len(players))
	}
}

// TestHandleDisconnectCleansConnToRoom verifies that handleDisconnect
// always cleans up connToRoom, even when RemoveClientByConn returns error.
func TestHandleDisconnectCleansConnToRoom(t *testing.T) {
	h := NewHub()

	// Create a room
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	// Verify connToRoom has the creator's connection
	h.mu.RLock()
	_, exists := h.connToRoom[creatorConn]
	h.mu.RUnlock()
	if !exists {
		t.Fatal("expected creator conn in connToRoom after CREATE_ROOM")
	}

	// Disconnect the creator
	h.handleDisconnect(creatorConn)

	// connToRoom should be cleaned up
	h.mu.RLock()
	_, exists = h.connToRoom[creatorConn]
	h.mu.RUnlock()
	if exists {
		t.Error("expected creator conn to be removed from connToRoom after disconnect")
	}

	// Disconnect an unknown connection — should not panic and should not leak
	unknownConn := NewFakeConnection()
	h.handleDisconnect(unknownConn)

	h.mu.RLock()
	_, exists = h.connToRoom[unknownConn]
	h.mu.RUnlock()
	if exists {
		t.Error("expected unknown conn to not be in connToRoom")
	}
}

// TestCreateRoomRegistersCreatorByConnection verifies that room creation
// registers the creator through the same room membership semantics as join.
func TestCreateRoomRegistersCreatorByConnection(t *testing.T) {
	h := NewHub()
	conn := NewFakeConnection()

	h.handleMessage(conn, ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 10,
	})

	roomID := conn.Messages()[0].(ServerMessage).RoomID
	playerID := h.rm.GetPlayerByConn(roomID, conn)
	if playerID != "creator" {
		t.Fatalf("expected creator registered by connection, got %q", playerID)
	}
}
