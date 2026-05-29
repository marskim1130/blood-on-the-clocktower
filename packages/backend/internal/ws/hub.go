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

// Hub is the transport layer: WebSocket upgrade, message routing, broadcast dispatch.
// Implements Broadcaster. No game logic, no room state.
type Hub struct {
	rm         *RoomManager
	sessions   map[string]*GameSession // roomID -> GameSession
	mu         sync.RWMutex
	connToRoom map[Connection]string // conn -> roomID (for disconnect lookup)
}

func NewHub() *Hub {
	return &Hub{
		rm:         NewRoomManager(),
		sessions:   make(map[string]*GameSession),
		connToRoom: make(map[Connection]string),
	}
}

// HandleWebSocket handles WebSocket upgrade and message routing.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	conn := newWSConn(wsConn)
	defer conn.Close()

	for {
		_, raw, err := wsConn.ReadMessage()
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

func (h *Hub) handleMessage(conn Connection, msg ClientMessage) {
	switch msg.Type {
	case "CREATE_ROOM":
		h.handleCreateRoom(conn, msg)
	case "JOIN_ROOM":
		h.handleJoinRoom(conn, msg)
	case "LEAVE_ROOM":
		h.handleLeaveRoom(conn, msg)
	case "SET_STORYTELLER":
		h.handleSetStoryteller(conn, msg)
	case "ASSIGN_CHARACTERS":
		h.handleAssignCharacters(conn, msg)
	case "SUBMIT_EVENT":
		h.handleSubmitEvent(conn, msg)
	}
}

func (h *Hub) handleCreateRoom(conn Connection, msg ClientMessage) {
	room := h.rm.CreateRoom(msg.PlayerID, msg.MaxPlayers)

	room.mu.Lock()
	room.clients[msg.PlayerID] = &Client{Conn: conn, PlayerID: msg.PlayerID}
	room.mu.Unlock()

	gs := NewGameSession()
	gs.AddPlayer(game.Player{ID: msg.PlayerID, Name: msg.PlayerName, IsAlive: true})

	h.mu.Lock()
	h.sessions[room.id] = gs
	h.connToRoom[conn] = room.id
	h.mu.Unlock()

	conn.SendJSON(ServerMessage{
		Type:   "ROOM_STATE",
		RoomID: room.id,
		State:  h.buildRoomState(room.id),
	})
}

func (h *Hub) handleJoinRoom(conn Connection, msg ClientMessage) {
	err := h.rm.JoinRoom(msg.RoomID, conn, msg.PlayerID, msg.PlayerName)
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	h.mu.Lock()
	h.connToRoom[conn] = msg.RoomID
	gs := h.sessions[msg.RoomID]
	h.mu.Unlock()

	if gs != nil {
		gs.AddPlayer(game.Player{ID: msg.PlayerID, Name: msg.PlayerName, IsAlive: true})
	}

	conn.SendJSON(ServerMessage{
		Type:   "ROOM_STATE",
		RoomID: msg.RoomID,
		State:  h.buildRoomState(msg.RoomID),
	})

	h.BroadcastExcept(msg.RoomID, msg.PlayerID, ServerMessage{
		Type: "EVENT_BROADCAST",
		Event: &game.GameEvent{
			PlayerJoined: &game.PlayerJoined{
				Player: game.Player{ID: msg.PlayerID, Name: msg.PlayerName, IsAlive: true},
			},
		},
	})
}

func (h *Hub) handleLeaveRoom(conn Connection, msg ClientMessage) {
	roomID, err := h.rm.LeaveRoom(msg.PlayerID)
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	h.mu.Lock()
	delete(h.connToRoom, conn)
	gs := h.sessions[roomID]
	h.mu.Unlock()

	if gs != nil {
		gs.RemovePlayer(msg.PlayerID)
	}

	h.Broadcast(roomID, ServerMessage{
		Type: "EVENT_BROADCAST",
		Event: &game.GameEvent{
			PlayerLeft: &game.PlayerLeft{PlayerID: msg.PlayerID},
		},
	})

	// Clean up session if room destroyed
	if h.rm.GetRoom(roomID) == nil {
		h.mu.Lock()
		delete(h.sessions, roomID)
		h.mu.Unlock()
	}
}

func (h *Hub) handleSetStoryteller(conn Connection, msg ClientMessage) {
	roomID := h.connToRoom[conn]
	creatorID := h.rm.CreatorID(roomID)

	senderID := ""
	if client, rid := h.rm.GetClient(msg.PlayerID); client != nil {
		senderID = msg.PlayerID
		_ = rid
	}
	_ = senderID

	// Verify sender is the creator
	if msg.PlayerID != creatorID {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only room creator can set storyteller"})
		return
	}

	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if gs == nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "no game session"})
		return
	}

	result, err := gs.Apply(SetStorytellerCmd{
		SenderID:       msg.PlayerID,
		TargetPlayerID: msg.TargetPlayerID,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	if result.Updated {
		state := h.buildRoomState(roomID)
		h.Broadcast(roomID, ServerMessage{
			Type:   "ROOM_STATE",
			RoomID: roomID,
			State:  state,
		})
	}
}

func (h *Hub) handleAssignCharacters(conn Connection, msg ClientMessage) {
	roomID := h.connToRoom[conn]

	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if gs == nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "no game session"})
		return
	}

	result, err := gs.Apply(AssignCharactersCmd{
		SenderID:   msg.PlayerID,
		Assignments: msg.Assignments,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{
			Type:  "EVENT_BROADCAST",
			Event: &eventCopy,
		})
	}
}

func (h *Hub) handleSubmitEvent(conn Connection, msg ClientMessage) {
	if msg.Event == nil {
		return
	}

	roomID := h.connToRoom[conn]

	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if gs == nil {
		return
	}

	result, err := gs.Apply(SubmitEventCmd{
		SenderID: msg.PlayerID,
		Event:    *msg.Event,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{
			Type:  "EVENT_BROADCAST",
			Event: &eventCopy,
		})
	}
}

func (h *Hub) handleDisconnect(conn Connection) {
	roomID, playerID, err := h.rm.RemoveClientByConn(conn)
	if err != nil {
		return
	}

	h.mu.Lock()
	delete(h.connToRoom, conn)
	gs := h.sessions[roomID]
	h.mu.Unlock()

	if gs != nil {
		gs.RemovePlayer(playerID)
	}

	h.Broadcast(roomID, ServerMessage{
		Type: "EVENT_BROADCAST",
		Event: &game.GameEvent{
			PlayerLeft: &game.PlayerLeft{PlayerID: playerID},
		},
	})

	if h.rm.GetRoom(roomID) == nil {
		h.mu.Lock()
		delete(h.sessions, roomID)
		h.mu.Unlock()
	}
}

// --- Broadcaster implementation ---

func (h *Hub) Broadcast(roomID string, msg ServerMessage) {
	clients := h.rm.GetClientsByRoom(roomID)
	for _, client := range clients {
		client.Conn.SendJSON(msg)
	}
}

func (h *Hub) BroadcastExcept(roomID string, excludePlayerID string, msg ServerMessage) {
	clients := h.rm.GetClientsByRoom(roomID)
	for pid, client := range clients {
		if pid != excludePlayerID {
			client.Conn.SendJSON(msg)
		}
	}
}

func (h *Hub) SendTo(playerID string, msg ServerMessage) {
	client, _ := h.rm.GetClient(playerID)
	if client != nil {
		client.Conn.SendJSON(msg)
	}
}

func (h *Hub) buildRoomState(roomID string) *RoomState {
	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if gs == nil {
		return &RoomState{RoomID: roomID}
	}

	state := gs.StateForRoom(roomID)
	state.MaxPlayers = h.rm.MaxPlayers(roomID)
	return state
}
