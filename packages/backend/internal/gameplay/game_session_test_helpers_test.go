package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
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

func markAllPlayersReady(gs *GameSession) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	for index := range gs.players {
		gs.players[index].IsReady = true
	}
}

func openNominationVoting(t *testing.T, gs *GameSession) {
	t.Helper()
	for gs.nomination != nil && (gs.nomination.Stage == game.NominationStageAccusation || gs.nomination.Stage == game.NominationStageDefense) {
		if _, err := gs.Apply(AdvanceNominationStageCmd{SenderID: gs.storytellerID}); err != nil {
			t.Fatal(err)
		}
	}
}

func markAllPlayersConfirmed(gs *GameSession) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	for index := range gs.players {
		gs.players[index].HasConfirmedCharacter = true
	}
}
