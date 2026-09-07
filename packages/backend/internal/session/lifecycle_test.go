package session

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func TestCreatorTransfersOwnershipToMember(t *testing.T) {
	s, _, actor := newTestSession(t)
	result, err := s.Execute(context.Background(), actor, Command{Kind: "TRANSFER_OWNERSHIP", TargetID: "member", Fingerprint: "transfer", Payload: 0})
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata.CreatorID != "member" {
		t.Fatalf("creator = %s", result.Metadata.CreatorID)
	}
}

type restartEngine struct {
	*testEngine
	RestartedMembers []string `json:"restartedMembers"`
}

func (e *restartEngine) Clone() Engine {
	return &restartEngine{testEngine: e.testEngine.Clone().(*testEngine), RestartedMembers: slices.Clone(e.RestartedMembers)}
}
func (e *restartEngine) Restart(memberIDs []string) error {
	e.FinishedFlag = false
	e.Value = 0
	e.RestartedMembers = slices.Clone(memberIDs)
	return nil
}
func (e *restartEngine) Marshal() ([]byte, error) { return json.Marshal(e) }

func TestCreatorRestartsFinishedRoomKeepingMembers(t *testing.T) {
	s, _, actor := newTestSession(t)
	current := s.committed.Load()
	current.record.ParticipantSetFrozen = true
	e := current.engine.(*testEngine)
	e.FinishedFlag = true
	e.Value = 99
	current.engine = &restartEngine{testEngine: e}
	result, err := s.Execute(context.Background(), actor, Command{Kind: "RESTART_GAME", Fingerprint: "restart", Payload: 0})
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata.ParticipantSetFrozen {
		t.Fatal("participants remain frozen")
	}
	updated := s.committed.Load().engine.(*restartEngine)
	if updated.Finished() || updated.Value != 0 || !slices.Equal(updated.RestartedMembers, []string{"creator", "member"}) {
		t.Fatalf("not restarted: %+v", updated)
	}
}

func TestLifecycleCommandsRejectUnauthorizedAndInvalidRequests(t *testing.T) {
	for _, tc := range []struct {
		name          string
		kind          CommandKind
		actor, target string
		want          error
	}{
		{"member cannot transfer", CommandTransferOwnership, "member", "creator", ErrForbidden},
		{"target must be member", CommandTransferOwnership, "creator", "missing", ErrInvalidCommand},
		{"member cannot restart", CommandRestartGame, "member", "", ErrForbidden},
		{"unfinished cannot restart", CommandRestartGame, "creator", "", ErrInvalidCommand},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, _, _ := newTestSession(t)
			_, err := s.Execute(context.Background(), Actor{PlayerID: tc.actor, Credential: "room:" + tc.actor + ":nonce", ClientSequence: 1}, Command{Kind: tc.kind, TargetID: tc.target, Fingerprint: tc.name})
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if s.committed.Load().record.RoomRevision != 1 {
				t.Fatal("rejected command mutated revision")
			}
		})
	}
}

func TestLifecyclePersistenceFailureDoesNotPublishOrConsumeSequence(t *testing.T) {
	for _, kind := range []CommandKind{CommandTransferOwnership, CommandRestartGame} {
		t.Run(string(kind), func(t *testing.T) {
			s, store, actor := newTestSession(t)
			current := s.committed.Load()
			current.record.ParticipantSetFrozen = true
			e := current.engine.(*testEngine)
			e.FinishedFlag = true
			e.Value = 99
			current.engine = &restartEngine{testEngine: e}
			store.failReplace = true
			command := Command{Kind: kind, TargetID: "member", Fingerprint: string(kind)}
			_, err := s.Execute(context.Background(), actor, command)
			if !errors.Is(err, ErrPersistenceUnavailable) {
				t.Fatal(err)
			}
			unchanged := s.committed.Load()
			if unchanged.record.CreatorID != "creator" || !unchanged.record.ParticipantSetFrozen || unchanged.record.Members["creator"].LastSequence != 0 || !unchanged.engine.Finished() {
				t.Fatal("uncommitted lifecycle state leaked")
			}
			store.failReplace = false
			if _, err := s.Execute(context.Background(), actor, command); err != nil {
				t.Fatal(err)
			}
			duplicate, err := s.Execute(context.Background(), actor, command)
			if err != nil || !duplicate.Duplicate {
				t.Fatalf("retry = %+v, %v", duplicate, err)
			}
		})
	}
}
