package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestUndertakerAutoComputesExecutedCharacter(t *testing.T) {
	gs := newStartedUndertakerGame(t)
	completeUndertakerFirstNight(t, gs)
	executeChefAndEnterNight(t, gs)
	skipUndertakerGameToCharacter(t, gs, "undertaker")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnExecuted),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "Chef" {
		t.Fatalf("expected Undertaker result 'Chef', got %v", event.Result)
	}
}

func TestUndertakerAutoComputesNoneWhenNoExecutionToday(t *testing.T) {
	gs := newStartedUndertakerGame(t)
	completeUndertakerFirstNight(t, gs)

	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err != nil {
		t.Fatalf("ChangePhase failed: %v", err)
	}
	skipUndertakerGameToCharacter(t, gs, "undertaker")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnExecuted),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "none" {
		t.Fatalf("expected Undertaker result 'none', got %v", event.Result)
	}
}

func TestPoisonedUndertakerDoesNotAutoCompute(t *testing.T) {
	gs := newStartedUndertakerGame(t)
	completeUndertakerFirstNight(t, gs)
	executeChefAndEnterNight(t, gs)

	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p1"},
	}); err != nil {
		t.Fatalf("Poison failed: %v", err)
	}
	skipUndertakerGameToCharacter(t, gs, "undertaker")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnExecuted),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result != nil && *event.Result != "" {
		t.Fatalf("expected poisoned Undertaker to have no auto result, got %v", event.Result)
	}
}

func TestUndertakerManualResultOverridesAutoCompute(t *testing.T) {
	gs := newStartedUndertakerGame(t)
	completeUndertakerFirstNight(t, gs)
	executeChefAndEnterNight(t, gs)
	skipUndertakerGameToCharacter(t, gs, "undertaker")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnExecuted),
		Result:     "Washerwoman",
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "Washerwoman" {
		t.Fatalf("expected manual Undertaker result 'Washerwoman', got %v", event.Result)
	}
}

func newStartedUndertakerGame(t *testing.T) *GameSession {
	t.Helper()

	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // Undertaker
		{ID: "p2", Name: "P2", IsAlive: true}, // Washerwoman
		{ID: "p3", Name: "P3", IsAlive: true}, // Chef
		{ID: "p4", Name: "P4", IsAlive: true}, // Poisoner
		{ID: "p5", Name: "P5", IsAlive: true}, // Imp
	})

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}
	if _, err := gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "undertaker",
			"p2": "washerwoman",
			"p3": "chef",
			"p4": "poisoner",
			"p5": "imp",
		},
	}); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}
	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	return gs
}

func completeUndertakerFirstNight(t *testing.T, gs *GameSession) {
	t.Helper()

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	actions := []SubmitNightActionCmd{
		{SenderID: "storyteller", ActionType: string(game.NightActionPoison), TargetIDs: []string{"p3"}},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnTownsfolk), TargetIDs: []string{"p1", "p2"}},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnEvilPairs)},
	}
	for _, action := range actions {
		if _, err := gs.Apply(action); err != nil {
			t.Fatalf("SubmitNightAction failed: %v", err)
		}
	}
	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
}

func executeChefAndEnterNight(t *testing.T, gs *GameSession) {
	t.Helper()

	if _, err := gs.Apply(ExecutePlayerCmd{SenderID: "storyteller", PlayerID: "p3"}); err != nil {
		t.Fatalf("ExecutePlayer failed: %v", err)
	}
	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err != nil {
		t.Fatalf("ChangePhase failed: %v", err)
	}
}

func skipUndertakerGameToCharacter(t *testing.T, gs *GameSession, targetCharID string) {
	t.Helper()

	for {
		step := gs.currentNightWakeStepLocked()
		if step == nil {
			t.Fatal("no more wake steps")
		}
		if step.CharacterID == targetCharID {
			return
		}

		targets := make([]string, 0, step.MinTargets)
		if step.MinTargets > 0 {
			for _, player := range gs.players {
				if player.ID == "storyteller" || !player.IsAlive || (player.Character != nil && player.Character.ID == targetCharID) {
					continue
				}
				targets = append(targets, player.ID)
				if len(targets) == step.MinTargets {
					break
				}
			}
		}

		if _, err := gs.Apply(SubmitNightActionCmd{
			SenderID:   "storyteller",
			ActionType: string(step.ActionType),
			TargetIDs:  targets,
		}); err != nil {
			t.Fatalf("skip action failed: %v", err)
		}
	}
}
