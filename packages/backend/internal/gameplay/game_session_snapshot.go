package gameplay

import (
	"encoding/json"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

const defaultMaxPlayers = 10

type gameSessionSnapshot struct {
	Players                   []game.Player           `json:"players"`
	StorytellerID             string                  `json:"storytellerId"`
	StorytellerName           string                  `json:"storytellerName"`
	OriginalPlayers           int                     `json:"originalPlayers"`
	ScriptID                  string                  `json:"scriptId"`
	Phase                     game.GamePhase          `json:"phase"`
	DayNumber                 int32                   `json:"dayNumber"`
	NightNumber               int32                   `json:"nightNumber"`
	NightWakeIndex            int                     `json:"nightWakeIndex"`
	Nomination                *game.Nomination        `json:"nomination,omitempty"`
	NominationResults         []game.NominationResult `json:"nominationResults,omitempty"`
	NightActions              []game.NightAction      `json:"nightActions,omitempty"`
	PendingNightAction        *game.NightAction       `json:"pendingNightAction,omitempty"`
	ConfirmedNightAction      *game.NightAction       `json:"confirmedNightAction,omitempty"`
	NightAcknowledged         map[string]bool         `json:"nightAcknowledged,omitempty"`
	DawnReviewPending         bool                    `json:"dawnReviewPending,omitempty"`
	PendingDawnDeathIDs       []string                `json:"pendingDawnDeathIds,omitempty"`
	Deaths                    []game.DeathRecord      `json:"deaths,omitempty"`
	GhostVotesUsed            map[string]bool         `json:"ghostVotesUsed,omitempty"`
	SlayerUsed                map[string]bool         `json:"slayerUsed,omitempty"`
	NominatorsToday           map[string]bool         `json:"nominatorsToday,omitempty"`
	NomineesToday             map[string]bool         `json:"nomineesToday,omitempty"`
	VirginAbilityUsed         map[string]bool         `json:"virginAbilityUsed,omitempty"`
	ButlerMasters             map[string]string       `json:"butlerMasters,omitempty"`
	FortuneTellerRedHerringID string                  `json:"fortuneTellerRedHerringId,omitempty"`
	DemonBluffCharacterIDs    []string                `json:"demonBluffCharacterIds,omitempty"`
	GrimoireRevealed          bool                    `json:"grimoireRevealed,omitempty"`
	Winner                    *game.GameEndedEvent    `json:"winner,omitempty"`
}

type persistedGameSessionSnapshot struct {
	gameSessionSnapshot
	History gameHistory `json:"operationHistory"`
}

func (gs *GameSession) Clone() *GameSession {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	cloned := newGameSessionFromSnapshot(gs.snapshotLocked())
	cloned.history = gs.history.clone()
	return cloned
}

func (gs *GameSession) MarshalSnapshot() ([]byte, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return json.Marshal(persistedGameSessionSnapshot{gameSessionSnapshot: gs.snapshotLocked(), History: gs.history})
}

func LoadGameSessionSnapshot(data []byte) (*GameSession, error) {
	var snapshot persistedGameSessionSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	loaded := newGameSessionFromSnapshot(snapshot.gameSessionSnapshot)
	loaded.history = snapshot.History
	return loaded, nil
}

func (gs *GameSession) snapshot() gameSessionSnapshot {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.snapshotLocked()
}

func (gs *GameSession) snapshotLocked() gameSessionSnapshot {
	return gameSessionSnapshot{
		Players:                   clonePlayers(gs.players),
		StorytellerID:             gs.storytellerID,
		StorytellerName:           gs.storytellerName,
		OriginalPlayers:           gs.originalPlayers,
		ScriptID:                  gs.scriptID,
		Phase:                     gs.phase,
		DayNumber:                 gs.dayNumber,
		NightNumber:               gs.nightNumber,
		NightWakeIndex:            gs.nightWakeIndex,
		Nomination:                cloneNomination(gs.nomination),
		NominationResults:         cloneNominationResults(gs.nominationResults),
		NightActions:              cloneNightActions(gs.nightActions),
		PendingNightAction:        cloneNightAction(gs.pendingNightAction),
		ConfirmedNightAction:      cloneNightAction(gs.confirmedNightAction),
		NightAcknowledged:         cloneBoolMap(gs.nightAcknowledged),
		DawnReviewPending:         gs.dawnReviewPending,
		PendingDawnDeathIDs:       append([]string(nil), gs.pendingDawnDeathIDs...),
		Deaths:                    cloneDeaths(gs.deaths),
		GhostVotesUsed:            cloneBoolMap(gs.ghostVotesUsed),
		SlayerUsed:                cloneBoolMap(gs.slayerUsed),
		NominatorsToday:           cloneBoolMap(gs.nominatorsToday),
		NomineesToday:             cloneBoolMap(gs.nomineesToday),
		VirginAbilityUsed:         cloneBoolMap(gs.virginAbilityUsed),
		ButlerMasters:             cloneStringMap(gs.butlerMasters),
		FortuneTellerRedHerringID: gs.fortuneTellerRedHerringID,
		DemonBluffCharacterIDs:    append([]string(nil), gs.demonBluffCharacterIDs...),
		GrimoireRevealed:          gs.grimoireRevealed,
		Winner:                    cloneWinner(gs.winner),
	}
}

func newGameSessionFromSnapshot(snapshot gameSessionSnapshot) *GameSession {
	gs := &GameSession{}
	gs.restoreSnapshotLocked(snapshot)
	return gs
}

// restoreSnapshotLocked replaces only game state; the mutex and history stay live.
func (gs *GameSession) restoreSnapshotLocked(snapshot gameSessionSnapshot) {
	scriptID := snapshot.ScriptID
	if scriptID == "" {
		scriptID = game.TroubleBrewingScriptID
	}
	gs.players = clonePlayers(snapshot.Players)
	gs.storytellerID = snapshot.StorytellerID
	gs.storytellerName = snapshot.StorytellerName
	gs.originalPlayers = snapshot.OriginalPlayers
	gs.scriptID = scriptID
	gs.phase = snapshot.Phase
	gs.dayNumber = snapshot.DayNumber
	gs.nightNumber = snapshot.NightNumber
	gs.nightWakeIndex = snapshot.NightWakeIndex
	gs.nomination = cloneNomination(snapshot.Nomination)
	gs.nominationResults = cloneNominationResults(snapshot.NominationResults)
	gs.nightActions = cloneNightActions(snapshot.NightActions)
	gs.pendingNightAction = cloneNightAction(snapshot.PendingNightAction)
	gs.confirmedNightAction = cloneNightAction(snapshot.ConfirmedNightAction)
	gs.nightAcknowledged = cloneBoolMap(snapshot.NightAcknowledged)
	gs.dawnReviewPending = snapshot.DawnReviewPending
	gs.pendingDawnDeathIDs = append([]string(nil), snapshot.PendingDawnDeathIDs...)
	gs.deaths = cloneDeaths(snapshot.Deaths)
	gs.ghostVotesUsed = nonNilBoolMap(snapshot.GhostVotesUsed)
	gs.slayerUsed = nonNilBoolMap(snapshot.SlayerUsed)
	gs.nominatorsToday = nonNilBoolMap(snapshot.NominatorsToday)
	gs.nomineesToday = nonNilBoolMap(snapshot.NomineesToday)
	gs.virginAbilityUsed = nonNilBoolMap(snapshot.VirginAbilityUsed)
	gs.butlerMasters = nonNilStringMap(snapshot.ButlerMasters)
	gs.fortuneTellerRedHerringID = snapshot.FortuneTellerRedHerringID
	gs.demonBluffCharacterIDs = append([]string(nil), snapshot.DemonBluffCharacterIDs...)
	gs.grimoireRevealed = snapshot.GrimoireRevealed
	gs.winner = cloneWinner(snapshot.Winner)
}

func nonNilBoolMap(values map[string]bool) map[string]bool {
	cloned := cloneBoolMap(values)
	if cloned == nil {
		return make(map[string]bool)
	}
	return cloned
}

func nonNilStringMap(values map[string]string) map[string]string {
	cloned := cloneStringMap(values)
	if cloned == nil {
		return make(map[string]string)
	}
	return cloned
}

func clonePlayers(players []game.Player) []game.Player {
	result := make([]game.Player, len(players))
	for i, player := range players {
		result[i] = player
		if player.Character != nil {
			character := *player.Character
			result[i].Character = &character
		}
		if player.ShownCharacter != nil {
			shownCharacter := *player.ShownCharacter
			result[i].ShownCharacter = &shownCharacter
		}
	}
	return result
}

func cloneNomination(nomination *game.Nomination) *game.Nomination {
	if nomination == nil {
		return nil
	}
	result := *nomination
	if nomination.Votes != nil {
		result.Votes = make(map[string]bool, len(nomination.Votes))
		for playerID, decision := range nomination.Votes {
			result.Votes[playerID] = decision
		}
	}
	result.VoterOrder = append([]string(nil), nomination.VoterOrder...)
	return &result
}

func cloneNominationResults(results []game.NominationResult) []game.NominationResult {
	cloned := make([]game.NominationResult, len(results))
	for index, result := range results {
		cloned[index] = result
		cloned[index].Votes = cloneBoolMap(result.Votes)
	}
	return cloned
}

func cloneNightActions(actions []game.NightAction) []game.NightAction {
	result := make([]game.NightAction, len(actions))
	for i, action := range actions {
		result[i] = action
		result[i].TargetIDs = append([]string{}, action.TargetIDs...)
	}
	return result
}

func cloneNightAction(action *game.NightAction) *game.NightAction {
	if action == nil {
		return nil
	}
	result := *action
	result.TargetIDs = append([]string(nil), action.TargetIDs...)
	return &result
}

func cloneDeaths(deaths []game.DeathRecord) []game.DeathRecord {
	return append([]game.DeathRecord(nil), deaths...)
}

func cloneBoolMap(values map[string]bool) map[string]bool {
	if values == nil {
		return nil
	}
	result := make(map[string]bool, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func cloneWinner(winner *game.GameEndedEvent) *game.GameEndedEvent {
	if winner == nil {
		return nil
	}
	result := *winner
	return &result
}
