package ws

import (
	"errors"
	"sync"
)

var errStaleActiveConnection = errors.New("stale active connection")

type connectionIdentity struct {
	RoomID   string
	PlayerID string
}

type activeConnection struct {
	Connection Connection
	Generation uint64
}

type ActiveConnectionRegistry struct {
	mu          sync.RWMutex
	connections map[connectionIdentity]activeConnection
	identities  map[Connection]connectionIdentity
	generations map[connectionIdentity]uint64
}

func NewActiveConnectionRegistry() *ActiveConnectionRegistry {
	return &ActiveConnectionRegistry{
		connections: make(map[connectionIdentity]activeConnection),
		identities:  make(map[Connection]connectionIdentity),
		generations: make(map[connectionIdentity]uint64),
	}
}

func (r *ActiveConnectionRegistry) Takeover(roomID, playerID string, connection Connection) (uint64, Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()

	identity := connectionIdentity{RoomID: roomID, PlayerID: playerID}
	generation := r.generations[identity] + 1
	r.generations[identity] = generation

	previous := r.connections[identity].Connection
	if previous != nil {
		delete(r.identities, previous)
	}
	r.connections[identity] = activeConnection{Connection: connection, Generation: generation}
	r.identities[connection] = identity
	return generation, previous
}

func (r *ActiveConnectionRegistry) IsCurrent(roomID, playerID string, connection Connection, generation uint64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	current, ok := r.connections[connectionIdentity{RoomID: roomID, PlayerID: playerID}]
	return ok && current.Connection == connection && current.Generation == generation
}

func (r *ActiveConnectionRegistry) WithCurrent(roomID, playerID string, connection Connection, action func() error) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	current, ok := r.connections[connectionIdentity{RoomID: roomID, PlayerID: playerID}]
	if !ok || current.Connection != connection {
		return errStaleActiveConnection
	}
	return action()
}

func (r *ActiveConnectionRegistry) Identity(connection Connection) (string, string, uint64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	identity, ok := r.identities[connection]
	if !ok {
		return "", "", 0, false
	}
	current := r.connections[identity]
	return identity.RoomID, identity.PlayerID, current.Generation, true
}

func (r *ActiveConnectionRegistry) Remove(connection Connection) (string, string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	identity, ok := r.identities[connection]
	if !ok {
		return "", "", false
	}
	delete(r.identities, connection)
	current := r.connections[identity]
	if current.Connection == connection {
		delete(r.connections, identity)
	}
	return identity.RoomID, identity.PlayerID, true
}

func (r *ActiveConnectionRegistry) Get(roomID, playerID string) (Connection, uint64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	current, ok := r.connections[connectionIdentity{RoomID: roomID, PlayerID: playerID}]
	return current.Connection, current.Generation, ok
}

func (r *ActiveConnectionRegistry) Room(roomID string) map[string]Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	connections := make(map[string]Connection)
	for identity, active := range r.connections {
		if identity.RoomID == roomID {
			connections[identity.PlayerID] = active.Connection
		}
	}
	return connections
}

func (r *ActiveConnectionRegistry) RemoveRoom(roomID string) []Connection {
	r.mu.Lock()
	defer r.mu.Unlock()
	var connections []Connection
	for identity, active := range r.connections {
		if identity.RoomID != roomID {
			continue
		}
		connections = append(connections, active.Connection)
		delete(r.identities, active.Connection)
		delete(r.connections, identity)
	}
	return connections
}
