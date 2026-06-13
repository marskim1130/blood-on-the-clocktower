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

// DeathCause represents the reason a player died
type DeathCause string

const (
	DeathCauseExecution DeathCause = "execution"
	DeathCauseNightKill DeathCause = "night_kill"
	DeathCauseAbility   DeathCause = "ability"
)

// NightActionType represents the type of night action
type NightActionType string

const (
	NightActionPoison             NightActionType = "poison"
	NightActionProtect            NightActionType = "protect"
	NightActionKill               NightActionType = "kill"
	NightActionLearnTownsfolk     NightActionType = "learn_townsfolk"
	NightActionLearnOutsider      NightActionType = "learn_outsider"
	NightActionLearnMinion        NightActionType = "learn_minion"
	NightActionLearnEvilPairs     NightActionType = "learn_evil_pairs"
	NightActionLearnEvilNeighbors NightActionType = "learn_evil_neighbors"
	NightActionCheckDemon         NightActionType = "check_demon"
	NightActionLearnExecuted      NightActionType = "learn_executed"
	NightActionLearnDied          NightActionType = "learn_died"
	NightActionLearnMaster        NightActionType = "learn_master"
	NightActionLearnDemon         NightActionType = "learn_demon"
	NightActionChoosePlayer       NightActionType = "choose_player"
	NightActionNone               NightActionType = "none"
)

// WinReason represents the reason the game ended
type WinReason string

const (
	WinReasonImpExecuted         WinReason = "imp_executed"
	WinReasonMayorEndgame        WinReason = "mayor_endgame"
	WinReasonEvilMajority        WinReason = "evil_majority"
	WinReasonSaintExecuted       WinReason = "saint_executed"
	WinReasonImpStarpass         WinReason = "imp_starpass"
	WinReasonStorytellerDecision WinReason = "storyteller_decision"
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
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Character     *Character `json:"character,omitempty"`
	IsAlive       bool       `json:"isAlive"`
	Votes         int32      `json:"votes"`
	PoisonedUntil *int32     `json:"poisonedUntil,omitempty"`
}

// DeathRecord represents a record of a player's death
type DeathRecord struct {
	PlayerID  string     `json:"playerId"`
	Cause     DeathCause `json:"cause"`
	DayNumber int32      `json:"dayNumber"`
	KilledBy  string     `json:"killedBy,omitempty"`
}

// Nomination represents an active nomination for execution
type Nomination struct {
	NominatorID string          `json:"nominatorId"`
	NomineeID   string          `json:"nomineeId"`
	Votes       map[string]bool `json:"votes"` // voterID -> decision (true=yes, false=no)
	Resolved    bool            `json:"resolved"`
}

// NightAction represents a night action submitted by a player
type NightAction struct {
	ActorID    string          `json:"actorId"`
	ActionType NightActionType `json:"actionType"`
	TargetIDs  []string        `json:"targetIds"`
	Result     string          `json:"result,omitempty"`
}

// NightWakeStep describes one storyteller-facing wake step for the current script.
type NightWakeStep struct {
	CharacterID   string          `json:"characterId"`
	CharacterType string          `json:"characterType,omitempty"`
	Order         int             `json:"order"`
	ActionType    NightActionType `json:"actionType"`
	Prompt        string          `json:"prompt"`
	MinTargets    int             `json:"minTargets"`
	MaxTargets    int             `json:"maxTargets"`
}

// GameState represents the complete state of a game
type GameState struct {
	ID           string            `json:"id"`
	Phase        GamePhase         `json:"phase"`
	Players      []Player          `json:"players"`
	DayNumber    int32             `json:"dayNumber"`
	Storyteller  *string           `json:"storyteller,omitempty"`
	Votes        map[string]string `json:"votes"` // voterID -> targetID
	Deaths       []DeathRecord     `json:"deaths"`
	Nomination   *Nomination       `json:"nomination,omitempty"`
	NightActions []NightAction     `json:"nightActions"`
	Winner       *Team             `json:"winner,omitempty"`
}

// GameEvent represents events that can occur in the game
type GameEvent struct {
	PlayerJoined         *PlayerJoined            `json:"playerJoined,omitempty"`
	PlayerLeft           *PlayerLeft              `json:"playerLeft,omitempty"`
	PhaseChanged         *PhaseChanged            `json:"phaseChanged,omitempty"`
	VoteCast             *VoteCast                `json:"voteCast,omitempty"`
	CharacterAssigned    *CharacterAssigned       `json:"characterAssigned,omitempty"`
	PlayerDied           *PlayerDiedEvent         `json:"playerDied,omitempty"`
	NominationStarted    *NominationStartedEvent  `json:"nominationStarted,omitempty"`
	NominationResolved   *NominationResolvedEvent `json:"nominationResolved,omitempty"`
	NightActionSubmitted *NightActionEvent        `json:"nightActionSubmitted,omitempty"`
	GameEnded            *GameEndedEvent          `json:"gameEnded,omitempty"`
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
	Decision *bool   `json:"decision,omitempty"`
}

type CharacterAssigned struct {
	PlayerID  string    `json:"playerId"`
	Character Character `json:"character"`
}

// PlayerDiedEvent represents a player death event
type PlayerDiedEvent struct {
	PlayerID  string     `json:"playerId"`
	Cause     DeathCause `json:"cause"`
	DayNumber int32      `json:"dayNumber"`
}

// NominationStartedEvent represents the start of a nomination
type NominationStartedEvent struct {
	NominatorID string `json:"nominatorId"`
	NomineeID   string `json:"nomineeId"`
}

// NominationResolvedEvent represents the result of a nomination vote
type NominationResolvedEvent struct {
	NomineeID     string `json:"nomineeId"`
	Executed      bool   `json:"executed"`
	YesVotes      int    `json:"yesVotes"`
	NoVotes       int    `json:"noVotes"`
	RequiredVotes int    `json:"requiredVotes"`
}

// NightActionEvent represents a night action submitted by a player
type NightActionEvent struct {
	ActorID    string          `json:"actorId"`
	ActionType NightActionType `json:"actionType"`
	TargetIDs  []string        `json:"targetIds"`
	Result     *string         `json:"result,omitempty"`
}

// GameEndedEvent represents the end of the game
type GameEndedEvent struct {
	Winner      Team      `json:"winner"`
	Reason      WinReason `json:"reason"`
	Description string    `json:"description"`
}
