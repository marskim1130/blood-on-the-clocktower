package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestCloneNightActionsUsesEmptyTargetArrays(t *testing.T) {
	cloned := cloneNightActions([]game.NightAction{{ActionType: game.NightActionLearnDemon}})

	if len(cloned) != 1 || cloned[0].TargetIDs == nil || len(cloned[0].TargetIDs) != 0 {
		t.Fatalf("empty night targets must project as [], got %#v", cloned)
	}
}

func TestPlayerProjectionHidesPrivateDeathCauseAndKiller(t *testing.T) {
	gs := &GameSession{
		players:       []game.Player{{ID: "p1", IsAlive: true}, {ID: "p2", IsAlive: false}},
		storytellerID: "storyteller",
		phase:         game.GamePhaseDay,
		deaths: []game.DeathRecord{{
			PlayerID:  "p2",
			Cause:     game.DeathCauseNightKill,
			DayNumber: 2,
			KilledBy:  "p1",
		}},
	}

	playerDeath := gs.ProjectionFor("p1").Deaths[0]
	if playerDeath.Cause != "" || playerDeath.KilledBy != "" {
		t.Fatalf("player projection disclosed private death metadata: %+v", playerDeath)
	}
	storytellerDeath := gs.ProjectionFor("storyteller").Deaths[0]
	if storytellerDeath.Cause != game.DeathCauseNightKill || storytellerDeath.KilledBy != "p1" {
		t.Fatalf("storyteller lost private death metadata: %+v", storytellerDeath)
	}
}
