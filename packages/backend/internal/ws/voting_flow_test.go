package ws

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestProtocolV2StorytellerFinalizesTheUniqueExecutionCandidate(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")
	p3 := harness.join("p3", "P3")
	p4 := harness.join("p4", "P4")
	p5 := harness.join("p5", "P5")
	harness.command(storyteller, ClientMessage{Type: MsgSetStoryteller, TargetPlayerID: storyteller.id})
	ready := true
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4, p5} {
		harness.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}
	harness.command(storyteller, ClientMessage{
		Type: MsgAssignCharacters,
		Assignments: map[string]string{
			p1.id: "slayer",
			p2.id: "soldier",
			p3.id: "mayor",
			p4.id: "poisoner",
			p5.id: "imp",
		},
	})
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4, p5} {
		harness.command(player, ClientMessage{Type: MsgConfirmCharacter})
	}
	harness.command(storyteller, ClientMessage{Type: MsgStartGame})
	harness.command(storyteller, ClientMessage{Type: MsgSubmitNightAction, ActionType: string(game.NightActionPoison), TargetIDs: []string{p3.id}})
	harness.command(storyteller, ClientMessage{Type: MsgPrepareDawn})
	harness.command(storyteller, ClientMessage{Type: MsgConfirmDawn})
	harness.command(p1, ClientMessage{Type: MsgNominate, NomineeID: p2.id})
	harness.command(storyteller, ClientMessage{Type: MsgAdvanceNominationStage})
	harness.command(storyteller, ClientMessage{Type: MsgAdvanceNominationStage})
	yes := true
	no := false
	harness.command(p3, ClientMessage{Type: MsgCastVote, Decision: &yes})
	proxy := harness.command(storyteller, ClientMessage{Type: MsgRecordVote, TargetPlayerID: p4.id, Decision: &no})
	if proxy.direct.State.Nomination == nil || proxy.direct.State.Nomination.CurrentVoterIndex != 2 || proxy.direct.State.Nomination.Votes[p4.id] {
		t.Fatalf("storyteller proxy no must advance the public clockwise ballot: %+v", proxy.direct.State.Nomination)
	}
	harness.command(storyteller, ClientMessage{Type: MsgRecordVote, TargetPlayerID: p5.id, Decision: &no})
	harness.command(p1, ClientMessage{Type: MsgCastVote, Decision: &yes})
	harness.command(p2, ClientMessage{Type: MsgCastVote, Decision: &yes})
	resolved := harness.command(storyteller, ClientMessage{Type: MsgResolveNomination})

	if resolved.direct.State.ExecutionCandidateID != p2.id || resolved.direct.State.ExecutionCandidateVotes != 3 || resolved.direct.State.ExecutionTied {
		t.Fatalf("resolved projection must put p2 uniquely on the block: %+v", resolved.direct.State)
	}
	if len(resolved.direct.State.NominationResults) != 1 {
		t.Fatalf("resolved projection must retain one public ballot: %+v", resolved.direct.State.NominationResults)
	}

	finalized := harness.command(storyteller, ClientMessage{Type: MsgFinalizeDay})
	if playerByID(t, finalized.direct.State, p2.id).IsAlive || finalized.direct.State.Phase != game.GamePhaseNight {
		t.Fatalf("finalized candidate must be executed before night: %+v", finalized.direct.State)
	}
}
