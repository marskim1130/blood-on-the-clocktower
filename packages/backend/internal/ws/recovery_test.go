package ws

import "testing"

func TestRecoveryWaitsForApprovalBeforeReplacingTheOriginalDevice(t *testing.T) {
	hub := NewHub()
	owner, created := createContractRoom(t, hub, 5)
	player := joinContractPlayer(t, hub, created.RoomID, "p1")
	s, _ := hub.registry.Get(created.RoomID)
	code, _ := s.RecoveryCredentialFor("p1")
	newDevice := newFakeConnection()
	hub.handleMessageV2(newDevice, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "recover-one", RecoveryCredential: code})
	waiting := waitForContractMessage(t, newDevice, func(m ServerMessage) bool { return m.Type == ServerMsgRecoveryStatus })
	if waiting.RecoveryStatus != "pending" || waiting.State != nil || waiting.ResumeCredential != "" {
		t.Fatalf("approval was bypassed: %+v", waiting)
	}
	current, _, _ := hub.active.Get(created.RoomID, "p1")
	if current != player.connection {
		t.Fatal("pending request replaced original device")
	}
	yes := true
	hub.handleMessageV2(owner, ClientMessage{Type: MsgReviewRecovery, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: created.NextClientSequence, RecoveryRequestID: "recover-one", Decision: &yes})
	approved := waitForContractMessage(t, newDevice, func(m ServerMessage) bool { return m.Type == ServerMsgRecoveryStatus && m.RecoveryStatus == "approved" })
	if approved.ResumeCredential == "" || approved.RecoveryCredential == "" || approved.State != nil {
		t.Fatal("approval response is missing credentials or leaking state")
	}
	hub.handleMessageV2(newDevice, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "p1", ResumeCredential: approved.ResumeCredential})
	resumed := waitForContractMessage(t, newDevice, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if resumed.State == nil {
		t.Fatal("approved device did not resume")
	}
	oldCodeAttempt := newFakeConnection()
	hub.handleMessageV2(oldCodeAttempt, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "p1", ResumeCredential: player.credential})
	if message := lastServerMessage(t, oldCodeAttempt); message.Type != ServerMsgError {
		t.Fatal("old resume credential survived takeover")
	}
}
