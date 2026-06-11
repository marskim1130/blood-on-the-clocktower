package ws

import "github.com/your-org/blood-on-the-clocktower/internal/game"

// Client-to-server message type constants.
const (
	MsgCreateRoom        = "CREATE_ROOM"
	MsgJoinRoom          = "JOIN_ROOM"
	MsgLeaveRoom         = "LEAVE_ROOM"
	MsgSetStoryteller    = "SET_STORYTELLER"
	MsgAssignCharacters  = "ASSIGN_CHARACTERS"
	MsgSubmitEvent       = "SUBMIT_EVENT"
	MsgStartGame         = "START_GAME"
	MsgChangePhase       = "CHANGE_PHASE"
	MsgNominate          = "NOMINATE"
	MsgCastVote          = "CAST_VOTE"
	MsgResolveNomination = "RESOLVE_NOMINATION"
	MsgExecutePlayer     = "EXECUTE_PLAYER"
	MsgSubmitNightAction = "SUBMIT_NIGHT_ACTION"
	MsgResolveNight      = "RESOLVE_NIGHT"
)

// ClientMessage represents a message from client to server
type ClientMessage struct {
	Type           string            `json:"type"`
	RoomID         string            `json:"roomId,omitempty"`
	PlayerName     string            `json:"playerName,omitempty"`
	PlayerID       string            `json:"playerId,omitempty"`
	TargetPlayerID string            `json:"targetPlayerId,omitempty"`
	MaxPlayers     int               `json:"maxPlayers,omitempty"`
	Assignments    map[string]string `json:"assignments,omitempty"` // playerID -> characterID
	Event          *game.GameEvent   `json:"event,omitempty"`
	NomineeID      string            `json:"nomineeId,omitempty"`   // NOMINATE target
	Decision       *bool             `json:"decision,omitempty"`    // CAST_VOTE value
	Phase          game.GamePhase    `json:"phase,omitempty"`      // CHANGE_PHASE target
	ActionType     string            `json:"actionType,omitempty"` // SUBMIT_NIGHT_ACTION type
	TargetIDs      []string          `json:"targetIds,omitempty"`  // SUBMIT_NIGHT_ACTION targets
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
	RoomID        string           `json:"roomId"`
	Players       []game.Player    `json:"players"`
	MaxPlayers    int              `json:"maxPlayers"`
	StorytellerID string           `json:"storytellerId,omitempty"`
	Phase         game.GamePhase   `json:"phase"`
	DayNumber     int32            `json:"dayNumber"`
	Nomination    *game.Nomination `json:"nomination,omitempty"`
	Deaths        []game.DeathRecord   `json:"deaths,omitempty"`
	Winner        *game.GameEndedEvent `json:"winner,omitempty"`
}
