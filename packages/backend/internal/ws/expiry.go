package ws

import (
	"context"
	"log"
	"time"
)

// CleanupInactiveRooms only notifies clients after authoritative durable deletion.
func (h *Hub) CleanupInactiveRooms(ctx context.Context, now time.Time) {
	removed, failures := h.registry.CleanupInactive(ctx, now)
	for _, err := range failures {
		log.Printf("inactive room cleanup failed: %v", err)
	}
	for _, roomID := range removed {
		for _, connection := range h.active.RemoveRoom(roomID) {
			h.outbound.sendAndClose(connection, ServerMessage{Type: ServerMsgRoomClosed, RoomID: roomID})
		}
	}
	if len(removed) > 0 {
		log.Printf("deleted %d rooms inactive for more than seven days", len(removed))
	}
}

// RunInactiveRoomCleanup is owned by the server process and stops with its context.
func (h *Hub) RunInactiveRoomCleanup(ctx context.Context) {
	h.CleanupInactiveRooms(ctx, time.Now())
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	recoveryTicker := time.NewTicker(time.Minute)
	defer recoveryTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			h.CleanupInactiveRooms(ctx, now)
		case <-recoveryTicker.C:
			h.recoveryMu.Lock()
			h.pruneRecoveryRequestsLocked()
			h.recoveryMu.Unlock()
		}
	}
}
