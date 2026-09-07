package ws

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestProtocolV2RavenkeeperCompletesBeforeMultipleDawnDeathsArePublished(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	raven := harness.join("raven", "Ravenkeeper")
	soldier := harness.join("soldier", "Soldier")
	slayer := harness.join("slayer", "Slayer")
	poisoner := harness.join("poisoner", "Poisoner")
	imp := harness.join("imp", "Imp")
	harness.command(storyteller, ClientMessage{Type: MsgSetStoryteller, TargetPlayerID: storyteller.id})
	ready := true
	for _, player := range []*protocolV2GameClient{raven, soldier, slayer, poisoner, imp} {
		harness.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}
	harness.command(storyteller, ClientMessage{Type: MsgAssignCharacters, Assignments: map[string]string{raven.id: "ravenkeeper", soldier.id: "soldier", slayer.id: "slayer", poisoner.id: "poisoner", imp.id: "imp"}})
	for _, player := range []*protocolV2GameClient{raven, soldier, slayer, poisoner, imp} {
		harness.command(player, ClientMessage{Type: MsgConfirmCharacter})
	}
	harness.command(storyteller, ClientMessage{Type: MsgStartGame})
	harness.command(storyteller, ClientMessage{Type: MsgSkipNightAction})
	harness.command(storyteller, ClientMessage{Type: MsgPrepareDawn})
	locked := harness.command(storyteller, ClientMessage{Type: MsgConfirmDawn, TargetIDs: []string{raven.id, slayer.id}})
	if locked.direct.State.DawnReviewPending || len(locked.direct.State.PendingDawnDeathIDs) != 2 {
		t.Fatal("storyteller lost locked dawn deaths")
	}
	for _, player := range []*protocolV2GameClient{soldier, slayer, poisoner, imp} {
		view := locked.broadcasts[player.id].State
		if view.Phase != game.GamePhaseNight || len(view.Deaths) != 0 || len(view.PendingDawnDeathIDs) != 0 || view.CurrentNightWakeStep != nil || !playerByID(t, view, raven.id).IsAlive {
			t.Fatal("private post-death wake leaked before dawn")
		}
	}
	harness.command(raven, ClientMessage{Type: MsgSubmitNightAction, ActionType: string(game.NightActionLearnDied), TargetIDs: []string{soldier.id}})
	confirmed := harness.command(storyteller, ClientMessage{Type: MsgConfirmNightAction})
	if result := confirmed.broadcasts[raven.id].State.ConfirmedNightAction; result == nil || result.Result != "Soldier" {
		t.Fatal("Ravenkeeper did not receive private character result")
	}
	finished := harness.command(raven, ClientMessage{Type: MsgAcknowledgeNightAction})
	view := finished.broadcasts[soldier.id].State
	if view.Phase != game.GamePhaseDay || len(view.Deaths) != 2 || playerByID(t, view, raven.id).IsAlive || playerByID(t, view, slayer.id).IsAlive {
		t.Fatal("dawn did not publish all confirmed deaths together")
	}
	for _, death := range view.Deaths {
		if death.Cause != "" || death.KilledBy != "" {
			t.Fatal("public dawn disclosed a death cause")
		}
	}
}

func TestProtocolV2PlayerNightChoiceReviewAndAcknowledgement(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	harness.command(storyteller, ClientMessage{Type: MsgUpdateRoomSettings, MaxPlayers: 7})
	poisoner := harness.join("p1", "Poisoner")
	washerwoman := harness.join("p2", "Washerwoman")
	slayer := harness.join("p3", "Slayer")
	mayor := harness.join("p4", "Mayor")
	imp := harness.join("p5", "Imp")
	soldier := harness.join("p6", "Soldier")
	chef := harness.join("p7", "Chef")
	harness.command(storyteller, ClientMessage{Type: MsgSetStoryteller, TargetPlayerID: storyteller.id})
	ready := true
	for _, player := range []*protocolV2GameClient{poisoner, washerwoman, slayer, mayor, imp, soldier, chef} {
		harness.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}
	harness.command(storyteller, ClientMessage{Type: MsgAssignCharacters, Assignments: map[string]string{
		poisoner.id:    "poisoner",
		washerwoman.id: "washerwoman",
		slayer.id:      "slayer",
		mayor.id:       "mayor",
		imp.id:         "imp",
		soldier.id:     "soldier",
		chef.id:        "chef",
	}})
	for _, player := range []*protocolV2GameClient{poisoner, washerwoman, slayer, mayor, imp, soldier, chef} {
		harness.command(player, ClientMessage{Type: MsgConfirmCharacter})
	}

	started := harness.command(storyteller, ClientMessage{Type: MsgStartGame})
	poisonerStart := started.broadcasts[poisoner.id].State
	if poisonerStart.CurrentNightWakeStep == nil || poisonerStart.NightTurnStatus != "awaiting_player" {
		t.Fatalf("current Minion did not receive private first-night turn: %+v", poisonerStart)
	}
	if bystander := started.broadcasts[washerwoman.id].State; bystander.CurrentNightWakeStep != nil || bystander.NightTurnStatus != "" {
		t.Fatalf("bystander received the Minion turn: %+v", bystander)
	}

	choice := harness.command(poisoner, ClientMessage{Type: MsgSubmitNightAction, ActionType: string(game.NightActionLearnDemon)})
	if choice.direct.State.PendingNightAction == nil || choice.direct.State.NightTurnStatus != "awaiting_storyteller" {
		t.Fatalf("player choice did not enter storyteller review: %+v", choice.direct.State)
	}
	if storytellerView := choice.broadcasts[storyteller.id].State; storytellerView.PendingNightAction == nil {
		t.Fatalf("storyteller did not receive pending choice: %+v", storytellerView)
	}
	if bystander := choice.broadcasts[washerwoman.id].State; bystander.PendingNightAction != nil || bystander.CurrentNightWakeStep != nil {
		t.Fatalf("pending Minion choice leaked to bystander: %+v", bystander)
	}

	confirmed := harness.command(storyteller, ClientMessage{Type: MsgConfirmNightAction})
	poisonerConfirmed := confirmed.broadcasts[poisoner.id].State
	if poisonerConfirmed.ConfirmedNightAction == nil || poisonerConfirmed.ConfirmedNightAction.Result == "" || poisonerConfirmed.NightTurnStatus != "awaiting_acknowledgement" {
		t.Fatalf("confirmed private result did not reach acting player: %+v", poisonerConfirmed)
	}
	if confirmed.direct.State.CurrentNightWakeIndex != 0 {
		t.Fatalf("storyteller confirmation advanced before acknowledgement: %+v", confirmed.direct.State)
	}

	acknowledged := harness.command(poisoner, ClientMessage{Type: MsgAcknowledgeNightAction})
	if acknowledged.direct.State.CurrentNightWakeStep != nil || acknowledged.direct.State.NightTurnStatus != "" {
		t.Fatalf("completed Minion lost neither private turn nor transient state: %+v", acknowledged.direct.State)
	}
	impView := acknowledged.broadcasts[imp.id].State
	if impView.CurrentNightWakeStep == nil || impView.CurrentNightWakeStep.ActionType != game.NightActionLearnMinion || impView.NightTurnStatus != "awaiting_player" {
		t.Fatalf("night did not advance to the Demon information step: %+v", impView)
	}
}
