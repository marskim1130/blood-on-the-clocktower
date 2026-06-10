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
	room.addClient(conn, msg.PlayerID)
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
		State:  h.buildRoomStateForRecipient(room.id, msg.PlayerID),
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
		State:  h.buildRoomStateForRecipient(msg.RoomID, msg.PlayerID),
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

func (h *Hub) handleLeaveRoom(conn Connection, _ ClientMessage) {
	// Derive room and player identity from the connection, not from the message.
	h.mu.RLock()
	roomID := h.connToRoom[conn]
	h.mu.RUnlock()

	if roomID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	playerID := h.rm.GetPlayerByConn(roomID, conn)
	if playerID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	_, err := h.rm.LeaveRoom(playerID)
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
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

	// Clean up session if room destroyed
	if h.rm.GetRoom(roomID) == nil {
		h.mu.Lock()
		delete(h.sessions, roomID)
		h.mu.Unlock()
	}
}

func (h *Hub) handleSetStoryteller(conn Connection, msg ClientMessage) {
	h.mu.RLock()
	roomID := h.connToRoom[conn]
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if roomID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	// Derive sender identity from the connection, not from the message
	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	creatorID := h.rm.CreatorID(roomID)
	if senderID != creatorID {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only room creator can set storyteller"})
		return
	}

	if gs == nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "no game session"})
		return
	}

	result, err := gs.Apply(SetStorytellerCmd{
		SenderID:       senderID,
		TargetPlayerID: msg.TargetPlayerID,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	if result.Updated {
		h.BroadcastRoomState(roomID)
	}
}

func (h *Hub) handleAssignCharacters(conn Connection, msg ClientMessage) {
	h.mu.RLock()
	roomID := h.connToRoom[conn]
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if roomID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	if gs == nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "no game session"})
		return
	}

	// Derive sender identity from the connection, not from the message
	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	result, err := gs.Apply(AssignCharactersCmd{
		SenderID:    senderID,
		Assignments: msg.Assignments,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		if event.CharacterAssigned != nil {
			h.sendCharacterAssignment(roomID, senderID, event.CharacterAssigned)
			continue
		}
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{
			Type:  "EVENT_BROADCAST",
			Event: &eventCopy,
		})
	}

	if result.Updated {
		h.BroadcastRoomState(roomID)
	}
}

func (h *Hub) handleSubmitEvent(conn Connection, msg ClientMessage) {
	if msg.Event == nil {
		return
	}

	h.mu.RLock()
	roomID := h.connToRoom[conn]
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if roomID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	if gs == nil {
		return
	}

	// Derive sender identity from the connection, not from the message
	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	result, err := gs.Apply(SubmitEventCmd{
		SenderID: senderID,
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
	// Always clean up connToRoom, even if RemoveClientByConn fails
	h.mu.Lock()
	delete(h.connToRoom, conn)
	h.mu.Unlock()

	roomID, playerID, err := h.rm.RemoveClientByConn(conn)
	if err != nil {
		return
	}

	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()

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
	for pid, client := range clients {
		if err := client.Conn.SendJSON(msg); err != nil {
			log.Printf("broadcast to %s in room %s failed: %v", pid, roomID, err)
		}
	}
}

func (h *Hub) BroadcastExcept(roomID string, excludePlayerID string, msg ServerMessage) {
	clients := h.rm.GetClientsByRoom(roomID)
	for pid, client := range clients {
		if pid != excludePlayerID {
			if err := client.Conn.SendJSON(msg); err != nil {
				log.Printf("broadcast to %s in room %s failed: %v", pid, roomID, err)
			}
		}
	}
}

func (h *Hub) SendTo(playerID string, msg ServerMessage) {
	client, _ := h.rm.GetClient(playerID)
	if client != nil {
		if err := client.Conn.SendJSON(msg); err != nil {
			log.Printf("send to player %s failed: %v", playerID, err)
		}
	}
}

func (h *Hub) BroadcastRoomState(roomID string) {
	clients := h.rm.GetClientsByRoom(roomID)
	for pid, client := range clients {
		msg := ServerMessage{
			Type:   "ROOM_STATE",
			RoomID: roomID,
			State:  h.buildRoomStateForRecipient(roomID, pid),
		}
		if err := client.Conn.SendJSON(msg); err != nil {
			log.Printf("room state to %s in room %s failed: %v", pid, roomID, err)
		}
	}
}

func (h *Hub) sendCharacterAssignment(roomID, storytellerID string, assignment *game.CharacterAssigned) {
	clients := h.rm.GetClientsByRoom(roomID)

	playerEvent := game.GameEvent{CharacterAssigned: assignment}
	if playerClient := clients[assignment.PlayerID]; playerClient != nil {
		if err := playerClient.Conn.SendJSON(ServerMessage{
			Type:  "EVENT_BROADCAST",
			Event: &playerEvent,
		}); err != nil {
			log.Printf("send assignment to player %s in room %s failed: %v", assignment.PlayerID, roomID, err)
		}
	}

	if storytellerID != "" && storytellerID != assignment.PlayerID {
		storytellerEvent := game.GameEvent{CharacterAssigned: assignment}
		if storytellerClient := clients[storytellerID]; storytellerClient != nil {
			if err := storytellerClient.Conn.SendJSON(ServerMessage{
				Type:  "EVENT_BROADCAST",
				Event: &storytellerEvent,
			}); err != nil {
				log.Printf("send assignment to storyteller %s in room %s failed: %v", storytellerID, roomID, err)
			}
		}
	}
}

func (h *Hub) buildRoomStateForRecipient(roomID, recipientID string) *RoomState {
	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if gs == nil {
		return &RoomState{RoomID: roomID, MaxPlayers: h.rm.MaxPlayers(roomID)}
	}

	state := gs.StateForRoomForRecipient(roomID, recipientID)
	state.MaxPlayers = h.rm.MaxPlayers(roomID)
	return state
}
