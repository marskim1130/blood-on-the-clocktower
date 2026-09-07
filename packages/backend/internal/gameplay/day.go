package gameplay

import (
	"fmt"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

const (
	defaultAccusationDuration = 30 * time.Second
	defaultDefenseDuration    = 30 * time.Second
	defaultVoterDuration      = 3 * time.Second
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
		if !p.HasConfirmedCharacter {
			return ApplyResult{}, fmt.Errorf("all seated players must confirm their character before the game starts")
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
	if gs.phase == game.GamePhaseDay && cmd.Phase == game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("day phase must be finalized with FINALIZE_DAY")
	}
	return ApplyResult{}, fmt.Errorf("invalid phase transition from %d to %d", gs.phase, cmd.Phase)
}

func (gs *GameSession) applyFinalizeDay(cmd FinalizeDayCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can finalize the day")
	}
	if gs.phase != game.GamePhaseDay || gs.nomination != nil {
		return ApplyResult{}, fmt.Errorf("day can only be finalized without an active nomination")
	}

	events := make([]game.GameEvent, 0, 2)
	candidateID, _, _ := gs.executionBlockLocked()
	if candidateID != "" {
		playerIndex := gs.findPlayerIndex(candidateID)
		if playerIndex == -1 || !gs.players[playerIndex].IsAlive {
			return ApplyResult{}, fmt.Errorf("execution candidate %s is not alive", candidateID)
		}
		aliveBeforeDeath := gs.alivePlayerCountLocked()
		demonDeathAliveCount := gs.demonDeathAliveCountLocked(playerIndex, aliveBeforeDeath)
		gs.players[playerIndex].IsAlive = false
		gs.deaths = append(gs.deaths, game.DeathRecord{
			PlayerID:  candidateID,
			Cause:     game.DeathCauseExecution,
			DayNumber: gs.dayNumber,
		})
		events = append(events, game.GameEvent{PlayerDied: &game.PlayerDiedEvent{
			PlayerID:  candidateID,
			Cause:     game.DeathCauseExecution,
			DayNumber: gs.dayNumber,
		}})
		if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
			gs.winner = won
			gs.phase = game.GamePhaseFinished
			events = append(events, game.GameEvent{GameEnded: won})
			return ApplyResult{Events: events, Updated: true}, nil
		}
	} else if won := gs.mayorEndgameWinnerLocked(); won != nil {
		gs.winner = won
		gs.phase = game.GamePhaseFinished
		events = append(events, game.GameEvent{GameEnded: won})
		return ApplyResult{Events: events, Updated: true}, nil
	}

	gs.clearExpiredPoisonLocked()
	gs.startNightLocked()
	gs.phase = game.GamePhaseNight
	events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseNight}})
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
		NominatorID:       cmd.SenderID,
		NomineeID:         cmd.NomineeID,
		Votes:             make(map[string]bool),
		VoterOrder:        gs.clockwiseVoterOrderLocked(nomineeIdx),
		CurrentVoterIndex: 0,
		Stage:             game.NominationStageAccusation,
		DeadlineUnixMs:    time.Now().Add(defaultAccusationDuration).UnixMilli(),
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

func (gs *GameSession) clockwiseVoterOrderLocked(startIndex int) []string {
	order := make([]string, 0, len(gs.players))
	for offset := 0; offset < len(gs.players); offset++ {
		index := (startIndex + offset + 1) % len(gs.players)
		order = append(order, gs.players[index].ID)
	}
	return order
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
	if gs.nomination.Stage != "" && gs.nomination.Stage != game.NominationStageVoting {
		return ApplyResult{}, fmt.Errorf("voting is not open during the %s stage", gs.nomination.Stage)
	}
	if gs.nomination.Paused {
		return ApplyResult{}, fmt.Errorf("nomination timer is paused")
	}
	if gs.nomination.CurrentVoterIndex < 0 || gs.nomination.CurrentVoterIndex >= len(gs.nomination.VoterOrder) {
		return ApplyResult{}, fmt.Errorf("the clockwise ballot is complete")
	}
	currentVoterID := gs.nomination.VoterOrder[gs.nomination.CurrentVoterIndex]
	if cmd.SenderID != currentVoterID {
		return ApplyResult{}, fmt.Errorf("waiting for player %s to vote", currentVoterID)
	}

	voterIdx := gs.findPlayerIndex(cmd.SenderID)
	if voterIdx == -1 {
		return ApplyResult{}, fmt.Errorf("voter %s not found", cmd.SenderID)
	}
	voter := gs.players[voterIdx]

	// Check for duplicate vote
	if _, already := gs.nomination.Votes[cmd.SenderID]; already {
		return ApplyResult{}, fmt.Errorf("player %s has already voted", cmd.SenderID)
	}
	// A dead player's one-use ghost vote is spent only after a yes vote passes validation.
	if !voter.IsAlive && cmd.Decision {
		if gs.ghostVotesUsed[cmd.SenderID] {
			return ApplyResult{}, fmt.Errorf("ghost vote already used")
		}
		gs.ghostVotesUsed[cmd.SenderID] = true
	}

	gs.nomination.Votes[cmd.SenderID] = cmd.Decision
	gs.nomination.CurrentVoterIndex++
	if gs.nomination.CurrentVoterIndex < len(gs.nomination.VoterOrder) {
		gs.nomination.DeadlineUnixMs = time.Now().Add(defaultVoterDuration).UnixMilli()
	} else {
		gs.nomination.DeadlineUnixMs = 0
	}

	events := []game.GameEvent{
		{VoteCast: &game.VoteCast{
			VoterID:  cmd.SenderID,
			Decision: &cmd.Decision,
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func (gs *GameSession) applyAdvanceNominationStage(cmd AdvanceNominationStageCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can advance the nomination stage")
	}
	if gs.phase != game.GamePhaseVoting || gs.nomination == nil {
		return ApplyResult{}, fmt.Errorf("no active nomination can be advanced")
	}
	if gs.nomination.Paused {
		return ApplyResult{}, fmt.Errorf("resume the nomination timer before advancing")
	}
	switch gs.nomination.Stage {
	case game.NominationStageAccusation:
		gs.nomination.Stage = game.NominationStageDefense
		gs.nomination.DeadlineUnixMs = time.Now().Add(defaultDefenseDuration).UnixMilli()
	case game.NominationStageDefense:
		gs.nomination.Stage = game.NominationStageVoting
		gs.nomination.DeadlineUnixMs = time.Now().Add(defaultVoterDuration).UnixMilli()
	default:
		return ApplyResult{}, fmt.Errorf("nomination stage %s cannot be advanced", gs.nomination.Stage)
	}
	gs.nomination.RemainingMs = 0
	return ApplyResult{Updated: true}, nil
}

func (gs *GameSession) applyRecordVote(cmd RecordVoteCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can record a player's vote")
	}
	return gs.applyCastVote(CastVoteCmd{SenderID: cmd.VoterID, Decision: cmd.Decision})
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
	if len(gs.nomination.VoterOrder) == 0 || gs.nomination.CurrentVoterIndex < len(gs.nomination.VoterOrder) {
		return ApplyResult{}, fmt.Errorf("the clockwise ballot is not complete")
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
	nomineeID := gs.nomination.NomineeID
	gs.nominationResults = append(gs.nominationResults, game.NominationResult{
		DayNumber:     gs.dayNumber,
		NominatorID:   gs.nomination.NominatorID,
		NomineeID:     nomineeID,
		Votes:         cloneBoolMap(gs.nomination.Votes),
		YesVotes:      yesVotes,
		NoVotes:       noVotes,
		RequiredVotes: requiredVotes,
	})

	gs.nomination.Resolved = true
	gs.nomination = nil
	gs.phase = game.GamePhaseDay

	events := []game.GameEvent{
		{NominationResolved: &game.NominationResolvedEvent{
			NomineeID:     nomineeID,
			Executed:      false,
			YesVotes:      yesVotes,
			NoVotes:       noVotes,
			RequiredVotes: requiredVotes,
		}},
		{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay}},
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
