package ws

import (
	"context"
	"log"
	"time"
)

// CleanupInactiveRooms only notifies clients after authoritative durable deletion.
func (h *Hub) CleanupInactiveRooms(ctx context.Context, now time.Time) {
	roomIDs := h.registry.RoomIDs()
	type outcome struct {
		roomID  string
		removed bool
		err     error
	}
	results := make(chan outcome, len(roomIDs))
	for _, roomID := range roomIDs {
		roomID := roomID
		h.rooms.submit(roomID, func() {
			removed, err := h.registry.CleanupInactiveRoom(ctx, roomID, now)
			if removed && err == nil {
				for _, connection := range h.active.RemoveRoom(roomID) {
					h.outbound.sendAndClose(connection, ServerMessage{Type: ServerMsgRoomClosed, RoomID: roomID})
				}
			}
			results <- outcome{roomID: roomID, removed: removed, err: err}
		})
	}
	removedCount := 0
	for range roomIDs {
		result := <-results
		if result.err != nil {
			log.Printf("inactive room %s cleanup failed: %v", result.roomID, result.err)
		} else if result.removed {
			removedCount++
		}
	}
	if removedCount > 0 {
		log.Printf("deleted %d rooms inactive for more than seven days", removedCount)
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
		case now := <-recoveryTicker.C:
			h.cleanupRecoveryRequests(now)
		}
	}
}
