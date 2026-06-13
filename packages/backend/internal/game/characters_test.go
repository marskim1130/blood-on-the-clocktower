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

func TestGetScriptByID(t *testing.T) {
	script := GetScriptByID(TroubleBrewingScriptID)
	if script == nil {
		t.Fatal("expected to find Trouble Brewing script")
	}
	if script.Name != "Trouble Brewing" {
		t.Errorf("expected Trouble Brewing, got %s", script.Name)
	}
	if len(script.Characters) != 22 {
		t.Errorf("expected 22 characters, got %d", len(script.Characters))
	}

	if GetScriptByID("bad_moon_rising") != nil {
		t.Error("expected nil for unsupported script")
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

func TestValidateAssignmentRejectsDuplicateCharacters(t *testing.T) {
	duplicate := map[string]string{
		"p1": "washerwoman",
		"p2": "washerwoman",
		"p3": "investigator",
		"p4": "poisoner",
		"p5": "imp",
	}

	if ValidateAssignment(duplicate, 5) {
		t.Error("expected invalid assignment with duplicate character")
	}
}

func TestValidateAssignmentWithBaronSetupModifier(t *testing.T) {
	validWithBaron := map[string]string{
		"p1": "washerwoman",
		"p2": "librarian",
		"p3": "investigator",
		"p4": "butler",
		"p5": "saint",
		"p6": "baron",
		"p7": "imp",
	}
	if !ValidateAssignment(validWithBaron, 7) {
		t.Error("expected valid Baron assignment with two extra Outsiders")
	}

	invalidBaseCountWithBaron := map[string]string{
		"p1": "washerwoman",
		"p2": "librarian",
		"p3": "investigator",
		"p4": "chef",
		"p5": "empath",
		"p6": "baron",
		"p7": "imp",
	}
	if ValidateAssignment(invalidBaseCountWithBaron, 7) {
		t.Error("expected invalid Baron assignment without extra Outsiders")
	}
}

func TestGetScriptWakeOrderUsesCanonicalFortuneTellerID(t *testing.T) {
	firstNight := GetScriptWakeOrder(TroubleBrewingScriptID, 1)
	subsequentNight := GetScriptWakeOrder(TroubleBrewingScriptID, 2)

	if !nightWakeOrderContains(firstNight, "fortuneteller") {
		t.Fatal("expected first night to contain canonical fortuneteller id")
	}
	if !nightWakeOrderContains(subsequentNight, "fortuneteller") {
		t.Fatal("expected subsequent night to contain canonical fortuneteller id")
	}
	if nightWakeOrderContains(firstNight, "fortune_teller") {
		t.Fatal("first night should not contain deprecated fortune_teller id")
	}
}

func TestGetActiveNightWakeStepsFiltersToAliveAssignedCharacters(t *testing.T) {
	players := []Player{
		{ID: "p1", IsAlive: true, Character: &Character{ID: "washerwoman", Name: "Washerwoman", Team: TeamGood}},
		{ID: "p2", IsAlive: false, Character: &Character{ID: "poisoner", Name: "Poisoner", Team: TeamEvil}},
		{ID: "p3", IsAlive: true, Character: &Character{ID: "imp", Name: "Imp", Team: TeamEvil}},
	}

	active := GetActiveNightWakeSteps(TroubleBrewingScriptID, 1, players)
	if len(active) != 3 {
		t.Fatalf("expected 3 active wake steps, got %#v", active)
	}
	if active[0].CharacterType != NightWakeCharacterTypeDemon ||
		active[1].CharacterID != "washerwoman" ||
		active[2].CharacterID != "imp" {
		t.Fatalf("expected demon info, washerwoman, then imp, got %#v", active)
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

func nightWakeOrderContains(steps []NightWakeStep, characterID string) bool {
	for _, step := range steps {
		if step.CharacterID == characterID {
			return true
		}
	}
	return false
}

func TestValidateAssignmentInvalidPlayerCount(t *testing.T) {
	if ValidateAssignment(map[string]string{}, 4) {
		t.Error("expected invalid for 4 players")
	}
	if ValidateAssignment(map[string]string{}, 16) {
		t.Error("expected invalid for 16 players")
	}
}
