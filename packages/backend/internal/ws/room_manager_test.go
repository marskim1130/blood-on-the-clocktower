package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestRoomManagerCreateRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)

	if len(room.id) != 6 {
		t.Errorf("expected 6-digit room ID, got %q", room.id)
	}
	if room.maxPlayers != 10 {
		t.Errorf("expected maxPlayers=10, got %d", room.maxPlayers)
	}
	if room.creatorID != "p1" {
		t.Errorf("expected creatorID=p1, got %s", room.creatorID)
	}
	if room.scriptID != game.TroubleBrewingScriptID {
		t.Errorf("expected default script=%s, got %s", game.TroubleBrewingScriptID, room.scriptID)
	}
}

func TestRoomManagerCreateRoomStoresScript(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10, game.TroubleBrewingScriptID)

	if got := rm.ScriptID(room.id); got != game.TroubleBrewingScriptID {
		t.Errorf("expected scriptID=%s, got %s", game.TroubleBrewingScriptID, got)
	}
}

func TestRoomManagerCreateRoomDefaultMaxPlayers(t *testing.T) {
	rm := NewRoomManager()

	room := rm.CreateRoom("p1", 0)
	if room.maxPlayers != defaultMaxPlayers {
		t.Errorf("expected default maxPlayers=%d, got %d", defaultMaxPlayers, room.maxPlayers)
	}

	room = rm.CreateRoom("p2", 20)
	if room.maxPlayers != defaultMaxPlayers {
		t.Errorf("expected default maxPlayers=%d for out-of-range, got %d", defaultMaxPlayers, room.maxPlayers)
	}
}

func TestRoomManagerJoinRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)
	conn := NewFakeConnection()

	err := rm.JoinRoom(room.id, conn, "p2", "Bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client, rid := rm.GetClient("p2")
	if client == nil {
		t.Fatal("expected to find p2")
	}
	if rid != room.id {
		t.Errorf("expected room %s, got %s", room.id, rid)
	}
	if client.PlayerID != "p2" {
		t.Errorf("expected PlayerID=p2, got %s", client.PlayerID)
	}
}

func TestRoomManagerJoinNonExistentRoom(t *testing.T) {
	rm := NewRoomManager()
	conn := NewFakeConnection()

	err := rm.JoinRoom("999999", conn, "p1", "Alice")
	if err == nil {
		t.Error("expected error for non-existent room")
	}
}

func TestRoomManagerJoinFullRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("storyteller", 5)

	for _, playerID := range []string{"storyteller", "p1", "p2", "p3", "p4", "p5"} {
		conn := NewFakeConnection()
		if err := rm.JoinRoom(room.id, conn, playerID, "Player"); err != nil {
			t.Fatalf("unexpected join error for %s: %v", playerID, err)
		}
	}

	conn := NewFakeConnection()
	err := rm.JoinRoom(room.id, conn, "p6", "Overflow")
	if err == nil {
		t.Error("expected error for full room")
	}
}

func TestRoomManagerLeaveRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)
	conn1 := NewFakeConnection()
	conn2 := NewFakeConnection()

	rm.JoinRoom(room.id, conn1, "p1", "Alice")
	rm.JoinRoom(room.id, conn2, "p2", "Bob")

	roomID, err := rm.LeaveRoom("p2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if roomID != room.id {
		t.Errorf("expected room %s, got %s", room.id, roomID)
	}

	client, _ := rm.GetClient("p2")
	if client != nil {
		t.Error("expected p2 to be removed")
	}
}

func TestRoomManagerLeaveRoomKeepsRoomLifecycleExternal(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)
	conn := NewFakeConnection()
	rm.JoinRoom(room.id, conn, "p1", "Alice")

	rm.LeaveRoom("p1")

	if rm.GetRoom(room.id) == nil {
		t.Error("expected room to remain until the hub destroys it")
	}
}

func TestRoomManagerKickPlayerPreventsRejoin(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("creator", 5)
	conn := NewFakeConnection()
	if err := rm.JoinRoom(room.id, conn, "p1", "Alice"); err != nil {
		t.Fatalf("unexpected join error: %v", err)
	}

	kickedClient, err := rm.KickPlayer(room.id, "p1")
	if err != nil {
		t.Fatalf("unexpected kick error: %v", err)
	}
	if kickedClient == nil || kickedClient.PlayerID != "p1" {
		t.Fatalf("expected kicked client for p1, got %#v", kickedClient)
	}
	if client, _ := rm.GetClient("p1"); client != nil {
		t.Fatal("expected kicked player to be removed from connected clients")
	}

	rejoinConn := NewFakeConnection()
	err = rm.JoinRoom(room.id, rejoinConn, "p1", "Alice")
	if err == nil {
		t.Fatal("expected kicked player rejoin to be rejected")
	}
	if err.Error() != "player was kicked from room" {
		t.Fatalf("expected kicked rejoin error, got %q", err.Error())
	}
}

func TestRoomManagerUpdateRoomSettings(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("creator", 5)

	if err := rm.UpdateRoomSettings(room.id, 6, game.TroubleBrewingScriptID); err != nil {
		t.Fatalf("unexpected update settings error: %v", err)
	}
	if got := rm.MaxPlayers(room.id); got != 6 {
		t.Fatalf("expected maxPlayers=6, got %d", got)
	}
	if got := rm.ScriptID(room.id); got != game.TroubleBrewingScriptID {
		t.Fatalf("expected scriptID=%s, got %s", game.TroubleBrewingScriptID, got)
	}
}

func TestRoomManagerUpdateRoomSettingsRejectsMissingRoom(t *testing.T) {
	rm := NewRoomManager()

	err := rm.UpdateRoomSettings("999999", 6, game.TroubleBrewingScriptID)
	if err == nil {
		t.Fatal("expected missing room update to be rejected")
	}
	if err.Error() != "room not found" {
		t.Fatalf("expected room not found error, got %q", err.Error())
	}
}

func TestRoomManagerDestroyRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)

	rm.DestroyRoom(room.id)

	if rm.GetRoom(room.id) != nil {
		t.Error("expected room to be destroyed explicitly")
	}
}

func TestRoomManagerRemoveClientByConn(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)
	conn := NewFakeConnection()
	rm.JoinRoom(room.id, conn, "p1", "Alice")

	roomID, playerID, err := rm.RemoveClientByConn(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if roomID != room.id {
		t.Errorf("expected room %s, got %s", room.id, roomID)
	}
	if playerID != "p1" {
		t.Errorf("expected playerID=p1, got %s", playerID)
	}
	if rm.GetRoom(room.id) == nil {
		t.Error("expected disconnect cleanup to keep the room available for reconnect")
	}
}

func TestRoomManagerGetClientsByRoom(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)
	conn1 := NewFakeConnection()
	conn2 := NewFakeConnection()

	rm.JoinRoom(room.id, conn1, "p1", "Alice")
	rm.JoinRoom(room.id, conn2, "p2", "Bob")

	clients := rm.GetClientsByRoom(room.id)
	if len(clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(clients))
	}

	if rm.GetClientsByRoom("nonexistent") != nil {
		t.Error("expected nil for non-existent room")
	}
}

func TestRoomManagerPlayerCount(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)

	if rm.PlayerCount(room.id) != 0 {
		t.Error("expected 0 players in empty room")
	}

	conn := NewFakeConnection()
	rm.JoinRoom(room.id, conn, "p1", "Alice")

	if rm.PlayerCount(room.id) != 1 {
		t.Errorf("expected 1 player, got %d", rm.PlayerCount(room.id))
	}
}
