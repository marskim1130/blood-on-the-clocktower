package ws

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/marskim1130/blood-on-the-clocktower/internal/sessionstore"
)

// A real in-memory store with one controllable write, so the tests exercise the
// complete Hub commit/delivery path without replacing the session or gameplay.
type pausedRoomStore struct {
	sessionstore.RoomRecordStore
	mu         sync.Mutex
	roomID     string
	next       *pausedRoomWrite
	nextDelete *pausedRoomWrite
	nextCreate *pausedRoomWrite
}

type pausedRoomWrite struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	err     error
}

func (s *pausedRoomStore) pause(roomID string, err error) *pausedRoomWrite {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := &pausedRoomWrite{started: make(chan struct{}), release: make(chan struct{}), err: err}
	s.roomID, s.next = roomID, p
	return p
}

func (p *pausedRoomWrite) unblock() { p.once.Do(func() { close(p.release) }) }

func (s *pausedRoomStore) Replace(ctx context.Context, revision uint64, record sessionstore.Record) error {
	s.mu.Lock()
	var paused *pausedRoomWrite
	if record.RoomID == s.roomID {
		paused, s.next = s.next, nil
	}
	s.mu.Unlock()
	if paused != nil {
		close(paused.started)
		<-paused.release
		if paused.err != nil {
			return paused.err
		}
	}
	return s.RoomRecordStore.Replace(ctx, revision, record)
}

func (s *pausedRoomStore) pauseDelete(roomID string) *pausedRoomWrite {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := &pausedRoomWrite{started: make(chan struct{}), release: make(chan struct{})}
	s.roomID, s.nextDelete = roomID, p
	return p
}

func (s *pausedRoomStore) Delete(ctx context.Context, roomID string, revision uint64) error {
	s.mu.Lock()
	var paused *pausedRoomWrite
	if roomID == s.roomID {
		paused, s.nextDelete = s.nextDelete, nil
	}
	s.mu.Unlock()
	if paused != nil {
		close(paused.started)
		<-paused.release
	}
	return s.RoomRecordStore.Delete(ctx, roomID, revision)
}

func (s *pausedRoomStore) pauseCreate() *pausedRoomWrite {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := &pausedRoomWrite{started: make(chan struct{}), release: make(chan struct{})}
	s.nextCreate = p
	return p
}

func (s *pausedRoomStore) Create(ctx context.Context, record sessionstore.Record) error {
	s.mu.Lock()
	paused := s.nextCreate
	s.nextCreate = nil
	s.mu.Unlock()
	if paused != nil {
		close(paused.started)
		<-paused.release
	}
	return s.RoomRecordStore.Create(ctx, record)
}

func awaitRoomFlow(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("room operation did not finish")
	}
}

func TestTakeoverReturnsStateCommittedWhileItWaited(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	commandDone := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	awaitRoomFlow(t, paused.started)
	device := newFakeConnection()
	resumeDone := hub.submitMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	select {
	case <-resumeDone:
		t.Fatal("takeover completed before the preceding write")
	default:
	}
	paused.unblock()
	awaitRoomFlow(t, commandDone)
	awaitRoomFlow(t, resumeDone)
	result := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if result.RoomRevision != 2 || result.NextClientSequence != 2 || result.State == nil || result.State.MaxPlayers != 7 {
		t.Fatalf("takeover used a pre-commit view: revision=%d sequence=%d", result.RoomRevision, result.NextClientSequence)
	}
}

func TestAdmittedCommandsSurviveDisconnectInReceiveOrder(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	first := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	awaitRoomFlow(t, paused.started)
	second := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 2, MaxPlayers: 9})
	disconnected := make(chan struct{})
	go func() { hub.handleDisconnect(owner); close(disconnected) }()
	awaitRoomFlow(t, disconnected)
	device := newFakeConnection()
	resumed := hub.submitMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	paused.unblock()
	awaitRoomFlow(t, first)
	awaitRoomFlow(t, second)
	awaitRoomFlow(t, resumed)
	result := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if result.RoomRevision != 3 || result.NextClientSequence != 3 || result.State == nil || result.State.MaxPlayers != 9 {
		t.Fatalf("disconnect lost admitted commands: revision=%d sequence=%d", result.RoomRevision, result.NextClientSequence)
	}
}

func TestRecoveryWriteDoesNotBlockAnotherRoomsTakeover(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	player := joinContractPlayer(t, hub, created.RoomID, "p1")
	joined := waitForContractMessage(t, player.connection, func(m ServerMessage) bool { return m.Type == ServerMsgJoinRoomResult })
	device := newFakeConnection()
	hub.handleMessageV2(device, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "flow-recovery", RecoveryCredential: joined.RecoveryCredential})
	other := newFakeConnection()
	hub.handleMessageV2(other, ClientMessage{Type: MsgCreateRoom, RequestID: "other-room", PlayerID: "other", PlayerName: "Other", MaxPlayers: 5})
	otherCreated := waitForContractMessage(t, other, func(m ServerMessage) bool { return m.Type == ServerMsgCreateRoomResult })
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	yes := true
	reviewed := hub.submitMessageV2(owner, ClientMessage{Type: MsgReviewRecovery, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, RecoveryRequestID: "flow-recovery", Decision: &yes})
	awaitRoomFlow(t, paused.started)
	otherDevice := newFakeConnection()
	resumed := hub.submitMessageV2(otherDevice, ClientMessage{Type: MsgResumeRoom, RoomID: otherCreated.RoomID, PlayerID: "other", ResumeCredential: otherCreated.ResumeCredential})
	awaitRoomFlow(t, resumed)
	result := waitForContractMessage(t, otherDevice, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if result.RoomRevision != 1 {
		t.Fatalf("unrelated room revision=%d", result.RoomRevision)
	}
	paused.unblock()
	awaitRoomFlow(t, reviewed)
}

func TestCreateRetryWaitsForPriorCommandsBeforeAttaching(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	command := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	awaitRoomFlow(t, paused.started)
	device := newFakeConnection()
	retry := hub.submitMessageV2(device, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "contract-create", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5, ScriptID: "trouble_brewing"})
	select {
	case <-retry:
		t.Fatal("create retry bypassed the preceding room command")
	default:
	}
	paused.unblock()
	awaitRoomFlow(t, command)
	awaitRoomFlow(t, retry)
	result := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgCreateRoomResult })
	if result.RoomID != created.RoomID || result.RoomRevision != 2 || result.NextClientSequence != 2 || result.State == nil || result.State.MaxPlayers != 7 {
		t.Fatalf("create retry attached with stale state: revision=%d sequence=%d", result.RoomRevision, result.NextClientSequence)
	}
}

func TestExpiryWriteDoesNotBlockCreatingAnotherRoom(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	_, created := createContractRoom(t, hub, 5)
	paused := store.pauseDelete(created.RoomID)
	t.Cleanup(paused.unblock)
	cleaned := make(chan struct{})
	go func() { hub.CleanupInactiveRooms(context.Background(), time.Now().Add(8*24*time.Hour)); close(cleaned) }()
	awaitRoomFlow(t, paused.started)
	other := newFakeConnection()
	newRoom := hub.submitMessageV2(other, ClientMessage{Type: MsgCreateRoom, RequestID: "during-expiry", PlayerID: "other", PlayerName: "Other", MaxPlayers: 5})
	awaitRoomFlow(t, newRoom)
	result := waitForContractMessage(t, other, func(m ServerMessage) bool { return m.Type == ServerMsgCreateRoomResult })
	if result.RoomID == created.RoomID || result.RoomRevision != 1 {
		t.Fatal("independent create did not succeed")
	}
	paused.unblock()
	awaitRoomFlow(t, cleaned)
}

func TestResumeOnCurrentConnectionKeepsItOpen(t *testing.T) {
	hub := NewHub()
	owner, created := createContractRoom(t, hub, 5)
	hub.handleMessageV2(owner, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	if owner.IsClosed() {
		t.Fatal("resuming the current connection closed that connection")
	}
	hub.handleMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	result := waitForContractMessage(t, owner, func(m ServerMessage) bool { return m.Type == ServerMsgCommandResult })
	if result.RoomRevision != 2 {
		t.Fatal("current connection lost command eligibility")
	}
}

func TestCreateRetryDoesNotImportAnotherRoomsSlowCreate(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	_, created := createContractRoom(t, hub, 5)
	paused := store.pauseCreate()
	t.Cleanup(paused.unblock)
	creating := hub.submitMessageV2(newFakeConnection(), ClientMessage{Type: MsgCreateRoom, RequestID: "slow-new-room", PlayerID: "other", PlayerName: "Other", MaxPlayers: 5})
	awaitRoomFlow(t, paused.started)
	retrying := hub.submitMessageV2(newFakeConnection(), ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "contract-create", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5, ScriptID: "trouble_brewing"})
	device := newFakeConnection()
	resuming := hub.submitMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	awaitRoomFlow(t, resuming)
	result := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if result.RoomRevision != 1 {
		t.Fatal("existing room could not resume during unrelated create")
	}
	paused.unblock()
	awaitRoomFlow(t, creating)
	awaitRoomFlow(t, retrying)
}

func TestRoomFlowOrdersJoinCommandTakeoverAndRejectsSupersededCommands(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	joining := newFakeConnection()
	join := hub.submitMessageV2(joining, ClientMessage{Type: MsgJoinRoom, RoomID: created.RoomID, PlayerID: "p1", PlayerName: "Player", JoinRequestID: "flow-join"})
	awaitRoomFlow(t, paused.started)
	command := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	device := newFakeConnection()
	resume := hub.submitMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	obsolete := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 2, MaxPlayers: 10})
	paused.unblock()
	for _, done := range []<-chan struct{}{join, command, resume, obsolete} {
		awaitRoomFlow(t, done)
	}
	joined := waitForContractMessage(t, joining, func(m ServerMessage) bool { return m.Type == ServerMsgJoinRoomResult })
	resumed := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if joined.RoomRevision != 2 || resumed.RoomRevision != 3 || resumed.NextClientSequence != 2 || resumed.State == nil || resumed.State.MaxPlayers != 7 {
		t.Fatal("join, command and takeover were not applied in admission order")
	}
	hub.handleMessageV2(device, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 2, MaxPlayers: 9})
	result := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgCommandResult })
	if result.RoomRevision != 4 || result.NextClientSequence != 3 || result.State.MaxPlayers != 9 {
		t.Fatal("superseded command consumed a sequence or changed state")
	}
}

func TestSameConnectionResumePreservesFollowingAdmittedCommands(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	first := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	awaitRoomFlow(t, paused.started)
	resume := hub.submitMessageV2(owner, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	second := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 2, MaxPlayers: 9})
	paused.unblock()
	for _, done := range []<-chan struct{}{first, resume, second} {
		awaitRoomFlow(t, done)
	}
	result := waitForContractMessage(t, owner, func(m ServerMessage) bool { return m.Type == ServerMsgCommandResult && m.AcceptedSequence == 2 })
	if owner.IsClosed() || result.RoomRevision != 3 || result.State.MaxPlayers != 9 {
		t.Fatal("same-connection resume invalidated admitted work")
	}
}

func TestFailedCommitLeavesTakeoverStateAndRetryDeduplicationIntact(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	witness := joinContractPlayer(t, hub, created.RoomID, "witness")
	paused := store.pause(created.RoomID, errors.New("controlled persistence failure"))
	t.Cleanup(paused.unblock)
	message := ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7}
	failed := hub.submitMessageV2(owner, message)
	awaitRoomFlow(t, paused.started)
	device := newFakeConnection()
	resume := hub.submitMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	paused.unblock()
	awaitRoomFlow(t, failed)
	awaitRoomFlow(t, resume)
	result := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if result.RoomRevision != 2 || result.NextClientSequence != 1 || result.State == nil || result.State.MaxPlayers != 5 {
		t.Fatal("failed write was exposed by takeover")
	}
	hub.handleMessageV2(device, message)
	committed := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgCommandResult })
	if committed.RoomRevision != 3 || committed.NextClientSequence != 2 {
		t.Fatal("retry after failed write did not commit once")
	}
	hub.handleMessageV2(device, message)
	// A query on the witness connection is an outbound ordering barrier.
	hub.handleMessageV2(witness.connection, ClientMessage{Type: MsgGetRoomState, RoomID: created.RoomID, PlayerID: "witness", ResumeCredential: witness.credential})
	state := waitForContractMessage(t, witness.connection, func(m ServerMessage) bool { return m.Type == ServerMsgRoomState })
	if state.RoomRevision != 3 {
		t.Fatal("duplicate command advanced the revision")
	}
	updates := 0
	for _, raw := range witness.connection.Messages() {
		if m, ok := raw.(ServerMessage); ok && m.Type == ServerMsgRoomStateChanged && m.RoomRevision >= 3 {
			updates++
		}
	}
	if updates != 1 {
		t.Fatalf("retry produced %d committed broadcasts, want 1", updates)
	}
}

func TestRecoveryAndFollowingCommandDeliverInCommitOrder(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	player := joinContractPlayer(t, hub, created.RoomID, "p1")
	witness := joinContractPlayer(t, hub, created.RoomID, "witness")
	joined := waitForContractMessage(t, player.connection, func(m ServerMessage) bool { return m.Type == ServerMsgJoinRoomResult })
	device := newFakeConnection()
	hub.handleMessageV2(device, ClientMessage{Type: MsgRequestRecovery, RoomID: created.RoomID, PlayerID: "p1", RequestID: "ordered-recovery", RecoveryCredential: joined.RecoveryCredential})
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	yes := true
	review := hub.submitMessageV2(owner, ClientMessage{Type: MsgReviewRecovery, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, RecoveryRequestID: "ordered-recovery", Decision: &yes})
	awaitRoomFlow(t, paused.started)
	command := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 2, MaxPlayers: 7})
	oldCommand := hub.submitMessageV2(player.connection, ClientMessage{Type: MsgSetReady, RoomID: created.RoomID, PlayerID: "p1", ResumeCredential: player.credential, ClientSequence: 1, Ready: &yes})
	paused.unblock()
	for _, done := range []<-chan struct{}{review, command, oldCommand} {
		awaitRoomFlow(t, done)
	}
	approved := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.RecoveryStatus == "approved" })
	if approved.State != nil || approved.ResumeCredential == "" {
		t.Fatal("approval leaked a projection or failed to grant identity")
	}
	hub.handleMessageV2(witness.connection, ClientMessage{Type: MsgGetRoomState, RoomID: created.RoomID, PlayerID: "witness", ResumeCredential: witness.credential})
	state := waitForContractMessage(t, witness.connection, func(m ServerMessage) bool { return m.Type == ServerMsgRoomState })
	var revisions []uint64
	for _, raw := range witness.connection.Messages() {
		if m, ok := raw.(ServerMessage); ok && m.Type == ServerMsgRoomStateChanged && m.RoomRevision >= 4 {
			revisions = append(revisions, m.RoomRevision)
		}
	}
	if state.RoomRevision != 5 || len(revisions) != 2 || revisions[0] != 4 || revisions[1] != 5 {
		t.Fatalf("recovery/command delivery order=%v", revisions)
	}
	hub.handleMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "p1", ResumeCredential: approved.ResumeCredential})
	resumed := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgResumeRoomResult })
	if resumed.RoomRevision != 5 || resumed.NextClientSequence != 1 {
		t.Fatal("revoked device mutated the recovered identity")
	}
}

func TestCloseFollowsAdmittedCommandAndPrecedesTakeover(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	owner, created := createContractRoom(t, hub, 5)
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	command := hub.submitMessageV2(owner, ClientMessage{Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	awaitRoomFlow(t, paused.started)
	closed := hub.submitMessageV2(owner, ClientMessage{Type: MsgCloseRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 2})
	device := newFakeConnection()
	resume := hub.submitMessageV2(device, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	paused.unblock()
	for _, done := range []<-chan struct{}{command, closed, resume} {
		awaitRoomFlow(t, done)
	}
	result := waitForContractMessage(t, device, func(m ServerMessage) bool { return m.Type == ServerMsgError })
	if result.Code != ProtocolErrorRoomNotFound {
		t.Fatalf("takeover after close returned %s", result.Code)
	}
	// Restore through the actual persistence seam to prove the close was durable.
	restored, failures, err := NewHubWithRoomRecordStore(store, developmentCredentialKey())
	if err != nil || len(failures) != 0 {
		t.Fatal("restore failed")
	}
	restoredDevice := newFakeConnection()
	restored.handleMessageV2(restoredDevice, ClientMessage{Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	if lastServerMessage(t, restoredDevice).Code != ProtocolErrorRoomNotFound {
		t.Fatal("closed room reappeared after restore")
	}
}

func TestWireDisconnectDoesNotCancelReceivedCommand(t *testing.T) {
	store := &pausedRoomStore{RoomRecordStore: sessionstore.NewMemoryStore()}
	hub := newHubWithSessionStore(store, developmentCredentialKey())
	server := httptest.NewServer(handlerForHub(hub))
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	owner, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if err := owner.WriteJSON(ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "wire-flow", PlayerID: "creator", PlayerName: "Owner", MaxPlayers: 5}); err != nil {
		t.Fatal(err)
	}
	created := readWireMessage(t, owner)
	paused := store.pause(created.RoomID, nil)
	t.Cleanup(paused.unblock)
	if err := owner.WriteJSON(ClientMessage{ProtocolVersion: 2, Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7}); err != nil {
		t.Fatal(err)
	}
	awaitRoomFlow(t, paused.started)
	_ = owner.Close()
	device, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer device.Close()
	if err := device.WriteJSON(ClientMessage{ProtocolVersion: 2, Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential}); err != nil {
		t.Fatal(err)
	}
	paused.unblock()
	resumed := readWireMessage(t, device)
	if resumed.Type != ServerMsgResumeRoomResult || resumed.RoomRevision != 2 || resumed.NextClientSequence != 2 || resumed.State == nil || resumed.State.MaxPlayers != 7 {
		t.Fatal("wire disconnect cancelled an admitted command or resume used stale state")
	}
}

func TestRoomAdmissionPreservesProtocolErrorKinds(t *testing.T) {
	hub := NewHub()
	_, created := createContractRoom(t, hub, 5)
	for _, test := range []struct{ messageType, code string }{
		{MsgGetRoomState, ProtocolErrorInvalidCredential},
		{"UNKNOWN_MESSAGE", ProtocolErrorInternal},
	} {
		t.Run(test.messageType, func(t *testing.T) {
			unattached := newFakeConnection()
			hub.handleMessageV2(unattached, ClientMessage{Type: test.messageType, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
			result := waitForContractMessage(t, unattached, func(m ServerMessage) bool { return m.Type == ServerMsgError })
			if result.Code != test.code {
				t.Fatalf("error code=%s, want %s", result.Code, test.code)
			}
		})
	}
}
