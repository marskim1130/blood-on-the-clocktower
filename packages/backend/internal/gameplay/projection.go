package gameplay

import "github.com/marskim1130/blood-on-the-clocktower/internal/game"

const (
	NightTurnAwaitingPlayer         = "awaiting_player"
	NightTurnAwaitingStoryteller    = "awaiting_storyteller"
	NightTurnAwaitingAcknowledgment = "awaiting_acknowledgement"
)

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
	NominationResults         []game.NominationResult
	ExecutionCandidateID      string
	ExecutionCandidateVotes   int
	ExecutionTied             bool
	Deaths                    []game.DeathRecord
	GhostVotesRemaining       []string
	NightWakeSteps            []game.NightWakeStep
	CurrentNightWakeIndex     int
	CurrentNightWakeStep      *game.NightWakeStep
	Winner                    *game.GameEndedEvent
	NightActions              []game.NightAction
	NightTurnStatus           string
	PendingNightAction        *game.NightAction
	ConfirmedNightAction      *game.NightAction
	DemonBluffCharacterIDs    []string
	GrimoireRevealed          bool
	DawnReviewPending         bool
	PendingDawnDeathIDs       []string
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
	deaths := cloneDeaths(gs.deaths)
	if !capabilities.seeDeathCauses {
		for index := range deaths {
			deaths[index].Cause = ""
			deaths[index].KilledBy = ""
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
	nightTurnStatus := ""
	var pendingNightAction *game.NightAction
	var confirmedNightAction *game.NightAction
	dawnReviewPending := false
	var pendingDawnDeathIDs []string
	fortuneTellerRedHerringID := ""
	var demonBluffCharacterIDs []string
	if forceSeeAll || (gs.storytellerID != "" && recipientID == gs.storytellerID) {
		fortuneTellerRedHerringID = gs.fortuneTellerRedHerringID
		demonBluffCharacterIDs = append([]string(nil), gs.demonBluffCharacterIDs...)
	}
	if gs.phase == game.GamePhaseNight && capabilities.seeNightManagement {
		nightWakeSteps = gs.activeNightWakeStepsLocked()
		currentNightWakeIndex = gs.nightWakeIndex
		currentNightWakeStep = gs.currentNightWakeStepLocked()
		nightActions = cloneNightActions(gs.nightActions)
		nightTurnStatus = gs.nightTurnStatusLocked(currentNightWakeStep)
		pendingNightAction = cloneNightAction(gs.pendingNightAction)
		confirmedNightAction = cloneNightAction(gs.confirmedNightAction)
		dawnReviewPending = gs.dawnReviewPending
		pendingDawnDeathIDs = append([]string(nil), gs.pendingDawnDeathIDs...)
	} else if gs.phase == game.GamePhaseNight {
		step := gs.currentNightWakeStepLocked()
		playerIndex := gs.findPlayerIndex(recipientID)
		if step != nil && playerIndex != -1 && gs.players[playerIndex].IsAlive && !gs.nightAcknowledged[recipientID] && gs.playerMatchesNightWakeStepLocked(playerIndex, *step) {
			currentNightWakeStep = step
			if storytellerChoosesNightTargets(*step) {
				currentNightWakeStep.MinTargets = 0
				currentNightWakeStep.MaxTargets = 0
			}
			nightTurnStatus = gs.nightTurnStatusLocked(step)
			if gs.pendingNightAction != nil && gs.pendingNightAction.ActorID == recipientID {
				pendingNightAction = cloneNightAction(gs.pendingNightAction)
			}
			confirmedNightAction = cloneNightAction(gs.confirmedNightAction)
		}
	}
	executionCandidateID, executionCandidateVotes, executionTied := gs.executionBlockLocked()

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
		NominationResults:         cloneNominationResults(gs.nominationResults),
		ExecutionCandidateID:      executionCandidateID,
		ExecutionCandidateVotes:   executionCandidateVotes,
		ExecutionTied:             executionTied,
		Deaths:                    deaths,
		GhostVotesRemaining:       ghostVotesRemaining,
		NightWakeSteps:            nightWakeSteps,
		CurrentNightWakeIndex:     currentNightWakeIndex,
		CurrentNightWakeStep:      currentNightWakeStep,
		Winner:                    cloneWinner(gs.winner),
		NightActions:              nightActions,
		NightTurnStatus:           nightTurnStatus,
		PendingNightAction:        pendingNightAction,
		ConfirmedNightAction:      confirmedNightAction,
		DemonBluffCharacterIDs:    demonBluffCharacterIDs,
		GrimoireRevealed:          gs.grimoireRevealed,
		DawnReviewPending:         dawnReviewPending,
		PendingDawnDeathIDs:       pendingDawnDeathIDs,
		FortuneTellerRedHerringID: fortuneTellerRedHerringID,
	}
}

func (gs *GameSession) nightTurnStatusLocked(step *game.NightWakeStep) string {
	if step == nil {
		return ""
	}
	if gs.confirmedNightAction != nil {
		return NightTurnAwaitingAcknowledgment
	}
	if gs.pendingNightAction != nil {
		return NightTurnAwaitingStoryteller
	}
	return NightTurnAwaitingPlayer
}

func (gs *GameSession) executionBlockLocked() (string, int, bool) {
	candidateID := ""
	highestVotes := 0
	tied := false
	for _, result := range gs.nominationResults {
		if result.DayNumber != gs.dayNumber || result.YesVotes < result.RequiredVotes {
			continue
		}
		switch {
		case result.YesVotes > highestVotes:
			candidateID = result.NomineeID
			highestVotes = result.YesVotes
			tied = false
		case result.YesVotes == highestVotes:
			candidateID = ""
			tied = true
		}
	}
	return candidateID, highestVotes, tied
}

type projectionCapabilities struct {
	seeAllCharacters   bool
	seePoisoning       bool
	seeNightManagement bool
	seeDeathCauses     bool
}

func (gs *GameSession) projectionCapabilitiesLocked(forceSeeAll bool, recipientID string) projectionCapabilities {
	if forceSeeAll || (gs.storytellerID != "" && recipientID == gs.storytellerID) {
		return projectionCapabilities{seeAllCharacters: true, seePoisoning: true, seeNightManagement: true, seeDeathCauses: true}
	}
	if (gs.phase == game.GamePhaseFinished || gs.winner != nil) && gs.grimoireRevealed {
		return projectionCapabilities{seeAllCharacters: true}
	}
	return projectionCapabilities{}
}
