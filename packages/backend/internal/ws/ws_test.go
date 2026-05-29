package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestWebSocketServerAcceptsConnection(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Connection should be established
	if ws == nil {
		t.Fatal("expected non-nil websocket connection")
	}
}

func TestPlayerCanJoinRoom(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send JOIN_ROOM message
	joinMsg := ClientMessage{
		Type:       "JOIN_ROOM",
		RoomID:     "room-1",
		PlayerName: "Alice",
		PlayerID:   "p1",
	}
	if err := ws.WriteJSON(joinMsg); err != nil {
		t.Fatalf("failed to send join message: %v", err)
	}

	// Read response
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var resp ServerMessage
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Type != "ROOM_STATE" {
		t.Errorf("expected ROOM_STATE, got %s", resp.Type)
	}
	if resp.RoomID != "room-1" {
		t.Errorf("expected room-1, got %s", resp.RoomID)
	}
}

func TestBroadcastToRoomMembers(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect two players to the same room
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("player1 failed to connect: %v", err)
	}
	defer ws1.Close()

	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("player2 failed to connect: %v", err)
	}
	defer ws2.Close()

	// Both join room-1
	join1 := ClientMessage{Type: "JOIN_ROOM", RoomID: "room-1", PlayerName: "Alice", PlayerID: "p1"}
	join2 := ClientMessage{Type: "JOIN_ROOM", RoomID: "room-1", PlayerName: "Bob", PlayerID: "p2"}

	ws1.WriteJSON(join1)
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws1.ReadMessage() // consume ROOM_STATE

	ws2.WriteJSON(join2)
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage() // consume ROOM_STATE

	// Player 1 should receive broadcast about player 2 joining
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("player1 failed to receive broadcast: %v", err)
	}

	var broadcast ServerMessage
	if err := json.Unmarshal(raw, &broadcast); err != nil {
		t.Fatalf("failed to unmarshal broadcast: %v", err)
	}

	if broadcast.Type != "EVENT_BROADCAST" {
		t.Errorf("expected EVENT_BROADCAST, got %s", broadcast.Type)
	}
	if broadcast.Event == nil || broadcast.Event.PlayerJoined == nil {
		t.Error("expected playerJoined event in broadcast")
	}
}

func TestBroadcastOnlyToRoomMembers(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect player to room-1 and player to room-2
	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	ws1.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: "room-1", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws1.ReadMessage() // consume ROOM_STATE

	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: "room-2", PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage() // consume ROOM_STATE

	// Connect a third player to room-1
	ws3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws3.Close()
	ws3.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: "room-1", PlayerName: "Charlie", PlayerID: "p3"})

	// Player 1 (room-1) should receive broadcast about Charlie joining
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("player1 should receive broadcast: %v", err)
	}

	// Player 2 (room-2) should NOT receive anything
	ws2.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err = ws2.ReadMessage()
	if err == nil {
		t.Error("player2 should NOT receive broadcast from room-1")
	}
}

func TestEndToEndEventFlow(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Two players join room-1
	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	ws1.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: "room-1", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws1.ReadMessage() // consume own ROOM_STATE

	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: "room-1", PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage() // consume own ROOM_STATE
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws1.ReadMessage() // consume PLAYER_JOINED broadcast

	// Player 1 submits a phase change event
	phaseEvent := game.GameEvent{
		PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay},
	}
	ws1.WriteJSON(ClientMessage{Type: "SUBMIT_EVENT", Event: &phaseEvent})

	// Both players should receive the broadcast
	for _, ws := range []*websocket.Conn{ws1, ws2} {
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, raw, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("failed to receive event broadcast: %v", err)
		}

		var msg ServerMessage
		json.Unmarshal(raw, &msg)

		if msg.Type != "EVENT_BROADCAST" {
			t.Errorf("expected EVENT_BROADCAST, got %s", msg.Type)
		}
		if msg.Event == nil || msg.Event.PhaseChanged == nil {
			t.Error("expected phaseChanged event")
		}
		if msg.Event.PhaseChanged.Phase != game.GamePhaseDay {
			t.Errorf("expected phase DAY (%d), got %d", game.GamePhaseDay, msg.Event.PhaseChanged.Phase)
		}
	}
}
