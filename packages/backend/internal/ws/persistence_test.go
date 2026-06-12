package ws

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

type fakeRedisSnapshotClient struct {
	values map[string]string
	closed bool
}

func newFakeRedisSnapshotClient() *fakeRedisSnapshotClient {
	return &fakeRedisSnapshotClient{values: make(map[string]string)}
}

func (c *fakeRedisSnapshotClient) Get(_ context.Context, key string) (string, error) {
	value, exists := c.values[key]
	if !exists {
		return "", redis.Nil
	}
	return value, nil
}

func (c *fakeRedisSnapshotClient) Set(_ context.Context, key string, value string) error {
	c.values[key] = value
	return nil
}

func (c *fakeRedisSnapshotClient) Close() error {
	c.closed = true
	return nil
}

func TestRedisSnapshotStoreReturnsEmptySnapshotForMissingKey(t *testing.T) {
	store := newRedisSnapshotStoreWithClient(newFakeRedisSnapshotClient(), "clocktower:test")

	snapshot, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if snapshot.Version != snapshotVersion {
		t.Fatalf("expected snapshot version %d, got %d", snapshotVersion, snapshot.Version)
	}
	if len(snapshot.Rooms) != 0 {
		t.Fatalf("expected no rooms for missing Redis key, got %#v", snapshot.Rooms)
	}
	if snapshot.Sessions == nil {
		t.Fatal("expected sessions map to be initialized")
	}
}

func TestRedisSnapshotStoreSavesAndLoadsSnapshot(t *testing.T) {
	client := newFakeRedisSnapshotClient()
	store := newRedisSnapshotStoreWithClient(client, "clocktower:test")
	original := &hubSnapshot{
		Version: snapshotVersion,
		Rooms: []roomSnapshot{
			{
				ID:              "123456",
				MaxPlayers:      5,
				CreatorID:       "storyteller",
				ScriptID:        game.TroubleBrewingScriptID,
				KickedPlayerIDs: []string{"p9"},
			},
		},
		Sessions: map[string]gameSessionSnapshot{
			"123456": {
				Players: []game.Player{
					{ID: "p1", Name: "Alice", IsAlive: false},
				},
				StorytellerID:  "storyteller",
				ScriptID:       game.TroubleBrewingScriptID,
				Phase:          game.GamePhaseDay,
				DayNumber:      2,
				GhostVotesUsed: map[string]bool{"p1": true},
				SlayerUsed:     map[string]bool{"p2": true},
				Deaths: []game.DeathRecord{
					{PlayerID: "p1", Cause: game.DeathCauseNightKill, DayNumber: 1},
				},
			},
		},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}

	if len(loaded.Rooms) != 1 || loaded.Rooms[0].ID != "123456" {
		t.Fatalf("expected restored room, got %#v", loaded.Rooms)
	}
	if len(loaded.Rooms[0].KickedPlayerIDs) != 1 || loaded.Rooms[0].KickedPlayerIDs[0] != "p9" {
		t.Fatalf("expected restored kicked player ids, got %#v", loaded.Rooms[0].KickedPlayerIDs)
	}
	session := loaded.Sessions["123456"]
	if session.Phase != game.GamePhaseDay || session.DayNumber != 2 {
		t.Fatalf("expected restored day 2 session, got %#v", session)
	}
	if !session.GhostVotesUsed["p1"] {
		t.Fatalf("expected restored ghost vote usage, got %#v", session.GhostVotesUsed)
	}
	if !session.SlayerUsed["p2"] {
		t.Fatalf("expected restored Slayer ability usage, got %#v", session.SlayerUsed)
	}
	if len(session.Deaths) != 1 || session.Deaths[0].PlayerID != "p1" {
		t.Fatalf("expected restored death record, got %#v", session.Deaths)
	}
}

func TestRedisSnapshotStoreUsesDefaultKeyAndClosesClient(t *testing.T) {
	client := newFakeRedisSnapshotClient()
	store := newRedisSnapshotStoreWithClient(client, "")

	if err := store.Save(emptyHubSnapshot()); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	if _, exists := client.values[defaultRedisSnapshotKey]; !exists {
		t.Fatalf("expected save to use default Redis key %q", defaultRedisSnapshotKey)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !client.closed {
		t.Fatal("expected Redis client to be closed")
	}
}

func TestNewRedisSnapshotStoreRejectsInvalidURL(t *testing.T) {
	_, err := NewRedisSnapshotStore("://bad-url", "clocktower:test")
	if err == nil {
		t.Fatal("expected invalid Redis URL to be rejected")
	}
}
