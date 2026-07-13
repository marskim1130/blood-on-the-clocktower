package ws

import (
	"testing"
	"time"
)

func lastServerMessage(t *testing.T, connection *fakeConnection) ServerMessage {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	var messages []any
	for time.Now().Before(deadline) {
		messages = connection.Messages()
		if len(messages) > 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if len(messages) == 0 {
		t.Fatal("expected server message")
	}
	message, ok := messages[len(messages)-1].(ServerMessage)
	if !ok {
		t.Fatalf("unexpected message type %T", messages[len(messages)-1])
	}
	return message
}

func roomStateHasPlayer(state *RoomState, playerID string) bool {
	if state == nil {
		return false
	}
	for _, player := range state.Players {
		if player.ID == playerID {
			return true
		}
	}
	return false
}

func TestProtocolV2CreateResumeAndSequencedCommand(t *testing.T) {
	hub := NewHub()
	creator := newFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create-1", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5, ScriptID: "trouble_brewing"})
	created := lastServerMessage(t, creator)
	if created.Type != "CREATE_ROOM_RESULT" || created.ResumeCredential == "" || created.RoomRevision != 1 || created.NextClientSequence != 1 {
		t.Fatalf("unexpected create result: %+v", created)
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	command := lastServerMessage(t, creator)
	if command.Type != "COMMAND_RESULT" || command.AcceptedSequence != 1 || command.NextClientSequence != 2 || command.RoomRevision != 2 {
		t.Fatalf("unexpected command result: %+v", command)
	}
	if command.IdentityStatus != nil {
		t.Fatalf("member command must not expose room state as identity status: %+v", command.IdentityStatus)
	}

	replacement := newFakeConnection()
	hub.handleMessageV2(replacement, ClientMessage{ProtocolVersion: 2, Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	if !creator.IsClosed() {
		t.Fatal("resume should close the previous connection")
	}
	resumed := lastServerMessage(t, replacement)
	if resumed.Type != "RESUME_ROOM_RESULT" || resumed.NextClientSequence != 2 || resumed.RoomRevision != 2 {
		t.Fatalf("unexpected resume result: %+v", resumed)
	}

	replacement.ClearMessages()
	hub.handleMessageV2(replacement, ClientMessage{ProtocolVersion: 2, Type: MsgGetRoomState, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	state := lastServerMessage(t, replacement)
	if state.Type != "ROOM_STATE" || state.RoomRevision != 2 || state.NextClientSequence != 2 {
		t.Fatalf("unexpected state result: %+v", state)
	}
}

func TestProtocolV2RejectsWrongCredentialWithoutIdentityDisclosure(t *testing.T) {
	hub := NewHub()
	creator := newFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create-1", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5})
	created := lastServerMessage(t, creator)

	attacker := newFakeConnection()
	hub.handleMessageV2(attacker, ClientMessage{ProtocolVersion: 2, Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: "wrong"})
	result := lastServerMessage(t, attacker)
	if result.Code != "INVALID_CREDENTIAL" {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestProtocolV2JoinPublishesCommittedProjection(t *testing.T) {
	hub := NewHub()
	creator := newFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5})
	created := lastServerMessage(t, creator)
	creator.ClearMessages()

	player := newFakeConnection()
	hub.handleMessageV2(player, ClientMessage{ProtocolVersion: 2, Type: MsgJoinRoom, JoinRequestID: "join", RoomID: created.RoomID, PlayerID: "player", PlayerName: "Bob"})
	joined := lastServerMessage(t, player)
	if joined.Type != "JOIN_ROOM_RESULT" || joined.RoomRevision != 2 || joined.ResumeCredential == "" {
		t.Fatalf("unexpected join result: %+v", joined)
	}

	projection := lastServerMessage(t, creator)
	if projection.Type != "ROOM_STATE_CHANGED" || projection.RoomRevision != 2 || projection.State == nil {
		t.Fatalf("unexpected committed projection: %+v", projection)
	}
	if !roomStateHasPlayer(projection.State, "player") {
		t.Fatal("creator projection must contain the committed room member")
	}
}

func TestProtocolV2RejectsConnectionIdentityForgery(t *testing.T) {
	hub := NewHub()
	creator := newFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5})
	created := lastServerMessage(t, creator)

	player := newFakeConnection()
	hub.handleMessageV2(player, ClientMessage{ProtocolVersion: 2, Type: MsgJoinRoom, JoinRequestID: "join", RoomID: created.RoomID, PlayerID: "player", PlayerName: "Bob"})
	_ = lastServerMessage(t, player)
	player.ClearMessages()

	hub.handleMessageV2(player, ClientMessage{ProtocolVersion: 2, Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	result := lastServerMessage(t, player)
	if result.Code != "STALE_CONNECTION" {
		t.Fatalf("forged identity must fail at the active-connection seam: %+v", result)
	}
}

func TestProtocolV2TakeoverInvalidatesPreviousConnection(t *testing.T) {
	hub := NewHub()
	creator := newFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5})
	created := lastServerMessage(t, creator)

	replacement := newFakeConnection()
	hub.handleMessageV2(replacement, ClientMessage{ProtocolVersion: 2, Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	_ = lastServerMessage(t, replacement)
	if !creator.IsClosed() {
		t.Fatal("connection takeover must close the previous active connection")
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	result := lastServerMessage(t, creator)
	if result.Code != "STALE_CONNECTION" {
		t.Fatalf("previous connection must not mutate the room: %+v", result)
	}
}
