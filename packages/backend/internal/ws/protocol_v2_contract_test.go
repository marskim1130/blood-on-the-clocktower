package ws

import (
	"strconv"
	"testing"
	"time"
)

type contractClient struct {
	connection *fakeConnection
	credential string
}

func waitForContractMessage(t *testing.T, connection *fakeConnection, matches func(ServerMessage) bool) ServerMessage {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		for _, raw := range connection.Messages() {
			message, ok := raw.(ServerMessage)
			if ok && matches(message) {
				return message
			}
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("expected matching server message")
	return ServerMessage{}
}

func createContractRoom(t *testing.T, hub *Hub, maxPlayers int) (*fakeConnection, ServerMessage) {
	t.Helper()
	creator := newFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{
		ProtocolVersion: 2,
		Type:            MsgCreateRoom,
		RequestID:       "contract-create",
		PlayerID:        "creator",
		PlayerName:      "Alice",
		MaxPlayers:      maxPlayers,
		ScriptID:        "trouble_brewing",
	})
	return creator, lastServerMessage(t, creator)
}

func joinContractPlayer(t *testing.T, hub *Hub, roomID, playerID string) contractClient {
	t.Helper()
	connection := newFakeConnection()
	hub.handleMessageV2(connection, ClientMessage{
		ProtocolVersion: 2,
		Type:            MsgJoinRoom,
		JoinRequestID:   "join-" + playerID,
		RoomID:          roomID,
		PlayerID:        playerID,
		PlayerName:      playerID,
	})
	message := lastServerMessage(t, connection)
	if message.Type != ServerMsgJoinRoomResult {
		t.Fatalf("join %s failed: %+v", playerID, message)
	}
	return contractClient{connection: connection, credential: message.ResumeCredential}
}

func TestProtocolV2CreateValidatesCapacityAndScript(t *testing.T) {
	tests := []struct {
		name       string
		maxPlayers int
		scriptID   string
	}{
		{name: "too few players", maxPlayers: 4, scriptID: "trouble_brewing"},
		{name: "too many players", maxPlayers: 16, scriptID: "trouble_brewing"},
		{name: "unsupported script", maxPlayers: 5, scriptID: "not-a-script"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hub := NewHub()
			connection := newFakeConnection()
			hub.handleMessageV2(connection, ClientMessage{
				ProtocolVersion: 2,
				Type:            MsgCreateRoom,
				RequestID:       "invalid-create",
				PlayerID:        "creator",
				PlayerName:      "Alice",
				MaxPlayers:      test.maxPlayers,
				ScriptID:        test.scriptID,
			})
			message := lastServerMessage(t, connection)
			if message.Code != ProtocolErrorInvalidCommand {
				t.Fatalf("expected INVALID_COMMAND, got %+v", message)
			}
		})
	}
}

func TestProtocolV2RoomFullUsesDedicatedError(t *testing.T) {
	hub := NewHub()
	_, created := createContractRoom(t, hub, 5)
	for i := 1; i <= 5; i++ {
		joinContractPlayer(t, hub, created.RoomID, "p"+strconv.Itoa(i))
	}

	overflow := newFakeConnection()
	hub.handleMessageV2(overflow, ClientMessage{
		ProtocolVersion: 2,
		Type:            MsgJoinRoom,
		JoinRequestID:   "join-overflow",
		RoomID:          created.RoomID,
		PlayerID:        "overflow",
		PlayerName:      "Overflow",
	})
	message := lastServerMessage(t, overflow)
	if message.Code != ProtocolErrorRoomFull {
		t.Fatalf("expected ROOM_FULL, got %+v", message)
	}
}

func TestProtocolV2CreatorCanSetSelfAsStoryteller(t *testing.T) {
	hub := NewHub()
	creator, created := createContractRoom(t, hub, 5)
	if !roomStateHasPlayer(created.State, "creator") {
		t.Fatal("creator must be present in the game engine before storyteller selection")
	}
	member := joinContractPlayer(t, hub, created.RoomID, "p1")

	member.connection.ClearMessages()
	hub.handleMessageV2(member.connection, ClientMessage{
		ProtocolVersion:  2,
		Type:             MsgSetStoryteller,
		RoomID:           created.RoomID,
		PlayerID:         "p1",
		TargetPlayerID:   "p1",
		ResumeCredential: member.credential,
		ClientSequence:   1,
	})
	if message := lastServerMessage(t, member.connection); message.Code != ProtocolErrorForbidden {
		t.Fatalf("only creator may select storyteller, got %+v", message)
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{
		ProtocolVersion:  2,
		Type:             MsgSetStoryteller,
		RoomID:           created.RoomID,
		PlayerID:         "creator",
		TargetPlayerID:   "creator",
		ResumeCredential: created.ResumeCredential,
		ClientSequence:   1,
	})
	message := waitForContractMessage(t, creator, func(message ServerMessage) bool {
		return message.Type == ServerMsgCommandResult && message.AcceptedSequence == 1
	})
	if message.Type != ServerMsgCommandResult || message.State == nil {
		t.Fatalf("unexpected storyteller command result: %+v", message)
	}
	if message.State.StorytellerID != "creator" || message.State.StorytellerName != "Alice" {
		t.Fatalf("unexpected storyteller projection: %+v", message.State)
	}
	if roomStateHasPlayer(message.State, "creator") {
		t.Fatal("storyteller must not count as an actual player")
	}
}

func TestProtocolV2AssignmentFreezesOnlyAfterSuccessAndKeepsRedHerringPrivate(t *testing.T) {
	hub := NewHub()
	creator, created := createContractRoom(t, hub, 5)
	clients := map[string]contractClient{}
	for _, playerID := range []string{"p1", "p2", "p3", "p4"} {
		clients[playerID] = joinContractPlayer(t, hub, created.RoomID, playerID)
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgSetStoryteller, RoomID: created.RoomID, PlayerID: "creator", TargetPlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1})
	if message := waitForContractMessage(t, creator, func(message ServerMessage) bool {
		return message.Type == ServerMsgCommandResult && message.AcceptedSequence == 1
	}); message.Type != ServerMsgCommandResult {
		t.Fatalf("set storyteller failed: %+v", message)
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgAssignCharacters, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 2, Assignments: map[string]string{}})
	if message := waitForContractMessage(t, creator, func(message ServerMessage) bool {
		return message.Code == ProtocolErrorInvalidCommand && message.NextClientSequence == 2
	}); message.Code != ProtocolErrorInvalidCommand || message.NextClientSequence != 2 {
		t.Fatalf("invalid assignment must not commit or consume sequence: %+v", message)
	}

	clients["p5"] = joinContractPlayer(t, hub, created.RoomID, "p5")
	ready := true
	for playerID, client := range clients {
		hub.handleMessageV2(client.connection, ClientMessage{
			ProtocolVersion:  2,
			Type:             MsgSetReady,
			RoomID:           created.RoomID,
			PlayerID:         playerID,
			ResumeCredential: client.credential,
			ClientSequence:   1,
			Ready:            &ready,
		})
		if message := waitForContractMessage(t, client.connection, func(message ServerMessage) bool {
			return message.Type == ServerMsgCommandResult && message.AcceptedSequence == 1
		}); message.Type != ServerMsgCommandResult {
			t.Fatalf("set ready for %s failed: %+v", playerID, message)
		}
	}
	for _, client := range clients {
		client.connection.ClearMessages()
	}
	creator.ClearMessages()
	assignments := map[string]string{
		"p1": "fortuneteller",
		"p2": "slayer",
		"p3": "soldier",
		"p4": "poisoner",
		"p5": "imp",
	}
	hub.handleMessageV2(creator, ClientMessage{
		ProtocolVersion:           2,
		Type:                      MsgAssignCharacters,
		RoomID:                    created.RoomID,
		PlayerID:                  "creator",
		ResumeCredential:          created.ResumeCredential,
		ClientSequence:            2,
		Assignments:               assignments,
		FortuneTellerRedHerringID: "p2",
	})
	assigned := waitForContractMessage(t, creator, func(message ServerMessage) bool {
		return message.Type == ServerMsgCommandResult && message.AcceptedSequence == 2
	})
	if assigned.Type != ServerMsgCommandResult || assigned.State == nil || assigned.State.FortuneTellerRedHerringID != "p2" {
		t.Fatalf("unexpected assignment result: %+v", assigned)
	}
	if assigned.IdentityStatus == nil || assigned.IdentityStatus.Status != "member" || !assigned.IdentityStatus.ParticipantSetFrozen || assigned.IdentityStatus.NextClientSequence != 3 {
		t.Fatalf("storyteller did not receive authoritative frozen identity status: %+v", assigned.IdentityStatus)
	}
	playerProjection := waitForContractMessage(t, clients["p1"].connection, func(message ServerMessage) bool {
		return message.Type == ServerMsgRoomStateChanged && message.RoomRevision == assigned.RoomRevision
	})
	if playerProjection.State == nil || playerProjection.State.FortuneTellerRedHerringID != "" {
		t.Fatalf("red herring leaked to player: %+v", playerProjection)
	}
	if playerProjection.IdentityStatus == nil || playerProjection.IdentityStatus.Status != "member" || !playerProjection.IdentityStatus.ParticipantSetFrozen || playerProjection.IdentityStatus.NextClientSequence != 2 {
		t.Fatalf("player did not receive authoritative frozen identity status: %+v", playerProjection.IdentityStatus)
	}
	for playerID, client := range clients {
		hub.handleMessageV2(client.connection, ClientMessage{
			ProtocolVersion:  2,
			Type:             MsgConfirmCharacter,
			RoomID:           created.RoomID,
			PlayerID:         playerID,
			ResumeCredential: client.credential,
			ClientSequence:   2,
		})
		if message := waitForContractMessage(t, client.connection, func(message ServerMessage) bool {
			return message.Type == ServerMsgCommandResult && message.AcceptedSequence == 2
		}); message.Type != ServerMsgCommandResult {
			t.Fatalf("confirm character for %s failed: %+v", playerID, message)
		}
	}

	resumedPlayer := newFakeConnection()
	hub.handleMessageV2(resumedPlayer, ClientMessage{ProtocolVersion: 2, Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "p1", ResumeCredential: clients["p1"].credential})
	resumed := lastServerMessage(t, resumedPlayer)
	if resumed.Type != ServerMsgResumeRoomResult || resumed.State == nil || resumed.IdentityStatus == nil || !resumed.IdentityStatus.ParticipantSetFrozen {
		t.Fatalf("resume did not restore frozen member status and projection: %+v", resumed)
	}

	overflow := newFakeConnection()
	hub.handleMessageV2(overflow, ClientMessage{ProtocolVersion: 2, Type: MsgJoinRoom, JoinRequestID: "join-after-freeze", RoomID: created.RoomID, PlayerID: "late", PlayerName: "Late"})
	if message := lastServerMessage(t, overflow); message.Code != ProtocolErrorParticipantSetFrozen {
		t.Fatalf("successful assignment must freeze participants: %+v", message)
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgAssignCharacters, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 3, Assignments: assignments, FortuneTellerRedHerringID: "p2"})
	if message := waitForContractMessage(t, creator, func(message ServerMessage) bool {
		return message.Code == ProtocolErrorInvalidCommand && message.NextClientSequence == 3
	}); message.Code != ProtocolErrorInvalidCommand {
		t.Fatalf("second assignment must be rejected: %+v", message)
	}

	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgStartGame, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 3})
	started := waitForContractMessage(t, creator, func(message ServerMessage) bool {
		return message.Type == ServerMsgCommandResult && message.AcceptedSequence == 3
	})
	if started.Type != ServerMsgCommandResult || started.State == nil || started.State.DayNumber != 0 || started.State.NightNumber != 1 {
		t.Fatalf("first night must be day=0/night=1: %+v", started)
	}
}
