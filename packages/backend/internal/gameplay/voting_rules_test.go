package gameplay

import (
	"slices"
	"testing"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestDeadPlayerVotingNoKeepsGhostVote(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "dead", Name: "Dead", IsAlive: false},
			{ID: "alive", Name: "Alive", IsAlive: true},
		},
		phase: game.GamePhaseVoting,
		nomination: &game.Nomination{
			NominatorID:       "alive",
			NomineeID:         "dead",
			Votes:             map[string]bool{},
			VoterOrder:        []string{"dead", "alive"},
			CurrentVoterIndex: 0,
		},
		ghostVotesUsed: map[string]bool{},
	}

	if _, err := gs.Apply(CastVoteCmd{SenderID: "dead", Decision: false}); err != nil {
		t.Fatalf("dead player no vote failed: %v", err)
	}

	if gs.ghostVotesUsed["dead"] {
		t.Fatal("a no vote must not consume the dead player's ghost vote")
	}
	projection := gs.Projection()
	if len(projection.GhostVotesRemaining) != 1 || projection.GhostVotesRemaining[0] != "dead" {
		t.Fatalf("dead player must retain their ghost vote: %+v", projection.GhostVotesRemaining)
	}
}

func TestRejectedDuplicateYesVoteDoesNotSpendGhostVote(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "dead", Name: "Dead", IsAlive: false},
			{ID: "alive", Name: "Alive", IsAlive: true},
		},
		phase: game.GamePhaseVoting,
		nomination: &game.Nomination{
			NominatorID:       "alive",
			NomineeID:         "dead",
			Votes:             map[string]bool{"dead": false},
			VoterOrder:        []string{"dead", "alive"},
			CurrentVoterIndex: 0,
		},
		ghostVotesUsed: map[string]bool{},
	}

	if _, err := gs.Apply(CastVoteCmd{SenderID: "dead", Decision: true}); err == nil {
		t.Fatal("duplicate vote must be rejected")
	}
	if gs.ghostVotesUsed["dead"] {
		t.Fatal("a rejected duplicate vote must not consume the ghost vote")
	}
}

func TestResolvedMajorityPutsNomineeOnBlockWithoutExecuting(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true},
			{ID: "p2", IsAlive: true},
			{ID: "p3", IsAlive: true},
			{ID: "p4", IsAlive: true},
			{ID: "p5", IsAlive: true},
		},
		storytellerID: "storyteller",
		phase:         game.GamePhaseVoting,
		dayNumber:     1,
		nomination: &game.Nomination{
			NominatorID:       "p1",
			NomineeID:         "p2",
			Votes:             map[string]bool{"p1": true, "p2": true, "p3": true, "p4": false, "p5": false},
			VoterOrder:        []string{"p2", "p3", "p4", "p5", "p1"},
			CurrentVoterIndex: 5,
		},
		ghostVotesUsed: map[string]bool{},
	}

	if _, err := gs.Apply(ResolveNominationCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("resolve nomination failed: %v", err)
	}

	if !gs.players[1].IsAlive || len(gs.deaths) != 0 || gs.phase != game.GamePhaseDay {
		t.Fatalf("majority must not execute immediately: player=%+v deaths=%+v phase=%v", gs.players[1], gs.deaths, gs.phase)
	}
	projection := gs.Projection()
	if projection.ExecutionCandidateID != "p2" || projection.ExecutionCandidateVotes != 3 || projection.ExecutionTied {
		t.Fatalf("p2 must be solely on the block with three votes: %+v", projection)
	}
}

func TestStorytellerFinalizesDayByExecutingUniqueCandidateAndStartingNight(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true, Character: testCharacter(t, "slayer")},
			{ID: "p2", IsAlive: true, Character: testCharacter(t, "soldier")},
			{ID: "p3", IsAlive: true, Character: testCharacter(t, "mayor")},
			{ID: "p4", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "p5", IsAlive: true, Character: testCharacter(t, "imp")},
		},
		storytellerID: "storyteller",
		scriptID:      game.TroubleBrewingScriptID,
		phase:         game.GamePhaseDay,
		dayNumber:     1,
		nominationResults: []game.NominationResult{{
			DayNumber:     1,
			NominatorID:   "p2",
			NomineeID:     "p1",
			Votes:         map[string]bool{"p1": true, "p2": true, "p3": true},
			YesVotes:      3,
			RequiredVotes: 3,
		}},
		ghostVotesUsed: map[string]bool{},
	}

	result, err := gs.Apply(FinalizeDayCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("finalize day failed: %v", err)
	}

	if gs.players[0].IsAlive || gs.phase != game.GamePhaseNight {
		t.Fatalf("unique candidate must be executed before night: player=%+v phase=%v", gs.players[0], gs.phase)
	}
	if len(gs.deaths) != 1 || gs.deaths[0].PlayerID != "p1" || gs.deaths[0].Cause != game.DeathCauseExecution {
		t.Fatalf("unexpected execution record: %+v", gs.deaths)
	}
	if !eventListContainsPlayerDied(result.Events, "p1", game.DeathCauseExecution) {
		t.Fatalf("finalize result must publish p1 execution: %+v", result.Events)
	}
}

func TestChangePhaseCannotBypassDayFinalization(t *testing.T) {
	gs := &GameSession{
		storytellerID: "storyteller",
		phase:         game.GamePhaseDay,
		dayNumber:     1,
	}

	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err == nil {
		t.Fatal("CHANGE_PHASE must not bypass FINALIZE_DAY")
	}
	if gs.phase != game.GamePhaseDay {
		t.Fatalf("rejected phase bypass changed phase to %v", gs.phase)
	}
}

func TestNominationBuildsClockwiseVoterOrderWithNomineeLast(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true},
			{ID: "p2", IsAlive: true},
			{ID: "p3", IsAlive: false},
			{ID: "p4", IsAlive: true},
		},
		phase:             game.GamePhaseDay,
		dayNumber:         1,
		nominatorsToday:   map[string]bool{},
		nomineesToday:     map[string]bool{},
		virginAbilityUsed: map[string]bool{},
	}

	if _, err := gs.Apply(NominateCmd{SenderID: "p2", NomineeID: "p4"}); err != nil {
		t.Fatalf("nomination failed: %v", err)
	}

	want := []string{"p1", "p2", "p3", "p4"}
	if gs.nomination == nil || !slices.Equal(gs.nomination.VoterOrder, want) || gs.nomination.CurrentVoterIndex != 0 {
		t.Fatalf("clockwise voter order = %+v, index=%d, want %+v at 0", gs.nomination.VoterOrder, gs.nomination.CurrentVoterIndex, want)
	}
}

func TestOnlyCurrentClockwiseSeatCanVote(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true},
			{ID: "p2", IsAlive: true},
			{ID: "p3", IsAlive: true},
		},
		phase: game.GamePhaseVoting,
		nomination: &game.Nomination{
			NominatorID:       "p1",
			NomineeID:         "p2",
			Votes:             map[string]bool{},
			VoterOrder:        []string{"p2", "p3", "p1"},
			CurrentVoterIndex: 0,
		},
		ghostVotesUsed: map[string]bool{},
	}

	if _, err := gs.Apply(CastVoteCmd{SenderID: "p1", Decision: true}); err == nil {
		t.Fatal("a player must not vote before their clockwise turn")
	}
	if len(gs.nomination.Votes) != 0 || gs.nomination.CurrentVoterIndex != 0 {
		t.Fatalf("rejected out-of-turn vote changed ballot: %+v", gs.nomination)
	}
}

func TestCurrentClockwiseVoteAdvancesToNextSeat(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true},
			{ID: "p2", IsAlive: true},
			{ID: "p3", IsAlive: true},
		},
		phase: game.GamePhaseVoting,
		nomination: &game.Nomination{
			NominatorID:       "p1",
			NomineeID:         "p2",
			Votes:             map[string]bool{},
			VoterOrder:        []string{"p2", "p3", "p1"},
			CurrentVoterIndex: 0,
		},
		ghostVotesUsed: map[string]bool{},
	}

	if _, err := gs.Apply(CastVoteCmd{SenderID: "p2", Decision: true}); err != nil {
		t.Fatalf("current voter failed: %v", err)
	}
	if decision, ok := gs.nomination.Votes["p2"]; !ok || !decision {
		t.Fatalf("current vote was not recorded: %+v", gs.nomination.Votes)
	}
	if gs.nomination.CurrentVoterIndex != 1 {
		t.Fatalf("current voter index = %d, want 1", gs.nomination.CurrentVoterIndex)
	}
}

func TestStorytellerCannotResolveBeforeEverySeatHasARecordedDecision(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true},
			{ID: "p2", IsAlive: true},
			{ID: "p3", IsAlive: true},
		},
		storytellerID: "storyteller",
		phase:         game.GamePhaseVoting,
		nomination: &game.Nomination{
			NominatorID:       "p1",
			NomineeID:         "p2",
			Votes:             map[string]bool{"p2": true},
			VoterOrder:        []string{"p2", "p3", "p1"},
			CurrentVoterIndex: 1,
		},
	}

	if _, err := gs.Apply(ResolveNominationCmd{SenderID: "storyteller"}); err == nil {
		t.Fatal("storyteller must not resolve an incomplete clockwise ballot")
	}
	if gs.phase != game.GamePhaseVoting || gs.nomination == nil || len(gs.nominationResults) != 0 {
		t.Fatalf("rejected early resolution changed state: phase=%v nomination=%+v results=%+v", gs.phase, gs.nomination, gs.nominationResults)
	}
}

func TestStorytellerCanRecordNoForTheCurrentUnresponsiveSeat(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "dead", IsAlive: false},
			{ID: "alive", IsAlive: true},
		},
		storytellerID: "storyteller",
		phase:         game.GamePhaseVoting,
		nomination: &game.Nomination{
			NominatorID:       "alive",
			NomineeID:         "dead",
			Votes:             map[string]bool{},
			VoterOrder:        []string{"dead", "alive"},
			CurrentVoterIndex: 0,
		},
		ghostVotesUsed: map[string]bool{},
	}

	if _, err := gs.Apply(RecordVoteCmd{SenderID: "storyteller", VoterID: "dead", Decision: false}); err != nil {
		t.Fatalf("storyteller proxy no failed: %v", err)
	}
	if decision, ok := gs.nomination.Votes["dead"]; !ok || decision {
		t.Fatalf("proxy no was not recorded: %+v", gs.nomination.Votes)
	}
	if gs.nomination.CurrentVoterIndex != 1 || gs.ghostVotesUsed["dead"] {
		t.Fatalf("proxy no must advance without spending ghost vote: nomination=%+v ghost=%+v", gs.nomination, gs.ghostVotesUsed)
	}
}

func TestNominationStartsWithTimedAccusationBeforeVotesOpen(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "p1", IsAlive: true},
			{ID: "p2", IsAlive: true},
			{ID: "p3", IsAlive: true},
		},
		storytellerID:   "storyteller",
		phase:           game.GamePhaseDay,
		dayNumber:       1,
		nominatorsToday: map[string]bool{},
		nomineesToday:   map[string]bool{},
	}

	if _, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p2"}); err != nil {
		t.Fatalf("nomination failed: %v", err)
	}
	if gs.nomination == nil || gs.nomination.Stage != game.NominationStageAccusation || gs.nomination.DeadlineUnixMs <= 0 {
		t.Fatalf("nomination did not open the timed accusation: %+v", gs.nomination)
	}
	if _, err := gs.Apply(CastVoteCmd{SenderID: "p2", Decision: true}); err == nil {
		t.Fatal("vote must remain closed during the accusation")
	}
}

func TestStorytellerAdvancesAccusationToDefenseAndVoting(t *testing.T) {
	gs := &GameSession{
		players:         []game.Player{{ID: "p1", IsAlive: true}, {ID: "p2", IsAlive: true}},
		storytellerID:   "storyteller",
		phase:           game.GamePhaseDay,
		dayNumber:       1,
		nominatorsToday: map[string]bool{},
		nomineesToday:   map[string]bool{},
	}
	if _, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p2"}); err != nil {
		t.Fatalf("nomination failed: %v", err)
	}

	if _, err := gs.Apply(AdvanceNominationStageCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("advance to defense failed: %v", err)
	}
	if gs.nomination.Stage != game.NominationStageDefense || gs.nomination.DeadlineUnixMs <= time.Now().UnixMilli() {
		t.Fatalf("unexpected defense timer: %+v", gs.nomination)
	}
	if _, err := gs.Apply(AdvanceNominationStageCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("advance to voting failed: %v", err)
	}
	if gs.nomination.Stage != game.NominationStageVoting || gs.nomination.DeadlineUnixMs <= time.Now().UnixMilli() {
		t.Fatalf("unexpected first-voter timer: %+v", gs.nomination)
	}
}
