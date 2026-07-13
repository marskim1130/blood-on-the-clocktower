package session

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/your-org/blood-on-the-clocktower/internal/sessionstore"
)

type committedView struct {
	record RoomRecord
	engine Engine
	closed bool
}

type AuthoritativeGameSession struct {
	commandMu sync.Mutex
	store     Store
	codec     CredentialCodec
	committed atomic.Pointer[committedView]
	conflict  atomic.Bool
}

func NewAuthoritativeGameSession(store Store, codec CredentialCodec, record RoomRecord, engine Engine) *AuthoritativeGameSession {
	s := &AuthoritativeGameSession{store: store, codec: codec}
	s.committed.Store(&committedView{record: cloneRecord(record), engine: engine})
	return s
}

func (s *AuthoritativeGameSession) RoomID() string { return s.committed.Load().record.RoomID }

func (s *AuthoritativeGameSession) Metadata() RoomMetadata {
	record := s.committed.Load().record
	return RoomMetadata{RoomID: record.RoomID, CreatorID: record.CreatorID, MaxPlayers: record.MaxPlayers, ScriptID: record.ScriptID, ParticipantSetFrozen: record.ParticipantSetFrozen}
}

func (s *AuthoritativeGameSession) Healthy() bool { return !s.conflict.Load() }

func (s *AuthoritativeGameSession) CredentialFor(playerID string) (string, bool) {
	view := s.committed.Load()
	if view.closed {
		return "", false
	}
	identity, ok := view.record.Members[playerID]
	if !ok {
		return "", false
	}
	return s.codec.Encode(view.record.RoomID, playerID, identity.CredentialNonce), true
}

func (s *AuthoritativeGameSession) Execute(ctx context.Context, actor Actor, command Command) (CommandResult, error) {
	return s.ExecuteObserved(ctx, actor, command, nil)
}

func (s *AuthoritativeGameSession) ExecuteObserved(ctx context.Context, actor Actor, command Command, observer CommitObserver) (CommandResult, error) {
	s.commandMu.Lock()
	defer s.commandMu.Unlock()
	if s.conflict.Load() {
		return CommandResult{}, ErrPersistenceConflict
	}
	current := s.committed.Load()
	if current.closed {
		return CommandResult{}, ErrRoomNotFound
	}
	candidate := cloneRecord(current.record)
	engine := current.engine.Clone()
	identity, retained, err := authenticate(candidate, s.codec, actor)
	if err != nil {
		return CommandResult{}, err
	}
	if actor.ClientSequence == identity.LastSequence && identity.LastSequence != 0 {
		if command.Fingerprint != identity.LastFingerprint {
			return CommandResult{}, ErrSequenceConflict
		}
		result := CommandResult{AcceptedSequence: identity.LastResult.AcceptedSequence, NextClientSequence: identity.LastResult.NextClientSequence, RoomRevision: current.record.RoomRevision, Duplicate: true, DirectResponse: project(current, actor.PlayerID, retained), Metadata: metadataFromRecord(current.record)}
		if observer != nil {
			observer(result)
		}
		return result, nil
	}
	if actor.ClientSequence != identity.LastSequence+1 {
		return CommandResult{}, ErrUnexpectedSequence
	}

	result := CommandResult{AcceptedSequence: actor.ClientSequence, NextClientSequence: actor.ClientSequence + 1, TargetPlayerID: command.TargetID}
	switch command.Kind {
	case CommandRejoin:
		if !retained || candidate.ParticipantSetFrozen {
			return CommandResult{}, ErrParticipantSetFrozen
		}
		delete(candidate.Retained, actor.PlayerID)
		identity.State = IdentityMember
		candidate.Members[actor.PlayerID] = identity
		if err := engine.AddPlayer(identity.PlayerID, identity.Name); err != nil {
			return CommandResult{}, err
		}
	case CommandLeave:
		if retained || candidate.ParticipantSetFrozen {
			return CommandResult{}, ErrParticipantSetFrozen
		}
		delete(candidate.Members, actor.PlayerID)
		identity.State = IdentityRetained
		candidate.Retained[actor.PlayerID] = identity
		engine.RemovePlayer(actor.PlayerID)
		result.ConnectionEffects = []ConnectionEffect{CloseSender}
	case CommandKick:
		if actor.PlayerID != candidate.CreatorID || command.TargetID == candidate.CreatorID || candidate.ParticipantSetFrozen {
			return CommandResult{}, ErrForbidden
		}
		if _, ok := candidate.Members[command.TargetID]; !ok {
			return CommandResult{}, ErrInvalidCredential
		}
		delete(candidate.Members, command.TargetID)
		delete(candidate.Retained, command.TargetID)
		candidate.Bans[command.TargetID] = true
		engine.RemovePlayer(command.TargetID)
		result.ConnectionEffects = []ConnectionEffect{CloseTarget}
	case CommandUpdateSettings:
		if actor.PlayerID != candidate.CreatorID || retained || candidate.ParticipantSetFrozen {
			return CommandResult{}, ErrForbidden
		}
		if command.MaxPlayers > 0 {
			candidate.MaxPlayers = command.MaxPlayers
		}
		if command.ScriptID != "" {
			candidate.ScriptID = command.ScriptID
		}
		if command.Payload != nil {
			if _, err := engine.Execute(actor.PlayerID, command.Payload); err != nil {
				return CommandResult{}, err
			}
		}
	case CommandClose:
		if actor.PlayerID != candidate.CreatorID {
			return CommandResult{}, ErrForbidden
		}
		if err := s.store.Delete(ctx, candidate.RoomID, candidate.RoomRevision); err != nil {
			if errors.Is(err, sessionstore.ErrRevisionConflict) {
				s.conflict.Store(true)
				return CommandResult{}, ErrPersistenceConflict
			}
			return CommandResult{}, ErrPersistenceUnavailable
		}
		s.committed.Store(&committedView{record: candidate, engine: engine, closed: true})
		result.RoomRevision = candidate.RoomRevision
		result.Metadata = metadataFromRecord(candidate)
		result.ConnectionEffects = []ConnectionEffect{CloseAll}
		if observer != nil {
			observer(result)
		}
		return result, nil
	default:
		if retained {
			return CommandResult{}, ErrForbidden
		}
		updated, err := engine.Execute(actor.PlayerID, command.Payload)
		if err != nil {
			return CommandResult{}, err
		}
		_ = updated
		if command.FreezeParticipants {
			candidate.ParticipantSetFrozen = true
		}
	}

	identity.LastSequence = actor.ClientSequence
	identity.LastFingerprint = command.Fingerprint
	candidate.RoomRevision++
	identity.LastResult = ResultSummary{RoomRevision: candidate.RoomRevision, AcceptedSequence: actor.ClientSequence, NextClientSequence: actor.ClientSequence + 1}
	if identity.State == IdentityRetained {
		candidate.Retained[actor.PlayerID] = identity
	} else {
		candidate.Members[actor.PlayerID] = identity
	}
	gameData, err := engine.Marshal()
	if err != nil {
		return CommandResult{}, err
	}
	candidate.Game = gameData
	data, err := json.Marshal(candidate)
	if err != nil {
		return CommandResult{}, err
	}
	if err := s.store.Replace(ctx, current.record.RoomRevision, StoreRecord{RoomID: candidate.RoomID, Revision: candidate.RoomRevision, Data: data}); err != nil {
		if errors.Is(err, sessionstore.ErrRevisionConflict) {
			s.conflict.Store(true)
			return CommandResult{}, ErrPersistenceConflict
		}
		return CommandResult{}, ErrPersistenceUnavailable
	}
	view := &committedView{record: candidate, engine: engine}
	s.committed.Store(view)
	result.RoomRevision = candidate.RoomRevision
	result.Metadata = metadataFromRecord(candidate)
	result.DirectResponse = project(view, actor.PlayerID, identity.State == IdentityRetained)
	for playerID := range candidate.Members {
		result.Deliveries = append(result.Deliveries, Delivery{PlayerID: playerID, Payload: engine.Project(playerID)})
	}
	if observer != nil {
		observer(result)
	}
	return result, nil
}

func (s *AuthoritativeGameSession) Query(actor Actor) (QueryResult, error) {
	view := s.committed.Load()
	if view.closed {
		return QueryResult{}, ErrRoomNotFound
	}
	identity, retained, err := authenticate(view.record, s.codec, actor)
	if err != nil {
		return QueryResult{}, err
	}
	result := QueryResult{RoomRevision: view.record.RoomRevision, NextClientSequence: identity.LastSequence + 1, Metadata: metadataFromRecord(view.record)}
	if retained {
		result.Identity = &IdentityStatus{Status: IdentityRetained, CanRejoin: !view.record.ParticipantSetFrozen, NextClientSequence: identity.LastSequence + 1, ParticipantSetFrozen: view.record.ParticipantSetFrozen}
	} else {
		result.Room = view.engine.Project(actor.PlayerID)
	}
	return result, nil
}

func metadataFromRecord(record RoomRecord) RoomMetadata {
	return RoomMetadata{RoomID: record.RoomID, CreatorID: record.CreatorID, MaxPlayers: record.MaxPlayers, ScriptID: record.ScriptID, ParticipantSetFrozen: record.ParticipantSetFrozen}
}

func authenticate(record RoomRecord, codec CredentialCodec, actor Actor) (Identity, bool, error) {
	if identity, ok := record.Members[actor.PlayerID]; ok && codec.Verify(record.RoomID, actor.PlayerID, identity.CredentialNonce, actor.Credential) {
		return identity, false, nil
	}
	if identity, ok := record.Retained[actor.PlayerID]; ok && codec.Verify(record.RoomID, actor.PlayerID, identity.CredentialNonce, actor.Credential) {
		return identity, true, nil
	}
	return Identity{}, false, ErrInvalidCredential
}

func project(view *committedView, playerID string, retained bool) any {
	if retained {
		identity := view.record.Retained[playerID]
		return IdentityStatus{Status: IdentityRetained, CanRejoin: !view.record.ParticipantSetFrozen, NextClientSequence: identity.LastSequence + 1, ParticipantSetFrozen: view.record.ParticipantSetFrozen}
	}
	return view.engine.Project(playerID)
}

func cloneRecord(record RoomRecord) RoomRecord {
	data, _ := json.Marshal(record)
	var clone RoomRecord
	_ = json.Unmarshal(data, &clone)
	if clone.Members == nil {
		clone.Members = make(map[string]Identity)
	}
	if clone.Retained == nil {
		clone.Retained = make(map[string]Identity)
	}
	if clone.Bans == nil {
		clone.Bans = make(map[string]bool)
	}
	if clone.CreateRequests == nil {
		clone.CreateRequests = make(map[string]IdempotencyRecord)
	}
	if clone.JoinRequests == nil {
		clone.JoinRequests = make(map[string]IdempotencyRecord)
	}
	return clone
}
