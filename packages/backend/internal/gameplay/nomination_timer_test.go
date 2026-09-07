package gameplay

import (
	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"testing"
	"time"
)

func TestNominationTimerPauseResumeAndTimeoutRecordsNo(t *testing.T) {
	gs := &GameSession{storytellerID: "st", phase: game.GamePhaseVoting, players: []game.Player{{ID: "p1", IsAlive: false}, {ID: "p2", IsAlive: true}}, ghostVotesUsed: map[string]bool{}, nomination: &game.Nomination{Stage: game.NominationStageVoting, VoterOrder: []string{"p1", "p2"}, Votes: map[string]bool{}, DeadlineUnixMs: time.Now().Add(2 * time.Second).UnixMilli()}}
	if _, err := gs.Apply(ControlNominationTimerCmd{SenderID: "st", Action: "pause"}); err != nil {
		t.Fatal(err)
	}
	if !gs.Projection().Nomination.Paused {
		t.Fatal("timer must pause")
	}
	if _, err := gs.Apply(CastVoteCmd{SenderID: "p1", Decision: false}); err == nil {
		t.Fatal("paused ballot accepted vote")
	}
	if _, err := gs.Apply(ControlNominationTimerCmd{SenderID: "st", Action: "resume"}); err != nil {
		t.Fatal(err)
	}
	gs.nomination.DeadlineUnixMs = time.Now().Add(-time.Second).UnixMilli()
	deadline := gs.nomination.DeadlineUnixMs
	if _, err := gs.Apply(ExpireNominationTimerCmd{SenderID: "p2", DeadlineUnixMs: deadline}); err != nil {
		t.Fatal(err)
	}
	ballot := gs.Projection().Nomination
	if value, ok := ballot.Votes["p1"]; !ok || value || ballot.CurrentVoterIndex != 1 {
		t.Fatalf("timeout must record no and advance: %+v", ballot)
	}
	if gs.ghostVotesUsed["p1"] {
		t.Fatal("timeout spent ghost vote")
	}
	if _, err := gs.Apply(ExpireNominationTimerCmd{SenderID: "p2", DeadlineUnixMs: deadline}); err == nil {
		t.Fatal("stale timeout affected next seat")
	}
}
