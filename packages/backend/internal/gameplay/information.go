package gameplay

import (
	"fmt"
	"strings"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

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
	if err := gs.validateNightSelfTargetLocked(step, targetIDs); err != nil {
		return err
	}
	return nil
}

func (gs *GameSession) validateNightSelfTargetLocked(step game.NightWakeStep, targetIDs []string) error {
	if len(targetIDs) == 0 || step.CharacterID == "" {
		return nil
	}
	actorIdx := gs.findLivingVisibleCharacterIndexLocked(step.CharacterID)
	if actorIdx == -1 || targetIDs[0] != gs.players[actorIdx].ID {
		return nil
	}

	switch step.ActionType {
	case game.NightActionProtect:
		if step.CharacterID == "monk" {
			return fmt.Errorf("monk cannot protect themself")
		}
	case game.NightActionLearnMaster:
		if step.CharacterID == "butler" {
			return fmt.Errorf("butler cannot choose themself as master")
		}
	}
	return nil
}

func (gs *GameSession) playerMatchesNightWakeStepLocked(playerIdx int, step game.NightWakeStep) bool {
	if step.CharacterID != "" {
		return gs.playerCanUseVisibleCharacterLocked(playerIdx, step.CharacterID)
	}

	switch step.CharacterType {
	case game.NightWakeCharacterTypeMinion:
		return gs.playerHasCharacterTypeLocked(playerIdx, game.CharacterTypeMinion)
	case game.NightWakeCharacterTypeDemon:
		return gs.playerHasCharacterTypeLocked(playerIdx, game.CharacterTypeDemon)
	default:
		return false
	}
}

func (gs *GameSession) findLivingVisibleCharacterIndexLocked(characterID string) int {
	for i := range gs.players {
		if gs.players[i].IsAlive && gs.playerCanUseVisibleCharacterLocked(i, characterID) {
			return i
		}
	}
	return -1
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
