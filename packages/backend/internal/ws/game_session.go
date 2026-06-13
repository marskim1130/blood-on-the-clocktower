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
	gs.resetDailyNominationLimitsLocked()
}

func (gs *GameSession) resetDailyNominationLimitsLocked() {
	gs.nominatorsToday = make(map[string]bool)
	gs.nomineesToday = make(map[string]bool)
}

func (gs *GameSession) applyPoisonEffectLocked(targetIDs []string) {
	expiresAt := gs.dayNumber + 1
	for _, targetID := range targetIDs {
		idx := gs.findPlayerIndex(targetID)
		if idx != -1 {
			gs.players[idx].PoisonedUntil = &expiresAt
		}
	}
}

func (gs *GameSession) clearExpiredPoisonLocked() {
	for i := range gs.players {
		if gs.players[i].PoisonedUntil != nil && *gs.players[i].PoisonedUntil <= gs.dayNumber {
			gs.players[i].PoisonedUntil = nil
		}
	}
}

func (gs *GameSession) findLivingCharacterIndexLocked(characterID string) int {
	for i, p := range gs.players {
		if p.Character != nil && p.Character.ID == characterID && p.IsAlive {
			return i
		}
	}
	return -1
}

func (gs *GameSession) playerIsPoisonedLocked(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(gs.players) {
		return false
	}
	poisonedUntil := gs.players[playerIdx].PoisonedUntil
	return poisonedUntil != nil && *poisonedUntil >= gs.dayNumber
}

func (gs *GameSession) playerAbilityMalfunctioningLocked(playerIdx int) bool {
	if gs.playerIsPoisonedLocked(playerIdx) {
		return true
	}
	return playerIdx >= 0 &&
		playerIdx < len(gs.players) &&
		gs.players[playerIdx].Character != nil &&
		gs.players[playerIdx].Character.ID == "drunk"
}

func (gs *GameSession) playerCanUseVisibleCharacterLocked(playerIdx int, characterID string) bool {
	if playerIdx < 0 || playerIdx >= len(gs.players) || gs.players[playerIdx].Character == nil {
		return false
	}
	player := gs.players[playerIdx]
	if player.Character.ID == characterID {
		return true
	}
	return player.Character.ID == "drunk" &&
		player.ShownCharacter != nil &&
		player.ShownCharacter.ID == characterID
}

func (gs *GameSession) characterCanAutoResolveLocked(characterID string) (int, bool) {
	playerIdx := gs.findLivingCharacterIndexLocked(characterID)
	if playerIdx == -1 || gs.playerAbilityMalfunctioningLocked(playerIdx) {
		return -1, false
	}
	return playerIdx, true
}

func (gs *GameSession) computeDemonInfoResultLocked() string {
	return "Demon: " + gs.formatPlayersByCharacterTypeLocked(game.CharacterTypeDemon)
}

func (gs *GameSession) computeMinionInfoResultLocked() string {
	return "Minions: " + gs.formatPlayersByCharacterTypeLocked(game.CharacterTypeMinion)
}

func (gs *GameSession) formatPlayersByCharacterTypeLocked(characterType game.CharacterType) string {
	summaries := make([]string, 0)
	for i := range gs.players {
		if !gs.playerHasCharacterTypeLocked(i, characterType) {
			continue
		}
		summaries = append(summaries, fmt.Sprintf("%s (%s)", gs.players[i].Name, gs.players[i].Character.Name))
	}
	if len(summaries) == 0 {
		return "none"
	}
	return strings.Join(summaries, ", ")
}

func (gs *GameSession) computeWasherwomanResultLocked(washerwomanCharID string, targetIDs []string) string {
	return gs.computeCharacterTypeHintResultLocked(washerwomanCharID, targetIDs, game.CharacterTypeTownsfolk, false)
}

func (gs *GameSession) computeLibrarianResultLocked(librarianCharID string, targetIDs []string) string {
	return gs.computeCharacterTypeHintResultLocked(librarianCharID, targetIDs, game.CharacterTypeOutsider, true)
}

func (gs *GameSession) computeInvestigatorResultLocked(investigatorCharID string, targetIDs []string) string {
	return gs.computeCharacterTypeHintResultLocked(investigatorCharID, targetIDs, game.CharacterTypeMinion, false)
}

func (gs *GameSession) computeCharacterTypeHintResultLocked(characterID string, targetIDs []string, characterType game.CharacterType, allowNone bool) string {
	if _, ok := gs.characterCanAutoResolveLocked(characterID); !ok {
		return ""
	}
	if gs.targetsIncludeAmbiguousRegistrationLocked(targetIDs) {
		return ""
	}

	for _, targetID := range targetIDs {
		targetIdx := gs.findPlayerIndex(targetID)
		if gs.playerHasCharacterTypeLocked(targetIdx, characterType) {
			return gs.players[targetIdx].Character.Name
		}
	}
	if allowNone && !gs.characterTypeInPlayLocked(characterType) {
		return "none"
	}
	return "unknown"
}

func (gs *GameSession) computeEmpathResultLocked(empathCharID string) string {
	empathIdx, ok := gs.characterCanAutoResolveLocked(empathCharID)
	if !ok {
		return ""
	}

	leftIdx := (empathIdx - 1 + len(gs.players)) % len(gs.players)
	rightIdx := (empathIdx + 1) % len(gs.players)
	if (gs.players[leftIdx].IsAlive && gs.playerHasAmbiguousRegistrationLocked(leftIdx)) ||
		(gs.players[rightIdx].IsAlive && gs.playerHasAmbiguousRegistrationLocked(rightIdx)) {
		return ""
	}

	evilCount := 0
	if gs.players[leftIdx].IsAlive && gs.isPlayerEvilLocked(leftIdx) {
		evilCount++
	}
	if gs.players[rightIdx].IsAlive && gs.isPlayerEvilLocked(rightIdx) {
		evilCount++
	}

	return fmt.Sprintf("%d", evilCount)
}

func (gs *GameSession) computeChefResultLocked(chefCharID string) string {
	if _, ok := gs.characterCanAutoResolveLocked(chefCharID); !ok {
		return ""
	}
	if len(gs.players) < 2 {
		return "0"
	}
	if gs.anyAliveAmbiguousRegistrationLocked() {
		return ""
	}

	evilPairs := 0
	for i := range gs.players {
		nextIdx := (i + 1) % len(gs.players)
		if gs.isPlayerEvilLocked(i) && gs.isPlayerEvilLocked(nextIdx) {
			evilPairs++
		}
	}

	return fmt.Sprintf("%d", evilPairs)
}

func (gs *GameSession) computeFortuneTellerResultLocked(fortuneTellerCharID string, targetIDs []string) string {
	if _, ok := gs.characterCanAutoResolveLocked(fortuneTellerCharID); !ok {
		return ""
	}

	for _, targetID := range targetIDs {
		targetIdx := gs.findPlayerIndex(targetID)
		if gs.playerIsDemonLocked(targetIdx) {
			return "yes"
		}
	}
	if gs.fortuneTellerRedHerringID == "" {
		return ""
	}
	for _, targetID := range targetIDs {
		if targetID == gs.fortuneTellerRedHerringID {
			return "yes"
		}
	}
	if gs.targetsIncludeAmbiguousRegistrationLocked(targetIDs) {
		return ""
	}

	return "no"
}

func (gs *GameSession) computeUndertakerResultLocked(undertakerCharID string) string {
	if _, ok := gs.characterCanAutoResolveLocked(undertakerCharID); !ok {
		return ""
	}

	for i := len(gs.deaths) - 1; i >= 0; i-- {
		death := gs.deaths[i]
		if death.Cause != game.DeathCauseExecution || death.DayNumber != gs.dayNumber {
			continue
		}
		playerIdx := gs.findPlayerIndex(death.PlayerID)
		if playerIdx == -1 || gs.players[playerIdx].Character == nil {
			return "unknown"
		}
		return gs.players[playerIdx].Character.Name
	}

	return "none"
}

func (gs *GameSession) computeRavenkeeperResultLocked(ravenkeeperCharID string, targetIDs []string) string {
	ravenkeeperIdx, ok := gs.characterCanAutoResolveLocked(ravenkeeperCharID)
	if !ok {
		return ""
	}
	if !gs.ravenkeeperDiesTonightLocked(ravenkeeperIdx) {
		return "none"
	}
	if len(targetIDs) == 0 {
		return ""
	}

	targetIdx := gs.findPlayerIndex(targetIDs[0])
	if targetIdx == -1 || gs.players[targetIdx].Character == nil {
		return "unknown"
	}
	return gs.players[targetIdx].Character.Name
}

func (gs *GameSession) applyButlerMasterSelectionLocked(butlerCharID string, targetIDs []string) {
	if butlerCharID != "butler" || len(targetIDs) == 0 {
		return
	}
	butlerIdx := gs.findLivingCharacterIndexLocked(butlerCharID)
	if butlerIdx == -1 {
		return
	}
	if gs.butlerMasters == nil {
		gs.butlerMasters = make(map[string]string)
	}
	gs.butlerMasters[gs.players[butlerIdx].ID] = targetIDs[0]
}

func (gs *GameSession) ravenkeeperDiesTonightLocked(ravenkeeperIdx int) bool {
	protectedTargets := gs.nightProtectedTargetsLocked()
	ravenkeeper := gs.players[ravenkeeperIdx]
	for _, action := range gs.nightActions {
		if action.ActorID != gs.storytellerID || action.ActionType != game.NightActionKill {
			continue
		}
		for _, targetID := range action.TargetIDs {
			if targetID == ravenkeeper.ID && !gs.nightKillPreventedLocked(ravenkeeperIdx, protectedTargets) {
				return true
			}
		}
	}
	return false
}

func (gs *GameSession) isPlayerEvilLocked(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(gs.players) {
		return false
	}
	player := gs.players[playerIdx]
	if player.Character == nil {
		return false
	}
	charDef := game.GetCharacterByID(player.Character.ID)
	return charDef != nil && charDef.Team == game.TeamEvil
}

func (gs *GameSession) targetsIncludeAmbiguousRegistrationLocked(targetIDs []string) bool {
	for _, targetID := range targetIDs {
		if gs.playerHasAmbiguousRegistrationLocked(gs.findPlayerIndex(targetID)) {
			return true
		}
	}
	return false
}

func (gs *GameSession) anyAliveAmbiguousRegistrationLocked() bool {
	for i := range gs.players {
		if gs.players[i].IsAlive && gs.playerHasAmbiguousRegistrationLocked(i) {
			return true
		}
	}
	return false
}

func (gs *GameSession) playerHasAmbiguousRegistrationLocked(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(gs.players) || gs.players[playerIdx].Character == nil {
		return false
	}
	if gs.playerIsPoisonedLocked(playerIdx) {
		return false
	}
	switch gs.players[playerIdx].Character.ID {
	case "recluse", "spy":
		return true
	default:
		return false
	}
}

func (gs *GameSession) characterTypeInPlayLocked(characterType game.CharacterType) bool {
	for i := range gs.players {
		if gs.playerHasCharacterTypeLocked(i, characterType) {
			return true
		}
	}
	return false
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
	assignments := make(map[string]string, len(gs.players))
	for _, p := range gs.players {
		if p.Character == nil {
			return ApplyResult{}, fmt.Errorf("player %s has no character assigned", p.ID)
		}
		assignments[p.ID] = p.Character.ID
	}
	if !game.ValidateScriptAssignment(gs.scriptID, assignments, len(gs.players)) {
		return ApplyResult{}, fmt.Errorf("invalid character assignment for player count")
	}
	shownCharacters := make(map[string]string, len(gs.players))
	for _, p := range gs.players {
		if p.ShownCharacter != nil {
			shownCharacters[p.ID] = p.ShownCharacter.ID
		}
	}
	if err := validateShownCharacters(gs.scriptID, assignments, shownCharacters); err != nil {
		return ApplyResult{}, err
	}
	if err := validateFortuneTellerRedHerring(gs.scriptID, assignments, gs.fortuneTellerRedHerringID); err != nil {
		return ApplyResult{}, err
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
		gs.resetDailyNominationLimitsLocked()
	}
	if cmd.Phase == game.GamePhaseNight {
		if won := gs.mayorEndgameWinnerLocked(); won != nil {
			gs.winner = won
			gs.phase = game.GamePhaseFinished
			return ApplyResult{Events: []game.GameEvent{{GameEnded: won}}, Updated: true}, nil
		}
		// Clear expired poison at dusk (Day→Night transition)
		gs.clearExpiredPoisonLocked()
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
	if gs.nominatorsToday == nil {
		gs.nominatorsToday = make(map[string]bool)
	}
	if gs.nomineesToday == nil {
		gs.nomineesToday = make(map[string]bool)
	}
	if gs.nominatorsToday[cmd.SenderID] {
		return ApplyResult{}, fmt.Errorf("player %s has already nominated today", cmd.SenderID)
	}
	if gs.nomineesToday[cmd.NomineeID] {
		return ApplyResult{}, fmt.Errorf("player %s has already been nominated today", cmd.NomineeID)
	}

	gs.nominatorsToday[cmd.SenderID] = true
	gs.nomineesToday[cmd.NomineeID] = true

	if gs.virginAbilityUsed == nil {
		gs.virginAbilityUsed = make(map[string]bool)
	}
	if gs.players[nomineeIdx].Character != nil &&
		gs.players[nomineeIdx].Character.ID == "virgin" &&
		!gs.virginAbilityUsed[cmd.NomineeID] {
		gs.virginAbilityUsed[cmd.NomineeID] = true
		if !gs.playerAbilityMalfunctioningLocked(nomineeIdx) &&
			gs.playerHasCharacterTypeLocked(nomIdx, game.CharacterTypeTownsfolk) {
			return gs.applyVirginExecutionLocked(cmd, nomIdx)
		}
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

func (gs *GameSession) playerHasCharacterTypeLocked(playerIndex int, characterType game.CharacterType) bool {
	if playerIndex < 0 || playerIndex >= len(gs.players) || gs.players[playerIndex].Character == nil {
		return false
	}
	charDef := game.GetCharacterByID(gs.players[playerIndex].Character.ID)
	return charDef != nil && charDef.Type == characterType
}

func (gs *GameSession) applyVirginExecutionLocked(cmd NominateCmd, nominatorIndex int) (ApplyResult, error) {
	gs.players[nominatorIndex].IsAlive = false
	gs.deaths = append(gs.deaths, game.DeathRecord{
		PlayerID:  cmd.SenderID,
		Cause:     game.DeathCauseExecution,
		DayNumber: gs.dayNumber,
	})

	events := []game.GameEvent{
		{NominationStarted: &game.NominationStartedEvent{
			NominatorID: cmd.SenderID,
			NomineeID:   cmd.NomineeID,
		}},
		{PlayerDied: &game.PlayerDiedEvent{
			PlayerID:  cmd.SenderID,
			Cause:     game.DeathCauseExecution,
			DayNumber: gs.dayNumber,
		}},
	}

	if won := gs.checkWinConditions(0); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
		return ApplyResult{Events: events, Updated: true}, nil
	}

	gs.phase = game.GamePhaseNight
	gs.nomination = nil
	gs.startNightLocked()
	events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseNight}})

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
	if err := gs.validateButlerVoteLocked(voterIdx, cmd.Decision); err != nil {
		return ApplyResult{}, err
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

func (gs *GameSession) validateButlerVoteLocked(voterIdx int, decision bool) error {
	if !decision || voterIdx < 0 || voterIdx >= len(gs.players) {
		return nil
	}
	voter := gs.players[voterIdx]
	if !voter.IsAlive || voter.Character == nil || voter.Character.ID != "butler" {
		return nil
	}
	if gs.playerAbilityMalfunctioningLocked(voterIdx) {
		return nil
	}
	masterID := gs.butlerMasters[voter.ID]
	if masterID == "" {
		return fmt.Errorf("butler %s has not chosen a master", voter.ID)
	}
	if !gs.nomination.Votes[masterID] {
		return fmt.Errorf("butler %s cannot vote until master %s votes yes", voter.ID, masterID)
	}
	return nil
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
			aliveBeforeDeath := gs.alivePlayerCountLocked()
			demonDeathAliveCount := gs.demonDeathAliveCountLocked(pIdx, aliveBeforeDeath)
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
			if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
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

	aliveBeforeDeath := gs.alivePlayerCountLocked()
	demonDeathAliveCount := gs.demonDeathAliveCountLocked(pIdx, aliveBeforeDeath)
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

	if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

// ────────────────────────────────────────────────
// UseSlayerAbilityCmd
// ────────────────────────────────────────────────

func (gs *GameSession) applyUseSlayerAbility(cmd UseSlayerAbilityCmd) (ApplyResult, error) {
	if gs.phase != game.GamePhaseDay {
		return ApplyResult{}, fmt.Errorf("Slayer ability can only be used during the day phase")
	}
	if cmd.TargetPlayerID == "" {
		return ApplyResult{}, fmt.Errorf("target player is required")
	}

	slayerIdx := gs.findPlayerIndex(cmd.SenderID)
	if slayerIdx == -1 {
		return ApplyResult{}, fmt.Errorf("Slayer %s not found", cmd.SenderID)
	}
	slayer := gs.players[slayerIdx]
	if !slayer.IsAlive {
		return ApplyResult{}, fmt.Errorf("dead players cannot use the Slayer ability")
	}
	if !gs.playerCanUseVisibleCharacterLocked(slayerIdx, "slayer") {
		return ApplyResult{}, fmt.Errorf("only the Slayer can use this ability")
	}
	if gs.slayerUsed == nil {
		gs.slayerUsed = make(map[string]bool)
	}
	if gs.slayerUsed[cmd.SenderID] {
		return ApplyResult{}, fmt.Errorf("Slayer ability already used")
	}

	targetIdx := gs.findPlayerIndex(cmd.TargetPlayerID)
	if targetIdx == -1 {
		return ApplyResult{}, fmt.Errorf("target player %s not found", cmd.TargetPlayerID)
	}
	if !gs.players[targetIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("cannot target a dead player")
	}

	gs.slayerUsed[cmd.SenderID] = true
	events := []game.GameEvent{}
	var targetDef *game.CharacterDefinition
	if gs.players[targetIdx].Character != nil {
		targetDef = game.GetCharacterByID(gs.players[targetIdx].Character.ID)
	}
	if targetDef != nil &&
		targetDef.Type == game.CharacterTypeDemon &&
		slayer.Character != nil &&
		slayer.Character.ID == "slayer" &&
		!gs.playerAbilityMalfunctioningLocked(slayerIdx) {
		aliveBeforeDeath := gs.alivePlayerCountLocked()
		demonDeathAliveCount := gs.demonDeathAliveCountLocked(targetIdx, aliveBeforeDeath)
		gs.players[targetIdx].IsAlive = false
		gs.deaths = append(gs.deaths, game.DeathRecord{
			PlayerID:  cmd.TargetPlayerID,
			Cause:     game.DeathCauseAbility,
			DayNumber: gs.dayNumber,
			KilledBy:  cmd.SenderID,
		})
		events = append(events, game.GameEvent{
			PlayerDied: &game.PlayerDiedEvent{
				PlayerID:  cmd.TargetPlayerID,
				Cause:     game.DeathCauseAbility,
				DayNumber: gs.dayNumber,
			},
		})
		if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
			gs.winner = won
			events = append(events, game.GameEvent{GameEnded: won})
			gs.phase = game.GamePhaseFinished
		}
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

	var step *game.NightWakeStep
	if cmd.SenderID != gs.storytellerID {
		actorIdx := gs.findPlayerIndex(cmd.SenderID)
		if actorIdx == -1 {
			return ApplyResult{}, fmt.Errorf("actor %s not found", cmd.SenderID)
		}
		if !gs.players[actorIdx].IsAlive {
			return ApplyResult{}, fmt.Errorf("dead players cannot submit night actions")
		}
	} else {
		step = gs.currentNightWakeStepLocked()
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

	actorCharID := ""
	if cmd.SenderID == gs.storytellerID {
		if step != nil {
			actorCharID = step.CharacterID
		}
	}

	action := game.NightAction{
		ActorID:    cmd.SenderID,
		ActionType: game.NightActionType(cmd.ActionType),
		TargetIDs:  cmd.TargetIDs,
		Result:     strings.TrimSpace(cmd.Result),
	}

	// Auto-compute ability results for information roles (if not poisoned)
	if cmd.SenderID == gs.storytellerID && action.Result == "" {
		switch action.ActionType {
		case game.NightActionLearnDemon:
			if step != nil && step.CharacterType == game.NightWakeCharacterTypeMinion {
				action.Result = gs.computeDemonInfoResultLocked()
			}
		case game.NightActionLearnTownsfolk:
			action.Result = gs.computeWasherwomanResultLocked(actorCharID, action.TargetIDs)
		case game.NightActionLearnOutsider:
			action.Result = gs.computeLibrarianResultLocked(actorCharID, action.TargetIDs)
		case game.NightActionLearnMinion:
			if step != nil && step.CharacterType == game.NightWakeCharacterTypeDemon {
				action.Result = gs.computeMinionInfoResultLocked()
			} else {
				action.Result = gs.computeInvestigatorResultLocked(actorCharID, action.TargetIDs)
			}
		case game.NightActionLearnEvilPairs:
			action.Result = gs.computeChefResultLocked(actorCharID)
		case game.NightActionLearnEvilNeighbors:
			action.Result = gs.computeEmpathResultLocked(actorCharID)
		case game.NightActionCheckDemon:
			action.Result = gs.computeFortuneTellerResultLocked(actorCharID, action.TargetIDs)
		case game.NightActionLearnExecuted:
			action.Result = gs.computeUndertakerResultLocked(actorCharID)
		case game.NightActionLearnDied:
			action.Result = gs.computeRavenkeeperResultLocked(actorCharID, action.TargetIDs)
		}
	}

	gs.nightActions = append(gs.nightActions, action)
	if cmd.SenderID == gs.storytellerID {
		gs.nightWakeIndex++
		// Apply poison effect immediately when storyteller submits poison action
		if action.ActionType == game.NightActionPoison {
			gs.applyPoisonEffectLocked(action.TargetIDs)
		}
		if action.ActionType == game.NightActionLearnMaster {
			gs.applyButlerMasterSelectionLocked(actorCharID, action.TargetIDs)
		}
	}

	events := []game.GameEvent{
		{NightActionSubmitted: &game.NightActionEvent{
			ActorID:    cmd.SenderID,
			ActionType: game.NightActionType(cmd.ActionType),
			TargetIDs:  cmd.TargetIDs,
			Result:     optionalString(action.Result),
		}},
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
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
	demonDeathAliveCount := 0
	protectedTargets := gs.nightProtectedTargetsLocked()
	for _, action := range gs.nightActions {
		if action.ActorID == gs.storytellerID && action.ActionType == game.NightActionKill {
			for _, targetID := range action.TargetIDs {
				tIdx := gs.findPlayerIndex(targetID)
				if tIdx != -1 && gs.players[tIdx].IsAlive {
					if gs.nightKillPreventedLocked(tIdx, protectedTargets) {
						continue
					}
					aliveBeforeDeath := gs.alivePlayerCountLocked()
					impSelfKill := gs.playerIsImpLocked(tIdx)
					if aliveAtDemonDeath := gs.demonDeathAliveCountLocked(tIdx, aliveBeforeDeath); aliveAtDemonDeath > demonDeathAliveCount {
						demonDeathAliveCount = aliveAtDemonDeath
					}
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
					if impSelfKill && gs.applyImpSelfStarpassLocked(tIdx) {
						demonDeathAliveCount = 0
					}
				}
			}
		}
	}

	// Clear night actions for next night
	gs.nightActions = nil

	// Transition to Day
	gs.phase = game.GamePhaseDay
	gs.dayNumber++
	gs.resetDailyNominationLimitsLocked()
	events = append(events, game.GameEvent{PhaseChanged: &game.PhaseChanged{Phase: game.GamePhaseDay}})

	if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func (gs *GameSession) nightProtectedTargetsLocked() map[string]bool {
	protected := map[string]bool{}
	monkIdx := gs.findLivingCharacterIndexLocked("monk")
	if monkIdx == -1 || gs.playerAbilityMalfunctioningLocked(monkIdx) {
		return protected
	}
	for _, action := range gs.nightActions {
		if action.ActorID != gs.storytellerID || action.ActionType != game.NightActionProtect {
			continue
		}
		for _, targetID := range action.TargetIDs {
			protected[targetID] = true
		}
	}
	return protected
}

func (gs *GameSession) nightKillPreventedLocked(targetIndex int, protectedTargets map[string]bool) bool {
	target := gs.players[targetIndex]
	if protectedTargets[target.ID] {
		return true
	}
	return target.Character != nil &&
		target.Character.ID == "soldier" &&
		!gs.playerAbilityMalfunctioningLocked(targetIndex)
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
	if cmd.SenderID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only the storyteller can kill players")
	}
	if gs.phase == game.GamePhaseSetup {
		return ApplyResult{}, fmt.Errorf("game cannot kill players before it starts")
	}
	if gs.phase == game.GamePhaseFinished {
		return ApplyResult{}, fmt.Errorf("game is already finished")
	}
	if cmd.PlayerID == "" {
		return ApplyResult{}, fmt.Errorf("target player is required")
	}
	if cmd.Cause == "" {
		return ApplyResult{}, fmt.Errorf("death cause is required")
	}
	if !isSupportedDeathCause(cmd.Cause) {
		return ApplyResult{}, fmt.Errorf("unsupported death cause %s", cmd.Cause)
	}

	pIdx := gs.findPlayerIndex(cmd.PlayerID)
	if pIdx == -1 {
		return ApplyResult{}, fmt.Errorf("player %s not found", cmd.PlayerID)
	}
	if !gs.players[pIdx].IsAlive {
		return ApplyResult{}, fmt.Errorf("player %s is already dead", cmd.PlayerID)
	}

	aliveBeforeDeath := gs.alivePlayerCountLocked()
	demonDeathAliveCount := gs.demonDeathAliveCountLocked(pIdx, aliveBeforeDeath)
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

	if won := gs.checkWinConditions(demonDeathAliveCount); won != nil {
		gs.winner = won
		events = append(events, game.GameEvent{GameEnded: won})
		gs.phase = game.GamePhaseFinished
	}

	return ApplyResult{Events: events, Updated: true}, nil
}

func isSupportedDeathCause(cause game.DeathCause) bool {
	switch cause {
	case game.DeathCauseExecution, game.DeathCauseNightKill, game.DeathCauseAbility:
		return true
	default:
		return false
	}
}

// ────────────────────────────────────────────────
// Win condition checking
// ────────────────────────────────────────────────

// checkWinConditions evaluates current state and returns a GameEndedEvent if
// the game should end, or nil if it continues. Must be called with gs.mu held.
func (gs *GameSession) checkWinConditions(demonDeathAliveCount int) *game.GameEndedEvent {
	aliveGood := 0
	aliveEvil := 0
	hasAliveDemon := false

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
	}

	// Check if any executed player was the Saint
	for _, d := range gs.deaths {
		if d.Cause == game.DeathCauseExecution && d.DayNumber == gs.dayNumber {
			pIdx := gs.findPlayerIndex(d.PlayerID)
			if pIdx != -1 && gs.players[pIdx].Character != nil &&
				gs.players[pIdx].Character.ID == "saint" &&
				!gs.playerAbilityMalfunctioningLocked(pIdx) {
				return &game.GameEndedEvent{
					Winner:      game.TeamEvil,
					Reason:      game.WinReasonSaintExecuted,
					Description: "The Saint was executed — evil wins!",
				}
			}
		}
	}

	totalAlive := aliveGood + aliveEvil

	// Demon dead => good wins (by execution)
	if !hasAliveDemon {
		if gs.applyScarletWomanStarpassLocked(demonDeathAliveCount) {
			return nil
		}
		return &game.GameEndedEvent{
			Winner:      game.TeamGood,
			Reason:      game.WinReasonImpExecuted,
			Description: "The Demon is dead — good wins!",
		}
	}

	// Evil wins when the game reaches the final two living players.
	if totalAlive <= 2 && totalAlive > 0 {
		return &game.GameEndedEvent{
			Winner:      game.TeamEvil,
			Reason:      game.WinReasonEvilMajority,
			Description: "Only two players remain alive — evil wins!",
		}
	}

	return nil
}

func (gs *GameSession) mayorEndgameWinnerLocked() *game.GameEndedEvent {
	totalAlive := 0
	mayorIdx := -1
	for i, player := range gs.players {
		if !player.IsAlive {
			continue
		}
		totalAlive++
		if player.Character != nil && player.Character.ID == "mayor" {
			mayorIdx = i
		}
	}
	if totalAlive != 3 || mayorIdx == -1 || gs.playerAbilityMalfunctioningLocked(mayorIdx) {
		return nil
	}
	for _, death := range gs.deaths {
		if death.DayNumber == gs.dayNumber && death.Cause == game.DeathCauseExecution {
			return nil
		}
	}
	return &game.GameEndedEvent{
		Winner:      game.TeamGood,
		Reason:      game.WinReasonMayorEndgame,
		Description: "Only 3 players remain with no execution — Mayor wins for good!",
	}
}

func (gs *GameSession) demonDeathAliveCountLocked(playerIndex int, aliveBeforeDeath int) int {
	if !gs.playerIsDemonLocked(playerIndex) {
		return 0
	}
	return aliveBeforeDeath
}

func (gs *GameSession) playerIsDemonLocked(playerIndex int) bool {
	if playerIndex < 0 || playerIndex >= len(gs.players) || gs.players[playerIndex].Character == nil {
		return false
	}
	charDef := game.GetCharacterByID(gs.players[playerIndex].Character.ID)
	return charDef != nil && charDef.Type == game.CharacterTypeDemon
}

func (gs *GameSession) playerIsImpLocked(playerIndex int) bool {
	return playerIndex >= 0 &&
		playerIndex < len(gs.players) &&
		gs.players[playerIndex].Character != nil &&
		gs.players[playerIndex].Character.ID == "imp"
}

func (gs *GameSession) applyImpSelfStarpassLocked(deadImpIndex int) bool {
	imp := game.GetCharacterByID("imp")
	if imp == nil {
		return false
	}
	for i := range gs.players {
		if i == deadImpIndex || !gs.players[i].IsAlive || !gs.playerHasCharacterTypeLocked(i, game.CharacterTypeMinion) {
			continue
		}
		gs.players[i].Character = &game.Character{
			ID:      imp.ID,
			Name:    imp.Name,
			Team:    imp.Team,
			Ability: imp.Ability,
		}
		return true
	}
	return false
}

func (gs *GameSession) applyScarletWomanStarpassLocked(demonDeathAliveCount int) bool {
	if demonDeathAliveCount < 5 || !gs.hasDeadDemonLocked() {
		return false
	}

	imp := game.GetCharacterByID("imp")
	if imp == nil {
		return false
	}

	for i := range gs.players {
		player := &gs.players[i]
		if !player.IsAlive ||
			player.Character == nil ||
			player.Character.ID != "scarletwoman" ||
			gs.playerAbilityMalfunctioningLocked(i) {
			continue
		}
		player.Character = &game.Character{
			ID:      imp.ID,
			Name:    imp.Name,
			Team:    imp.Team,
			Ability: imp.Ability,
		}
		return true
	}
	return false
}

func (gs *GameSession) hasDeadDemonLocked() bool {
	for _, player := range gs.players {
		if player.IsAlive || player.Character == nil {
			continue
		}
		charDef := game.GetCharacterByID(player.Character.ID)
		if charDef != nil && charDef.Type == game.CharacterTypeDemon {
			return true
		}
	}
	return false
}

// StateForRoom builds a complete RoomState snapshot.
func (gs *GameSession) StateForRoom(roomID string) *RoomState {
	return gs.stateForRoom(roomID, true, "")
}

// StateForRoomForRecipient builds a recipient-specific RoomState snapshot.
// Storyteller sees all character assignments; players usually only see their own.
func (gs *GameSession) StateForRoomForRecipient(roomID, recipientID string) *RoomState {
	return gs.stateForRoom(roomID, false, recipientID)
}

func (gs *GameSession) stateForRoom(roomID string, forceSeeAll bool, recipientID string) *RoomState {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	canSeeAll := gs.recipientCanSeeAllLocked(forceSeeAll, recipientID)
	players := make([]game.Player, len(gs.players))
	for i, player := range gs.players {
		players[i] = player
		if player.Character != nil {
			character := *player.Character
			players[i].Character = &character
		}
		if player.ShownCharacter != nil {
			shownCharacter := *player.ShownCharacter
			players[i].ShownCharacter = &shownCharacter
		}
		if player.PoisonedUntil != nil {
			poisonedUntil := *player.PoisonedUntil
			players[i].PoisonedUntil = &poisonedUntil
		}
		// Hide character from non-storyteller recipients, and show the Drunk only their false Townsfolk.
		if !canSeeAll {
			if player.ID == recipientID && player.Character != nil && player.Character.ID == "drunk" && player.ShownCharacter != nil {
				shownCharacter := *player.ShownCharacter
				players[i].Character = &shownCharacter
			} else if player.ID != recipientID {
				players[i].Character = nil
			}
			players[i].ShownCharacter = nil
		}
		// Hide PoisonedUntil from all non-storyteller recipients (including the poisoned player)
		if !canSeeAll {
			players[i].PoisonedUntil = nil
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

func (gs *GameSession) recipientCanSeeAllLocked(forceSeeAll bool, recipientID string) bool {
	if forceSeeAll ||
		(gs.storytellerID != "" && recipientID == gs.storytellerID) ||
		gs.phase == game.GamePhaseFinished ||
		gs.winner != nil {
		return true
	}
	if gs.phase != game.GamePhaseNight {
		return false
	}
	playerIdx := gs.findPlayerIndex(recipientID)
	if playerIdx == -1 || !gs.players[playerIdx].IsAlive || gs.players[playerIdx].Character == nil {
		return false
	}
	return gs.players[playerIdx].Character.ID == "spy" && !gs.playerIsPoisonedLocked(playerIdx)
}
