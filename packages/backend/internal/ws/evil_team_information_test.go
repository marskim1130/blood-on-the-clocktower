package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestFirstNightAutoComputesEvilTeamInformation(t *testing.T) {
	gs := newStartedTypeHintGame(t, "washerwoman", "librarian", "investigator", "poisoner", "imp")

	step := gs.currentNightWakeStepLocked()
	if step == nil || step.CharacterType != game.NightWakeCharacterTypeMinion || step.ActionType != game.NightActionLearnDemon {
		t.Fatalf("expected Minion demon-info step first, got %#v", step)
	}
	minionEvent := submitTypeHintNightAction(t, gs, game.NightActionLearnDemon, nil, "")
	if minionEvent.Result == nil || *minionEvent.Result != "Demon: P5 (Imp)" {
		t.Fatalf("expected Minions to learn Demon, got %v", minionEvent.Result)
	}

	step = gs.currentNightWakeStepLocked()
	if step == nil || step.CharacterType != game.NightWakeCharacterTypeDemon || step.ActionType != game.NightActionLearnMinion {
		t.Fatalf("expected Demon minion-info step second, got %#v", step)
	}
	demonEvent := submitTypeHintNightAction(t, gs, game.NightActionLearnMinion, nil, "")
	if demonEvent.Result == nil || *demonEvent.Result != "Minions: P4 (Poisoner)" {
		t.Fatalf("expected Demon to learn Minions, got %v", demonEvent.Result)
	}
}
