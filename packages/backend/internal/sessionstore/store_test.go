package sessionstore

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

type storeFactory struct {
	name string
	new  func(*testing.T) RoomRecordStore
}

func TestRoomRecordStoreContract(t *testing.T) {
	factories := []storeFactory{
		{name: "memory", new: func(*testing.T) RoomRecordStore { return NewMemoryStore() }},
		{name: "file", new: func(t *testing.T) RoomRecordStore {
			store, err := NewFileStore(filepath.Join(t.TempDir(), "records"))
			if err != nil {
				t.Fatalf("create file store: %v", err)
			}
			return store
		}},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) { testStoreContract(t, factory.new(t)) })
	}
}

func testStoreContract(t *testing.T, store RoomRecordStore) {
	t.Helper()
	ctx := context.Background()
	first := Record{RoomID: "room-a", Revision: 1, Data: []byte(`{"phase":"lobby"}`)}
	second := Record{RoomID: "room-b", Revision: 3, Data: []byte(`{"phase":"night"}`)}

	if err := store.Create(ctx, first); err != nil {
		t.Fatalf("create first record: %v", err)
	}
	if err := store.Create(ctx, second); err != nil {
		t.Fatalf("create second record: %v", err)
	}
	if err := store.Create(ctx, first); !errors.Is(err, ErrRecordExists) {
		t.Fatalf("duplicate create error = %v, want ErrRecordExists", err)
	}

	loaded, err := store.LoadAll(ctx)
	if err != nil {
		t.Fatalf("load records: %v", err)
	}
	if len(loaded.Errors) != 0 {
		t.Fatalf("unexpected load errors: %#v", loaded.Errors)
	}
	if !reflect.DeepEqual(loaded.Records, []Record{first, second}) {
		t.Fatalf("loaded records = %#v", loaded.Records)
	}
	loaded.Records[0].Data[0] = 'x'
	reloaded, err := store.LoadAll(ctx)
	if err != nil {
		t.Fatalf("reload records: %v", err)
	}
	if !reflect.DeepEqual(reloaded.Records[0], first) {
		t.Fatalf("store leaked mutable data: %#v", reloaded.Records[0])
	}

	replacement := Record{RoomID: first.RoomID, Revision: 2, Data: []byte(`{"phase":"day"}`)}
	if err := store.Replace(ctx, 99, replacement); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("replace conflict error = %v, want ErrRevisionConflict", err)
	}
	if err := store.Replace(ctx, first.Revision, replacement); err != nil {
		t.Fatalf("replace record: %v", err)
	}
	if err := store.Replace(ctx, 1, Record{RoomID: "missing", Revision: 2, Data: []byte(`{}`)}); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("replace missing error = %v, want ErrRecordNotFound", err)
	}
	if err := store.Delete(ctx, replacement.RoomID, first.Revision); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("delete conflict error = %v, want ErrRevisionConflict", err)
	}
	if err := store.Delete(ctx, replacement.RoomID, replacement.Revision); err != nil {
		t.Fatalf("delete record: %v", err)
	}
	if err := store.Delete(ctx, replacement.RoomID, replacement.Revision); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("delete missing error = %v, want ErrRecordNotFound", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store twice: %v", err)
	}
	if _, err := store.LoadAll(ctx); !errors.Is(err, ErrClosed) {
		t.Fatalf("load after close error = %v, want ErrClosed", err)
	}
}

func TestStoresRejectInvalidRecords(t *testing.T) {
	store := NewMemoryStore()
	invalid := []Record{
		{RoomID: "", Revision: 1, Data: []byte(`{}`)},
		{RoomID: "../escape", Revision: 1, Data: []byte(`{}`)},
		{RoomID: "room-a", Revision: 0, Data: []byte(`{}`)},
		{RoomID: "room-a", Revision: 1, Data: []byte(`not-json`)},
	}
	for _, record := range invalid {
		if err := store.Create(context.Background(), record); !errors.Is(err, ErrInvalidRecord) {
			t.Fatalf("create %#v error = %v, want ErrInvalidRecord", record, err)
		}
	}
}
