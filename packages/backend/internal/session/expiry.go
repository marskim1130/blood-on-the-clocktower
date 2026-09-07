package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/sessionstore"
)

const RoomInactivityLimit = 7 * 24 * time.Hour

// CleanupInactive serializes expiry with room commands and removes registry entries
// only after the persisted room has been deleted with its current revision.
func (r *Registry) CleanupInactive(ctx context.Context, now time.Time) ([]string, []error) {
	r.createMu.Lock()
	defer r.createMu.Unlock()
	r.mu.RLock()
	sessions := make([]*AuthoritativeGameSession, 0, len(r.sessions))
	for _, s := range r.sessions {
		sessions = append(sessions, s)
	}
	r.mu.RUnlock()
	var removed []string
	var failures []error
	for _, s := range sessions {
		expired, err := s.expireInactive(ctx, now)
		if err != nil {
			failures = append(failures, fmt.Errorf("room %s expiry: %w", s.RoomID(), err))
			continue
		}
		if expired {
			r.Remove(s.RoomID())
			removed = append(removed, s.RoomID())
		}
	}
	return removed, failures
}

func (s *AuthoritativeGameSession) expireInactive(ctx context.Context, now time.Time) (bool, error) {
	s.commandMu.Lock()
	defer s.commandMu.Unlock()
	if s.conflict.Load() {
		return false, ErrPersistenceConflict
	}
	current := s.committed.Load()
	if current.closed {
		return false, nil
	}
	if current.record.LastActiveAt.IsZero() {
		candidate := cloneRecord(current.record)
		candidate.LastActiveAt = now.UTC()
		candidate.RoomRevision++
		data, err := json.Marshal(candidate)
		if err != nil {
			return false, err
		}
		if err := s.store.Replace(ctx, current.record.RoomRevision, StoreRecord{RoomID: candidate.RoomID, Revision: candidate.RoomRevision, Data: data}); err != nil {
			if errors.Is(err, sessionstore.ErrRevisionConflict) {
				s.conflict.Store(true)
				return false, ErrPersistenceConflict
			}
			return false, ErrPersistenceUnavailable
		}
		s.committed.Store(&committedView{record: candidate, engine: current.engine})
		return false, nil
	}
	if !current.record.LastActiveAt.Before(now.Add(-RoomInactivityLimit)) {
		return false, nil
	}
	if err := s.store.Delete(ctx, current.record.RoomID, current.record.RoomRevision); err != nil {
		if errors.Is(err, sessionstore.ErrRevisionConflict) {
			s.conflict.Store(true)
			return false, ErrPersistenceConflict
		}
		return false, ErrPersistenceUnavailable
	}
	s.committed.Store(&committedView{record: current.record, engine: current.engine, closed: true})
	return true, nil
}
