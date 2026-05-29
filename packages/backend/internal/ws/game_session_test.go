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

	// Set storyteller (p1), remaining players: p2-p5 (4 players)
	gs.Apply(SetStorytellerCmd{SenderID: "p1", TargetPlayerID: "p1"})

	result, err := gs.Apply(AssignCharactersCmd{
		SenderID: "p1",
		Assignments: map[string]string{
			"p2": "washerwoman",
			"p3": "librarian",
			"p4": "investigator",
			"p5": "imp",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 4 {
		t.Errorf("expected 4 events, got %d", len(result.Events))
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

func TestGameSessionSubmitEvent(t *testing.T) {
	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: "p1", Name: "Alice", IsAlive: true})

	result, err := gs.Apply(SubmitEventCmd{
		SenderID: "p1",
		Event: game.GameEvent{
			PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(result.Events))
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

type unknownCmd struct{}

func (unknownCmd) commandTag() {}

func TestGameSessionUnknownCommand(t *testing.T) {
	gs := NewGameSession()
	_, err := gs.Apply(unknownCmd{})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}
