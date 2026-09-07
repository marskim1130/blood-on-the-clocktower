package gameplay

import (
	"fmt"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func (gs *GameSession) Restart(memberIDs []string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if gs.phase != game.GamePhaseFinished {
		return fmt.Errorf("game must be finished before restarting")
	}
	members := map[string]bool{}
	for _, id := range memberIDs {
		members[id] = true
	}
	next := NewGameSession(gs.scriptID)
	if members[gs.storytellerID] {
		next.storytellerID = gs.storytellerID
		next.storytellerName = gs.storytellerName
	}
	for _, p := range gs.players {
		if members[p.ID] {
			next.players = append(next.players, game.Player{ID: p.ID, Name: p.Name, IsAlive: true})
		}
	}
	// 逐项清理游戏状态，不复制互斥锁；增加游戏字段时必须同步更新这里及重赛回归测试。
	gs.players = next.players
	gs.storytellerID = next.storytellerID
	gs.storytellerName = next.storytellerName
	gs.originalPlayers = 0
	gs.phase = game.GamePhaseSetup
	gs.dayNumber = 0
	gs.nightNumber = 0
	gs.nightWakeIndex = 0
	gs.nomination = nil
	gs.nominationResults = nil
	gs.nightActions = nil
	gs.pendingNightAction = nil
	gs.confirmedNightAction = nil
	gs.nightAcknowledged = nil
	gs.dawnReviewPending = false
	gs.pendingDawnDeathIDs = nil
	gs.deaths = nil
	gs.ghostVotesUsed = next.ghostVotesUsed
	gs.slayerUsed = next.slayerUsed
	gs.nominatorsToday = next.nominatorsToday
	gs.nomineesToday = next.nomineesToday
	gs.virginAbilityUsed = next.virginAbilityUsed
	gs.butlerMasters = next.butlerMasters
	gs.fortuneTellerRedHerringID = ""
	gs.demonBluffCharacterIDs = nil
	gs.grimoireRevealed = false
	gs.winner = nil
	return nil
}
