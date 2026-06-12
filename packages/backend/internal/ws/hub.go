package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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
	rm            *RoomManager
	sessions      map[string]*GameSession // roomID -> GameSession
	mu            sync.RWMutex
	connToRoom    map[Connection]string // conn -> roomID (for disconnect lookup)
	snapshotStore snapshotStore
}

func NewHub() *Hub {
	return newHub()
}

func newHub() *Hub {
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
	case MsgCreateRoom:
		h.handleCreateRoom(conn, msg)
	case MsgJoinRoom:
		h.handleJoinRoom(conn, msg)
	case MsgLeaveRoom:
		h.handleLeaveRoom(conn, msg)
	case MsgKickPlayer:
		h.handleKickPlayer(conn, msg)
	case MsgUpdateRoomSettings:
		h.handleUpdateRoomSettings(conn, msg)
	case MsgSetStoryteller:
		h.handleSetStoryteller(conn, msg)
	case MsgAssignCharacters:
		h.handleAssignCharacters(conn, msg)
	case MsgSubmitEvent:
		h.handleSubmitEvent(conn, msg)
	case MsgStartGame:
		h.handleStartGame(conn, msg)
	case MsgChangePhase:
		h.handleChangePhase(conn, msg)
	case MsgNominate:
		h.handleNominate(conn, msg)
	case MsgCastVote:
		h.handleCastVote(conn, msg)
	case MsgResolveNomination:
		h.handleResolveNomination(conn, msg)
	case MsgExecutePlayer:
		h.handleExecutePlayer(conn, msg)
	case MsgSubmitNightAction:
		h.handleSubmitNightAction(conn, msg)
	case MsgResolveNight:
		h.handleResolveNight(conn, msg)
	case MsgEndGame:
		h.handleEndGame(conn, msg)
	}
}

func (h *Hub) handleCreateRoom(conn Connection, msg ClientMessage) {
	scriptID := msg.ScriptID
	if scriptID == "" {
		scriptID = game.TroubleBrewingScriptID
	}
	if game.GetScriptByID(scriptID) == nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "unsupported script"})
		return
	}

	room := h.rm.CreateRoom(msg.PlayerID, msg.MaxPlayers, scriptID)

	room.mu.Lock()
	room.addClient(conn, msg.PlayerID)
	room.mu.Unlock()

	gs := NewGameSession(scriptID)
	gs.AddPlayer(game.Player{ID: msg.PlayerID, Name: msg.PlayerName, IsAlive: true})

	h.mu.Lock()
	h.sessions[room.id] = gs
	h.connToRoom[conn] = room.id
	h.mu.Unlock()
	h.persistSnapshot()

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

	addedParticipant := true
	if gs != nil {
		added, err := gs.AddOrReconnectPlayer(game.Player{ID: msg.PlayerID, Name: msg.PlayerName, IsAlive: true}, h.rm.MaxPlayers(msg.RoomID))
		if err != nil {
			h.rm.RemoveClient(msg.RoomID, msg.PlayerID)
			h.mu.Lock()
			delete(h.connToRoom, conn)
			h.mu.Unlock()
			conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
			return
		}
		addedParticipant = added
	}
	if gs != nil {
		h.persistSnapshot()
	}

	conn.SendJSON(ServerMessage{
		Type:   "ROOM_STATE",
		RoomID: msg.RoomID,
		State:  h.buildRoomStateForRecipient(msg.RoomID, msg.PlayerID),
	})

	if addedParticipant {
		h.BroadcastExcept(msg.RoomID, msg.PlayerID, ServerMessage{
			Type: "EVENT_BROADCAST",
			Event: &game.GameEvent{
				PlayerJoined: &game.PlayerJoined{
					Player: game.Player{ID: msg.PlayerID, Name: msg.PlayerName, IsAlive: true},
				},
			},
		})
	}
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

	shouldDestroy := gs == nil || gs.ParticipantCount() == 0
	if shouldDestroy {
		h.rm.DestroyRoom(roomID)
		h.mu.Lock()
		delete(h.sessions, roomID)
		h.mu.Unlock()
	}
	h.persistSnapshot()
}

func (h *Hub) handleKickPlayer(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}
	if senderID != h.rm.CreatorID(roomID) {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only room creator can kick players"})
		return
	}

	result, err := gs.Apply(KickPlayerCmd{
		SenderID:       senderID,
		TargetPlayerID: msg.TargetPlayerID,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	kickedClient, err := h.rm.KickPlayer(roomID, msg.TargetPlayerID)
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}
	if kickedClient != nil {
		h.mu.Lock()
		delete(h.connToRoom, kickedClient.Conn)
		h.mu.Unlock()
		if err := kickedClient.Conn.SendJSON(ServerMessage{Type: "ERROR", Error: "kicked from room"}); err != nil {
			log.Printf("notify kicked player %s in room %s failed: %v", msg.TargetPlayerID, roomID, err)
		}
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleUpdateRoomSettings(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}
	if senderID != h.rm.CreatorID(roomID) {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only room creator can update room settings"})
		return
	}

	result, err := gs.Apply(UpdateRoomSettingsCmd{
		SenderID:   senderID,
		MaxPlayers: msg.MaxPlayers,
		ScriptID:   msg.ScriptID,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	scriptID := strings.TrimSpace(msg.ScriptID)
	if err := h.rm.UpdateRoomSettings(roomID, msg.MaxPlayers, scriptID); err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
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
		h.commitRoomUpdate(roomID, result)
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
		h.commitRoomUpdate(roomID, result)
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

func (h *Hub) handleStartGame(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	result, err := gs.Apply(StartGameCmd{SenderID: senderID})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleChangePhase(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	// Storyteller-only command.
	if senderID != gs.StorytellerID() {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only the storyteller can change phase"})
		return
	}

	result, err := gs.Apply(ChangePhaseCmd{
		SenderID: senderID,
		Phase:    msg.Phase.GamePhase(),
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleNominate(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	result, err := gs.Apply(NominateCmd{
		SenderID:  senderID,
		NomineeID: msg.NomineeID,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleCastVote(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	if msg.Decision == nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "decision is required"})
		return
	}

	result, err := gs.Apply(CastVoteCmd{
		SenderID: senderID,
		Decision: *msg.Decision,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleResolveNomination(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	// Storyteller-only command.
	if senderID != gs.StorytellerID() {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only the storyteller can resolve a nomination"})
		return
	}

	result, err := gs.Apply(ResolveNominationCmd{SenderID: senderID})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleExecutePlayer(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	// Storyteller-only command.
	if senderID != gs.StorytellerID() {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only the storyteller can execute a player"})
		return
	}

	result, err := gs.Apply(ExecutePlayerCmd{
		SenderID: senderID,
		PlayerID: msg.targetPlayerID(),
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleSubmitNightAction(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	result, err := gs.Apply(SubmitNightActionCmd{
		SenderID:   senderID,
		ActionType: msg.ActionType,
		TargetIDs:  msg.TargetIDs,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		if event.NightActionSubmitted != nil {
			h.sendNightActionSubmitted(roomID, gs.StorytellerID(), event.NightActionSubmitted)
			continue
		}
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleResolveNight(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	// Storyteller-only command.
	if senderID != gs.StorytellerID() {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "only the storyteller can resolve the night"})
		return
	}

	result, err := gs.Apply(ResolveNightCmd{SenderID: senderID})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleEndGame(conn Connection, msg ClientMessage) {
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

	senderID := h.rm.GetPlayerByConn(roomID, conn)
	if senderID == "" {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: "not in any room"})
		return
	}

	result, err := gs.Apply(EndGameCmd{
		SenderID:    senderID,
		Winner:      msg.Winner.Team(),
		Reason:      game.WinReason(msg.Reason),
		Description: msg.Description,
	})
	if err != nil {
		conn.SendJSON(ServerMessage{Type: "ERROR", Error: err.Error()})
		return
	}

	for _, event := range result.Events {
		eventCopy := event
		h.Broadcast(roomID, ServerMessage{Type: "EVENT_BROADCAST", Event: &eventCopy})
	}

	if result.Updated {
		h.commitRoomUpdate(roomID, result)
	}
}

func (h *Hub) handleDisconnect(conn Connection) {
	// Always clean up connToRoom, even if RemoveClientByConn fails
	h.mu.Lock()
	delete(h.connToRoom, conn)
	h.mu.Unlock()

	_, _, err := h.rm.RemoveClientByConn(conn)
	if err != nil {
		return
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

func (h *Hub) commitRoomUpdate(roomID string, result ApplyResult) {
	if !result.Updated {
		return
	}
	if h.roomPhase(roomID) == game.GamePhaseFinished {
		h.BroadcastRoomState(roomID)
		h.destroyFinishedRoom(roomID)
		h.persistSnapshot()
		return
	}

	h.persistSnapshot()
	h.BroadcastRoomState(roomID)
}

func (h *Hub) roomPhase(roomID string) game.GamePhase {
	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()
	if gs == nil {
		return game.GamePhaseUnspecified
	}
	return gs.Phase()
}

func (h *Hub) destroyFinishedRoom(roomID string) {
	clients := h.rm.GetClientsByRoom(roomID)

	h.rm.DestroyRoom(roomID)
	h.mu.Lock()
	delete(h.sessions, roomID)
	for _, client := range clients {
		delete(h.connToRoom, client.Conn)
	}
	h.mu.Unlock()
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

func (h *Hub) sendNightActionSubmitted(roomID, storytellerID string, action *game.NightActionEvent) {
	clients := h.rm.GetClientsByRoom(roomID)

	actionEvent := game.GameEvent{NightActionSubmitted: action}
	if actorClient := clients[action.ActorID]; actorClient != nil {
		if err := actorClient.Conn.SendJSON(ServerMessage{
			Type:  "EVENT_BROADCAST",
			Event: &actionEvent,
		}); err != nil {
			log.Printf("send night action to actor %s in room %s failed: %v", action.ActorID, roomID, err)
		}
	}

	if storytellerID != "" && storytellerID != action.ActorID {
		storytellerEvent := game.GameEvent{NightActionSubmitted: action}
		if storytellerClient := clients[storytellerID]; storytellerClient != nil {
			if err := storytellerClient.Conn.SendJSON(ServerMessage{
				Type:  "EVENT_BROADCAST",
				Event: &storytellerEvent,
			}); err != nil {
				log.Printf("send night action to storyteller %s in room %s failed: %v", storytellerID, roomID, err)
			}
		}
	}
}

func (h *Hub) buildRoomStateForRecipient(roomID, recipientID string) *RoomState {
	h.mu.RLock()
	gs := h.sessions[roomID]
	h.mu.RUnlock()

	if gs == nil {
		state := &RoomState{RoomID: roomID, MaxPlayers: h.rm.MaxPlayers(roomID)}
		state.CreatorID = h.rm.CreatorID(roomID)
		applyScriptToRoomState(state, h.rm.ScriptID(roomID))
		return state
	}

	state := gs.StateForRoomForRecipient(roomID, recipientID)
	state.MaxPlayers = h.rm.MaxPlayers(roomID)
	state.CreatorID = h.rm.CreatorID(roomID)
	applyScriptToRoomState(state, h.rm.ScriptID(roomID))
	return state
}

func applyScriptToRoomState(state *RoomState, scriptID string) {
	if state == nil {
		return
	}
	if scriptID == "" {
		scriptID = game.TroubleBrewingScriptID
	}
	state.ScriptID = scriptID
	if script := game.GetScriptByID(scriptID); script != nil {
		state.ScriptName = script.Name
	}
}
