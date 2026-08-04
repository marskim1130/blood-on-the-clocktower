package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestCloneNightActionsUsesEmptyTargetArrays(t *testing.T) {
	cloned := cloneNightActions([]game.NightAction{{ActionType: game.NightActionLearnDemon}})

	if len(cloned) != 1 || cloned[0].TargetIDs == nil || len(cloned[0].TargetIDs) != 0 {
		t.Fatalf("empty night targets must project as [], got %#v", cloned)
	}
}
