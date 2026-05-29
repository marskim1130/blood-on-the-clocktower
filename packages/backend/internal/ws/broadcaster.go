package ws

// Broadcaster is the seam for message delivery.
// Hub implements this; Room/GameSession trigger events through it.
type Broadcaster interface {
	Broadcast(roomID string, msg ServerMessage)
	BroadcastExcept(roomID string, excludePlayerID string, msg ServerMessage)
	SendTo(playerID string, msg ServerMessage)
}
