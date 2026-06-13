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

func TestSubmitEventFromRoomMemberIsRejected(t *testing.T) {
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

	// Player 1 attempts to submit a raw phase change event. The server must stay
	// authoritative and reject raw client-provided GameEvent payloads.
	phaseEvent := game.GameEvent{
		PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay},
	}
	ws1.WriteJSON(ClientMessage{Type: "SUBMIT_EVENT", Event: &phaseEvent})

	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("failed to receive submit event rejection: %v", err)
	}

	var msg ServerMessage
	json.Unmarshal(raw, &msg)
	if msg.Type != "ERROR" {
		t.Fatalf("expected ERROR, got %s", msg.Type)
	}
	if msg.Error != "raw event submission is disabled; use explicit game commands" {
		t.Fatalf("expected raw submission rejection, got %q", msg.Error)
	}

	ws2.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	_, _, err = ws2.ReadMessage()
	if err == nil {
		t.Fatal("expected no raw event broadcast to other room members")
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

	// Fill room to capacity: maxPlayers counts actual players, plus one storyteller seat.
	for i := 2; i <= 6; i++ {
		ws, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
		defer ws.Close()
		ws.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: fmt.Sprintf("P%d", i), PlayerID: fmt.Sprintf("p%d", i)})
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		ws.ReadMessage() // consume ROOM_STATE
		// Drain broadcast from ws1
		ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
		ws1.ReadMessage()
	}

	// Seventh connection should be rejected.
	ws3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws3.Close()
	ws3.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Overflow", PlayerID: "p7"})
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

func TestRoomSurvivesDisconnectForReconnect(t *testing.T) {
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

	// Reconnect with the same player identity.
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()
	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomID, PlayerName: "Alice Again", PlayerID: "p1"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var stateResp ServerMessage
	json.Unmarshal(raw, &stateResp)

	if stateResp.Type != "ROOM_STATE" {
		t.Errorf("expected ROOM_STATE after reconnect, got %s", stateResp.Type)
	}
	if stateResp.RoomID != roomID {
		t.Errorf("expected room %s after reconnect, got %s", roomID, stateResp.RoomID)
	}
}

func TestWebSocketCompleteMVPGameFlow(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	storyteller := dialTestWebSocket(t, wsURL)
	defer storyteller.Close()

	storyteller.WriteJSON(ClientMessage{
		Type:       "CREATE_ROOM",
		PlayerName: "Storyteller",
		PlayerID:   "storyteller",
		MaxPlayers: 5,
	})
	createResp := readNextServerMessage(t, storyteller)
	if createResp.Type != "ROOM_STATE" || createResp.RoomID == "" {
		t.Fatalf("expected initial ROOM_STATE with room id, got %#v", createResp)
	}
	roomID := createResp.RoomID

	playerSockets := map[string]*websocket.Conn{}
	for _, playerID := range []string{"p1", "p2", "p3", "p4", "p5"} {
		conn := dialTestWebSocket(t, wsURL)
		defer conn.Close()
		playerSockets[playerID] = conn
		conn.WriteJSON(ClientMessage{
			Type:       "JOIN_ROOM",
			RoomID:     roomID,
			PlayerName: playerID,
			PlayerID:   playerID,
		})
		joinResp := readNextServerMessage(t, conn)
		if joinResp.Type != "ROOM_STATE" || joinResp.RoomID != roomID {
			t.Fatalf("expected %s to receive room state, got %#v", playerID, joinResp)
		}
	}

	storyteller.WriteJSON(ClientMessage{
		Type:           "SET_STORYTELLER",
		TargetPlayerID: "storyteller",
	})
	stateMsg := readUntilServerMessage(t, storyteller, func(msg ServerMessage) bool {
		return msg.Type == "ROOM_STATE" &&
			msg.State != nil &&
			msg.State.StorytellerID == "storyteller" &&
			len(msg.State.Players) == 5
	})
	if stateMsg.State == nil {
		t.Fatal("expected storyteller room state after setting storyteller")
	}

	storyteller.WriteJSON(ClientMessage{
		Type: "ASSIGN_CHARACTERS",
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "librarian",
			"p3": "investigator",
			"p4": "poisoner",
			"p5": "imp",
		},
	})
	stateMsg = readUntilServerMessage(t, storyteller, func(msg ServerMessage) bool {
		return msg.Type == "ROOM_STATE" && msg.State != nil && allPlayersHaveCharacters(msg.State)
	})
	if stateMsg.State == nil {
		t.Fatal("expected character assignment snapshot")
	}

	storyteller.WriteJSON(ClientMessage{Type: "START_GAME"})
	stateMsg = readUntilServerMessage(t, storyteller, func(msg ServerMessage) bool {
		return msg.Type == "ROOM_STATE" && msg.State != nil && msg.State.Phase == game.GamePhaseNight
	})
	if stateMsg.State == nil || stateMsg.State.Phase != game.GamePhaseNight {
		t.Fatalf("expected night phase after start, got %#v", stateMsg)
	}

	writeStorytellerFirstNightActions(t, storyteller)

	storyteller.WriteJSON(ClientMessage{Type: "RESOLVE_NIGHT"})
	stateMsg = readUntilServerMessage(t, storyteller, func(msg ServerMessage) bool {
		return msg.Type == "ROOM_STATE" &&
			msg.State != nil &&
			msg.State.Phase == game.GamePhaseDay &&
			playerAliveInState(msg.State, "p1")
	})
	if stateMsg.State == nil || stateMsg.State.Winner != nil {
		t.Fatalf("expected game to continue after first night, got %#v", stateMsg.State)
	}
	if containsString(stateMsg.State.GhostVotesRemaining, "p1") {
		t.Fatalf("expected no p1 ghost vote while alive, got %#v", stateMsg.State.GhostVotesRemaining)
	}

	playerSockets["p2"].WriteJSON(ClientMessage{
		Type:      "NOMINATE",
		NomineeID: "p5",
	})
	stateMsg = readUntilServerMessage(t, storyteller, func(msg ServerMessage) bool {
		return msg.Type == "ROOM_STATE" &&
			msg.State != nil &&
			msg.State.Phase == game.GamePhaseVoting &&
			msg.State.Nomination != nil &&
			msg.State.Nomination.NomineeID == "p5"
	})
	if stateMsg.State == nil || stateMsg.State.Nomination == nil {
		t.Fatalf("expected active nomination snapshot, got %#v", stateMsg)
	}

	for _, voterID := range []string{"p1", "p2", "p3"} {
		yes := true
		playerSockets[voterID].WriteJSON(ClientMessage{
			Type:     "CAST_VOTE",
			Decision: &yes,
		})
	}
	stateMsg = readUntilServerMessage(t, storyteller, func(msg ServerMessage) bool {
		return msg.Type == "ROOM_STATE" &&
			msg.State != nil &&
			msg.State.Nomination != nil &&
			len(msg.State.Nomination.Votes) == 3
	})
	if stateMsg.State == nil || stateMsg.State.Nomination == nil {
		t.Fatalf("expected three recorded votes before resolving nomination, got %#v", stateMsg)
	}

	storyteller.WriteJSON(ClientMessage{Type: "RESOLVE_NOMINATION"})
	stateMsg = readUntilServerMessage(t, storyteller, func(msg ServerMessage) bool {
		return msg.Type == "ROOM_STATE" &&
			msg.State != nil &&
			msg.State.Phase == game.GamePhaseFinished &&
			msg.State.Winner != nil
	})
	if stateMsg.State == nil || stateMsg.State.Winner == nil {
		t.Fatalf("expected finished game snapshot, got %#v", stateMsg)
	}
	if stateMsg.State.Winner.Winner != game.TeamGood || stateMsg.State.Winner.Reason != game.WinReasonImpExecuted {
		t.Fatalf("expected good win by Imp execution, got %#v", stateMsg.State.Winner)
	}
	if playerAliveInState(stateMsg.State, "p5") {
		t.Fatal("expected p5 Imp to be dead after execution")
	}
	if containsString(stateMsg.State.GhostVotesRemaining, "p1") {
		t.Fatalf("expected p1 ghost vote to be spent, got %#v", stateMsg.State.GhostVotesRemaining)
	}
}

// --- Storyteller tests (Issue #4) ---

func TestSetStoryteller(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	// Create room, player 1 is creator
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
	ws1.ReadMessage() // PLAYER_JOINED

	// Creator sets player 2 as Storyteller
	ws1.WriteJSON(ClientMessage{Type: "SET_STORYTELLER", RoomID: roomResp.RoomID, PlayerID: "p1", TargetPlayerID: "p2"})

	// Both should receive ROOM_STATE update with storytellerId
	for _, ws := range []*websocket.Conn{ws1, ws2} {
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, raw, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read response: %v", err)
		}
		var msg ServerMessage
		json.Unmarshal(raw, &msg)
		if msg.Type != "ROOM_STATE" {
			t.Errorf("expected ROOM_STATE, got %s", msg.Type)
		}
		if msg.State == nil || msg.State.StorytellerID != "p2" {
			t.Errorf("expected storytellerId=p2, got %v", msg.State)
		}
	}
}

func TestStorytellerRemovedFromPlayerList(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage()
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws1.ReadMessage()

	// Set player 2 as Storyteller
	ws1.WriteJSON(ClientMessage{Type: "SET_STORYTELLER", RoomID: roomResp.RoomID, PlayerID: "p1", TargetPlayerID: "p2"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ = ws1.ReadMessage()
	var stateMsg ServerMessage
	json.Unmarshal(raw, &stateMsg)

	// Storyteller (p2) should not be in player list
	if stateMsg.State == nil {
		t.Fatal("expected non-nil state")
	}
	for _, p := range stateMsg.State.Players {
		if p.ID == "p2" {
			t.Error("storyteller should be removed from player list")
		}
	}
}

func TestOnlyOneStoryteller(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()
	ws3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws3.Close()

	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage()
	ws1.ReadMessage() // broadcast

	ws3.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Charlie", PlayerID: "p3"})
	ws3.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws3.ReadMessage()
	ws1.ReadMessage() // broadcast
	ws2.ReadMessage() // broadcast

	// Set p2 as Storyteller
	ws1.WriteJSON(ClientMessage{Type: "SET_STORYTELLER", RoomID: roomResp.RoomID, PlayerID: "p1", TargetPlayerID: "p2"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws1.ReadMessage() // ROOM_STATE
	ws2.ReadMessage() // broadcast
	ws3.ReadMessage() // broadcast

	// Try to set p3 as Storyteller (should fail)
	ws1.WriteJSON(ClientMessage{Type: "SET_STORYTELLER", RoomID: roomResp.RoomID, PlayerID: "p1", TargetPlayerID: "p3"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	var errResp ServerMessage
	json.Unmarshal(raw, &errResp)
	if errResp.Type != "ERROR" {
		t.Errorf("expected ERROR, got %s", errResp.Type)
	}
}

func TestNonStorytellerCannotPerformStorytellerActions(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1"})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage()
	ws1.ReadMessage()

	// Non-storyteller player tries to set storyteller (should fail, only creator can)
	ws2.WriteJSON(ClientMessage{Type: "SET_STORYTELLER", RoomID: roomResp.RoomID, PlayerID: "p2", TargetPlayerID: "p1"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	var errResp ServerMessage
	json.Unmarshal(raw, &errResp)
	if errResp.Type != "ERROR" {
		t.Errorf("expected ERROR for non-creator, got %s", errResp.Type)
	}
}

// --- Character assignment tests (Issue #5) ---

func TestStorytellerCanAssignCharacters(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()
	ws3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws3.Close()
	ws4, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws4.Close()
	ws5, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws5.Close()
	ws6, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws6.Close()

	// Create room and join one storyteller plus 5 actual players.
	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1", MaxPlayers: 5})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	connected := []*websocket.Conn{ws1}
	for i, ws := range []*websocket.Conn{ws2, ws3, ws4, ws5, ws6} {
		ws.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: fmt.Sprintf("P%d", i+2), PlayerID: fmt.Sprintf("p%d", i+2)})
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		ws.ReadMessage() // ROOM_STATE for new player
		// Drain PLAYER_JOINED broadcast from all existing connected clients
		time.Sleep(20 * time.Millisecond)
		for _, existing := range connected {
			existing.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			existing.ReadMessage()
		}
		connected = append(connected, ws)
	}

	// Set p1 as storyteller.
	ws1.WriteJSON(ClientMessage{Type: "SET_STORYTELLER", RoomID: roomResp.RoomID, PlayerID: "p1", TargetPlayerID: "p1"})
	// Read ROOM_STATE from all clients.
	for _, ws := range []*websocket.Conn{ws1, ws2, ws3, ws4, ws5, ws6} {
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, raw, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read storyteller broadcast: %v", err)
		}
		var msg ServerMessage
		json.Unmarshal(raw, &msg)
		if msg.Type != "ROOM_STATE" {
			t.Errorf("expected ROOM_STATE, got %s", msg.Type)
		}
	}

	// Storyteller assigns characters to 5 actual players.
	assignments := map[string]string{
		"p2": "washerwoman",
		"p3": "librarian",
		"p4": "investigator",
		"p5": "poisoner",
		"p6": "imp",
	}
	ws1.WriteJSON(ClientMessage{Type: "ASSIGN_CHARACTERS", RoomID: roomResp.RoomID, PlayerID: "p1", Assignments: assignments})

	// Storyteller receives every private assignment plus a complete ROOM_STATE.
	for i := 0; i < 6; i++ {
		ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, raw, err := ws1.ReadMessage()
		if err != nil {
			t.Fatalf("storyteller failed to read character assignment/state %d: %v", i, err)
		}
		var msg ServerMessage
		json.Unmarshal(raw, &msg)
		if msg.Type != "EVENT_BROADCAST" && msg.Type != "ROOM_STATE" {
			t.Errorf("expected EVENT_BROADCAST or ROOM_STATE, got %s", msg.Type)
		}
	}

	// Each actual player receives only their own private assignment plus their own ROOM_STATE.
	for _, ws := range []*websocket.Conn{ws2, ws3, ws4, ws5, ws6} {
		for i := 0; i < 2; i++ {
			ws.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, raw, err := ws.ReadMessage()
			if err != nil {
				t.Fatalf("player failed to read private assignment/state %d: %v", i, err)
			}
			var msg ServerMessage
			json.Unmarshal(raw, &msg)
			if msg.Type != "EVENT_BROADCAST" && msg.Type != "ROOM_STATE" {
				t.Errorf("expected EVENT_BROADCAST or ROOM_STATE, got %s", msg.Type)
			}
		}
	}
}

func TestInvalidAssignmentRejected(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()

	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1", MaxPlayers: 5})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage()
	ws1.ReadMessage()

	// Invalid: 2 players but trying to assign 5-player distribution
	assignments := map[string]string{
		"p1": "imp",
		"p2": "imp",
	}
	ws1.WriteJSON(ClientMessage{Type: "ASSIGN_CHARACTERS", RoomID: roomResp.RoomID, PlayerID: "p1", Assignments: assignments})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	var errResp ServerMessage
	json.Unmarshal(raw, &errResp)
	if errResp.Type != "ERROR" {
		t.Errorf("expected ERROR for invalid assignment, got %s", errResp.Type)
	}
}

func TestNonStorytellerCannotAssignCharacters(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws1.Close()
	ws2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws2.Close()
	ws3, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer ws3.Close()

	ws1.WriteJSON(ClientMessage{Type: "CREATE_ROOM", PlayerName: "Alice", PlayerID: "p1", MaxPlayers: 5})
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, _ := ws1.ReadMessage()
	var roomResp ServerMessage
	json.Unmarshal(raw, &roomResp)

	ws2.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Bob", PlayerID: "p2"})
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws2.ReadMessage()
	ws1.ReadMessage()

	ws3.WriteJSON(ClientMessage{Type: "JOIN_ROOM", RoomID: roomResp.RoomID, PlayerName: "Charlie", PlayerID: "p3"})
	ws3.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws3.ReadMessage()
	ws1.ReadMessage()
	ws2.ReadMessage()

	// Set p2 as storyteller
	ws1.WriteJSON(ClientMessage{Type: "SET_STORYTELLER", RoomID: roomResp.RoomID, PlayerID: "p1", TargetPlayerID: "p2"})
	ws1.ReadMessage() // ROOM_STATE
	ws2.ReadMessage()
	ws3.ReadMessage()

	// Non-storyteller (p3) tries to assign characters
	ws3.WriteJSON(ClientMessage{Type: "ASSIGN_CHARACTERS", RoomID: roomResp.RoomID, PlayerID: "p3", Assignments: map[string]string{"p1": "imp", "p3": "washerwoman"}})
	ws3.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := ws3.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	var errResp ServerMessage
	json.Unmarshal(raw, &errResp)
	if errResp.Type != "ERROR" {
		t.Errorf("expected ERROR for non-storyteller, got %s", errResp.Type)
	}
}

func dialTestWebSocket(t *testing.T, wsURL string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect websocket: %v", err)
	}
	return conn
}

func readNextServerMessage(t *testing.T, conn *websocket.Conn) ServerMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	var msg ServerMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("failed to unmarshal server message %s: %v", string(raw), err)
	}
	return msg
}

func readUntilServerMessage(t *testing.T, conn *websocket.Conn, matches func(ServerMessage) bool) ServerMessage {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	conn.SetReadDeadline(deadline)
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("timed out waiting for matching server message: %v", err)
		}

		var msg ServerMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("failed to unmarshal server message %s: %v", string(raw), err)
		}
		if matches(msg) {
			return msg
		}
	}
}

func writeStorytellerFirstNightActions(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	actions := []ClientMessage{
		{Type: "SUBMIT_NIGHT_ACTION", ActionType: string(game.NightActionLearnDemon)},
		{Type: "SUBMIT_NIGHT_ACTION", ActionType: string(game.NightActionLearnMinion)},
		{Type: "SUBMIT_NIGHT_ACTION", ActionType: string(game.NightActionPoison), TargetIDs: []string{"p2"}},
		{Type: "SUBMIT_NIGHT_ACTION", ActionType: string(game.NightActionLearnTownsfolk), TargetIDs: []string{"p1", "p2"}},
		{Type: "SUBMIT_NIGHT_ACTION", ActionType: string(game.NightActionLearnOutsider)},
		{Type: "SUBMIT_NIGHT_ACTION", ActionType: string(game.NightActionLearnMinion), TargetIDs: []string{"p4", "p5"}},
	}
	for _, action := range actions {
		if err := conn.WriteJSON(action); err != nil {
			t.Fatalf("failed to write night action %s: %v", action.ActionType, err)
		}
	}
}

func allPlayersHaveCharacters(state *RoomState) bool {
	if state == nil || len(state.Players) == 0 {
		return false
	}
	for _, player := range state.Players {
		if player.Character == nil {
			return false
		}
	}
	return true
}

func playerAliveInState(state *RoomState, playerID string) bool {
	if state == nil {
		return false
	}
	for _, player := range state.Players {
		if player.ID == playerID {
			return player.IsAlive
		}
	}
	return false
}
