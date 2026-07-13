package ws

import (
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

func TestGameSessionSetStoryteller(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p3", Name: "Charlie", IsAlive: true})

	result, err := gs.Apply(SetStorytellerCmd{
		SenderID:       "p1",
		TargetPlayerID: "p2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Updated {
		t.Error("expected Updated=true")
	}
	if gs.StorytellerID() != "p2" {
		t.Errorf("expected storyteller=p2, got %s", gs.StorytellerID())
	}
	if gs.OriginalPlayers() != 3 {
		t.Errorf("expected originalPlayers=3, got %d", gs.OriginalPlayers())
	}

	// Storyteller removed from player list
	players := gs.Players()
	for _, p := range players {
		if p.ID == "p2" {
			t.Error("storyteller should be removed from player list")
		}
	}
	if len(players) != 2 {
		t.Errorf("expected 2 players after storyteller set, got %d", len(players))
	}
}

func TestGameSessionSetStorytellerAlreadySet(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})

	gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p2"})

	_, err := gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p1"})
	if err == nil {
		t.Error("expected error when storyteller already set")
	}
}

func TestGameSessionSetStorytellerTargetNotFound(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})

	_, err := gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "nonexistent"})
	if err == nil {
		t.Error("expected error for non-existent target")
	}
}

func TestGameSessionAssignCharacters(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p3", Name: "Charlie", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p4", Name: "Dave", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p5", Name: "Eve", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p6", Name: "Frank", IsAlive: true})

	// Set storyteller (p1), remaining players: p2-p6 (5 players)
	gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p1"})

	result, err := gs.Apply(AssignCharactersCmd{
		SenderID: "p1",
		Assignments: map[string]string{
			"p2": "washerwoman",
			"p3": "librarian",
			"p4": "investigator",
			"p5": "poisoner",
			"p6": "imp",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 5 {
		t.Errorf("expected 5 events, got %d", len(result.Events))
	}
}

func TestGameSessionAssignCharactersInvalid(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})

	gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p1"})

	// 1 player but assigning 2 characters
	_, err := gs.Apply(AssignCharactersCmd{
		SenderID: "p1",
		Assignments: map[string]string{
			"p2": "imp",
			"p1": "washerwoman",
		},
	})
	if err == nil {
		t.Error("expected error for invalid assignment")
	}
}

func TestGameSessionAssignCharactersNonStoryteller(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p3", Name: "Charlie", IsAlive: true})

	gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p1"})

	// p3 (not storyteller) tries to assign
	_, err := gs.Apply(AssignCharactersCmd{
		SenderID:    "p3",
		Assignments: map[string]string{"p2": "imp"},
	})
	if err == nil {
		t.Error("expected error for non-storyteller assignment")
	}
}

func TestGameSessionAssignCharactersBogusPlayerIDs(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p3", Name: "Charlie", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p4", Name: "Dave", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p5", Name: "Eve", IsAlive: true})

	gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p1"})

	// Use non-existent playerIDs — should be rejected
	_, err := gs.Apply(AssignCharactersCmd{
		SenderID: "p1",
		Assignments: map[string]string{
			"x1": "washerwoman",
			"x2": "librarian",
			"x3": "investigator",
			"x4": "imp",
		},
	})
	if err == nil {
		t.Error("expected error for non-existent playerIDs in assignments")
	}
}

func TestStartGameRejectsNoActualPlayers(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "storyteller", Name: "Storyteller", IsAlive: true})

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}

	result, err := gs.Apply(StartGameCmd{SenderID: "storyteller"})
	if err == nil {
		t.Fatal("expected start game with no actual players to be rejected")
	}
	if err.Error() != "invalid character assignment for player count" {
		t.Fatalf("expected invalid assignment error, got %q", err.Error())
	}
	if result.Updated {
		t.Fatalf("expected rejected start to make no changes, got %#v", result)
	}
	if phase := gs.Phase(); phase != game.GamePhaseSetup {
		t.Fatalf("expected phase to remain setup, got %d", phase)
	}
}

func TestStartGameRevalidatesAssignedRoleDistribution(t *testing.T) {
	gs := NewGameSession()
	gs.SetPlayers([]game.Player{
		{ID: "storyteller", Name: "Storyteller", IsAlive: true},
		{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "librarian")},
		{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "investigator")},
		{ID: "p4", Name: "P4", IsAlive: true, Character: testCharacter(t, "chef")},
		{ID: "p5", Name: "P5", IsAlive: true, Character: testCharacter(t, "imp")},
	})

	if _, err := gs.Apply(SetStorytellerCmd{SenderID: "storyteller", TargetPlayerID: "storyteller"}); err != nil {
		t.Fatalf("SetStoryteller failed: %v", err)
	}

	_, err := gs.Apply(StartGameCmd{SenderID: "storyteller"})
	if err == nil {
		t.Fatal("expected invalid manual role distribution to be rejected")
	}
	if err.Error() != "invalid character assignment for player count" {
		t.Fatalf("expected invalid assignment error, got %q", err.Error())
	}
}

func TestGameSessionRejectsRawSubmittedEvents(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})

	result, err := gs.Apply(SubmitEventCmd{
		SenderID: "p1",
		Event: game.GameEvent{
			PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay},
		},
	})
	if err == nil {
		t.Fatal("expected raw submitted events to be rejected")
	}
	if err.Error() != "raw event submission is disabled; use explicit game commands" {
		t.Fatalf("expected raw event submission error, got %q", err.Error())
	}
	if result.Updated || len(result.Events) != 0 {
		t.Fatalf("expected rejected submit event to make no changes, got %#v", result)
	}
}

func TestGameSessionKickPlayerDuringSetup(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "creator", Name: "Creator", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})

	result, err := gs.Apply(KickPlayerCmd{
		SenderID:       "creator",
		TargetPlayerID: "p1",
	})
	if err != nil {
		t.Fatalf("unexpected kick error: %v", err)
	}
	if !result.Updated {
		t.Fatal("expected kick to update state")
	}
	if len(result.Events) != 1 || result.Events[0].PlayerLeft == nil || result.Events[0].PlayerLeft.PlayerID != "p1" {
		t.Fatalf("expected player left event for p1, got %#v", result.Events)
	}
	for _, player := range gs.Players() {
		if player.ID == "p1" {
			t.Fatal("expected p1 to be removed from session players")
		}
	}
}

func TestGameSessionKickPlayerRejectsAfterSetup(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "creator", Name: "Creator", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.phase = game.GamePhaseDay

	result, err := gs.Apply(KickPlayerCmd{
		SenderID:       "creator",
		TargetPlayerID: "p1",
	})
	if err == nil {
		t.Fatal("expected kick after setup to be rejected")
	}
	if err.Error() != "players can only be kicked during setup phase" {
		t.Fatalf("expected setup-only error, got %q", err.Error())
	}
	if result.Updated || len(result.Events) != 0 {
		t.Fatalf("expected rejected kick to make no changes, got %#v", result)
	}
}

func TestGameSessionUpdateRoomSettingsDuringSetup(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "creator", Name: "Creator", IsAlive: true})
	for _, playerID := range []string{"p1", "p2", "p3", "p4", "p5"} {
		gs.AddPlayer(game.Player{ID: playerID, Name: playerID, IsAlive: true})
	}
	gs.Apply(SetStorytellerCmd{SenderID: "creator", TargetPlayerID: "creator"})

	result, err := gs.Apply(UpdateRoomSettingsCmd{
		SenderID:   "creator",
		MaxPlayers: 6,
		ScriptID:   game.TroubleBrewingScriptID,
	})
	if err != nil {
		t.Fatalf("unexpected room settings error: %v", err)
	}
	if !result.Updated {
		t.Fatal("expected room settings update to mark session updated")
	}

	state := gs.StateForRoom("room-1")
	if state.ScriptID != game.TroubleBrewingScriptID {
		t.Fatalf("expected script %s, got %s", game.TroubleBrewingScriptID, state.ScriptID)
	}
}

func TestGameSessionUpdateRoomSettingsRejectsAfterSetup(t *testing.T) {
	gs := NewGameSession()
	gs.phase = game.GamePhaseDay

	result, err := gs.Apply(UpdateRoomSettingsCmd{
		SenderID:   "creator",
		MaxPlayers: 6,
	})
	if err == nil {
		t.Fatal("expected settings update after setup to be rejected")
	}
	if err.Error() != "room settings can only be updated during setup phase" {
		t.Fatalf("expected setup-only settings error, got %q", err.Error())
	}
	if result.Updated || len(result.Events) != 0 {
		t.Fatalf("expected rejected settings update to make no changes, got %#v", result)
	}
}

func TestGameSessionUpdateRoomSettingsRejectsTooFewMaxPlayers(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "creator", Name: "Creator", IsAlive: true})
	for _, playerID := range []string{"p1", "p2", "p3", "p4", "p5", "p6"} {
		gs.AddPlayer(game.Player{ID: playerID, Name: playerID, IsAlive: true})
	}
	gs.Apply(SetStorytellerCmd{SenderID: "creator", TargetPlayerID: "creator"})

	result, err := gs.Apply(UpdateRoomSettingsCmd{
		SenderID:   "creator",
		MaxPlayers: 5,
	})
	if err == nil {
		t.Fatal("expected maxPlayers below current player count to be rejected")
	}
	if err.Error() != "maxPlayers cannot be less than current player count" {
		t.Fatalf("expected current player count error, got %q", err.Error())
	}
	if result.Updated || len(result.Events) != 0 {
		t.Fatalf("expected rejected settings update to make no changes, got %#v", result)
	}
}

func TestGameSessionEndGameByStoryteller(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay

	result, err := gs.Apply(EndGameCmd{
		SenderID:    "storyteller",
		Winner:      game.TeamEvil,
		Description: "The town conceded.",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Updated {
		t.Fatal("expected end game to update state")
	}
	if len(result.Events) != 1 || result.Events[0].GameEnded == nil {
		t.Fatalf("expected one game ended event, got %#v", result.Events)
	}

	state := gs.StateForRoom("room-1")
	if state.Phase != game.GamePhaseFinished {
		t.Fatalf("expected finished phase, got %d", state.Phase)
	}
	if state.Winner == nil {
		t.Fatal("expected winner in room state")
	}
	if state.Winner.Winner != game.TeamEvil ||
		state.Winner.Reason != game.WinReasonStorytellerDecision ||
		state.Winner.Description != "The town conceded." {
		t.Fatalf("unexpected winner payload: %#v", state.Winner)
	}
}

func TestGameSessionEndGameRejectsInvalidWinner(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay

	result, err := gs.Apply(EndGameCmd{
		SenderID: "storyteller",
		Winner:   game.TeamUnspecified,
	})
	if err == nil {
		t.Fatal("expected invalid winner error")
	}
	if err.Error() != "winner must be good or evil" {
		t.Fatalf("expected invalid winner error, got %q", err.Error())
	}
	if result.Updated || len(result.Events) != 0 {
		t.Fatalf("expected rejected end game to make no changes, got %#v", result)
	}
	if phase := gs.Phase(); phase != game.GamePhaseDay {
		t.Fatalf("expected phase to remain day, got %d", phase)
	}
}

func TestGameSessionScarletWomanBecomesImpWhenDemonDiesWithFiveAlive(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay
	gs.dayNumber = 1
	gs.players = []game.Player{
		{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "librarian")},
		{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "investigator")},
		{ID: "scarlet", Name: "Scarlet", IsAlive: true, Character: testCharacter(t, "scarletwoman")},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
	}

	result, err := gs.Apply(ExecutePlayerCmd{
		SenderID: "storyteller",
		PlayerID: "imp",
	})
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	for _, event := range result.Events {
		if event.GameEnded != nil {
			t.Fatalf("expected game to continue after starpass, got gameEnded %#v", event.GameEnded)
		}
		if event.CharacterAssigned != nil {
			t.Fatalf("expected starpass not to broadcast public assignment, got %#v", event.CharacterAssigned)
		}
	}

	state := gs.StateForRoom("room-1")
	if state.Phase == game.GamePhaseFinished {
		t.Fatalf("expected game to continue after Scarlet Woman starpass, got phase %d", state.Phase)
	}
	scarlet := findPlayerInState(t, state, "scarlet")
	if scarlet.Character == nil || scarlet.Character.ID != "imp" {
		t.Fatalf("expected Scarlet Woman to become Imp, got %#v", scarlet.Character)
	}
	deadImp := findPlayerInState(t, state, "imp")
	if deadImp.IsAlive {
		t.Fatal("expected original Imp to be dead")
	}
}

func TestGameSessionPoisonedScarletWomanDoesNotStarpass(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay
	gs.dayNumber = 1
	poisonedUntil := gs.dayNumber
	gs.players = []game.Player{
		{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "librarian")},
		{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "investigator")},
		{ID: "scarlet", Name: "Scarlet", IsAlive: true, Character: testCharacter(t, "scarletwoman"), PoisonedUntil: &poisonedUntil},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
	}

	result, err := gs.Apply(ExecutePlayerCmd{
		SenderID: "storyteller",
		PlayerID: "imp",
	})
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if len(result.Events) == 0 || result.Events[len(result.Events)-1].GameEnded == nil {
		t.Fatalf("expected gameEnded after demon death with poisoned Scarlet Woman, got %#v", result.Events)
	}
	ended := result.Events[len(result.Events)-1].GameEnded
	if ended.Winner != game.TeamGood || ended.Reason != game.WinReasonImpExecuted {
		t.Fatalf("expected good win when poisoned Scarlet Woman cannot starpass, got %#v", ended)
	}
	scarlet := findPlayerInState(t, gs.StateForRoom("room-1"), "scarlet")
	if scarlet.Character == nil || scarlet.Character.ID != "scarletwoman" {
		t.Fatalf("expected poisoned Scarlet Woman to remain unchanged, got %#v", scarlet.Character)
	}
}

func TestGameSessionDemonDeathWinsWhenScarletWomanCannotStarpassWithFourAlive(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.phase = game.GamePhaseDay
	gs.dayNumber = 1
	gs.players = []game.Player{
		{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "librarian")},
		{ID: "scarlet", Name: "Scarlet", IsAlive: true, Character: testCharacter(t, "scarletwoman")},
		{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
	}

	result, err := gs.Apply(ExecutePlayerCmd{
		SenderID: "storyteller",
		PlayerID: "imp",
	})
	if err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if len(result.Events) == 0 || result.Events[len(result.Events)-1].GameEnded == nil {
		t.Fatalf("expected gameEnded after demon death without starpass, got %#v", result.Events)
	}

	state := gs.StateForRoom("room-1")
	if state.Phase != game.GamePhaseFinished {
		t.Fatalf("expected finished phase, got %d", state.Phase)
	}
	if state.Winner == nil || state.Winner.Winner != game.TeamGood || state.Winner.Reason != game.WinReasonImpExecuted {
		t.Fatalf("expected good win by demon death, got %#v", state.Winner)
	}
	scarlet := findPlayerInState(t, state, "scarlet")
	if scarlet.Character == nil || scarlet.Character.ID != "scarletwoman" {
		t.Fatalf("expected Scarlet Woman to remain unchanged, got %#v", scarlet.Character)
	}
}

func TestGameSessionResolveNightDoesNotKillMonkProtectedTarget(t *testing.T) {
	gs := nightProtectionSession(t)
	gs.nightActions = []game.NightAction{
		{ActorID: "storyteller", ActionType: game.NightActionProtect, TargetIDs: []string{"p1"}},
		{ActorID: "storyteller", ActionType: game.NightActionKill, TargetIDs: []string{"p1"}},
	}

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("unexpected resolve night error: %v", err)
	}
	for _, event := range result.Events {
		if event.PlayerDied != nil {
			t.Fatalf("expected protected target not to die, got %#v", event.PlayerDied)
		}
	}

	state := gs.StateForRoom("room-1")
	protected := findPlayerInState(t, state, "p1")
	if !protected.IsAlive {
		t.Fatal("expected Monk-protected target to remain alive")
	}
	if len(state.Deaths) != 0 {
		t.Fatalf("expected no death records for protected target, got %#v", state.Deaths)
	}
}

func TestGameSessionRejectsMonkSelfProtection(t *testing.T) {
	gs := nightProtectionSession(t)
	gs.nightWakeIndex = 0

	_, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionProtect),
		TargetIDs:  []string{"monk"},
	})
	if err == nil {
		t.Fatal("expected Monk self-protection to be rejected")
	}
	if err.Error() != "monk cannot protect themself" {
		t.Fatalf("expected Monk self-protection error, got %q", err.Error())
	}
}

func TestGameSessionResolveNightDoesNotKillSoldier(t *testing.T) {
	gs := nightProtectionSession(t)
	gs.nightActions = []game.NightAction{
		{ActorID: "storyteller", ActionType: game.NightActionKill, TargetIDs: []string{"soldier"}},
	}

	result, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"})
	if err != nil {
		t.Fatalf("unexpected resolve night error: %v", err)
	}
	for _, event := range result.Events {
		if event.PlayerDied != nil {
			t.Fatalf("expected Soldier not to die, got %#v", event.PlayerDied)
		}
	}

	state := gs.StateForRoom("room-1")
	soldier := findPlayerInState(t, state, "soldier")
	if !soldier.IsAlive {
		t.Fatal("expected Soldier to remain alive after night kill")
	}
	if len(state.Deaths) != 0 {
		t.Fatalf("expected no death records for Soldier, got %#v", state.Deaths)
	}
}

func TestGameSessionSubmitNightActionStoresResult(t *testing.T) {
	gs := nightProtectionSession(t)
	gs.nightWakeIndex = 1

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   "storyteller",
		ActionType: string(game.NightActionKill),
		TargetIDs:  []string{"p1"},
		Result:     "P1 dies at dawn.",
	})
	if err != nil {
		t.Fatalf("unexpected night action error: %v", err)
	}
	if len(gs.nightActions) != 1 {
		t.Fatalf("expected one stored night action, got %#v", gs.nightActions)
	}
	if gs.nightActions[0].Result != "P1 dies at dawn." {
		t.Fatalf("expected stored result, got %q", gs.nightActions[0].Result)
	}
	if len(result.Events) != 1 ||
		result.Events[0].NightActionSubmitted == nil ||
		result.Events[0].NightActionSubmitted.Result == nil ||
		*result.Events[0].NightActionSubmitted.Result != "P1 dies at dawn." {
		t.Fatalf("expected night action submitted result event, got %#v", result.Events)
	}
}

func TestGameSessionSlayerAbilityKillsDemon(t *testing.T) {
	gs := slayerSession(t)

	result, err := gs.Apply(UseSlayerAbilityCmd{
		SenderID:       "slayer",
		TargetPlayerID: "imp",
	})
	if err != nil {
		t.Fatalf("unexpected Slayer ability error: %v", err)
	}
	if len(result.Events) < 2 {
		t.Fatalf("expected death and game ended events, got %#v", result.Events)
	}
	if result.Events[0].PlayerDied == nil ||
		result.Events[0].PlayerDied.PlayerID != "imp" ||
		result.Events[0].PlayerDied.Cause != game.DeathCauseAbility {
		t.Fatalf("expected ability death for Imp, got %#v", result.Events)
	}

	state := gs.StateForRoom("room-1")
	imp := findPlayerInState(t, state, "imp")
	if imp.IsAlive {
		t.Fatal("expected Imp to die from Slayer ability")
	}
	if state.Winner == nil || state.Winner.Winner != game.TeamGood {
		t.Fatalf("expected good win after Slayer kills demon, got %#v", state.Winner)
	}
}

func TestGameSessionSlayerAbilityMissConsumesUse(t *testing.T) {
	gs := slayerSession(t)

	result, err := gs.Apply(UseSlayerAbilityCmd{
		SenderID:       "slayer",
		TargetPlayerID: "p1",
	})
	if err != nil {
		t.Fatalf("unexpected Slayer ability error: %v", err)
	}
	if len(result.Events) != 0 {
		t.Fatalf("expected missed Slayer shot to broadcast no events, got %#v", result.Events)
	}
	if !result.Updated {
		t.Fatal("expected missed Slayer shot to update used ability state")
	}
	state := gs.StateForRoom("room-1")
	p1 := findPlayerInState(t, state, "p1")
	if !p1.IsAlive {
		t.Fatal("expected non-demon target to remain alive")
	}

	_, err = gs.Apply(UseSlayerAbilityCmd{
		SenderID:       "slayer",
		TargetPlayerID: "imp",
	})
	if err == nil {
		t.Fatal("expected second Slayer ability use to be rejected")
	}
	if err.Error() != "Slayer ability already used" {
		t.Fatalf("expected already used error, got %q", err.Error())
	}
}

func TestGameSessionStorytellerCanKillPlayer(t *testing.T) {
	gs := slayerSession(t)

	result, err := gs.Apply(KillPlayerCmd{
		SenderID: "storyteller",
		PlayerID: "p1",
		Cause:    game.DeathCauseAbility,
	})
	if err != nil {
		t.Fatalf("unexpected manual kill error: %v", err)
	}
	if len(result.Events) != 1 ||
		result.Events[0].PlayerDied == nil ||
		result.Events[0].PlayerDied.PlayerID != "p1" ||
		result.Events[0].PlayerDied.Cause != game.DeathCauseAbility {
		t.Fatalf("expected p1 ability death event, got %#v", result.Events)
	}

	state := gs.StateForRoom("room-1")
	p1 := findPlayerInState(t, state, "p1")
	if p1.IsAlive {
		t.Fatal("expected p1 to be dead after storyteller kill")
	}
	if len(state.Deaths) != 1 ||
		state.Deaths[0].PlayerID != "p1" ||
		state.Deaths[0].Cause != game.DeathCauseAbility ||
		state.Deaths[0].KilledBy != "storyteller" {
		t.Fatalf("expected manual death record, got %#v", state.Deaths)
	}
}

func TestGameSessionKillPlayerRejectsNonStoryteller(t *testing.T) {
	gs := slayerSession(t)

	result, err := gs.Apply(KillPlayerCmd{
		SenderID: "p1",
		PlayerID: "p2",
		Cause:    game.DeathCauseAbility,
	})
	if err == nil {
		t.Fatal("expected non-storyteller manual kill to be rejected")
	}
	if err.Error() != "only the storyteller can kill players" {
		t.Fatalf("expected storyteller-only error, got %q", err.Error())
	}
	if result.Updated || len(result.Events) != 0 {
		t.Fatalf("expected rejected kill to make no changes, got %#v", result)
	}

	state := gs.StateForRoom("room-1")
	p2 := findPlayerInState(t, state, "p2")
	if !p2.IsAlive {
		t.Fatal("expected p2 to remain alive after rejected manual kill")
	}
}

func TestGameSessionKillPlayerRequiresDeathCause(t *testing.T) {
	gs := slayerSession(t)

	_, err := gs.Apply(KillPlayerCmd{
		SenderID: "storyteller",
		PlayerID: "p1",
	})
	if err == nil {
		t.Fatal("expected manual kill without death cause to be rejected")
	}
	if err.Error() != "death cause is required" {
		t.Fatalf("expected death cause error, got %q", err.Error())
	}
}

func TestGameSessionRejectsDuplicateNominatorForSameDay(t *testing.T) {
	gs := nominationLimitSession(t)

	if _, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p2"}); err != nil {
		t.Fatalf("unexpected first nomination error: %v", err)
	}
	resolveWithoutExecution(t, gs)

	_, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p3"})
	if err == nil {
		t.Fatal("expected duplicate nominator to be rejected")
	}
	if err.Error() != "player p1 has already nominated today" {
		t.Fatalf("expected duplicate nominator error, got %q", err.Error())
	}
}

func TestGameSessionRejectsDuplicateNomineeForSameDay(t *testing.T) {
	gs := nominationLimitSession(t)

	if _, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p2"}); err != nil {
		t.Fatalf("unexpected first nomination error: %v", err)
	}
	resolveWithoutExecution(t, gs)

	_, err := gs.Apply(NominateCmd{SenderID: "p3", NomineeID: "p2"})
	if err == nil {
		t.Fatal("expected duplicate nominee to be rejected")
	}
	if err.Error() != "player p2 has already been nominated today" {
		t.Fatalf("expected duplicate nominee error, got %q", err.Error())
	}
}

func TestGameSessionNominationLimitsResetOnNewDay(t *testing.T) {
	gs := nominationLimitSession(t)

	if _, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p2"}); err != nil {
		t.Fatalf("unexpected first nomination error: %v", err)
	}
	resolveWithoutExecution(t, gs)
	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err != nil {
		t.Fatalf("unexpected change to night error: %v", err)
	}
	gs.nightWakeIndex = len(gs.activeNightWakeStepsLocked())
	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("unexpected resolve night error: %v", err)
	}

	if _, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p2"}); err != nil {
		t.Fatalf("expected nomination to be allowed on a new day, got %v", err)
	}
}

func TestGameSessionNominationLimitsSurviveSnapshotRestore(t *testing.T) {
	gs := nominationLimitSession(t)

	if _, err := gs.Apply(NominateCmd{SenderID: "p1", NomineeID: "p2"}); err != nil {
		t.Fatalf("unexpected first nomination error: %v", err)
	}
	resolveWithoutExecution(t, gs)

	restored := newGameSessionFromSnapshot(gs.snapshot())

	_, err := restored.Apply(NominateCmd{SenderID: "p1", NomineeID: "p3"})
	if err == nil {
		t.Fatal("expected restored duplicate nominator to be rejected")
	}
	if err.Error() != "player p1 has already nominated today" {
		t.Fatalf("expected restored duplicate nominator error, got %q", err.Error())
	}
}

func TestGameSessionVirginExecutesTownsfolkNominatorOnFirstNomination(t *testing.T) {
	gs := virginSession(t)

	result, err := gs.Apply(NominateCmd{SenderID: "townsfolk", NomineeID: "virgin"})
	if err != nil {
		t.Fatalf("unexpected Virgin nomination error: %v", err)
	}

	if !eventListContainsPlayerDied(result.Events, "townsfolk", game.DeathCauseExecution) {
		t.Fatalf("expected Virgin ability to execute townsfolk nominator, got %#v", result.Events)
	}
	if phase := gs.Phase(); phase != game.GamePhaseNight {
		t.Fatalf("expected Virgin execution to end day and enter night, got %d", phase)
	}
	if nomination := gs.Nomination(); nomination != nil {
		t.Fatalf("expected no active voting nomination after Virgin execution, got %#v", nomination)
	}
	state := gs.StateForRoom("room-1")
	townsfolk := findPlayerInState(t, state, "townsfolk")
	if townsfolk.IsAlive {
		t.Fatal("expected townsfolk nominator to be dead")
	}
}

func TestGameSessionVirginFirstNominationByNonTownsfolkConsumesAbilityWithoutExecution(t *testing.T) {
	gs := virginSession(t)

	result, err := gs.Apply(NominateCmd{SenderID: "poisoner", NomineeID: "virgin"})
	if err != nil {
		t.Fatalf("unexpected minion nomination error: %v", err)
	}
	if eventListContainsPlayerDied(result.Events, "poisoner", game.DeathCauseExecution) {
		t.Fatalf("expected non-Townsfolk nominator not to be executed, got %#v", result.Events)
	}
	if phase := gs.Phase(); phase != game.GamePhaseVoting {
		t.Fatalf("expected normal voting after non-Townsfolk nominates Virgin, got %d", phase)
	}

	resolveNominationWithoutExecutionBy(t, gs, "virgin")
	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err != nil {
		t.Fatalf("unexpected change to night error: %v", err)
	}
	gs.nightWakeIndex = len(gs.activeNightWakeStepsLocked())
	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("unexpected resolve night error: %v", err)
	}

	secondResult, err := gs.Apply(NominateCmd{SenderID: "townsfolk", NomineeID: "virgin"})
	if err != nil {
		t.Fatalf("unexpected second Virgin nomination error: %v", err)
	}
	if eventListContainsPlayerDied(secondResult.Events, "townsfolk", game.DeathCauseExecution) {
		t.Fatalf("expected Virgin ability to be spent after first nomination, got %#v", secondResult.Events)
	}
	if phase := gs.Phase(); phase != game.GamePhaseVoting {
		t.Fatalf("expected second Virgin nomination to proceed to voting, got %d", phase)
	}
}

func TestGameSessionPoisonedVirginFirstNominationConsumesAbilityWithoutExecution(t *testing.T) {
	gs := virginSession(t)
	poisonedUntil := gs.dayNumber
	gs.players[0].PoisonedUntil = &poisonedUntil

	result, err := gs.Apply(NominateCmd{SenderID: "townsfolk", NomineeID: "virgin"})
	if err != nil {
		t.Fatalf("unexpected poisoned Virgin nomination error: %v", err)
	}
	if eventListContainsPlayerDied(result.Events, "townsfolk", game.DeathCauseExecution) {
		t.Fatalf("expected poisoned Virgin not to execute townsfolk nominator, got %#v", result.Events)
	}
	if phase := gs.Phase(); phase != game.GamePhaseVoting {
		t.Fatalf("expected normal voting after poisoned Virgin nomination, got %d", phase)
	}

	resolveNominationWithoutExecutionBy(t, gs, "virgin")
	if _, err := gs.Apply(ChangePhaseCmd{SenderID: "storyteller", Phase: game.GamePhaseNight}); err != nil {
		t.Fatalf("unexpected change to night error: %v", err)
	}
	gs.nightWakeIndex = len(gs.activeNightWakeStepsLocked())
	if _, err := gs.Apply(ResolveNightCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("unexpected resolve night error: %v", err)
	}

	secondResult, err := gs.Apply(NominateCmd{SenderID: "townsfolk", NomineeID: "virgin"})
	if err != nil {
		t.Fatalf("unexpected second Virgin nomination error: %v", err)
	}
	if eventListContainsPlayerDied(secondResult.Events, "townsfolk", game.DeathCauseExecution) {
		t.Fatalf("expected poisoned first nomination to spend Virgin ability, got %#v", secondResult.Events)
	}
}

func TestGameSessionVirginAbilityUsageSurvivesSnapshotRestore(t *testing.T) {
	gs := virginSession(t)
	gs.virginAbilityUsed = map[string]bool{"virgin": true}

	restored := newGameSessionFromSnapshot(gs.snapshot())

	result, err := restored.Apply(NominateCmd{SenderID: "townsfolk", NomineeID: "virgin"})
	if err != nil {
		t.Fatalf("unexpected restored Virgin nomination error: %v", err)
	}
	if eventListContainsPlayerDied(result.Events, "townsfolk", game.DeathCauseExecution) {
		t.Fatalf("expected restored Virgin ability usage to prevent execution, got %#v", result.Events)
	}
	if phase := restored.Phase(); phase != game.GamePhaseVoting {
		t.Fatalf("expected restored nomination to proceed to voting, got %d", phase)
	}
}

func TestGameSessionRemovePlayer(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})

	gs.RemovePlayer("p2")

	players := gs.Players()
	if len(players) != 1 {
		t.Errorf("expected 1 player, got %d", len(players))
	}
	if players[0].ID != "p1" {
		t.Errorf("expected p1, got %s", players[0].ID)
	}
}

func TestGameSessionStateForRoom(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})
	gs.AddPlayer(game.Player{ID: "p2", Name: "Bob", IsAlive: true})

	gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p2"})

	state := gs.StateForRoom("room-1")
	if state.RoomID != "room-1" {
		t.Errorf("expected room-1, got %s", state.RoomID)
	}
	if state.StorytellerID != "p2" {
		t.Errorf("expected storyteller=p2, got %s", state.StorytellerID)
	}
	if len(state.Players) != 1 {
		t.Errorf("expected 1 player (storyteller removed), got %d", len(state.Players))
	}
}

func TestGameSessionFinishedStateRevealsCharactersToPlayers(t *testing.T) {
	gs := NewGameSession()
	gs.storytellerID = "storyteller"
	gs.players = []game.Player{
		{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
		{ID: "p2", Name: "P2", IsAlive: false, Character: testCharacter(t, "imp")},
	}

	hiddenState := gs.StateForRoomForRecipient("room-1", "p1")
	hidden := findPlayerInState(t, hiddenState, "p2")
	if hidden.Character != nil {
		t.Fatalf("expected non-recipient character to be hidden before finish, got %#v", hidden.Character)
	}

	gs.phase = game.GamePhaseFinished
	gs.winner = &game.GameEndedEvent{
		Winner:      game.TeamGood,
		Reason:      game.WinReasonImpExecuted,
		Description: "The Demon is dead — good wins!",
	}

	revealedState := gs.StateForRoomForRecipient("room-1", "p1")
	revealed := findPlayerInState(t, revealedState, "p2")
	if revealed.Character == nil || revealed.Character.ID != "imp" {
		t.Fatalf("expected finished state to reveal p2 character, got %#v", revealed.Character)
	}
}

func TestGameSessionSpySeesAllCharactersAtNight(t *testing.T) {
	gs := spyVisibilitySession(t)

	state := gs.StateForRoomForRecipient("room-1", "spy")

	for _, player := range state.Players {
		if player.Character == nil {
			t.Fatalf("expected Spy to see %s character at night", player.ID)
		}
	}
	if state.CurrentNightWakeStep != nil || len(state.NightWakeSteps) != 0 {
		t.Fatal("Spy must not see storyteller night management")
	}
	for _, player := range state.Players {
		if player.PoisonedUntil != nil {
			t.Fatal("Spy must not see poisoning state")
		}
	}
}

func TestGameSessionSpyDoesNotSeeAllCharactersDuringDay(t *testing.T) {
	gs := spyVisibilitySession(t)
	gs.phase = game.GamePhaseDay

	state := gs.StateForRoomForRecipient("room-1", "spy")

	visible := findPlayerInState(t, state, "spy")
	if visible.Character == nil || visible.Character.ID != "spy" {
		t.Fatalf("expected Spy to see their own character during day, got %#v", visible.Character)
	}
	hidden := findPlayerInState(t, state, "imp")
	if hidden.Character != nil {
		t.Fatalf("expected Spy not to see other characters during day, got %#v", hidden.Character)
	}
}

func TestGameSessionPoisonedSpyDoesNotSeeAllCharactersAtNight(t *testing.T) {
	gs := spyVisibilitySession(t)
	poisonedUntil := gs.dayNumber
	gs.players[0].PoisonedUntil = &poisonedUntil

	state := gs.StateForRoomForRecipient("room-1", "spy")

	visible := findPlayerInState(t, state, "spy")
	if visible.Character == nil || visible.Character.ID != "spy" {
		t.Fatalf("expected poisoned Spy to see their own character, got %#v", visible.Character)
	}
	hidden := findPlayerInState(t, state, "imp")
	if hidden.Character != nil {
		t.Fatalf("expected poisoned Spy not to see other characters, got %#v", hidden.Character)
	}
}

func TestGameSessionStateForRoomReturnsImmutableSnapshot(t *testing.T) {
	gs := NewGameSession()
	gs.players = []game.Player{
		{ID: "p1", Name: "Alice", IsAlive: true},
		{ID: "p2", Name: "Bob", IsAlive: true},
	}
	gs.phase = game.GamePhaseVoting
	gs.nomination = &game.Nomination{
		NominatorID: "p1",
		NomineeID:   "p2",
		Votes:       map[string]bool{"p1": true},
	}
	gs.deaths = []game.DeathRecord{{PlayerID: "p2", Cause: game.DeathCauseExecution, DayNumber: 1}}
	gs.winner = &game.GameEndedEvent{
		Winner:      game.TeamGood,
		Reason:      game.WinReasonImpExecuted,
		Description: "good wins",
	}

	state := gs.StateForRoom("room-1")

	gs.nomination.Resolved = true
	gs.nomination.Votes["p2"] = false
	gs.deaths[0].PlayerID = "p1"
	gs.winner.Winner = game.TeamEvil

	if state.Nomination == nil {
		t.Fatal("expected nomination snapshot")
	}
	if state.Nomination.Resolved {
		t.Fatal("expected nomination snapshot not to reflect later resolved mutation")
	}
	if _, exists := state.Nomination.Votes["p2"]; exists {
		t.Fatal("expected nomination vote map snapshot not to reflect later mutation")
	}
	if state.Deaths[0].PlayerID != "p2" {
		t.Fatalf("expected death snapshot to remain p2, got %s", state.Deaths[0].PlayerID)
	}
	if state.Winner == nil || state.Winner.Winner != game.TeamGood {
		t.Fatalf("expected winner snapshot to remain good, got %#v", state.Winner)
	}
}

type unknownCmd struct{}

func (unknownCmd) commandTag() {}

func TestGameSessionUnknownCommand(t *testing.T) {
	gs := NewGameSession()
	_, err := gs.Apply(unknownCmd{})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func testCharacter(t *testing.T, characterID string) *game.Character {
	t.Helper()
	character := game.GetCharacterByID(characterID)
	if character == nil {
		t.Fatalf("expected test character %s to exist", characterID)
	}
	return &game.Character{
		ID:      character.ID,
		Name:    character.Name,
		Team:    character.Team,
		Ability: character.Ability,
	}
}

func spyVisibilitySession(t *testing.T) *GameSession {
	t.Helper()
	return &GameSession{
		players: []game.Player{
			{ID: "spy", Name: "Spy", IsAlive: true, Character: testCharacter(t, "spy")},
			{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
			{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "chef")},
			{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
		},
		storytellerID:     "storyteller",
		scriptID:          game.TroubleBrewingScriptID,
		phase:             game.GamePhaseNight,
		dayNumber:         1,
		nightNumber:       1,
		ghostVotesUsed:    map[string]bool{},
		slayerUsed:        map[string]bool{},
		nominatorsToday:   map[string]bool{},
		nomineesToday:     map[string]bool{},
		virginAbilityUsed: map[string]bool{},
		butlerMasters:     map[string]string{},
	}
}

func nightProtectionSession(t *testing.T) *GameSession {
	t.Helper()
	return &GameSession{
		players: []game.Player{
			{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
			{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "librarian")},
			{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "investigator")},
			{ID: "monk", Name: "Monk", IsAlive: true, Character: testCharacter(t, "monk")},
			{ID: "soldier", Name: "Soldier", IsAlive: true, Character: testCharacter(t, "soldier")},
			{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
		},
		storytellerID:  "storyteller",
		scriptID:       game.TroubleBrewingScriptID,
		phase:          game.GamePhaseNight,
		dayNumber:      1,
		nightNumber:    2,
		nightWakeIndex: 2,
		ghostVotesUsed: map[string]bool{},
	}
}

func slayerSession(t *testing.T) *GameSession {
	t.Helper()
	return &GameSession{
		players: []game.Player{
			{ID: "slayer", Name: "Slayer", IsAlive: true, Character: testCharacter(t, "slayer")},
			{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
			{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "librarian")},
			{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "investigator")},
			{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
		},
		storytellerID:  "storyteller",
		scriptID:       game.TroubleBrewingScriptID,
		phase:          game.GamePhaseDay,
		dayNumber:      1,
		ghostVotesUsed: map[string]bool{},
		slayerUsed:     map[string]bool{},
	}
}

func nominationLimitSession(t *testing.T) *GameSession {
	t.Helper()
	return &GameSession{
		players: []game.Player{
			{ID: "p1", Name: "P1", IsAlive: true, Character: testCharacter(t, "washerwoman")},
			{ID: "p2", Name: "P2", IsAlive: true, Character: testCharacter(t, "librarian")},
			{ID: "p3", Name: "P3", IsAlive: true, Character: testCharacter(t, "investigator")},
			{ID: "p4", Name: "P4", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "p5", Name: "P5", IsAlive: true, Character: testCharacter(t, "imp")},
		},
		storytellerID:   "storyteller",
		scriptID:        game.TroubleBrewingScriptID,
		phase:           game.GamePhaseDay,
		dayNumber:       1,
		ghostVotesUsed:  map[string]bool{},
		slayerUsed:      map[string]bool{},
		nominatorsToday: map[string]bool{},
		nomineesToday:   map[string]bool{},
	}
}

func resolveWithoutExecution(t *testing.T, gs *GameSession) {
	t.Helper()
	resolveNominationWithoutExecutionBy(t, gs, "p1")
}

func resolveNominationWithoutExecutionBy(t *testing.T, gs *GameSession, voterID string) {
	t.Helper()
	no := false
	if _, err := gs.Apply(CastVoteCmd{SenderID: voterID, Decision: no}); err != nil {
		t.Fatalf("unexpected vote error: %v", err)
	}
	if _, err := gs.Apply(ResolveNominationCmd{SenderID: "storyteller"}); err != nil {
		t.Fatalf("unexpected resolve nomination error: %v", err)
	}
}

func virginSession(t *testing.T) *GameSession {
	t.Helper()
	return &GameSession{
		players: []game.Player{
			{ID: "virgin", Name: "Virgin", IsAlive: true, Character: testCharacter(t, "virgin")},
			{ID: "townsfolk", Name: "Townsfolk", IsAlive: true, Character: testCharacter(t, "washerwoman")},
			{ID: "outsider", Name: "Outsider", IsAlive: true, Character: testCharacter(t, "librarian")},
			{ID: "poisoner", Name: "Poisoner", IsAlive: true, Character: testCharacter(t, "poisoner")},
			{ID: "imp", Name: "Imp", IsAlive: true, Character: testCharacter(t, "imp")},
		},
		storytellerID:     "storyteller",
		scriptID:          game.TroubleBrewingScriptID,
		phase:             game.GamePhaseDay,
		dayNumber:         1,
		ghostVotesUsed:    map[string]bool{},
		slayerUsed:        map[string]bool{},
		nominatorsToday:   map[string]bool{},
		nomineesToday:     map[string]bool{},
		virginAbilityUsed: map[string]bool{},
	}
}

func eventListContainsPlayerDied(events []game.GameEvent, playerID string, cause game.DeathCause) bool {
	for _, event := range events {
		if event.PlayerDied == nil {
			continue
		}
		if event.PlayerDied.PlayerID == playerID && event.PlayerDied.Cause == cause {
			return true
		}
	}
	return false
}
