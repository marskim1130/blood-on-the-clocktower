package gameplay_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/gameplay"
)

func TestGameplayRecordsSuccessfulCommandsForStoryteller(t *testing.T) {
	gs := gameplay.NewGameSession()
	gs.AddPlayer(game.Player{ID: "st", Name: "主持", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p1", Name: "玩家", IsAlive: true})
	if _, err := gs.Apply(gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(gameplay.SetReadyCmd{SenderID: "p1", Ready: true}); err != nil {
		t.Fatal(err)
	}
	view := gs.ProjectionFor("st")
	if !view.CanUndo || view.CanRedo || len(view.OperationLog) != 2 {
		t.Fatalf("successful game action was not recorded: %+v", view)
	}
	entry := view.OperationLog[1]
	if entry.ID != 2 || entry.ActorID != "p1" || entry.Action != "SetReady" || entry.FromPhase != game.GamePhaseSetup || entry.ToPhase != game.GamePhaseSetup {
		t.Fatalf("wrong action recorded: %+v", entry)
	}
}

func TestStorytellerCanUndoAndRedoWithinGameplay(t *testing.T) {
	gs := gameplay.NewGameSession()
	gs.AddPlayer(game.Player{ID: "st", Name: "主持", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p1", Name: "玩家", IsAlive: true})
	for _, cmd := range []gameplay.Command{
		gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"},
		gameplay.SetReadyCmd{SenderID: "p1", Ready: true},
	} {
		if _, err := gs.Apply(cmd); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := gs.Apply(gameplay.UndoGameCmd{SenderID: "p1"}); err == nil {
		t.Fatal("player undid a game action")
	}
	if _, err := gs.Apply(gameplay.UndoGameCmd{SenderID: "st"}); err != nil {
		t.Fatal(err)
	}
	undone := gs.ProjectionFor("st")
	if undone.Players[0].IsReady || !undone.CanRedo || undone.CanUndo || len(undone.OperationLog) != 3 || undone.OperationLog[2].Action != "UndoGame" {
		t.Fatalf("undo did not restore game and advance history: %+v", undone)
	}
	if _, err := gs.Apply(gameplay.RedoGameCmd{SenderID: "st"}); err != nil {
		t.Fatal(err)
	}
	redone := gs.ProjectionFor("st")
	if !redone.Players[0].IsReady || !redone.CanUndo || redone.CanRedo || len(redone.OperationLog) != 4 || redone.OperationLog[3].ID != 4 || redone.OperationLog[3].Action != "RedoGame" {
		t.Fatalf("redo did not restore game and advance history: %+v", redone)
	}
}

func TestJoiningPlayerInvalidatesGameplayUndoAndRedo(t *testing.T) {
	gs := gameplay.NewGameSession()
	gs.AddPlayer(game.Player{ID: "st", Name: "主持", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p1", Name: "玩家", IsAlive: true})
	for _, cmd := range []gameplay.Command{
		gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"},
		gameplay.SetReadyCmd{SenderID: "p1", Ready: true},
		gameplay.SetReadyCmd{SenderID: "p1", Ready: false},
		gameplay.UndoGameCmd{SenderID: "st"},
	} {
		if _, err := gs.Apply(cmd); err != nil {
			t.Fatal(err)
		}
	}
	before := gs.ProjectionFor("st")
	if !before.CanUndo || !before.CanRedo {
		t.Fatal("test must start with both undo and redo available")
	}
	gs.AddPlayer(game.Player{ID: "p2", Name: "新玩家", IsAlive: true})
	after := gs.ProjectionFor("st")
	if after.CanUndo || after.CanRedo || len(after.Players) != 2 || len(after.OperationLog) != len(before.OperationLog) {
		t.Fatalf("joining player did not invalidate restores while preserving logs: %+v", after)
	}
}

func TestReconnectingWithChangedNameInvalidatesOldGameHistory(t *testing.T) {
	gs := gameplay.NewGameSession()
	gs.AddPlayer(game.Player{ID: "st", Name: "主持", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p1", Name: "旧名字", IsAlive: true})
	for _, cmd := range []gameplay.Command{
		gameplay.SetStorytellerCmd{SenderID: "st", TargetPlayerID: "st"},
		gameplay.SetReadyCmd{SenderID: "p1", Ready: true},
	} {
		if _, err := gs.Apply(cmd); err != nil {
			t.Fatal(err)
		}
	}
	if added, err := gs.AddOrReconnectPlayer(game.Player{ID: "p1", Name: "旧名字"}, 5); err != nil || added || !gs.ProjectionFor("st").CanUndo {
		t.Fatal("unchanged reconnect should preserve game history")
	}
	if added, err := gs.AddOrReconnectPlayer(game.Player{ID: "p1", Name: "新名字"}, 5); err != nil || added {
		t.Fatalf("rename failed: added=%v err=%v", added, err)
	}
	if _, err := gs.Apply(gameplay.UndoGameCmd{SenderID: "st"}); err == nil {
		t.Fatal("game undo can roll back a changed player name")
	}
	if view := gs.ProjectionFor("st"); view.Players[0].Name != "新名字" || view.CanUndo || view.CanRedo {
		t.Fatal("rename did not establish a new history boundary")
	}
}

func TestLegacyHistorySurvivesGameplayReloadAndStaysPrivate(t *testing.T) {
	// This is the persisted shape produced by the original WebSocket history owner.
	before := []byte(`{"storytellerId":"st","phase":1,"players":[{"id":"p1","name":"玩家","isAlive":true,"isReady":false}]}`)
	after := []byte(`{"storytellerId":"st","phase":1,"players":[{"id":"p1","name":"玩家","isAlive":true,"isReady":true}]}`)
	var legacy map[string]any
	if err := json.Unmarshal(after, &legacy); err != nil {
		t.Fatal(err)
	}
	legacy["operationHistory"] = map[string]any{
		"undo":   []any{map[string]any{"before": before, "after": after, "fromPhase": game.GamePhaseSetup, "toPhase": game.GamePhaseSetup}},
		"logs":   []any{map[string]any{"id": 7, "actorId": "p1", "action": "SetReady", "createdAtUnixMs": 1234, "fromPhase": game.GamePhaseSetup, "toPhase": game.GamePhaseSetup, "disclosureWarning": true}},
		"nextId": 7,
	}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := gameplay.LoadGameSessionSnapshot(raw)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := loaded.Clone().MarshalSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	var oldWire, newWire map[string]any
	if err := json.Unmarshal(raw, &oldWire); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(saved, &newWire); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(oldWire["operationHistory"], newWire["operationHistory"]) {
		t.Fatal("persisted history no longer matches the legacy wire format")
	}
	reloaded, err := gameplay.LoadGameSessionSnapshot(saved)
	if err != nil {
		t.Fatal(err)
	}
	for _, recipient := range []string{"st", "p1", ""} {
		history := reloaded.ProjectionFor(recipient)
		if recipient == "st" {
			if !history.CanUndo || len(history.OperationLog) != 1 {
				t.Fatalf("legacy storyteller history was lost: %+v", history)
			}
			entry := history.OperationLog[0]
			if entry.ID != 7 || entry.Action != "SetReady" || entry.ActorID != "p1" || entry.CreatedAtUnixMs != 1234 || !entry.DisclosureWarning {
				t.Fatalf("legacy log changed: %+v", entry)
			}
		} else if history.CanUndo || len(history.OperationLog) != 0 {
			t.Fatalf("private history leaked to %q", recipient)
		}
	}
	if _, err := reloaded.Apply(gameplay.UndoGameCmd{SenderID: "st"}); err != nil {
		t.Fatal(err)
	}
	undone := reloaded.ProjectionFor("st")
	if undone.Players[0].IsReady || !undone.CanRedo || len(undone.OperationLog) != 2 || undone.OperationLog[1].ID != 8 {
		t.Fatal("legacy undo payload or log sequence could not be resumed")
	}
	clone := reloaded.Clone()
	if _, err := clone.Apply(gameplay.RedoGameCmd{SenderID: "st"}); err != nil {
		t.Fatal(err)
	}
	redone := clone.ProjectionFor("st")
	if !redone.Players[0].IsReady || redone.CanRedo || redone.OperationLog[len(redone.OperationLog)-1].ID != 9 {
		t.Fatal("legacy redo payload or log sequence could not be resumed")
	}
	original := reloaded.ProjectionFor("st")
	if original.Players[0].IsReady || !original.CanRedo || len(original.OperationLog) != 2 {
		t.Fatal("candidate restore mutated the committed game or history")
	}
}
