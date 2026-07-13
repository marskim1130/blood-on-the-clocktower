package gameplay

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestChefAutoComputesAdjacentEvilPairs(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // Chef
		{ID: "p2", Name: "P2", IsAlive: true}, // Poisoner
		{ID: "p3", Name: "P3", IsAlive: true}, // Imp
		{ID: "p4", Name: "P4", IsAlive: true}, // Washerwoman
		{ID: "p5", Name: "P5", IsAlive: true}, // Empath
	})

	setupAndStartChefGame(t, gs, map[string]string{
		"p1": "chef",
		"p2": "poisoner",
		"p3": "imp",
		"p4": "washerwoman",
		"p5": "empath",
	})

	skipChefGameToCharacter(t, gs, "chef")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilPairs),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "1" {
		t.Fatalf("expected Chef result '1', got %v", event.Result)
	}
}

func TestChefCountsCircularAdjacentEvilPair(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // Poisoner
		{ID: "p2", Name: "P2", IsAlive: true}, // Chef
		{ID: "p3", Name: "P3", IsAlive: true}, // Washerwoman
		{ID: "p4", Name: "P4", IsAlive: true}, // Empath
		{ID: "p5", Name: "P5", IsAlive: true}, // Imp
	})

	setupAndStartChefGame(t, gs, map[string]string{
		"p1": "poisoner",
		"p2": "chef",
		"p3": "washerwoman",
		"p4": "empath",
		"p5": "imp",
	})

	skipChefGameToCharacter(t, gs, "chef")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilPairs),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "1" {
		t.Fatalf("expected circular Chef result '1', got %v", event.Result)
	}
}

func TestPoisonedChefDoesNotAutoCompute(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // Poisoner
		{ID: "p2", Name: "P2", IsAlive: true}, // Chef
		{ID: "p3", Name: "P3", IsAlive: true}, // Imp
		{ID: "p4", Name: "P4", IsAlive: true}, // Washerwoman
		{ID: "p5", Name: "P5", IsAlive: true}, // Empath
	})

	setupAndStartChefGame(t, gs, map[string]string{
		"p1": "poisoner",
		"p2": "chef",
		"p3": "imp",
		"p4": "washerwoman",
		"p5": "empath",
	})

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p2"},
	}); err != nil {
		t.Fatalf("Poison failed: %v", err)
	}

	skipChefGameToCharacter(t, gs, "chef")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilPairs),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result != nil && *event.Result != "" {
		t.Fatalf("expected poisoned Chef to have no auto result, got %v", event.Result)
	}
}

func TestChefManualResultOverridesAutoCompute(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true},
		{ID: "p2", Name: "P2", IsAlive: true},
		{ID: "p3", Name: "P3", IsAlive: true},
		{ID: "p4", Name: "P4", IsAlive: true},
		{ID: "p5", Name: "P5", IsAlive: true},
	})

	setupAndStartChefGame(t, gs, map[string]string{
		"p1": "chef",
		"p2": "poisoner",
		"p3": "imp",
		"p4": "washerwoman",
		"p5": "empath",
	})

	skipChefGameToCharacter(t, gs, "chef")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilPairs),
		Result:     "0",
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "0" {
		t.Fatalf("expected manual Chef result '0', got %v", event.Result)
	}
}

func setupAndStartChefGame(t *testing.T, gs *GameSession, assignments map[string]string) {
	t.Helper()

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}
	if _, err := gs.Apply(AssignCharactersCmd{SenderID: "storyteller", Assignments: assignments}); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}
	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}
}

func skipChefGameToCharacter(t *testing.T, gs *GameSession, targetCharID string) {
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
