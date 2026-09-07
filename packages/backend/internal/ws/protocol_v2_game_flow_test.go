package ws

import (
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/gameplay"
)

type protocolV2GameClient struct {
	id               string
	name             string
	connection       *websocket.Conn
	resumeCredential string
	nextSequence     uint64
}

type protocolV2GameCommit struct {
	direct     ServerMessage
	broadcasts map[string]ServerMessage
}

type protocolV2GameHarness struct {
	t        *testing.T
	server   *httptest.Server
	url      string
	roomID   string
	revision uint64
	clients  []*protocolV2GameClient
}

func newProtocolV2GameHarness(t *testing.T) *protocolV2GameHarness {
	t.Helper()

	hub := NewHub()
	server := httptest.NewServer(handlerForHub(hub))
	harness := &protocolV2GameHarness{
		t:      t,
		server: server,
		url:    "ws" + strings.TrimPrefix(server.URL, "http") + "/ws",
	}
	t.Cleanup(func() {
		for _, client := range harness.clients {
			if client.connection != nil {
				_ = client.connection.Close()
			}
		}
		server.Close()
	})
	return harness
}

func (h *protocolV2GameHarness) dial(id, name string) *protocolV2GameClient {
	h.t.Helper()
	connection, _, err := websocket.DefaultDialer.Dial(h.url, nil)
	if err != nil {
		h.t.Fatalf("dial %s websocket: %v", id, err)
	}
	return &protocolV2GameClient{id: id, name: name, connection: connection}
}

func (h *protocolV2GameHarness) create(id, name string) *protocolV2GameClient {
	h.t.Helper()
	client := h.dial(id, name)
	request := ClientMessage{
		ProtocolVersion: 2,
		Type:            MsgCreateRoom,
		RequestID:       "complete-game-create",
		PlayerID:        id,
		PlayerName:      name,
		MaxPlayers:      5,
		ScriptID:        game.TroubleBrewingScriptID,
	}
	h.write(client, request)
	result := readWireMessage(h.t, client.connection)
	if result.Type != ServerMsgCreateRoomResult || result.RoomID == "" || result.ResumeCredential == "" {
		h.t.Fatalf("unexpected create result: %+v", result)
	}
	if result.RoomRevision != 1 || result.NextClientSequence != 1 || result.State == nil {
		h.t.Fatalf("create must establish revision 1 and sequence 1: %+v", result)
	}
	h.roomID = result.RoomID
	h.revision = result.RoomRevision
	client.resumeCredential = result.ResumeCredential
	client.nextSequence = result.NextClientSequence
	h.clients = append(h.clients, client)
	return client
}

func (h *protocolV2GameHarness) join(id, name string) *protocolV2GameClient {
	h.t.Helper()
	existing := append([]*protocolV2GameClient(nil), h.clients...)
	client := h.dial(id, name)
	expectedRevision := h.revision + 1
	h.write(client, ClientMessage{
		ProtocolVersion: 2,
		Type:            MsgJoinRoom,
		JoinRequestID:   "complete-game-join-" + id,
		RoomID:          h.roomID,
		PlayerID:        id,
		PlayerName:      name,
	})

	result := readWireMessage(h.t, client.connection)
	if result.Type != ServerMsgJoinRoomResult || result.RoomID != h.roomID || result.ResumeCredential == "" {
		h.t.Fatalf("unexpected join result for %s: %+v", id, result)
	}
	if result.RoomRevision != expectedRevision || result.NextClientSequence != 1 || result.State == nil {
		h.t.Fatalf("join %s must commit revision %d: %+v", id, expectedRevision, result)
	}
	for _, recipient := range existing {
		broadcast := readWireMessage(h.t, recipient.connection)
		h.assertBroadcast(recipient, broadcast, expectedRevision)
	}

	client.resumeCredential = result.ResumeCredential
	client.nextSequence = result.NextClientSequence
	h.revision = expectedRevision
	h.clients = append(h.clients, client)
	return client
}

func (h *protocolV2GameHarness) command(sender *protocolV2GameClient, message ClientMessage) protocolV2GameCommit {
	h.t.Helper()
	expectedSequence := sender.nextSequence
	expectedRevision := h.revision + 1
	message.ProtocolVersion = 2
	message.RoomID = h.roomID
	message.PlayerID = sender.id
	message.ResumeCredential = sender.resumeCredential
	message.ClientSequence = expectedSequence
	h.write(sender, message)

	direct := readWireMessage(h.t, sender.connection)
	if direct.Type != ServerMsgCommandResult || direct.RoomID != h.roomID || direct.State == nil {
		h.t.Fatalf("unexpected command result for %s %s: %+v", sender.id, message.Type, direct)
	}
	if direct.AcceptedSequence != expectedSequence || direct.NextClientSequence != expectedSequence+1 {
		h.t.Fatalf("%s %s sequence mismatch: got accepted=%d next=%d, want accepted=%d next=%d", sender.id, message.Type, direct.AcceptedSequence, direct.NextClientSequence, expectedSequence, expectedSequence+1)
	}
	if direct.RoomRevision != expectedRevision {
		h.t.Fatalf("%s %s revision mismatch: got %d, want %d", sender.id, message.Type, direct.RoomRevision, expectedRevision)
	}

	broadcasts := make(map[string]ServerMessage, len(h.clients)-1)
	for _, recipient := range h.clients {
		if recipient == sender {
			continue
		}
		broadcast := readWireMessage(h.t, recipient.connection)
		h.assertBroadcast(recipient, broadcast, expectedRevision)
		broadcasts[recipient.id] = broadcast
	}

	sender.nextSequence++
	h.revision = expectedRevision
	return protocolV2GameCommit{direct: direct, broadcasts: broadcasts}
}

func (h *protocolV2GameHarness) resume(client *protocolV2GameClient) ServerMessage {
	h.t.Helper()
	if err := client.connection.Close(); err != nil {
		h.t.Fatalf("close %s before resume: %v", client.id, err)
	}
	replacement := h.dial(client.id, client.name)
	client.connection = replacement.connection
	h.write(client, ClientMessage{
		ProtocolVersion:  2,
		Type:             MsgResumeRoom,
		RoomID:           h.roomID,
		PlayerID:         client.id,
		ResumeCredential: client.resumeCredential,
	})
	result := readWireMessage(h.t, client.connection)
	if result.Type != ServerMsgResumeRoomResult || result.RoomID != h.roomID || result.State == nil {
		h.t.Fatalf("unexpected resume result for %s: %+v", client.id, result)
	}
	if result.RoomRevision != h.revision || result.NextClientSequence != client.nextSequence {
		h.t.Fatalf("resume %s lost revision or sequence: %+v", client.id, result)
	}
	return result
}

func (h *protocolV2GameHarness) write(client *protocolV2GameClient, message ClientMessage) {
	h.t.Helper()
	if err := client.connection.WriteJSON(message); err != nil {
		h.t.Fatalf("write %s for %s: %v", message.Type, client.id, err)
	}
}

func (h *protocolV2GameHarness) assertBroadcast(recipient *protocolV2GameClient, message ServerMessage, revision uint64) {
	h.t.Helper()
	if message.Type != ServerMsgRoomStateChanged || message.RoomID != h.roomID || message.State == nil {
		h.t.Fatalf("unexpected broadcast for %s: %+v", recipient.id, message)
	}
	if message.RoomRevision != revision || message.AcceptedSequence != 0 || message.NextClientSequence != 0 {
		h.t.Fatalf("broadcast for %s must contain only committed revision %d: %+v", recipient.id, revision, message)
	}
}

func TestProtocolV2CompleteGameOverWebSocket(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")
	p3 := harness.join("p3", "P3")
	p4 := harness.join("p4", "P4")
	p5 := harness.join("p5", "P5")

	setStoryteller := harness.command(storyteller, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: storyteller.id,
	})
	if setStoryteller.direct.State.StorytellerID != storyteller.id || len(setStoryteller.direct.State.Players) != 5 {
		t.Fatalf("creator must become storyteller without occupying a player seat: %+v", setStoryteller.direct.State)
	}
	ready := true
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4, p5} {
		harness.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}

	assignments := map[string]string{
		"p1": "slayer",
		"p2": "soldier",
		"p3": "mayor",
		"p4": "poisoner",
		"p5": "imp",
	}
	assigned := harness.command(storyteller, ClientMessage{
		Type:                   MsgAssignCharacters,
		Assignments:            assignments,
		DemonBluffCharacterIDs: []string{"chef", "empath", "fortuneteller"},
	})
	if !slices.Equal(assigned.direct.State.DemonBluffCharacterIDs, []string{"chef", "empath", "fortuneteller"}) {
		t.Fatalf("storyteller projection lost Demon bluffs: %+v", assigned.direct.State.DemonBluffCharacterIDs)
	}
	assertCharacterVisibility(t, assigned.direct.State, "", assignments, true)
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4, p5} {
		assertCharacterVisibility(t, assigned.broadcasts[player.id].State, player.id, assignments, false)
	}
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4, p5} {
		harness.command(player, ClientMessage{Type: MsgConfirmCharacter})
	}

	started := harness.command(storyteller, ClientMessage{Type: MsgStartGame})
	assertStorytellerNightProjection(t, started.direct.State, 0)
	for id, broadcast := range started.broadcasts {
		assertPlayerNightPrivacy(t, broadcast.State, id, assignments)
	}

	poisoned := harness.command(storyteller, ClientMessage{
		Type:       MsgSubmitNightAction,
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{p2.id},
	})
	assertStorytellerNightProjection(t, poisoned.direct.State, 1)
	if playerByID(t, poisoned.direct.State, p2.id).PoisonedUntil == nil {
		t.Fatal("storyteller must see that p2 is poisoned")
	}
	for id, broadcast := range poisoned.broadcasts {
		assertPlayerNightPrivacy(t, broadcast.State, id, assignments)
	}

	preparedDawn := harness.command(storyteller, ClientMessage{Type: MsgPrepareDawn})
	if !preparedDawn.direct.State.DawnReviewPending || len(preparedDawn.direct.State.PendingDawnDeathIDs) != 0 || preparedDawn.direct.State.Phase != game.GamePhaseNight {
		t.Fatalf("first-night dawn proposal must be an empty private review: %+v", preparedDawn.direct.State)
	}
	for id, broadcast := range preparedDawn.broadcasts {
		if broadcast.State.DawnReviewPending || len(broadcast.State.PendingDawnDeathIDs) != 0 {
			t.Fatalf("dawn review leaked to %s: %+v", id, broadcast.State)
		}
	}
	day := harness.command(storyteller, ClientMessage{Type: MsgConfirmDawn})
	assertFirstDay(t, day.direct.State)
	for _, broadcast := range day.broadcasts {
		assertFirstDay(t, broadcast.State)
		assertNoPoisoningDisclosure(t, broadcast.State)
	}

	nominated := harness.command(p1, ClientMessage{Type: MsgNominate, NomineeID: "p5"})
	assertNomination(t, nominated.direct.State, map[string]bool{})
	for _, broadcast := range nominated.broadcasts {
		assertNomination(t, broadcast.State, map[string]bool{})
	}

	yes := true
	no := false
	harness.command(storyteller, ClientMessage{Type: MsgAdvanceNominationStage})
	harness.command(storyteller, ClientMessage{Type: MsgAdvanceNominationStage})
	p1Vote := harness.command(p1, ClientMessage{Type: MsgCastVote, Decision: &yes})
	assertNomination(t, p1Vote.direct.State, map[string]bool{"p1": true})
	for _, broadcast := range p1Vote.broadcasts {
		assertNomination(t, broadcast.State, map[string]bool{"p1": true})
	}

	resumedP2 := harness.resume(p2)
	assertNomination(t, resumedP2.State, map[string]bool{"p1": true})
	if resumedP2.RoomRevision != 26 || resumedP2.NextClientSequence != 3 {
		t.Fatalf("p2 resume must continue at revision 26 and sequence 3: %+v", resumedP2)
	}

	p2Vote := harness.command(p2, ClientMessage{Type: MsgCastVote, Decision: &yes})
	assertNomination(t, p2Vote.direct.State, map[string]bool{"p1": true, "p2": true})

	p3Vote := harness.command(p3, ClientMessage{Type: MsgCastVote, Decision: &yes})
	assertNomination(t, p3Vote.direct.State, map[string]bool{"p1": true, "p2": true, "p3": true})
	harness.command(storyteller, ClientMessage{Type: MsgRecordVote, TargetPlayerID: p4.id, Decision: &no})
	p5Vote := harness.command(storyteller, ClientMessage{Type: MsgRecordVote, TargetPlayerID: p5.id, Decision: &no})
	assertNomination(t, p5Vote.direct.State, map[string]bool{"p1": true, "p2": true, "p3": true, "p4": false, "p5": false})

	resolved := harness.command(storyteller, ClientMessage{Type: MsgResolveNomination})
	if resolved.direct.State.ExecutionCandidateID != p5.id {
		t.Fatalf("demon must be on the block before day finalization: %+v", resolved.direct.State)
	}
	finished := harness.command(storyteller, ClientMessage{Type: MsgFinalizeDay})
	if harness.revision != 32 {
		t.Fatalf("complete game revision = %d, want 32", harness.revision)
	}
	assertGoodGameOver(t, finished.direct.State, assignments)
	assertCharacterVisibility(t, finished.direct.State, "", assignments, true)
	for recipientID, broadcast := range finished.broadcasts {
		assertGoodGameOver(t, broadcast.State, assignments)
		assertCharacterVisibility(t, broadcast.State, recipientID, assignments, false)
		if broadcast.State.GrimoireRevealed {
			t.Fatalf("game end published the grimoire before storyteller confirmation: %+v", broadcast.State)
		}
	}

	published := harness.command(storyteller, ClientMessage{Type: MsgPublishGrimoire})
	if !published.direct.State.GrimoireRevealed || harness.revision != finished.direct.RoomRevision+1 {
		t.Fatalf("grimoire publication was not committed: %+v revision=%d", published.direct.State, harness.revision)
	}
	for _, broadcast := range published.broadcasts {
		assertGoodGameOver(t, broadcast.State, assignments)
		assertCharacterVisibility(t, broadcast.State, "", assignments, true)
	}
	restarted := harness.command(storyteller, ClientMessage{Type: MsgRestartGame})
	state := restarted.direct.State
	if state.Phase != game.GamePhaseSetup || len(state.Players) != 5 || state.Winner != nil || state.GrimoireRevealed || len(state.NightActions) != 0 || len(state.DemonBluffCharacterIDs) != 0 || len(state.Deaths) != 0 || len(state.NominationResults) != 0 {
		t.Fatalf("restart did not clear finished game: %+v", state)
	}
	for _, player := range state.Players {
		if player.Character != nil || player.ShownCharacter != nil || player.IsReady || player.HasConfirmedCharacter || !player.IsAlive {
			t.Fatalf("restart retained private state: %+v", player)
		}
	}
	swapped := harness.command(storyteller, ClientMessage{Type: MsgSetStoryteller, TargetPlayerID: p1.id})
	if swapped.direct.State.StorytellerID != p1.id || len(swapped.direct.State.Players) != 5 || playerByID(t, swapped.direct.State, storyteller.id).Name != "Storyteller" {
		t.Fatalf("restart storyteller swap failed: %+v", swapped.direct.State)
	}
}

func assertCharacterVisibility(t *testing.T, state *RoomState, recipientID string, assignments map[string]string, revealAll bool) {
	t.Helper()
	if state == nil || len(state.Players) != len(assignments) {
		t.Fatalf("unexpected player projection: %+v", state)
	}
	for playerID, characterID := range assignments {
		player := playerByID(t, state, playerID)
		shouldSee := revealAll || playerID == recipientID
		if !shouldSee && player.Character != nil {
			t.Fatalf("%s must not see %s role before game over: %+v", recipientID, playerID, player.Character)
		}
		if shouldSee && (player.Character == nil || player.Character.ID != characterID) {
			t.Fatalf("%s role visibility for %s = %+v, want %s", recipientID, playerID, player.Character, characterID)
		}
	}
}

func assertStorytellerNightProjection(t *testing.T, state *RoomState, completed int) {
	t.Helper()
	if state == nil || state.Phase != game.GamePhaseNight {
		t.Fatalf("expected storyteller night projection: %+v", state)
	}
	if state.DayNumber != 0 || state.NightNumber != 1 {
		t.Fatalf("first night counters = day %d/night %d, want day 0/night 1", state.DayNumber, state.NightNumber)
	}
	expected := []game.NightActionType{
		game.NightActionPoison,
	}
	if len(state.NightWakeSteps) != len(expected) {
		t.Fatalf("night wake steps = %+v, want %v", state.NightWakeSteps, expected)
	}
	for index, actionType := range expected {
		if state.NightWakeSteps[index].ActionType != actionType {
			t.Fatalf("night wake step %d = %s, want %s", index, state.NightWakeSteps[index].ActionType, actionType)
		}
	}
	if state.CurrentNightWakeIndex != completed {
		t.Fatalf("current night wake index = %d, want %d", state.CurrentNightWakeIndex, completed)
	}
	if completed < len(expected) {
		if state.CurrentNightWakeStep == nil || state.CurrentNightWakeStep.ActionType != expected[completed] {
			t.Fatalf("current night wake step = %+v, want %s", state.CurrentNightWakeStep, expected[completed])
		}
	} else if state.CurrentNightWakeStep != nil {
		t.Fatalf("completed wake order must have no current step: %+v", state.CurrentNightWakeStep)
	}
	if len(state.NightActions) != completed {
		t.Fatalf("night action history length = %d, want %d: %+v", len(state.NightActions), completed, state.NightActions)
	}
	for index := range state.NightActions {
		if state.NightActions[index].ActorID != "storyteller" || state.NightActions[index].ActionType != expected[index] {
			t.Fatalf("night action %d = %+v", index, state.NightActions[index])
		}
	}
	if completed >= 1 {
		if len(state.NightActions[0].TargetIDs) != 1 || state.NightActions[0].TargetIDs[0] != "p2" {
			t.Fatalf("poison action targets = %v, want [p2]", state.NightActions[0].TargetIDs)
		}
	}
}

func assertPlayerNightPrivacy(t *testing.T, state *RoomState, recipientID string, assignments map[string]string) {
	t.Helper()
	if state == nil || state.Phase != game.GamePhaseNight {
		t.Fatalf("expected player night projection for %s: %+v", recipientID, state)
	}
	assertCharacterVisibility(t, state, recipientID, assignments, false)
	if len(state.NightWakeSteps) != 0 || len(state.NightActions) != 0 || state.PendingNightAction != nil || state.ConfirmedNightAction != nil {
		t.Fatalf("%s must not see storyteller night management: %+v", recipientID, state)
	}
	if step := state.CurrentNightWakeStep; step != nil {
		character := game.GetScriptCharacterByID(game.TroubleBrewingScriptID, assignments[recipientID])
		matchesCharacter := step.CharacterID != "" && step.CharacterID == assignments[recipientID]
		matchesType := character != nil && ((step.CharacterType == game.NightWakeCharacterTypeMinion && character.Type == game.CharacterTypeMinion) ||
			(step.CharacterType == game.NightWakeCharacterTypeDemon && character.Type == game.CharacterTypeDemon))
		if (!matchesCharacter && !matchesType) || state.NightTurnStatus != gameplay.NightTurnAwaitingPlayer {
			t.Fatalf("%s received another role's private night turn: %+v", recipientID, state)
		}
	} else if state.NightTurnStatus != "" {
		t.Fatalf("%s received a night status without a private step: %+v", recipientID, state)
	}
	assertNoPoisoningDisclosure(t, state)
}

func assertNoPoisoningDisclosure(t *testing.T, state *RoomState) {
	t.Helper()
	for _, player := range state.Players {
		if player.PoisonedUntil != nil {
			t.Fatalf("player projection disclosed %s poisoning: %d", player.ID, *player.PoisonedUntil)
		}
	}
}

func assertFirstDay(t *testing.T, state *RoomState) {
	t.Helper()
	if state == nil || state.Phase != game.GamePhaseDay || state.DayNumber != 1 {
		t.Fatalf("first resolved night must enter day 1: %+v", state)
	}
	if state.NightNumber != 1 {
		t.Fatalf("first day must follow night 1, got night %d", state.NightNumber)
	}
	if len(state.NightWakeSteps) != 0 || state.CurrentNightWakeStep != nil || len(state.NightActions) != 0 {
		t.Fatalf("day projection must not retain night management: %+v", state)
	}
}

func assertNomination(t *testing.T, state *RoomState, expectedVotes map[string]bool) {
	t.Helper()
	if state == nil || state.Phase != game.GamePhaseVoting || state.Nomination == nil {
		t.Fatalf("expected active nomination: %+v", state)
	}
	if state.Nomination.NominatorID != "p1" || state.Nomination.NomineeID != "p5" {
		t.Fatalf("unexpected nomination: %+v", state.Nomination)
	}
	if len(state.Nomination.Votes) != len(expectedVotes) {
		t.Fatalf("votes = %v, want %v", state.Nomination.Votes, expectedVotes)
	}
	for voterID, decision := range expectedVotes {
		if actual, ok := state.Nomination.Votes[voterID]; !ok || actual != decision {
			t.Fatalf("vote %s = %t/%t, want %t", voterID, actual, ok, decision)
		}
	}
}

func assertGoodGameOver(t *testing.T, state *RoomState, assignments map[string]string) {
	t.Helper()
	if state == nil || state.Phase != game.GamePhaseFinished || state.DayNumber != 1 {
		t.Fatalf("expected finished game on day 1: %+v", state)
	}
	if state.Winner == nil || state.Winner.Winner != game.TeamGood || state.Winner.Reason != game.WinReasonImpExecuted {
		t.Fatalf("expected good imp_executed result: %+v", state.Winner)
	}
	if len(state.Deaths) != 1 {
		t.Fatalf("deaths = %+v, want one execution", state.Deaths)
	}
	death := state.Deaths[0]
	if death.PlayerID != "p5" || death.DayNumber != 1 {
		t.Fatalf("unexpected Imp death: %+v", death)
	}
	if playerByID(t, state, "p5").IsAlive {
		t.Fatal("executed Imp must be dead")
	}
}

func playerByID(t *testing.T, state *RoomState, playerID string) game.Player {
	t.Helper()
	if state != nil {
		for _, player := range state.Players {
			if player.ID == playerID {
				return player
			}
		}
	}
	t.Fatalf("player %s missing from projection: %+v", playerID, state)
	return game.Player{}
}
