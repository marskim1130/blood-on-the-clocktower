package gameplay

import "github.com/marskim1130/blood-on-the-clocktower/internal/game"

// Projection is the privacy-safe game view for one recipient.
// Room lifecycle metadata is added by the WebSocket adapter.
type Projection struct {
	Players                   []game.Player
	ScriptID                  string
	ScriptName                string
	StorytellerID             string
	StorytellerName           string
	Phase                     game.GamePhase
	DayNumber                 int32
	NightNumber               int32
	Nomination                *game.Nomination
	Deaths                    []game.DeathRecord
	GhostVotesRemaining       []string
	NightWakeSteps            []game.NightWakeStep
	CurrentNightWakeIndex     int
	CurrentNightWakeStep      *game.NightWakeStep
	Winner                    *game.GameEndedEvent
	NightActions              []game.NightAction
	FortuneTellerRedHerringID string
}

// Projection builds a complete game projection for trusted internal use.
func (gs *GameSession) Projection() *Projection {
	return gs.project(true, "")
}

// ProjectionFor builds a recipient-specific game projection.
// Storyteller sees all character assignments; players usually only see their own.
func (gs *GameSession) ProjectionFor(recipientID string) *Projection {
	return gs.project(false, recipientID)
}

func (gs *GameSession) project(forceSeeAll bool, recipientID string) *Projection {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	capabilities := gs.projectionCapabilitiesLocked(forceSeeAll, recipientID)
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
		if !capabilities.seeAllCharacters {
			if player.ID == recipientID && player.Character != nil && player.Character.ID == "drunk" && player.ShownCharacter != nil {
				shownCharacter := *player.ShownCharacter
				players[i].Character = &shownCharacter
			} else if player.ID != recipientID {
				players[i].Character = nil
			}
			players[i].ShownCharacter = nil
		}
		// Hide PoisonedUntil from all non-storyteller recipients (including the poisoned player)
		if !capabilities.seePoisoning {
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
	var nightActions []game.NightAction
	fortuneTellerRedHerringID := ""
	if forceSeeAll || (gs.storytellerID != "" && recipientID == gs.storytellerID) {
		fortuneTellerRedHerringID = gs.fortuneTellerRedHerringID
	}
	if gs.phase == game.GamePhaseNight && capabilities.seeNightManagement {
		nightWakeSteps = gs.activeNightWakeStepsLocked()
		currentNightWakeIndex = gs.nightWakeIndex
		currentNightWakeStep = gs.currentNightWakeStepLocked()
		nightActions = cloneNightActions(gs.nightActions)
	}

	return &Projection{
		Players:                   players,
		ScriptID:                  gs.scriptID,
		ScriptName:                scriptName,
		StorytellerID:             gs.storytellerID,
		StorytellerName:           gs.storytellerName,
		Phase:                     gs.phase,
		DayNumber:                 gs.dayNumber,
		NightNumber:               gs.nightNumber,
		Nomination:                cloneNomination(gs.nomination),
		Deaths:                    cloneDeaths(gs.deaths),
		GhostVotesRemaining:       ghostVotesRemaining,
		NightWakeSteps:            nightWakeSteps,
		CurrentNightWakeIndex:     currentNightWakeIndex,
		CurrentNightWakeStep:      currentNightWakeStep,
		Winner:                    cloneWinner(gs.winner),
		NightActions:              nightActions,
		FortuneTellerRedHerringID: fortuneTellerRedHerringID,
	}
}

type projectionCapabilities struct {
	seeAllCharacters   bool
	seePoisoning       bool
	seeNightManagement bool
}

func (gs *GameSession) projectionCapabilitiesLocked(forceSeeAll bool, recipientID string) projectionCapabilities {
	if forceSeeAll || (gs.storytellerID != "" && recipientID == gs.storytellerID) {
		return projectionCapabilities{seeAllCharacters: true, seePoisoning: true, seeNightManagement: true}
	}
	if gs.phase == game.GamePhaseFinished || gs.winner != nil {
		return projectionCapabilities{seeAllCharacters: true}
	}
	if gs.phase != game.GamePhaseNight {
		return projectionCapabilities{}
	}
	playerIdx := gs.findPlayerIndex(recipientID)
	if playerIdx == -1 || !gs.players[playerIdx].IsAlive || gs.players[playerIdx].Character == nil {
		return projectionCapabilities{}
	}
	return projectionCapabilities{seeAllCharacters: gs.players[playerIdx].Character.ID == "spy" && !gs.playerIsPoisonedLocked(playerIdx)}
}
