package sessionstore

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrClosed           = errors.New("room record store is closed")
	ErrInvalidRecord    = errors.New("invalid room record")
	ErrRecordExists     = errors.New("room record already exists")
	ErrRecordNotFound   = errors.New("room record not found")
	ErrRevisionConflict = errors.New("room record revision conflict")
)

type Record struct {
	RoomID   string
	Revision uint64
	Data     []byte
}

type LoadError struct {
	RoomID string
	Source string
	Err    error
}

func (e LoadError) Error() string {
	if e.RoomID == "" {
		return fmt.Sprintf("load room record from %s: %v", e.Source, e.Err)
	}
	return fmt.Sprintf("load room record %q from %s: %v", e.RoomID, e.Source, e.Err)
}

func (e LoadError) Unwrap() error { return e.Err }

type LoadAllResult struct {
	Records []Record
	Errors  []LoadError
}

type RevisionConflictError struct {
	RoomID   string
	Expected uint64
	Actual   uint64
}

func (e *RevisionConflictError) Error() string {
	return fmt.Sprintf("%s for room %q: expected %d, actual %d", ErrRevisionConflict, e.RoomID, e.Expected, e.Actual)
}

func (e *RevisionConflictError) Unwrap() error { return ErrRevisionConflict }

type RoomRecordStore interface {
	LoadAll(context.Context) (LoadAllResult, error)
	Create(context.Context, Record) error
	Replace(context.Context, uint64, Record) error
	Delete(context.Context, string, uint64) error
	Close() error
}

func cloneRecord(record Record) Record {
	cloned := record
	cloned.Data = append([]byte(nil), record.Data...)
	return cloned
}
