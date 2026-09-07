package gameplay

import (
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestMayorWinsWhenDayEndsWithThreeAliveAndNoExecution(t *testing.T) {
	gs := mayorEndgameSession(t)

	result, err := gs.Apply(FinalizeDayCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ChangePhase failed: %v", err)
	}

	if len(result.Events) != 1 || result.Events[0].GameEnded == nil {
		t.Fatalf("expected Mayor game ended event, got %#v", result.Events)
	}
	ended := result.Events[0].GameEnded
	if ended.Winner != game.TeamGood || ended.Reason != game.WinReasonMayorEndgame {
		t.Fatalf("expected Mayor good win, got %#v", ended)
	}
	if phase := gs.Phase(); phase != game.GamePhaseFinished {
		t.Fatalf("expected finished phase after Mayor win, got %d", phase)
	}
}

func TestMayorDoesNotWinImmediatelyAfterNightKillLeavesThreeAlive(t *testing.T) {
	gs := mayorEndgameSession(t)
	gs.players[3].IsAlive = true
	gs.phase = game.GamePhaseNight
	gs.dayNumber = 1
	gs.nightWakeIndex = len(gs.activeNightWakeStepsLocked())
	gs.nightActions = []game.NightAction{
		{ActorID: "storyteller", ActionType: game.NightActionKill, TargetIDs: []string{"p2"}},
	}

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}

	for _, event := range result.Events {
		if event.GameEnded != nil {
			t.Fatalf("expected no Mayor win immediately after night kill, got %#v", event.GameEnded)
		}
	}
	if phase := gs.Phase(); phase != game.GamePhaseDay {
		t.Fatalf("expected day phase after resolving night, got %d", phase)
	}
	if state := gs.Projection(); state.Winner != nil {
		t.Fatalf("expected no winner after night kill to 3 alive, got %#v", state.Winner)
	}
}

func TestMayorDoesNotWinWhenExecutionHappenedToday(t *testing.T) {
	gs := mayorEndgameSession(t)
	gs.deaths = append(gs.deaths, game.DeathRecord{
		PlayerID:  "p4",
		Cause:     game.DeathCauseExecution,
		DayNumber: gs.dayNumber,
	})

	result, err := gs.Apply(FinalizeDayCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ChangePhase failed: %v", err)
	}
	for _, event := range result.Events {
		if event.GameEnded != nil {
			t.Fatalf("expected no Mayor win after execution today, got %#v", event.GameEnded)
		}
	}
	if phase := gs.Phase(); phase != game.GamePhaseNight {
		t.Fatalf("expected night phase when Mayor does not win, got %d", phase)
	}
}

func TestPoisonedMayorDoesNotWinWhenDayEndsWithThreeAlive(t *testing.T) {
	gs := mayorEndgameSession(t)
	poisonedUntil := gs.dayNumber
	gs.players[0].PoisonedUntil = &poisonedUntil

	result, err := gs.Apply(FinalizeDayCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ChangePhase failed: %v", err)
	}
	for _, event := range result.Events {
		if event.GameEnded != nil {
			t.Fatalf("expected poisoned Mayor not to win, got %#v", event.GameEnded)
		}
	}
	if phase := gs.Phase(); phase != game.GamePhaseNight {
		t.Fatalf("expected night phase when poisoned Mayor does not win, got %d", phase)
	}
}

func TestMayorTargetedByDemonNightKillSurvivesForStorytellerRedirection(t *testing.T) {
	gs := mayorNightKillSession(t, false)

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	for _, event := range result.Events {
		if event.PlayerDied != nil {
			t.Fatalf("expected Mayor night kill redirection to prevent automatic death, got %#v", event.PlayerDied)
		}
	}

	mayor := findPlayerInState(t, gs.Projection(), "mayor")
	if !mayor.IsAlive {
		t.Fatal("expected Mayor to survive automatic demon night kill")
	}
	if deaths := gs.Deaths(); len(deaths) != 0 {
		t.Fatalf("expected no death record before storyteller redirects Mayor death, got %#v", deaths)
	}
}

func TestPoisonedMayorTargetedByDemonNightKillDies(t *testing.T) {
	gs := mayorNightKillSession(t, true)

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	if len(result.Events) == 0 ||
		result.Events[0].PlayerDied == nil ||
		result.Events[0].PlayerDied.PlayerID != "mayor" ||
		result.Events[0].PlayerDied.Cause != game.DeathCauseNightKill {
		t.Fatalf("expected poisoned Mayor night kill death event, got %#v", result.Events)
	}

	mayor := findPlayerInState(t, gs.Projection(), "mayor")
	if mayor.IsAlive {
		t.Fatal("expected poisoned Mayor to die from demon night kill")
	}
	if deaths := gs.Deaths(); len(deaths) != 1 || deaths[0].PlayerID != "mayor" || deaths[0].Cause != game.DeathCauseNightKill {
		t.Fatalf("expected poisoned Mayor death record, got %#v", deaths)
	}
}

func TestStorytellerCanRedirectMayorNightKillWithManualDeath(t *testing.T) {
	gs := mayorNightKillSession(t, false)

	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	result, err := gs.Apply(KillPlayerCmd{
		SenderID: "storyteller",
		PlayerID: "p2",
		Cause:    game.DeathCauseNightKill,
	})
	if err != nil {
		t.Fatalf("KillPlayer failed: %v", err)
	}
	if len(result.Events) != 1 ||
		result.Events[0].PlayerDied == nil ||
		result.Events[0].PlayerDied.PlayerID != "p2" ||
		result.Events[0].PlayerDied.Cause != game.DeathCauseNightKill {
		t.Fatalf("expected redirected night kill death event, got %#v", result.Events)
	}

	state := gs.Projection()
	mayor := findPlayerInState(t, state, "mayor")
	if !mayor.IsAlive {
		t.Fatal("expected Mayor to remain alive after redirected death")
	}
	p2 := findPlayerInState(t, state, "p2")
	if p2.IsAlive {
		t.Fatal("expected redirected target to be dead")
	}
}

func TestSaintExecutionWinsForEvil(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay
	gs.dayNumber = 2
	gs.players = []game.Player{
		{ID: "saint", Name: "Saint", IsAlive: true, Character: testCharacter(t, "saint")},
		{ID: "townsfolk", Name: "Townsfolk", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "poisoner", Name: "Poisoner", IsAlive: true, Character: testCharacter(t, "poisoner")},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
	}

	result, err := gs.Apply(ExecutePlayerCmd{SenderID: "storyteller", PlayerID: "saint"})
	if err != nil {
		t.Fatalf("ExecutePlayer failed: %v", err)
	}
	if len(result.Events) == 0 || result.Events[len(result.Events)-1].GameEnded == nil {
		t.Fatalf("expected Saint execution game ended event, got %#v", result.Events)
	}
	ended := result.Events[len(result.Events)-1].GameEnded
	if ended.Winner != game.TeamEvil || ended.Reason != game.WinReasonSaintExecuted {
		t.Fatalf("expected Saint evil win, got %#v", ended)
	}
}

func TestPoisonedSaintExecutionDoesNotWinForEvil(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay
	gs.dayNumber = 2
	poisonedUntil := gs.dayNumber
	gs.players = []game.Player{
		{ID: "saint", Name: "Saint", IsAlive: true, Character: testCharacter(t, "saint"), PoisonedUntil: &poisonedUntil},
		{ID: "townsfolk", Name: "Townsfolk", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "poisoner", Name: "Poisoner", IsAlive: true, Character: testCharacter(t, "poisoner")},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
	}

	result, err := gs.Apply(ExecutePlayerCmd{SenderID: "storyteller", PlayerID: "saint"})
	if err != nil {
		t.Fatalf("ExecutePlayer failed: %v", err)
	}
	for _, event := range result.Events {
		if event.GameEnded != nil {
			t.Fatalf("expected poisoned Saint execution not to end game, got %#v", event.GameEnded)
		}
	}
	if phase := gs.Phase(); phase != game.GamePhaseDay {
		t.Fatalf("expected phase to remain day after poisoned Saint execution, got %d", phase)
	}
	if winner := gs.Projection().Winner; winner != nil {
		t.Fatalf("expected no winner after poisoned Saint execution, got %#v", winner)
	}
}

func TestImpSelfKillMakesLivingMinionTheImp(t *testing.T) {
	gs := impSelfKillSession(t, true)

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	for _, event := range result.Events {
		if event.GameEnded != nil {
			t.Fatalf("expected game to continue after Imp self-kill starpass, got %#v", event.GameEnded)
		}
	}

	state := gs.Projection()
	oldImp := findPlayerInState(t, state, "imp")
	if oldImp.IsAlive {
		t.Fatal("expected original Imp to be dead after self-kill")
	}
	newImp := findPlayerInState(t, state, "minion")
	if newImp.Character == nil || newImp.Character.ID != "imp" {
		t.Fatalf("expected living minion to become Imp, got %#v", newImp.Character)
	}
	if state.Winner != nil {
		t.Fatalf("expected no winner after Imp self-kill starpass, got %#v", state.Winner)
	}
}

func TestImpSelfKillWithoutLivingMinionWinsForGood(t *testing.T) {
	gs := impSelfKillSession(t, false)

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("ResolveNight failed: %v", err)
	}
	if len(result.Events) == 0 || result.Events[len(result.Events)-1].GameEnded == nil {
		t.Fatalf("expected good win when Imp self-kills without minion, got %#v", result.Events)
	}
	ended := result.Events[len(result.Events)-1].GameEnded
	if ended.Winner != game.TeamGood || ended.Reason != game.WinReasonImpExecuted {
		t.Fatalf("expected good demon-dead win without minion, got %#v", ended)
	}
}

func mayorEndgameSession(t *testing.T) *GameSession {
	t.Helper()

	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay
	gs.dayNumber = 2
	gs.players = []game.Player{
		{ID: "mayor", Name: "Mayor", IsAlive: true, Character: testCharacter(t, "mayor")},
		{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
		{ID: "p4", Name: "P4", IsAlive: false, Character: testCharacter(t, "poisoner")},
	}
	return gs
}

func mayorNightKillSession(t *testing.T, poisonedMayor bool) *GameSession {
	t.Helper()

	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseNight
	gs.dayNumber = 1
	gs.nightNumber = 1
	gs.players = []game.Player{
		{ID: "mayor", Name: "Mayor", IsAlive: true, Character: testCharacter(t, "mayor")},
		{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "chef")},
		{ID: "poisoner", Name: "Poisoner", IsAlive: true, Character: testCharacter(t, "poisoner")},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
	}
	if poisonedMayor {
		poisonedUntil := gs.dayNumber
		gs.players[0].PoisonedUntil = &poisonedUntil
	}
	gs.nightWakeIndex = len(gs.activeNightWakeStepsLocked())
	gs.nightActions = []game.NightAction{
		{ActorID: "storyteller", ActionType: game.NightActionKill, TargetIDs: []string{"mayor"}},
	}
	return gs
}

func impSelfKillSession(t *testing.T, withMinion bool) *GameSession {
	t.Helper()

	players := []game.Player{
		{ID: "townsfolk1", Name: "Townsfolk 1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "townsfolk2", Name: "Townsfolk 2", IsAlive: true, Character: testCharacter(t, "chef")},
		{ID: "townsfolk3", Name: "Townsfolk 3", IsAlive: true, Character: testCharacter(t, "empath")},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
	}
	if withMinion {
		players = append(players, game.Player{ID: "minion", Name: "Minion", IsAlive: true, Character: testCharacter(t, "poisoner")})
	}

	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseNight
	gs.dayNumber = 1
	gs.nightNumber = 1
	gs.players = players
	gs.nightWakeIndex = len(gs.activeNightWakeStepsLocked())
	gs.nightActions = []game.NightAction{
		{ActorID: "storyteller", ActionType: game.NightActionKill, TargetIDs: []string{"imp"}},
	}
	return gs
}
