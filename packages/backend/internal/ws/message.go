package ws

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// Client-to-server message type constants.
const (
	MsgCreateRoom         = "CREATE_ROOM"
	MsgJoinRoom           = "JOIN_ROOM"
	MsgLeaveRoom          = "LEAVE_ROOM"
	MsgKickPlayer         = "KICK_PLAYER"
	MsgUpdateRoomSettings = "UPDATE_ROOM_SETTINGS"
	MsgSetStoryteller     = "SET_STORYTELLER"
	MsgAssignCharacters   = "ASSIGN_CHARACTERS"
	MsgSubmitEvent        = "SUBMIT_EVENT"
	MsgStartGame          = "START_GAME"
	MsgChangePhase        = "CHANGE_PHASE"
	MsgNominate           = "NOMINATE"
	MsgCastVote           = "CAST_VOTE"
	MsgResolveNomination  = "RESOLVE_NOMINATION"
	MsgExecutePlayer      = "EXECUTE_PLAYER"
	MsgUseSlayerAbility   = "USE_SLAYER_ABILITY"
	MsgSubmitNightAction  = "SUBMIT_NIGHT_ACTION"
	MsgResolveNight       = "RESOLVE_NIGHT"
	MsgEndGame            = "END_GAME"
)

// ClientMessage represents a message from client to server
type ClientMessage struct {
	Type            string            `json:"type"`
	RoomID          string            `json:"roomId,omitempty"`
	PlayerName      string            `json:"playerName,omitempty"`
	PlayerID        string            `json:"playerId,omitempty"`
	TargetPlayerID  string            `json:"targetPlayerId,omitempty"`
	ExecutePlayerID string            `json:"executePlayerId,omitempty"`
	MaxPlayers      int               `json:"maxPlayers,omitempty"`
	ScriptID        string            `json:"scriptId,omitempty"`
	Assignments     map[string]string `json:"assignments,omitempty"` // playerID -> characterID
	Event           *game.GameEvent   `json:"event,omitempty"`
	NomineeID       string            `json:"nomineeId,omitempty"`   // NOMINATE target
	Decision        *bool             `json:"decision,omitempty"`    // CAST_VOTE value
	Phase           ClientGamePhase   `json:"phase,omitempty"`       // CHANGE_PHASE target
	Winner          ClientTeam        `json:"winner,omitempty"`      // END_GAME winning team
	Reason          string            `json:"reason,omitempty"`      // END_GAME reason
	Description     string            `json:"description,omitempty"` // END_GAME description
	ActionType      string            `json:"actionType,omitempty"`  // SUBMIT_NIGHT_ACTION type
	TargetIDs       []string          `json:"targetIds,omitempty"`   // SUBMIT_NIGHT_ACTION targets
}

// ClientGamePhase accepts both the numeric protocol enum and the current
// frontend string names ("day", "night", etc.).
type ClientGamePhase game.GamePhase

func (p ClientGamePhase) GamePhase() game.GamePhase {
	return game.GamePhase(p)
}

func (p *ClientGamePhase) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		*p = ClientGamePhase(game.GamePhaseUnspecified)
		return nil
	}

	var numeric int
	if err := json.Unmarshal(raw, &numeric); err == nil {
		*p = ClientGamePhase(game.GamePhase(numeric))
		return nil
	}

	var named string
	if err := json.Unmarshal(raw, &named); err != nil {
		return err
	}

	phase, ok := parseClientGamePhase(named)
	if !ok {
		return fmt.Errorf("unknown game phase %q", named)
	}
	*p = ClientGamePhase(phase)
	return nil
}

func parseClientGamePhase(value string) (game.GamePhase, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "unspecified", "0":
		return game.GamePhaseUnspecified, true
	case "setup", "1":
		return game.GamePhaseSetup, true
	case "day", "2":
		return game.GamePhaseDay, true
	case "night", "3":
		return game.GamePhaseNight, true
	case "voting", "4":
		return game.GamePhaseVoting, true
	case "finished", "5":
		return game.GamePhaseFinished, true
	default:
		return game.GamePhaseUnspecified, false
	}
}

// ClientTeam accepts both the numeric protocol enum and compact frontend names.
type ClientTeam game.Team

func (t ClientTeam) Team() game.Team {
	return game.Team(t)
}

func (t *ClientTeam) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		*t = ClientTeam(game.TeamUnspecified)
		return nil
	}

	var numeric int
	if err := json.Unmarshal(raw, &numeric); err == nil {
		*t = ClientTeam(game.Team(numeric))
		return nil
	}

	var named string
	if err := json.Unmarshal(raw, &named); err != nil {
		return err
	}

	team, ok := parseClientTeam(named)
	if !ok {
		return fmt.Errorf("unknown team %q", named)
	}
	*t = ClientTeam(team)
	return nil
}

func parseClientTeam(value string) (game.Team, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "good", "1":
		return game.TeamGood, true
	case "evil", "2":
		return game.TeamEvil, true
	case "unspecified", "0":
		return game.TeamUnspecified, true
	default:
		return game.TeamUnspecified, false
	}
}

func (msg ClientMessage) targetPlayerID() string {
	if msg.TargetPlayerID != "" {
		return msg.TargetPlayerID
	}
	return msg.ExecutePlayerID
}

// ServerMessage represents a message from server to client
type ServerMessage struct {
	Type   string          `json:"type"`
	RoomID string          `json:"roomId,omitempty"`
	State  *RoomState      `json:"state,omitempty"`
	Event  *game.GameEvent `json:"event,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// RoomState represents the current state of a room
type RoomState struct {
	RoomID                string               `json:"roomId"`
	Players               []game.Player        `json:"players"`
	MaxPlayers            int                  `json:"maxPlayers"`
	ScriptID              string               `json:"scriptId"`
	ScriptName            string               `json:"scriptName"`
	CreatorID             string               `json:"creatorId,omitempty"`
	StorytellerID         string               `json:"storytellerId,omitempty"`
	Phase                 game.GamePhase       `json:"phase"`
	DayNumber             int32                `json:"dayNumber"`
	Nomination            *game.Nomination     `json:"nomination,omitempty"`
	Deaths                []game.DeathRecord   `json:"deaths,omitempty"`
	GhostVotesRemaining   []string             `json:"ghostVotesRemaining,omitempty"`
	NightWakeSteps        []game.NightWakeStep `json:"nightWakeSteps,omitempty"`
	CurrentNightWakeIndex int                  `json:"currentNightWakeIndex,omitempty"`
	CurrentNightWakeStep  *game.NightWakeStep  `json:"currentNightWakeStep,omitempty"`
	Winner                *game.GameEndedEvent `json:"winner,omitempty"`
}
