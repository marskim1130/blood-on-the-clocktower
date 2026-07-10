package sessionstore

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRedisStoreClient struct {
	values      map[string]string
	sets        map[string][]string
	scanPages   map[uint64]redisScanPage
	evalResults []interface{}
	evalErr     error
	closed      bool
	lastScript  string
	lastKeys    []string
	lastArgs    []interface{}
}

type redisScanPage struct {
	values []string
	next   uint64
	err    error
}

func newFakeRedisStoreClient() *fakeRedisStoreClient {
	return &fakeRedisStoreClient{values: make(map[string]string), sets: make(map[string][]string), scanPages: make(map[uint64]redisScanPage)}
}

func (c *fakeRedisStoreClient) Eval(_ context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	c.lastScript = script
	c.lastKeys = append([]string(nil), keys...)
	c.lastArgs = append([]interface{}(nil), args...)
	if c.evalErr != nil {
		return nil, c.evalErr
	}
	if len(c.evalResults) == 0 {
		return "ok", nil
	}
	result := c.evalResults[0]
	c.evalResults = c.evalResults[1:]
	return result, nil
}
func (c *fakeRedisStoreClient) Get(_ context.Context, key string) (string, error) {
	value, ok := c.values[key]
	if !ok {
		return "", errors.New("missing Redis value")
	}
	return value, nil
}
func (c *fakeRedisStoreClient) SScan(_ context.Context, _ string, cursor uint64, _ string, _ int64) ([]string, uint64, error) {
	page := c.scanPages[cursor]
	return append([]string(nil), page.values...), page.next, page.err
}
func (c *fakeRedisStoreClient) Close() error { c.closed = true; return nil }

func TestRedisStoreUsesPerRoomKeysAndCAS(t *testing.T) {
	client := newFakeRedisStoreClient()
	client.evalResults = []interface{}{"ok", "conflict:4", "missing"}
	store := newRedisStore(client, "clocktower:test:")
	record := Record{RoomID: "room-a", Revision: 1, Data: []byte(`{"phase":"lobby"}`)}
	if err := store.Create(context.Background(), record); err != nil {
		t.Fatalf("create record: %v", err)
	}
	if !reflect.DeepEqual(client.lastKeys, []string{"clocktower:test:room:room-a", "clocktower:test:rooms"}) {
		t.Fatalf("create keys = %#v", client.lastKeys)
	}
	err := store.Replace(context.Background(), 3, Record{RoomID: "room-a", Revision: 5, Data: []byte(`{}`)})
	var conflict *RevisionConflictError
	if !errors.As(err, &conflict) || conflict.Actual != 4 {
		t.Fatalf("replace error = %#v, want actual revision 4", err)
	}
	if err := store.Delete(context.Background(), "room-a", 5); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("delete error = %v, want ErrRecordNotFound", err)
	}
}

func TestRedisStoreLoadsPagesAndIsolatesBadRecords(t *testing.T) {
	client := newFakeRedisStoreClient()
	client.scanPages[0] = redisScanPage{values: []string{"room-b"}, next: 7}
	client.scanPages[7] = redisScanPage{values: []string{"room-a", "broken"}, next: 0}
	validA, _ := encodeStoredRecord(Record{RoomID: "room-a", Revision: 1, Data: []byte(`{"a":1}`)}, false)
	validB, _ := encodeStoredRecord(Record{RoomID: "room-b", Revision: 2, Data: []byte(`{"b":2}`)}, false)
	client.values["clocktower:room:room-a"] = string(validA)
	client.values["clocktower:room:room-b"] = string(validB)
	client.values["clocktower:room:broken"] = `not-json`
	store := newRedisStore(client, "")
	result, err := store.LoadAll(context.Background())
	if err != nil {
		t.Fatalf("load all: %v", err)
	}
	if len(result.Records) != 2 || result.Records[0].RoomID != "room-a" || result.Records[1].RoomID != "room-b" {
		t.Fatalf("records = %#v", result.Records)
	}
	if len(result.Errors) != 1 || result.Errors[0].RoomID != "broken" {
		t.Fatalf("errors = %#v", result.Errors)
	}
}

func TestRedisStoreCloseClosesClient(t *testing.T) {
	client := newFakeRedisStoreClient()
	store := newRedisStore(client, "test")
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	if !client.closed {
		t.Fatal("expected Redis client to close")
	}
	if err := store.Create(context.Background(), Record{RoomID: "room-a", Revision: 1, Data: []byte(`{}`)}); !errors.Is(err, ErrClosed) {
		t.Fatalf("create after close error = %v, want ErrClosed", err)
	}
}

func TestNewRedisStoreRejectsInvalidURL(t *testing.T) {
	if _, err := NewRedisStore("://bad-url", "test"); err == nil {
		t.Fatal("expected invalid Redis URL to be rejected")
	}
}
