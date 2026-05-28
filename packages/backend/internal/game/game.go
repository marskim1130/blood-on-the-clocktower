package game

// GamePhase represents the current phase of the game
type GamePhase int

const (
	GamePhaseUnspecified GamePhase = iota
	GamePhaseSetup
	GamePhaseDay
	GamePhaseNight
	GamePhaseVoting
	GamePhaseFinished
)

// Team represents the team a character belongs to
type Team int

const (
	TeamUnspecified Team = iota
	TeamGood
	TeamEvil
)

// Character represents a character role in the game
type Character struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Team    Team   `json:"team"`
	Ability string `json:"ability"`
}

// Player represents a player in the game
type Player struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Character *Character `json:"character,omitempty"`
	IsAlive   bool       `json:"isAlive"`
	Votes     int32      `json:"votes"`
}

// GameState represents the complete state of a game
type GameState struct {
	ID           string            `json:"id"`
	Phase        GamePhase         `json:"phase"`
	Players      []Player          `json:"players"`
	DayNumber    int32             `json:"dayNumber"`
	Storyteller  *string           `json:"storyteller,omitempty"`
	Votes        map[string]string `json:"votes"` // voterID -> targetID
}

// GameEvent represents events that can occur in the game
type GameEvent struct {
	PlayerJoined     *PlayerJoined     `json:"playerJoined,omitempty"`
	PlayerLeft       *PlayerLeft       `json:"playerLeft,omitempty"`
	PhaseChanged     *PhaseChanged     `json:"phaseChanged,omitempty"`
	VoteCast         *VoteCast         `json:"voteCast,omitempty"`
	CharacterAssigned *CharacterAssigned `json:"characterAssigned,omitempty"`
}

type PlayerJoined struct {
	Player Player `json:"player"`
}

type PlayerLeft struct {
	PlayerID string `json:"playerId"`
}

type PhaseChanged struct {
	Phase GamePhase `json:"phase"`
}

type VoteCast struct {
	VoterID  string  `json:"voterId"`
	TargetID *string `json:"targetId,omitempty"`
}

type CharacterAssigned struct {
	PlayerID  string    `json:"playerId"`
	Character Character `json:"character"`
}
