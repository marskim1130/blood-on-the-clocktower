package gameplay

import (
	"fmt"
	"strings"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func (gs *GameSession) applyPublishGrimoire(cmd PublishGrimoireCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can publish the grimoire")
	}
	if gs.phase != game.GamePhaseFinished || gs.winner == nil {
		return ApplyResult{}, fmt.Errorf("the grimoire can only be published after the game ends")
	}
	if gs.grimoireRevealed {
		return ApplyResult{}, nil
	}
	gs.grimoireRevealed = true
	return ApplyResult{Updated: true}, nil
}

// ────────────────────────────────────────────────
// EndGameCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyEndGame(cmd EndGameCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can end the game")
	}
	if gs.phase == game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("game cannot be ended before it starts")
	}
	if gs.phase == game.GamePhaseFinished {
		return ApplyResult{}, fmt.Errorf("game is already finished")
	}
	if cmd.Winner != game.TeamGood && cmd.Winner != game.TeamEvil {
		return ApplyResult{}, fmt.Errorf("winner must be good or evil")
	}

	reason := cmd.Reason
	if reason == "" {
		reason = game.WinReasonStorytellerDecision
	}
	description := strings.TrimSpace(cmd.Description)
	if description == "" {
		description = "Storyteller ended the game."
	}

	ended := &game.GameEndedEvent{
		Winner:      cmd.Winner,
		Reason:      reason,
		Description: description,
	}
	gs.winner = ended
	gs.phase = game.GamePhaseFinished

	return ApplyResult{
		Events:  []game.GameEvent{{GameEnded: ended}},
		Updated: true,
	}, nil
}

// ────────────────────────────────────────────────
// KillPlayerCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyKillPlayer(cmd KillPlayerCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can kill players")
	}
	if gs.phase == game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("game cannot kill players before it starts")
	}
	if gs.phase == game.GamePhaseFinished {
		return ApplyResult{}, fmt.Errorf("game is already finished")
	}
	if cmd.PlayerID == "" {
		return ApplyResult{}, fmt.Errorf("target player is required")
	}
	if cmd.Cause == "" {
		return ApplyResult{}, fmt.Errorf("death cause is required")
	}
	if !isSupportedDeathCause(cmd.Cause) {
		return ApplyResult{}, fmt.Errorf("unsupported death cause %s", cmd.Cause)
	}

	pIdx := gs.findPlayerIndex(cmd.PlayerID)
	if pIdx == -1 {
		return ApplyResult{}, fmt.Errorf("player %s not found", cmd.PlayerID)
	}
	if !gs.players[pIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("player %s is already dead", cmd.PlayerID)
	}

	aliveBeforeDeath := gs.alivePlayerCountLocked()
	demonDeathAliveCount := gs.demonDeathAliveCountLocked(pIdx, aliveBeforeDeath)
	gs.players[pIdx].IsAlive = false
	gs.deaths = append(gs.deaths, game.DeathRecord{
		PlayerID:  cmd.PlayerID,
		Cause:     cmd.Cause,
		DayNumber: gs.dayNumber,
		KilledBy:  cmd.SenderID,
	})

	events := []game.GameEvent{
		{PlayerDied: &game.PlayerDiedEvent{
			PlayerID:  cmd.PlayerID,
			Cause:     cmd.Cause,
			DayNumber: gs.dayNumber,
		}},
	}

	if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func isSupportedDeathCause(cause game.DeathCause) bool {
	switch cause {
	case game.DeathCauseExecution, game.DeathCauseNightKill, game.DeathCauseAbility:
		return true
	default:
		return false
	}
}

// ────────────────────────────────────────────────
// Win condition checking
// ────────────────────────────────────────────────

// checkWinConditions evaluates current state and returns a GameEndedEvent if
// the game should end, or nil if it continues. Must be called with gs.mu held.
func (gs *GameSession) checkWinConditions(demonDeathAliveCount int) *game.GameEndedEvent {
	aliveGood := 0
	aliveEvil := 0
	hasAliveDemon := false

	for _, p := range gs.players {
		if !p.IsAlive {
			continue
		}
		if p.Character == nil {
			continue
		}
		switch p.Character.Team {
		case game.TeamGood:
			aliveGood++
		case game.TeamEvil:
			aliveEvil++
		}
		// Check if this character is a demon by looking up its definition
		charDef := game.GetCharacterByID(p.Character.ID)
		if charDef != nil && charDef.Type == game.CharacterTypeDemon {
			hasAliveDemon = true
		}
	}

	// Check if any executed player was the Saint
	for _, d := range gs.deaths {
		if d.Cause == game.DeathCauseExecution && d.DayNumber == gs.dayNumber {
			pIdx := gs.findPlayerIndex(d.PlayerID)
			if pIdx != -1 && gs.players[pIdx].Character != nil &&
				gs.players[pIdx].Character.ID == "saint" &&
				!gs.playerAbilityMalfunctioningLocked(pIdx) {
				return &game.GameEndedEvent{
					Winner:      game.TeamEvil,
					Reason:      game.WinReasonSaintExecuted,
					Description: "The Saint was executed — evil wins!",
				}
			}
		}
	}

	totalAlive := aliveGood + aliveEvil

	// Demon dead => good wins (by execution)
	if !hasAliveDemon {
		if gs.applyScarletWomanStarpassLocked(demonDeathAliveCount) {
			return nil
		}
		return &game.GameEndedEvent{
			Winner:      game.TeamGood,
			Reason:      game.WinReasonImpExecuted,
			Description: "The Demon is dead — good wins!",
		}
	}

	// Evil wins when the game reaches the final two living players.
	if totalAlive <= 2 && totalAlive > 0 {
		return &game.GameEndedEvent{
			Winner:      game.TeamEvil,
			Reason:      game.WinReasonEvilMajority,
			Description: "Only two players remain alive — evil wins!",
		}
	}

	return nil
}

func (gs *GameSession) mayorEndgameWinnerLocked() *game.GameEndedEvent {
	totalAlive := 0
	mayorIdx := -1
	for i, player := range gs.players {
		if !player.IsAlive {
			continue
		}
		totalAlive++
		if player.Character != nil && player.Character.ID == "mayor" {
			mayorIdx = i
		}
	}
	if totalAlive != 3 || mayorIdx == -1 || gs.playerAbilityMalfunctioningLocked(mayorIdx) {
		return nil
	}
	for _, death := range gs.deaths {
		if death.DayNumber == gs.dayNumber && death.Cause == game.DeathCauseExecution {
			return nil
		}
	}
	return &game.GameEndedEvent{
		Winner:      game.TeamGood,
		Reason:      game.WinReasonMayorEndgame,
		Description: "Only 3 players remain with no execution — Mayor wins for good!",
	}
}

func (gs *GameSession) demonDeathAliveCountLocked(playerIndex int, aliveBeforeDeath int) int {
	if !gs.playerIsDemonLocked(playerIndex) {
		return 0
	}
	return aliveBeforeDeath
}

func (gs *GameSession) playerIsDemonLocked(playerIndex int) bool {
	if playerIndex < 0 || playerIndex >= len(gs.players) || gs.players[playerIndex].Character == nil {
		return false
	}
	charDef := game.GetCharacterByID(gs.players[playerIndex].Character.ID)
	return charDef != nil && charDef.Type == game.CharacterTypeDemon
}

func (gs *GameSession) playerIsImpLocked(playerIndex int) bool {
	return playerIndex >= 0 &&
		playerIndex < len(gs.players) &&
		gs.players[playerIndex].Character != nil &&
		gs.players[playerIndex].Character.ID == "imp"
}

func (gs *GameSession) applyImpSelfStarpassLocked(deadImpIndex int) bool {
	imp := game.GetCharacterByID("imp")
	if imp == nil {
		return false
	}
	for i := range gs.players {
		if i == deadImpIndex || !gs.players[i].IsAlive || !gs.playerHasCharacterTypeLocked(i, game.CharacterTypeMinion) {
			continue
		}
		gs.players[i].Character = &game.Character{
			ID:      imp.ID,
			Name:    imp.Name,
			Team:    imp.Team,
			Ability: imp.Ability,
		}
		return true
	}
	return false
}

func (gs *GameSession) applyScarletWomanStarpassLocked(demonDeathAliveCount int) bool {
	if demonDeathAliveCount < 5 || !gs.hasDeadDemonLocked() {
		return false
	}

	imp := game.GetCharacterByID("imp")
	if imp == nil {
		return false
	}

	for i := range gs.players {
		player := &gs.players[i]
		if !player.IsAlive ||
			player.Character == nil ||
			player.Character.ID != "scarletwoman" ||
			gs.playerAbilityMalfunctioningLocked(i) {
			continue
		}
		player.Character = &game.Character{
			ID:      imp.ID,
			Name:    imp.Name,
			Team:    imp.Team,
			Ability: imp.Ability,
		}
		return true
	}
	return false
}

func (gs *GameSession) hasDeadDemonLocked() bool {
	for _, player := range gs.players {
		if player.IsAlive || player.Character == nil {
			continue
		}
		charDef := game.GetCharacterByID(player.Character.ID)
		if charDef != nil && charDef.Type == game.CharacterTypeDemon {
			return true
		}
	}
	return false
}
