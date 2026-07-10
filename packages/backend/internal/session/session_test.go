package session

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/your-org/blood-on-the-clocktower/internal/sessionstore"
)

type memoryStore struct {
	mu          sync.Mutex
	records     map[string]StoreRecord
	failReplace bool
}

func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]StoreRecord)} }
func (s *memoryStore) LoadAll(context.Context) (sessionstore.LoadAllResult, error) {
	return sessionstore.LoadAllResult{}, nil
}
func (s *memoryStore) Create(_ context.Context, record StoreRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[record.RoomID]; ok {
		return errors.New("exists")
	}
	s.records[record.RoomID] = record
	return nil
}
func (s *memoryStore) Replace(_ context.Context, expected uint64, record StoreRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failReplace {
		return errors.New("failed")
	}
	current := s.records[record.RoomID]
	if current.Revision != expected {
		return ErrPersistenceConflict
	}
	s.records[record.RoomID] = record
	return nil
}
func (s *memoryStore) Delete(_ context.Context, roomID string, expected uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.records[roomID]
	if current.Revision != expected {
		return ErrPersistenceConflict
	}
	delete(s.records, roomID)
	return nil
}
func (s *memoryStore) Close() error { return nil }

type testCodec struct{}

func (testCodec) Issue(roomID, playerID string) (string, string, error) {
	return "nonce", roomID + ":" + playerID + ":nonce", nil
}
func (testCodec) Encode(roomID, playerID, nonce string) string {
	return roomID + ":" + playerID + ":" + nonce
}
func (testCodec) Verify(roomID, playerID, nonce, credential string) bool {
	return credential == roomID+":"+playerID+":"+nonce
}

type testEngine struct {
	Players map[string]string `json:"players"`
	Value   int               `json:"value"`
}

func (e *testEngine) Clone() Engine {
	clone := &testEngine{Players: map[string]string{}, Value: e.Value}
	for k, v := range e.Players {
		clone.Players[k] = v
	}
	return clone
}
func (e *testEngine) SetRoomID(string)                {}
func (e *testEngine) AddPlayer(id, name string) error { e.Players[id] = name; return nil }
func (e *testEngine) RemovePlayer(id string)          { delete(e.Players, id) }
func (e *testEngine) Execute(_ string, payload any) (bool, error) {
	delta := payload.(int)
	e.Value += delta
	return delta != 0, nil
}
func (e *testEngine) Project(id string) any {
	return map[string]any{"playerId": id, "value": e.Value, "players": len(e.Players)}
}
func (e *testEngine) Marshal() ([]byte, error) { return json.Marshal(e) }

func newTestSession(t *testing.T) (*AuthoritativeGameSession, *memoryStore, Actor) {
	t.Helper()
	store := newMemoryStore()
	engine := &testEngine{Players: map[string]string{"creator": "Creator"}}
	record := RoomRecord{SchemaVersion: 1, RoomID: "room", RoomRevision: 1, CreatorID: "creator", MaxPlayers: 5, Members: map[string]Identity{"creator": {PlayerID: "creator", Name: "Creator", CredentialNonce: "nonce", State: IdentityMember}}, Retained: map[string]Identity{}, Bans: map[string]bool{}, CreateRequests: map[string]IdempotencyRecord{}, JoinRequests: map[string]IdempotencyRecord{}}
	record.Game, _ = engine.Marshal()
	data, _ := json.Marshal(record)
	_ = store.Create(context.Background(), StoreRecord{RoomID: "room", Revision: 1, Data: data})
	return NewAuthoritativeGameSession(store, testCodec{}, record, engine), store, Actor{PlayerID: "creator", Credential: "room:creator:nonce", ClientSequence: 1}
}

func TestExecutePublishesOnlyAfterPersistence(t *testing.T) {
	session, store, actor := newTestSession(t)
	store.failReplace = true
	_, err := session.Execute(context.Background(), actor, Command{Kind: CommandGame, Fingerprint: "add", Payload: 1})
	if !errors.Is(err, ErrPersistenceUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	query, err := session.Query(Actor{PlayerID: "creator", Credential: actor.Credential})
	if err != nil {
		t.Fatal(err)
	}
	if query.RoomRevision != 1 {
		t.Fatalf("revision advanced to %d", query.RoomRevision)
	}
	if query.Room.(map[string]any)["value"].(int) != 0 {
		t.Fatal("candidate state leaked")
	}
}

func TestExecuteDeduplicatesAcceptedSequence(t *testing.T) {
	session, _, actor := newTestSession(t)
	first, err := session.Execute(context.Background(), actor, Command{Kind: CommandGame, Fingerprint: "add-one", Payload: 1})
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := session.Execute(context.Background(), actor, Command{Kind: CommandGame, Fingerprint: "add-one", Payload: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !duplicate.Duplicate || duplicate.RoomRevision != first.RoomRevision {
		t.Fatal("expected duplicate result")
	}
	query, _ := session.Query(Actor{PlayerID: "creator", Credential: actor.Credential})
	if query.Room.(map[string]any)["value"].(int) != 1 {
		t.Fatal("duplicate executed twice")
	}
}

func TestLeaveRequiresExplicitRejoin(t *testing.T) {
	session, _, actor := newTestSession(t)
	leave, err := session.Execute(context.Background(), actor, Command{Kind: CommandLeave, Fingerprint: "leave"})
	if err != nil {
		t.Fatal(err)
	}
	query, err := session.Query(Actor{PlayerID: "creator", Credential: actor.Credential})
	if err != nil {
		t.Fatal(err)
	}
	if query.Identity == nil || query.Identity.Status != IdentityRetained {
		t.Fatal("expected retained identity")
	}
	actor.ClientSequence = leave.NextClientSequence
	_, err = session.Execute(context.Background(), actor, Command{Kind: CommandRejoin, Fingerprint: "rejoin"})
	if err != nil {
		t.Fatal(err)
	}
	query, _ = session.Query(Actor{PlayerID: "creator", Credential: actor.Credential})
	if query.Room == nil {
		t.Fatal("expected member room projection")
	}
}

func TestNoOpCommandStillCommitsClientSequence(t *testing.T) {
	session, _, actor := newTestSession(t)
	result, err := session.Execute(context.Background(), actor, Command{Kind: CommandGame, Fingerprint: "no-op", Payload: 0})
	if err != nil {
		t.Fatal(err)
	}
	if result.AcceptedSequence != 1 || result.NextClientSequence != 2 || result.RoomRevision != 2 {
		t.Fatalf("unexpected no-op result: %+v", result)
	}
}

func TestClosedSessionRejectsHeldReferences(t *testing.T) {
	session, _, actor := newTestSession(t)
	result, err := session.Execute(context.Background(), actor, Command{Kind: CommandClose, Fingerprint: "close"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ConnectionEffects[0] != CloseAll {
		t.Fatalf("unexpected close result: %+v", result)
	}
	if _, err := session.Query(Actor{PlayerID: "creator", Credential: actor.Credential}); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("query after close returned %v", err)
	}
	if _, err := session.Execute(context.Background(), actor, Command{Kind: CommandGame, Fingerprint: "late", Payload: 1}); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("command after close returned %v", err)
	}
}
