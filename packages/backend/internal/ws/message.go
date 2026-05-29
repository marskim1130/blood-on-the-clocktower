package ws

import "github.com/your-org/blood-on-the-clocktower/internal/game"

// ClientMessage represents a message from client to server
type ClientMessage struct {
	Type       string          `json:"type"`
	RoomID     string          `json:"roomId,omitempty"`
	PlayerName string          `json:"playerName,omitempty"`
	PlayerID   string          `json:"playerId,omitempty"`
	MaxPlayers int             `json:"maxPlayers,omitempty"`
	Event      *game.GameEvent `json:"event,omitempty"`
}

// ServerMessage represents a message from server to client
type ServerMessage struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"roomId,omitempty"`
	State   *RoomState      `json:"state,omitempty"`
	Event   *game.GameEvent `json:"event,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// RoomState represents the current state of a room
type RoomState struct {
	RoomID     string        `json:"roomId"`
	Players    []game.Player `json:"players"`
	MaxPlayers int           `json:"maxPlayers"`
}
