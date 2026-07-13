package ws

import "testing"

func TestActiveConnectionRegistryTakeoverInvalidatesPreviousConnection(t *testing.T) {
	registry := NewActiveConnectionRegistry()
	first := newFakeConnection()
	second := newFakeConnection()

	firstGeneration, previous := registry.Takeover("room", "player", first)
	if previous != nil || !registry.IsCurrent("room", "player", first, firstGeneration) {
		t.Fatal("first connection should become current")
	}

	secondGeneration, previous := registry.Takeover("room", "player", second)
	if previous != first {
		t.Fatal("takeover should return the previous connection")
	}
	if registry.IsCurrent("room", "player", first, firstGeneration) {
		t.Fatal("previous connection should be stale")
	}
	if !registry.IsCurrent("room", "player", second, secondGeneration) {
		t.Fatal("new connection should be current")
	}
}

func TestActiveConnectionRegistryRemoveDoesNotRemoveReplacement(t *testing.T) {
	registry := NewActiveConnectionRegistry()
	first := newFakeConnection()
	second := newFakeConnection()
	registry.Takeover("room", "player", first)
	generation, _ := registry.Takeover("room", "player", second)

	registry.Remove(first)
	if !registry.IsCurrent("room", "player", second, generation) {
		t.Fatal("removing a stale connection must not remove its replacement")
	}
}
