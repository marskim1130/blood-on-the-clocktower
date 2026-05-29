package ws

import (
	"fmt"
	"sync"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// Command types for GameSession.Apply()
type SetStorytellerCmd struct {
	SenderID       string
	TargetPlayerID string
}

type AssignCharactersCmd struct {
	SenderID   string
	Assignments map[string]string // playerID -> characterID
}

type SubmitEventCmd struct {
	SenderID string
	Event    game.GameEvent
}

// Command is a marker interface for all commands.
type Command interface {
	commandTag()
}

func (SetStorytellerCmd) commandTag()  {}
func (AssignCharactersCmd) commandTag() {}
func (SubmitEventCmd) commandTag()      {}

// Result of applying a command.
type ApplyResult struct {
	Events  []game.GameEvent
	State   *RoomState
	Updated bool // true if state changed (triggers broadcast)
}

// GameSession owns game state and processes commands.
// No knowledge of rooms, connections, or broadcasting.
type GameSession struct {
	mu              sync.Mutex
	players         []game.Player
	storytellerID   string
	originalPlayers int
}

func NewGameSession() *GameSession {
	return &GameSession{}
}

func (gs *GameSession) SetPlayers(players []game.Player) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.players = players
}

func (gs *GameSession) AddPlayer(player game.Player) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.players = append(gs.players, player)
}

func (gs *GameSession) RemovePlayer(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	for i, p := range gs.players {
		if p.ID == playerID {
			gs.players = append(gs.players[:i], gs.players[i+1:]...)
			break
		}
	}
}

func (gs *GameSession) Players() []game.Player {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	result := make([]game.Player, len(gs.players))
	copy(result, gs.players)
	return result
}

func (gs *GameSession) StorytellerID() string {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.storytellerID
}

func (gs *GameSession) OriginalPlayers() int {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.originalPlayers
}

// Apply processes a command and returns events to be broadcast.
func (gs *GameSession) Apply(cmd Command) (ApplyResult, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	switch c := cmd.(type) {
	case SetStorytellerCmd:
		return gs.applySetStoryteller(c)
	case AssignCharactersCmd:
		return gs.applyAssignCharacters(c)
	case SubmitEventCmd:
		return gs.applySubmitEvent(c)
	default:
		return ApplyResult{}, fmt.Errorf("unknown command type")
	}
}

func (gs *GameSession) applySetStoryteller(cmd SetStorytellerCmd) (ApplyResult, error) {
	if gs.storytellerID != "" {
		return ApplyResult{}, fmt.Errorf("storyteller already set")
	}

	found := false
	for _, p := range gs.players {
		if p.ID == cmd.TargetPlayerID {
			found = true
			break
		}
	}
	if !found {
		return ApplyResult{}, fmt.Errorf("target player not found")
	}

	gs.storytellerID = cmd.TargetPlayerID
	gs.originalPlayers = len(gs.players)

	// Remove storyteller from player list
	var newPlayers []game.Player
	for _, p := range gs.players {
		if p.ID != cmd.TargetPlayerID {
			newPlayers = append(newPlayers, p)
		}
	}
	gs.players = newPlayers

	return ApplyResult{Updated: true}, nil
}

func (gs *GameSession) applyAssignCharacters(cmd AssignCharactersCmd) (ApplyResult, error) {
	if gs.storytellerID == "" || cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only storyteller can assign characters")
	}

	playerCount := len(gs.players)
	if !game.ValidateAssignment(cmd.Assignments, playerCount) {
		return ApplyResult{}, fmt.Errorf("invalid character assignment for player count")
	}

	// Assign characters
	for playerID, charID := range cmd.Assignments {
		charDef := game.GetCharacterByID(charID)
		if charDef == nil {
			continue
		}
		for i := range gs.players {
			if gs.players[i].ID == playerID {
				gs.players[i].Character = &game.Character{
					ID:      charDef.ID,
					Name:    charDef.Name,
					Team:    charDef.Team,
					Ability: charDef.Ability,
				}
			}
		}
	}

	// Build events
	var events []game.GameEvent
	for playerID, charID := range cmd.Assignments {
		charDef := game.GetCharacterByID(charID)
		if charDef == nil {
			continue
		}
		events = append(events, game.GameEvent{
			CharacterAssigned: &game.CharacterAssigned{
				PlayerID: playerID,
				Character: game.Character{
					ID:      charDef.ID,
					Name:    charDef.Name,
					Team:    charDef.Team,
					Ability: charDef.Ability,
				},
			},
		})
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func (gs *GameSession) applySubmitEvent(cmd SubmitEventCmd) (ApplyResult, error) {
	// Future: validate event against current state
	// For now, accept and broadcast
	return ApplyResult{Events: []game.GameEvent{cmd.Event}, Updated: true}, nil
}

// StateForRoom builds a RoomState snapshot.
func (gs *GameSession) StateForRoom(roomID string) *RoomState {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	players := make([]game.Player, len(gs.players))
	copy(players, gs.players)

	return &RoomState{
		RoomID:        roomID,
		Players:       players,
		StorytellerID: gs.storytellerID,
	}
}
