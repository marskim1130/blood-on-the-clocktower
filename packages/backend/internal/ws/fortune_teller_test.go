package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestFortuneTellerAutoComputesYesWhenTargetIncludesDemon(t *testing.T) {
	gs := newStartedFortuneTellerGame(t)
	skipFortuneTellerGameToCharacter(t, gs, "fortuneteller")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionCheckDemon),
		TargetIDs:  []string{"p2", "p5"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "yes" {
		t.Fatalf("expected Fortune Teller result 'yes', got %v", event.Result)
	}
}

func TestFortuneTellerAutoComputesNoWhenTargetsExcludeDemon(t *testing.T) {
	gs := newStartedFortuneTellerGame(t)
	skipFortuneTellerGameToCharacter(t, gs, "fortuneteller")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionCheckDemon),
		TargetIDs:  []string{"p2", "p3"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "no" {
		t.Fatalf("expected Fortune Teller result 'no', got %v", event.Result)
	}
}

func TestPoisonedFortuneTellerDoesNotAutoCompute(t *testing.T) {
	gs := newStartedFortuneTellerGame(t)

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p1"},
	}); err != nil {
		t.Fatalf("Poison failed: %v", err)
	}

	skipFortuneTellerGameToCharacter(t, gs, "fortuneteller")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionCheckDemon),
		TargetIDs:  []string{"p2", "p5"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result != nil && *event.Result != "" {
		t.Fatalf("expected poisoned Fortune Teller to have no auto result, got %v", event.Result)
	}
}

func TestFortuneTellerManualResultOverridesAutoCompute(t *testing.T) {
	gs := newStartedFortuneTellerGame(t)
	skipFortuneTellerGameToCharacter(t, gs, "fortuneteller")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionCheckDemon),
		TargetIDs:  []string{"p2", "p5"},
		Result:     "no",
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "no" {
		t.Fatalf("expected manual Fortune Teller result 'no', got %v", event.Result)
	}
}

func newStartedFortuneTellerGame(t *testing.T) *GameSession {
	t.Helper()

	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // Fortune Teller
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
			"p1": "fortuneteller",
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

func skipFortuneTellerGameToCharacter(t *testing.T, gs *GameSession, targetCharID string) {
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
				if player.ID == "storyteller" || (player.Character != nil && player.Character.ID == targetCharID) {
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
