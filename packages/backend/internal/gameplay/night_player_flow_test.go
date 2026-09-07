package gameplay

import (
	"strings"
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestLockedRavenkeeperDawnSurvivesSnapshotAndCannotBeReplaced(t *testing.T) {
	gs := &GameSession{players: []game.Player{
		{ID: "raven", IsAlive: true, Character: testCharacter(t, "ravenkeeper")},
		{ID: "target", IsAlive: true, Character: testCharacter(t, "soldier")},
	}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 1}
	if _, err := gs.Apply(PrepareDawnCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmDawnCmd{SenderID: "storyteller", DeathPlayerIDs: []string{"raven"}}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := gs.MarshalSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := LoadGameSessionSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := restored.Apply(PrepareDawnCmd{SenderID: "storyteller"}); err == nil {
		t.Fatal("locked deaths could be re-prepared")
	}
	if _, err := restored.Apply(ConfirmDawnCmd{SenderID: "storyteller", DeathPlayerIDs: []string{"target"}}); err == nil {
		t.Fatal("locked deaths could be replaced")
	}
	if step := restored.ProjectionFor("raven").CurrentNightWakeStep; step == nil || step.CharacterID != "ravenkeeper" {
		t.Fatal("snapshot lost the private death-triggered wake")
	}
	if _, err := restored.Apply(SkipNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	view := restored.ProjectionFor("target")
	if findPlayerInState(t, view, "raven").IsAlive || !findPlayerInState(t, view, "target").IsAlive {
		t.Fatal("skip did not publish original confirmed deaths")
	}
}

func TestStorytellerProxyFinishesLockedRavenkeeperDawn(t *testing.T) {
	gs := &GameSession{players: []game.Player{
		{ID: "raven", IsAlive: true, Character: testCharacter(t, "ravenkeeper")},
		{ID: "target", IsAlive: true, Character: testCharacter(t, "soldier")},
	}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 1}
	if _, err := gs.Apply(PrepareDawnCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmDawnCmd{SenderID: "storyteller", DeathPlayerIDs: []string{"raven"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "storyteller", ActionType: string(game.NightActionLearnDied), TargetIDs: []string{"target"}}); err != nil {
		t.Fatal(err)
	}
	if view := gs.ProjectionFor("target"); view.Phase == game.GamePhaseNight || findPlayerInState(t, view, "raven").IsAlive {
		t.Fatal("proxy completion left confirmed dawn stuck")
	}
}

func TestRavenkeeperWakesAfterConfirmedDeathBeforePublicDawn(t *testing.T) {
	gs := &GameSession{players: []game.Player{
		{ID: "raven", Name: "守鸦", IsAlive: true, Character: testCharacter(t, "ravenkeeper")},
		{ID: "imp", Name: "恶魔", IsAlive: true, Character: testCharacter(t, "imp")},
		{ID: "target", Name: "目标", IsAlive: true, Character: testCharacter(t, "soldier")},
		{ID: "other", Name: "其他", IsAlive: true, Character: testCharacter(t, "mayor")},
	}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 2}
	if _, err := gs.Apply(SkipNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	if step := gs.ProjectionFor("raven").CurrentNightWakeStep; step != nil {
		t.Fatal("unconfirmed Ravenkeeper death woke the player")
	}
	if _, err := gs.Apply(PrepareDawnCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmDawnCmd{SenderID: "storyteller", DeathPlayerIDs: []string{"raven"}}); err != nil {
		t.Fatal(err)
	}
	view := gs.ProjectionFor("target")
	if view.Phase != game.GamePhaseNight || !findPlayerInState(t, view, "raven").IsAlive || len(view.Deaths) != 0 {
		t.Fatal("dawn death leaked before Ravenkeeper finishes")
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "raven", ActionType: string(game.NightActionLearnDied), TargetIDs: []string{"target"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	if result := gs.ProjectionFor("raven").ConfirmedNightAction.Result; result != "Soldier" {
		t.Fatalf("wrong private result: %q", result)
	}
	if _, err := gs.Apply(AcknowledgeNightActionCmd{SenderID: "raven"}); err != nil {
		t.Fatal(err)
	}
	view = gs.ProjectionFor("target")
	if view.Phase != game.GamePhaseDay || findPlayerInState(t, view, "raven").IsAlive || len(view.Deaths) != 1 || view.Deaths[0].Cause != "" {
		t.Fatal("acknowledgement must publish only the confirmed death without cause")
	}
}

func TestPoisonedSpyReceivesOnlyStorytellerSuppliedGrimoire(t *testing.T) {
	gs := spyVisibilitySession(t)
	poisonedUntil := int32(2)
	gs.players[0].PoisonedUntil = &poisonedUntil
	for gs.ProjectionFor("storyteller").CurrentNightWakeStep.CharacterID != "spy" {
		if _, err := gs.Apply(SkipNightActionCmd{SenderID: "storyteller"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "spy", ActionType: "show_grimoire"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller"}); err == nil {
		t.Fatal("poisoned Spy must not get an automatic true grimoire")
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller", Result: "1号 Spy · Chef；2号 P1 · Imp"}); err != nil {
		t.Fatal(err)
	}
	view := gs.ProjectionFor("spy")
	if view.ConfirmedNightAction == nil || view.ConfirmedNightAction.Result != "1号 Spy · Chef；2号 P1 · Imp" {
		t.Fatal("manual false grimoire was not delivered")
	}
	for _, player := range view.Players {
		if player.PoisonedUntil != nil || (player.ID != "spy" && player.Character != nil) {
			t.Fatal("true state leaked beside false grimoire")
		}
	}
}

func TestSpyReceivesGrimoireOnlyThroughConfirmedNightResult(t *testing.T) {
	gs := spyVisibilitySession(t)
	view := gs.ProjectionFor("spy")
	if findPlayerInState(t, view, "p1").Character != nil {
		t.Fatal("spy sees roles before being shown the grimoire")
	}
	for gs.ProjectionFor("storyteller").CurrentNightWakeStep != nil && gs.ProjectionFor("storyteller").CurrentNightWakeStep.CharacterID != "spy" {
		if _, err := gs.Apply(SkipNightActionCmd{SenderID: "storyteller"}); err != nil {
			t.Fatal(err)
		}
	}
	if step := gs.ProjectionFor("spy").CurrentNightWakeStep; step == nil || string(step.ActionType) != "show_grimoire" {
		t.Fatal("spy has no dedicated grimoire wake step")
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "spy", ActionType: "show_grimoire"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	view = gs.ProjectionFor("spy")
	if view.ConfirmedNightAction == nil || !strings.Contains(view.ConfirmedNightAction.Result, "P1") || !strings.Contains(view.ConfirmedNightAction.Result, "Washerwoman") {
		t.Fatal("spy has no readable grimoire")
	}
	if gs.ProjectionFor("p1").ConfirmedNightAction != nil {
		t.Fatal("grimoire leaked to other player")
	}
	if _, err := gs.Apply(AcknowledgeNightActionCmd{SenderID: "spy"}); err != nil {
		t.Fatal(err)
	}
	view = gs.ProjectionFor("spy")
	if view.ConfirmedNightAction != nil || findPlayerInState(t, view, "p1").Character != nil {
		t.Fatal("acknowledged grimoire remains visible")
	}
}

func TestStorytellerDawnDeathDoesNotTurnAnUnrelatedImpAttackIntoStarpass(t *testing.T) {
	gs := impSelfKillSession(t, true)
	gs.nightActions[0].TargetIDs = []string{"townsfolk1"}
	if _, err := gs.Apply(PrepareDawnCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmDawnCmd{SenderID: "storyteller", DeathPlayerIDs: []string{"imp"}}); err != nil {
		t.Fatal(err)
	}
	view := gs.ProjectionFor("storyteller")
	if minion := findPlayerInState(t, view, "minion"); minion.Character.ID == "imp" {
		t.Fatal("storyteller-selected demon death caused an unearned starpass")
	}
	if view.Winner == nil || view.Winner.Winner != game.TeamGood {
		t.Fatal("demon death without self-kill must allow good to win")
	}
}

func TestFiveAndSixPlayerGamesSkipEvilIntroductions(t *testing.T) {
	for _, roles := range [][]string{{"washerwoman", "chef", "empath", "poisoner", "imp"}, {"washerwoman", "chef", "empath", "butler", "poisoner", "imp"}} {
		gs := newStartedTypeHintGame(t, roles...)
		for _, step := range gs.ProjectionFor("storyteller").NightWakeSteps {
			if step.CharacterType != "" {
				t.Fatalf("%d-player game includes evil introduction: %+v", len(roles), step)
			}
		}
	}
}

func TestPoisonedInformationRequiresStorytellerProvidedResult(t *testing.T) {
	poisonedUntil := int32(1)
	gs := &GameSession{players: []game.Player{{ID: "empath", IsAlive: true, Character: testCharacter(t, "empath"), PoisonedUntil: &poisonedUntil}}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 2}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "empath", ActionType: string(game.NightActionLearnEvilNeighbors)}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller"}); err == nil {
		t.Fatal("missing private information must not be confirmed")
	}
	if view := gs.ProjectionFor("empath"); view.PendingNightAction == nil || view.ConfirmedNightAction != nil {
		t.Fatal("rejected confirmation must retain the pending request")
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller", Result: "1"}); err != nil {
		t.Fatal(err)
	}
	if result := gs.ProjectionFor("empath").ConfirmedNightAction.Result; result != "1" {
		t.Fatal("manual information was not delivered")
	}
}

func TestEmpathReadsNearestLivingNeighborsAcrossDeadSeats(t *testing.T) {
	gs := &GameSession{players: []game.Player{
		{ID: "empath", IsAlive: true, Character: testCharacter(t, "empath")},
		{ID: "dead-right", Character: testCharacter(t, "chef")},
		{ID: "evil", IsAlive: true, Character: testCharacter(t, "imp")},
		{ID: "good", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "dead-left", Character: testCharacter(t, "soldier")},
	}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 2, nightWakeIndex: 1}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "empath", ActionType: string(game.NightActionLearnEvilNeighbors)}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	if result := gs.ProjectionFor("empath").ConfirmedNightAction.Result; result != "1" {
		t.Fatalf("expected one evil living neighbor, got %q", result)
	}
}

func TestEvilIntroductionNamesAllTeammatesWithoutRevealingTheirCharacters(t *testing.T) {
	gs := &GameSession{players: []game.Player{
		{ID: "first", Name: "张三", IsAlive: true, Character: testCharacter(t, "poisoner")},
		{ID: "second", Name: "李四", IsAlive: true, Character: testCharacter(t, "baron")},
		{ID: "demon", Name: "王五", IsAlive: true, Character: testCharacter(t, "imp")},
	}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 1}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "first", ActionType: string(game.NightActionLearnDemon)}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	result := gs.ProjectionFor("first").ConfirmedNightAction.Result
	if !strings.Contains(result, "张三") || !strings.Contains(result, "李四") || !strings.Contains(result, "王五") {
		t.Fatalf("teammate missing: %s", result)
	}
	if _, err := gs.Apply(AcknowledgeNightActionCmd{SenderID: "first"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(AcknowledgeNightActionCmd{SenderID: "second"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "demon", ActionType: string(game.NightActionLearnMinion)}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatal(err)
	}
	result += gs.ProjectionFor("demon").ConfirmedNightAction.Result
	for _, forbidden := range []string{"Poisoner", "Baron", "Imp"} {
		if strings.Contains(result, forbidden) {
			t.Fatalf("evil introduction disclosed role %s: %s", forbidden, result)
		}
	}
}

func TestAcknowledgedMinionNoLongerReceivesPrivateResultWhileOthersRead(t *testing.T) {
	gs := &GameSession{players: []game.Player{
		{ID: "first", IsAlive: true, Character: testCharacter(t, "poisoner")},
		{ID: "second", IsAlive: true, Character: testCharacter(t, "baron")},
	}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 1}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "first", ActionType: string(game.NightActionLearnDemon)}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller", Result: "秘密"}); err != nil {
		t.Fatal(err)
	}
	if _, err := gs.Apply(AcknowledgeNightActionCmd{SenderID: "first"}); err != nil {
		t.Fatal(err)
	}
	first, second := gs.ProjectionFor("first"), gs.ProjectionFor("second")
	if first.ConfirmedNightAction != nil || first.CurrentNightWakeStep != nil {
		t.Fatal("acknowledged minion still receives secret")
	}
	if second.ConfirmedNightAction == nil {
		t.Fatal("other minion lost unread result")
	}
}

func TestInformationRecipientRequestsInformationWithoutChoosingTargets(t *testing.T) {
	gs := &GameSession{players: []game.Player{
		{ID: "washer", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "target", IsAlive: true, Character: testCharacter(t, "chef")},
	}, storytellerID: "storyteller", scriptID: game.TroubleBrewingScriptID, phase: game.GamePhaseNight, nightNumber: 1}
	view := gs.ProjectionFor("washer")
	if view.CurrentNightWakeStep == nil || view.CurrentNightWakeStep.MaxTargets != 0 {
		t.Fatal("information recipient must not be offered a target picker")
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "washer", ActionType: string(game.NightActionLearnTownsfolk)}); err != nil {
		t.Fatalf("request information: %v", err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller", TargetIDs: []string{"washer", "target"}, Result: "其中一人是厨师"}); err != nil {
		t.Fatalf("storyteller supplies targets: %v", err)
	}
	if result := gs.ProjectionFor("washer").ConfirmedNightAction; result == nil || len(result.TargetIDs) != 2 {
		t.Fatal("recipient must receive storyteller-selected information")
	}
}

func TestPlayerNightChoiceWaitsForStorytellerReview(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "poisoner-player", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "target", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		},
		storytellerID: "storyteller",
		scriptID:      game.TroubleBrewingScriptID,
		phase:         game.GamePhaseNight,
		nightNumber:   2,
	}

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "poisoner-player",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"target"},
	})
	if err != nil {
		t.Fatalf("player night choice failed: %v", err)
	}
	if !result.Updated || gs.pendingNightAction == nil {
		t.Fatalf("player choice must become a pending review: result=%+v pending=%+v", result, gs.pendingNightAction)
	}
	if gs.pendingNightAction.ActorID != "poisoner-player" || len(gs.pendingNightAction.TargetIDs) != 1 || gs.pendingNightAction.TargetIDs[0] != "target" {
		t.Fatalf("unexpected pending choice: %+v", gs.pendingNightAction)
	}
	if gs.nightWakeIndex != 0 || len(gs.nightActions) != 0 || gs.players[1].PoisonedUntil != nil {
		t.Fatalf("unreviewed choice changed authoritative night state: index=%d actions=%+v target=%+v", gs.nightWakeIndex, gs.nightActions, gs.players[1])
	}
}

func TestStorytellerConfirmationAppliesPendingChoiceButWaitsForPlayerAcknowledgement(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "poisoner-player", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "target", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		},
		storytellerID: "storyteller",
		scriptID:      game.TroubleBrewingScriptID,
		phase:         game.GamePhaseNight,
		nightNumber:   2,
	}
	if _, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "poisoner-player",
		ActionType: string(game.NightActionPoison),
		TargetIDs:  []string{"target"},
	}); err != nil {
		t.Fatalf("player night choice failed: %v", err)
	}

	if _, err := gs.Apply(ConfirmNightActionCmd{
		SenderID:  "storyteller",
		TargetIDs: []string{"target"},
	}); err != nil {
		t.Fatalf("storyteller confirmation failed: %v", err)
	}
	if gs.pendingNightAction != nil || gs.confirmedNightAction == nil {
		t.Fatalf("confirmation must move pending choice to acknowledged state: pending=%+v confirmed=%+v", gs.pendingNightAction, gs.confirmedNightAction)
	}
	if gs.confirmedNightAction.ActorID != "poisoner-player" || len(gs.nightActions) != 1 || gs.players[1].PoisonedUntil == nil {
		t.Fatalf("confirmed choice was not applied: confirmed=%+v actions=%+v target=%+v", gs.confirmedNightAction, gs.nightActions, gs.players[1])
	}
	if gs.nightWakeIndex != 0 {
		t.Fatalf("night marker advanced before player acknowledgement: %d", gs.nightWakeIndex)
	}
}

func TestActingPlayerAcknowledgementAdvancesToTheNextNightStep(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "poisoner-player", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "target", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		},
		storytellerID: "storyteller",
		scriptID:      game.TroubleBrewingScriptID,
		phase:         game.GamePhaseNight,
		nightNumber:   2,
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "poisoner-player", ActionType: string(game.NightActionPoison), TargetIDs: []string{"target"}}); err != nil {
		t.Fatalf("player night choice failed: %v", err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller", TargetIDs: []string{"target"}}); err != nil {
		t.Fatalf("storyteller confirmation failed: %v", err)
	}

	if _, err := gs.Apply(AcknowledgeNightActionCmd{SenderID: "poisoner-player"}); err != nil {
		t.Fatalf("player acknowledgement failed: %v", err)
	}
	if gs.nightWakeIndex != 1 || gs.confirmedNightAction != nil {
		t.Fatalf("acknowledgement must advance and clear the confirmed step: index=%d confirmed=%+v", gs.nightWakeIndex, gs.confirmedNightAction)
	}
	if len(gs.nightActions) != 1 || gs.players[1].PoisonedUntil == nil {
		t.Fatalf("acknowledgement must retain the confirmed action and effect: actions=%+v target=%+v", gs.nightActions, gs.players[1])
	}
}

func TestOnlyTheActingPlayerReceivesThePrivateNightTurnProjection(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "poisoner-player", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "bystander", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		},
		storytellerID: "storyteller",
		scriptID:      game.TroubleBrewingScriptID,
		phase:         game.GamePhaseNight,
		nightNumber:   2,
	}

	actorView := gs.ProjectionFor("poisoner-player")
	if actorView.CurrentNightWakeStep == nil || actorView.NightTurnStatus != NightTurnAwaitingPlayer {
		t.Fatalf("actor did not receive current private turn: %+v", actorView)
	}
	bystanderView := gs.ProjectionFor("bystander")
	if bystanderView.CurrentNightWakeStep != nil || bystanderView.NightTurnStatus != "" || bystanderView.PendingNightAction != nil {
		t.Fatalf("bystander received private night state: %+v", bystanderView)
	}

	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "poisoner-player", ActionType: string(game.NightActionPoison), TargetIDs: []string{"bystander"}}); err != nil {
		t.Fatalf("player night choice failed: %v", err)
	}
	actorView = gs.ProjectionFor("poisoner-player")
	if actorView.NightTurnStatus != NightTurnAwaitingStoryteller || actorView.PendingNightAction == nil {
		t.Fatalf("actor did not receive pending review state: %+v", actorView)
	}
	bystanderView = gs.ProjectionFor("bystander")
	if bystanderView.CurrentNightWakeStep != nil || bystanderView.PendingNightAction != nil {
		t.Fatalf("pending choice leaked to bystander: %+v", bystanderView)
	}
}

func TestStorytellerCanSkipAConfirmedStepWhenTheActorCannotAcknowledge(t *testing.T) {
	gs := &GameSession{
		players: []game.Player{
			{ID: "poisoner-player", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "target", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		},
		storytellerID: "storyteller",
		scriptID:      game.TroubleBrewingScriptID,
		phase:         game.GamePhaseNight,
		nightNumber:   2,
	}
	if _, err := gs.Apply(SubmitNightActionCmd{SenderID: "poisoner-player", ActionType: string(game.NightActionPoison), TargetIDs: []string{"target"}}); err != nil {
		t.Fatalf("player night choice failed: %v", err)
	}
	if _, err := gs.Apply(ConfirmNightActionCmd{SenderID: "storyteller", TargetIDs: []string{"target"}}); err != nil {
		t.Fatalf("storyteller confirmation failed: %v", err)
	}

	if _, err := gs.Apply(SkipNightActionCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("storyteller acknowledgement override failed: %v", err)
	}
	if gs.nightWakeIndex != 1 || gs.pendingNightAction != nil || gs.confirmedNightAction != nil || gs.nightAcknowledged != nil {
		t.Fatalf("skip must advance and clear transient turn state: index=%d pending=%+v confirmed=%+v acknowledgements=%+v", gs.nightWakeIndex, gs.pendingNightAction, gs.confirmedNightAction, gs.nightAcknowledged)
	}
	if len(gs.nightActions) != 1 || gs.players[1].PoisonedUntil == nil {
		t.Fatalf("skipping acknowledgement must retain confirmed action and effect: actions=%+v target=%+v", gs.nightActions, gs.players[1])
	}
}
