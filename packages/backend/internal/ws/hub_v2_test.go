package ws

import "testing"

func lastServerMessage(t *testing.T, connection *FakeConnection) ServerMessage {
	t.Helper()
	messages := connection.Messages()
	if len(messages) == 0 {
		t.Fatal("expected server message")
	}
	message, ok := messages[len(messages)-1].(ServerMessage)
	if !ok {
		t.Fatalf("unexpected message type %T", messages[len(messages)-1])
	}
	return message
}

func TestProtocolV2CreateResumeAndSequencedCommand(t *testing.T) {
	hub := NewHub()
	creator := NewFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create-1", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5, ScriptID: "trouble_brewing"})
	created := lastServerMessage(t, creator)
	if created.Type != "CREATE_ROOM_RESULT" || created.ResumeCredential == "" || created.RoomRevision != 1 || created.NextClientSequence != 1 {
		t.Fatalf("unexpected create result: %+v", created)
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	command := creator.Messages()[0].(ServerMessage)
	if command.Type != "COMMAND_RESULT" || command.AcceptedSequence != 1 || command.NextClientSequence != 2 || command.RoomRevision != 2 {
		t.Fatalf("unexpected command result: %+v", command)
	}
	if command.IdentityStatus != nil {
		t.Fatalf("member command must not expose room state as identity status: %+v", command.IdentityStatus)
	}

	replacement := NewFakeConnection()
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
	creator := NewFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create-1", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5})
	created := lastServerMessage(t, creator)

	attacker := NewFakeConnection()
	hub.handleMessageV2(attacker, ClientMessage{ProtocolVersion: 2, Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: "wrong"})
	result := lastServerMessage(t, attacker)
	if result.Code != "INVALID_CREDENTIAL" {
		t.Fatalf("unexpected error: %+v", result)
	}
}
