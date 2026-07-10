package sessionstore

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]Record
	closed  bool
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{records: make(map[string]Record)} }

func (s *MemoryStore) LoadAll(ctx context.Context) (LoadAllResult, error) {
	if err := ctx.Err(); err != nil {
		return LoadAllResult{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return LoadAllResult{}, ErrClosed
	}
	roomIDs := make([]string, 0, len(s.records))
	for roomID := range s.records {
		roomIDs = append(roomIDs, roomID)
	}
	sort.Strings(roomIDs)
	result := LoadAllResult{Records: make([]Record, 0, len(roomIDs))}
	for _, roomID := range roomIDs {
		result.Records = append(result.Records, cloneRecord(s.records[roomID]))
	}
	return result, nil
}

func (s *MemoryStore) Create(ctx context.Context, record Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateRecord(record); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	if _, exists := s.records[record.RoomID]; exists {
		return fmt.Errorf("%w: %s", ErrRecordExists, record.RoomID)
	}
	s.records[record.RoomID] = cloneRecord(record)
	return nil
}

func (s *MemoryStore) Replace(ctx context.Context, expectedRevision uint64, record Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateRecord(record); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	current, exists := s.records[record.RoomID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrRecordNotFound, record.RoomID)
	}
	if current.Revision != expectedRevision {
		return &RevisionConflictError{RoomID: record.RoomID, Expected: expectedRevision, Actual: current.Revision}
	}
	s.records[record.RoomID] = cloneRecord(record)
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, roomID string, expectedRevision uint64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateRoomID(roomID); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	current, exists := s.records[roomID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrRecordNotFound, roomID)
	}
	if current.Revision != expectedRevision {
		return &RevisionConflictError{RoomID: roomID, Expected: expectedRevision, Actual: current.Revision}
	}
	delete(s.records, roomID)
	return nil
}

func (s *MemoryStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}
