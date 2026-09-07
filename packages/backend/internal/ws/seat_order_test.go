package ws

import (
	"slices"
	"testing"
)

func TestProtocolV2CreatorSetsClockwiseSeatOrderBeforeSetupIsLocked(t *testing.T) {
	harness := newProtocolV2GameHarness(t)
	storyteller := harness.create("storyteller", "Storyteller")
	p1 := harness.join("p1", "P1")
	p2 := harness.join("p2", "P2")
	p3 := harness.join("p3", "P3")

	harness.command(storyteller, ClientMessage{
		Type:           MsgSetStoryteller,
		TargetPlayerID: storyteller.id,
	})
	want := []string{p3.id, p1.id, p2.id}
	commit := harness.command(storyteller, ClientMessage{
		Type:      MsgSetSeatOrder,
		SeatOrder: want,
	})

	projections := []*RoomState{commit.direct.State}
	for _, player := range []*protocolV2GameClient{p1, p2, p3} {
		projections = append(projections, commit.broadcasts[player.id].State)
	}
	for _, state := range projections {
		got := make([]string, len(state.Players))
		for index, player := range state.Players {
			got[index] = player.ID
		}
		if !slices.Equal(got, want) {
			t.Fatalf("clockwise seat order = %v, want %v", got, want)
		}
	}
}
