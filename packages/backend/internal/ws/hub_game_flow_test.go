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
		Result:     "P2 is marked by the actor.",
	})

	if countNightActionSubmitted(storytellerConn.Messages()) != 1 {
		t.Fatalf("expected storyteller to receive night action, got %#v", storytellerConn.Messages())
	}
	storytellerAction := lastNightActionSubmitted(t, storytellerConn.Messages())
	if storytellerAction.Result == nil || *storytellerAction.Result != "P2 is marked by the actor." {
		t.Fatalf("expected storyteller to receive night action result, got %#v", storytellerAction)
	}
	if countNightActionSubmitted(playerConns["p1"].Messages()) != 1 {
		t.Fatalf("expected actor to receive night action, got %#v", playerConns["p1"].Messages())
	}
	actorAction := lastNightActionSubmitted(t, playerConns["p1"].Messages())
	if actorAction.Result == nil || *actorAction.Result != "P2 is marked by the actor." {
		t.Fatalf("expected actor to receive night action result, got %#v", actorAction)
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

func TestClientMessageWinnerAcceptsStringAndNumericValues(t *testing.T) {
	var stringMsg ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"END_GAME","winner":"evil"}`), &stringMsg); err != nil {
		t.Fatalf("unexpected string winner unmarshal error: %v", err)
	}
	if winner := stringMsg.Winner.Team(); winner != game.TeamEvil {
		t.Fatalf("expected evil winner from string, got %d", winner)
	}

	var numericMsg ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"END_GAME","winner":1}`), &numericMsg); err != nil {
		t.Fatalf("unexpected numeric winner unmarshal error: %v", err)
	}
	if winner := numericMsg.Winner.Team(); winner != game.TeamGood {
		t.Fatalf("expected good winner from number, got %d", winner)
	}
}

func TestClientMessageDeathCauseAcceptsStringAndNumericValues(t *testing.T) {
	var stringMsg ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"KILL_PLAYER","cause":"night_kill"}`), &stringMsg); err != nil {
		t.Fatalf("unexpected string death cause unmarshal error: %v", err)
	}
	if cause := stringMsg.Cause.DeathCause(); cause != game.DeathCauseNightKill {
		t.Fatalf("expected night kill cause from string, got %q", cause)
	}

	var numericMsg ClientMessage
	if err := json.Unmarshal([]byte(`{"type":"KILL_PLAYER","cause":3}`), &numericMsg); err != nil {
		t.Fatalf("unexpected numeric death cause unmarshal error: %v", err)
	}
	if cause := numericMsg.Cause.DeathCause(); cause != game.DeathCauseAbility {
		t.Fatalf("expected ability cause from number, got %q", cause)
	}
}

func TestCreateRoomRejectsUnsupportedScript(t *testing.T) {
	h := NewHub()
	conn := NewFakeConnection()

	h.handleMessage(conn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "storyteller",
		PlayerName: "Storyteller",
		MaxPlayers: 5,
		ScriptID:   "bad_moon_rising",
	})

	msgs := conn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "unsupported script" {
		t.Fatalf("expected unsupported script error, got %q", errMsg.Error)
	}
}

func TestRoomStateIncludesScriptMetadata(t *testing.T) {
	h := NewHub()
	conn := NewFakeConnection()

	h.handleMessage(conn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "storyteller",
		PlayerName: "Storyteller",
		MaxPlayers: 5,
		ScriptID:   game.TroubleBrewingScriptID,
	})

	msgs := conn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected room state response")
	}
	stateMsg, ok := msgs[0].(ServerMessage)
	if !ok || stateMsg.Type != "ROOM_STATE" || stateMsg.State == nil {
		t.Fatalf("expected ROOM_STATE message, got %#v", msgs[0])
	}
	if stateMsg.State.ScriptID != game.TroubleBrewingScriptID {
		t.Fatalf("expected scriptID=%s, got %s", game.TroubleBrewingScriptID, stateMsg.State.ScriptID)
	}
	if stateMsg.State.ScriptName != "Trouble Brewing" {
		t.Fatalf("expected Trouble Brewing script name, got %q", stateMsg.State.ScriptName)
	}
}

func TestRoomCreatorCanUpdateRoomSettingsDuringSetup(t *testing.T) {
	h := NewHub()
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 5,
	})
	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID

	playerConn := NewFakeConnection()
	h.handleMessage(playerConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "p1",
		PlayerName: "P1",
	})
	creatorConn.ClearMessages()
	playerConn.ClearMessages()

	h.handleMessage(creatorConn, ClientMessage{
		Type:       MsgUpdateRoomSettings,
		MaxPlayers: 6,
		ScriptID:   game.TroubleBrewingScriptID,
	})
	assertNoErrorMessages(t, creatorConn.Messages())

	if got := h.rm.MaxPlayers(roomID); got != 6 {
		t.Fatalf("expected room maxPlayers=6, got %d", got)
	}
	creatorState := lastRoomState(t, creatorConn.Messages())
	if creatorState.MaxPlayers != 6 {
		t.Fatalf("expected creator room state maxPlayers=6, got %d", creatorState.MaxPlayers)
	}
	if creatorState.ScriptID != game.TroubleBrewingScriptID {
		t.Fatalf("expected scriptID=%s, got %s", game.TroubleBrewingScriptID, creatorState.ScriptID)
	}

	playerState := lastRoomState(t, playerConn.Messages())
	if playerState.MaxPlayers != 6 {
		t.Fatalf("expected player room state maxPlayers=6, got %d", playerState.MaxPlayers)
	}
}

func TestUpdateRoomSettingsRejectsNonCreator(t *testing.T) {
	h := NewHub()
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 5,
	})
	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID

	playerConn := NewFakeConnection()
	h.handleMessage(playerConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "p1",
		PlayerName: "P1",
	})
	playerConn.ClearMessages()

	h.handleMessage(playerConn, ClientMessage{
		Type:       MsgUpdateRoomSettings,
		MaxPlayers: 6,
	})

	msgs := playerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected non-creator settings update error")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "only room creator can update room settings" {
		t.Fatalf("expected creator-only settings error, got %q", errMsg.Error)
	}
	if got := h.rm.MaxPlayers(roomID); got != 5 {
		t.Fatalf("expected maxPlayers to remain 5, got %d", got)
	}
}

func TestRoomCreatorCanKickPlayerDuringSetup(t *testing.T) {
	h := NewHub()
	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 5,
	})
	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID

	victimConn := NewFakeConnection()
	h.handleMessage(victimConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "victim",
		PlayerName: "Victim",
	})
	bystanderConn := NewFakeConnection()
	h.handleMessage(bystanderConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "bystander",
		PlayerName: "Bystander",
	})
	creatorConn.ClearMessages()
	victimConn.ClearMessages()
	bystanderConn.ClearMessages()

	h.handleMessage(creatorConn, ClientMessage{
		Type:           MsgKickPlayer,
		TargetPlayerID: "victim",
	})
	assertNoErrorMessages(t, creatorConn.Messages())

	victimMessages := victimConn.Messages()
	if len(victimMessages) == 0 {
		t.Fatal("expected kicked player notification")
	}
	kickedMsg, ok := victimMessages[0].(ServerMessage)
	if !ok || kickedMsg.Type != "ERROR" || kickedMsg.Error != "kicked from room" {
		t.Fatalf("expected kicked error for victim, got %#v", victimMessages)
	}

	if _, exists := h.rm.GetClientsByRoom(roomID)["victim"]; exists {
		t.Fatal("expected victim to be removed from room clients")
	}
	h.mu.RLock()
	_, mapped := h.connToRoom[victimConn]
	h.mu.RUnlock()
	if mapped {
		t.Fatal("expected victim connection mapping to be removed")
	}
	for _, player := range h.sessions[roomID].Players() {
		if player.ID == "victim" {
			t.Fatal("expected victim to be removed from game session")
		}
	}

	state := lastRoomState(t, bystanderConn.Messages())
	if findOptionalPlayerInState(state, "victim") != nil {
		t.Fatalf("expected broadcast room state without victim, got %#v", state.Players)
	}

	rejoinConn := NewFakeConnection()
	h.handleMessage(rejoinConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "victim",
		PlayerName: "Victim",
	})
	msgs := rejoinConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected kicked rejoin error")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" || errMsg.Error != "player was kicked from room" {
		t.Fatalf("expected kicked rejoin error, got %#v", msgs[0])
	}
}

func TestKickPlayerRejectsAfterGameStart(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:           MsgKickPlayer,
		TargetPlayerID: "p1",
	})

	msgs := storytellerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "players can only be kicked during setup phase" {
		t.Fatalf("expected setup-only kick error, got %q", errMsg.Error)
	}
	if findOptionalPlayerInState(h.buildRoomStateForRecipient(roomID, "storyteller"), "p1") == nil {
		t.Fatal("expected p1 to remain in game after rejected kick")
	}
	if len(playerConns["p1"].Messages()) != 0 {
		t.Fatalf("expected kicked target to receive no message after rejected kick, got %#v", playerConns["p1"].Messages())
	}
}

func TestUpdateRoomSettingsRejectsAfterGameStart(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:       MsgUpdateRoomSettings,
		MaxPlayers: 6,
	})

	msgs := storytellerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected setup-only settings error")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "room settings can only be updated during setup phase" {
		t.Fatalf("expected setup-only settings error, got %q", errMsg.Error)
	}
	if got := h.rm.MaxPlayers(roomID); got != 5 {
		t.Fatalf("expected maxPlayers to remain 5, got %d", got)
	}
	for playerID, conn := range playerConns {
		if len(conn.Messages()) != 0 {
			t.Fatalf("expected %s to receive no settings update broadcast, got %#v", playerID, conn.Messages())
		}
	}
}

func TestDisconnectReconnectPreservesPlayerSnapshot(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:           MsgExecutePlayer,
		TargetPlayerID: "p1",
	})
	assertNoErrorMessages(t, storytellerConn.Messages())
	clearAllMessages(storytellerConn, playerConns)

	h.handleDisconnect(playerConns["p1"])
	for playerID, conn := range playerConns {
		if playerID == "p1" {
			continue
		}
		if len(conn.Messages()) != 0 {
			t.Fatalf("expected disconnect not to broadcast playerLeft to %s, got %#v", playerID, conn.Messages())
		}
	}

	reconnectConn := NewFakeConnection()
	h.handleMessage(reconnectConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "p1",
		PlayerName: "p1 reconnected",
	})

	msgs := reconnectConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected room state after reconnect")
	}
	stateMsg, ok := msgs[0].(ServerMessage)
	if !ok || stateMsg.Type != "ROOM_STATE" || stateMsg.State == nil {
		t.Fatalf("expected ROOM_STATE after reconnect, got %#v", msgs[0])
	}

	var reconnected game.Player
	found := false
	for _, player := range stateMsg.State.Players {
		if player.ID == "p1" {
			reconnected = player
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected p1 in reconnect snapshot")
	}
	if reconnected.IsAlive {
		t.Fatal("expected p1 death state to survive reconnect")
	}
	if reconnected.Character == nil || reconnected.Character.ID != "washerwoman" {
		t.Fatalf("expected p1 character to survive reconnect, got %#v", reconnected.Character)
	}
	if len(stateMsg.State.Deaths) != 1 || stateMsg.State.Deaths[0].PlayerID != "p1" {
		t.Fatalf("expected p1 death record in reconnect snapshot, got %#v", stateMsg.State.Deaths)
	}

	count := 0
	for _, player := range h.sessions[roomID].Players() {
		if player.ID == "p1" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected one p1 participant after reconnect, got %d", count)
	}
}

func TestStorytellerReconnectDoesNotBecomePlayer(t *testing.T) {
	h, storytellerConn, _, roomID := setupAssignedRoom(t)

	h.handleDisconnect(storytellerConn)

	reconnectConn := NewFakeConnection()
	h.handleMessage(reconnectConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "storyteller",
		PlayerName: "Storyteller Again",
	})

	if h.sessions[roomID].StorytellerID() != "storyteller" {
		t.Fatalf("expected storyteller identity to survive reconnect")
	}
	for _, player := range h.sessions[roomID].Players() {
		if player.ID == "storyteller" {
			t.Fatal("storyteller reconnect should not add storyteller as a player")
		}
	}

	msgs := reconnectConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected room state after storyteller reconnect")
	}
	stateMsg, ok := msgs[0].(ServerMessage)
	if !ok || stateMsg.Type != "ROOM_STATE" || stateMsg.State == nil {
		t.Fatalf("expected ROOM_STATE after storyteller reconnect, got %#v", msgs[0])
	}
	for _, player := range stateMsg.State.Players {
		if player.Character == nil {
			t.Fatalf("expected storyteller reconnect snapshot to reveal all characters, got %#v", stateMsg.State.Players)
		}
	}
	if len(storytellerConn.Messages()) != 0 {
		t.Fatalf("expected disconnected storyteller connection to receive no messages, got %#v", storytellerConn.Messages())
	}
}

func TestCompleteMVPGameFlowFromFirstNightToGoodWin(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupAssignedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{Type: MsgStartGame})
	assertNoErrorMessages(t, storytellerConn.Messages())
	if phase := h.sessions[roomID].Phase(); phase != game.GamePhaseNight {
		t.Fatalf("expected game to start at night, got %d", phase)
	}
	stateAtNight := h.buildRoomStateForRecipient(roomID, "storyteller")
	if len(stateAtNight.NightWakeSteps) != 5 {
		t.Fatalf("expected 5 active first-night wake steps, got %#v", stateAtNight.NightWakeSteps)
	}
	if stateAtNight.CurrentNightWakeStep == nil || stateAtNight.CurrentNightWakeStep.ActionType != game.NightActionPoison {
		t.Fatalf("expected Poisoner to be current wake step, got %#v", stateAtNight.CurrentNightWakeStep)
	}
	clearAllMessages(storytellerConn, playerConns)

	submitStorytellerFirstNightActions(t, h, storytellerConn)

	h.handleMessage(storytellerConn, ClientMessage{Type: MsgResolveNight})
	assertNoErrorMessages(t, storytellerConn.Messages())

	stateAfterNight := h.buildRoomStateForRecipient(roomID, "storyteller")
	if stateAfterNight.Phase != game.GamePhaseDay {
		t.Fatalf("expected day phase after resolving night, got %d", stateAfterNight.Phase)
	}
	if stateAfterNight.Winner != nil {
		t.Fatalf("expected game to continue after first night death, got winner %#v", stateAfterNight.Winner)
	}
	p1 := findPlayerInState(t, stateAfterNight, "p1")
	if p1.IsAlive {
		t.Fatal("expected p1 to be dead after storyteller night kill")
	}
	if len(stateAfterNight.Deaths) != 1 ||
		stateAfterNight.Deaths[0].PlayerID != "p1" ||
		stateAfterNight.Deaths[0].Cause != game.DeathCauseNightKill {
		t.Fatalf("expected p1 night-kill death record, got %#v", stateAfterNight.Deaths)
	}
	if !containsString(stateAfterNight.GhostVotesRemaining, "p1") {
		t.Fatalf("expected p1 to have a ghost vote after death, got %#v", stateAfterNight.GhostVotesRemaining)
	}
	clearAllMessages(storytellerConn, playerConns)

	h.handleMessage(playerConns["p2"], ClientMessage{
		Type:      MsgNominate,
		NomineeID: "p5",
	})
	assertNoErrorMessages(t, playerConns["p2"].Messages())
	if phase := h.sessions[roomID].Phase(); phase != game.GamePhaseVoting {
		t.Fatalf("expected voting phase after nomination, got %d", phase)
	}

	for _, voterID := range []string{"p1", "p2", "p3"} {
		yes := true
		h.handleMessage(playerConns[voterID], ClientMessage{
			Type:     MsgCastVote,
			Decision: &yes,
		})
		assertNoErrorMessages(t, playerConns[voterID].Messages())
	}

	h.handleMessage(storytellerConn, ClientMessage{Type: MsgResolveNomination})
	assertNoErrorMessages(t, storytellerConn.Messages())

	finalState := lastRoomState(t, storytellerConn.Messages())
	if finalState.Phase != game.GamePhaseFinished {
		t.Fatalf("expected finished phase after executing the Imp, got %d", finalState.Phase)
	}
	if finalState.Winner == nil {
		t.Fatal("expected game winner after executing the Imp")
	}
	if finalState.Winner.Winner != game.TeamGood || finalState.Winner.Reason != game.WinReasonImpExecuted {
		t.Fatalf("expected good imp-executed win, got %#v", finalState.Winner)
	}
	p5 := findPlayerInState(t, finalState, "p5")
	if p5.IsAlive {
		t.Fatal("expected p5 Imp to be dead after execution")
	}
	if containsString(finalState.GhostVotesRemaining, "p1") {
		t.Fatalf("expected p1 ghost vote to be spent, got %#v", finalState.GhostVotesRemaining)
	}
	assertRoomDestroyed(t, h, roomID, append([]*FakeConnection{storytellerConn}, mapValues(playerConns)...)...)
}

func TestUseSlayerAbilityUsesConnectionIdentity(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupSlayerAssignedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{Type: MsgStartGame})
	assertNoErrorMessages(t, storytellerConn.Messages())
	h.handleMessage(storytellerConn, ClientMessage{
		Type:  MsgChangePhase,
		Phase: ClientGamePhase(game.GamePhaseDay),
	})
	assertNoErrorMessages(t, storytellerConn.Messages())
	clearAllMessages(storytellerConn, playerConns)

	h.handleMessage(playerConns["p1"], ClientMessage{
		Type:           MsgUseSlayerAbility,
		PlayerID:       "p2",
		TargetPlayerID: "p5",
	})
	assertNoErrorMessages(t, playerConns["p1"].Messages())

	death := lastPlayerDied(t, storytellerConn.Messages())
	if death.PlayerID != "p5" || death.Cause != game.DeathCauseAbility {
		t.Fatalf("expected p5 ability death, got %#v", death)
	}
	gameEnded := lastGameEnded(t, storytellerConn.Messages())
	if gameEnded.Winner != game.TeamGood || gameEnded.Reason != game.WinReasonImpExecuted {
		t.Fatalf("expected good demon-dead win, got %#v", gameEnded)
	}

	finalState := lastRoomState(t, storytellerConn.Messages())
	if finalState.Phase != game.GamePhaseFinished {
		t.Fatalf("expected finished phase after Slayer hit, got %d", finalState.Phase)
	}
	p5 := findPlayerInState(t, finalState, "p5")
	if p5.IsAlive {
		t.Fatal("expected p5 Imp to be dead after Slayer hit")
	}
	assertRoomDestroyed(t, h, roomID, append([]*FakeConnection{storytellerConn}, mapValues(playerConns)...)...)
}

func TestStorytellerNightActionMustMatchCurrentWakeStep(t *testing.T) {
	h, storytellerConn, _, _ := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:       MsgSubmitNightAction,
		ActionType: string(game.NightActionKill),
		TargetIDs:  []string{"p1"},
	})

	msgs := storytellerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "expected night action poison, got kill" {
		t.Fatalf("expected current wake step error, got %q", errMsg.Error)
	}
}

func TestStorytellerNightActionValidatesTargetCount(t *testing.T) {
	h, storytellerConn, _, _ := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:       MsgSubmitNightAction,
		ActionType: string(game.NightActionPoison),
	})

	msgs := storytellerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "night action poison requires at least 1 target(s)" {
		t.Fatalf("expected target count error, got %q", errMsg.Error)
	}
}

func TestResolveNightRejectsRemainingWakeSteps(t *testing.T) {
	h, storytellerConn, _, _ := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{Type: MsgResolveNight})

	msgs := storytellerConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "cannot resolve night with 5 wake step(s) remaining" {
		t.Fatalf("expected remaining wake steps error, got %q", errMsg.Error)
	}
}

func TestResolveNominationUsesAliveMajorityThreshold(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupDayAfterFirstNight(t)

	h.handleMessage(playerConns["p2"], ClientMessage{
		Type:      MsgNominate,
		NomineeID: "p5",
	})
	assertNoErrorMessages(t, playerConns["p2"].Messages())

	yes := true
	h.handleMessage(playerConns["p2"], ClientMessage{
		Type:     MsgCastVote,
		Decision: &yes,
	})
	assertNoErrorMessages(t, playerConns["p2"].Messages())

	h.handleMessage(storytellerConn, ClientMessage{Type: MsgResolveNomination})
	assertNoErrorMessages(t, storytellerConn.Messages())

	state := h.buildRoomStateForRecipient(roomID, "storyteller")
	if state.Phase != game.GamePhaseDay {
		t.Fatalf("expected failed nomination to return to day, got %d", state.Phase)
	}
	p5 := findPlayerInState(t, state, "p5")
	if !p5.IsAlive {
		t.Fatal("expected p5 to survive with only one yes vote below threshold")
	}

	resolved := lastNominationResolved(t, storytellerConn.Messages())
	if resolved.RequiredVotes != 2 {
		t.Fatalf("expected 2 required votes with 4 alive, got %d", resolved.RequiredVotes)
	}
	if resolved.Executed {
		t.Fatalf("expected nominee to be spared below threshold, got %#v", resolved)
	}
}

func TestExecutionByNominationEndsDayAndStartsNight(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupDayAfterFirstNight(t)

	h.handleMessage(playerConns["p3"], ClientMessage{
		Type:      MsgNominate,
		NomineeID: "p2",
	})
	assertNoErrorMessages(t, playerConns["p3"].Messages())

	yes := true
	for _, voterID := range []string{"p1", "p3"} {
		h.handleMessage(playerConns[voterID], ClientMessage{
			Type:     MsgCastVote,
			Decision: &yes,
		})
		assertNoErrorMessages(t, playerConns[voterID].Messages())
	}

	h.handleMessage(storytellerConn, ClientMessage{Type: MsgResolveNomination})
	assertNoErrorMessages(t, storytellerConn.Messages())

	state := h.buildRoomStateForRecipient(roomID, "storyteller")
	if state.Phase != game.GamePhaseNight {
		t.Fatalf("expected successful execution to start night, got %d", state.Phase)
	}
	p2 := findPlayerInState(t, state, "p2")
	if p2.IsAlive {
		t.Fatal("expected p2 to be dead after threshold execution")
	}
	if state.CurrentNightWakeStep == nil || state.CurrentNightWakeStep.ActionType != game.NightActionPoison {
		t.Fatalf("expected next night to start at Poisoner, got %#v", state.CurrentNightWakeStep)
	}

	resolved := lastNominationResolved(t, storytellerConn.Messages())
	if !resolved.Executed || resolved.RequiredVotes != 2 {
		t.Fatalf("expected executed nomination at threshold 2, got %#v", resolved)
	}
}

func TestEndGameRejectsNonStoryteller(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupStartedRoom(t)

	h.handleMessage(playerConns["p1"], ClientMessage{
		Type:   MsgEndGame,
		Winner: ClientTeam(game.TeamGood),
	})

	msgs := playerConns["p1"].Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "only the storyteller can end the game" {
		t.Fatalf("expected storyteller-only error, got %q", errMsg.Error)
	}
	if phase := h.sessions[roomID].Phase(); phase != game.GamePhaseNight {
		t.Fatalf("expected phase to remain night, got %d", phase)
	}
	if len(storytellerConn.Messages()) != 0 {
		t.Fatalf("expected no broadcast to storyteller, got %#v", storytellerConn.Messages())
	}
}

func TestStorytellerCanKillPlayerManually(t *testing.T) {
	h, storytellerConn, playerConns, _ := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:           MsgKillPlayer,
		TargetPlayerID: "p1",
		Cause:          ClientDeathCause(game.DeathCauseAbility),
	})
	assertNoErrorMessages(t, storytellerConn.Messages())

	death := lastPlayerDied(t, storytellerConn.Messages())
	if death.PlayerID != "p1" || death.Cause != game.DeathCauseAbility {
		t.Fatalf("expected p1 ability death, got %#v", death)
	}
	if playerDeath := lastPlayerDied(t, playerConns["p2"].Messages()); playerDeath.PlayerID != "p1" {
		t.Fatalf("expected other players to receive p1 death, got %#v", playerDeath)
	}

	state := lastRoomState(t, storytellerConn.Messages())
	if state.Phase != game.GamePhaseNight {
		t.Fatalf("expected manual death to preserve phase, got %d", state.Phase)
	}
	p1 := findPlayerInState(t, state, "p1")
	if p1.IsAlive {
		t.Fatal("expected p1 to be dead after manual storyteller kill")
	}
	if len(state.Deaths) != 1 ||
		state.Deaths[0].PlayerID != "p1" ||
		state.Deaths[0].Cause != game.DeathCauseAbility ||
		state.Deaths[0].KilledBy != "storyteller" {
		t.Fatalf("expected manual death record, got %#v", state.Deaths)
	}
}

func TestKillPlayerRejectsNonStorytellerIdentitySpoof(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupStartedRoom(t)

	h.handleMessage(playerConns["p1"], ClientMessage{
		Type:           MsgKillPlayer,
		PlayerID:       "storyteller",
		TargetPlayerID: "p2",
		Cause:          ClientDeathCause(game.DeathCauseAbility),
	})

	msgs := playerConns["p1"].Messages()
	if len(msgs) == 0 {
		t.Fatal("expected error response")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" {
		t.Fatalf("expected ERROR message, got %#v", msgs[0])
	}
	if errMsg.Error != "only the storyteller can kill players" {
		t.Fatalf("expected storyteller-only kill error, got %q", errMsg.Error)
	}
	if len(storytellerConn.Messages()) != 0 {
		t.Fatalf("expected no broadcast to storyteller, got %#v", storytellerConn.Messages())
	}
	state := h.buildRoomStateForRecipient(roomID, "storyteller")
	p2 := findPlayerInState(t, state, "p2")
	if !p2.IsAlive {
		t.Fatal("expected spoofed manual kill to leave p2 alive")
	}
}

func TestStorytellerCanEndGameManually(t *testing.T) {
	h, storytellerConn, playerConns, roomID := setupStartedRoom(t)

	h.handleMessage(storytellerConn, ClientMessage{
		Type:        MsgEndGame,
		Winner:      ClientTeam(game.TeamEvil),
		Description: "The town accepted evil's claim.",
	})
	assertNoErrorMessages(t, storytellerConn.Messages())

	state := lastRoomState(t, storytellerConn.Messages())
	if state.Phase != game.GamePhaseFinished {
		t.Fatalf("expected finished phase, got %d", state.Phase)
	}
	if state.Winner == nil {
		t.Fatal("expected winner after manual end game")
	}
	if state.Winner.Winner != game.TeamEvil ||
		state.Winner.Reason != game.WinReasonStorytellerDecision ||
		state.Winner.Description != "The town accepted evil's claim." {
		t.Fatalf("unexpected winner payload: %#v", state.Winner)
	}
	if lastGameEnded(t, storytellerConn.Messages()).Winner != game.TeamEvil {
		t.Fatalf("expected storyteller to receive evil gameEnded, got %#v", storytellerConn.Messages())
	}
	for playerID, conn := range playerConns {
		if lastGameEnded(t, conn.Messages()).Winner != game.TeamEvil {
			t.Fatalf("expected %s to receive evil gameEnded, got %#v", playerID, conn.Messages())
		}
	}
	assertRoomDestroyed(t, h, roomID, append([]*FakeConnection{storytellerConn}, mapValues(playerConns)...)...)
}

func TestPersistentHubRestoresGameAfterRestart(t *testing.T) {
	store := NewFileSnapshotStore(t.TempDir() + "/clocktower-snapshot.json")
	h, err := NewHubWithSnapshotStore(store)
	if err != nil {
		t.Fatalf("unexpected persistent hub error: %v", err)
	}

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
	h.handleMessage(storytellerConn, ClientMessage{Type: MsgStartGame})
	submitStorytellerFirstNightActions(t, h, storytellerConn)
	h.handleMessage(storytellerConn, ClientMessage{Type: MsgResolveNight})
	assertNoErrorMessages(t, storytellerConn.Messages())

	restored, err := NewHubWithSnapshotStore(store)
	if err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}

	reconnectConn := NewFakeConnection()
	restored.handleMessage(reconnectConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "p1",
		PlayerName: "p1 restored",
	})

	msgs := reconnectConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected restored room state after reconnect")
	}
	stateMsg, ok := msgs[0].(ServerMessage)
	if !ok || stateMsg.Type != "ROOM_STATE" || stateMsg.State == nil {
		t.Fatalf("expected restored ROOM_STATE, got %#v", msgs[0])
	}
	if stateMsg.RoomID != roomID {
		t.Fatalf("expected restored room %s, got %s", roomID, stateMsg.RoomID)
	}
	if stateMsg.State.Phase != game.GamePhaseDay {
		t.Fatalf("expected restored day phase, got %d", stateMsg.State.Phase)
	}
	p1 := findPlayerInState(t, stateMsg.State, "p1")
	if p1.IsAlive {
		t.Fatal("expected restored p1 death state")
	}
	if p1.Character == nil || p1.Character.ID != "washerwoman" {
		t.Fatalf("expected p1 to recover their own character, got %#v", p1.Character)
	}
	if len(stateMsg.State.Deaths) != 1 || stateMsg.State.Deaths[0].PlayerID != "p1" {
		t.Fatalf("expected restored death record for p1, got %#v", stateMsg.State.Deaths)
	}
	if !containsString(stateMsg.State.GhostVotesRemaining, "p1") {
		t.Fatalf("expected restored ghost vote for p1, got %#v", stateMsg.State.GhostVotesRemaining)
	}

	storytellerReconnect := NewFakeConnection()
	restored.handleMessage(storytellerReconnect, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "storyteller",
		PlayerName: "Storyteller restored",
	})
	stMsgs := storytellerReconnect.Messages()
	if len(stMsgs) == 0 {
		t.Fatal("expected restored storyteller room state")
	}
	stStateMsg, ok := stMsgs[0].(ServerMessage)
	if !ok || stStateMsg.State == nil {
		t.Fatalf("expected storyteller ROOM_STATE, got %#v", stMsgs[0])
	}
	for _, player := range stStateMsg.State.Players {
		if player.Character == nil {
			t.Fatalf("expected storyteller restored snapshot to reveal all characters, got %#v", stStateMsg.State.Players)
		}
	}
}

func TestPersistentHubRemovesDestroyedRoomFromSnapshot(t *testing.T) {
	store := NewFileSnapshotStore(t.TempDir() + "/clocktower-snapshot.json")
	h, err := NewHubWithSnapshotStore(store)
	if err != nil {
		t.Fatalf("unexpected persistent hub error: %v", err)
	}

	conn := NewFakeConnection()
	h.handleMessage(conn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 5,
	})
	roomID := conn.Messages()[0].(ServerMessage).RoomID
	conn.ClearMessages()

	h.handleMessage(conn, ClientMessage{Type: MsgLeaveRoom})
	assertNoErrorMessages(t, conn.Messages())

	restored, err := NewHubWithSnapshotStore(store)
	if err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}
	reconnectConn := NewFakeConnection()
	restored.handleMessage(reconnectConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "creator",
		PlayerName: "Creator",
	})

	msgs := reconnectConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected join error for destroyed room")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" || errMsg.Error != "room not found" {
		t.Fatalf("expected room not found after destroyed room restore, got %#v", msgs[0])
	}
}

func TestPersistentHubRestoresKickedPlayers(t *testing.T) {
	store := NewFileSnapshotStore(t.TempDir() + "/clocktower-snapshot.json")
	h, err := NewHubWithSnapshotStore(store)
	if err != nil {
		t.Fatalf("unexpected persistent hub error: %v", err)
	}

	creatorConn := NewFakeConnection()
	h.handleMessage(creatorConn, ClientMessage{
		Type:       MsgCreateRoom,
		PlayerID:   "creator",
		PlayerName: "Creator",
		MaxPlayers: 5,
	})
	roomID := creatorConn.Messages()[0].(ServerMessage).RoomID
	victimConn := NewFakeConnection()
	h.handleMessage(victimConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "victim",
		PlayerName: "Victim",
	})
	h.handleMessage(creatorConn, ClientMessage{
		Type:           MsgKickPlayer,
		TargetPlayerID: "victim",
	})
	assertNoErrorMessages(t, creatorConn.Messages())

	restored, err := NewHubWithSnapshotStore(store)
	if err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}
	rejoinConn := NewFakeConnection()
	restored.handleMessage(rejoinConn, ClientMessage{
		Type:       MsgJoinRoom,
		RoomID:     roomID,
		PlayerID:   "victim",
		PlayerName: "Victim",
	})

	msgs := rejoinConn.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected kicked rejoin error after restore")
	}
	errMsg, ok := msgs[0].(ServerMessage)
	if !ok || errMsg.Type != "ERROR" || errMsg.Error != "player was kicked from room" {
		t.Fatalf("expected restored kick rejection, got %#v", msgs[0])
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

func setupDayAfterFirstNight(t *testing.T) (*Hub, *FakeConnection, map[string]*FakeConnection, string) {
	t.Helper()

	h, storytellerConn, playerConns, roomID := setupStartedRoom(t)
	submitStorytellerFirstNightActions(t, h, storytellerConn)
	h.handleMessage(storytellerConn, ClientMessage{Type: MsgResolveNight})
	assertNoErrorMessages(t, storytellerConn.Messages())
	clearAllMessages(storytellerConn, playerConns)

	state := h.buildRoomStateForRecipient(roomID, "storyteller")
	if state.Phase != game.GamePhaseDay {
		t.Fatalf("expected setup helper to reach day phase, got %d", state.Phase)
	}
	return h, storytellerConn, playerConns, roomID
}

func setupAssignedRoom(t *testing.T) (*Hub, *FakeConnection, map[string]*FakeConnection, string) {
	t.Helper()

	return setupAssignedRoomWithAssignments(t, map[string]string{
		"p1": "washerwoman",
		"p2": "librarian",
		"p3": "investigator",
		"p4": "poisoner",
		"p5": "imp",
	})
}

func setupSlayerAssignedRoom(t *testing.T) (*Hub, *FakeConnection, map[string]*FakeConnection, string) {
	t.Helper()

	return setupAssignedRoomWithAssignments(t, map[string]string{
		"p1": "slayer",
		"p2": "librarian",
		"p3": "investigator",
		"p4": "poisoner",
		"p5": "imp",
	})
}

func setupAssignedRoomWithAssignments(t *testing.T, assignments map[string]string) (*Hub, *FakeConnection, map[string]*FakeConnection, string) {
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
		Type:        MsgAssignCharacters,
		Assignments: assignments,
	})
	assertNoErrorMessages(t, storytellerConn.Messages())
	clearAllMessages(storytellerConn, playerConns)

	return h, storytellerConn, playerConns, roomID
}

func submitStorytellerFirstNightActions(t *testing.T, h *Hub, storytellerConn *FakeConnection) {
	t.Helper()

	actions := []ClientMessage{
		{Type: MsgSubmitNightAction, ActionType: string(game.NightActionPoison), TargetIDs: []string{"p2"}},
		{Type: MsgSubmitNightAction, ActionType: string(game.NightActionLearnTownsfolk), TargetIDs: []string{"p1", "p2"}},
		{Type: MsgSubmitNightAction, ActionType: string(game.NightActionLearnOutsider)},
		{Type: MsgSubmitNightAction, ActionType: string(game.NightActionLearnMinion), TargetIDs: []string{"p4", "p5"}},
		{Type: MsgSubmitNightAction, ActionType: string(game.NightActionKill), TargetIDs: []string{"p1"}},
	}

	for _, action := range actions {
		storytellerConn.ClearMessages()
		h.handleMessage(storytellerConn, action)
		assertNoErrorMessages(t, storytellerConn.Messages())
	}
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

func lastNightActionSubmitted(t *testing.T, messages []any) *game.NightActionEvent {
	t.Helper()
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(ServerMessage)
		if !ok || msg.Event == nil || msg.Event.NightActionSubmitted == nil {
			continue
		}
		return msg.Event.NightActionSubmitted
	}
	t.Fatalf("expected night action submitted event in %#v", messages)
	return nil
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

func findPlayerInState(t *testing.T, state *RoomState, playerID string) game.Player {
	t.Helper()
	for _, player := range state.Players {
		if player.ID == playerID {
			return player
		}
	}
	t.Fatalf("expected player %s in room state", playerID)
	return game.Player{}
}

func findOptionalPlayerInState(state *RoomState, playerID string) *game.Player {
	if state == nil {
		return nil
	}
	for i := range state.Players {
		if state.Players[i].ID == playerID {
			return &state.Players[i]
		}
	}
	return nil
}

func lastNominationResolved(t *testing.T, messages []any) *game.NominationResolvedEvent {
	t.Helper()
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(ServerMessage)
		if !ok || msg.Event == nil || msg.Event.NominationResolved == nil {
			continue
		}
		return msg.Event.NominationResolved
	}
	t.Fatalf("expected nomination resolved event in %#v", messages)
	return nil
}

func lastGameEnded(t *testing.T, messages []any) *game.GameEndedEvent {
	t.Helper()
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(ServerMessage)
		if !ok || msg.Event == nil || msg.Event.GameEnded == nil {
			continue
		}
		return msg.Event.GameEnded
	}
	t.Fatalf("expected game ended event in %#v", messages)
	return nil
}

func lastPlayerDied(t *testing.T, messages []any) *game.PlayerDiedEvent {
	t.Helper()
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(ServerMessage)
		if !ok || msg.Event == nil || msg.Event.PlayerDied == nil {
			continue
		}
		return msg.Event.PlayerDied
	}
	t.Fatalf("expected player died event in %#v", messages)
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func mapValues(values map[string]*FakeConnection) []*FakeConnection {
	result := make([]*FakeConnection, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func assertRoomDestroyed(t *testing.T, h *Hub, roomID string, conns ...*FakeConnection) {
	t.Helper()
	if h.rm.GetRoom(roomID) != nil {
		t.Fatalf("expected room %s to be destroyed", roomID)
	}

	h.mu.RLock()
	_, hasSession := h.sessions[roomID]
	mapped := make([]*FakeConnection, 0)
	for _, conn := range conns {
		if _, ok := h.connToRoom[conn]; ok {
			mapped = append(mapped, conn)
		}
	}
	h.mu.RUnlock()

	if hasSession {
		t.Fatalf("expected session %s to be removed", roomID)
	}
	if len(mapped) != 0 {
		t.Fatalf("expected %d connection mapping(s) to be removed", len(mapped))
	}
}
