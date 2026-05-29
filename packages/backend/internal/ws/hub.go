package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const defaultMaxPlayers = 10

// Hub manages all rooms and WebSocket connections
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

// Room represents a game room with connected clients
type Room struct {
	mu         sync.RWMutex
	id         string
	clients    map[*websocket.Conn]string // conn -> playerID
	players    []game.Player
	maxPlayers int
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

// HandleWebSocket handles WebSocket upgrade and message routing
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer conn.Close()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			h.handleDisconnect(conn)
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}

		h.handleMessage(conn, msg)
	}
}

func (h *Hub) handleMessage(conn *websocket.Conn, msg ClientMessage) {
	switch msg.Type {
	case "CREATE_ROOM":
		h.handleCreateRoom(conn, msg)
	case "JOIN_ROOM":
		h.handleJoinRoom(conn, msg)
	case "LEAVE_ROOM":
		h.handleLeaveRoom(conn, msg)
	case "SUBMIT_EVENT":
		h.handleSubmitEvent(conn, msg)
	}
}

func (h *Hub) handleCreateRoom(conn *websocket.Conn, msg ClientMessage) {
	maxPlayers := msg.MaxPlayers
	if maxPlayers < 5 || maxPlayers > 15 {
		maxPlayers = defaultMaxPlayers
	}

	roomID := h.generateRoomID()

	h.mu.Lock()
	room := &Room{
		id:         roomID,
		clients:    make(map[*websocket.Conn]string),
		maxPlayers: maxPlayers,
	}
	h.rooms[roomID] = room
	h.mu.Unlock()

	player := game.Player{
		ID:      msg.PlayerID,
		Name:    msg.PlayerName,
		IsAlive: true,
	}

	room.mu.Lock()
	room.players = append(room.players, player)
	room.clients[conn] = msg.PlayerID
	room.mu.Unlock()

	conn.WriteJSON(ServerMessage{
		Type:   "ROOM_STATE",
		RoomID: roomID,
		State:  room.getState(),
	})
}

func (h *Hub) handleJoinRoom(conn *websocket.Conn, msg ClientMessage) {
	h.mu.RLock()
	room, exists := h.rooms[msg.RoomID]
	h.mu.RUnlock()

	if !exists {
		conn.WriteJSON(ServerMessage{
			Type:  "ERROR",
			Error: "room not found",
		})
		return
	}

	room.mu.Lock()
	if len(room.players) >= room.maxPlayers {
		room.mu.Unlock()
		conn.WriteJSON(ServerMessage{
			Type:  "ERROR",
			Error: "room is full",
		})
		return
	}

	player := game.Player{
		ID:      msg.PlayerID,
		Name:    msg.PlayerName,
		IsAlive: true,
	}
	room.players = append(room.players, player)
	room.clients[conn] = msg.PlayerID
	room.mu.Unlock()

	conn.WriteJSON(ServerMessage{
		Type:   "ROOM_STATE",
		RoomID: room.id,
		State:  room.getState(),
	})

	event := game.GameEvent{
		PlayerJoined: &game.PlayerJoined{Player: player},
	}
	room.broadcastExcept(conn, ServerMessage{
		Type:  "EVENT_BROADCAST",
		Event: &event,
	})
}

func (h *Hub) handleLeaveRoom(conn *websocket.Conn, msg ClientMessage) {
	var targetRoom *Room
	var playerID string

	h.mu.Lock()
	for _, room := range h.rooms {
		room.mu.Lock()
		if pid, ok := room.clients[conn]; ok {
			targetRoom = room
			playerID = pid
			room.mu.Unlock()
			break
		}
		room.mu.Unlock()
	}
	h.mu.Unlock()

	if targetRoom == nil {
		return
	}

	targetRoom.mu.Lock()
	delete(targetRoom.clients, conn)
	for i, p := range targetRoom.players {
		if p.ID == playerID {
			targetRoom.players = append(targetRoom.players[:i], targetRoom.players[i+1:]...)
			break
		}
	}
	targetRoom.broadcast(ServerMessage{
		Type:  "EVENT_BROADCAST",
		Event: &game.GameEvent{PlayerLeft: &game.PlayerLeft{PlayerID: playerID}},
	})
	shouldDestroy := len(targetRoom.clients) == 0
	roomID := targetRoom.id
	targetRoom.mu.Unlock()

	if shouldDestroy {
		h.mu.Lock()
		delete(h.rooms, roomID)
		h.mu.Unlock()
	}
}

func (h *Hub) handleSubmitEvent(conn *websocket.Conn, msg ClientMessage) {
	if msg.Event == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, room := range h.rooms {
		room.mu.RLock()
		if _, ok := room.clients[conn]; ok {
			room.broadcast(ServerMessage{
				Type:  "EVENT_BROADCAST",
				Event: msg.Event,
			})
		}
		room.mu.RUnlock()
	}
}

func (h *Hub) handleDisconnect(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for id, room := range h.rooms {
		room.mu.Lock()
		if playerID, ok := room.clients[conn]; ok {
			delete(room.clients, conn)
			for i, p := range room.players {
				if p.ID == playerID {
					room.players = append(room.players[:i], room.players[i+1:]...)
					break
				}
			}
			event := game.GameEvent{
				PlayerLeft: &game.PlayerLeft{PlayerID: playerID},
			}
			room.broadcast(ServerMessage{
				Type:  "EVENT_BROADCAST",
				Event: &event,
			})
			if len(room.clients) == 0 {
				delete(h.rooms, id)
			}
		}
		room.mu.Unlock()
	}
}

func (h *Hub) generateRoomID() string {
	for {
		id := fmt.Sprintf("%06d", rand.Intn(1000000))
		h.mu.RLock()
		_, exists := h.rooms[id]
		h.mu.RUnlock()
		if !exists {
			return id
		}
	}
}

func (r *Room) getState() *RoomState {
	return &RoomState{
		RoomID:     r.id,
		Players:    r.players,
		MaxPlayers: r.maxPlayers,
	}
}

func (r *Room) broadcast(msg ServerMessage) {
	for conn := range r.clients {
		conn.WriteJSON(msg)
	}
}

func (r *Room) broadcastExcept(exclude *websocket.Conn, msg ServerMessage) {
	for conn := range r.clients {
		if conn != exclude {
			conn.WriteJSON(msg)
		}
	}
}
