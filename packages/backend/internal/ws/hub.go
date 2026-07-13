package ws

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/your-org/blood-on-the-clocktower/internal/session"
	"github.com/your-org/blood-on-the-clocktower/internal/sessionstore"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub adapts protocol v2 WebSocket messages to authoritative game sessions.
// Room membership, game state, persistence, and projections live behind Registry.
type Hub struct {
	registry *session.Registry
	active   *ActiveConnectionRegistry
	outbound *outboundDispatcher
}

// NewHub creates an in-memory protocol v2 hub for local development and tests.
func NewHub() *Hub {
	return newHubWithSessionStore(sessionstore.NewMemoryStore(), developmentCredentialKey())
}

func NewEphemeralProductionHub(credentialKey []byte) (*Hub, error) {
	if len(credentialKey) < 32 {
		return nil, fmt.Errorf("CLOCKTOWER_CREDENTIAL_KEY must be at least 32 bytes")
	}
	return newHubWithSessionStore(sessionstore.NewMemoryStore(), credentialKey), nil
}

func developmentCredentialKey() []byte {
	sum := sha256.Sum256([]byte("clocktower-development-credential-key"))
	return sum[:]
}

func newHubWithSessionStore(store session.Store, credentialKey []byte) *Hub {
	registry := session.NewRegistry(
		store,
		session.NewHMACCredentialCodec(credentialKey),
		func(data []byte) (session.Engine, error) { return loadSessionGameEngine("", data) },
		nil,
	)
	hub := &Hub{
		registry: registry,
		active:   NewActiveConnectionRegistry(),
	}
	hub.outbound = newOutboundDispatcher(func(connection Connection) { hub.active.Remove(connection) })
	return hub
}

func NewHubWithRoomRecordStore(store session.Store, credentialKey []byte) (*Hub, []error, error) {
	if len(credentialKey) < 32 {
		return nil, nil, fmt.Errorf("CLOCKTOWER_CREDENTIAL_KEY must be at least 32 bytes")
	}
	hub := newHubWithSessionStore(store, credentialKey)
	recoveryErrors, err := hub.registry.Restore(context.Background())
	if err != nil {
		return nil, nil, err
	}
	return hub, recoveryErrors, nil
}

func NewHubWithFileRoomRecords(directory string, credentialKey []byte) (*Hub, []error, error) {
	store, err := sessionstore.NewFileStore(directory)
	if err != nil {
		return nil, nil, err
	}
	return NewHubWithRoomRecordStore(store, credentialKey)
}

func NewHubWithRedisRoomRecords(redisURL, prefix string, credentialKey []byte) (*Hub, []error, error) {
	store, err := sessionstore.NewRedisStore(redisURL, prefix)
	if err != nil {
		return nil, nil, err
	}
	return NewHubWithRoomRecordStore(store, credentialKey)
}

func (h *Hub) Healthy() bool { return h.registry.Healthy() }

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsConnection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	connection := newWSConn(wsConnection)
	defer connection.Close()

	for {
		_, raw, err := connection.ReadMessage()
		if err != nil {
			h.handleDisconnect(connection)
			return
		}

		var message ClientMessage
		if err := json.Unmarshal(raw, &message); err != nil {
			h.outbound.send(connection, ServerMessage{Type: ServerMsgError, Code: ProtocolErrorInvalidMessage, Error: "invalid JSON message"})
			continue
		}
		if message.ProtocolVersion != 2 {
			h.outbound.send(connection, ServerMessage{Type: ServerMsgError, Code: ProtocolErrorUnsupportedProtocol, Error: "protocol version 2 is required"})
			continue
		}
		h.handleMessageV2(connection, message)
	}
}

func (h *Hub) handleDisconnect(connection Connection) {
	h.outbound.remove(connection)
	h.active.Remove(connection)
}
