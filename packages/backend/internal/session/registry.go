package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
)

type randomIDGenerator struct{}

func (randomIDGenerator) RoomID() (string, error) {
	data := make([]byte, 18)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

type Registry struct {
	mu       sync.RWMutex
	createMu sync.Mutex
	store    Store
	codec    CredentialCodec
	ids      IDGenerator
	loader   EngineLoader
	sessions map[string]*AuthoritativeGameSession
	creates  map[string]string
}

func NewRegistry(store Store, codec CredentialCodec, loader EngineLoader, ids IDGenerator) *Registry {
	if ids == nil {
		ids = randomIDGenerator{}
	}
	return &Registry{store: store, codec: codec, loader: loader, ids: ids, sessions: make(map[string]*AuthoritativeGameSession), creates: make(map[string]string)}
}

func (r *Registry) Restore(ctx context.Context) ([]error, error) {
	loaded, err := r.store.LoadAll(ctx)
	if err != nil {
		return nil, err
	}
	errorsByRoom := make([]error, 0, len(loaded.Errors))
	for _, loadErr := range loaded.Errors {
		errorsByRoom = append(errorsByRoom, loadErr)
	}
	for _, stored := range loaded.Records {
		var record RoomRecord
		if err := json.Unmarshal(stored.Data, &record); err != nil {
			errorsByRoom = append(errorsByRoom, fmt.Errorf("room %s: %w", stored.RoomID, err))
			continue
		}
		if record.SchemaVersion != SchemaVersion || record.RoomID != stored.RoomID || record.RoomRevision != stored.Revision {
			errorsByRoom = append(errorsByRoom, fmt.Errorf("room %s: unsupported or inconsistent record", stored.RoomID))
			continue
		}
		engine, err := r.loader(record.Game)
		if err != nil {
			errorsByRoom = append(errorsByRoom, fmt.Errorf("room %s: %w", stored.RoomID, err))
			continue
		}
		engine.SetRoomID(record.RoomID)
		r.sessions[record.RoomID] = NewAuthoritativeGameSession(r.store, r.codec, record, engine)
		for requestID := range record.CreateRequests {
			if existing := r.creates[requestID]; existing != "" {
				delete(r.sessions, record.RoomID)
				delete(r.sessions, existing)
				errorsByRoom = append(errorsByRoom, fmt.Errorf("duplicate create request %s", requestID))
				continue
			}
			r.creates[requestID] = record.RoomID
		}
	}
	return errorsByRoom, nil
}

type CreateInput struct {
	RequestID, PlayerID, PlayerName, ScriptID, Fingerprint string
	MaxPlayers                                             int
	Engine                                                 Engine
}
type CreateResult struct {
	RoomID, PlayerID, ResumeCredential string
	RoomRevision, NextClientSequence   uint64
	State                              any
	Metadata                           RoomMetadata
}
type JoinInput struct{ RequestID, RoomID, PlayerID, PlayerName, Fingerprint string }
type JoinResult struct {
	PlayerID, ResumeCredential       string
	RoomRevision, NextClientSequence uint64
	State                            any
	Deliveries                       []Delivery
	Metadata                         RoomMetadata
}

func (r *Registry) Create(ctx context.Context, input CreateInput) (CreateResult, error) {
	r.createMu.Lock()
	defer r.createMu.Unlock()
	r.mu.RLock()
	if roomID := r.creates[input.RequestID]; roomID != "" {
		s := r.sessions[roomID]
		r.mu.RUnlock()
		view := s.committed.Load()
		record := view.record.CreateRequests[input.RequestID]
		if record.Fingerprint != input.Fingerprint {
			return CreateResult{}, ErrIdempotencyConflict
		}
		identity := view.record.Members[record.PlayerID]
		return CreateResult{RoomID: roomID, PlayerID: record.PlayerID, ResumeCredential: r.codec.Encode(roomID, record.PlayerID, identity.CredentialNonce), RoomRevision: view.record.RoomRevision, NextClientSequence: identity.LastSequence + 1, State: view.engine.Project(record.PlayerID), Metadata: metadataFromRecord(view.record)}, nil
	}
	r.mu.RUnlock()

	for attempts := 0; attempts < 8; attempts++ {
		roomID, err := r.ids.RoomID()
		if err != nil {
			return CreateResult{}, err
		}
		input.Engine.SetRoomID(roomID)
		nonce, credential, err := r.codec.Issue(roomID, input.PlayerID)
		if err != nil {
			return CreateResult{}, err
		}
		maxPlayers := input.MaxPlayers
		if maxPlayers == 0 {
			maxPlayers = 5
		}
		record := RoomRecord{SchemaVersion: SchemaVersion, RoomID: roomID, RoomRevision: 1, CreatorID: input.PlayerID, MaxPlayers: maxPlayers, ScriptID: input.ScriptID, Members: map[string]Identity{input.PlayerID: {PlayerID: input.PlayerID, Name: input.PlayerName, CredentialNonce: nonce, State: IdentityMember}}, Retained: map[string]Identity{}, Bans: map[string]bool{}, CreateRequests: map[string]IdempotencyRecord{input.RequestID: {Fingerprint: input.Fingerprint, PlayerID: input.PlayerID}}, JoinRequests: map[string]IdempotencyRecord{}}
		gameData, err := input.Engine.Marshal()
		if err != nil {
			return CreateResult{}, err
		}
		record.Game = gameData
		data, _ := json.Marshal(record)
		if err := r.store.Create(ctx, StoreRecord{RoomID: roomID, Revision: 1, Data: data}); err != nil {
			continue
		}
		s := NewAuthoritativeGameSession(r.store, r.codec, record, input.Engine)
		r.mu.Lock()
		r.sessions[roomID] = s
		r.creates[input.RequestID] = roomID
		r.mu.Unlock()
		return CreateResult{RoomID: roomID, PlayerID: input.PlayerID, ResumeCredential: credential, RoomRevision: 1, NextClientSequence: 1, State: input.Engine.Project(input.PlayerID), Metadata: metadataFromRecord(record)}, nil
	}
	return CreateResult{}, fmt.Errorf("create room: %w", ErrPersistenceUnavailable)
}

func (r *Registry) Get(roomID string) (*AuthoritativeGameSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[roomID]
	return s, ok
}

func (r *Registry) Healthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, activeSession := range r.sessions {
		if !activeSession.Healthy() {
			return false
		}
	}
	return true
}

func (r *Registry) Join(ctx context.Context, input JoinInput) (JoinResult, error) {
	return r.JoinObserved(ctx, input, nil)
}

func (r *Registry) JoinObserved(ctx context.Context, input JoinInput, observer JoinObserver) (JoinResult, error) {
	s, ok := r.Get(input.RoomID)
	if !ok {
		return JoinResult{}, ErrRoomNotFound
	}
	s.commandMu.Lock()
	defer s.commandMu.Unlock()
	current := s.committed.Load()
	if existing, ok := current.record.JoinRequests[input.RequestID]; ok {
		if existing.Fingerprint != input.Fingerprint {
			return JoinResult{}, ErrIdempotencyConflict
		}
		identity := current.record.Members[existing.PlayerID]
		result := JoinResult{PlayerID: identity.PlayerID, ResumeCredential: r.codec.Encode(input.RoomID, identity.PlayerID, identity.CredentialNonce), RoomRevision: current.record.RoomRevision, NextClientSequence: identity.LastSequence + 1, State: current.engine.Project(identity.PlayerID), Metadata: metadataFromRecord(current.record)}
		if observer != nil {
			observer(result)
		}
		return result, nil
	}
	if current.record.ParticipantSetFrozen {
		return JoinResult{}, ErrParticipantSetFrozen
	}
	if current.record.Bans[input.PlayerID] {
		return JoinResult{}, ErrInvalidCredential
	}
	if _, exists := current.record.Members[input.PlayerID]; exists {
		return JoinResult{}, ErrInvalidCredential
	}
	if _, exists := current.record.Retained[input.PlayerID]; exists {
		return JoinResult{}, ErrInvalidCredential
	}
	if len(current.record.Members) >= current.record.MaxPlayers+1 {
		return JoinResult{}, ErrForbidden
	}

	candidate := cloneRecord(current.record)
	engine := current.engine.Clone()
	if err := engine.AddPlayer(input.PlayerID, input.PlayerName); err != nil {
		return JoinResult{}, err
	}
	nonce, credential, err := r.codec.Issue(input.RoomID, input.PlayerID)
	if err != nil {
		return JoinResult{}, err
	}
	candidate.Members[input.PlayerID] = Identity{PlayerID: input.PlayerID, Name: input.PlayerName, CredentialNonce: nonce, State: IdentityMember}
	candidate.JoinRequests[input.RequestID] = IdempotencyRecord{Fingerprint: input.Fingerprint, PlayerID: input.PlayerID}
	candidate.RoomRevision++
	gameData, err := engine.Marshal()
	if err != nil {
		return JoinResult{}, err
	}
	candidate.Game = gameData
	data, err := json.Marshal(candidate)
	if err != nil {
		return JoinResult{}, err
	}
	if err := r.store.Replace(ctx, current.record.RoomRevision, StoreRecord{RoomID: input.RoomID, Revision: candidate.RoomRevision, Data: data}); err != nil {
		return JoinResult{}, ErrPersistenceUnavailable
	}
	s.committed.Store(&committedView{record: candidate, engine: engine})
	result := JoinResult{PlayerID: input.PlayerID, ResumeCredential: credential, RoomRevision: candidate.RoomRevision, NextClientSequence: 1, State: engine.Project(input.PlayerID), Metadata: metadataFromRecord(candidate)}
	for playerID := range candidate.Members {
		result.Deliveries = append(result.Deliveries, Delivery{PlayerID: playerID, Payload: engine.Project(playerID)})
	}
	if observer != nil {
		observer(result)
	}
	return result, nil
}

func (r *Registry) Remove(roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, roomID)
	for requestID, id := range r.creates {
		if id == roomID {
			delete(r.creates, requestID)
		}
	}
}
