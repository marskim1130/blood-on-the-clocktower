package gameplay

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestNightDeathIsAttributedToUpcomingDay(t *testing.T) {
	gs := setupPoisonedGameSession(t)
	completeFirstNight(t, gs)
	if state := gs.Projection(); state.DayNumber != 1 || state.NightNumber != 1 {
		t.Fatalf("expected first day after first night, got day=%d night=%d", state.DayNumber, state.NightNumber)
	}
	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err != nil {
		t.Fatalf("enter second night: %v", err)
	}

	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "storyteller", ActionType: string(game.NightActionPoison), TargetIDs: []string{"p2"}}); err != nil {
		t.Fatalf("submit poison action: %v", err)
	}
	skipNightWakeStepsUntilAction(t, gs, game.NightActionKill)
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "storyteller", ActionType: string(game.NightActionKill), TargetIDs: []string{"p1"}}); err != nil {
		t.Fatalf("submit kill action: %v", err)
	}
	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("resolve second night: %v", err)
	}

	state := gs.Projection()
	if state.DayNumber != 2 || state.NightNumber != 2 {
		t.Fatalf("expected second day after second night, got day=%d night=%d", state.DayNumber, state.NightNumber)
	}
	if len(state.Deaths) != 1 || state.Deaths[0].PlayerID != "p1" || state.Deaths[0].DayNumber != 2 {
		t.Fatalf("night death must belong to upcoming day 2, got %#v", state.Deaths)
	}
}

func TestStorytellerMetadataAndRedHerringSurviveSnapshotWithoutPrivacyLeak(t *testing.T) {
	gs := newStartedFortuneTellerGameWithRedHerring(t, "p2")
	restored := newGameSessionFromSnapshot(gs.snapshot())

	storytellerState := restored.ProjectionFor("storyteller")
	if storytellerState.StorytellerID != "storyteller" || storytellerState.StorytellerName != "Storyteller" {
		t.Fatalf("storyteller metadata did not survive snapshot: %#v", storytellerState)
	}
	if storytellerState.DayNumber != 0 || storytellerState.NightNumber != 1 {
		t.Fatalf("first-night counters did not survive snapshot: day=%d night=%d", storytellerState.DayNumber, storytellerState.NightNumber)
	}
	if storytellerState.FortuneTellerRedHerringID != "p2" {
		t.Fatalf("storyteller lost red herring after restore: %#v", storytellerState)
	}

	playerState := restored.ProjectionFor("p1")
	if playerState.FortuneTellerRedHerringID != "" {
		t.Fatalf("red herring leaked to player after restore: %#v", playerState)
	}
}
