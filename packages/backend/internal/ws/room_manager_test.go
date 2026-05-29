package ws

import (
	"testing"
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
	room := rm.CreateRoom("p1", 5)

	for i := 0; i < 5; i++ {
		conn := NewFakeConnection()
		rm.JoinRoom(room.id, conn, "p"+string(rune('1'+i)), "Player")
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

func TestRoomManagerLeaveRoomAutoDestroy(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("p1", 10)
	conn := NewFakeConnection()
	rm.JoinRoom(room.id, conn, "p1", "Alice")

	rm.LeaveRoom("p1")

	if rm.GetRoom(room.id) != nil {
		t.Error("expected room to be destroyed after last player left")
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
