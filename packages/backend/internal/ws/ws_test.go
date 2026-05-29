package ws

import (
	"encoding/json"
	"fmt"
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

	// Creator creates room
	creator, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer creator.Close()
	creator.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	creator.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := creator.ReadMessage()
	var createResp ServerMessage
	json.Unmarshal(raw, &createResp)

	// Second player joins
	ws, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws.Close()
	ws.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: createResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var resp ServerMessage
	json.Unmarshal(raw, &resp)

	if resp.Type != "ROOM_STATE" {
		t.Errorf("expected ROOM_STATE, got %s", resp.Type)
	}
	if resp.RoomID != createResp.RoomID {
		t.Errorf("expected %s, got %s", createResp.RoomID, resp.RoomID)
	}
}

func TestBroadcastToRoomMembers(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	// Create room with player 1
	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	// Player 2 joins
	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage() // consume ROOM_STATE

	// Player 1 should receive broadcast about player 2 joining
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("player1 failed to receive broadcast: %v", err)
	}

	var broadcast ServerMessage
	json.Unmarshal(raw, &broadcast)

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

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	// Create room-1 with player 1
	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw1, _ := ws1.ReadMessage()
	var room1Resp ServerMessage
	json.Unmarshal(raw1, &room1Resp)

	// Create room-2 with player 2
	ws2.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage() // consume ROOM_STATE

	// Player 3 joins room-1
	ws3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws3.Close()
	ws3.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: room1Resp.RoomID, PlayerName: "Charlie", PlayerID: "p3"})

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

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	// Create room with player 1
	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	// Player 2 joins
	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
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

func TestCreateRoomGeneratesUniqueID(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws.Close()

	ws.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1", MaxPlayers: 10})
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var resp ServerMessage
	json.Unmarshal(raw, &resp)

	if resp.Type != "ROOM_STATE" {
		t.Errorf("expected ROOM_STATE, got %s", resp.Type)
	}
	if len(resp.RoomID) != 6 {
		t.Errorf("expected 6-digit room ID, got %q (len=%d)", resp.RoomID, len(resp.RoomID))
	}
	if resp.State == nil || len(resp.State.Players) != 1 {
		t.Errorf("expected 1 player in room, got %v", resp.State)
	}
	if resp.State.MaxPlayers != 10 {
		t.Errorf("expected maxPlayers=10, got %d", resp.State.MaxPlayers)
	}
}

func TestJoinNonExistentRoomReturnsError(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws.Close()

	ws.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: "999999", PlayerName: "Alice", PlayerID: "p1"})
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var resp ServerMessage
	json.Unmarshal(raw, &resp)

	if resp.Type != "ERROR" {
		t.Errorf("expected ERROR, got %s", resp.Type)
	}
	if resp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestJoinFullRoomReturnsError(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create room with max 5 players (minimum valid)
	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1", MaxPlayers: 5})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	// Fill room to capacity (5 players)
	for i := 2; i <= 5; i++ {
		ws, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
		defer ws.Close()
		ws.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: fmt.Sprintf("P%d", i), PlayerID: fmt.Sprintf("p%d", i)})
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		ws.ReadMessage() // consume ROOM_STATE
		// Drain broadcast from ws1
		ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
		ws1.ReadMessage()
	}

	// Third player should be rejected
	ws3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws3.Close()
	ws3.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Charlie", PlayerID: "p3"})
	ws3.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws3.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var errResp ServerMessage
	json.Unmarshal(raw, &errResp)

	if errResp.Type != "ERROR" {
		t.Errorf("expected ERROR, got %s (roomId=%s, maxPlayers=%d)", errResp.Type, roomResp.RoomID, roomResp.State.MaxPlayers)
	}
	if errResp.Error != "room is full" {
		t.Errorf("expected 'room is full', got %q", errResp.Error)
	}
}

func TestLeaveRoomRemovesPlayer(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	// Create room with player 1
	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	// Player 2 joins
	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage()
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws1.ReadMessage() // consume PLAYER_JOINED

	// Player 2 leaves
	ws2.WriteJSON(ClientMessage{Type: "LEAVE_ROOM", RoomID: roomResp.RoomID, PlayerID: "p2"})

	// Player 1 should receive PLAYER_LEFT broadcast
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("player1 should receive PLAYER_LEFT: %v", err)
	}

	var broadcast ServerMessage
	json.Unmarshal(raw, &broadcast)

	if broadcast.Type != "EVENT_BROADCAST" {
		t.Errorf("expected EVENT_BROADCAST, got %s", broadcast.Type)
	}
	if broadcast.Event == nil || broadcast.Event.PlayerLeft == nil {
		t.Error("expected playerLeft event")
	}
	if broadcast.Event.PlayerLeft.PlayerID != "p2" {
		t.Errorf("expected playerID p2, got %s", broadcast.Event.PlayerLeft.PlayerID)
	}
}

func TestRoomAutoDestroysWhenEmpty(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create room
	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)
	roomID := roomResp.RoomID
	ws1.Close()

	// Wait for disconnect to propagate
	time.Sleep(100 * time.Millisecond)

	// Try to join the destroyed room
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()
	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var errResp ServerMessage
	json.Unmarshal(raw, &errResp)

	if errResp.Type != "ERROR" {
		t.Errorf("expected ERROR after room destroyed, got %s", errResp.Type)
	}
}
