package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func skipNightWakeStepsUntilAction(t *testing.T, gs *GameSession, actionType game.NightActionType) {
	t.Helper()

	for {
		step := gs.currentNightWakeStepLocked()
		if step == nil {
			t.Fatalf("no remaining wake steps before %s", actionType)
		}
		if step.ActionType == actionType {
			return
		}
		submitDefaultNightWakeStep(t, gs, step)
	}
}

func submitDefaultNightWakeStep(t *testing.T, gs *GameSession, step *game.NightWakeStep) {
	t.Helper()

	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(step.ActionType),
		TargetIDs:  defaultNightTargetsForStep(gs, step),
	}); err != nil {
		t.Fatalf("skip night wake step %s failed: %v", step.ActionType, err)
	}
}

func defaultNightTargetsForStep(gs *GameSession, step *game.NightWakeStep) []string {
	if step.MinTargets == 0 {
		return nil
	}

	targets := make([]string, 0, step.MinTargets)
	for _, player := range gs.players {
		if player.ID == "storyteller" || !player.IsAlive || (step.CharacterID != "" && player.Character != nil && player.Character.ID == step.CharacterID) {
			continue
		}
		targets = append(targets, player.ID)
		if len(targets) == step.MinTargets {
			return targets
		}
	}
	return targets
}
