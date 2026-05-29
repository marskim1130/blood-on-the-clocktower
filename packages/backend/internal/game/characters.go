package game

// CharacterType represents the type of character
type CharacterType int

const (
	CharacterTypeTownsfolk CharacterType = iota
	CharacterTypeOutsider
	CharacterTypeMinion
	CharacterTypeDemon
)

// CharacterDefinition defines a Trouble Brewing character
type CharacterDefinition struct {
	ID      string
	Name    string
	Type    CharacterType
	Team    Team
	Ability string
}

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
		Ability: "Each night, choose a player: they die. If you kill yourself this way, a Minion becomes the Imp."},
}

// GetCharacterByID returns a character definition by ID
func GetCharacterByID(id string) *CharacterDefinition {
	for _, c := range TroubleBrewing {
		if c.ID == id {
			return &c
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
	4:  {Townsfolk: 3, Outsiders: 0, Minions: 0, Demons: 1},
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

// ValidateAssignment checks if a character assignment is valid for the given player count
func ValidateAssignment(assignments map[string]string, playerCount int) bool {
	counts, ok := ValidRoleCounts[playerCount]
	if !ok {
		return false
	}

	typeCounts := map[CharacterType]int{}
	for _, charID := range assignments {
		def := GetCharacterByID(charID)
		if def == nil {
			return false
		}
		typeCounts[def.Type]++
	}

	return typeCounts[CharacterTypeTownsfolk] == counts.Townsfolk &&
		typeCounts[CharacterTypeOutsider] == counts.Outsiders &&
		typeCounts[CharacterTypeMinion] == counts.Minions &&
		typeCounts[CharacterTypeDemon] == counts.Demons
}
