package gameplay

import (
	"fmt"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

// ────────────────────────────────────────────────
// StartGameCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyStartGame(cmd StartGameCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("game can only be started from setup phase, current phase: %d", gs.phase)
	}
	if gs.storytellerID == "" {
		return ApplyResult{}, fmt.Errorf("storyteller must be set before starting the game")
	}
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can start the game")
	}

	// Verify all players have characters assigned
	assignments := make(map[string]string, len(gs.players))
	for _, p := range gs.players {
		if p.Character == nil {
			return ApplyResult{}, fmt.Errorf("player %s has no character assigned", p.ID)
		}
		assignments[p.ID] = p.Character.ID
	}
	if !game.ValidateScriptAssignment(gs.scriptID, assignments, len(gs.players)) {
		return ApplyResult{}, fmt.Errorf("invalid character assignment for player count")
	}
	shownCharacters := make(map[string]string, len(gs.players))
	for _, p := range gs.players {
		if p.ShownCharacter != nil {
			shownCharacters[p.ID] = p.ShownCharacter.ID
		}
	}
	if err := validateShownCharacters(gs.scriptID, assignments, shownCharacters); err != nil {
		return ApplyResult{}, err
	}
	if err := validateFortuneTellerRedHerring(gs.scriptID, assignments, gs.fortuneTellerRedHerringID); err != nil {
		return ApplyResult{}, err
	}

	gs.phase = game.GamePhaseNight
	gs.dayNumber = 0
	gs.startNightLocked()

	events := []game.GameEvent{
		{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseNight}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// ChangePhaseCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyChangePhase(cmd ChangePhaseCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can change phase")
	}
	if gs.phase == game.GamePhaseSetup || gs.phase == game.GamePhaseFinished {
		return ApplyResult{}, fmt.Errorf("cannot change phase from %d", gs.phase)
	}

	if gs.phase == game.GamePhaseNight && cmd.Phase == game.GamePhaseDay {
		return ApplyResult{}, fmt.Errorf("night phase must be resolved with RESOLVE_NIGHT")
	}
	if !(gs.phase == game.GamePhaseDay && cmd.Phase == game.GamePhaseNight) {
		return ApplyResult{}, fmt.Errorf("invalid phase transition from %d to %d", gs.phase, cmd.Phase)
	}

	if cmd.Phase == game.GamePhaseNight {
		if won := gs.mayorEndgameWinnerLocked(); won != nil {
			gs.winner = won
			gs.phase = game.GamePhaseFinished
			return ApplyResult{Events: []game.GameEvent{{GameEnded: won}}, Updated: true}, nil
		}
		// Clear expired poison at dusk (Day→Night transition)
		gs.clearExpiredPoisonLocked()
		gs.startNightLocked()
	}

	gs.phase = cmd.Phase

	events := []game.GameEvent{
		{PhaseChanged: &game.PhaseChanged{Phase: cmd.Phase}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// NominateCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyNominate(cmd NominateCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseDay {
		return ApplyResult{}, fmt.Errorf("nominations can only happen during the day phase")
	}
	if cmd.SenderID == cmd.NomineeID {
		return ApplyResult{}, fmt.Errorf("cannot nominate yourself")
	}

	nomIdx := gs.findPlayerIndex(cmd.SenderID)
	if nomIdx == -1 {
		return ApplyResult{}, fmt.Errorf("nominator %s not found", cmd.SenderID)
	}
	if !gs.players[nomIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("dead players cannot nominate")
	}

	nomineeIdx := gs.findPlayerIndex(cmd.NomineeID)
	if nomineeIdx == -1 {
		return ApplyResult{}, fmt.Errorf("nominee %s not found", cmd.NomineeID)
	}
	if !gs.players[nomineeIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("cannot nominate a dead player")
	}
	if gs.nominatorsToday == nil {
		gs.nominatorsToday = make(map[string]bool)
	}
	if gs.nomineesToday == nil {
		gs.nomineesToday = make(map[string]bool)
	}
	if gs.nominatorsToday[cmd.SenderID] {
		return ApplyResult{}, fmt.Errorf("player %s has already nominated today", cmd.SenderID)
	}
	if gs.nomineesToday[cmd.NomineeID] {
		return ApplyResult{}, fmt.Errorf("player %s has already been nominated today", cmd.NomineeID)
	}

	gs.nominatorsToday[cmd.SenderID] = true
	gs.nomineesToday[cmd.NomineeID] = true

	if gs.virginAbilityUsed == nil {
		gs.virginAbilityUsed = make(map[string]bool)
	}
	if gs.players[nomineeIdx].Character != nil &&
		gs.players[nomineeIdx].Character.ID == "virgin" &&
		!gs.virginAbilityUsed[cmd.NomineeID] {
		gs.virginAbilityUsed[cmd.NomineeID] = true
		if !gs.playerAbilityMalfunctioningLocked(nomineeIdx) &&
			gs.playerHasCharacterTypeLocked(nomIdx, game.CharacterTypeTownsfolk) {
			return gs.applyVirginExecutionLocked(cmd, nomIdx)
		}
	}

	gs.phase = game.GamePhaseVoting
	gs.nomination = &game.Nomination{
		NominatorID: cmd.SenderID,
		NomineeID:   cmd.NomineeID,
		Votes:       make(map[string]bool),
	}

	events := []game.GameEvent{
		{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseVoting}},
		{NominationStarted: &game.NominationStartedEvent{
			NominatorID: cmd.SenderID,
			NomineeID:   cmd.NomineeID,
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func (gs *GameSession) playerHasCharacterTypeLocked(playerIndex int, characterType game.CharacterType) bool {
	if playerIndex < 0 || playerIndex >= len(gs.players) || gs.players[playerIndex].Character == nil {
		return false
	}
	charDef := game.GetCharacterByID(gs.players[playerIndex].Character.ID)
	return charDef != nil && charDef.Type == characterType
}

func (gs *GameSession) applyVirginExecutionLocked(cmd NominateCmd, nominatorIndex int) (ApplyResult, error) {
	gs.players[nominatorIndex].IsAlive = false
	gs.deaths = append(gs.deaths, game.DeathRecord{
		PlayerID:  cmd.SenderID,
		Cause:     game.DeathCauseExecution,
		DayNumber: gs.dayNumber,
	})

	events := []game.GameEvent{
		{NominationStarted: &game.NominationStartedEvent{
			NominatorID: cmd.SenderID,
			NomineeID:   cmd.NomineeID,
		}},
		{PlayerDied: &game.PlayerDiedEvent{
			PlayerID:  cmd.SenderID,
			Cause:     game.DeathCauseExecution,
			DayNumber: gs.dayNumber,
		}},
	}

	if won := gs.checkWinConditions(0); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
		return ApplyResult{Events: events, Updated: true}, nil
	}

	gs.phase = game.GamePhaseNight
	gs.nomination = nil
	gs.startNightLocked()
	events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseNight}})

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// CastVoteCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyCastVote(cmd CastVoteCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseVoting || gs.nomination == nil {
		return ApplyResult{}, fmt.Errorf("no active nomination to vote on")
	}

	voterIdx := gs.findPlayerIndex(cmd.SenderID)
	if voterIdx == -1 {
		return ApplyResult{}, fmt.Errorf("voter %s not found", cmd.SenderID)
	}
	voter := gs.players[voterIdx]

	// Dead players may cast one ghost vote
	if !voter.IsAlive {
		if gs.ghostVotesUsed[cmd.SenderID] {
			return ApplyResult{}, fmt.Errorf("ghost vote already used")
		}
		gs.ghostVotesUsed[cmd.SenderID] = true
	}

	// Check for duplicate vote
	if _, already := gs.nomination.Votes[cmd.SenderID]; already {
		return ApplyResult{}, fmt.Errorf("player %s has already voted", cmd.SenderID)
	}
	if err := gs.validateButlerVoteLocked(voterIdx, cmd.Decision); err != nil {
		return ApplyResult{}, err
	}

	gs.nomination.Votes[cmd.SenderID] = cmd.Decision

	events := []game.GameEvent{
		{VoteCast: &game.VoteCast{
			VoterID:  cmd.SenderID,
			Decision: &cmd.Decision,
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func (gs *GameSession) validateButlerVoteLocked(voterIdx int, decision bool) error {
	if !decision || voterIdx < 0 || voterIdx >= len(gs.players) {
		return nil
	}
	voter := gs.players[voterIdx]
	if !voter.IsAlive || voter.Character == nil || voter.Character.ID != "butler" {
		return nil
	}
	if gs.playerAbilityMalfunctioningLocked(voterIdx) {
		return nil
	}
	masterID := gs.butlerMasters[voter.ID]
	if masterID == "" {
		return fmt.Errorf("butler %s has not chosen a master", voter.ID)
	}
	if !gs.nomination.Votes[masterID] {
		return fmt.Errorf("butler %s cannot vote until master %s votes yes", voter.ID, masterID)
	}
	return nil
}

// ────────────────────────────────────────────────
// ResolveNominationCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyResolveNomination(cmd ResolveNominationCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can resolve a nomination")
	}
	if gs.phase != game.GamePhaseVoting || gs.nomination == nil {
		return ApplyResult{}, fmt.Errorf("no active nomination to resolve")
	}

	yesVotes := 0
	noVotes := 0
	for _, decision := range gs.nomination.Votes {
		if decision {
			yesVotes++
		} else {
			noVotes++
		}
	}

	requiredVotes := majorityThreshold(gs.alivePlayerCountLocked())
	executed := yesVotes >= requiredVotes
	nomineeID := gs.nomination.NomineeID

	gs.nomination.Resolved = true
	gs.nomination = nil
	gs.phase = game.GamePhaseDay

	events := []game.GameEvent{
		{NominationResolved: &game.NominationResolvedEvent{
			NomineeID:     nomineeID,
			Executed:      executed,
			YesVotes:      yesVotes,
			NoVotes:       noVotes,
			RequiredVotes: requiredVotes,
		}},
	}

	// If executed, kill the player and check win conditions
	if executed {
		pIdx := gs.findPlayerIndex(nomineeID)
		if pIdx != -1 {
			aliveBeforeDeath := gs.alivePlayerCountLocked()
			demonDeathAliveCount := gs.demonDeathAliveCountLocked(pIdx, aliveBeforeDeath)
			gs.players[pIdx].IsAlive = false
			gs.deaths = append(gs.deaths, game.DeathRecord{
				PlayerID:  nomineeID,
				Cause:     game.DeathCauseExecution,
				DayNumber: gs.dayNumber,
			})
			events = append(events, game.GameEvent{
				PlayerDied: &game.PlayerDiedEvent{
					PlayerID:  nomineeID,
					Cause:     game.DeathCauseExecution,
					DayNumber: gs.dayNumber,
				},
			})
			if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
				gs.winner = won
				events = append(events, game.GameEvent{GameEnded: won})
				gs.phase = game.GamePhaseFinished
			}
		}
	}
	if gs.phase != game.GamePhaseFinished {
		if executed {
			gs.phase = game.GamePhaseNight
			gs.startNightLocked()
			events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseNight}})
		} else {
			gs.phase = game.GamePhaseDay
			events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay}})
		}
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// ExecutePlayerCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyExecutePlayer(cmd ExecutePlayerCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can execute a player")
	}
	if gs.phase != game.GamePhaseDay {
		return ApplyResult{}, fmt.Errorf("players can only be executed during the day phase")
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
		Cause:     game.DeathCauseExecution,
		DayNumber: gs.dayNumber,
	})

	events := []game.GameEvent{
		{PlayerDied: &game.PlayerDiedEvent{
			PlayerID:  cmd.PlayerID,
			Cause:     game.DeathCauseExecution,
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

// ────────────────────────────────────────────────
// UseSlayerAbilityCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyUseSlayerAbility(cmd UseSlayerAbilityCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseDay {
		return ApplyResult{}, fmt.Errorf("Slayer ability can only be used during the day phase")
	}
	if cmd.TargetPlayerID == "" {
		return ApplyResult{}, fmt.Errorf("target player is required")
	}

	slayerIdx := gs.findPlayerIndex(cmd.SenderID)
	if slayerIdx == -1 {
		return ApplyResult{}, fmt.Errorf("Slayer %s not found", cmd.SenderID)
	}
	slayer := gs.players[slayerIdx]
	if !slayer.IsAlive {
		return ApplyResult{}, fmt.Errorf("dead players cannot use the Slayer ability")
	}
	if !gs.playerCanUseVisibleCharacterLocked(slayerIdx, "slayer") {
		return ApplyResult{}, fmt.Errorf("only the Slayer can use this ability")
	}
	if gs.slayerUsed == nil {
		gs.slayerUsed = make(map[string]bool)
	}
	if gs.slayerUsed[cmd.SenderID] {
		return ApplyResult{}, fmt.Errorf("Slayer ability already used")
	}

	targetIdx := gs.findPlayerIndex(cmd.TargetPlayerID)
	if targetIdx == -1 {
		return ApplyResult{}, fmt.Errorf("target player %s not found", cmd.TargetPlayerID)
	}
	if !gs.players[targetIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("cannot target a dead player")
	}

	gs.slayerUsed[cmd.SenderID] = true
	events := []game.GameEvent{}
	var targetDef *game.CharacterDefinition
	if gs.players[targetIdx].Character != nil {
		targetDef = game.GetCharacterByID(gs.players[targetIdx].Character.ID)
	}
	if targetDef != nil &&
		targetDef.Type == game.CharacterTypeDemon &&
		slayer.Character != nil &&
		slayer.Character.ID == "slayer" &&
		!gs.playerAbilityMalfunctioningLocked(slayerIdx) {
		aliveBeforeDeath := gs.alivePlayerCountLocked()
		demonDeathAliveCount := gs.demonDeathAliveCountLocked(targetIdx, aliveBeforeDeath)
		gs.players[targetIdx].IsAlive = false
		gs.deaths = append(gs.deaths, game.DeathRecord{
			PlayerID:  cmd.TargetPlayerID,
			Cause:     game.DeathCauseAbility,
			DayNumber: gs.dayNumber,
			KilledBy:  cmd.SenderID,
		})
		events = append(events, game.GameEvent{
			PlayerDied: &game.PlayerDiedEvent{
				PlayerID:  cmd.TargetPlayerID,
				Cause:     game.DeathCauseAbility,
				DayNumber: gs.dayNumber,
			},
		})
		if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
			gs.winner = won
			events = append(events, game.GameEvent{GameEnded: won})
			gs.phase = game.GamePhaseFinished
		}
	}

	return ApplyResult{Events: events, Updated: true}, nil
}
