package gameplay

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestButlerCannotVoteYesBeforeMasterVotesYes(t *testing.T) {
	gs := preparedButlerDay(t, false)
	startButlerNomination(t, gs)

	_, err := gs.Apply(CastVoteCmd{SenderID: "p1", Decision: true})
	if err == nil {
		t.Fatal("expected Butler yes vote before master to be rejected")
	}
	if err.Error() != "butler p1 cannot vote until master p2 votes yes" {
		t.Fatalf("expected Butler master vote error, got %q", err.Error())
	}

	if _, err := gs.Apply(CastVoteCmd{SenderID: "p2", Decision: true}); err != nil {
		t.Fatalf("expected master yes vote to be accepted: %v", err)
	}
	if _, err := gs.Apply(CastVoteCmd{SenderID: "p1", Decision: true}); err != nil {
		t.Fatalf("expected Butler yes vote after master to be accepted: %v", err)
	}
}

func TestButlerCanVoteNoBeforeMasterVotes(t *testing.T) {
	gs := preparedButlerDay(t, false)
	startButlerNomination(t, gs)

	if _, err := gs.Apply(CastVoteCmd{SenderID: "p1", Decision: false}); err != nil {
		t.Fatalf("expected Butler no vote before master to be accepted: %v", err)
	}
}

func TestPoisonedButlerCanVoteYesBeforeMasterVotes(t *testing.T) {
	gs := preparedButlerDay(t, true)
	startButlerNomination(t, gs)

	if _, err := gs.Apply(CastVoteCmd{SenderID: "p1", Decision: true}); err != nil {
		t.Fatalf("expected poisoned Butler yes vote to be accepted: %v", err)
	}
}

func TestButlerWithoutMasterCannotVoteYes(t *testing.T) {
	gs := preparedButlerDay(t, false)
	delete(gs.butlerMasters, "p1")
	startButlerNomination(t, gs)

	_, err := gs.Apply(CastVoteCmd{SenderID: "p1", Decision: true})
	if err == nil {
		t.Fatal("expected Butler without master to be rejected")
	}
	if err.Error() != "butler p1 has not chosen a master" {
		t.Fatalf("expected missing Butler master error, got %q", err.Error())
	}
}

func TestButlerMasterSurvivesSnapshotRestore(t *testing.T) {
	gs := preparedButlerDay(t, false)
	restored := newGameSessionFromSnapshot(gs.snapshot())
	startButlerNomination(t, restored)

	if _, err := restored.Apply(CastVoteCmd{SenderID: "p2", Decision: true}); err != nil {
		t.Fatalf("expected restored master yes vote to be accepted: %v", err)
	}
	if _, err := restored.Apply(CastVoteCmd{SenderID: "p1", Decision: true}); err != nil {
		t.Fatalf("expected restored Butler yes vote after master to be accepted: %v", err)
	}
}

func TestButlerCannotChooseThemselfAsMaster(t *testing.T) {
	gs := newStartedTypeHintGame(t, "butler", "washerwoman", "librarian", "chef", "poisoner", "imp")
	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	submitTypeHintNightAction(t, gs, game.NightActionPoison, []string{"p3"}, "")
	skipTypeHintGameToCharacter(t, gs, "butler")

	_, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionLearnMaster),
		TargetIDs:  []string{"p1"},
	})
	if err == nil {
		t.Fatal("expected Butler self-master choice to be rejected")
	}
	if err.Error() != "butler cannot choose themself as master" {
		t.Fatalf("expected Butler self-master error, got %q", err.Error())
	}
}

func preparedButlerDay(t *testing.T, poisonButler bool) *GameSession {
	t.Helper()

	gs := newStartedTypeHintGame(t, "butler", "washerwoman", "librarian", "chef", "poisoner", "imp")
	poisonTargetID := "p3"
	if poisonButler {
		poisonTargetID = "p1"
	}
	skipNightWakeStepsUntilAction(t, gs, game.NightActionPoison)
	submitTypeHintNightAction(t, gs, game.NightActionPoison, []string{poisonTargetID}, "")
	skipTypeHintGameToCharacter(t, gs, "butler")
	submitTypeHintNightAction(t, gs, game.NightActionLearnMaster, []string{"p2"}, "")

	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	if phase := gs.Phase(); phase != game.GamePhaseDay {
		t.Fatalf("expected day phase after resolving night, got %d", phase)
	}
	return gs
}

func startButlerNomination(t *testing.T, gs *GameSession) {
	t.Helper()

	if _, err := gs.Apply(NominateCmd{SenderID: "p2", NomineeID: "p6"}); err != nil {
		t.Fatalf("Nominate failed: %v", err)
	}
}
