package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// TestEmpathAutoComputesEvilNeighbors verifies that Empath result is auto-calculated
func TestEmpathAutoComputesEvilNeighbors(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // washerwoman (good)
		{ID: "p2", Name: "P2", IsAlive: true}, // empath (good) - neighbors: p1 (good), p3 (evil)
		{ID: "p3", Name: "P3", IsAlive: true}, // poisoner (evil)
		{ID: "p4", Name: "P4", IsAlive: true}, // librarian (good)
		{ID: "p5", Name: "P5", IsAlive: true}, // imp (evil)
	})

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}

	if _, err := gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "empath",
			"p3": "poisoner",
			"p4": "librarian",
			"p5": "imp",
		},
	}); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}

	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p4"},
	}); err != nil {
		t.Fatalf("Poison failed: %v", err)
	}
	skipToCharacter(t, gs, "empath")

	// Now submit Empath action WITHOUT providing a result
	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilNeighbors),
		TargetIDs:  nil,
		Result:     "", // Empty - should be auto-computed
	})
	if err != nil {
		t.Fatalf("SubmitNightAction (Empath) failed: %v", err)
	}

	// Verify the event contains the computed result
	if len(result.Events) != 1 || result.Events[0].NightActionSubmitted == nil {
		t.Fatalf("expected NightActionSubmitted event, got %#v", result.Events)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil {
		t.Fatal("expected Empath result to be auto-computed, got nil")
	}

	// Empath (p2) has neighbors p1 (good) and p3 (evil) -> result should be "1"
	if *event.Result != "1" {
		t.Errorf("expected Empath result '1', got '%s'", *event.Result)
	}
}

// TestEmpathPoisonedNoAutoCompute verifies that poisoned Empath gets no auto-calculation
func TestEmpathPoisonedNoAutoCompute(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // washerwoman
		{ID: "p2", Name: "P2", IsAlive: true}, // empath
		{ID: "p3", Name: "P3", IsAlive: true}, // poisoner
		{ID: "p4", Name: "P4", IsAlive: true}, // librarian
		{ID: "p5", Name: "P5", IsAlive: true}, // imp
	})

	setupAndStartGame(t, gs, map[string]string{
		"p1": "washerwoman",
		"p2": "empath",
		"p3": "poisoner",
		"p4": "librarian",
		"p5": "imp",
	})

	// Poisoner poisons Empath (p2)
	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p2"},
	}); err != nil {
		t.Fatalf("Poison failed: %v", err)
	}

	// Skip to Empath
	skipToCharacter(t, gs, "empath")

	// Submit Empath action without result - should NOT auto-compute because poisoned
	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilNeighbors),
		TargetIDs:  nil,
		Result:     "",
	})
	if err != nil {
		t.Fatalf("SubmitNightAction (poisoned Empath) failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	// Poisoned Empath should get empty result (storyteller must manually provide false info)
	if event.Result != nil && *event.Result != "" {
		t.Errorf("expected poisoned Empath to have empty result, got '%s'", *event.Result)
	}
}

// TestEmpathManualResultOverridesAutoCompute verifies that manual result takes precedence
func TestEmpathManualResultOverridesAutoCompute(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true},
		{ID: "p2", Name: "P2", IsAlive: true}, // empath
		{ID: "p3", Name: "P3", IsAlive: true},
		{ID: "p4", Name: "P4", IsAlive: true},
		{ID: "p5", Name: "P5", IsAlive: true},
	})

	setupAndStartGame(t, gs, map[string]string{
		"p1": "washerwoman",
		"p2": "empath",
		"p3": "poisoner",
		"p4": "librarian",
		"p5": "imp",
	})

	skipToCharacter(t, gs, "empath")

	// Storyteller provides manual result (false info)
	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilNeighbors),
		TargetIDs:  nil,
		Result:     "0", // Manual override
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "0" {
		t.Errorf("expected manual result '0', got %v", event.Result)
	}
}

// TestEmpathBothNeighborsEvil verifies result when both neighbors are evil
func TestEmpathBothNeighborsEvil(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // poisoner (evil)
		{ID: "p2", Name: "P2", IsAlive: true}, // empath (good)
		{ID: "p3", Name: "P3", IsAlive: true}, // imp (evil)
		{ID: "p4", Name: "P4", IsAlive: true}, // washerwoman (good)
		{ID: "p5", Name: "P5", IsAlive: true}, // librarian (good)
	})

	setupAndStartGame(t, gs, map[string]string{
		"p1": "poisoner",
		"p2": "empath",
		"p3": "imp",
		"p4": "washerwoman",
		"p5": "librarian",
	})

	skipToCharacter(t, gs, "empath")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilNeighbors),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	// Empath (p2) neighbors: p1 (poisoner, evil) and p3 (imp, evil) -> "2"
	if event.Result == nil || *event.Result != "2" {
		t.Errorf("expected result '2', got %v", event.Result)
	}
}

// TestEmpathDeadNeighborNotCounted verifies dead neighbors are not counted
func TestEmpathDeadNeighborNotCounted(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: false}, // poisoner (evil, DEAD)
		{ID: "p2", Name: "P2", IsAlive: true},  // empath (good)
		{ID: "p3", Name: "P3", IsAlive: true},  // imp (evil)
		{ID: "p4", Name: "P4", IsAlive: true},  // washerwoman
		{ID: "p5", Name: "P5", IsAlive: true},  // librarian
	})

	setupAndStartGame(t, gs, map[string]string{
		"p1": "poisoner",
		"p2": "empath",
		"p3": "imp",
		"p4": "washerwoman",
		"p5": "librarian",
	})

	// Mark p1 as dead
	gs.players[0].IsAlive = false

	skipToCharacter(t, gs, "empath")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilNeighbors),
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	// Empath (p2) left neighbor p1 is dead, right neighbor p3 (imp) is alive and evil -> "1"
	if event.Result == nil || *event.Result != "1" {
		t.Errorf("expected result '1' (only right evil neighbor), got %v", event.Result)
	}
}

// Test helpers

func setupAndStartGame(t *testing.T, gs *GameSession, assignments map[string]string) {
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

func skipToCharacter(t *testing.T, gs *GameSession, targetCharID string) {
	t.Helper()

	// Submit dummy actions until we reach the target character
	for {
		step := gs.currentNightWakeStepLocked()
		if step == nil {
			t.Fatal("no more wake steps")
		}
		if step.CharacterID == targetCharID {
			return
		}

		// Submit a dummy action for this step
		var targets []string
		if step.MinTargets > 0 {
			// Provide distinct dummy targets based on required count
			playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}
			for i := 0; i < step.MinTargets && i < len(playerIDs); i++ {
				targets = append(targets, playerIDs[i])
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
