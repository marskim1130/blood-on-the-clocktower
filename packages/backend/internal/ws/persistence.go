package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

const snapshotVersion = 1
const defaultRedisSnapshotKey = "clocktower:snapshot"
const snapshotStoreTimeout = 3 * time.Second

type snapshotStore interface {
	Load() (*hubSnapshot, error)
	Save(*hubSnapshot) error
}

type hubSnapshot struct {
	Version  int                            `json:"version"`
	Rooms    []roomSnapshot                 `json:"rooms"`
	Sessions map[string]gameSessionSnapshot `json:"sessions"`
}

type roomSnapshot struct {
	ID              string   `json:"id"`
	MaxPlayers      int      `json:"maxPlayers"`
	CreatorID       string   `json:"creatorId"`
	ScriptID        string   `json:"scriptId"`
	KickedPlayerIDs []string `json:"kickedPlayerIds,omitempty"`
}

type gameSessionSnapshot struct {
	Players         []game.Player        `json:"players"`
	StorytellerID   string               `json:"storytellerId"`
	OriginalPlayers int                  `json:"originalPlayers"`
	ScriptID        string               `json:"scriptId"`
	Phase           game.GamePhase       `json:"phase"`
	DayNumber       int32                `json:"dayNumber"`
	NightNumber     int32                `json:"nightNumber"`
	NightWakeIndex  int                  `json:"nightWakeIndex"`
	Nomination      *game.Nomination     `json:"nomination,omitempty"`
	NightActions    []game.NightAction   `json:"nightActions,omitempty"`
	Deaths          []game.DeathRecord   `json:"deaths,omitempty"`
	GhostVotesUsed  map[string]bool      `json:"ghostVotesUsed,omitempty"`
	SlayerUsed      map[string]bool      `json:"slayerUsed,omitempty"`
	Winner          *game.GameEndedEvent `json:"winner,omitempty"`
}

type FileSnapshotStore struct {
	path string
	mu   sync.Mutex
}

type redisSnapshotClient interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string) error
	Close() error
}

type RedisSnapshotStore struct {
	client redisSnapshotClient
	key    string
	mu     sync.Mutex
}

type redisClientAdapter struct {
	client *redis.Client
}

func NewFileSnapshotStore(path string) *FileSnapshotStore {
	return &FileSnapshotStore{path: path}
}

func NewRedisSnapshotStore(redisURL, key string) (*RedisSnapshotStore, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return newRedisSnapshotStoreWithClient(&redisClientAdapter{client: redis.NewClient(options)}, key), nil
}

func newRedisSnapshotStoreWithClient(client redisSnapshotClient, key string) *RedisSnapshotStore {
	if key == "" {
		key = defaultRedisSnapshotKey
	}
	return &RedisSnapshotStore{client: client, key: key}
}

func (c *redisClientAdapter) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisClientAdapter) Set(ctx context.Context, key string, value string) error {
	return c.client.Set(ctx, key, value, 0).Err()
}

func (c *redisClientAdapter) Close() error {
	return c.client.Close()
}

func (s *FileSnapshotStore) Load() (*hubSnapshot, error) {
	if s == nil || s.path == "" {
		return emptyHubSnapshot(), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return emptyHubSnapshot(), nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return emptyHubSnapshot(), nil
	}

	var snapshot hubSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	if snapshot.Sessions == nil {
		snapshot.Sessions = make(map[string]gameSessionSnapshot)
	}
	return &snapshot, nil
}

func (s *FileSnapshotStore) Save(snapshot *hubSnapshot) error {
	if s == nil || s.path == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if snapshot == nil {
		snapshot = emptyHubSnapshot()
	}
	snapshot.Version = snapshotVersion
	if snapshot.Sessions == nil {
		snapshot.Sessions = make(map[string]gameSessionSnapshot)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *RedisSnapshotStore) Load() (*hubSnapshot, error) {
	if s == nil || s.client == nil {
		return emptyHubSnapshot(), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), snapshotStoreTimeout)
	defer cancel()

	data, err := s.client.Get(ctx, s.key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return emptyHubSnapshot(), nil
		}
		return nil, err
	}
	if data == "" {
		return emptyHubSnapshot(), nil
	}

	var snapshot hubSnapshot
	if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
		return nil, err
	}
	if snapshot.Sessions == nil {
		snapshot.Sessions = make(map[string]gameSessionSnapshot)
	}
	return &snapshot, nil
}

func (s *RedisSnapshotStore) Save(snapshot *hubSnapshot) error {
	if s == nil || s.client == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if snapshot == nil {
		snapshot = emptyHubSnapshot()
	}
	snapshot.Version = snapshotVersion
	if snapshot.Sessions == nil {
		snapshot.Sessions = make(map[string]gameSessionSnapshot)
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), snapshotStoreTimeout)
	defer cancel()
	return s.client.Set(ctx, s.key, string(data))
}

func (s *RedisSnapshotStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

func emptyHubSnapshot() *hubSnapshot {
	return &hubSnapshot{
		Version:  snapshotVersion,
		Rooms:    []roomSnapshot{},
		Sessions: make(map[string]gameSessionSnapshot),
	}
}

func NewHubWithSnapshotStore(store snapshotStore) (*Hub, error) {
	h := newHub()
	h.snapshotStore = store
	if store == nil {
		return h, nil
	}

	snapshot, err := store.Load()
	if err != nil {
		return nil, err
	}
	if err := h.restoreSnapshot(snapshot); err != nil {
		return nil, err
	}
	return h, nil
}

func NewHubWithFileSnapshot(path string) (*Hub, error) {
	return NewHubWithSnapshotStore(NewFileSnapshotStore(path))
}

func NewHubWithRedisSnapshot(redisURL, key string) (*Hub, error) {
	store, err := NewRedisSnapshotStore(redisURL, key)
	if err != nil {
		return nil, err
	}
	return NewHubWithSnapshotStore(store)
}

func (h *Hub) snapshot() *hubSnapshot {
	h.mu.RLock()
	sessions := make(map[string]*GameSession, len(h.sessions))
	for roomID, session := range h.sessions {
		sessions[roomID] = session
	}
	h.mu.RUnlock()

	sessionSnapshots := make(map[string]gameSessionSnapshot, len(sessions))
	for roomID, session := range sessions {
		sessionSnapshots[roomID] = session.snapshot()
	}

	return &hubSnapshot{
		Version:  snapshotVersion,
		Rooms:    h.rm.snapshot(),
		Sessions: sessionSnapshots,
	}
}

func (h *Hub) restoreSnapshot(snapshot *hubSnapshot) error {
	if snapshot == nil {
		return nil
	}
	if snapshot.Version != 0 && snapshot.Version != snapshotVersion {
		return fmt.Errorf("unsupported snapshot version %d", snapshot.Version)
	}
	if err := h.rm.restoreSnapshot(snapshot.Rooms); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions = make(map[string]*GameSession, len(snapshot.Sessions))
	for roomID, sessionSnapshot := range snapshot.Sessions {
		if h.rm.GetRoom(roomID) == nil {
			continue
		}
		h.sessions[roomID] = newGameSessionFromSnapshot(sessionSnapshot)
	}
	return nil
}

func (h *Hub) persistSnapshot() {
	if h.snapshotStore == nil {
		return
	}
	if err := h.snapshotStore.Save(h.snapshot()); err != nil {
		log.Printf("failed to persist game snapshot: %v", err)
	}
}

func (rm *RoomManager) snapshot() []roomSnapshot {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]roomSnapshot, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		room.mu.RLock()
		rooms = append(rooms, roomSnapshot{
			ID:              room.id,
			MaxPlayers:      room.maxPlayers,
			CreatorID:       room.creatorID,
			ScriptID:        room.scriptID,
			KickedPlayerIDs: cloneKickedPlayerIDs(room.kicked),
		})
		room.mu.RUnlock()
	}
	return rooms
}

func (rm *RoomManager) restoreSnapshot(rooms []roomSnapshot) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	restored := make(map[string]*Room, len(rooms))
	for _, room := range rooms {
		if room.ID == "" {
			return fmt.Errorf("snapshot room id is required")
		}
		if _, exists := restored[room.ID]; exists {
			return fmt.Errorf("duplicate snapshot room id %s", room.ID)
		}

		maxPlayers := room.MaxPlayers
		if maxPlayers < 5 || maxPlayers > 15 {
			maxPlayers = defaultMaxPlayers
		}
		scriptID := room.ScriptID
		if scriptID == "" {
			scriptID = game.TroubleBrewingScriptID
		}
		if game.GetScriptByID(scriptID) == nil {
			return fmt.Errorf("snapshot room %s has unsupported script %s", room.ID, scriptID)
		}

		restored[room.ID] = &Room{
			id:         room.ID,
			clients:    make(map[string]*Client),
			maxPlayers: maxPlayers,
			creatorID:  room.CreatorID,
			scriptID:   scriptID,
			kicked:     kickedPlayerSet(room.KickedPlayerIDs),
		}
	}
	rm.rooms = restored
	return nil
}

func (gs *GameSession) snapshot() gameSessionSnapshot {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	return gameSessionSnapshot{
		Players:         clonePlayers(gs.players),
		StorytellerID:   gs.storytellerID,
		OriginalPlayers: gs.originalPlayers,
		ScriptID:        gs.scriptID,
		Phase:           gs.phase,
		DayNumber:       gs.dayNumber,
		NightNumber:     gs.nightNumber,
		NightWakeIndex:  gs.nightWakeIndex,
		Nomination:      cloneNomination(gs.nomination),
		NightActions:    cloneNightActions(gs.nightActions),
		Deaths:          cloneDeaths(gs.deaths),
		GhostVotesUsed:  cloneGhostVotesUsed(gs.ghostVotesUsed),
		SlayerUsed:      cloneGhostVotesUsed(gs.slayerUsed),
		Winner:          cloneWinner(gs.winner),
	}
}

func newGameSessionFromSnapshot(snapshot gameSessionSnapshot) *GameSession {
	scriptID := snapshot.ScriptID
	if scriptID == "" {
		scriptID = game.TroubleBrewingScriptID
	}
	ghostVotesUsed := cloneGhostVotesUsed(snapshot.GhostVotesUsed)
	if ghostVotesUsed == nil {
		ghostVotesUsed = make(map[string]bool)
	}
	slayerUsed := cloneGhostVotesUsed(snapshot.SlayerUsed)
	if slayerUsed == nil {
		slayerUsed = make(map[string]bool)
	}

	return &GameSession{
		players:         clonePlayers(snapshot.Players),
		storytellerID:   snapshot.StorytellerID,
		originalPlayers: snapshot.OriginalPlayers,
		scriptID:        scriptID,
		phase:           snapshot.Phase,
		dayNumber:       snapshot.DayNumber,
		nightNumber:     snapshot.NightNumber,
		nightWakeIndex:  snapshot.NightWakeIndex,
		nomination:      cloneNomination(snapshot.Nomination),
		nightActions:    cloneNightActions(snapshot.NightActions),
		deaths:          cloneDeaths(snapshot.Deaths),
		ghostVotesUsed:  ghostVotesUsed,
		slayerUsed:      slayerUsed,
		winner:          cloneWinner(snapshot.Winner),
	}
}

func clonePlayers(players []game.Player) []game.Player {
	result := make([]game.Player, len(players))
	for i, player := range players {
		result[i] = player
		if player.Character != nil {
			character := *player.Character
			result[i].Character = &character
		}
	}
	return result
}

func cloneNomination(nomination *game.Nomination) *game.Nomination {
	if nomination == nil {
		return nil
	}
	result := *nomination
	if nomination.Votes != nil {
		result.Votes = make(map[string]bool, len(nomination.Votes))
		for playerID, decision := range nomination.Votes {
			result.Votes[playerID] = decision
		}
	}
	return &result
}

func cloneNightActions(actions []game.NightAction) []game.NightAction {
	result := make([]game.NightAction, len(actions))
	for i, action := range actions {
		result[i] = action
		result[i].TargetIDs = append([]string(nil), action.TargetIDs...)
	}
	return result
}

func cloneDeaths(deaths []game.DeathRecord) []game.DeathRecord {
	result := make([]game.DeathRecord, len(deaths))
	copy(result, deaths)
	return result
}

func cloneGhostVotesUsed(ghostVotesUsed map[string]bool) map[string]bool {
	if ghostVotesUsed == nil {
		return nil
	}
	result := make(map[string]bool, len(ghostVotesUsed))
	for playerID, used := range ghostVotesUsed {
		result[playerID] = used
	}
	return result
}

func cloneWinner(winner *game.GameEndedEvent) *game.GameEndedEvent {
	if winner == nil {
		return nil
	}
	result := *winner
	return &result
}

func cloneKickedPlayerIDs(kicked map[string]bool) []string {
	if len(kicked) == 0 {
		return nil
	}
	result := make([]string, 0, len(kicked))
	for playerID, isKicked := range kicked {
		if isKicked {
			result = append(result, playerID)
		}
	}
	return result
}

func kickedPlayerSet(playerIDs []string) map[string]bool {
	result := make(map[string]bool, len(playerIDs))
	for _, playerID := range playerIDs {
		if playerID != "" {
			result[playerID] = true
		}
	}
	return result
}
