package game

import "testing"

func TestTroubleBrewingCharacterCount(t *testing.T) {
	// Trouble Brewing should have 13 townsfolk, 4 outsiders, 4 minions, 1 demon = 22 total
	typeCounts := map[CharacterType]int{}
	for _, c := range TroubleBrewing {
		typeCounts[c.Type]++
	}

	if typeCounts[CharacterTypeTownsfolk] != 13 {
		t.Errorf("expected 13 townsfolk, got %d", typeCounts[CharacterTypeTownsfolk])
	}
	if typeCounts[CharacterTypeOutsider] != 4 {
		t.Errorf("expected 4 outsiders, got %d", typeCounts[CharacterTypeOutsider])
	}
	if typeCounts[CharacterTypeMinion] != 4 {
		t.Errorf("expected 4 minions, got %d", typeCounts[CharacterTypeMinion])
	}
	if typeCounts[CharacterTypeDemon] != 1 {
		t.Errorf("expected 1 demon, got %d", typeCounts[CharacterTypeDemon])
	}
}

func TestAllCharactersHaveAbilities(t *testing.T) {
	for _, c := range TroubleBrewing {
		if c.Ability == "" {
			t.Errorf("character %s has no ability description", c.Name)
		}
	}
}

func TestGetCharacterByID(t *testing.T) {
	c := GetCharacterByID("imp")
	if c == nil {
		t.Fatal("expected to find imp")
	}
	if c.Name != "Imp" {
		t.Errorf("expected Imp, got %s", c.Name)
	}
	if c.Type != CharacterTypeDemon {
		t.Errorf("expected Demon type, got %d", c.Type)
	}

	if GetCharacterByID("nonexistent") != nil {
		t.Error("expected nil for nonexistent character")
	}
}

func TestValidateAssignment5Players(t *testing.T) {
	// 5 players: 3 townsfolk, 0 outsiders, 1 minion, 1 demon
	valid := map[string]string{
		"p1": "washerwoman",
		"p2": "librarian",
		"p3": "investigator",
		"p4": "poisoner",
		"p5": "imp",
	}
	if !ValidateAssignment(valid, 5) {
		t.Error("expected valid assignment for 5 players")
	}

	// Invalid: 2 demons
	invalid := map[string]string{
		"p1": "washerwoman",
		"p2": "librarian",
		"p3": "investigator",
		"p4": "imp",
		"p5": "imp",
	}
	if ValidateAssignment(invalid, 5) {
		t.Error("expected invalid assignment with 2 demons")
	}
}

func TestValidateAssignment10Players(t *testing.T) {
	// 10 players: 7 townsfolk, 0 outsiders, 2 minions, 1 demon
	valid := map[string]string{
		"p1":  "washerwoman",
		"p2":  "librarian",
		"p3":  "investigator",
		"p4":  "chef",
		"p5":  "empath",
		"p6":  "fortuneteller",
		"p7":  "undertaker",
		"p8":  "poisoner",
		"p9":  "spy",
		"p10": "imp",
	}
	if !ValidateAssignment(valid, 10) {
		t.Error("expected valid assignment for 10 players")
	}

	// Invalid: wrong minion count
	invalid := map[string]string{
		"p1":  "washerwoman",
		"p2":  "librarian",
		"p3":  "investigator",
		"p4":  "chef",
		"p5":  "empath",
		"p6":  "fortuneteller",
		"p7":  "undertaker",
		"p8":  "monk",
		"p9":  "poisoner",
		"p10": "imp",
	}
	if ValidateAssignment(invalid, 10) {
		t.Error("expected invalid with 1 minion instead of 2")
	}
}

func TestValidateAssignmentInvalidPlayerCount(t *testing.T) {
	if ValidateAssignment(map[string]string{}, 4) {
		t.Error("expected invalid for 4 players")
	}
	if ValidateAssignment(map[string]string{}, 16) {
		t.Error("expected invalid for 16 players")
	}
}
