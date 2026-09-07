package ws

import "testing"

func TestProtocolV2PlayerConfirmsOwnAssignedCharacter(t *testing.T) {
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
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4, p5} {
		harness.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}
	harness.command(storyteller, ClientMessage{
		Type: MsgAssignCharacters,
		Assignments: map[string]string{
			p1.id: "slayer",
			p2.id: "soldier",
			p3.id: "mayor",
			p4.id: "poisoner",
			p5.id: "imp",
		},
	})

	confirmed := harness.command(p1, ClientMessage{Type: MsgConfirmCharacter})

	if !playerByID(t, confirmed.direct.State, p1.id).HasConfirmedCharacter {
		t.Fatalf("p1 must see their character confirmation: %+v", confirmed.direct.State.Players)
	}
	storytellerView := confirmed.broadcasts[storyteller.id].State
	if !playerByID(t, storytellerView, p1.id).HasConfirmedCharacter {
		t.Fatalf("storyteller must see p1 confirmation: %+v", storytellerView.Players)
	}
	p2View := confirmed.broadcasts[p2.id].State
	if !playerByID(t, p2View, p1.id).HasConfirmedCharacter {
		t.Fatalf("other players must see p1 confirmation: %+v", p2View.Players)
	}
	if playerByID(t, p2View, p1.id).Character != nil {
		t.Fatal("confirmation must not disclose p1 character to p2")
	}
	if playerByID(t, p2View, p2.id).HasConfirmedCharacter {
		t.Fatal("confirming p1 must not confirm p2")
	}
}

func TestProtocolV2StorytellerCannotStartUntilEveryPlayerConfirmsCharacter(t *testing.T) {
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
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4, p5} {
		harness.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}
	harness.command(storyteller, ClientMessage{
		Type: MsgAssignCharacters,
		Assignments: map[string]string{
			p1.id: "slayer",
			p2.id: "soldier",
			p3.id: "mayor",
			p4.id: "poisoner",
			p5.id: "imp",
		},
	})
	for _, player := range []*protocolV2GameClient{p1, p2, p3, p4} {
		harness.command(player, ClientMessage{Type: MsgConfirmCharacter})
	}

	harness.write(storyteller, ClientMessage{
		ProtocolVersion:  2,
		Type:             MsgStartGame,
		RoomID:           harness.roomID,
		PlayerID:         storyteller.id,
		ResumeCredential: storyteller.resumeCredential,
		ClientSequence:   storyteller.nextSequence,
	})
	result := readWireMessage(t, storyteller.connection)

	if result.Type != ServerMsgError || result.Code != ProtocolErrorInvalidCommand {
		t.Fatalf("start with an unconfirmed player must be rejected: %+v", result)
	}
	if result.NextClientSequence != storyteller.nextSequence {
		t.Fatalf("rejected start next sequence = %d, want %d", result.NextClientSequence, storyteller.nextSequence)
	}
}
