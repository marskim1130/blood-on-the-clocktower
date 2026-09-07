package ws

import "testing"

func TestProtocolV2SeatedPlayerPublishesOwnReadyState(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")

	harness.command(storyteller, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: storyteller.id,
	})
	ready := true
	commit := harness.command(p1, ClientMessage{
		Type:  MsgSetReady,
		Ready: &ready,
	})

	projections := []*RoomState{
		commit.direct.State,
		commit.broadcasts[storyteller.id].State,
		commit.broadcasts[p2.id].State,
	}
	for _, state := range projections {
		if !playerByID(t, state, p1.id).IsReady || playerByID(t, state, p2.id).IsReady {
			t.Fatalf("ready state must identify only p1 as ready: %+v", state.Players)
		}
	}
}

func TestProtocolV2StorytellerCannotAssignCharactersUntilEveryPlayerIsReady(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")
	p3 := harness.join("p3", "P3")
	p4 := harness.join("p4", "P4")
	p5 := harness.join("p5", "P5")
	harness.command(storyteller, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: storyteller.id,
	})
	ready := true
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4} {
		harness.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}

	harness.write(storyteller, ClientMessage{
		ProtocolVersion:  2,
		Type:             MsgAssignCharacters,
		RoomID:           harness.roomID,
		PlayerID:         storyteller.id,
		ResumeCredential: storyteller.resumeCredential,
		ClientSequence:   storyteller.nextSequence,
		Assignments: map[string]string{
			p1.id: "slayer",
			p2.id: "soldier",
			p3.id: "mayor",
			p4.id: "poisoner",
			p5.id: "imp",
		},
	})
	result := readWireMessage(t, storyteller.connection)

	if result.Type != ServerMsgError || result.Code != ProtocolErrorInvalidCommand {
		t.Fatalf("assignment with an unready player must be rejected: %+v", result)
	}
	if result.NextClientSequence != storyteller.nextSequence {
		t.Fatalf("rejected assignment next sequence = %d, want %d", result.NextClientSequence, storyteller.nextSequence)
	}
}

func TestProtocolV2ChangingSeatOrderClearsEveryPlayersReadyState(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")
	harness.command(storyteller, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: storyteller.id,
	})
	ready := true
	harness.command(p1, ClientMessage{Type: MsgSetReady, Ready: &ready})
	harness.command(p2, ClientMessage{Type: MsgSetReady, Ready: &ready})

	reordered := harness.command(storyteller, ClientMessage{
		Type:      MsgSetSeatOrder,
		SeatOrder: []string{p2.id, p1.id},
	})

	for _, player := range reordered.direct.State.Players {
		if player.IsReady {
			t.Fatalf("seat change must clear %s ready state: %+v", player.ID, reordered.direct.State.Players)
		}
	}
}

func TestProtocolV2JoiningPlayerClearsExistingReadyState(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	harness.command(storyteller, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: storyteller.id,
	})
	ready := true
	harness.command(p1, ClientMessage{Type: MsgSetReady, Ready: &ready})
	p2 := harness.join("p2", "P2")

	restored := harness.resume(p1)
	if playerByID(t, restored.State, p1.id).IsReady || playerByID(t, restored.State, p2.id).IsReady {
		t.Fatalf("membership change must clear every ready state: %+v", restored.State.Players)
	}
}

func TestProtocolV2LeavingPlayerClearsRemainingReadyState(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")
	harness.command(storyteller, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: storyteller.id,
	})
	ready := true
	harness.command(p1, ClientMessage{Type: MsgSetReady, Ready: &ready})

	harness.write(p2, ClientMessage{
		ProtocolVersion:  2,
		Type:             MsgLeaveRoom,
		RoomID:           harness.roomID,
		PlayerID:         p2.id,
		ResumeCredential: p2.resumeCredential,
		ClientSequence:   p2.nextSequence,
	})
	direct := readWireMessage(t, p2.connection)
	if direct.Type != ServerMsgCommandResult || direct.State != nil {
		t.Fatalf("unexpected leave result: %+v", direct)
	}
	readWireMessage(t, storyteller.connection)
	remaining := readWireMessage(t, p1.connection)
	if remaining.State == nil || playerByID(t, remaining.State, p1.id).IsReady {
		t.Fatalf("membership change must clear remaining ready state: %+v", remaining.State)
	}
}

func TestProtocolV2SelectingStorytellerClearsRemainingReadyState(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	creator := harness.create("creator", "Creator")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")
	ready := true
	harness.command(p1, ClientMessage{Type: MsgSetReady, Ready: &ready})

	selected := harness.command(creator, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: p2.id,
	})

	if playerByID(t, selected.direct.State, p1.id).IsReady {
		t.Fatalf("storyteller selection must clear remaining ready state: %+v", selected.direct.State.Players)
	}
}
