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
