package ws

import "testing"

func TestClientCannotBypassDawnReviewWithLegacyResolveNight(t *testing.T) {
	if _, err := toGameCommand(ClientMessage{Type: MsgResolveNight, PlayerID: "storyteller"}); err == nil {
		t.Fatal("legacy command must not bypass the private dawn review")
	}
}
