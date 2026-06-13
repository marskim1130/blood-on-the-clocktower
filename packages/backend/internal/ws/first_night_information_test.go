package ws

import (
	"fmt"
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestWasherwomanAutoComputesTownsfolkNameFromTargets(t *testing.T) {
	gs := newStartedTypeHintGame(t, "washerwoman", "empath", "chef", "poisoner", "imp")
	skipTypeHintGameToCharacter(t, gs, "washerwoman")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnTownsfolk, []string{"p2", "p4"}, "")
	if event.Result == nil || *event.Result != "Empath" {
		t.Fatalf("expected Washerwoman result 'Empath', got %v", event.Result)
	}
}

func TestPoisonedWasherwomanDoesNotAutoCompute(t *testing.T) {
	gs := newStartedTypeHintGame(t, "washerwoman", "empath", "chef", "poisoner", "imp")

	event := submitTypeHintNightAction(t, gs, game.NightActionPoison, []string{"p1"}, "")
	if event.Result != nil {
		t.Fatalf("expected poison action to have no result, got %v", event.Result)
	}

	event = submitTypeHintNightAction(t, gs, game.NightActionLearnTownsfolk, []string{"p2", "p4"}, "")
	if event.Result != nil && *event.Result != "" {
		t.Fatalf("expected poisoned Washerwoman to have no auto result, got %v", event.Result)
	}
}

func TestWasherwomanManualResultOverridesAutoCompute(t *testing.T) {
	gs := newStartedTypeHintGame(t, "washerwoman", "empath", "chef", "poisoner", "imp")
	skipTypeHintGameToCharacter(t, gs, "washerwoman")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnTownsfolk, []string{"p2", "p4"}, "Chef")
	if event.Result == nil || *event.Result != "Chef" {
		t.Fatalf("expected manual Washerwoman result 'Chef', got %v", event.Result)
	}
}

func TestLibrarianAutoComputesOutsiderNameFromTargets(t *testing.T) {
	gs := newStartedTypeHintGame(t, "librarian", "butler", "washerwoman", "chef", "poisoner", "imp")
	skipTypeHintGameToCharacter(t, gs, "librarian")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnOutsider, []string{"p2", "p5"}, "")
	if event.Result == nil || *event.Result != "Butler" {
		t.Fatalf("expected Librarian result 'Butler', got %v", event.Result)
	}
}

func TestLibrarianAutoComputesNoneWhenNoOutsidersAreInPlay(t *testing.T) {
	gs := newStartedTypeHintGame(t, "librarian", "empath", "chef", "poisoner", "imp")
	skipTypeHintGameToCharacter(t, gs, "librarian")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnOutsider, nil, "")
	if event.Result == nil || *event.Result != "none" {
		t.Fatalf("expected Librarian result 'none', got %v", event.Result)
	}
}

func TestInvestigatorAutoComputesMinionNameFromTargets(t *testing.T) {
	gs := newStartedTypeHintGame(t, "investigator", "washerwoman", "chef", "poisoner", "imp")
	skipTypeHintGameToCharacter(t, gs, "investigator")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnMinion, []string{"p2", "p4"}, "")
	if event.Result == nil || *event.Result != "Poisoner" {
		t.Fatalf("expected Investigator result 'Poisoner', got %v", event.Result)
	}
}

func TestPoisonedInvestigatorDoesNotAutoCompute(t *testing.T) {
	gs := newStartedTypeHintGame(t, "investigator", "washerwoman", "chef", "poisoner", "imp")

	submitTypeHintNightAction(t, gs, game.NightActionPoison, []string{"p1"}, "")
	skipTypeHintGameToCharacter(t, gs, "investigator")

	event := submitTypeHintNightAction(t, gs, game.NightActionLearnMinion, []string{"p2", "p4"}, "")
	if event.Result != nil && *event.Result != "" {
		t.Fatalf("expected poisoned Investigator to have no auto result, got %v", event.Result)
	}
}

func newStartedTypeHintGame(t *testing.T, characterIDs ...string) *GameSession {
	t.Helper()

	gs := NewGameSession()
	players := []game.Player{{ID: "storyteller", Name: "Storyteller", IsAlive: true}}
	assignments := make(map[string]string, len(characterIDs))
	for i, characterID := range characterIDs {
		playerID := fmt.Sprintf("p%d", i+1)
		players = append(players, game.Player{ID: playerID, Name: fmt.Sprintf("P%d", i+1), IsAlive: true})
		assignments[playerID] = characterID
	}

	gs.SetPlayers(players)
	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}
	if _, err := gs.Apply(AssignCharactersCmd{SenderID: "storyteller", Assignments: assignments}); err != nil {
		t.Fatalf("AssignCharacters failed: %v", err)
	}
	if _, err := gs.Apply(StartGameCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}
	return gs
}

func skipTypeHintGameToCharacter(t *testing.T, gs *GameSession, targetCharID string) {
	t.Helper()

	for {
		step := gs.currentNightWakeStepLocked()
		if step == nil {
			t.Fatal("no remaining wake steps")
		}
		if step.CharacterID == targetCharID {
			return
		}
		submitTypeHintNightAction(t, gs, step.ActionType, typeHintTargetsForStep(gs, step, targetCharID), "")
	}
}

func typeHintTargetsForStep(gs *GameSession, step *game.NightWakeStep, targetCharID string) []string {
	if step.MinTargets == 0 {
		return nil
	}
	targets := make([]string, 0, step.MinTargets)
	for _, player := range gs.players {
		if player.Character != nil && player.Character.ID == targetCharID {
			continue
		}
		targets = append(targets, player.ID)
		if len(targets) == step.MinTargets {
			return targets
		}
	}
	return targets
}

func submitTypeHintNightAction(t *testing.T, gs *GameSession, actionType game.NightActionType, targetIDs []string, result string) *game.NightActionEvent {
	t.Helper()

	applyResult, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(actionType),
		TargetIDs:  targetIDs,
		Result:     result,
	})
	if err != nil {
		t.Fatalf("SubmitNightAction %s failed: %v", actionType, err)
	}
	if len(applyResult.Events) != 1 || applyResult.Events[0].NightActionSubmitted == nil {
		t.Fatalf("expected NightActionSubmitted event, got %#v", applyResult.Events)
	}
	return applyResult.Events[0].NightActionSubmitted
}
