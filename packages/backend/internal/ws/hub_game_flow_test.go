package ws

import (
	"encoding/json"
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestStartGameRejectsNonStoryteller(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupAssignedRoom(t)

	h.handleMessage(playerConns["p1"], ClientMessage{Type: MsgStartGame})

	msgs := playerConns["p1"].Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "only the storyteller can start the game" {
		t.Fatalf("expected storyteller-only error, got %q", errMsg.Error)
	}
	if phase := h.sessions[roomID].Phase(); phase != game.GamePhaseSetup {
		t.Fatalf("expected phase to remain setup, got %d", phase)
	}
	if len(storytellerConn.Messages()) != 0 {
		t.Fatalf("expected no broadcast to storyteller, got %#v", storytellerConn.Messages())
	}
}

func TestExecutePlayerAcceptsLegacyExecutePlayerID(t *testing.T) {
	h, storytellerConn, _, roomID := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:            MsgExecutePlayer,
		ExecutePlayerID: "p1",
	})

	for _, player := range h.sessions[roomID].Players() {
		if player.ID == "p1" && player.IsAlive {
			t.Fatal("expected p1 to be executed when executePlayerId is supplied")
		}
	}
}

func TestSubmitNightActionOnlyNotifiesActorAndStoryteller(t *testing.T) {
	h, storytellerConn, playerConns, _ := setupStartedRoom(t)

	h.handleMessage(playerConns["p1"], ClientMessage{
		Type:       MsgSubmitNightAction,
		ActionType: string(game.NightActionKill),
		TargetIDs:  []string{"p2"},
	})

	if countNightActionSubmitted(storytellerConn.Messages()) != 1 {
		t.Fatalf("expected storyteller to receive night action, got %#v", storytellerConn.Messages())
	}
	if countNightActionSubmitted(playerConns["p1"].Messages()) != 1 {
		t.Fatalf("expected actor to receive night action, got %#v", playerConns["p1"].Messages())
	}
	for _, playerID := range []string{"p2", "p3", "p4", "p5"} {
		if count := countNightActionSubmitted(playerConns[playerID].Messages()); count != 0 {
			t.Fatalf("expected %s not to receive night action details, got %d in %#v", playerID, count, playerConns[playerID].Messages())
		}
	}
}

func TestClientMessagePhaseAcceptsStringAndNumericValues(t *testing.T) {
	var stringMsg ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"CHANGE_PHASE","phase":"day"}`), &stringMsg); err != nil {
		t.Fatalf("unexpected string phase unmarshal error: %v", err)
	}
	if phase := stringMsg.Phase.GamePhase(); phase != game.GamePhaseDay {
		t.Fatalf("expected day phase from string, got %d", phase)
	}

	var numericMsg ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"CHANGE_PHASE","phase":3}`), &numericMsg); err != nil {
		t.Fatalf("unexpected numeric phase unmarshal error: %v", err)
	}
	if phase := numericMsg.Phase.GamePhase(); phase != game.GamePhaseNight {
		t.Fatalf("expected night phase from number, got %d", phase)
	}
}

func setupStartedRoom(t *testing.T) (*Hub, *FakeConnection, map[string]*FakeConnection, string) {
	t.Helper()

	h, storytellerConn, playerConns, roomID := setupAssignedRoom(t)
	h.handleMessage(storytellerConn, ClientMessage{Type: MsgStartGame})
	assertNoErrorMessages(t, storytellerConn.Messages())
	clearAllMessages(storytellerConn, playerConns)

	return h, storytellerConn, playerConns, roomID
}

func setupAssignedRoom(t *testing.T) (*Hub, *FakeConnection, map[string]*FakeConnection, string) {
	t.Helper()

	h := NewHub()
	storytellerConn := NewFakeConnection()
	h.handleMessage(storytellerConn, ClientMessage{
		Type:       MsgCreateRoom,
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
			Type:       MsgJoinRoom,
			RoomID:     roomID,
			PlayerID:   playerID,
			PlayerName: playerID,
		})
	}

	h.handleMessage(storytellerConn, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: "storyteller",
	})
	h.handleMessage(storytellerConn, ClientMessage{
		Type: MsgAssignCharacters,
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "librarian",
			"p3": "investigator",
			"p4": "poisoner",
			"p5": "imp",
		},
	})
	assertNoErrorMessages(t, storytellerConn.Messages())
	clearAllMessages(storytellerConn, playerConns)

	return h, storytellerConn, playerConns, roomID
}

func countNightActionSubmitted(messages []any) int {
	count := 0
	for _, raw := range messages {
		msg, ok := raw.(ServerMessage)
		if !ok || msg.Event == nil || msg.Event.NightActionSubmitted == nil {
			continue
		}
		count++
	}
	return count
}

func assertNoErrorMessages(t *testing.T, messages []any) {
	t.Helper()
	for _, raw := range messages {
		msg, ok := raw.(ServerMessage)
		if ok && msg.Type == "ERROR" {
			t.Fatalf("unexpected error message: %#v", msg)
		}
	}
}

func clearAllMessages(storytellerConn *FakeConnection, playerConns map[string]*FakeConnection) {
	storytellerConn.ClearMessages()
	for _, conn := range playerConns {
		conn.ClearMessages()
	}
}
