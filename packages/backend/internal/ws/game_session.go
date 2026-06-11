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
	SenderID   string
	NomineeID  string
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

type KillPlayerCmd struct {
	SenderID string
	PlayerID string
	Cause    game.DeathCause
}

// Command is a marker interface for all commands.
type Command interface {
	commandTag()
}

func (SetStorytellerCmd) commandTag()    {}
func (AssignCharactersCmd) commandTag()  {}
func (SubmitEventCmd) commandTag()       {}
func (StartGameCmd) commandTag()         {}
func (ChangePhaseCmd) commandTag()       {}
func (NominateCmd) commandTag()          {}
func (CastVoteCmd) commandTag()          {}
func (ResolveNominationCmd) commandTag() {}
func (ExecutePlayerCmd) commandTag()     {}
func (SubmitNightActionCmd) commandTag() {}
func (ResolveNightCmd) commandTag()      {}
func (KillPlayerCmd) commandTag()        {}

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

	// Game state fields (populated after game starts)
	phase          game.GamePhase
	dayNumber      int32
	nomination     *game.Nomination
	nightActions   []game.NightAction
	deaths         []game.DeathRecord
	ghostVotesUsed map[string]bool      // playerID -> whether ghost vote was used
	winner          *game.GameEndedEvent // set when game ends
}

func NewGameSession() *GameSession {
	return &GameSession{
		phase:          game.GamePhaseSetup,
		ghostVotesUsed: make(map[string]bool),
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
			gs.players[i] = player
			return
		}
	}
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

	// Verify all players have characters assigned
	for _, p := range gs.players {
		if p.Character == nil {
			return ApplyResult{}, fmt.Errorf("player %s has no character assigned", p.ID)
		}
	}

	gs.phase = game.GamePhaseNight
	gs.dayNumber = 1

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

	// Majority = more yes than no
	executed := yesVotes > noVotes
	nomineeID := gs.nomination.NomineeID

	gs.nomination.Resolved = true
	gs.nomination = nil
	gs.phase = game.GamePhaseDay

	events := []game.GameEvent{
		{NominationResolved: &game.NominationResolvedEvent{
			NomineeID: nomineeID,
			Executed:  executed,
			YesVotes:  yesVotes,
			NoVotes:   noVotes,
		}},
		{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay}},
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

	actorIdx := gs.findPlayerIndex(cmd.SenderID)
	if actorIdx == -1 {
		return ApplyResult{}, fmt.Errorf("actor %s not found", cmd.SenderID)
	}
	if !gs.players[actorIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("dead players cannot submit night actions")
	}

	action := game.NightAction{
		ActorID:    cmd.SenderID,
		ActionType: game.NightActionType(cmd.ActionType),
		TargetIDs:  cmd.TargetIDs,
	}
	gs.nightActions = append(gs.nightActions, action)

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

	// Process night actions — for MVP, the storyteller determines outcomes.
	// Actions with results stored in Result field are applied.
	var events []game.GameEvent
	for _, action := range gs.nightActions {
		if action.Result != "" && action.ActionType == game.NightActionKill {
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

	// Evil majority (evil >= good, with at least 1 good dead)
	if aliveEvil >= aliveGood && aliveGood > 0 {
		return &game.GameEndedEvent{
			Winner:      game.TeamEvil,
			Reason:      game.WinReasonEvilMajority,
			Description: "Evil outnumbers good — evil wins!",
		}
	}

	// Mayor endgame: only 3 alive and no execution happened today
	totalAlive := aliveGood + aliveEvil
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

	return &RoomState{
		RoomID:        roomID,
		Players:       players,
		StorytellerID: gs.storytellerID,
		Phase:         gs.phase,
		DayNumber:     gs.dayNumber,
		Nomination:    gs.nomination,
		Deaths:        gs.deaths,
		Winner:        gs.winner,
	}
}
