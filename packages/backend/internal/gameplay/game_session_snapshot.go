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

func (gs *GameSession) Clone() *GameSession {
	return newGameSessionFromSnapshot(gs.snapshot())
}

func (gs *GameSession) MarshalSnapshot() ([]byte, error) {
	return json.Marshal(gs.snapshot())
}

func LoadGameSessionSnapshot(data []byte) (*GameSession, error) {
	var snapshot gameSessionSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return newGameSessionFromSnapshot(snapshot), nil
}

func (gs *GameSession) snapshot() gameSessionSnapshot {
	gs.mu.Lock()
	defer gs.mu.Unlock()

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
	scriptID := snapshot.ScriptID
	if scriptID == "" {
		scriptID = game.TroubleBrewingScriptID
	}
	return &GameSession{
		players:                   clonePlayers(snapshot.Players),
		storytellerID:             snapshot.StorytellerID,
		storytellerName:           snapshot.StorytellerName,
		originalPlayers:           snapshot.OriginalPlayers,
		scriptID:                  scriptID,
		phase:                     snapshot.Phase,
		dayNumber:                 snapshot.DayNumber,
		nightNumber:               snapshot.NightNumber,
		nightWakeIndex:            snapshot.NightWakeIndex,
		nomination:                cloneNomination(snapshot.Nomination),
		nominationResults:         cloneNominationResults(snapshot.NominationResults),
		nightActions:              cloneNightActions(snapshot.NightActions),
		pendingNightAction:        cloneNightAction(snapshot.PendingNightAction),
		confirmedNightAction:      cloneNightAction(snapshot.ConfirmedNightAction),
		nightAcknowledged:         cloneBoolMap(snapshot.NightAcknowledged),
		dawnReviewPending:         snapshot.DawnReviewPending,
		pendingDawnDeathIDs:       append([]string(nil), snapshot.PendingDawnDeathIDs...),
		deaths:                    cloneDeaths(snapshot.Deaths),
		ghostVotesUsed:            nonNilBoolMap(snapshot.GhostVotesUsed),
		slayerUsed:                nonNilBoolMap(snapshot.SlayerUsed),
		nominatorsToday:           nonNilBoolMap(snapshot.NominatorsToday),
		nomineesToday:             nonNilBoolMap(snapshot.NomineesToday),
		virginAbilityUsed:         nonNilBoolMap(snapshot.VirginAbilityUsed),
		butlerMasters:             nonNilStringMap(snapshot.ButlerMasters),
		fortuneTellerRedHerringID: snapshot.FortuneTellerRedHerringID,
		demonBluffCharacterIDs:    append([]string(nil), snapshot.DemonBluffCharacterIDs...),
		grimoireRevealed:          snapshot.GrimoireRevealed,
		winner:                    cloneWinner(snapshot.Winner),
	}
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
