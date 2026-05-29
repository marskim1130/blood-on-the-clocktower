package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/your-org/blood-on-the-clocktower/internal/game"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub manages all rooms and WebSocket connections
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

// Room represents a game room with connected clients
type Room struct {
	mu      sync.RWMutex
	id      string
	clients map[*websocket.Conn]string // conn -> playerID
	players []game.Player
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
	case "JOIN_ROOM":
		h.handleJoinRoom(conn, msg)
	case "SUBMIT_EVENT":
		h.handleSubmitEvent(conn, msg)
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

func (h *Hub) handleJoinRoom(conn *websocket.Conn, msg ClientMessage) {
	h.mu.Lock()
	room, exists := h.rooms[msg.RoomID]
	if !exists {
		room = &Room{
			id:      msg.RoomID,
			clients: make(map[*websocket.Conn]string),
		}
		h.rooms[msg.RoomID] = room
	}
	h.mu.Unlock()

	room.mu.Lock()
	player := game.Player{
		ID:      msg.PlayerID,
		Name:    msg.PlayerName,
		IsAlive: true,
	}
	room.players = append(room.players, player)
	room.clients[conn] = msg.PlayerID
	room.mu.Unlock()

	// Send ROOM_STATE to the joining player
	state := ServerMessage{
		Type:   "ROOM_STATE",
		RoomID: room.id,
		State:  room.getState(),
	}
	conn.WriteJSON(state)

	// Broadcast PLAYER_JOINED to other room members
	event := game.GameEvent{
		PlayerJoined: &game.PlayerJoined{Player: player},
	}
	room.broadcastExcept(conn, ServerMessage{
		Type:  "EVENT_BROADCAST",
		Event: &event,
	})
}

func (h *Hub) handleDisconnect(conn *websocket.Conn) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, room := range h.rooms {
		room.mu.Lock()
		if playerID, ok := room.clients[conn]; ok {
			delete(room.clients, conn)
			// Remove player from room
			for i, p := range room.players {
				if p.ID == playerID {
					room.players = append(room.players[:i], room.players[i+1:]...)
					break
				}
			}
			// Broadcast player left
			event := game.GameEvent{
				PlayerLeft: &game.PlayerLeft{PlayerID: playerID},
			}
			room.broadcast(ServerMessage{
				Type:  "EVENT_BROADCAST",
				Event: &event,
			})
		}
		room.mu.Unlock()
	}
}

func (r *Room) getState() *RoomState {
	return &RoomState{
		RoomID:  r.id,
		Players: r.players,
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
