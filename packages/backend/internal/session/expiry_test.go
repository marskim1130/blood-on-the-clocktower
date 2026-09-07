package session

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/sessionstore"
)

func TestCleanupInactiveDeletesExpiredRoomAfterDurableDelete(t *testing.T) {
	s, store, _ := newTestSession(t)
	now := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	s.committed.Load().record.LastActiveAt = now.Add(-7*24*time.Hour - time.Second)
	r := NewRegistry(store, testCodec{}, nil, nil)
	r.sessions["room"], r.creates["request"] = s, "room"
	removed, failures := r.CleanupInactive(context.Background(), now)
	if len(failures) != 0 || len(removed) != 1 || removed[0] != "room" {
		t.Fatalf("cleanup = %v, %v", removed, failures)
	}
	if _, exists := r.Get("room"); exists {
		t.Fatal("expired room remains registered")
	}
	if _, exists := store.records["room"]; exists {
		t.Fatal("expired room remains persisted")
	}
	if r.creates["request"] != "" {
		t.Fatal("create alias survived deletion")
	}
	if _, ok := s.CredentialFor("creator"); ok {
		t.Fatal("expired credentials still accepted")
	}
}

func TestOnlySuccessfulRoomWritesRefreshActivity(t *testing.T) {
	s, store, actor := newTestSession(t)
	old := time.Now().UTC().Add(-8 * 24 * time.Hour)
	s.committed.Load().record.LastActiveAt = old
	store.failReplace = true
	command := Command{Kind: CommandGame, Fingerprint: "activity", Payload: 1}
	if _, err := s.Execute(context.Background(), actor, command); err == nil {
		t.Fatal("expected persistence failure")
	}
	if !s.committed.Load().record.LastActiveAt.Equal(old) {
		t.Fatal("failed command refreshed activity")
	}
	store.failReplace = false
	if _, err := s.Execute(context.Background(), actor, command); err != nil {
		t.Fatal(err)
	}
	committed := s.committed.Load().record.LastActiveAt
	if !committed.After(old) {
		t.Fatal("successful command did not refresh activity")
	}
	if _, err := s.Execute(context.Background(), actor, command); err != nil {
		t.Fatal(err)
	}
	if !s.committed.Load().record.LastActiveAt.Equal(committed) {
		t.Fatal("deduplicated retry refreshed activity")
	}
}

func TestCreatingAndJoiningRoomPersistActivityTime(t *testing.T) {
	store := newMemoryStore()
	r := NewRegistry(store, testCodec{}, nil, nil)
	created, err := r.Create(context.Background(), CreateInput{RequestID: "create", PlayerID: "creator", PlayerName: "Creator", Fingerprint: "create", Engine: &testEngine{Players: map[string]string{}}})
	if err != nil {
		t.Fatal(err)
	}
	s, _ := r.Get(created.RoomID)
	if s.committed.Load().record.LastActiveAt.IsZero() {
		t.Fatal("new room lacks activity timestamp")
	}
	old := time.Now().UTC().Add(-8 * 24 * time.Hour)
	s.committed.Load().record.LastActiveAt = old
	if _, err := s.JoinObserved(context.Background(), JoinInput{RequestID: "join", PlayerID: "new", PlayerName: "New", Fingerprint: "join"}, nil); err != nil {
		t.Fatal(err)
	}
	if !s.committed.Load().record.LastActiveAt.After(old) {
		t.Fatal("joining did not refresh activity")
	}
}

type expiryStore struct {
	*memoryStore
	failDelete bool
}

func (s *expiryStore) Delete(ctx context.Context, id string, revision uint64) error {
	if s.failDelete {
		return errors.New("delete unavailable")
	}
	return s.memoryStore.Delete(ctx, id, revision)
}
func (s *expiryStore) LoadAll(context.Context) (sessionstore.LoadAllResult, error) {
	var result sessionstore.LoadAllResult
	for _, record := range s.records {
		result.Records = append(result.Records, record)
	}
	return result, nil
}
func loadExpiryEngine(data []byte) (Engine, error) {
	e := &testEngine{}
	err := json.Unmarshal(data, e)
	return e, err
}

func TestLegacyRoomGetsPersistedSevenDayGracePeriodAcrossRestarts(t *testing.T) {
	_, memory, _ := newTestSession(t)
	store := &expiryStore{memoryStore: memory}
	r := NewRegistry(store, testCodec{}, loadExpiryEngine, nil)
	if failures, err := r.Restore(context.Background()); err != nil || len(failures) != 0 {
		t.Fatalf("restore = %v, %v", failures, err)
	}
	now := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	removed, failures := r.CleanupInactive(context.Background(), now)
	if len(removed) != 0 || len(failures) != 0 {
		t.Fatalf("legacy cleanup = %v, %v", removed, failures)
	}
	var persisted RoomRecord
	if err := json.Unmarshal(store.records["room"].Data, &persisted); err != nil {
		t.Fatal(err)
	}
	if !persisted.LastActiveAt.Equal(now) {
		t.Fatal("legacy grace timestamp not persisted")
	}
	restored := NewRegistry(store, testCodec{}, loadExpiryEngine, nil)
	if _, err := restored.Restore(context.Background()); err != nil {
		t.Fatal(err)
	}
	removed, failures = restored.CleanupInactive(context.Background(), now.Add(RoomInactivityLimit+time.Second))
	if len(removed) != 1 || len(failures) != 0 {
		t.Fatalf("legacy room did not expire after grace: %v %v", removed, failures)
	}
}

func TestCleanupKeepsActiveBoundaryAndFutureRooms(t *testing.T) {
	now := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	for _, activity := range []time.Time{now, now.Add(-RoomInactivityLimit), now.Add(time.Hour)} {
		s, store, _ := newTestSession(t)
		s.committed.Load().record.LastActiveAt = activity
		r := NewRegistry(store, testCodec{}, nil, nil)
		r.sessions["room"] = s
		removed, failures := r.CleanupInactive(context.Background(), now)
		if len(removed) != 0 || len(failures) != 0 {
			t.Fatalf("active room removed: %v %v", removed, failures)
		}
		if _, exists := r.Get("room"); !exists {
			t.Fatal("active room missing")
		}
	}
}

func TestCleanupPersistenceFailuresKeepRoomAndCredentials(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		s, memory, _ := newTestSession(t)
		now := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
		if !legacy {
			s.committed.Load().record.LastActiveAt = now.Add(-RoomInactivityLimit - time.Second)
		}
		store := &expiryStore{memoryStore: memory, failDelete: !legacy}
		store.failReplace = legacy
		s.store = store
		r := NewRegistry(store, testCodec{}, nil, nil)
		r.sessions["room"] = s
		removed, failures := r.CleanupInactive(context.Background(), now)
		if len(removed) != 0 || len(failures) != 1 {
			t.Fatalf("failure cleanup = %v %v", removed, failures)
		}
		if _, exists := r.Get("room"); !exists {
			t.Fatal("failed deletion removed registry")
		}
		if _, exists := store.records["room"]; !exists {
			t.Fatal("failed deletion removed data")
		}
		if _, ok := s.CredentialFor("creator"); !ok {
			t.Fatal("failed deletion revoked credentials")
		}
		if legacy && !s.committed.Load().record.LastActiveAt.IsZero() {
			t.Fatal("failed migration published activity")
		}
	}
}

func TestCleanupCASConflictDoesNotRemoveRoom(t *testing.T) {
	s, memory, _ := newTestSession(t)
	now := time.Now().UTC()
	s.committed.Load().record.LastActiveAt = now.Add(-RoomInactivityLimit - time.Second)
	record := memory.records["room"]
	record.Revision++
	memory.records["room"] = record
	r := NewRegistry(memory, testCodec{}, nil, nil)
	r.sessions["room"] = s
	removed, failures := r.CleanupInactive(context.Background(), now)
	if len(removed) != 0 || len(failures) != 1 {
		t.Fatalf("CAS cleanup = %v %v", removed, failures)
	}
	if _, exists := r.Get("room"); !exists {
		t.Fatal("CAS conflict removed registry")
	}
}
