package ws

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/gameplay"
	"github.com/marskim1130/blood-on-the-clocktower/internal/session"
	"github.com/marskim1130/blood-on-the-clocktower/internal/sessionstore"
)

func TestStorytellerCanUndoAndRedoGameActionWithoutChangingRoster(t *testing.T) {
	e := newSessionGameEngine("room", "trouble_brewing")
	_ = e.AddPlayer("st", "主持")
	_ = e.AddPlayer("p1", "玩家")
	if _, err := e.Execute("st", gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Execute("p1", gameplay.SetReadyCmd{SenderID: "p1", Ready: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Execute("st", historyCommand{}); err != nil {
		t.Fatal(err)
	}
	if e.session.Players()[0].IsReady || e.session.Players()[0].ID != "p1" {
		t.Fatal("undo did not restore ready state and roster")
	}
	if _, err := e.Execute("st", historyCommand{redo: true}); err != nil {
		t.Fatal(err)
	}
	if !e.session.Players()[0].IsReady {
		t.Fatal("redo did not restore ready")
	}
}

func TestMembershipChangeInvalidatesUndoSnapshots(t *testing.T) {
	e := newSessionGameEngine("room", "trouble_brewing")
	_ = e.AddPlayer("st", "主持")
	_ = e.AddPlayer("p1", "玩家")
	_, _ = e.Execute("st", gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"})
	_, _ = e.Execute("p1", gameplay.SetReadyCmd{SenderID: "p1", Ready: true})
	_ = e.AddPlayer("p2", "新玩家")
	if _, err := e.Execute("st", historyCommand{}); err == nil {
		t.Fatal("undo can restore an obsolete roster")
	}
	if len(e.session.Players()) != 2 {
		t.Fatal("membership was rolled back")
	}
}

func TestUndoRestoresVotingTimerPaused(t *testing.T) {
	snapshot, _ := json.Marshal(map[string]any{"storytellerId": "st", "phase": game.GamePhaseVoting, "players": []game.Player{{ID: "p1", IsAlive: true}}, "nomination": game.Nomination{NominatorID: "p1", NomineeID: "p1", VoterOrder: []string{"p1"}, Votes: map[string]bool{}, Stage: game.NominationStageVoting, DeadlineUnixMs: time.Now().Add(3 * time.Second).UnixMilli(), RemainingMs: 3000}})
	loaded, err := loadSessionGameEngine("room", snapshot)
	if err != nil {
		t.Fatal(err)
	}
	e := loaded.(*sessionGameEngine)
	if _, err := e.Execute("p1", gameplay.CastVoteCmd{SenderID: "p1", Decision: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Execute("st", historyCommand{}); err != nil {
		t.Fatal(err)
	}
	nomination := e.session.ProjectionFor("st").Nomination
	if !nomination.Paused || nomination.DeadlineUnixMs != 0 || nomination.RemainingMs <= 0 {
		t.Fatalf("restored timer is running or expired: %+v", nomination)
	}
}

func TestHistorySurvivesReloadRequiresStorytellerAndExplicitPhaseConfirmation(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{"storytellerId": "st", "phase": game.GamePhaseDay, "players": []game.Player{{ID: "p1", IsAlive: true}}})
	loaded, _ := loadSessionGameEngine("room", raw)
	e := loaded.(*sessionGameEngine)
	if _, err := e.Execute("st", gameplay.EndGameCmd{SenderID: "st", Winner: game.TeamGood}); err != nil {
		t.Fatal(err)
	}
	saved, err := e.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err = loadSessionGameEngine("room", saved)
	if err != nil {
		t.Fatal(err)
	}
	e = loaded.(*sessionGameEngine)
	if _, err = e.Execute("p1", historyCommand{confirmPhaseChange: true}); err == nil {
		t.Fatal("player undid storyteller action")
	}
	if _, err = e.Execute("st", historyCommand{}); err == nil {
		t.Fatal("phase crossing did not require explicit confirmation")
	}
	if len(e.Project("p1").(*RoomState).OperationLog) != 0 || e.Project("p1").(*RoomState).CanUndo {
		t.Fatal("private log leaked")
	}
	view := e.Project("st").(*RoomState)
	if !view.CanUndo || !view.UndoCrossesPhase || len(view.OperationLog) != 1 {
		t.Fatalf("storyteller history missing: %+v", view)
	}
	if _, err = e.Execute("st", historyCommand{confirmPhaseChange: true}); err != nil {
		t.Fatal(err)
	}
	if e.session.Phase() != game.GamePhaseDay {
		t.Fatal("undo did not restore day")
	}
	if _, err = e.Execute("st", gameplay.EndGameCmd{SenderID: "st", Winner: game.TeamEvil}); err != nil {
		t.Fatal(err)
	}
	if e.Project("st").(*RoomState).CanRedo {
		t.Fatal("new game action retained redo")
	}
}

type historyFailStore struct {
	sessionstore.RoomRecordStore
	fail bool
}

func (s *historyFailStore) Replace(ctx context.Context, revision uint64, record sessionstore.Record) error {
	if s.fail {
		return errors.New("storage unavailable")
	}
	return s.RoomRecordStore.Replace(ctx, revision, record)
}

func TestHistoryPersistenceFailureRollsBackGameAndHistoryAtomically(t *testing.T) {
	ctx := context.Background()
	store := &historyFailStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	registry := session.NewRegistry(store, session.NewHMACCredentialCodec(developmentCredentialKey()), nil, nil)
	created, err := registry.Create(ctx, session.CreateInput{RequestID: "create-history", PlayerID: "st", PlayerName: "主持", Fingerprint: "create", Engine: newSessionGameEngine("", "trouble_brewing")})
	if err != nil {
		t.Fatal(err)
	}
	room, _ := registry.Get(created.RoomID)
	joined, err := room.JoinObserved(ctx, session.JoinInput{RequestID: "join-history", PlayerID: "p1", PlayerName: "玩家", Fingerprint: "join"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	actor := session.Actor{PlayerID: "st", Credential: created.ResumeCredential, ClientSequence: 1}
	if _, err = room.Execute(ctx, actor, session.Command{Kind: session.CommandSetStoryteller, Fingerprint: "set-st", Payload: gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = room.Execute(ctx, session.Actor{PlayerID: "p1", Credential: joined.ResumeCredential, ClientSequence: 1}, session.Command{Kind: session.CommandGame, Fingerprint: "ready", Payload: gameplay.SetReadyCmd{SenderID: "p1", Ready: true}}); err != nil {
		t.Fatal(err)
	}
	actor.ClientSequence = 2
	command := session.Command{Kind: session.CommandUndoGame, Fingerprint: "undo", Payload: historyCommand{}}
	before, _ := room.Query(actor)
	store.fail = true
	if _, err = room.Execute(ctx, actor, command); !errors.Is(err, session.ErrPersistenceUnavailable) {
		t.Fatalf("undo error=%v", err)
	}
	after, _ := room.Query(actor)
	if after.RoomRevision != before.RoomRevision || !after.Room.(*RoomState).Players[0].IsReady || !after.Room.(*RoomState).CanUndo || after.Room.(*RoomState).CanRedo {
		t.Fatal("uncommitted undo leaked")
	}
	store.fail = false
	result, err := room.Execute(ctx, actor, command)
	if err != nil {
		t.Fatal(err)
	}
	if result.NextClientSequence != 3 {
		t.Fatal("undo changed sequence lineage")
	}
	state, _ := room.Query(actor)
	if state.Room.(*RoomState).Players[0].IsReady || !state.Room.(*RoomState).CanRedo {
		t.Fatal("committed undo missing")
	}
	if credential, ok := room.CredentialFor("p1"); !ok || credential != joined.ResumeCredential {
		t.Fatal("undo changed member credentials")
	}
}

func TestProtocolV2UndoRedoProjectsHistoryOnlyToStoryteller(t *testing.T) {
	h := newProtocolV2GameHarness(t)
	st := h.create("st", "主持")
	p1 := h.join("p1", "玩家")
	h.command(st, ClientMessage{Type: MsgSetStoryteller, TargetPlayerID: st.id})
	ready := true
	h.command(p1, ClientMessage{Type: MsgSetReady, Ready: &ready})
	undone := h.command(st, ClientMessage{Type: MsgUndoGame})
	if undone.direct.State.Players[0].IsReady || !undone.direct.State.CanRedo || len(undone.direct.State.OperationLog) == 0 {
		t.Fatal("undo projection missing")
	}
	for recipient, broadcast := range undone.broadcasts {
		if recipient != st.id && (len(broadcast.State.OperationLog) > 0 || broadcast.State.CanUndo || broadcast.State.CanRedo) {
			t.Fatal("history leaked to player broadcast")
		}
	}
	redone := h.command(st, ClientMessage{Type: MsgRedoGame})
	if !redone.direct.State.Players[0].IsReady || redone.direct.State.CanRedo {
		t.Fatal("redo failed")
	}
}

func TestHistoryIsBoundedAndSnapshotsDoNotNestHistory(t *testing.T) {
	e := newSessionGameEngine("room", "trouble_brewing")
	_ = e.AddPlayer("st", "主持")
	_ = e.AddPlayer("p1", "玩家")
	_, _ = e.Execute("st", gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"})
	for index := 0; index < 110; index++ {
		if _, err := e.Execute("p1", gameplay.SetReadyCmd{SenderID: "p1", Ready: index%2 == 0}); err != nil {
			t.Fatal(err)
		}
	}
	if len(e.history.Undo) != maxUndoSteps || len(e.history.Logs) != maxOperationLogs {
		t.Fatal("history exceeded retention limits")
	}
	for _, entry := range e.history.Undo {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(entry.Before, &raw); err != nil {
			t.Fatal(err)
		}
		if _, exists := raw["operationHistory"]; exists {
			t.Fatal("history snapshot recursively contains history")
		}
	}
}
