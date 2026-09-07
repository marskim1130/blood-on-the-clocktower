package gameplay

import (
	"slices"
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestStorytellerReviewsAndEditsDawnDeathsBeforeTheyBecomePublic(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
			{ID: "p2", IsAlive: true, Character: testCharacter(t, "librarian")},
			{ID: "p3", IsAlive: true, Character: testCharacter(t, "investigator")},
			{ID: "p4", IsAlive: true, Character: testCharacter(t, "chef")},
			{ID: "imp", IsAlive: true, Character: testCharacter(t, "imp")},
		},
		storytellerID: "storyteller",
		phase:         game.GamePhaseNight,
		nightNumber:   1,
		nightActions: []game.NightAction{{
			ActorID:    "imp",
			ActionType: game.NightActionKill,
			TargetIDs:  []string{"p1"},
		}},
	}

	if _, err := gs.Apply(PrepareDawnCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("prepare dawn failed: %v", err)
	}
	if !gs.dawnReviewPending || !slices.Equal(gs.pendingDawnDeathIDs, []string{"p1"}) {
		t.Fatalf("unexpected dawn proposal: pending=%v deaths=%v", gs.dawnReviewPending, gs.pendingDawnDeathIDs)
	}
	if gs.phase != game.GamePhaseNight || !gs.players[0].IsAlive || len(gs.deaths) != 0 {
		t.Fatalf("proposal changed public state before confirmation: phase=%v player=%+v deaths=%+v", gs.phase, gs.players[0], gs.deaths)
	}

	if _, err := gs.Apply(ConfirmDawnCmd{SenderID: "storyteller", DeathPlayerIDs: []string{"p1", "p2"}}); err != nil {
		t.Fatalf("confirm edited dawn failed: %v", err)
	}
	if gs.phase != game.GamePhaseDay || gs.players[0].IsAlive || gs.players[1].IsAlive {
		t.Fatalf("edited dawn was not applied atomically: phase=%v players=%+v", gs.phase, gs.players)
	}
	if len(gs.deaths) != 2 || gs.deaths[0].DayNumber != 1 || gs.deaths[1].DayNumber != 1 {
		t.Fatalf("unexpected public dawn batch: %+v", gs.deaths)
	}
}
