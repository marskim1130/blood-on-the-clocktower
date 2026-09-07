package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestFirstNightAutoComputesEvilTeamInformation(t *testing.T) {
	gs := newStartedTypeHintGame(t, "washerwoman", "librarian", "investigator", "poisoner", "imp", "soldier", "slayer")

	step := gs.currentNightWakeStepLocked()
	if step == nil || step.CharacterType != game.NightWakeCharacterTypeMinion || step.ActionType != game.NightActionLearnDemon {
		t.Fatalf("expected Minion demon-info step first, got %#v", step)
	}
	minionEvent := submitTypeHintNightAction(t, gs, game.NightActionLearnDemon, nil, "")
	if minionEvent.Result == nil || *minionEvent.Result != "Demon: P5 | Minions: P4" {
		t.Fatalf("expected Minions to learn Demon, got %v", minionEvent.Result)
	}

	step = gs.currentNightWakeStepLocked()
	if step == nil || step.CharacterType != game.NightWakeCharacterTypeDemon || step.ActionType != game.NightActionLearnMinion {
		t.Fatalf("expected Demon minion-info step second, got %#v", step)
	}
	demonEvent := submitTypeHintNightAction(t, gs, game.NightActionLearnMinion, nil, "")
	want := "Minions: P4 | Bluffs: Chef, Empath, Fortune Teller"
	if demonEvent.Result == nil || *demonEvent.Result != want {
		t.Fatalf("expected Demon to learn Minions and three bluffs, got %v", demonEvent.Result)
	}
}
