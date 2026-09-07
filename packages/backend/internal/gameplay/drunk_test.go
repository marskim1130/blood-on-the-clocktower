package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestAssignCharactersSupportsDrunkShownTownsfolk(t *testing.T) {
	gs := newDrunkAssignmentSession(t)

	result, err := assignDrunkWasherwomanGame(t, gs)
	if err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}

	player := playerByID(t, gs.Players(), "p1")
	if player.Character == nil || player.Character.ID != "drunk" {
		t.Fatalf("expected p1 to actually be Drunk, got %#v", player.Character)
	}
	if player.ShownCharacter == nil || player.ShownCharacter.ID != "washerwoman" {
		t.Fatalf("expected p1 to be shown Washerwoman, got %#v", player.ShownCharacter)
	}

	found := false
	for _, event := range result.Events {
		if event.CharacterAssigned == nil || event.CharacterAssigned.PlayerID != "p1" {
			continue
		}
		found = true
		if event.CharacterAssigned.Character.ID != "drunk" {
			t.Fatalf("expected storyteller assignment event to keep actual Drunk, got %#v", event.CharacterAssigned.Character)
		}
		if event.CharacterAssigned.ShownCharacter == nil || event.CharacterAssigned.ShownCharacter.ID != "washerwoman" {
			t.Fatalf("expected assignment event to include shown Washerwoman, got %#v", event.CharacterAssigned.ShownCharacter)
		}
	}
	if !found {
		t.Fatal("expected character assignment event for p1")
	}
}

func TestAssignCharactersRejectsDrunkWithoutShownTownsfolk(t *testing.T) {
	gs := newDrunkAssignmentSession(t)

	markAllPlayersReady(gs)
	_, err := gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "drunk",
			"p2": "saint",
			"p3": "chef",
			"p4": "baron",
			"p5": "imp",
		},
	})
	if err == nil {
		t.Fatal("expected Drunk assignment without shown Townsfolk to be rejected")
	}
	if err.Error() != "Drunk player p1 requires a shown Townsfolk character" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssignCharactersRejectsDrunkShownAssignedCharacter(t *testing.T) {
	gs := newDrunkAssignmentSession(t)

	markAllPlayersReady(gs)
	_, err := gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "drunk",
			"p2": "saint",
			"p3": "chef",
			"p4": "baron",
			"p5": "imp",
		},
		ShownCharacters: map[string]string{"p1": "chef"},
	})
	if err == nil {
		t.Fatal("expected Drunk shown assigned Townsfolk to be rejected")
	}
	if err.Error() != "Drunk shown character chef is already assigned" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrunkVisibilityUsesShownCharacterForDrunkOnly(t *testing.T) {
	gs := newDrunkAssignmentSession(t)
	if _, err := assignDrunkWasherwomanGame(t, gs); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}

	storytellerState := gs.ProjectionFor("storyteller")
	storytellerView := playerByID(t, storytellerState.Players, "p1")
	if storytellerView.Character == nil || storytellerView.Character.ID != "drunk" {
		t.Fatalf("expected storyteller to see actual Drunk, got %#v", storytellerView.Character)
	}
	if storytellerView.ShownCharacter == nil || storytellerView.ShownCharacter.ID != "washerwoman" {
		t.Fatalf("expected storyteller to see shown Washerwoman, got %#v", storytellerView.ShownCharacter)
	}

	drunkState := gs.ProjectionFor("p1")
	drunkView := playerByID(t, drunkState.Players, "p1")
	if drunkView.Character == nil || drunkView.Character.ID != "washerwoman" {
		t.Fatalf("expected Drunk player to see Washerwoman, got %#v", drunkView.Character)
	}
	if drunkView.ShownCharacter != nil {
		t.Fatalf("expected Drunk player not to see shownCharacter metadata, got %#v", drunkView.ShownCharacter)
	}

	otherState := gs.ProjectionFor("p2")
	hiddenView := playerByID(t, otherState.Players, "p1")
	if hiddenView.Character != nil || hiddenView.ShownCharacter != nil {
		t.Fatalf("expected other players not to see Drunk identity, got character=%#v shown=%#v", hiddenView.Character, hiddenView.ShownCharacter)
	}
}

func TestDrunkShownInformationRoleWakesButDoesNotAutoResolve(t *testing.T) {
	gs := newStartedDrunkWasherwomanGame(t)

	skipNightWakeStepsUntilAction(t, gs, game.NightActionLearnTownsfolk)
	step := gs.currentNightWakeStepLocked()
	if step == nil || step.CharacterID != "washerwoman" {
		t.Fatalf("expected Drunk shown Washerwoman wake step, got %#v", step)
	}

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnTownsfolk),
		TargetIDs:  []string{"p1", "p2"},
	})
	if err != nil {
		t.Fatalf("SubmitNightAction failed: %v", err)
	}
	event := result.Events[0].NightActionSubmitted
	if event == nil {
		t.Fatalf("expected night action event, got %#v", result.Events)
	}
	if event.Result != nil {
		t.Fatalf("expected Drunk shown information role to have no auto result, got %v", *event.Result)
	}
}

func TestStartGameRejectsDrunkWithoutShownCharacter(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID, gs.storytellerName = "storyteller", "Storyteller"
	gs.SetPlayers([]game.Player{
		{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "drunk")},
		{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "saint")},
		{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "chef")},
		{ID: "p4", Name: "P4", IsAlive: true, Character: testCharacter(t, "baron")},
		{ID: "p5", Name: "P5", IsAlive: true, Character: testCharacter(t, "imp")},
	})

	markAllPlayersConfirmed(gs)
	_, err := gs.Apply(StartGameCmd{SenderID: "storyteller"})
	if err == nil {
		t.Fatal("expected start game with Drunk missing shown character to be rejected")
	}
	if err.Error() != "Drunk player p1 requires a shown Townsfolk character" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newStartedDrunkWasherwomanGame(t *testing.T) *GameSession {
	t.Helper()

	gs := newDrunkAssignmentSession(t)
	if _, err := assignDrunkWasherwomanGame(t, gs); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}
	markAllPlayersConfirmed(gs)
	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}
	return gs
}

func newDrunkAssignmentSession(t *testing.T) *GameSession {
	t.Helper()

	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "storyteller", Name: "Storyteller", IsAlive: true})
	for _, playerID := range []string{"p1", "p2", "p3", "p4", "p5"} {
		gs.AddPlayer(game.Player{ID: playerID, Name: playerID, IsAlive: true})
	}
	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}
	return gs
}

func assignDrunkWasherwomanGame(t *testing.T, gs *GameSession) (ApplyResult, error) {
	t.Helper()

	markAllPlayersReady(gs)
	return gs.Apply(AssignCharactersCmd{
		SenderID: "storyteller",
		Assignments: map[string]string{
			"p1": "drunk",
			"p2": "saint",
			"p3": "chef",
			"p4": "baron",
			"p5": "imp",
		},
		ShownCharacters: map[string]string{"p1": "washerwoman"},
	})
}

func playerByID(t *testing.T, players []game.Player, playerID string) game.Player {
	t.Helper()

	for _, player := range players {
		if player.ID == playerID {
			return player
		}
	}
	t.Fatalf("player %s not found in %#v", playerID, players)
	return game.Player{}
}
