package gameplay

import (
	"fmt"
	"strings"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

// ────────────────────────────────────────────────
// SubmitNightActionCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applySubmitNightAction(cmd SubmitNightActionCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("night actions can only be submitted during the night phase")
	}
	if gs.confirmedNightAction != nil {
		return ApplyResult{}, fmt.Errorf("the current night result is waiting for player acknowledgement")
	}

	actionType := game.NightActionType(cmd.ActionType)
	var step *game.NightWakeStep
	if cmd.SenderID != gs.storytellerID {
		if gs.pendingNightAction != nil {
			return ApplyResult{}, fmt.Errorf("the current night choice is already waiting for storyteller review")
		}
		actorIdx := gs.findPlayerIndex(cmd.SenderID)
		if actorIdx == -1 {
			return ApplyResult{}, fmt.Errorf("actor %s not found", cmd.SenderID)
		}
		if !gs.players[actorIdx].IsAlive {
			return ApplyResult{}, fmt.Errorf("dead players cannot submit night actions")
		}
		step = gs.currentNightWakeStepLocked()
		if step == nil {
			return ApplyResult{}, fmt.Errorf("no remaining night wake steps")
		}
		if actionType != step.ActionType {
			return ApplyResult{}, fmt.Errorf("expected night action %s, got %s", step.ActionType, actionType)
		}
		if !gs.playerMatchesNightWakeStepLocked(actorIdx, *step) {
			return ApplyResult{}, fmt.Errorf("player %s cannot act during night action %s", cmd.SenderID, step.ActionType)
		}
		if storytellerChoosesNightTargets(*step) {
			if len(cmd.TargetIDs) != 0 {
				return ApplyResult{}, fmt.Errorf("the storyteller chooses information targets")
			}
		} else if err := gs.validateNightTargetsLocked(*step, cmd.TargetIDs); err != nil {
			return ApplyResult{}, err
		}
	} else {
		if gs.pendingNightAction != nil {
			return ApplyResult{}, fmt.Errorf("the current player choice is waiting for storyteller confirmation")
		}
		step = gs.currentNightWakeStepLocked()
		if step == nil {
			return ApplyResult{}, fmt.Errorf("no remaining night wake steps")
		}
		if actionType != step.ActionType {
			return ApplyResult{}, fmt.Errorf("expected night action %s, got %s", step.ActionType, actionType)
		}
		if err := gs.validateNightTargetsLocked(*step, cmd.TargetIDs); err != nil {
			return ApplyResult{}, err
		}
	}

	action := game.NightAction{
		ActorID:    cmd.SenderID,
		ActionType: actionType,
		TargetIDs:  append([]string(nil), cmd.TargetIDs...),
		Result:     strings.TrimSpace(cmd.Result),
	}
	if cmd.SenderID != gs.storytellerID {
		gs.pendingNightAction = cloneNightAction(&action)
		return ApplyResult{
			Events: []game.GameEvent{{NightActionSubmitted: &game.NightActionEvent{
				ActorID:    action.ActorID,
				ActionType: action.ActionType,
				TargetIDs:  append([]string(nil), action.TargetIDs...),
				Result:     optionalString(action.Result),
			}}},
			Updated: true,
		}, nil
	}

	gs.populateNightActionResultLocked(step, &action)

	gs.nightActions = append(gs.nightActions, action)
	if cmd.SenderID == gs.storytellerID {
		gs.nightWakeIndex++
		gs.applyConfirmedNightEffectLocked(step, action)
	}

	events := []game.GameEvent{
		{NightActionSubmitted: &game.NightActionEvent{
			ActorID:    cmd.SenderID,
			ActionType: game.NightActionType(cmd.ActionType),
			TargetIDs:  cmd.TargetIDs,
			Result:     optionalString(action.Result),
		}},
	}

	if gs.dawnDeathsLocked() && gs.currentNightWakeStepLocked() == nil {
		dawn, err := gs.resolveNightWithDeathsLocked(gs.pendingDawnDeathIDs)
		if err != nil {
			return ApplyResult{}, err
		}
		dawn.Events = append(events, dawn.Events...)
		return dawn, nil
	}
	return ApplyResult{Events: events, Updated: true}, nil
}

func (gs *GameSession) applyConfirmNightAction(cmd ConfirmNightActionCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can confirm a night action")
	}
	if gs.phase != game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("night actions can only be confirmed during the night phase")
	}
	if gs.confirmedNightAction != nil {
		return ApplyResult{}, fmt.Errorf("the current night result is waiting for player acknowledgement")
	}
	if gs.pendingNightAction == nil {
		return ApplyResult{}, fmt.Errorf("no player night choice is waiting for review")
	}
	step := gs.currentNightWakeStepLocked()
	if step == nil || gs.pendingNightAction.ActionType != step.ActionType {
		return ApplyResult{}, fmt.Errorf("pending night choice does not match the current wake step")
	}

	targetIDs := gs.pendingNightAction.TargetIDs
	if cmd.TargetIDs != nil {
		targetIDs = cmd.TargetIDs
	}
	if err := gs.validateNightTargetsLocked(*step, targetIDs); err != nil {
		return ApplyResult{}, err
	}
	action := *gs.pendingNightAction
	action.TargetIDs = append([]string(nil), targetIDs...)
	action.Result = strings.TrimSpace(cmd.Result)
	gs.populateNightActionResultLocked(step, &action)
	if nightActionRequiresInformation(action.ActionType) && action.Result == "" {
		return ApplyResult{}, fmt.Errorf("the storyteller must provide a private information result before confirming")
	}
	gs.nightActions = append(gs.nightActions, action)
	gs.applyConfirmedNightEffectLocked(step, action)
	gs.pendingNightAction = nil
	gs.confirmedNightAction = cloneNightAction(&action)
	gs.nightAcknowledged = make(map[string]bool)

	return ApplyResult{
		Events: []game.GameEvent{{NightActionSubmitted: &game.NightActionEvent{
			ActorID:    action.ActorID,
			ActionType: action.ActionType,
			TargetIDs:  append([]string(nil), action.TargetIDs...),
			Result:     optionalString(action.Result),
		}}},
		Updated: true,
	}, nil
}

func (gs *GameSession) applyAcknowledgeNightAction(cmd AcknowledgeNightActionCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseNight || gs.confirmedNightAction == nil {
		return ApplyResult{}, fmt.Errorf("no confirmed night result is waiting for acknowledgement")
	}
	step := gs.currentNightWakeStepLocked()
	actorIdx := gs.findPlayerIndex(cmd.SenderID)
	if step == nil || actorIdx == -1 || !gs.playerMatchesNightWakeStepLocked(actorIdx, *step) {
		return ApplyResult{}, fmt.Errorf("player %s cannot acknowledge the current night result", cmd.SenderID)
	}
	if gs.nightAcknowledged == nil {
		gs.nightAcknowledged = make(map[string]bool)
	}
	if gs.nightAcknowledged[cmd.SenderID] {
		return ApplyResult{}, fmt.Errorf("player %s already acknowledged the current night result", cmd.SenderID)
	}
	gs.nightAcknowledged[cmd.SenderID] = true

	allAcknowledged := true
	for index := range gs.players {
		if !gs.players[index].IsAlive || !gs.playerMatchesNightWakeStepLocked(index, *step) {
			continue
		}
		if !gs.nightAcknowledged[gs.players[index].ID] {
			allAcknowledged = false
			break
		}
	}
	if allAcknowledged {
		gs.nightWakeIndex++
		gs.confirmedNightAction = nil
		gs.nightAcknowledged = nil
		if gs.dawnDeathsLocked() && gs.currentNightWakeStepLocked() == nil {
			return gs.resolveNightWithDeathsLocked(gs.pendingDawnDeathIDs)
		}
	}
	return ApplyResult{Updated: true}, nil
}

func (gs *GameSession) applySkipNightAction(cmd SkipNightActionCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can skip a night step")
	}
	if gs.phase != game.GamePhaseNight || gs.currentNightWakeStepLocked() == nil {
		return ApplyResult{}, fmt.Errorf("no current night step can be skipped")
	}
	gs.nightWakeIndex++
	gs.pendingNightAction = nil
	gs.confirmedNightAction = nil
	gs.nightAcknowledged = nil
	if gs.dawnDeathsLocked() && gs.currentNightWakeStepLocked() == nil {
		return gs.resolveNightWithDeathsLocked(gs.pendingDawnDeathIDs)
	}
	return ApplyResult{Updated: true}, nil
}

func (gs *GameSession) populateNightActionResultLocked(step *game.NightWakeStep, action *game.NightAction) {
	if step == nil || action == nil || action.Result != "" {
		return
	}
	actorCharID := step.CharacterID
	switch action.ActionType {
	case game.NightActionLearnDemon:
		if step.CharacterType == game.NightWakeCharacterTypeMinion {
			action.Result = gs.computeDemonInfoResultLocked()
		}
	case game.NightActionLearnTownsfolk:
		action.Result = gs.computeWasherwomanResultLocked(actorCharID, action.TargetIDs)
	case game.NightActionLearnOutsider:
		action.Result = gs.computeLibrarianResultLocked(actorCharID, action.TargetIDs)
	case game.NightActionLearnMinion:
		if step.CharacterType == game.NightWakeCharacterTypeDemon {
			action.Result = gs.computeMinionInfoResultLocked()
		} else {
			action.Result = gs.computeInvestigatorResultLocked(actorCharID, action.TargetIDs)
		}
	case game.NightActionLearnEvilPairs:
		action.Result = gs.computeChefResultLocked(actorCharID)
	case game.NightActionLearnEvilNeighbors:
		action.Result = gs.computeEmpathResultLocked(actorCharID)
	case game.NightActionCheckDemon:
		action.Result = gs.computeFortuneTellerResultLocked(actorCharID, action.TargetIDs)
	case game.NightActionLearnExecuted:
		action.Result = gs.computeUndertakerResultLocked(actorCharID)
	case game.NightActionLearnDied:
		action.Result = gs.computeRavenkeeperResultLocked(actorCharID, action.TargetIDs)
	case game.NightActionShowGrimoire:
		action.Result = gs.computeSpyGrimoireLocked()
	}
}

func (gs *GameSession) applyConfirmedNightEffectLocked(step *game.NightWakeStep, action game.NightAction) {
	if step == nil {
		return
	}
	if action.ActionType == game.NightActionPoison {
		gs.applyPoisonEffectLocked(action.TargetIDs)
	}
	if action.ActionType == game.NightActionLearnMaster {
		gs.applyButlerMasterSelectionLocked(step.CharacterID, action.TargetIDs)
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nightActionRequiresInformation(actionType game.NightActionType) bool {
	switch actionType {
	case game.NightActionLearnDemon, game.NightActionLearnTownsfolk, game.NightActionLearnOutsider,
		game.NightActionLearnMinion, game.NightActionLearnEvilPairs, game.NightActionLearnEvilNeighbors,
		game.NightActionCheckDemon, game.NightActionLearnExecuted, game.NightActionLearnDied, game.NightActionShowGrimoire:
		return true
	default:
		return false
	}
}

// ────────────────────────────────────────────────
// ResolveNightCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyPrepareDawn(cmd PrepareDawnCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can prepare dawn")
	}
	if gs.phase != game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("dawn can only be prepared during the night phase")
	}
	if gs.dawnReviewPending {
		return ApplyResult{}, fmt.Errorf("dawn is already waiting for storyteller confirmation")
	}
	if gs.dawnDeathsLocked() {
		return ApplyResult{}, fmt.Errorf("confirmed dawn is waiting for the Ravenkeeper")
	}
	if remaining := len(gs.activeNightWakeStepsLocked()) - gs.nightWakeIndex; remaining > 0 {
		return ApplyResult{}, fmt.Errorf("cannot prepare dawn with %d wake step(s) remaining", remaining)
	}
	if gs.pendingNightAction != nil || gs.confirmedNightAction != nil {
		return ApplyResult{}, fmt.Errorf("cannot prepare dawn while a night step is unresolved")
	}
	gs.pendingDawnDeathIDs = gs.suggestedDawnDeathsLocked()
	gs.dawnReviewPending = true
	return ApplyResult{Updated: true}, nil
}

func (gs *GameSession) applyConfirmDawn(cmd ConfirmDawnCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can confirm dawn")
	}
	if gs.phase != game.GamePhaseNight || !gs.dawnReviewPending {
		return ApplyResult{}, fmt.Errorf("no dawn proposal is waiting for confirmation")
	}
	seen := make(map[string]bool, len(cmd.DeathPlayerIDs))
	for _, playerID := range cmd.DeathPlayerIDs {
		playerIndex := gs.findPlayerIndex(playerID)
		if playerIndex == -1 || !gs.players[playerIndex].IsAlive {
			return ApplyResult{}, fmt.Errorf("dawn death player %s must be alive and seated", playerID)
		}
		if seen[playerID] {
			return ApplyResult{}, fmt.Errorf("duplicate dawn death player %s", playerID)
		}
		seen[playerID] = true
	}
	gs.pendingDawnDeathIDs = append([]string(nil), cmd.DeathPlayerIDs...)
	gs.dawnReviewPending = false
	if gs.currentNightWakeStepLocked() != nil {
		return ApplyResult{Updated: true}, nil
	}
	return gs.resolveNightWithDeathsLocked(cmd.DeathPlayerIDs)
}

func (gs *GameSession) applyResolveNight(cmd ResolveNightCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can resolve the night")
	}
	if gs.phase != game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("night can only be resolved during the night phase")
	}
	if remaining := len(gs.activeNightWakeStepsLocked()) - gs.nightWakeIndex; remaining > 0 {
		return ApplyResult{}, fmt.Errorf("cannot resolve night with %d wake step(s) remaining", remaining)
	}
	if gs.dawnReviewPending {
		return ApplyResult{}, fmt.Errorf("dawn must be confirmed with CONFIRM_DAWN")
	}
	return gs.resolveNightWithDeathsLocked(gs.suggestedDawnDeathsLocked())
}

func (gs *GameSession) suggestedDawnDeathsLocked() []string {
	if !gs.nightDemonCanKillLocked() {
		return []string{}
	}
	protectedTargets := gs.nightProtectedTargetsLocked()
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, action := range gs.nightActions {
		if action.ActionType != game.NightActionKill {
			continue
		}
		for _, targetID := range action.TargetIDs {
			targetIndex := gs.findPlayerIndex(targetID)
			if targetIndex == -1 || !gs.players[targetIndex].IsAlive || seen[targetID] || gs.nightKillPreventedLocked(targetIndex, protectedTargets) {
				continue
			}
			seen[targetID] = true
			result = append(result, targetID)
		}
	}
	return result
}

func (gs *GameSession) resolveNightWithDeathsLocked(deathPlayerIDs []string) (ApplyResult, error) {
	var events []game.GameEvent
	demonDeathAliveCount := 0
	upcomingDayNumber := gs.dayNumber + 1
	killedBy := gs.storytellerID
	for _, action := range gs.nightActions {
		if action.ActionType == game.NightActionKill && action.ActorID != "" {
			killedBy = action.ActorID
			break
		}
	}
	for _, targetID := range deathPlayerIDs {
		targetIndex := gs.findPlayerIndex(targetID)
		if targetIndex == -1 || !gs.players[targetIndex].IsAlive {
			continue
		}
		aliveBeforeDeath := gs.alivePlayerCountLocked()
		impSelfKill := gs.impChoseSelfKillLocked(targetIndex)
		if aliveAtDemonDeath := gs.demonDeathAliveCountLocked(targetIndex, aliveBeforeDeath); aliveAtDemonDeath > demonDeathAliveCount {
			demonDeathAliveCount = aliveAtDemonDeath
		}
		gs.players[targetIndex].IsAlive = false
		gs.deaths = append(gs.deaths, game.DeathRecord{PlayerID: targetID, Cause: game.DeathCauseNightKill, DayNumber: upcomingDayNumber, KilledBy: killedBy})
		events = append(events, game.GameEvent{PlayerDied: &game.PlayerDiedEvent{PlayerID: targetID, Cause: game.DeathCauseNightKill, DayNumber: upcomingDayNumber}})
		if impSelfKill && gs.applyImpSelfStarpassLocked(targetIndex) {
			demonDeathAliveCount = 0
		}
	}

	// Clear night actions for next night
	gs.nightActions = nil
	gs.pendingNightAction = nil
	gs.confirmedNightAction = nil
	gs.nightAcknowledged = nil
	gs.dawnReviewPending = false
	gs.pendingDawnDeathIDs = nil

	// Transition to Day
	gs.phase = game.GamePhaseDay
	gs.dayNumber++
	gs.resetDailyNominationLimitsLocked()
	events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay}})

	if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func (gs *GameSession) nightDemonCanKillLocked() bool {
	impIdx := gs.findLivingCharacterIndexLocked("imp")
	return impIdx != -1 && !gs.playerAbilityMalfunctioningLocked(impIdx)
}

func (gs *GameSession) impChoseSelfKillLocked(playerIndex int) bool {
	if !gs.playerIsImpLocked(playerIndex) || gs.playerAbilityMalfunctioningLocked(playerIndex) {
		return false
	}
	playerID := gs.players[playerIndex].ID
	for _, action := range gs.nightActions {
		if action.ActionType != game.NightActionKill || (action.ActorID != playerID && action.ActorID != gs.storytellerID) {
			continue
		}
		for _, targetID := range action.TargetIDs {
			if targetID == playerID {
				return true
			}
		}
	}
	return false
}

func (gs *GameSession) nightProtectedTargetsLocked() map[string]bool {
	protected := map[string]bool{}
	monkIdx := gs.findLivingCharacterIndexLocked("monk")
	if monkIdx == -1 || gs.playerAbilityMalfunctioningLocked(monkIdx) {
		return protected
	}
	for _, action := range gs.nightActions {
		if action.ActionType != game.NightActionProtect {
			continue
		}
		for _, targetID := range action.TargetIDs {
			protected[targetID] = true
		}
	}
	return protected
}

func (gs *GameSession) nightKillPreventedLocked(targetIndex int, protectedTargets map[string]bool) bool {
	target := gs.players[targetIndex]
	if protectedTargets[target.ID] {
		return true
	}
	if target.Character != nil &&
		target.Character.ID == "mayor" &&
		!gs.playerAbilityMalfunctioningLocked(targetIndex) {
		return true
	}
	return target.Character != nil &&
		target.Character.ID == "soldier" &&
		!gs.playerAbilityMalfunctioningLocked(targetIndex)
}
