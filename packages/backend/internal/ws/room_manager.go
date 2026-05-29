package ws

import (
	"fmt"
	"math/rand"
	"sync"
)

const defaultMaxPlayers = 10

// Room holds room metadata and connected clients.
// No GameState — that lives in GameSession.
type Room struct {
	mu         sync.RWMutex
	id         string
	clients    map[string]*Client // playerID -> Client
	maxPlayers int
	creatorID  string
}

// RoomManager manages room lifecycle.
// No game rules, no broadcasting.
type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) CreateRoom(creatorID string, maxPlayers int) *Room {
	if maxPlayers < 5 || maxPlayers > 15 {
		maxPlayers = defaultMaxPlayers
	}

	roomID := rm.generateRoomID()

	rm.mu.Lock()
	room := &Room{
		id:         roomID,
		clients:    make(map[string]*Client),
		maxPlayers: maxPlayers,
		creatorID:  creatorID,
	}
	rm.rooms[roomID] = room
	rm.mu.Unlock()

	return room
}

func (rm *RoomManager) JoinRoom(roomID string, conn Connection, playerID, playerName string) error {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.clients) >= room.maxPlayers {
		return fmt.Errorf("room is full")
	}

	room.clients[playerID] = &Client{
		Conn:     conn,
		PlayerID: playerID,
	}
	return nil
}

func (rm *RoomManager) LeaveRoom(playerID string) (roomID string, err error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, room := range rm.rooms {
		room.mu.Lock()
		if _, ok := room.clients[playerID]; ok {
			delete(room.clients, playerID)
			roomID = id
			shouldDestroy := len(room.clients) == 0
			room.mu.Unlock()

			if shouldDestroy {
				delete(rm.rooms, id)
			}
			return roomID, nil
		}
		room.mu.Unlock()
	}
	return "", fmt.Errorf("player not in any room")
}

func (rm *RoomManager) RemoveClientByConn(conn Connection) (roomID, playerID string, err error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for id, room := range rm.rooms {
		room.mu.Lock()
		for pid, client := range room.clients {
			if client.Conn == conn {
				delete(room.clients, pid)
				roomID = id
				playerID = pid
				shouldDestroy := len(room.clients) == 0
				room.mu.Unlock()

				if shouldDestroy {
					delete(rm.rooms, id)
				}
				return roomID, playerID, nil
			}
		}
		room.mu.Unlock()
	}
	return "", "", fmt.Errorf("connection not in any room")
}

func (rm *RoomManager) GetRoom(roomID string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) GetClient(playerID string) (*Client, string) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for roomID, room := range rm.rooms {
		room.mu.RLock()
		if client, ok := room.clients[playerID]; ok {
			room.mu.RUnlock()
			return client, roomID
		}
		room.mu.RUnlock()
	}
	return nil, ""
}

func (rm *RoomManager) GetClientsByRoom(roomID string) map[string]*Client {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	result := make(map[string]*Client, len(room.clients))
	for pid, client := range room.clients {
		result[pid] = client
	}
	return result
}

func (rm *RoomManager) CreatorID(roomID string) string {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return ""
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	return room.creatorID
}

func (rm *RoomManager) MaxPlayers(roomID string) int {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return 0
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	return room.maxPlayers
}

func (rm *RoomManager) PlayerCount(roomID string) int {
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	rm.mu.RUnlock()

	if !exists {
		return 0
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	return len(room.clients)
}

func (rm *RoomManager) generateRoomID() string {
	for {
		id := fmt.Sprintf("%06d", rand.Intn(1000000))
		rm.mu.RLock()
		_, exists := rm.rooms[id]
		rm.mu.RUnlock()
		if !exists {
			return id
		}
	}
}
