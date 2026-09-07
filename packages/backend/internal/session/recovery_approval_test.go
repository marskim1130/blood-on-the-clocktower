package session

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func recoveryApprovalFixture(t *testing.T) (*AuthoritativeGameSession, *memoryStore, CredentialCodec, Actor, Actor) {
	t.Helper()
	_, store, _ := newTestSession(t)
	var record RoomRecord
	if err := json.Unmarshal(store.records["room"].Data, &record); err != nil {
		t.Fatal(err)
	}
	var engine testEngine
	if err := json.Unmarshal(record.Game, &engine); err != nil {
		t.Fatal(err)
	}
	codec := NewHMACCredentialCodec([]byte("recovery-approval-test-key-at-least-32-bytes"))
	s := NewAuthoritativeGameSession(store, codec, record, &engine)
	return s, store, codec,
		Actor{PlayerID: "creator", Credential: codec.Encode("room", "creator", "nonce"), ClientSequence: 1},
		Actor{PlayerID: "member", Credential: codec.Encode("room", "member", "nonce"), ClientSequence: 1}
}

func TestLaterRecoveryRotationInvalidatesEarlierApprovedGrant(t *testing.T) {
	s, _, _, approver, target := recoveryApprovalFixture(t)
	firstRecovery, _ := s.RecoveryCredentialFor(target.PlayerID)
	if _, err := s.Execute(context.Background(), approver, Command{Kind: CommandReviewRecovery, Fingerprint: "first", TargetID: target.PlayerID, RecoveryCredential: firstRecovery, RecoveryRequestID: "first", Decision: true}); err != nil {
		t.Fatal(err)
	}
	firstResume, ok := s.ApprovedRecovery("first", target.PlayerID, firstRecovery)
	if !ok {
		t.Fatal("missing first approval")
	}
	secondRecovery, _ := s.RecoveryCredentialFor(target.PlayerID)
	approver.ClientSequence++
	if _, err := s.Execute(context.Background(), approver, Command{Kind: CommandReviewRecovery, Fingerprint: "second", TargetID: target.PlayerID, RecoveryCredential: secondRecovery, RecoveryRequestID: "second", Decision: true}); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.ApprovedRecovery("first", target.PlayerID, firstRecovery); ok {
		t.Fatal("old grant can mint a superseded credential")
	}
	secondResume, ok := s.ApprovedRecovery("second", target.PlayerID, secondRecovery)
	if !ok || secondResume == firstResume {
		t.Fatal("second approval did not rotate again")
	}
	if _, err := s.Query(Actor{PlayerID: target.PlayerID, Credential: firstResume}); !errors.Is(err, ErrInvalidCredential) {
		t.Fatalf("superseded credential stayed valid: %v", err)
	}
}

func TestApprovedRecoveryRetriesAfterRestartWithoutReapplyingApproval(t *testing.T) {
	s, store, codec, approver, target := recoveryApprovalFixture(t)
	recovery, _ := s.RecoveryCredentialFor(target.PlayerID)
	command := Command{Kind: CommandReviewRecovery, Fingerprint: "approve-persist", TargetID: target.PlayerID, RecoveryCredential: recovery, RecoveryRequestID: "request-persist", Decision: true}
	if _, err := s.Execute(context.Background(), approver, command); err != nil {
		t.Fatal(err)
	}
	credential, ok := s.ApprovedRecovery("request-persist", target.PlayerID, recovery)
	if !ok {
		t.Fatal("missing first grant")
	}
	duplicate, err := s.Execute(context.Background(), approver, command)
	if err != nil || !duplicate.Duplicate || duplicate.RoomRevision != 2 {
		t.Fatalf("approval retry reapplied state: %+v %v", duplicate, err)
	}
	var record RoomRecord
	if err := json.Unmarshal(store.records["room"].Data, &record); err != nil {
		t.Fatal(err)
	}
	var engine testEngine
	if err := json.Unmarshal(record.Game, &engine); err != nil {
		t.Fatal(err)
	}
	restarted := NewAuthoritativeGameSession(store, codec, record, &engine)
	for attempt := 0; attempt < 2; attempt++ {
		retried, ok := restarted.ApprovedRecovery("request-persist", target.PlayerID, recovery)
		if !ok || retried != credential {
			t.Fatal("restart lost idempotent approved credential")
		}
	}
	if _, err := restarted.Query(Actor{PlayerID: target.PlayerID, Credential: credential}); err != nil {
		t.Fatalf("persisted new credential cannot resume: %v", err)
	}
	if _, err := restarted.Query(target); !errors.Is(err, ErrInvalidCredential) {
		t.Fatalf("restart restored revoked credential: %v", err)
	}
	for _, input := range []struct{ request, player, secret string }{
		{"wrong-request", target.PlayerID, recovery},
		{"request-persist", approver.PlayerID, recovery},
		{"request-persist", target.PlayerID, target.Credential},
		{"request-persist", target.PlayerID, recovery + "tampered"},
	} {
		if _, ok := restarted.ApprovedRecovery(input.request, input.player, input.secret); ok {
			t.Fatal("grant leaked to wrong request, player, or credential")
		}
	}
}

func TestRecoveryReviewRejectsUnauthorizedAndSelfApproval(t *testing.T) {
	for _, scenario := range []struct{ name, actor, target string }{
		{"member is not approver", "member", "creator"},
		{"member cannot self approve", "member", "member"},
		{"creator cannot self approve", "creator", "creator"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			s, _, _, creator, member := recoveryApprovalFixture(t)
			actor := creator
			if scenario.actor == member.PlayerID {
				actor = member
			}
			recovery, _ := s.RecoveryCredentialFor(scenario.target)
			_, err := s.Execute(context.Background(), actor, Command{Kind: CommandReviewRecovery, Fingerprint: "forbidden", TargetID: scenario.target, RecoveryCredential: recovery, RecoveryRequestID: "forbidden", Decision: true})
			if !errors.Is(err, ErrForbidden) {
				t.Fatalf("unauthorized approval result: %v", err)
			}
			if _, ok := s.ApprovedRecovery("forbidden", scenario.target, recovery); ok {
				t.Fatal("forbidden grant was published")
			}
			if view, err := s.Query(actor); err != nil || view.RoomRevision != 1 || view.NextClientSequence != 1 {
				t.Fatalf("forbidden command changed state: %+v %v", view, err)
			}
		})
	}
}

func TestRecoveryApprovalPersistenceFailureLeavesOriginalCredentialsAndSequenceIntact(t *testing.T) {
	s, store, _, approver, target := recoveryApprovalFixture(t)
	recovery, _ := s.RecoveryCredentialFor(target.PlayerID)
	command := Command{Kind: CommandReviewRecovery, Fingerprint: "approve-rollback", TargetID: target.PlayerID, RecoveryCredential: recovery, RecoveryRequestID: "request-rollback", Decision: true}
	store.failReplace = true
	if _, err := s.Execute(context.Background(), approver, command); !errors.Is(err, ErrPersistenceUnavailable) {
		t.Fatalf("expected persistence failure: %v", err)
	}
	if _, ok := s.ApprovedRecovery("request-rollback", target.PlayerID, recovery); ok {
		t.Fatal("uncommitted approval grant leaked")
	}
	if _, err := s.Query(target); err != nil {
		t.Fatalf("old resume revoked before persistence: %v", err)
	}
	if _, _, err := s.ValidateRecovery(target.PlayerID, recovery); err != nil {
		t.Fatalf("old recovery revoked before persistence: %v", err)
	}
	view, err := s.Query(approver)
	if err != nil || view.RoomRevision != 1 || view.NextClientSequence != 1 {
		t.Fatalf("uncommitted approval advanced room or sequence: %+v %v", view, err)
	}
	store.failReplace = false
	if _, err := s.Execute(context.Background(), approver, command); err != nil {
		t.Fatalf("same approval cannot retry after store recovery: %v", err)
	}
	if credential, ok := s.ApprovedRecovery("request-rollback", target.PlayerID, recovery); !ok || credential == target.Credential {
		t.Fatal("retried approval did not rotate nonce")
	}
}

func TestRecoveryApprovalRotatesSignedNonceAndInvalidatesOldResume(t *testing.T) {
	s, _, _, approver, target := recoveryApprovalFixture(t)
	recovery, ok := s.RecoveryCredentialFor(target.PlayerID)
	if !ok {
		t.Fatal("missing recovery credential")
	}
	name, expectedApprover, err := s.ValidateRecovery(target.PlayerID, recovery)
	if err != nil || name != "Member" || expectedApprover != approver.PlayerID {
		t.Fatalf("recovery validation: %q %q %v", name, expectedApprover, err)
	}
	command := Command{Kind: CommandReviewRecovery, Fingerprint: "approve-1", TargetID: target.PlayerID, RecoveryCredential: recovery, RecoveryRequestID: "request-1", Decision: true}
	result, err := s.Execute(context.Background(), approver, command)
	if err != nil {
		t.Fatal(err)
	}
	if result.RoomRevision != 2 {
		t.Fatalf("approval revision = %d", result.RoomRevision)
	}
	credential, ok := s.ApprovedRecovery("request-1", target.PlayerID, recovery)
	if !ok || credential == "" || credential == target.Credential || credential == recovery {
		t.Fatal("approval did not issue a new signed resume credential")
	}
	if _, err := s.Query(target); !errors.Is(err, ErrInvalidCredential) {
		t.Fatalf("old resume remained valid: %v", err)
	}
	if _, err := s.Query(Actor{PlayerID: target.PlayerID, Credential: credential}); err != nil {
		t.Fatalf("new resume rejected: %v", err)
	}
	if _, _, err := s.ValidateRecovery(target.PlayerID, recovery); !errors.Is(err, ErrInvalidCredential) {
		t.Fatalf("old recovery code remained reusable: %v", err)
	}
	newRecovery, ok := s.RecoveryCredentialFor(target.PlayerID)
	if !ok || newRecovery == recovery {
		t.Fatal("recovery credential did not rotate")
	}
}
