package gameplay

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func findPlayerInState(t *testing.T, state *Projection, playerID string) game.Player {
	t.Helper()
	for _, player := range state.Players {
		if player.ID == playerID {
			return player
		}
	}
	t.Fatalf("expected player %s in room state", playerID)
	return game.Player{}
}
