package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestWasherwomanDoesNotAutoComputeWhenTargetIncludesSpy(t *testing.T) {
	gs := newStartedTypeHintGame(t, "washerwoman", "empath", "chef", "spy", "imp")
	skipTypeHintGameToCharacter(t, gs, "washerwoman")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnTownsfolk, []string{"p4", "p2"}, "")

	assertNoAutoResult(t, event, "Washerwoman target includes Spy")
}

func TestInvestigatorDoesNotAutoComputeWhenTargetIncludesSpy(t *testing.T) {
	gs := newStartedTypeHintGame(t, "investigator", "washerwoman", "chef", "spy", "imp")
	skipTypeHintGameToCharacter(t, gs, "investigator")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnMinion, []string{"p4", "p2"}, "")

	assertNoAutoResult(t, event, "Investigator target includes Spy")
}

func TestFortuneTellerDoesNotAutoComputeWhenTargetIncludesRecluseWithoutDemon(t *testing.T) {
	gs := newStartedTypeHintGame(t, "fortuneteller", "recluse", "washerwoman", "chef", "spy", "imp")
	skipTypeHintGameToCharacter(t, gs, "fortuneteller")

	event := submitTypeHintNightAction(t, gs, game.NightActionCheckDemon, []string{"p2", "p3"}, "")

	assertNoAutoResult(t, event, "Fortune Teller target includes Recluse")
}

func TestChefDoesNotAutoComputeWhenSpyIsAlive(t *testing.T) {
	gs := newStartedTypeHintGame(t, "chef", "washerwoman", "empath", "spy", "imp")
	skipTypeHintGameToCharacter(t, gs, "chef")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnEvilPairs, nil, "")

	assertNoAutoResult(t, event, "Chef has alive Spy in play")
}

func TestEmpathDoesNotAutoComputeWhenAliveNeighborIsRecluse(t *testing.T) {
	gs := newStartedTypeHintGame(t, "empath", "recluse", "washerwoman", "chef", "spy", "imp")
	skipTypeHintGameToCharacter(t, gs, "empath")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnEvilNeighbors, nil, "")

	assertNoAutoResult(t, event, "Empath neighbor is Recluse")
}

func TestPoisonedSpyDoesNotBlockAutoRegistrationResult(t *testing.T) {
	gs := newStartedTypeHintGame(t, "washerwoman", "empath", "chef", "spy", "imp")
	poisonedUntil := gs.dayNumber
	gs.players[3].PoisonedUntil = &poisonedUntil
	skipTypeHintGameToCharacter(t, gs, "washerwoman")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnTownsfolk, []string{"p4", "p2"}, "")
	if event.Result == nil || *event.Result != "Empath" {
		t.Fatalf("expected poisoned Spy not to create registration ambiguity, got %v", event.Result)
	}
}

func assertNoAutoResult(t *testing.T, event *game.NightActionEvent, context string) {
	t.Helper()
	if event.Result != nil && *event.Result != "" {
		t.Fatalf("expected no auto result when %s, got %v", context, event.Result)
	}
}
