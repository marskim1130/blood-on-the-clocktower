package ws

import (
	"reflect"
	"testing"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

func TestProtocolV2FinishedMemberLeaveClearsHistoryAndPreservesFinalSnapshot(t *testing.T) {
	h := newProtocolV2GameHarness(t)
	st := h.create("st", "主持")
	players := []*protocolV2GameClient{
		h.join("p1", "玩家一"),
		h.join("p2", "玩家二"),
		h.join("p3", "玩家三"),
		h.join("p4", "玩家四"),
		h.join("p5", "玩家五"),
	}
	h.command(st, ClientMessage{Type: MsgSetStoryteller, TargetPlayerID: st.id})
	ready := true
	for _, player := range players {
		h.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}
	h.command(st, ClientMessage{
		Type: MsgAssignCharacters,
		Assignments: map[string]string{
			"p1": "slayer",
			"p2": "soldier",
			"p3": "mayor",
			"p4": "poisoner",
			"p5": "imp",
		},
		DemonBluffCharacterIDs: []string{"chef", "empath", "fortuneteller"},
	})
	for _, player := range players {
		h.command(player, ClientMessage{Type: MsgConfirmCharacter})
	}
	h.command(st, ClientMessage{Type: MsgStartGame})
	h.command(st, ClientMessage{Type: MsgEndGame, Winner: ClientTeam(game.TeamGood)})
	h.command(st, ClientMessage{Type: MsgPublishGrimoire})
	undone := h.command(st, ClientMessage{Type: MsgUndoGame})
	before := undone.direct.State
	if before.Phase != game.GamePhaseFinished || !before.CanUndo || !before.CanRedo || len(before.OperationLog) == 0 {
		t.Fatalf("退席前必须处于终局且同时存在撤销、重做和日志: %+v", before)
	}

	departing := players[0]
	left := h.command(departing, ClientMessage{Type: MsgLeaveRoom})
	after := left.broadcasts[st.id].State
	if after.CanUndo || after.CanRedo {
		t.Fatalf("终局退席后仍可撤销或重做成员变更前的游戏: %+v", after)
	}
	if !reflect.DeepEqual(after.OperationLog, before.OperationLog) {
		t.Fatal("终局退席改变了说书人私密操作日志")
	}
	if after.Phase != game.GamePhaseFinished || !reflect.DeepEqual(after.Winner, before.Winner) || !reflect.DeepEqual(after.Players, before.Players) {
		t.Fatal("终局退席改变了胜负或最终玩家快照")
	}
	if left.direct.State.Phase != game.GamePhaseFinished || !reflect.DeepEqual(left.direct.State.Winner, before.Winner) || len(left.direct.State.Players) != len(players) {
		t.Fatalf("退席应答未保留最终对局: %+v", left.direct.State)
	}
	finalPlayer := playerByID(t, left.direct.State, departing.id)
	previousPlayer := playerByID(t, before, departing.id)
	if finalPlayer.IsAlive != previousPlayer.IsAlive || !reflect.DeepEqual(finalPlayer.Character, previousPlayer.Character) {
		t.Fatal("退席应答丢失了离开玩家的最终身份或生死状态")
	}

	for _, messageType := range []string{MsgUndoGame, MsgRedoGame} {
		h.write(st, ClientMessage{
			ProtocolVersion:    2,
			Type:               messageType,
			RoomID:             h.roomID,
			PlayerID:           st.id,
			ResumeCredential:   st.resumeCredential,
			ClientSequence:     st.nextSequence,
			ConfirmPhaseChange: true,
		})
		rejected := readWireMessage(t, st.connection)
		if rejected.Type != ServerMsgError || rejected.Code != ProtocolErrorInvalidCommand || rejected.NextClientSequence != st.nextSequence {
			t.Fatalf("退席后 %s 未被拒绝或消耗了命令序号: %+v", messageType, rejected)
		}
	}
}

func TestProtocolV2AssignmentUndoRedoSynchronizesParticipantLockAndPreservesIdentity(t *testing.T) {
	h := newProtocolV2GameHarness(t)
	st := h.create("st", "主持")
	players := []*protocolV2GameClient{
		h.join("p1", "玩家一"),
		h.join("p2", "玩家二"),
		h.join("p3", "玩家三"),
		h.join("p4", "玩家四"),
		h.join("p5", "玩家五"),
	}
	h.command(st, ClientMessage{Type: MsgSetStoryteller, TargetPlayerID: st.id})
	ready := true
	for _, player := range players {
		h.command(player, ClientMessage{Type: MsgSetReady, Ready: &ready})
	}
	assignments := map[string]string{
		"p1": "slayer",
		"p2": "soldier",
		"p3": "mayor",
		"p4": "poisoner",
		"p5": "imp",
	}
	assertCommit := func(label string, committed protocolV2GameCommit, frozen bool) {
		t.Helper()
		identity := committed.direct.IdentityStatus
		if identity == nil || identity.Status != "member" || identity.ParticipantSetFrozen != frozen || identity.NextClientSequence != st.nextSequence {
			t.Fatalf("%s 未同步说书人的成员锁定或命令序号: %+v", label, identity)
		}
		state := committed.direct.State
		if state.Phase != game.GamePhaseSetup || len(state.Players) != len(players) {
			t.Fatalf("%s 改变了准备阶段或成员数量: %+v", label, state)
		}
		for index, player := range state.Players {
			if player.ID != players[index].id {
				t.Fatalf("%s 改变了成员座位顺序: %+v", label, state.Players)
			}
			if !frozen && (player.Character != nil || player.ShownCharacter != nil) {
				t.Fatalf("撤销分配后说书人投影仍有角色: %+v", player)
			}
		}
		if frozen {
			assertCharacterVisibility(t, state, st.id, assignments, true)
		}
		for _, player := range players {
			broadcast := committed.broadcasts[player.id]
			identity := broadcast.IdentityStatus
			if identity == nil || identity.Status != "member" || identity.ParticipantSetFrozen != frozen || identity.NextClientSequence != player.nextSequence {
				t.Fatalf("%s 未同步 %s 的成员锁定或保留序号: %+v", label, player.id, identity)
			}
			if len(broadcast.State.OperationLog) != 0 || broadcast.State.CanUndo || broadcast.State.CanRedo || broadcast.ResumeCredential != "" || broadcast.RecoveryCredential != "" {
				t.Fatalf("%s 向 %s 泄露了私密历史或凭据", label, player.id)
			}
			if frozen {
				assertCharacterVisibility(t, broadcast.State, player.id, assignments, false)
			} else {
				for _, projected := range broadcast.State.Players {
					if projected.Character != nil || projected.ShownCharacter != nil {
						t.Fatalf("撤销分配后 %s 投影仍有角色: %+v", player.id, projected)
					}
				}
			}
		}
	}

	assigned := h.command(st, ClientMessage{
		Type:                   MsgAssignCharacters,
		Assignments:            assignments,
		DemonBluffCharacterIDs: []string{"chef", "empath", "fortuneteller"},
	})
	assertCommit("角色分配", assigned, true)
	undone := h.command(st, ClientMessage{Type: MsgUndoGame})
	assertCommit("撤销分配", undone, false)
	if !undone.direct.State.CanRedo {
		t.Fatal("撤销角色分配后无法重做")
	}
	redone := h.command(st, ClientMessage{Type: MsgRedoGame})
	assertCommit("重做分配", redone, true)
	if redone.direct.State.CanRedo {
		t.Fatal("重做角色分配后仍有未消费的重做步骤")
	}

	player := players[0]
	resumed := h.resume(player)
	if resumed.IdentityStatus == nil || resumed.IdentityStatus.Status != "member" || !resumed.IdentityStatus.ParticipantSetFrozen || resumed.IdentityStatus.NextClientSequence != player.nextSequence {
		t.Fatalf("原凭据恢复未保留成员身份、锁定或序号: %+v", resumed.IdentityStatus)
	}
	assertCharacterVisibility(t, resumed.State, player.id, assignments, false)
	h.write(player, ClientMessage{
		ProtocolVersion:  2,
		Type:             MsgLeaveRoom,
		RoomID:           h.roomID,
		PlayerID:         player.id,
		ResumeCredential: player.resumeCredential,
		ClientSequence:   player.nextSequence,
	})
	rejected := readWireMessage(t, player.connection)
	if rejected.Type != ServerMsgError || rejected.Code != ProtocolErrorParticipantSetFrozen || rejected.NextClientSequence != player.nextSequence {
		t.Fatalf("重做分配后成员锁定未阻止退席或错误消耗序号: %+v", rejected)
	}
}
