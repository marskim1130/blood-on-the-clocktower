package ws

import (
	"context"
	"testing"
	"time"
)

func TestInactiveRoomCleanupNotifiesAndDetachesOnlineClients(t *testing.T) {
	hub := NewHub()
	creator := newFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "expiry-create", PlayerID: "creator", PlayerName: "Creator", MaxPlayers: 5, ScriptID: "trouble_brewing"})
	created := lastServerMessage(t, creator)
	creator.ClearMessages()
	hub.CleanupInactiveRooms(context.Background(), time.Now().Add(8*24*time.Hour))
	closed := lastServerMessage(t, creator)
	if closed.Type != ServerMsgRoomClosed || closed.RoomID != created.RoomID {
		t.Fatalf("expiry notification = %+v", closed)
	}
	if _, exists := hub.registry.Get(created.RoomID); exists {
		t.Fatal("expired room still registered")
	}
	if len(hub.active.Room(created.RoomID)) != 0 {
		t.Fatal("expired room connections still active")
	}
}
