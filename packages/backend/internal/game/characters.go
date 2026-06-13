package game

// CharacterType represents the type of character
type CharacterType int

const (
	CharacterTypeTownsfolk CharacterType = iota
	CharacterTypeOutsider
	CharacterTypeMinion
	CharacterTypeDemon
)

const (
	NightWakeCharacterTypeMinion = "minion"
	NightWakeCharacterTypeDemon  = "demon"
)

// CharacterDefinition defines a Trouble Brewing character
type CharacterDefinition struct {
	ID      string
	Name    string
	Type    CharacterType
	Team    Team
	Ability string
}

// ScriptDefinition defines a playable script and its character pool.
type ScriptDefinition struct {
	ID         string
	Name       string
	Characters []CharacterDefinition
}

const TroubleBrewingScriptID = "trouble_brewing"

// TroubleBrewing contains all Trouble Brewing character definitions
var TroubleBrewing = []CharacterDefinition{
	// Townsfolk
	{ID: "washerwoman", Name: "Washerwoman", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "You start knowing that one of two players is a particular Townsfolk."},
	{ID: "librarian", Name: "Librarian", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "You start knowing that one of two players is a particular Outsider."},
	{ID: "investigator", Name: "Investigator", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "You start knowing that one of two players is a particular Minion."},
	{ID: "chef", Name: "Chef", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "You start knowing how many pairs of evil players there are."},
	{ID: "empath", Name: "Empath", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "Each night, you learn how many of your alive neighbours are evil."},
	{ID: "fortuneteller", Name: "Fortune Teller", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "Each night, choose 2 players: you learn if either is a Demon."},
	{ID: "undertaker", Name: "Undertaker", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "Each night, you learn which character died by execution today."},
	{ID: "monk", Name: "Monk", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "Each night, choose a player: they are safe from the Demon tonight."},
	{ID: "ravenkeeper", Name: "Ravenkeeper", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "If you die at night, you are woken to choose a player: you learn their character."},
	{ID: "virgin", Name: "Virgin", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "The first time you are nominated, if the nominator is a Townsfolk, they are executed immediately."},
	{ID: "slayer", Name: "Slayer", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "Once per game, during the day, choose a player: if they are the Demon, they die."},
	{ID: "soldier", Name: "Soldier", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "You are safe from the Demon at night."},
	{ID: "mayor", Name: "Mayor", Type: CharacterTypeTownsfolk, Team: TeamGood,
		Ability: "If only 3 players live & no execution occurs, your team wins. If you die at night, another player might die instead."},

	// Outsiders
	{ID: "butler", Name: "Butler", Type: CharacterTypeOutsider, Team: TeamGood,
		Ability: "Each night, choose a player: you may only vote if they vote."},
	{ID: "drunk", Name: "Drunk", Type: CharacterTypeOutsider, Team: TeamGood,
		Ability: "You do not know you are the Drunk. You think you are a Townsfolk, but your ability malfunctions."},
	{ID: "recluse", Name: "Recluse", Type: CharacterTypeOutsider, Team: TeamGood,
		Ability: "You might register as evil, even if you are good."},
	{ID: "saint", Name: "Saint", Type: CharacterTypeOutsider, Team: TeamGood,
		Ability: "If you are executed, your team loses."},

	// Minions
	{ID: "poisoner", Name: "Poisoner", Type: CharacterTypeMinion, Team: TeamEvil,
		Ability: "Each night, choose a player: they are poisoned until next dusk."},
	{ID: "spy", Name: "Spy", Type: CharacterTypeMinion, Team: TeamEvil,
		Ability: "Each night, you see the Grimoire. You might register as good or as a Townsfolk or Outsider."},
	{ID: "baron", Name: "Baron", Type: CharacterTypeMinion, Team: TeamEvil,
		Ability: "There are extra Outsiders in play."},
	{ID: "scarletwoman", Name: "Scarlet Woman", Type: CharacterTypeMinion, Team: TeamEvil,
		Ability: "If there are 5 or more players alive & the Demon dies, you become the Demon."},

	// Demon
	{ID: "imp", Name: "Imp", Type: CharacterTypeDemon, Team: TeamEvil,
		Ability: "Each night except the first, choose a player: they die. If you kill yourself this way, a Minion becomes the Imp."},
}

// TroubleBrewingScript is the default supported script.
var TroubleBrewingScript = ScriptDefinition{
	ID:         TroubleBrewingScriptID,
	Name:       "Trouble Brewing",
	Characters: TroubleBrewing,
}

var scriptsByID = map[string]*ScriptDefinition{
	TroubleBrewingScriptID: &TroubleBrewingScript,
}

// GetScriptByID returns a playable script definition by ID.
func GetScriptByID(id string) *ScriptDefinition {
	return scriptsByID[id]
}

// GetCharacterByID returns a character definition by ID
func GetCharacterByID(id string) *CharacterDefinition {
	return GetScriptCharacterByID(TroubleBrewingScriptID, id)
}

// GetScriptCharacterByID returns a character definition from a specific script.
func GetScriptCharacterByID(scriptID, characterID string) *CharacterDefinition {
	script := GetScriptByID(scriptID)
	if script == nil {
		return nil
	}
	for i := range script.Characters {
		if script.Characters[i].ID == characterID {
			return &script.Characters[i]
		}
	}
	return nil
}

// RoleCount defines the number of each character type per player count
type RoleCount struct {
	Townsfolk int
	Outsiders int
	Minions   int
	Demons    int
}

// ValidRoleCounts maps player count to valid role distribution
// Source: Trouble Brewing rulebook
// Note: These are for actual players (excluding storyteller)
var ValidRoleCounts = map[int]RoleCount{
	5:  {Townsfolk: 3, Outsiders: 0, Minions: 1, Demons: 1},
	6:  {Townsfolk: 3, Outsiders: 1, Minions: 1, Demons: 1},
	7:  {Townsfolk: 5, Outsiders: 0, Minions: 1, Demons: 1},
	8:  {Townsfolk: 5, Outsiders: 1, Minions: 1, Demons: 1},
	9:  {Townsfolk: 5, Outsiders: 2, Minions: 1, Demons: 1},
	10: {Townsfolk: 7, Outsiders: 0, Minions: 2, Demons: 1},
	11: {Townsfolk: 7, Outsiders: 1, Minions: 2, Demons: 1},
	12: {Townsfolk: 7, Outsiders: 2, Minions: 2, Demons: 1},
	13: {Townsfolk: 9, Outsiders: 0, Minions: 3, Demons: 1},
	14: {Townsfolk: 9, Outsiders: 1, Minions: 3, Demons: 1},
	15: {Townsfolk: 9, Outsiders: 2, Minions: 3, Demons: 1},
}

// TroubleBrewingFirstNightOrder defines the storyteller wake order for night one.
var TroubleBrewingFirstNightOrder = []NightWakeStep{
	{CharacterType: NightWakeCharacterTypeMinion, Order: 1, ActionType: NightActionLearnDemon, Prompt: "Minions learn which player is the Demon.", MinTargets: 0, MaxTargets: 0},
	{CharacterType: NightWakeCharacterTypeDemon, Order: 2, ActionType: NightActionLearnMinion, Prompt: "Demon learns which players are Minions.", MinTargets: 0, MaxTargets: 0},
	{CharacterID: "poisoner", Order: 3, ActionType: NightActionPoison, Prompt: "Poisoner chooses one player to poison until dusk.", MinTargets: 1, MaxTargets: 1},
	{CharacterID: "washerwoman", Order: 4, ActionType: NightActionLearnTownsfolk, Prompt: "Washerwoman learns that one of two players is a specific Townsfolk.", MinTargets: 2, MaxTargets: 2},
	{CharacterID: "librarian", Order: 5, ActionType: NightActionLearnOutsider, Prompt: "Librarian learns that one of two players is a specific Outsider, or that none are in play.", MinTargets: 0, MaxTargets: 2},
	{CharacterID: "investigator", Order: 6, ActionType: NightActionLearnMinion, Prompt: "Investigator learns that one of two players is a specific Minion.", MinTargets: 2, MaxTargets: 2},
	{CharacterID: "chef", Order: 7, ActionType: NightActionLearnEvilPairs, Prompt: "Chef learns the number of adjacent evil pairs.", MinTargets: 0, MaxTargets: 0},
	{CharacterID: "empath", Order: 8, ActionType: NightActionLearnEvilNeighbors, Prompt: "Empath learns how many alive neighbours are evil.", MinTargets: 0, MaxTargets: 0},
	{CharacterID: "fortuneteller", Order: 9, ActionType: NightActionCheckDemon, Prompt: "Fortune Teller chooses two players and learns if either registers as the Demon.", MinTargets: 2, MaxTargets: 2},
	{CharacterID: "butler", Order: 10, ActionType: NightActionLearnMaster, Prompt: "Butler chooses their master for tomorrow.", MinTargets: 1, MaxTargets: 1},
}

// TroubleBrewingSubsequentNightOrder defines the storyteller wake order after night one.
var TroubleBrewingSubsequentNightOrder = []NightWakeStep{
	{CharacterID: "poisoner", Order: 1, ActionType: NightActionPoison, Prompt: "Poisoner chooses one player to poison until dusk.", MinTargets: 1, MaxTargets: 1},
	{CharacterID: "monk", Order: 2, ActionType: NightActionProtect, Prompt: "Monk chooses one player other than themself to protect from the Demon.", MinTargets: 1, MaxTargets: 1},
	{CharacterID: "imp", Order: 3, ActionType: NightActionKill, Prompt: "Imp chooses one player to die.", MinTargets: 1, MaxTargets: 1},
	{CharacterID: "empath", Order: 4, ActionType: NightActionLearnEvilNeighbors, Prompt: "Empath learns how many alive neighbours are evil.", MinTargets: 0, MaxTargets: 0},
	{CharacterID: "fortuneteller", Order: 5, ActionType: NightActionCheckDemon, Prompt: "Fortune Teller chooses two players and learns if either registers as the Demon.", MinTargets: 2, MaxTargets: 2},
	{CharacterID: "undertaker", Order: 6, ActionType: NightActionLearnExecuted, Prompt: "Undertaker learns which character died by execution today.", MinTargets: 0, MaxTargets: 0},
	{CharacterID: "butler", Order: 7, ActionType: NightActionLearnMaster, Prompt: "Butler chooses their master for tomorrow.", MinTargets: 1, MaxTargets: 1},
	{CharacterID: "ravenkeeper", Order: 8, ActionType: NightActionLearnDied, Prompt: "If Ravenkeeper died tonight, they choose one player and learn their character.", MinTargets: 1, MaxTargets: 1},
}

// GetScriptWakeOrder returns the full wake order for a script and night number.
func GetScriptWakeOrder(scriptID string, nightNumber int32) []NightWakeStep {
	if scriptID != TroubleBrewingScriptID {
		return nil
	}
	if nightNumber <= 1 {
		return cloneNightWakeSteps(TroubleBrewingFirstNightOrder)
	}
	return cloneNightWakeSteps(TroubleBrewingSubsequentNightOrder)
}

// GetActiveNightWakeSteps filters the script wake order down to assigned, alive characters.
func GetActiveNightWakeSteps(scriptID string, nightNumber int32, players []Player) []NightWakeStep {
	inPlay := map[string]bool{}
	characterTypesInPlay := map[string]bool{}
	for _, player := range players {
		if player.Character == nil || !player.IsAlive {
			continue
		}
		inPlay[player.Character.ID] = true
		if player.Character.ID == "drunk" && player.ShownCharacter != nil {
			inPlay[player.ShownCharacter.ID] = true
		}
		if charDef := GetScriptCharacterByID(scriptID, player.Character.ID); charDef != nil {
			if characterType := characterTypeKey(charDef.Type); characterType != "" {
				characterTypesInPlay[characterType] = true
			}
		}
	}

	steps := GetScriptWakeOrder(scriptID, nightNumber)
	active := make([]NightWakeStep, 0, len(steps))
	for _, step := range steps {
		if step.CharacterID != "" && inPlay[step.CharacterID] {
			active = append(active, step)
			continue
		}
		if step.CharacterType != "" && characterTypesInPlay[step.CharacterType] {
			active = append(active, step)
		}
	}
	return active
}

func characterTypeKey(characterType CharacterType) string {
	switch characterType {
	case CharacterTypeMinion:
		return NightWakeCharacterTypeMinion
	case CharacterTypeDemon:
		return NightWakeCharacterTypeDemon
	default:
		return ""
	}
}

func cloneNightWakeSteps(steps []NightWakeStep) []NightWakeStep {
	result := make([]NightWakeStep, len(steps))
	copy(result, steps)
	return result
}

// ValidateAssignment checks if a character assignment is valid for the given player count
func ValidateAssignment(assignments map[string]string, playerCount int) bool {
	return ValidateScriptAssignment(TroubleBrewingScriptID, assignments, playerCount)
}

// ValidateScriptAssignment checks if a character assignment is valid for a script.
func ValidateScriptAssignment(scriptID string, assignments map[string]string, playerCount int) bool {
	if GetScriptByID(scriptID) == nil {
		return false
	}

	counts, ok := ValidRoleCounts[playerCount]
	if !ok {
		return false
	}
	if len(assignments) != playerCount {
		return false
	}

	typeCounts := map[CharacterType]int{}
	seenCharacters := map[string]bool{}
	hasBaron := false
	for _, charID := range assignments {
		if seenCharacters[charID] {
			return false
		}
		seenCharacters[charID] = true
		if charID == "baron" {
			hasBaron = true
		}

		def := GetScriptCharacterByID(scriptID, charID)
		if def == nil {
			return false
		}
		typeCounts[def.Type]++
	}

	if hasBaron {
		counts.Townsfolk -= 2
		counts.Outsiders += 2
		if counts.Townsfolk < 0 {
			return false
		}
	}

	return typeCounts[CharacterTypeTownsfolk] == counts.Townsfolk &&
		typeCounts[CharacterTypeOutsider] == counts.Outsiders &&
		typeCounts[CharacterTypeMinion] == counts.Minions &&
		typeCounts[CharacterTypeDemon] == counts.Demons
}
