package gameplay

import (
	"fmt"
	"strings"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// ────────────────────────────────────────────────
// SubmitNightActionCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applySubmitNightAction(cmd SubmitNightActionCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("night actions can only be submitted during the night phase")
	}

	actionType := game.NightActionType(cmd.ActionType)
	var step *game.NightWakeStep
	if cmd.SenderID != gs.storytellerID {
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
		if err := gs.validateNightTargetsLocked(*step, cmd.TargetIDs); err != nil {
			return ApplyResult{}, err
		}
	} else {
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

	actorCharID := ""
	if cmd.SenderID == gs.storytellerID {
		if step != nil {
			actorCharID = step.CharacterID
		}
	}

	action := game.NightAction{
		ActorID:    cmd.SenderID,
		ActionType: actionType,
		TargetIDs:  cmd.TargetIDs,
		Result:     strings.TrimSpace(cmd.Result),
	}

	// Auto-compute ability results for information roles (if not poisoned)
	if cmd.SenderID == gs.storytellerID && action.Result == "" {
		switch action.ActionType {
		case game.NightActionLearnDemon:
			if step != nil && step.CharacterType == game.NightWakeCharacterTypeMinion {
				action.Result = gs.computeDemonInfoResultLocked()
			}
		case game.NightActionLearnTownsfolk:
			action.Result = gs.computeWasherwomanResultLocked(actorCharID, action.TargetIDs)
		case game.NightActionLearnOutsider:
			action.Result = gs.computeLibrarianResultLocked(actorCharID, action.TargetIDs)
		case game.NightActionLearnMinion:
			if step != nil && step.CharacterType == game.NightWakeCharacterTypeDemon {
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
		}
	}

	gs.nightActions = append(gs.nightActions, action)
	if cmd.SenderID == gs.storytellerID {
		gs.nightWakeIndex++
		// Apply poison effect immediately when storyteller submits poison action
		if action.ActionType == game.NightActionPoison {
			gs.applyPoisonEffectLocked(action.TargetIDs)
		}
		if action.ActionType == game.NightActionLearnMaster {
			gs.applyButlerMasterSelectionLocked(actorCharID, action.TargetIDs)
		}
	}

	events := []game.GameEvent{
		{NightActionSubmitted: &game.NightActionEvent{
			ActorID:    cmd.SenderID,
			ActionType: game.NightActionType(cmd.ActionType),
			TargetIDs:  cmd.TargetIDs,
			Result:     optionalString(action.Result),
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// ────────────────────────────────────────────────
// ResolveNightCmd
// ────────────────────────────────────────────────

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

	// Process adjudicated night actions. Player-submitted actions are treated as
	// private choices for the storyteller; only the storyteller can resolve deaths.
	var events []game.GameEvent
	demonDeathAliveCount := 0
	protectedTargets := gs.nightProtectedTargetsLocked()
	for _, action := range gs.nightActions {
		if action.ActorID == gs.storytellerID && action.ActionType == game.NightActionKill {
			if !gs.nightDemonCanKillLocked() {
				continue
			}
			for _, targetID := range action.TargetIDs {
				tIdx := gs.findPlayerIndex(targetID)
				if tIdx != -1 && gs.players[tIdx].IsAlive {
					if gs.nightKillPreventedLocked(tIdx, protectedTargets) {
						continue
					}
					aliveBeforeDeath := gs.alivePlayerCountLocked()
					impSelfKill := gs.playerIsImpLocked(tIdx)
					if aliveAtDemonDeath := gs.demonDeathAliveCountLocked(tIdx, aliveBeforeDeath); aliveAtDemonDeath > demonDeathAliveCount {
						demonDeathAliveCount = aliveAtDemonDeath
					}
					gs.players[tIdx].IsAlive = false
					gs.deaths = append(gs.deaths, game.DeathRecord{
						PlayerID:  targetID,
						Cause:     game.DeathCauseNightKill,
						DayNumber: gs.dayNumber,
						KilledBy:  action.ActorID,
					})
					events = append(events, game.GameEvent{
						PlayerDied: &game.PlayerDiedEvent{
							PlayerID:  targetID,
							Cause:     game.DeathCauseNightKill,
							DayNumber: gs.dayNumber,
						},
					})
					if impSelfKill && gs.applyImpSelfStarpassLocked(tIdx) {
						demonDeathAliveCount = 0
					}
				}
			}
		}
	}

	// Clear night actions for next night
	gs.nightActions = nil

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

func (gs *GameSession) nightProtectedTargetsLocked() map[string]bool {
	protected := map[string]bool{}
	monkIdx := gs.findLivingCharacterIndexLocked("monk")
	if monkIdx == -1 || gs.playerAbilityMalfunctioningLocked(monkIdx) {
		return protected
	}
	for _, action := range gs.nightActions {
		if action.ActorID != gs.storytellerID || action.ActionType != game.NightActionProtect {
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
