package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

// TestPoisonPlayerSetsPoisonedUntil verifies that submitting a poison action sets PoisonedUntil field
func TestPoisonPlayerSetsPoisonedUntil(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true},
		{ID: "p2", Name: "P2", IsAlive: true},
		{ID: "p3", Name: "P3", IsAlive: true},
		{ID: "p4", Name: "P4", IsAlive: true},
		{ID: "p5", Name: "P5", IsAlive: true},
	})

	_, err := gs.Apply(SetStorytellerCmd{
		SenderID:       "storyteller",
		TargetPlayerID: "storyteller",
	})
	if err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}

	_, err = gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "librarian",
			"p3": "investigator",
			"p4": "poisoner",
			"p5": "imp",
		},
	})
	if err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}

	_, err = gs.Apply(StartGameCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	// Storyteller poisons p2
	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p2"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}

	if !result.Updated {
		t.Fatal("expected state to be updated after poison action")
	}

	// Verify p2 is poisoned through day 1. The first night uses dayNumber=0.
	players := gs.Players()
	var p2 *game.Player
	for i := range players {
		if players[i].ID == "p2" {
			p2 = &players[i]
			break
		}
	}

	if p2 == nil {
		t.Fatal("p2 not found in player list")
	}

	if p2.PoisonedUntil == nil {
		t.Fatal("expected p2 to be poisoned, but PoisonedUntil is nil")
	}

	expectedExpiration := int32(1)
	if *p2.PoisonedUntil != expectedExpiration {
		t.Errorf("expected PoisonedUntil=%d, got %d", expectedExpiration, *p2.PoisonedUntil)
	}
}

func TestPoisonedImpNightKillDoesNotKillTarget(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true},
		{ID: "p2", Name: "P2", IsAlive: true},
		{ID: "p3", Name: "P3", IsAlive: true},
		{ID: "p4", Name: "P4", IsAlive: true},
		{ID: "p5", Name: "P5", IsAlive: true},
	})

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}
	if _, err := gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "librarian",
			"p3": "investigator",
			"p4": "poisoner",
			"p5": "imp",
		},
	}); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}
	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	completeFullFirstNight(t, gs)
	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err != nil {
		t.Fatalf("ChangePhase to second night failed: %v", err)
	}

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p5"},
	}); err != nil {
		t.Fatalf("Poison Imp failed: %v", err)
	}

	skipNightWakeStepsUntilAction(t, gs, game.NightActionKill)
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionKill),
		TargetIDs:  []string{"p1"},
	}); err != nil {
		t.Fatalf("Imp kill failed: %v", err)
	}

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	for _, event := range result.Events {
		if event.PlayerDied != nil {
			t.Fatalf("expected poisoned Imp kill not to kill anyone, got %#v", event.PlayerDied)
		}
	}
	p1 := findPlayerInState(t, gs.Projection(), "p1")
	if !p1.IsAlive {
		t.Fatal("expected target to survive poisoned Imp kill")
	}
	if deaths := gs.Deaths(); len(deaths) != 0 {
		t.Fatalf("expected no death records from poisoned Imp kill, got %#v", deaths)
	}
}

func completeFullFirstNight(t *testing.T, gs *GameSession) {
	t.Helper()

	actions := []SubmitNightActionCmd{
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnDemon)},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnMinion)},
		{SenderID: "storyteller", ActionType: string(game.NightActionPoison), TargetIDs: []string{"p1"}},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnTownsfolk), TargetIDs: []string{"p1", "p2"}},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnOutsider)},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnMinion), TargetIDs: []string{"p4", "p5"}},
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

// TestPoisonExpiresAtDusk verifies that poison status clears at dusk (Day→Night transition)
func TestPoisonExpiresAtDusk(t *testing.T) {
	gs := setupPoisonedGameSession(t)

	// p2 is poisoned through day 1
	// Fast-forward through first night (ResolveNight automatically transitions to Day)
	completeFirstNight(t, gs)

	// Verify we're in day phase and p2 is still poisoned
	if gs.Phase() != game.GamePhaseDay {
		t.Fatalf("expected to be in Day phase after ResolveNight, got %v", gs.Phase())
	}

	players := gs.Players()
	p2 := findPlayer(players, "p2")
	if p2.PoisonedUntil == nil || *p2.PoisonedUntil != 1 {
		t.Fatal("expected p2 to still be poisoned during day 1")
	}

	// Transition to Night 2 (Day→Night = dusk)
	_, err := gs.Apply(ChangePhaseCmd{
		SenderID: "storyteller",
		Phase:    game.GamePhaseNight,
	})
	if err != nil {
		t.Fatalf("ChangePhase to Night failed: %v", err)
	}

	// Verify p2's poison has expired
	players = gs.Players()
	p2 = findPlayer(players, "p2")
	if p2.PoisonedUntil != nil {
		t.Errorf("expected p2's poison to expire at dusk, but PoisonedUntil=%d", *p2.PoisonedUntil)
	}
}

// TestStorytellerCanSeePoisonedState verifies storyteller can see PoisonedUntil field
func TestStorytellerCanSeePoisonedState(t *testing.T) {
	gs := setupPoisonedGameSession(t)

	state := gs.ProjectionFor("storyteller")

	p2 := findPlayer(state.Players, "p2")
	if p2.PoisonedUntil == nil {
		t.Fatal("expected storyteller to see p2's PoisonedUntil field")
	}

	if *p2.PoisonedUntil != 1 {
		t.Errorf("expected storyteller to see PoisonedUntil=1, got %d", *p2.PoisonedUntil)
	}
}

// TestPlayerCannotSeePoisonedState verifies players cannot see poison status (including their own)
func TestPlayerCannotSeePoisonedState(t *testing.T) {
	gs := setupPoisonedGameSession(t)

	// Test p1 (not poisoned) cannot see anyone's poison status
	state := gs.ProjectionFor("p1")
	for _, player := range state.Players {
		if player.PoisonedUntil != nil {
			t.Errorf("expected p1 not to see any PoisonedUntil fields, but saw %s with PoisonedUntil=%d",
				player.ID, *player.PoisonedUntil)
		}
	}

	// Test p2 (poisoned) cannot see their own poison status
	state = gs.ProjectionFor("p2")
	p2 := findPlayer(state.Players, "p2")
	if p2.PoisonedUntil != nil {
		t.Errorf("expected p2 not to see their own PoisonedUntil field, got %d", *p2.PoisonedUntil)
	}
}

// TestRepoisonUpdatesExpiration verifies that poisoning the same player again updates the expiration
func TestRepoisonUpdatesExpiration(t *testing.T) {
	gs := setupPoisonedGameSession(t)

	// Complete first night (transitions to Day 1)
	completeFirstNight(t, gs)

	// Enter night 2 (p2's poison expires at dusk)
	_, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight})
	if err != nil {
		t.Fatalf("ChangePhase to Night 2 failed: %v", err)
	}

	// Verify p2's poison has expired
	players := gs.Players()
	p2 := findPlayer(players, "p2")
	if p2.PoisonedUntil != nil {
		t.Fatal("expected p2's poison to be cleared at start of night 2")
	}

	// Poisoner acts again, re-poisons p2
	_, err = gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p2"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction (re-poison) failed: %v", err)
	}

	// Verify p2 is now poisoned until day 3
	players = gs.Players()
	p2 = findPlayer(players, "p2")
	if p2.PoisonedUntil == nil {
		t.Fatal("expected p2 to be re-poisoned, but PoisonedUntil is nil")
	}

	expectedExpiration := int32(2)
	if *p2.PoisonedUntil != expectedExpiration {
		t.Errorf("expected re-poisoned PoisonedUntil=%d, got %d", expectedExpiration, *p2.PoisonedUntil)
	}
}

// Helper: setup a game session with p2 poisoned on night 1
func setupPoisonedGameSession(t *testing.T) *GameSession {
	t.Helper()

	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true},
		{ID: "p2", Name: "P2", IsAlive: true},
		{ID: "p3", Name: "P3", IsAlive: true},
		{ID: "p4", Name: "P4", IsAlive: true},
		{ID: "p5", Name: "P5", IsAlive: true},
	})

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}

	if _, err := gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "washerwoman",
			"p2": "librarian",
			"p3": "investigator",
			"p4": "poisoner",
			"p5": "imp",
		},
	}); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}

	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	// Poisoner poisons p2
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"p2"},
	}); err != nil {
		t.Fatalf("SubmitNightAction (poison) failed: %v", err)
	}

	return gs
}

// Helper: complete first night wake order
func completeFirstNight(t *testing.T, gs *GameSession) {
	t.Helper()

	// Remaining first night actions: Washerwoman, Librarian, Investigator
	actions := []SubmitNightActionCmd{
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnTownsfolk), TargetIDs: []string{"p1", "p2"}},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnOutsider)},
		{SenderID: "storyteller", ActionType: string(game.NightActionLearnMinion), TargetIDs: []string{"p4", "p5"}},
	}

	for _, action := range actions {
		if _, err := gs.Apply(action); err != nil {
			t.Fatalf("SubmitNightAction failed: %v", err)
		}
	}

	// Resolve night
	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
}

// Helper: find player by ID in slice
func findPlayer(players []game.Player, id string) game.Player {
	for _, p := range players {
		if p.ID == id {
			return p
		}
	}
	panic("player not found: " + id)
}
