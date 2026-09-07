package gameplay

import (
	"fmt"
	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"time"
)

func (gs *GameSession) applyControlNominationTimer(cmd ControlNominationTimerCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can control nomination timers")
	}
	n := gs.nomination
	if gs.phase != game.GamePhaseVoting || n == nil {
		return ApplyResult{}, fmt.Errorf("no active nomination")
	}
	if n.Stage == game.NominationStageVoting && n.CurrentVoterIndex >= len(n.VoterOrder) {
		return ApplyResult{}, fmt.Errorf("the clockwise ballot is complete")
	}
	now := time.Now().UnixMilli()
	switch cmd.Action {
	case "pause":
		if n.Paused {
			return ApplyResult{}, nil
		}
		n.RemainingMs = max(0, n.DeadlineUnixMs-now)
		n.Paused = true
		n.DeadlineUnixMs = 0
	case "resume":
		if !n.Paused {
			return ApplyResult{}, fmt.Errorf("nomination timer is not paused")
		}
		n.DeadlineUnixMs = now + max(1, n.RemainingMs)
		n.RemainingMs = 0
		n.Paused = false
	case "restart":
		duration := defaultVoterDuration
		if n.Stage == game.NominationStageAccusation {
			duration = defaultAccusationDuration
		}
		if n.Stage == game.NominationStageDefense {
			duration = defaultDefenseDuration
		}
		n.DeadlineUnixMs = now + duration.Milliseconds()
		n.RemainingMs = 0
		n.Paused = false
	default:
		return ApplyResult{}, fmt.Errorf("unknown nomination timer action")
	}
	return ApplyResult{Updated: true}, nil
}

func (gs *GameSession) applyExpireNominationTimer(cmd ExpireNominationTimerCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID && gs.findPlayerIndex(cmd.SenderID) < 0 {
		return ApplyResult{}, fmt.Errorf("only participants can advance expired timers")
	}
	n := gs.nomination
	if gs.phase != game.GamePhaseVoting || n == nil || n.Paused || n.DeadlineUnixMs <= 0 || n.DeadlineUnixMs != cmd.DeadlineUnixMs || time.Now().UnixMilli() < n.DeadlineUnixMs {
		return ApplyResult{}, fmt.Errorf("nomination deadline is not due or has changed")
	}
	if n.Stage == game.NominationStageAccusation || n.Stage == game.NominationStageDefense {
		return gs.applyAdvanceNominationStage(AdvanceNominationStageCmd{SenderID: gs.storytellerID})
	}
	if n.CurrentVoterIndex < 0 || n.CurrentVoterIndex >= len(n.VoterOrder) {
		return ApplyResult{}, fmt.Errorf("the clockwise ballot is complete")
	}
	return gs.applyCastVote(CastVoteCmd{SenderID: n.VoterOrder[n.CurrentVoterIndex], Decision: false})
}
