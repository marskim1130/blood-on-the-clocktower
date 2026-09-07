package session

import "testing"

func TestRecoveryCredentialCannotResumeWithoutApproval(t *testing.T) {
	s, _, actor := newTestSession(t)
	s.codec = NewHMACCredentialCodec([]byte("test-key-with-at-least-thirty-two-bytes"))
	credential, ok := s.RecoveryCredentialFor(actor.PlayerID)
	if !ok || credential == "" {
		t.Fatal("missing separate recovery credential")
	}
	if _, err := s.Query(Actor{PlayerID: actor.PlayerID, Credential: credential}); err == nil {
		t.Fatal("recovery credential bypassed approval")
	}
	if _, _, err := s.ValidateRecovery(actor.PlayerID, credential); err != nil {
		t.Fatal(err)
	}
}
