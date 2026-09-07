package gameplay

import (
	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"testing"
)

func TestRestartKeepsSeatsAndClearsEveryGameSecret(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "st"
	gs.storytellerName = "主持"
	gs.phase = game.GamePhaseFinished
	gs.players = []game.Player{{ID: "p2", Name: "二", IsAlive: false, Character: testCharacter(t, "imp")}, {ID: "left", Name: "已离开"}, {ID: "p1", Name: "一", Character: testCharacter(t, "chef")}}
	gs.winner = &game.GameEndedEvent{Winner: game.TeamGood}
	gs.grimoireRevealed = true
	gs.demonBluffCharacterIDs = []string{"monk"}
	if err := gs.Restart([]string{"st", "p1", "p2"}); err != nil {
		t.Fatal(err)
	}
	p := gs.ProjectionFor("st")
	if p.Phase != game.GamePhaseSetup || len(p.Players) != 2 || p.Players[0].ID != "p2" || p.Players[1].ID != "p1" {
		t.Fatalf("seats not preserved: %+v", p)
	}
	if p.Winner != nil || p.GrimoireRevealed || len(p.DemonBluffCharacterIDs) != 0 || p.Players[0].Character != nil || !p.Players[0].IsAlive || p.Players[0].IsReady {
		t.Fatalf("old game secret survived: %+v", p)
	}
}

func TestRestartAllowsNewStorytellerAndKeepsTheReplacedSeat(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID, gs.storytellerName = "st", "主持"
	gs.phase = game.GamePhaseFinished
	gs.players = []game.Player{{ID: "p1", Name: "一", Character: testCharacter(t, "imp")}}
	if err := gs.Restart([]string{"st", "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "st", TargetPlayerID: "p1"}); err != nil {
		t.Fatal(err)
	}
	if gs.StorytellerID() != "p1" || gs.Players()[0].ID != "st" || gs.Players()[0].Character != nil {
		t.Fatal("restart storyteller swap failed")
	}
}

func TestStorytellerCannotChangeAfterIdentityAssignment(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "st"
	gs.players = []game.Player{{ID: "p1", Character: testCharacter(t, "imp")}}
	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "st", TargetPlayerID: "p1"}); err == nil {
		t.Fatal("assigned identities must lock storyteller")
	}
	if gs.StorytellerID() != "st" || gs.Players()[0].ID != "p1" {
		t.Fatal("rejected swap mutated roster")
	}
}
