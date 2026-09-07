package session

import "time"

const RecoveryRequestLifetime = 10 * time.Minute

type RecoveryGrant struct {
	PlayerID  string    `json:"playerId"`
	OldNonce  string    `json:"oldNonce"`
	NewNonce  string    `json:"newNonce"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type storytellerEngine interface{ StorytellerID() string }

func recoveryApprover(record RoomRecord, engine Engine, targetID string) string {
	if source, ok := engine.(storytellerEngine); ok {
		if storyteller := source.StorytellerID(); storyteller != "" && storyteller != targetID {
			return storyteller
		}
	}
	return record.CreatorID
}

func (s *AuthoritativeGameSession) RecoveryCredentialFor(playerID string) (string, bool) {
	view := s.committed.Load()
	identity, ok := view.record.Members[playerID]
	if view.closed || !ok {
		return "", false
	}
	return s.codec.Encode(view.record.RoomID, playerID, "recovery-"+identity.CredentialNonce), true
}

func (s *AuthoritativeGameSession) ValidateRecovery(playerID, credential string) (string, string, error) {
	view := s.committed.Load()
	identity, ok := view.record.Members[playerID]
	if view.closed {
		return "", "", ErrRoomNotFound
	}
	if !ok || !s.codec.Verify(view.record.RoomID, playerID, "recovery-"+identity.CredentialNonce, credential) {
		return "", "", ErrInvalidCredential
	}
	return identity.Name, recoveryApprover(view.record, view.engine, playerID), nil
}

// ApprovedRecovery retries a committed approval after a dropped response or restart.
func (s *AuthoritativeGameSession) ApprovedRecovery(requestID, playerID, credential string) (string, bool) {
	view := s.committed.Load()
	grant, ok := view.record.RecoveryGrants[requestID]
	identity, member := view.record.Members[playerID]
	if view.closed || !ok || !member || grant.PlayerID != playerID || !time.Now().Before(grant.ExpiresAt) || identity.CredentialNonce != grant.NewNonce {
		return "", false
	}
	if !s.codec.Verify(view.record.RoomID, playerID, "recovery-"+grant.OldNonce, credential) {
		return "", false
	}
	return s.codec.Encode(view.record.RoomID, playerID, grant.NewNonce), true
}

func (s *AuthoritativeGameSession) applyRecoveryReview(record *RoomRecord, engine Engine, actor Actor, command Command) error {
	if actor.PlayerID == command.TargetID || actor.PlayerID != recoveryApprover(*record, engine, command.TargetID) {
		return ErrForbidden
	}
	target, ok := record.Members[command.TargetID]
	if !ok || !s.codec.Verify(record.RoomID, command.TargetID, "recovery-"+target.CredentialNonce, command.RecoveryCredential) {
		return ErrInvalidCredential
	}
	if !command.Decision {
		return nil
	}
	if command.RecoveryRequestID == "" || len(command.RecoveryRequestID) > 128 {
		return ErrInvalidCommand
	}
	if record.RecoveryGrants == nil {
		record.RecoveryGrants = map[string]RecoveryGrant{}
	}
	for id, grant := range record.RecoveryGrants {
		if !time.Now().Before(grant.ExpiresAt) {
			delete(record.RecoveryGrants, id)
		}
	}
	if len(record.RecoveryGrants) >= 64 {
		return ErrInvalidCommand
	}
	nonce, _, err := s.codec.Issue(record.RoomID, command.TargetID)
	if err != nil {
		return err
	}
	record.RecoveryGrants[command.RecoveryRequestID] = RecoveryGrant{PlayerID: command.TargetID, OldNonce: target.CredentialNonce, NewNonce: nonce, ExpiresAt: time.Now().Add(RecoveryRequestLifetime)}
	target.CredentialNonce = nonce
	record.Members[command.TargetID] = target
	return nil
}
