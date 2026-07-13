package gameplay

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
	SenderID                  string
	Assignments               map[string]string // playerID -> characterID
	ShownCharacters           map[string]string // playerID -> townsfolk characterID shown to the Drunk
	FortuneTellerRedHerringID string
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

type UseSlayerAbilityCmd struct {
	SenderID       string
	TargetPlayerID string
}

type SubmitNightActionCmd struct {
	SenderID   string
	ActionType string
	TargetIDs  []string
	Result     string
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
func (UseSlayerAbilityCmd) commandTag()   {}
func (SubmitNightActionCmd) commandTag()  {}
func (ResolveNightCmd) commandTag()       {}
func (KickPlayerCmd) commandTag()         {}
func (UpdateRoomSettingsCmd) commandTag() {}
func (EndGameCmd) commandTag()            {}
func (KillPlayerCmd) commandTag()         {}

// Result of applying a command.
type ApplyResult struct {
	Events  []game.GameEvent
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
	phase                     game.GamePhase
	dayNumber                 int32
	nightNumber               int32
	nightWakeIndex            int
	nomination                *game.Nomination
	nightActions              []game.NightAction
	deaths                    []game.DeathRecord
	ghostVotesUsed            map[string]bool   // playerID -> whether ghost vote was used
	slayerUsed                map[string]bool   // playerID -> whether Slayer ability was used
	nominatorsToday           map[string]bool   // playerID -> whether they nominated today
	nomineesToday             map[string]bool   // playerID -> whether they were nominated today
	virginAbilityUsed         map[string]bool   // playerID -> whether Virgin ability was checked
	butlerMasters             map[string]string // butler playerID -> selected master playerID
	fortuneTellerRedHerringID string
	winner                    *game.GameEndedEvent // set when game ends
}

func NewGameSession(scriptIDs ...string) *GameSession {
	scriptID := game.TroubleBrewingScriptID
	if len(scriptIDs) > 0 && scriptIDs[0] != "" {
		scriptID = scriptIDs[0]
	}

	return &GameSession{
		phase:             game.GamePhaseSetup,
		ghostVotesUsed:    make(map[string]bool),
		slayerUsed:        make(map[string]bool),
		nominatorsToday:   make(map[string]bool),
		nomineesToday:     make(map[string]bool),
		virginAbilityUsed: make(map[string]bool),
		butlerMasters:     make(map[string]string),
		scriptID:          scriptID,
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
	for i, player := range gs.players {
		result[i] = player
		if player.Character != nil {
			character := *player.Character
			result[i].Character = &character
		}
		if player.ShownCharacter != nil {
			shownCharacter := *player.ShownCharacter
			result[i].ShownCharacter = &shownCharacter
		}
		if player.PoisonedUntil != nil {
			poisonedUntil := *player.PoisonedUntil
			result[i].PoisonedUntil = &poisonedUntil
		}
	}
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
	case UseSlayerAbilityCmd:
		return gs.applyUseSlayerAbility(c)
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
	if err := validateShownCharacters(gs.scriptID, cmd.Assignments, cmd.ShownCharacters); err != nil {
		return ApplyResult{}, err
	}
	redHerringID := strings.TrimSpace(cmd.FortuneTellerRedHerringID)
	if err := validateFortuneTellerRedHerring(gs.scriptID, cmd.Assignments, redHerringID); err != nil {
		return ApplyResult{}, err
	}
	gs.fortuneTellerRedHerringID = redHerringID

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
				gs.players[i].ShownCharacter = nil
				if shownCharID := cmd.ShownCharacters[playerID]; shownCharID != "" {
					shownDef := game.GetScriptCharacterByID(gs.scriptID, shownCharID)
					gs.players[i].ShownCharacter = &game.Character{
						ID:      shownDef.ID,
						Name:    shownDef.Name,
						Team:    shownDef.Team,
						Ability: shownDef.Ability,
					}
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
		assignment := &game.CharacterAssigned{
			PlayerID: playerID,
			Character: game.Character{
				ID:      charDef.ID,
				Name:    charDef.Name,
				Team:    charDef.Team,
				Ability: charDef.Ability,
			},
		}
		if shownCharID := cmd.ShownCharacters[playerID]; shownCharID != "" {
			shownDef := game.GetScriptCharacterByID(gs.scriptID, shownCharID)
			assignment.ShownCharacter = &game.Character{
				ID:      shownDef.ID,
				Name:    shownDef.Name,
				Team:    shownDef.Team,
				Ability: shownDef.Ability,
			}
		}
		events = append(events, game.GameEvent{
			CharacterAssigned: assignment,
		})
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func validateShownCharacters(scriptID string, assignments map[string]string, shownCharacters map[string]string) error {
	if shownCharacters == nil {
		shownCharacters = map[string]string{}
	}

	for playerID, shownCharID := range shownCharacters {
		actualCharID, ok := assignments[playerID]
		if !ok {
			return fmt.Errorf("shown character player %s not found in assignments", playerID)
		}
		if actualCharID != "drunk" {
			return fmt.Errorf("shown characters can only be assigned to the Drunk")
		}
		shownDef := game.GetScriptCharacterByID(scriptID, shownCharID)
		if shownDef == nil {
			return fmt.Errorf("shown character %s not found in script", shownCharID)
		}
		if shownDef.Type != game.CharacterTypeTownsfolk {
			return fmt.Errorf("Drunk shown character must be a Townsfolk")
		}
		for _, actualAssignedCharID := range assignments {
			if actualAssignedCharID == shownCharID {
				return fmt.Errorf("Drunk shown character %s is already assigned", shownCharID)
			}
		}
	}

	for playerID, actualCharID := range assignments {
		if actualCharID == "drunk" && shownCharacters[playerID] == "" {
			return fmt.Errorf("Drunk player %s requires a shown Townsfolk character", playerID)
		}
	}

	return nil
}

func validateFortuneTellerRedHerring(scriptID string, assignments map[string]string, redHerringID string) error {
	hasFortuneTeller := false
	for _, characterID := range assignments {
		if characterID == "fortuneteller" {
			hasFortuneTeller = true
			break
		}
	}
	if redHerringID == "" {
		return nil
	}
	if !hasFortuneTeller {
		return fmt.Errorf("Fortune Teller red herring requires Fortune Teller in play")
	}

	characterID, ok := assignments[redHerringID]
	if !ok {
		return fmt.Errorf("Fortune Teller red herring player %s not found in assignments", redHerringID)
	}
	if characterID == "fortuneteller" {
		return fmt.Errorf("Fortune Teller cannot be their own red herring")
	}
	charDef := game.GetScriptCharacterByID(scriptID, characterID)
	if charDef == nil {
		return fmt.Errorf("Fortune Teller red herring character %s not found in script", characterID)
	}
	if charDef.Team != game.TeamGood {
		return fmt.Errorf("Fortune Teller red herring must be a good player")
	}
	return nil
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
