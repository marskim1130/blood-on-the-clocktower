package sessionstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type storedRecord struct {
	RoomID   string `json:"roomId"`
	Revision uint64 `json:"revision,string"`
	Data     []byte `json:"data"`
}

type FileStore struct {
	directory string
	mu        sync.RWMutex
	closed    bool
}

func NewFileStore(directory string) (*FileStore, error) {
	if directory == "" {
		return nil, fmt.Errorf("%w: storage directory is required", ErrInvalidRecord)
	}
	info, err := os.Stat(directory)
	if err == nil && !info.IsDir() {
		return nil, fmt.Errorf("session store path %q is not a directory", directory)
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{directory: directory}, nil
}

func (s *FileStore) LoadAll(ctx context.Context) (LoadAllResult, error) {
	if err := ctx.Err(); err != nil {
		return LoadAllResult{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return LoadAllResult{}, ErrClosed
	}
	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return LoadAllResult{}, err
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Name() < entries[right].Name() })
	result := LoadAllResult{}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return LoadAllResult{}, err
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(s.directory, entry.Name())
		record, err := readStoredRecord(path)
		if err != nil {
			result.Errors = append(result.Errors, LoadError{RoomID: strings.TrimSuffix(entry.Name(), ".json"), Source: path, Err: err})
			continue
		}
		if entry.Name() != record.RoomID+".json" {
			result.Errors = append(result.Errors, LoadError{RoomID: record.RoomID, Source: path, Err: fmt.Errorf("%w: file name does not match room ID", ErrInvalidRecord)})
			continue
		}
		result.Records = append(result.Records, record)
	}
	return result, nil
}

func (s *FileStore) Create(ctx context.Context, record Record) error {
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
	path := s.recordPath(record.RoomID)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrRecordExists, record.RoomID)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return s.writeRecord(record)
}

func (s *FileStore) Replace(ctx context.Context, expectedRevision uint64, record Record) error {
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
	current, err := readStoredRecord(s.recordPath(record.RoomID))
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrRecordNotFound, record.RoomID)
	}
	if err != nil {
		return err
	}
	if current.Revision != expectedRevision {
		return &RevisionConflictError{RoomID: record.RoomID, Expected: expectedRevision, Actual: current.Revision}
	}
	return s.writeRecord(record)
}

func (s *FileStore) Delete(ctx context.Context, roomID string, expectedRevision uint64) error {
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
	path := s.recordPath(roomID)
	current, err := readStoredRecord(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrRecordNotFound, roomID)
	}
	if err != nil {
		return err
	}
	if current.Revision != expectedRevision {
		return &RevisionConflictError{RoomID: roomID, Expected: expectedRevision, Actual: current.Revision}
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	syncDirectory(s.directory)
	return nil
}

func (s *FileStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

func (s *FileStore) recordPath(roomID string) string {
	return filepath.Join(s.directory, roomID+".json")
}

func (s *FileStore) writeRecord(record Record) (returnErr error) {
	payload, err := encodeStoredRecord(record, true)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(s.directory, "."+record.RoomID+"-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		if returnErr != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := replaceFile(temporaryPath, s.recordPath(record.RoomID)); err != nil {
		return err
	}
	syncDirectory(s.directory)
	return nil
}

func readStoredRecord(path string) (Record, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return Record{}, err
	}
	return decodeStoredRecord(payload)
}

func encodeStoredRecord(record Record, indent bool) ([]byte, error) {
	stored := storedRecord{RoomID: record.RoomID, Revision: record.Revision, Data: record.Data}
	if indent {
		payload, err := json.MarshalIndent(stored, "", "  ")
		return append(payload, '\n'), err
	}
	return json.Marshal(stored)
}

func decodeStoredRecord(payload []byte) (Record, error) {
	var stored storedRecord
	if err := json.Unmarshal(payload, &stored); err != nil {
		return Record{}, err
	}
	record := Record{RoomID: stored.RoomID, Revision: stored.Revision, Data: append([]byte(nil), stored.Data...)}
	if err := validateRecord(record); err != nil {
		return Record{}, err
	}
	return record, nil
}

func syncDirectory(directory string) {
	handle, err := os.Open(directory)
	if err != nil {
		return
	}
	defer handle.Close()
	_ = handle.Sync()
}
