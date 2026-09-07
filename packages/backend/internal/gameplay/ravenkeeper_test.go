package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestRavenkeeperAutoComputesChosenPlayerCharacterWhenKilledTonight(t *testing.T) {
	gs := newStartedRavenkeeperGame(t)
	enterRavenkeeperSecondNight(t, gs)
	submitRavenkeeperNightPrefix(t, gs, "p2", "p4")
	submitRavenkeeperImpKill(t, gs, "p1")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnDied),
		TargetIDs:  []string{"p5"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "Imp" {
		t.Fatalf("expected Ravenkeeper result 'Imp', got %v", event.Result)
	}
}

func TestRavenkeeperIsNotWokenWhenNotKilledTonight(t *testing.T) {
	gs := newStartedRavenkeeperGame(t)
	enterRavenkeeperSecondNight(t, gs)
	submitRavenkeeperNightPrefix(t, gs, "p2", "p4")
	submitRavenkeeperImpKill(t, gs, "p2")

	_, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnDied),
		TargetIDs:  []string{"p5"},
	})
	if err == nil {
		t.Fatal("surviving Ravenkeeper must not receive a night action")
	}
}

func TestProtectedRavenkeeperDoesNotAutoRevealCharacter(t *testing.T) {
	gs := newStartedRavenkeeperGame(t)
	enterRavenkeeperSecondNight(t, gs)
	submitRavenkeeperNightPrefix(t, gs, "p3", "p1")
	submitRavenkeeperImpKill(t, gs, "p1")

	_, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnDied),
		TargetIDs:  []string{"p5"},
	})
	if err == nil {
		t.Fatal("protected Ravenkeeper must not receive a night action")
	}
}

func TestPoisonedMonkDoesNotProtectRavenkeeper(t *testing.T) {
	gs := newStartedRavenkeeperGame(t)
	enterRavenkeeperSecondNight(t, gs)
	submitRavenkeeperNightPrefix(t, gs, "p2", "p1")
	submitRavenkeeperImpKill(t, gs, "p1")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnDied),
		TargetIDs:  []string{"p5"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "Imp" {
		t.Fatalf("expected poisoned Monk not to protect Ravenkeeper, got %v", event.Result)
	}
}

func TestPoisonedRavenkeeperDoesNotAutoCompute(t *testing.T) {
	gs := newStartedRavenkeeperGame(t)
	enterRavenkeeperSecondNight(t, gs)
	submitRavenkeeperNightPrefix(t, gs, "p1", "p4")
	submitRavenkeeperImpKill(t, gs, "p1")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnDied),
		TargetIDs:  []string{"p5"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result != nil && *event.Result != "" {
		t.Fatalf("expected poisoned Ravenkeeper to have no auto result, got %v", event.Result)
	}
}

func TestRavenkeeperManualResultOverridesAutoCompute(t *testing.T) {
	gs := newStartedRavenkeeperGame(t)
	enterRavenkeeperSecondNight(t, gs)
	submitRavenkeeperNightPrefix(t, gs, "p2", "p4")
	submitRavenkeeperImpKill(t, gs, "p1")

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnDied),
		TargetIDs:  []string{"p5"},
		Result:     "Washerwoman",
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	event := result.Events[0].NightActionSubmitted
	if event.Result == nil || *event.Result != "Washerwoman" {
		t.Fatalf("expected manual Ravenkeeper result 'Washerwoman', got %v", event.Result)
	}
}

func newStartedRavenkeeperGame(t *testing.T) *GameSession {
	t.Helper()

	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true}, // Ravenkeeper
		{ID: "p2", Name: "P2", IsAlive: true}, // Monk
		{ID: "p3", Name: "P3", IsAlive: true}, // Empath
		{ID: "p4", Name: "P4", IsAlive: true}, // Poisoner
		{ID: "p5", Name: "P5", IsAlive: true}, // Imp
	})

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}
	markAllPlayersReady(gs)
	if _, err := gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "ravenkeeper",
			"p2": "monk",
			"p3": "empath",
			"p4": "poisoner",
			"p5": "imp",
		},
	}); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}
	markAllPlayersConfirmed(gs)
	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	return gs
}

func enterRavenkeeperSecondNight(t *testing.T, gs *GameSession) {
	t.Helper()

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	actions := []SubmitNightActionCmd{
		{SenderID: "storyteller", ActionType: string(game.NightActionPoison), TargetIDs: []string{"p2"}},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnEvilNeighbors)},
	}
	for _, action := range actions {
		if _, err := gs.Apply(action); err != nil {
			t.Fatalf("SubmitNightAction failed: %v", err)
		}
	}
	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	if _, err := gs.Apply(FinalizeDayCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("ChangePhase failed: %v", err)
	}
}

func submitRavenkeeperNightPrefix(t *testing.T, gs *GameSession, poisonTargetID, monkProtectTargetID string) {
	t.Helper()

	actions := []SubmitNightActionCmd{
		{SenderID: "storyteller", ActionType: string(game.NightActionPoison), TargetIDs: []string{poisonTargetID}},
		{SenderID: "storyteller", ActionType: string(game.NightActionProtect), TargetIDs: []string{monkProtectTargetID}},
	}
	for _, action := range actions {
		if _, err := gs.Apply(action); err != nil {
			t.Fatalf("SubmitNightAction failed: %v", err)
		}
	}
}

func submitRavenkeeperImpKill(t *testing.T, gs *GameSession, targetID string) {
	t.Helper()

	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionKill),
		TargetIDs:  []string{targetID},
	}); err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnEvilNeighbors),
	}); err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}
	if _, err := gs.Apply(PrepareDawnCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	deaths := gs.ProjectionFor("storyteller").PendingDawnDeathIDs
	if _, err := gs.Apply(ConfirmDawnCmd{SenderID: "storyteller", DeathPlayerIDs: deaths}); err != nil {
		t.Fatal(err)
	}
}
