package ws

import (
	"github.com/marskim1130/blood-on-the-clocktower/internal/sessionstore"
	"testing"
	"time"
)

func TestRecoveryApprovalReplayDoesNotDisconnectAlreadyRecoveredDevice(t *testing.T) {
	hub := NewHub()
	owner, created := createContractRoom(t, hub, 5)
	_ = joinContractPlayer(t, hub, created.RoomID, "p1")
	s, _ := hub.registry.Get(created.RoomID)
	code, _ := s.RecoveryCredentialFor("p1")
	device := newFakeConnection()
	hub.handleMessageV2(device, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "edge-replay", RecoveryCredential: code})
	_ = waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgRecoveryStatus })
	yes := true
	approval := ClientMessage{Type: MsgReviewRecovery, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: created.NextClientSequence, RecoveryRequestID: "edge-replay", Decision: &yes}
	hub.handleMessageV2(owner, approval)
	granted := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "approved" })
	hub.handleMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "p1", ResumeCredential: granted.ResumeCredential})
	_ = waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	hub.handleMessageV2(owner, approval)
	active, _, ok := hub.active.Get(created.RoomID, "p1")
	if !ok || active != device {
		t.Fatal("replaying an approved command disconnected the recovered device")
	}
}

func TestRejectedRecoveryCannotBeApprovedWithANewSequence(t *testing.T) {
	hub := NewHub()
	owner, created := createContractRoom(t, hub, 5)
	_ = joinContractPlayer(t, hub, created.RoomID, "p1")
	s, _ := hub.registry.Get(created.RoomID)
	code, _ := s.RecoveryCredentialFor("p1")
	device := newFakeConnection()
	hub.handleMessageV2(device, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "edge-rejected", RecoveryCredential: code})
	_ = waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "pending" })
	no := false
	hub.handleMessageV2(owner, ClientMessage{Type: MsgReviewRecovery, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: created.NextClientSequence, RecoveryRequestID: "edge-rejected", Decision: &no})
	_ = waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "rejected" })
	yes := true
	hub.handleMessageV2(owner, ClientMessage{Type: MsgReviewRecovery, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: created.NextClientSequence + 1, RecoveryRequestID: "edge-rejected", Decision: &yes})
	if _, _, err := s.ValidateRecovery("p1", code); err != nil {
		t.Fatal("rejected request was approved with a new sequence")
	}
}

func TestApprovedRecoveryCanReplayAfterHubRestartWithoutLeakingProjection(t *testing.T) {
	store := sessionstore.NewMemoryStore()
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	_ = joinContractPlayer(t, hub, created.RoomID, "p1")
	s, _ := hub.registry.Get(created.RoomID)
	code, _ := s.RecoveryCredentialFor("p1")
	device := newFakeConnection()
	request := ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "edge-restart", RecoveryCredential: code}
	hub.handleMessageV2(device, request)
	_ = waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "pending" })
	yes := true
	hub.handleMessageV2(owner, ClientMessage{Type: MsgReviewRecovery, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: created.NextClientSequence, RecoveryRequestID: request.RequestID, Decision: &yes})
	granted := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "approved" })
	restarted, failures, err := NewHubWithRoomRecordStore(store, developmentCredentialKey())
	if err != nil || len(failures) != 0 {
		t.Fatalf("restart=%v %v", err, failures)
	}
	retry := newFakeConnection()
	restarted.handleMessageV2(retry, request)
	replayed := waitForContractMessage(t, retry, func(m ServerMessage) bool { return m.RecoveryStatus == "approved" })
	if replayed.ResumeCredential != granted.ResumeCredential || replayed.RecoveryCredential != granted.RecoveryCredential || replayed.State != nil {
		t.Fatal("dropped approval replay changed credentials or leaked projection")
	}
}

func TestExpiredRecoveryDoesNotRotateCredentials(t *testing.T) {
	hub := NewHub()
	owner, created := createContractRoom(t, hub, 5)
	_ = joinContractPlayer(t, hub, created.RoomID, "p1")
	s, _ := hub.registry.Get(created.RoomID)
	code, _ := s.RecoveryCredentialFor("p1")
	device := newFakeConnection()
	hub.handleMessageV2(device, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "edge-expired", RecoveryCredential: code})
	_ = waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "pending" })
	hub.recoveryMu.Lock()
	hub.recoveryRequests[recoveryKey(created.RoomID, "edge-expired")].request.RequestedAtUnixMs = time.Now().Add(-11 * time.Minute).UnixMilli()
	hub.recoveryMu.Unlock()
	hub.handleMessageV2(owner, ClientMessage{Type: MsgGetRecoveryRequests, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	expired := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "expired" })
	if expired.ResumeCredential != "" || expired.State != nil {
		t.Fatal("expired request leaked identity")
	}
	if _, _, err := s.ValidateRecovery("p1", code); err != nil {
		t.Fatal("expiry rotated recovery credential")
	}
}

func TestRecoveryRequestsStayPrivateToAuthorizedApprover(t *testing.T) {
	hub := NewHub()
	_, created := createContractRoom(t, hub, 5)
	_ = joinContractPlayer(t, hub, created.RoomID, "p1")
	unrelated := joinContractPlayer(t, hub, created.RoomID, "p2")
	s, _ := hub.registry.Get(created.RoomID)
	code, _ := s.RecoveryCredentialFor("p1")
	device := newFakeConnection()
	hub.handleMessageV2(device, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "edge-private", RecoveryCredential: code})
	_ = waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "pending" })
	hub.handleMessageV2(unrelated.connection, ClientMessage{Type: MsgGetRecoveryRequests, RoomID: created.RoomID, PlayerID: "p2", ResumeCredential: unrelated.credential})
	list := waitForContractMessage(t, unrelated.connection, func(m ServerMessage) bool { return m.Type == ServerMsgRecoveryRequests })
	if len(list.RecoveryRequests) != 0 || list.State != nil || list.ResumeCredential != "" {
		t.Fatal("unrelated player saw a private recovery request")
	}
	invalid := newFakeConnection()
	hub.handleMessageV2(invalid, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "edge-invalid", RecoveryCredential: "invalid"})
	rejected := waitForContractMessage(t, invalid, func(m ServerMessage) bool { return m.Type == ServerMsgError })
	if rejected.State != nil || rejected.ResumeCredential != "" || rejected.RecoveryCredential != "" || len(rejected.RecoveryRequests) > 0 {
		t.Fatal("invalid recovery request leaked private data")
	}
}
