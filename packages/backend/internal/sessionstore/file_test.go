package sessionstore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorePersistsAcrossInstances(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "records")
	store, err := NewFileStore(directory)
	if err != nil {
		t.Fatalf("create file store: %v", err)
	}
	record := Record{RoomID: "persistent-room", Revision: 1, Data: []byte(`{"players":[]}`)}
	if err := store.Create(context.Background(), record); err != nil {
		t.Fatalf("create record: %v", err)
	}
	reopened, err := NewFileStore(directory)
	if err != nil {
		t.Fatalf("reopen file store: %v", err)
	}
	result, err := reopened.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("load records: %v", err)
	}
	if len(result.Records) != 1 || result.Records[0].RoomID != record.RoomID {
		t.Fatalf("loaded records = %#v", result.Records)
	}
}

func TestFileStoreIsolatesCorruptRecords(t *testing.T) {
	directory := t.TempDir()
	store, err := NewFileStore(directory)
	if err != nil {
		t.Fatalf("create file store: %v", err)
	}
	valid := Record{RoomID: "valid-room", Revision: 1, Data: []byte(`{"ok":true}`)}
	if err := store.Create(context.Background(), valid); err != nil {
		t.Fatalf("create valid record: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "broken-room.json"), []byte(`{"roomId":`), 0o600); err != nil {
		t.Fatalf("write broken record: %v", err)
	}
	mismatched, err := encodeStoredRecord(Record{RoomID: "different-room", Revision: 1, Data: []byte(`{}`)}, false)
	if err != nil {
		t.Fatalf("encode mismatched record: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "wrong-name.json"), mismatched, 0o600); err != nil {
		t.Fatalf("write mismatched record: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, ".valid-room-leftover.tmp"), []byte("partial"), 0o600); err != nil {
		t.Fatalf("write temporary file: %v", err)
	}
	result, err := store.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("load records: %v", err)
	}
	if len(result.Records) != 1 || result.Records[0].RoomID != valid.RoomID {
		t.Fatalf("valid records = %#v", result.Records)
	}
	if len(result.Errors) != 2 {
		t.Fatalf("load errors = %#v, want 2", result.Errors)
	}
}

func TestFileStoreRejectsFilePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
	if _, err := NewFileStore(path); err == nil {
		t.Fatal("expected file path to be rejected")
	}
}

func TestFileStoreLeavesOldRecordOnRevisionConflict(t *testing.T) {
	store, err := NewFileStore(t.TempDir())
	if err != nil {
		t.Fatalf("create file store: %v", err)
	}
	original := Record{RoomID: "room-a", Revision: 4, Data: []byte(`{"value":"old"}`)}
	if err := store.Create(context.Background(), original); err != nil {
		t.Fatalf("create record: %v", err)
	}
	err = store.Replace(context.Background(), 3, Record{RoomID: original.RoomID, Revision: 5, Data: []byte(`{"value":"new"}`)})
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("replace error = %v, want ErrRevisionConflict", err)
	}
	result, err := store.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("load records: %v", err)
	}
	if string(result.Records[0].Data) != string(original.Data) {
		t.Fatalf("record changed after conflict: %s", result.Records[0].Data)
	}
}
