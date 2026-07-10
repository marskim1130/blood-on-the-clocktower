package sessionstore

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

const defaultRedisPrefix = "clocktower"

const createRecordScript = `
if redis.call("EXISTS", KEYS[1]) == 1 then return "exists" end
redis.call("SET", KEYS[1], ARGV[2])
redis.call("SADD", KEYS[2], ARGV[1])
return "ok"`

const replaceRecordScript = `
local current = redis.call("GET", KEYS[1])
if not current then return "missing" end
local decoded = cjson.decode(current)
if tostring(decoded.revision) ~= ARGV[1] then return "conflict:" .. tostring(decoded.revision) end
redis.call("SET", KEYS[1], ARGV[2])
return "ok"`

const deleteRecordScript = `
local current = redis.call("GET", KEYS[1])
if not current then return "missing" end
local decoded = cjson.decode(current)
if tostring(decoded.revision) ~= ARGV[1] then return "conflict:" .. tostring(decoded.revision) end
redis.call("DEL", KEYS[1])
redis.call("SREM", KEYS[2], ARGV[2])
return "ok"`

type redisStoreClient interface {
	Eval(context.Context, string, []string, ...interface{}) (interface{}, error)
	Get(context.Context, string) (string, error)
	SScan(context.Context, string, uint64, string, int64) ([]string, uint64, error)
	Close() error
}

type goRedisClient struct{ client *redis.Client }

type RedisStore struct {
	client redisStoreClient
	prefix string
	mu     sync.RWMutex
	closed bool
}

func NewRedisStore(redisURL, prefix string) (*RedisStore, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return newRedisStore(&goRedisClient{client: redis.NewClient(options)}, prefix), nil
}

func newRedisStore(client redisStoreClient, prefix string) *RedisStore {
	prefix = strings.TrimSuffix(prefix, ":")
	if prefix == "" {
		prefix = defaultRedisPrefix
	}
	return &RedisStore{client: client, prefix: prefix}
}

func (c *goRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}
func (c *goRedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}
func (c *goRedisClient) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return c.client.SScan(ctx, key, cursor, match, count).Result()
}
func (c *goRedisClient) Close() error { return c.client.Close() }

func (s *RedisStore) LoadAll(ctx context.Context) (LoadAllResult, error) {
	if err := ctx.Err(); err != nil {
		return LoadAllResult{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return LoadAllResult{}, ErrClosed
	}
	var roomIDs []string
	var cursor uint64
	for {
		page, next, err := s.client.SScan(ctx, s.roomsKey(), cursor, "*", 100)
		if err != nil {
			return LoadAllResult{}, err
		}
		roomIDs = append(roomIDs, page...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	sort.Strings(roomIDs)
	result := LoadAllResult{}
	for _, roomID := range roomIDs {
		payload, err := s.client.Get(ctx, s.recordKey(roomID))
		if err != nil {
			result.Errors = append(result.Errors, LoadError{RoomID: roomID, Source: s.recordKey(roomID), Err: err})
			continue
		}
		record, err := decodeStoredRecord([]byte(payload))
		if err != nil || record.RoomID != roomID {
			if err == nil {
				err = fmt.Errorf("%w: Redis key does not match room ID", ErrInvalidRecord)
			}
			result.Errors = append(result.Errors, LoadError{RoomID: roomID, Source: s.recordKey(roomID), Err: err})
			continue
		}
		result.Records = append(result.Records, record)
	}
	return result, nil
}

func (s *RedisStore) Create(ctx context.Context, record Record) error {
	if err := validateRecord(record); err != nil {
		return err
	}
	payload, err := encodeStoredRecord(record, false)
	if err != nil {
		return err
	}
	result, err := s.eval(ctx, createRecordScript, []string{s.recordKey(record.RoomID), s.roomsKey()}, record.RoomID, string(payload))
	if err != nil {
		return err
	}
	if result == "exists" {
		return fmt.Errorf("%w: %s", ErrRecordExists, record.RoomID)
	}
	return expectRedisResult(result)
}

func (s *RedisStore) Replace(ctx context.Context, expectedRevision uint64, record Record) error {
	if err := validateRecord(record); err != nil {
		return err
	}
	payload, err := encodeStoredRecord(record, false)
	if err != nil {
		return err
	}
	result, err := s.eval(ctx, replaceRecordScript, []string{s.recordKey(record.RoomID)}, strconv.FormatUint(expectedRevision, 10), string(payload))
	if err != nil {
		return err
	}
	return mapRedisCASResult(record.RoomID, expectedRevision, result)
}

func (s *RedisStore) Delete(ctx context.Context, roomID string, expectedRevision uint64) error {
	if err := validateRoomID(roomID); err != nil {
		return err
	}
	result, err := s.eval(ctx, deleteRecordScript, []string{s.recordKey(roomID), s.roomsKey()}, strconv.FormatUint(expectedRevision, 10), roomID)
	if err != nil {
		return err
	}
	return mapRedisCASResult(roomID, expectedRevision, result)
}

func (s *RedisStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.client.Close()
}

func (s *RedisStore) eval(ctx context.Context, script string, keys []string, args ...interface{}) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return "", ErrClosed
	}
	result, err := s.client.Eval(ctx, script, keys, args...)
	if err != nil {
		return "", err
	}
	text, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected Redis script result %T", result)
	}
	return text, nil
}

func (s *RedisStore) roomsKey() string               { return s.prefix + ":rooms" }
func (s *RedisStore) recordKey(roomID string) string { return s.prefix + ":room:" + roomID }

func mapRedisCASResult(roomID string, expectedRevision uint64, result string) error {
	switch {
	case result == "ok":
		return nil
	case result == "missing":
		return fmt.Errorf("%w: %s", ErrRecordNotFound, roomID)
	case strings.HasPrefix(result, "conflict:"):
		actual, err := strconv.ParseUint(strings.TrimPrefix(result, "conflict:"), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid Redis revision conflict result %q: %w", result, err)
		}
		return &RevisionConflictError{RoomID: roomID, Expected: expectedRevision, Actual: actual}
	default:
		return expectRedisResult(result)
	}
}

func expectRedisResult(result string) error {
	if result == "ok" {
		return nil
	}
	return fmt.Errorf("unexpected Redis script result %q", result)
}
