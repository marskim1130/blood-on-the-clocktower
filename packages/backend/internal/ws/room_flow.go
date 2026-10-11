package ws

import "sync"

// roomFlow orders admission, committed projections and connection effects for
// one room. The map lock only protects queue bookkeeping, never persistence or
// delivery. Idle queues have neither a worker nor a retained map entry.
type roomFlow struct {
	mu     sync.Mutex
	queues map[string][]roomOperation
}

type roomOperation struct {
	run  func()
	done chan struct{}
}

func newRoomFlow() *roomFlow { return &roomFlow{queues: make(map[string][]roomOperation)} }

func (f *roomFlow) submit(roomID string, run func()) <-chan struct{} {
	done := make(chan struct{})
	f.mu.Lock()
	_, running := f.queues[roomID]
	f.queues[roomID] = append(f.queues[roomID], roomOperation{run: run, done: done})
	f.mu.Unlock()
	if !running {
		go f.drain(roomID)
	}
	return done
}

func (f *roomFlow) drain(roomID string) {
	for {
		f.mu.Lock()
		queue := f.queues[roomID]
		if len(queue) == 0 {
			delete(f.queues, roomID)
			f.mu.Unlock()
			return
		}
		op := queue[0]
		queue[0] = roomOperation{}
		f.queues[roomID] = queue[1:]
		f.mu.Unlock()
		op.run()
		close(op.done)
	}
}

// submitMessageV2 returns once a decoded request is admitted, not when it is
// committed. All callers and behavior tests use the same admission path.
func (h *Hub) submitMessageV2(conn Connection, msg ClientMessage) <-chan struct{} {
	if msg.Type == MsgCreateRoom {
		return h.submitCreateRoom(conn, msg)
	}
	roomID, playerID, generation, connected := h.active.Identity(conn)
	if !connected || roomID != msg.RoomID || playerID != msg.PlayerID {
		generation = 0
	}
	return h.rooms.submit(msg.RoomID, func() {
		h.dispatchRoomMessage(conn, msg, generation)
	})
}

func (h *Hub) submitCreateRoom(conn Connection, msg ClientMessage) <-chan struct{} {
	done := make(chan struct{})
	roomID := h.registry.CreatedRoomID(msg.RequestID)
	create := func() {
		result, err := h.createRoomV2(msg)
		if err != nil {
			h.sendV2Error(conn, err, 0)
			close(done)
			return
		}
		attach := func() { h.attachCreatedRoom(conn, result); close(done) }
		if roomID == result.RoomID {
			attach()
		} else {
			// A new room, including a retry after a preceding close deleted the
			// create alias, must attach through its own flow.
			h.rooms.submit(result.RoomID, attach)
		}
	}
	if roomID != "" {
		h.rooms.submit(roomID, create)
	} else {
		go create()
	}
	return done
}

func (h *Hub) handleMessageV2(conn Connection, msg ClientMessage) {
	<-h.submitMessageV2(conn, msg)
}

// attachConnection is only called from the room flow, after any preceding
// commit. Replaying an attachment on the same connection is idempotent.
func (h *Hub) attachConnection(roomID, playerID string, conn Connection) {
	_, previous := h.active.Takeover(roomID, playerID, conn)
	if previous != nil {
		h.outbound.remove(previous)
		_ = previous.Close()
	}
}
