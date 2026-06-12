package ws

import (
	"fmt"
	"strings"
	"sync"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

// Command types for GameSession.Apply()
type SetStorytellerCmd struct {
	SenderID       string
	TargetPlayerID string
}

type AssignCharactersCmd struct {
	SenderID    string
	Assignments map[string]string // playerID -> characterID
}

type SubmitEventCmd struct {
	SenderID string
	Event    game.GameEvent
}

type StartGameCmd struct {
	SenderID string
}

type ChangePhaseCmd struct {
	SenderID string
	Phase    game.GamePhase
}

type NominateCmd struct {
	SenderID  string
	NomineeID string
}

type CastVoteCmd struct {
	SenderID string
	Decision bool
}

type ResolveNominationCmd struct {
	SenderID string
}

type ExecutePlayerCmd struct {
	SenderID string
	PlayerID string
}

type SubmitNightActionCmd struct {
	SenderID   string
	ActionType string
	TargetIDs  []string
}

type ResolveNightCmd struct {
	SenderID string
}

type KickPlayerCmd struct {
	SenderID       string
	TargetPlayerID string
}

type UpdateRoomSettingsCmd struct {
	SenderID   string
	MaxPlayers int
	ScriptID   string
}

type EndGameCmd struct {
	SenderID    string
	Winner      game.Team
	Reason      game.WinReason
	Description string
}

type KillPlayerCmd struct {
	SenderID string
	PlayerID string
	Cause    game.DeathCause
}

// Command is a marker interface for all commands.
type Command interface {
	commandTag()
}

func (SetStorytellerCmd) commandTag()     {}
func (AssignCharactersCmd) commandTag()   {}
func (SubmitEventCmd) commandTag()        {}
func (StartGameCmd) commandTag()          {}
func (ChangePhaseCmd) commandTag()        {}
func (NominateCmd) commandTag()           {}
func (CastVoteCmd) commandTag()           {}
func (ResolveNominationCmd) commandTag()  {}
func (ExecutePlayerCmd) commandTag()      {}
func (SubmitNightActionCmd) commandTag()  {}
func (ResolveNightCmd) commandTag()       {}
func (KickPlayerCmd) commandTag()         {}
func (UpdateRoomSettingsCmd) commandTag() {}
func (EndGameCmd) commandTag()            {}
func (KillPlayerCmd) commandTag()         {}

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
	scriptID        string

	// Game state fields (populated after game starts)
	phase          game.GamePhase
	dayNumber      int32
	nightNumber    int32
	nightWakeIndex int
	nomination     *game.Nomination
	nightActions   []game.NightAction
	deaths         []game.DeathRecord
	ghostVotesUsed map[string]bool      // playerID -> whether ghost vote was used
	winner         *game.GameEndedEvent // set when game ends
}

func NewGameSession(scriptIDs ...string) *GameSession {
	scriptID := game.TroubleBrewingScriptID
	if len(scriptIDs) > 0 && scriptIDs[0] != "" {
		scriptID = scriptIDs[0]
	}

	return &GameSession{
		phase:          game.GamePhaseSetup,
		ghostVotesUsed: make(map[string]bool),
		scriptID:       scriptID,
	}
}

func (gs *GameSession) SetPlayers(players []game.Player) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.players = players
}

func (gs *GameSession) AddPlayer(player game.Player) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	for i, p := range gs.players {
		if p.ID == player.ID {
			if player.Name != "" {
				gs.players[i].Name = player.Name
			}
			if gs.players[i].Character == nil && player.Character != nil {
				gs.players[i].Character = player.Character
			}
			return
		}
	}
	gs.players = append(gs.players, player)
}

func (gs *GameSession) AddOrReconnectPlayer(player game.Player, maxPlayers int) (bool, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if player.ID == gs.storytellerID {
		return false, nil
	}

	for i, p := range gs.players {
		if p.ID == player.ID {
			if player.Name != "" {
				gs.players[i].Name = player.Name
			}
			return false, nil
		}
	}

	capacity := maxPlayers
	if capacity <= 0 {
		capacity = defaultMaxPlayers
	}
	if gs.storytellerID == "" {
		capacity++
	}
	if len(gs.players) >= capacity {
		return false, fmt.Errorf("room is full")
	}

	gs.players = append(gs.players, player)
	return true, nil
}

func (gs *GameSession) RemovePlayer(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if playerID == gs.storytellerID {
		gs.storytellerID = ""
	}
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

func (gs *GameSession) ParticipantCount() int {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	count := len(gs.players)
	if gs.storytellerID != "" {
		count++
	}
	return count
}

func (gs *GameSession) Phase() game.GamePhase {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.phase
}

func (gs *GameSession) DayNumber() int32 {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.dayNumber
}

func (gs *GameSession) Deaths() []game.DeathRecord {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	result := make([]game.DeathRecord, len(gs.deaths))
	copy(result, gs.deaths)
	return result
}

func (gs *GameSession) Nomination() *game.Nomination {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.nomination
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
	case StartGameCmd:
		return gs.applyStartGame(c)
	case ChangePhaseCmd:
		return gs.applyChangePhase(c)
	case NominateCmd:
		return gs.applyNominate(c)
	case CastVoteCmd:
		return gs.applyCastVote(c)
	case ResolveNominationCmd:
		return gs.applyResolveNomination(c)
	case ExecutePlayerCmd:
		return gs.applyExecutePlayer(c)
	case SubmitNightActionCmd:
		return gs.applySubmitNightAction(c)
	case ResolveNightCmd:
		return gs.applyResolveNight(c)
	case KickPlayerCmd:
		return gs.applyKickPlayer(c)
	case UpdateRoomSettingsCmd:
		return gs.applyUpdateRoomSettings(c)
	case EndGameCmd:
		return gs.applyEndGame(c)
	case KillPlayerCmd:
		return gs.applyKillPlayer(c)
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

	// Verify all playerIDs in assignments match actual players
	playerSet := make(map[string]bool, len(gs.players))
	for _, p := range gs.players {
		playerSet[p.ID] = true
	}
	for playerID := range cmd.Assignments {
		if !playerSet[playerID] {
			return ApplyResult{}, fmt.Errorf("player %s not found in game", playerID)
		}
	}

	playerCount := len(gs.players)
	if !game.ValidateScriptAssignment(gs.scriptID, cmd.Assignments, playerCount) {
		return ApplyResult{}, fmt.Errorf("invalid character assignment for player count")
	}

	// Assign characters
	for playerID, charID := range cmd.Assignments {
		charDef := game.GetScriptCharacterByID(gs.scriptID, charID)
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
		charDef := game.GetScriptCharacterByID(gs.scriptID, charID)
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
	return ApplyResult{}, fmt.Errorf("raw event submission is disabled; use explicit game commands")
}

func (gs *GameSession) applyKickPlayer(cmd KickPlayerCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("players can only be kicked during setup phase")
	}
	if cmd.TargetPlayerID == "" {
		return ApplyResult{}, fmt.Errorf("target player is required")
	}
	if cmd.SenderID == cmd.TargetPlayerID {
		return ApplyResult{}, fmt.Errorf("room creator cannot kick themselves")
	}

	found := false
	if gs.storytellerID == cmd.TargetPlayerID {
		gs.storytellerID = ""
		gs.originalPlayers = 0
		found = true
	}

	for i, player := range gs.players {
		if player.ID == cmd.TargetPlayerID {
			gs.players = append(gs.players[:i], gs.players[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ApplyResult{}, fmt.Errorf("player %s not found in game", cmd.TargetPlayerID)
	}

	return ApplyResult{
		Events: []game.GameEvent{
			{PlayerLeft: &game.PlayerLeft{PlayerID: cmd.TargetPlayerID}},
		},
		Updated: true,
	}, nil
}

func (gs *GameSession) applyUpdateRoomSettings(cmd UpdateRoomSettingsCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("room settings can only be updated during setup phase")
	}

	scriptID := strings.TrimSpace(cmd.ScriptID)
	if cmd.MaxPlayers == 0 && scriptID == "" {
		return ApplyResult{}, fmt.Errorf("at least one room setting is required")
	}
	if cmd.MaxPlayers != 0 {
		if cmd.MaxPlayers < 5 || cmd.MaxPlayers > 15 {
			return ApplyResult{}, fmt.Errorf("maxPlayers must be between 5 and 15")
		}
		if currentPlayers := gs.effectivePlayerCountLocked(); currentPlayers > cmd.MaxPlayers {
			return ApplyResult{}, fmt.Errorf("maxPlayers cannot be less than current player count")
		}
	}
	if scriptID != "" {
		if game.GetScriptByID(scriptID) == nil {
			return ApplyResult{}, fmt.Errorf("unsupported script")
		}
		if scriptID != gs.scriptID && gs.hasAssignedCharactersLocked() {
			return ApplyResult{}, fmt.Errorf("script cannot be changed after characters are assigned")
		}
		gs.scriptID = scriptID
	}

	return ApplyResult{Updated: true}, nil
}

// ────────────────────────────────────────────────
// Helper: find player by ID in gs.players
// ────────────────────────────────────────────────

func (gs *GameSession) findPlayerIndex(playerID string) int {
	for i, p := range gs.players {
		if p.ID == playerID {
			return i
		}
	}
	return -1
}

func (gs *GameSession) effectivePlayerCountLocked() int {
	count := len(gs.players)
	if gs.storytellerID == "" && count > 0 {
		count--
	}
	return count
}

func (gs *GameSession) hasAssignedCharactersLocked() bool {
	for _, player := range gs.players {
		if player.Character != nil {
			return true
		}
	}
	return false
}

func (gs *GameSession) startNightLocked() {
	gs.nightNumber++
	if gs.nightNumber == 0 {
		gs.nightNumber = 1
	}
	gs.nightWakeIndex = 0
	gs.nightActions = nil
}

func (gs *GameSession) activeNightWakeStepsLocked() []game.NightWakeStep {
	return game.GetActiveNightWakeSteps(gs.scriptID, gs.nightNumber, gs.players)
}

func (gs *GameSession) currentNightWakeStepLocked() *game.NightWakeStep {
	steps := gs.activeNightWakeStepsLocked()
	if gs.nightWakeIndex < 0 || gs.nightWakeIndex >= len(steps) {
		return nil
	}
	step := steps[gs.nightWakeIndex]
	return &step
}

func (gs *GameSession) validateNightTargetsLocked(step game.NightWakeStep, targetIDs []string) error {
	if len(targetIDs) < step.MinTargets {
		return fmt.Errorf("night action %s requires at least %d target(s)", step.ActionType, step.MinTargets)
	}
	if len(targetIDs) > step.MaxTargets {
		return fmt.Errorf("night action %s allows at most %d target(s)", step.ActionType, step.MaxTargets)
	}

	seen := map[string]bool{}
	for _, targetID := range targetIDs {
		if seen[targetID] {
			return fmt.Errorf("duplicate night action target %s", targetID)
		}
		seen[targetID] = true
		if gs.findPlayerIndex(targetID) == -1 {
			return fmt.Errorf("night action target %s not found", targetID)
		}
	}
	return nil
}

func (gs *GameSession) alivePlayerCountLocked() int {
	count := 0
	for _, player := range gs.players {
		if player.IsAlive {
			count++
		}
	}
	return count
}

func majorityThreshold(alivePlayerCount int) int {
	if alivePlayerCount <= 0 {
		return 0
	}
	return (alivePlayerCount + 1) / 2
}

// ────────────────────────────────────────────────
// StartGameCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyStartGame(cmd StartGameCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("game can only be started from setup phase, current phase: %d", gs.phase)
	}
	if gs.storytellerID == "" {
		return ApplyResult{}, fmt.Errorf("storyteller must be set before starting the game")
	}
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can start the game")
	}

	// Verify all players have characters assigned
	for _, p := range gs.players {
		if p.Character == nil {
			return ApplyResult{}, fmt.Errorf("player %s has no character assigned", p.ID)
		}
	}

	gs.phase = game.GamePhaseNight
	gs.dayNumber = 1
	gs.startNightLocked()

	events := []game.GameEvent{
		{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseNight}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// ChangePhaseCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyChangePhase(cmd ChangePhaseCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can change phase")
	}
	if gs.phase == game.GamePhaseSetup || gs.phase == game.GamePhaseFinished {
		return ApplyResult{}, fmt.Errorf("cannot change phase from %d", gs.phase)
	}

	// Validate transition: Night<->Day (storyteller-initiated)
	valid := false
	switch {
	case gs.phase == game.GamePhaseNight && cmd.Phase == game.GamePhaseDay:
		valid = true
	case gs.phase == game.GamePhaseDay && cmd.Phase == game.GamePhaseNight:
		valid = true
	}

	if !valid {
		return ApplyResult{}, fmt.Errorf("invalid phase transition from %d to %d", gs.phase, cmd.Phase)
	}

	if cmd.Phase == game.GamePhaseDay {
		gs.dayNumber++
	}
	if cmd.Phase == game.GamePhaseNight {
		gs.startNightLocked()
	}

	gs.phase = cmd.Phase

	events := []game.GameEvent{
		{PhaseChanged: &game.PhaseChanged{Phase: cmd.Phase}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// NominateCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyNominate(cmd NominateCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseDay {
		return ApplyResult{}, fmt.Errorf("nominations can only happen during the day phase")
	}
	if cmd.SenderID == cmd.NomineeID {
		return ApplyResult{}, fmt.Errorf("cannot nominate yourself")
	}

	nomIdx := gs.findPlayerIndex(cmd.SenderID)
	if nomIdx == -1 {
		return ApplyResult{}, fmt.Errorf("nominator %s not found", cmd.SenderID)
	}
	if !gs.players[nomIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("dead players cannot nominate")
	}

	nomineeIdx := gs.findPlayerIndex(cmd.NomineeID)
	if nomineeIdx == -1 {
		return ApplyResult{}, fmt.Errorf("nominee %s not found", cmd.NomineeID)
	}
	if !gs.players[nomineeIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("cannot nominate a dead player")
	}

	gs.phase = game.GamePhaseVoting
	gs.nomination = &game.Nomination{
		NominatorID: cmd.SenderID,
		NomineeID:   cmd.NomineeID,
		Votes:       make(map[string]bool),
	}

	events := []game.GameEvent{
		{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseVoting}},
		{NominationStarted: &game.NominationStartedEvent{
			NominatorID: cmd.SenderID,
			NomineeID:   cmd.NomineeID,
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// CastVoteCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyCastVote(cmd CastVoteCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseVoting || gs.nomination == nil {
		return ApplyResult{}, fmt.Errorf("no active nomination to vote on")
	}

	voterIdx := gs.findPlayerIndex(cmd.SenderID)
	if voterIdx == -1 {
		return ApplyResult{}, fmt.Errorf("voter %s not found", cmd.SenderID)
	}
	voter := gs.players[voterIdx]

	// Dead players may cast one ghost vote
	if !voter.IsAlive {
		if gs.ghostVotesUsed[cmd.SenderID] {
			return ApplyResult{}, fmt.Errorf("ghost vote already used")
		}
		gs.ghostVotesUsed[cmd.SenderID] = true
	}

	// Check for duplicate vote
	if _, already := gs.nomination.Votes[cmd.SenderID]; already {
		return ApplyResult{}, fmt.Errorf("player %s has already voted", cmd.SenderID)
	}

	gs.nomination.Votes[cmd.SenderID] = cmd.Decision

	events := []game.GameEvent{
		{VoteCast: &game.VoteCast{
			VoterID:  cmd.SenderID,
			Decision: &cmd.Decision,
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// ResolveNominationCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyResolveNomination(cmd ResolveNominationCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can resolve a nomination")
	}
	if gs.phase != game.GamePhaseVoting || gs.nomination == nil {
		return ApplyResult{}, fmt.Errorf("no active nomination to resolve")
	}

	yesVotes := 0
	noVotes := 0
	for _, decision := range gs.nomination.Votes {
		if decision {
			yesVotes++
		} else {
			noVotes++
		}
	}

	requiredVotes := majorityThreshold(gs.alivePlayerCountLocked())
	executed := yesVotes >= requiredVotes
	nomineeID := gs.nomination.NomineeID

	gs.nomination.Resolved = true
	gs.nomination = nil
	gs.phase = game.GamePhaseDay

	events := []game.GameEvent{
		{NominationResolved: &game.NominationResolvedEvent{
			NomineeID:     nomineeID,
			Executed:      executed,
			YesVotes:      yesVotes,
			NoVotes:       noVotes,
			RequiredVotes: requiredVotes,
		}},
	}

	// If executed, kill the player and check win conditions
	if executed {
		pIdx := gs.findPlayerIndex(nomineeID)
		if pIdx != -1 {
			gs.players[pIdx].IsAlive = false
			gs.deaths = append(gs.deaths, game.DeathRecord{
				PlayerID:  nomineeID,
				Cause:     game.DeathCauseExecution,
				DayNumber: gs.dayNumber,
			})
			events = append(events, game.GameEvent{
				PlayerDied: &game.PlayerDiedEvent{
					PlayerID:  nomineeID,
					Cause:     game.DeathCauseExecution,
					DayNumber: gs.dayNumber,
				},
			})
			if won := gs.checkWinConditions(); won != nil {
				gs.winner = won
				events = append(events, game.GameEvent{GameEnded: won})
				gs.phase = game.GamePhaseFinished
			}
		}
	}
	if gs.phase != game.GamePhaseFinished {
		if executed {
			gs.phase = game.GamePhaseNight
			gs.startNightLocked()
			events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseNight}})
		} else {
			gs.phase = game.GamePhaseDay
			events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay}})
		}
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// ExecutePlayerCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyExecutePlayer(cmd ExecutePlayerCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can execute a player")
	}
	if gs.phase == game.GamePhaseSetup || gs.phase == game.GamePhaseFinished {
		return ApplyResult{}, fmt.Errorf("cannot execute player in phase %d", gs.phase)
	}

	pIdx := gs.findPlayerIndex(cmd.PlayerID)
	if pIdx == -1 {
		return ApplyResult{}, fmt.Errorf("player %s not found", cmd.PlayerID)
	}
	if !gs.players[pIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("player %s is already dead", cmd.PlayerID)
	}

	gs.players[pIdx].IsAlive = false
	gs.deaths = append(gs.deaths, game.DeathRecord{
		PlayerID:  cmd.PlayerID,
		Cause:     game.DeathCauseExecution,
		DayNumber: gs.dayNumber,
	})

	events := []game.GameEvent{
		{PlayerDied: &game.PlayerDiedEvent{
			PlayerID:  cmd.PlayerID,
			Cause:     game.DeathCauseExecution,
			DayNumber: gs.dayNumber,
		}},
	}

	if won := gs.checkWinConditions(); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// SubmitNightActionCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applySubmitNightAction(cmd SubmitNightActionCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("night actions can only be submitted during the night phase")
	}

	if cmd.SenderID != gs.storytellerID {
		actorIdx := gs.findPlayerIndex(cmd.SenderID)
		if actorIdx == -1 {
			return ApplyResult{}, fmt.Errorf("actor %s not found", cmd.SenderID)
		}
		if !gs.players[actorIdx].IsAlive {
			return ApplyResult{}, fmt.Errorf("dead players cannot submit night actions")
		}
	} else {
		step := gs.currentNightWakeStepLocked()
		if step == nil {
			return ApplyResult{}, fmt.Errorf("no remaining night wake steps")
		}
		actionType := game.NightActionType(cmd.ActionType)
		if actionType != step.ActionType {
			return ApplyResult{}, fmt.Errorf("expected night action %s, got %s", step.ActionType, actionType)
		}
		if err := gs.validateNightTargetsLocked(*step, cmd.TargetIDs); err != nil {
			return ApplyResult{}, err
		}
	}

	action := game.NightAction{
		ActorID:    cmd.SenderID,
		ActionType: game.NightActionType(cmd.ActionType),
		TargetIDs:  cmd.TargetIDs,
	}
	gs.nightActions = append(gs.nightActions, action)
	if cmd.SenderID == gs.storytellerID {
		gs.nightWakeIndex++
	}

	events := []game.GameEvent{
		{NightActionSubmitted: &game.NightActionEvent{
			ActorID:    cmd.SenderID,
			ActionType: game.NightActionType(cmd.ActionType),
			TargetIDs:  cmd.TargetIDs,
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// ResolveNightCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyResolveNight(cmd ResolveNightCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can resolve the night")
	}
	if gs.phase != game.GamePhaseNight {
		return ApplyResult{}, fmt.Errorf("night can only be resolved during the night phase")
	}
	if remaining := len(gs.activeNightWakeStepsLocked()) - gs.nightWakeIndex; remaining > 0 {
		return ApplyResult{}, fmt.Errorf("cannot resolve night with %d wake step(s) remaining", remaining)
	}

	// Process adjudicated night actions. Player-submitted actions are treated as
	// private choices for the storyteller; only the storyteller can resolve deaths.
	var events []game.GameEvent
	for _, action := range gs.nightActions {
		if action.ActorID == gs.storytellerID && action.ActionType == game.NightActionKill {
			for _, targetID := range action.TargetIDs {
				tIdx := gs.findPlayerIndex(targetID)
				if tIdx != -1 && gs.players[tIdx].IsAlive {
					gs.players[tIdx].IsAlive = false
					gs.deaths = append(gs.deaths, game.DeathRecord{
						PlayerID:  targetID,
						Cause:     game.DeathCauseNightKill,
						DayNumber: gs.dayNumber,
						KilledBy:  action.ActorID,
					})
					events = append(events, game.GameEvent{
						PlayerDied: &game.PlayerDiedEvent{
							PlayerID:  targetID,
							Cause:     game.DeathCauseNightKill,
							DayNumber: gs.dayNumber,
						},
					})
				}
			}
		}
	}

	// Clear night actions for next night
	gs.nightActions = nil

	// Transition to Day
	gs.phase = game.GamePhaseDay
	gs.dayNumber++
	events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay}})

	if won := gs.checkWinConditions(); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// EndGameCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyEndGame(cmd EndGameCmd) (ApplyResult, error) {
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can end the game")
	}
	if gs.phase == game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("game cannot be ended before it starts")
	}
	if gs.phase == game.GamePhaseFinished {
		return ApplyResult{}, fmt.Errorf("game is already finished")
	}
	if cmd.Winner != game.TeamGood && cmd.Winner != game.TeamEvil {
		return ApplyResult{}, fmt.Errorf("winner must be good or evil")
	}

	reason := cmd.Reason
	if reason == "" {
		reason = game.WinReasonStorytellerDecision
	}
	description := strings.TrimSpace(cmd.Description)
	if description == "" {
		description = "Storyteller ended the game."
	}

	ended := &game.GameEndedEvent{
		Winner:      cmd.Winner,
		Reason:      reason,
		Description: description,
	}
	gs.winner = ended
	gs.phase = game.GamePhaseFinished

	return ApplyResult{
		Events:  []game.GameEvent{{GameEnded: ended}},
		Updated: true,
	}, nil
}

// ────────────────────────────────────────────────
// KillPlayerCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyKillPlayer(cmd KillPlayerCmd) (ApplyResult, error) {
	pIdx := gs.findPlayerIndex(cmd.PlayerID)
	if pIdx == -1 {
		return ApplyResult{}, fmt.Errorf("player %s not found", cmd.PlayerID)
	}
	if !gs.players[pIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("player %s is already dead", cmd.PlayerID)
	}

	gs.players[pIdx].IsAlive = false
	gs.deaths = append(gs.deaths, game.DeathRecord{
		PlayerID:  cmd.PlayerID,
		Cause:     cmd.Cause,
		DayNumber: gs.dayNumber,
		KilledBy:  cmd.SenderID,
	})

	events := []game.GameEvent{
		{PlayerDied: &game.PlayerDiedEvent{
			PlayerID:  cmd.PlayerID,
			Cause:     cmd.Cause,
			DayNumber: gs.dayNumber,
		}},
	}

	if won := gs.checkWinConditions(); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// Win condition checking
// ────────────────────────────────────────────────

// checkWinConditions evaluates current state and returns a GameEndedEvent if
// the game should end, or nil if it continues. Must be called with gs.mu held.
func (gs *GameSession) checkWinConditions() *game.GameEndedEvent {
	aliveGood := 0
	aliveEvil := 0
	hasAliveDemon := false
	var aliveMayor bool

	for _, p := range gs.players {
		if !p.IsAlive {
			continue
		}
		if p.Character == nil {
			continue
		}
		switch p.Character.Team {
		case game.TeamGood:
			aliveGood++
		case game.TeamEvil:
			aliveEvil++
		}
		// Check if this character is a demon by looking up its definition
		charDef := game.GetCharacterByID(p.Character.ID)
		if charDef != nil && charDef.Type == game.CharacterTypeDemon {
			hasAliveDemon = true
		}
		if p.Character.ID == "mayor" {
			aliveMayor = true
		}
	}

	// Check if any executed player was the Saint
	for _, d := range gs.deaths {
		if d.Cause == game.DeathCauseExecution {
			pIdx := gs.findPlayerIndex(d.PlayerID)
			if pIdx != -1 && gs.players[pIdx].Character != nil &&
				gs.players[pIdx].Character.ID == "saint" {
				return &game.GameEndedEvent{
					Winner:      game.TeamEvil,
					Reason:      game.WinReasonSaintExecuted,
					Description: "The Saint was executed — evil wins!",
				}
			}
		}
	}

	// Demon dead => good wins (by execution)
	if !hasAliveDemon {
		return &game.GameEndedEvent{
			Winner:      game.TeamGood,
			Reason:      game.WinReasonImpExecuted,
			Description: "The Demon is dead — good wins!",
		}
	}

	totalAlive := aliveGood + aliveEvil

	// Evil wins when the game reaches the final two living players.
	if totalAlive <= 2 && totalAlive > 0 {
		return &game.GameEndedEvent{
			Winner:      game.TeamEvil,
			Reason:      game.WinReasonEvilMajority,
			Description: "Only two players remain alive — evil wins!",
		}
	}

	// Mayor endgame: only 3 alive and no execution happened today
	if totalAlive == 3 && aliveMayor {
		// Check that no execution happened this day cycle
		executedToday := false
		for _, d := range gs.deaths {
			if d.DayNumber == gs.dayNumber && d.Cause == game.DeathCauseExecution {
				executedToday = true
				break
			}
		}
		if !executedToday {
			return &game.GameEndedEvent{
				Winner:      game.TeamGood,
				Reason:      game.WinReasonMayorEndgame,
				Description: "Only 3 players remain with no execution — Mayor wins for good!",
			}
		}
	}

	return nil
}

// StateForRoom builds a complete RoomState snapshot.
func (gs *GameSession) StateForRoom(roomID string) *RoomState {
	return gs.stateForRoom(roomID, true, "")
}

// StateForRoomForRecipient builds a recipient-specific RoomState snapshot.
// Storyteller sees all character assignments; players only see their own.
func (gs *GameSession) StateForRoomForRecipient(roomID, recipientID string) *RoomState {
	return gs.stateForRoom(roomID, false, recipientID)
}

func (gs *GameSession) stateForRoom(roomID string, forceSeeAll bool, recipientID string) *RoomState {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	canSeeAll := forceSeeAll || (gs.storytellerID != "" && recipientID == gs.storytellerID)
	players := make([]game.Player, len(gs.players))
	for i, player := range gs.players {
		players[i] = player
		if player.Character != nil {
			character := *player.Character
			players[i].Character = &character
		}
		if !canSeeAll && player.ID != recipientID {
			players[i].Character = nil
		}
	}

	ghostVotesRemaining := make([]string, 0)
	for _, player := range gs.players {
		if !player.IsAlive && !gs.ghostVotesUsed[player.ID] {
			ghostVotesRemaining = append(ghostVotesRemaining, player.ID)
		}
	}

	scriptName := ""
	if script := game.GetScriptByID(gs.scriptID); script != nil {
		scriptName = script.Name
	}
	nightWakeSteps := []game.NightWakeStep(nil)
	currentNightWakeIndex := 0
	var currentNightWakeStep *game.NightWakeStep
	if gs.phase == game.GamePhaseNight {
		nightWakeSteps = gs.activeNightWakeStepsLocked()
		currentNightWakeIndex = gs.nightWakeIndex
		currentNightWakeStep = gs.currentNightWakeStepLocked()
	}

	return &RoomState{
		RoomID:                roomID,
		Players:               players,
		ScriptID:              gs.scriptID,
		ScriptName:            scriptName,
		StorytellerID:         gs.storytellerID,
		Phase:                 gs.phase,
		DayNumber:             gs.dayNumber,
		Nomination:            cloneNomination(gs.nomination),
		Deaths:                cloneDeaths(gs.deaths),
		GhostVotesRemaining:   ghostVotesRemaining,
		NightWakeSteps:        nightWakeSteps,
		CurrentNightWakeIndex: currentNightWakeIndex,
		CurrentNightWakeStep:  currentNightWakeStep,
		Winner:                cloneWinner(gs.winner),
	}
}
