package ws

import (
	"context"
	"sort"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/session"
)

type pendingRecovery struct {
	request        RecoveryRequest
	roomID         string
	credential     string
	connection     Connection
	status         string
	reviewerID     string
	reviewSequence uint64
}

func (h *Hub) recoveryCredential(roomID, playerID string) (string, bool) {
	s, ok := h.registry.Get(roomID)
	if !ok {
		return "", false
	}
	return s.RecoveryCredentialFor(playerID)
}

func recoveryKey(roomID, requestID string) string { return roomID + ":" + requestID }

// Request contents are owned by the room flow; the global lock only protects
// the bounded index. Never hold it while executing a room operation.
func (h *Hub) recoveryRequestsForRoom(roomID string) map[string]*pendingRecovery {
	h.recoveryMu.Lock()
	defer h.recoveryMu.Unlock()
	requests := make(map[string]*pendingRecovery)
	for key, request := range h.recoveryRequests {
		if request.roomID == roomID {
			requests[key] = request
		}
	}
	return requests
}

func (h *Hub) addRecoveryRequest(key string, request *pendingRecovery) bool {
	h.recoveryMu.Lock()
	defer h.recoveryMu.Unlock()
	if len(h.recoveryRequests) >= 256 {
		return false
	}
	if h.recoveryRequests == nil {
		h.recoveryRequests = make(map[string]*pendingRecovery)
	}
	h.recoveryRequests[key] = request
	return true
}

func (h *Hub) pruneRecoveryRequests(roomID string, now time.Time) {
	for key, pending := range h.recoveryRequestsForRoom(roomID) {
		if now.Sub(time.UnixMilli(pending.request.RequestedAtUnixMs)) >= session.RecoveryRequestLifetime {
			if pending.status == "pending" {
				h.outbound.send(pending.connection, ServerMessage{Type: ServerMsgRecoveryStatus, RoomID: pending.roomID, PlayerID: pending.request.PlayerID, RecoveryRequestID: pending.request.RequestID, RecoveryStatus: "expired"})
			}
			h.recoveryMu.Lock()
			delete(h.recoveryRequests, key)
			h.recoveryMu.Unlock()
		}
	}
}

func (h *Hub) handleRequestRecovery(conn Connection, msg ClientMessage) {
	h.pruneRecoveryRequests(msg.RoomID, time.Now())
	if msg.RequestID == "" || len(msg.RequestID) > 128 || len(msg.RecoveryCredential) > 1024 {
		h.sendV2Error(conn, session.ErrInvalidCommand, 0)
		return
	}
	if _, _, _, active := h.active.Identity(conn); active {
		h.sendV2Error(conn, session.ErrForbidden, 0)
		return
	}
	s, ok := h.registry.Get(msg.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	if credential, approved := s.ApprovedRecovery(msg.RequestID, msg.PlayerID, msg.RecoveryCredential); approved {
		h.sendRecoveryApproval(conn, msg.RoomID, msg.PlayerID, msg.RequestID, credential)
		return
	}
	name, approver, err := s.ValidateRecovery(msg.PlayerID, msg.RecoveryCredential)
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	if approver == "" || approver == msg.PlayerID {
		h.sendV2Error(conn, session.ErrForbidden, 0)
		return
	}
	requests := h.recoveryRequestsForRoom(msg.RoomID)
	key := recoveryKey(msg.RoomID, msg.RequestID)
	pending := requests[key]
	if pending != nil && (pending.request.PlayerID != msg.PlayerID || pending.credential != msg.RecoveryCredential) {
		h.sendV2Error(conn, session.ErrIdempotencyConflict, 0)
		return
	}
	if pending == nil {
		for _, existing := range requests {
			if existing.roomID == msg.RoomID && existing.request.PlayerID == msg.PlayerID && existing.status == "pending" {
				h.sendV2Error(conn, session.ErrInvalidCommand, 0)
				return
			}
		}
		pending = &pendingRecovery{roomID: msg.RoomID, credential: msg.RecoveryCredential, request: RecoveryRequest{RequestID: msg.RequestID, PlayerID: msg.PlayerID, PlayerName: name, RequestedAtUnixMs: time.Now().UnixMilli()}, status: "pending"}
		if !h.addRecoveryRequest(key, pending) {
			h.sendV2Error(conn, session.ErrInvalidCommand, 0)
			return
		}
	}
	pending.connection = conn
	h.outbound.send(conn, ServerMessage{Type: ServerMsgRecoveryStatus, RoomID: msg.RoomID, PlayerID: msg.PlayerID, RecoveryRequestID: msg.RequestID, RecoveryStatus: pending.status})
	if target, _, connected := h.active.Get(msg.RoomID, approver); connected {
		h.sendRecoveryRequests(target, msg.RoomID, approver)
	}
}

func (h *Hub) sendRecoveryApproval(conn Connection, roomID, playerID, requestID, credential string) {
	recoveryCredential, _ := h.recoveryCredential(roomID, playerID)
	h.outbound.send(conn, ServerMessage{Type: ServerMsgRecoveryStatus, RoomID: roomID, PlayerID: playerID, RecoveryRequestID: requestID, RecoveryStatus: "approved", ResumeCredential: credential, RecoveryCredential: recoveryCredential})
}

func (h *Hub) sendRecoveryRequests(conn Connection, roomID, playerID string) {
	requests := []RecoveryRequest{}
	s, ok := h.registry.Get(roomID)
	if !ok {
		return
	}
	for _, pending := range h.recoveryRequestsForRoom(roomID) {
		if pending.roomID != roomID || pending.status != "pending" {
			continue
		}
		_, approver, err := s.ValidateRecovery(pending.request.PlayerID, pending.credential)
		if err == nil && approver == playerID {
			requests = append(requests, pending.request)
		}
	}
	sort.Slice(requests, func(i, j int) bool { return requests[i].RequestedAtUnixMs < requests[j].RequestedAtUnixMs })
	h.outbound.send(conn, ServerMessage{Type: ServerMsgRecoveryRequests, RoomID: roomID, RecoveryRequests: requests})
}

func (h *Hub) handleGetRecoveryRequests(conn Connection, msg ClientMessage) {
	h.pruneRecoveryRequests(msg.RoomID, time.Now())
	s, ok := h.registry.Get(msg.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	if _, err := s.Query(session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential}); err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	h.sendRecoveryRequests(conn, msg.RoomID, msg.PlayerID)
}

func (h *Hub) handleReviewRecovery(conn Connection, msg ClientMessage) {
	h.pruneRecoveryRequests(msg.RoomID, time.Now())
	pending := h.recoveryRequestsForRoom(msg.RoomID)[recoveryKey(msg.RoomID, msg.RecoveryRequestID)]
	if pending == nil || msg.Decision == nil {
		h.sendV2Error(conn, session.ErrInvalidCommand, 0)
		return
	}
	s, ok := h.registry.Get(msg.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	actor := session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential, ClientSequence: msg.ClientSequence}
	if pending.status != "pending" && (pending.reviewerID != msg.PlayerID || pending.reviewSequence != msg.ClientSequence || (pending.status == "approved") != *msg.Decision) {
		next := uint64(0)
		if query, err := s.Query(actor); err == nil {
			next = query.NextClientSequence
		}
		h.sendV2Error(conn, session.ErrInvalidCommand, next)
		return
	}
	command := session.Command{Kind: session.CommandReviewRecovery, TargetID: pending.request.PlayerID, Fingerprint: fingerprint(msg), RecoveryCredential: pending.credential, RecoveryRequestID: pending.request.RequestID, Decision: *msg.Decision}
	result, err := s.Execute(context.Background(), actor, command)
	if err != nil {
		next := uint64(0)
		if query, queryErr := s.Query(actor); queryErr == nil {
			next = query.NextClientSequence
		}
		h.sendV2Error(conn, err, next)
		return
	}
	h.enqueueCommandResult(conn, msg.RoomID, result)
	if !result.Duplicate {
		pending.reviewerID = msg.PlayerID
		pending.reviewSequence = msg.ClientSequence
		h.enqueueDeliveries(msg.RoomID, msg.PlayerID, result.RoomRevision, result.Deliveries, result.Metadata)
	}
	if *msg.Decision {
		credential, approved := s.ApprovedRecovery(pending.request.RequestID, pending.request.PlayerID, pending.credential)
		if !approved {
			h.sendV2Error(conn, session.ErrInvalidCredential, 0)
			return
		}
		pending.status = "approved"
		if !result.Duplicate {
			if old := h.active.Revoke(msg.RoomID, pending.request.PlayerID); old != nil {
				h.outbound.sendAndClose(old, ServerMessage{Type: ServerMsgSessionReplaced, RoomID: msg.RoomID})
			}
		}
		h.sendRecoveryApproval(pending.connection, msg.RoomID, pending.request.PlayerID, pending.request.RequestID, credential)
	} else {
		pending.status = "rejected"
		h.outbound.send(pending.connection, ServerMessage{Type: ServerMsgRecoveryStatus, RoomID: msg.RoomID, PlayerID: pending.request.PlayerID, RecoveryRequestID: pending.request.RequestID, RecoveryStatus: "rejected"})
	}
	h.sendRecoveryRequests(conn, msg.RoomID, msg.PlayerID)
}

func (h *Hub) cleanupRecoveryRequests(now time.Time) {
	h.recoveryMu.Lock()
	roomIDs := make(map[string]bool)
	for _, pending := range h.recoveryRequests {
		roomIDs[pending.roomID] = true
	}
	h.recoveryMu.Unlock()
	var completions []<-chan struct{}
	for roomID := range roomIDs {
		roomID := roomID
		completions = append(completions, h.rooms.submit(roomID, func() { h.pruneRecoveryRequests(roomID, now) }))
	}
	for _, done := range completions {
		<-done
	}
}
