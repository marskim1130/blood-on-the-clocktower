package ws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/gameplay"
	"github.com/marskim1130/blood-on-the-clocktower/internal/session"
)

func (h *Hub) dispatchRoomMessage(conn Connection, msg ClientMessage, generation uint64) {
	var handle func(Connection, ClientMessage)
	requiresConnection := true
	connectionError := session.ErrStaleConnection
	switch msg.Type {
	case MsgJoinRoom:
		handle, requiresConnection = h.handleJoinRoomV2, false
	case MsgResumeRoom:
		handle, requiresConnection = h.handleResumeRoomV2, false
	case MsgRequestRecovery:
		handle, requiresConnection = h.handleRequestRecovery, false
	case MsgReviewRecovery:
		handle = h.handleReviewRecovery
	case MsgGetRecoveryRequests:
		handle = h.handleGetRecoveryRequests
	case MsgRejoinRoom, MsgLeaveRoom, MsgKickPlayer, MsgUpdateRoomSettings, MsgSetStoryteller,
		MsgSetSeatOrder, MsgSetReady, MsgConfirmCharacter,
		MsgAdvanceNominationStage, MsgControlNominationTimer, MsgExpireNominationTimer,
		MsgTransferOwnership, MsgRestartGame, MsgUndoGame, MsgRedoGame,
		MsgAssignCharacters, MsgSubmitEvent, MsgStartGame, MsgChangePhase, MsgNominate, MsgCastVote,
		MsgRecordVote, MsgResolveNomination, MsgFinalizeDay, MsgExecutePlayer, MsgUseSlayerAbility, MsgKillPlayer, MsgSubmitNightAction,
		MsgConfirmNightAction, MsgAcknowledgeNightAction, MsgSkipNightAction, MsgPrepareDawn, MsgConfirmDawn,
		MsgResolveNight, MsgEndGame, MsgPublishGrimoire, MsgCloseRoom:
		handle, requiresConnection = h.handleCommandV2, msg.Type != MsgCloseRoom
	case MsgGetRoomState:
		handle, connectionError = h.handleGetRoomStateV2, session.ErrInvalidCredential
	default:
		h.sendV2Error(conn, errors.New("unsupported message type"), 0)
		return
	}
	if requiresConnection && !h.active.AcceptsGeneration(msg.RoomID, msg.PlayerID, generation) {
		h.sendV2Error(conn, connectionError, 0)
		return
	}
	handle(conn, msg)
}

func (h *Hub) createRoomV2(msg ClientMessage) (session.CreateResult, error) {
	if msg.RequestID == "" || msg.PlayerID == "" {
		return session.CreateResult{}, errors.New("requestId and playerId are required")
	}
	if msg.MaxPlayers < 5 || msg.MaxPlayers > 15 {
		return session.CreateResult{}, fmt.Errorf("%w: maxPlayers must be between 5 and 15", session.ErrInvalidCommand)
	}
	scriptID := strings.TrimSpace(msg.ScriptID)
	if scriptID == "" {
		scriptID = game.TroubleBrewingScriptID
	}
	if game.GetScriptByID(scriptID) == nil {
		return session.CreateResult{}, fmt.Errorf("%w: unsupported script", session.ErrInvalidCommand)
	}
	engine := newSessionGameEngine("", scriptID)
	return h.registry.Create(context.Background(), session.CreateInput{RequestID: msg.RequestID, PlayerID: msg.PlayerID, PlayerName: msg.PlayerName, MaxPlayers: msg.MaxPlayers, ScriptID: scriptID, Fingerprint: fingerprint(msg), Engine: engine})
}

func (h *Hub) attachCreatedRoom(conn Connection, result session.CreateResult) {
	s, ok := h.registry.Get(result.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	view, err := s.Query(session.Actor{PlayerID: result.PlayerID, Credential: result.ResumeCredential})
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	h.attachConnection(result.RoomID, result.PlayerID, conn)
	state, _ := view.Room.(*RoomState)
	decorateRoomState(state, view.Metadata)
	recoveryCredential, _ := h.recoveryCredential(result.RoomID, result.PlayerID)
	h.outbound.send(conn, ServerMessage{Type: ServerMsgCreateRoomResult, RoomID: result.RoomID, State: state, IdentityStatus: view.Identity, ResumeCredential: result.ResumeCredential, RecoveryCredential: recoveryCredential, RoomRevision: view.RoomRevision, NextClientSequence: view.NextClientSequence})
}

func (h *Hub) handleJoinRoomV2(conn Connection, msg ClientMessage) {
	s, ok := h.registry.Get(msg.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	result, err := s.JoinObserved(context.Background(), session.JoinInput{RequestID: msg.JoinRequestID, PlayerID: msg.PlayerID, PlayerName: msg.PlayerName, Fingerprint: fingerprint(msg)}, nil)
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	h.attachConnection(msg.RoomID, msg.PlayerID, conn)
	state, _ := result.State.(*RoomState)
	decorateRoomState(state, result.Metadata)
	recoveryCredential, _ := h.recoveryCredential(msg.RoomID, msg.PlayerID)
	h.outbound.send(conn, ServerMessage{Type: ServerMsgJoinRoomResult, RoomID: msg.RoomID, State: state, IdentityStatus: memberIdentityStatus(result.Metadata, result.NextClientSequence), ResumeCredential: result.ResumeCredential, RecoveryCredential: recoveryCredential, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence})
	h.enqueueDeliveries(msg.RoomID, msg.PlayerID, result.RoomRevision, result.Deliveries, result.Metadata)
}

func (h *Hub) handleResumeRoomV2(conn Connection, msg ClientMessage) {
	s, ok := h.registry.Get(msg.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	result, err := s.Query(session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential})
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	h.attachConnection(msg.RoomID, msg.PlayerID, conn)
	state, _ := result.Room.(*RoomState)
	decorateRoomState(state, result.Metadata)
	recoveryCredential, _ := s.RecoveryCredentialFor(msg.PlayerID)
	h.outbound.send(conn, ServerMessage{Type: ServerMsgResumeRoomResult, RoomID: msg.RoomID, State: state, IdentityStatus: result.Identity, RecoveryCredential: recoveryCredential, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence})
}

func (h *Hub) handleGetRoomStateV2(conn Connection, msg ClientMessage) {
	s, ok := h.registry.Get(msg.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	result, err := s.Query(session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential})
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	state, _ := result.Room.(*RoomState)
	decorateRoomState(state, result.Metadata)
	h.outbound.send(conn, ServerMessage{Type: ServerMsgRoomState, RoomID: msg.RoomID, State: state, IdentityStatus: result.Identity, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence})
}

func (h *Hub) handleCommandV2(conn Connection, msg ClientMessage) {
	detached := msg.Type == MsgCloseRoom
	s, ok := h.registry.Get(msg.RoomID)
	if !ok {
		h.sendV2Error(conn, session.ErrRoomNotFound, 0)
		return
	}
	command, err := toSessionCommand(msg)
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	result, err := s.Execute(context.Background(), session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential, ClientSequence: msg.ClientSequence, Detached: detached}, command)
	if err != nil {
		nextClientSequence := uint64(0)
		if queryResult, queryErr := s.Query(session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential}); queryErr == nil {
			nextClientSequence = queryResult.NextClientSequence
		}
		h.sendV2Error(conn, err, nextClientSequence)
		return
	}
	h.enqueueCommandResult(conn, msg.RoomID, result)
	if !result.Duplicate {
		h.enqueueDeliveries(msg.RoomID, msg.PlayerID, result.RoomRevision, result.Deliveries, result.Metadata)
	}
	if msg.Type == MsgCloseRoom {
		h.registry.Remove(msg.RoomID)
	}
	for _, effect := range result.ConnectionEffects {
		switch effect {
		case session.CloseSender:
			h.active.Revoke(msg.RoomID, msg.PlayerID)
			h.outbound.closeWhenDrained(conn)
		case session.CloseTarget:
			if target := h.active.Revoke(msg.RoomID, result.TargetPlayerID); target != nil {
				h.outbound.sendAndClose(target, ServerMessage{Type: ServerMsgKicked, RoomID: msg.RoomID})
			}
		case session.CloseAll:
			for _, target := range h.active.RemoveRoom(msg.RoomID) {
				h.outbound.sendAndClose(target, ServerMessage{Type: ServerMsgRoomClosed, RoomID: msg.RoomID})
			}
		}
	}
}

func (h *Hub) enqueueCommandResult(connection Connection, roomID string, result session.CommandResult) {
	// Disconnection cancels delivery, not an admitted command. Detached close
	// may legitimately reply on a connection that has no member attachment.
	if _, _, _, connected := h.active.Identity(connection); !connected && !hasConnectionEffect(result.ConnectionEffects, session.CloseAll) {
		return
	}
	state, _ := result.DirectResponse.(*RoomState)
	var identityStatus *session.IdentityStatus
	if status, ok := result.DirectResponse.(session.IdentityStatus); ok {
		identityStatus = &status
	} else if state != nil && !hasConnectionEffect(result.ConnectionEffects, session.CloseSender) {
		identityStatus = memberIdentityStatus(result.Metadata, result.NextClientSequence)
	}
	decorateRoomState(state, result.Metadata)
	h.outbound.send(connection, ServerMessage{Type: ServerMsgCommandResult, RoomID: roomID, State: state, IdentityStatus: identityStatus, RoomRevision: result.RoomRevision, AcceptedSequence: result.AcceptedSequence, NextClientSequence: result.NextClientSequence})
}

func (h *Hub) enqueueDeliveries(roomID, excludePlayerID string, revision uint64, deliveries []session.Delivery, metadata session.RoomMetadata) {
	for _, delivery := range deliveries {
		if delivery.PlayerID == excludePlayerID {
			continue
		}
		connection, _, ok := h.active.Get(roomID, delivery.PlayerID)
		if !ok {
			continue
		}
		state, _ := delivery.Payload.(*RoomState)
		decorateRoomState(state, metadata)
		h.outbound.send(connection, ServerMessage{Type: ServerMsgRoomStateChanged, RoomID: roomID, State: state, IdentityStatus: memberIdentityStatus(metadata, delivery.NextClientSequence), RoomRevision: revision})
	}
}

func memberIdentityStatus(metadata session.RoomMetadata, nextClientSequence uint64) *session.IdentityStatus {
	return &session.IdentityStatus{
		Status:               session.IdentityMember,
		CanRejoin:            false,
		NextClientSequence:   nextClientSequence,
		ParticipantSetFrozen: metadata.ParticipantSetFrozen,
	}
}

func hasConnectionEffect(effects []session.ConnectionEffect, target session.ConnectionEffect) bool {
	for _, effect := range effects {
		if effect == target {
			return true
		}
	}
	return false
}

func toSessionCommand(msg ClientMessage) (session.Command, error) {
	command := session.Command{Fingerprint: fingerprint(msg), TargetID: msg.targetPlayerID(), MaxPlayers: msg.MaxPlayers, ScriptID: msg.ScriptID, FreezeParticipants: msg.Type == MsgAssignCharacters}
	switch msg.Type {
	case MsgRejoinRoom:
		command.Kind = session.CommandRejoin
	case MsgLeaveRoom:
		command.Kind = session.CommandLeave
	case MsgKickPlayer:
		command.Kind = session.CommandKick
	case MsgCloseRoom:
		command.Kind = session.CommandClose
	case MsgTransferOwnership:
		command.Kind = session.CommandTransferOwnership
	case MsgRestartGame:
		command.Kind = session.CommandRestartGame
	case MsgUndoGame:
		command.Kind = session.CommandUndoGame
		command.Payload = gameplay.UndoGameCmd{SenderID: msg.PlayerID, ConfirmPhaseChange: msg.ConfirmPhaseChange}
	case MsgRedoGame:
		command.Kind = session.CommandRedoGame
		command.Payload = gameplay.RedoGameCmd{SenderID: msg.PlayerID, ConfirmPhaseChange: msg.ConfirmPhaseChange}
	case MsgUpdateRoomSettings:
		command.Kind = session.CommandUpdateSettings
		command.Payload = gameplay.UpdateRoomSettingsCmd{SenderID: msg.PlayerID, MaxPlayers: msg.MaxPlayers, ScriptID: msg.ScriptID}
	case MsgSetStoryteller:
		command.Kind = session.CommandSetStoryteller
		command.Payload = gameplay.SetStorytellerCmd{SenderID: msg.PlayerID, TargetPlayerID: msg.TargetPlayerID}
	case MsgSetSeatOrder:
		command.Kind = session.CommandSetSeatOrder
		command.Payload = gameplay.SetSeatOrderCmd{SenderID: msg.PlayerID, SeatOrder: msg.SeatOrder}
	default:
		gameCommand, err := toGameCommand(msg)
		if err != nil {
			return session.Command{}, err
		}
		command.Kind = session.CommandGame
		command.Payload = gameCommand
	}
	return command, nil
}

func toGameCommand(msg ClientMessage) (gameplay.Command, error) {
	switch msg.Type {
	case MsgSetStoryteller:
		return gameplay.SetStorytellerCmd{SenderID: msg.PlayerID, TargetPlayerID: msg.TargetPlayerID}, nil
	case MsgSetReady:
		if msg.Ready == nil {
			return nil, errors.New("ready is required")
		}
		return gameplay.SetReadyCmd{SenderID: msg.PlayerID, Ready: *msg.Ready}, nil
	case MsgConfirmCharacter:
		return gameplay.ConfirmCharacterCmd{SenderID: msg.PlayerID}, nil
	case MsgAssignCharacters:
		return gameplay.AssignCharactersCmd{
			SenderID:                  msg.PlayerID,
			Assignments:               msg.Assignments,
			ShownCharacters:           msg.ShownCharacters,
			FortuneTellerRedHerringID: msg.FortuneTellerRedHerringID,
			DemonBluffCharacterIDs:    msg.DemonBluffCharacterIDs,
		}, nil
	case MsgSubmitEvent:
		if msg.Event == nil {
			return nil, errors.New("event is required")
		}
		return gameplay.SubmitEventCmd{SenderID: msg.PlayerID, Event: *msg.Event}, nil
	case MsgStartGame:
		return gameplay.StartGameCmd{SenderID: msg.PlayerID}, nil
	case MsgChangePhase:
		return gameplay.ChangePhaseCmd{SenderID: msg.PlayerID, Phase: msg.Phase.GamePhase()}, nil
	case MsgNominate:
		return gameplay.NominateCmd{SenderID: msg.PlayerID, NomineeID: msg.NomineeID}, nil
	case MsgCastVote:
		if msg.Decision == nil {
			return nil, errors.New("decision is required")
		}
		return gameplay.CastVoteCmd{SenderID: msg.PlayerID, Decision: *msg.Decision}, nil
	case MsgAdvanceNominationStage:
		return gameplay.AdvanceNominationStageCmd{SenderID: msg.PlayerID}, nil
	case MsgControlNominationTimer:
		return gameplay.ControlNominationTimerCmd{SenderID: msg.PlayerID, Action: msg.TimerAction}, nil
	case MsgExpireNominationTimer:
		return gameplay.ExpireNominationTimerCmd{SenderID: msg.PlayerID, DeadlineUnixMs: msg.DeadlineUnixMs}, nil
	case MsgRecordVote:
		if msg.Decision == nil {
			return nil, errors.New("decision is required")
		}
		if msg.TargetPlayerID == "" {
			return nil, errors.New("targetPlayerId is required")
		}
		return gameplay.RecordVoteCmd{SenderID: msg.PlayerID, VoterID: msg.TargetPlayerID, Decision: *msg.Decision}, nil
	case MsgResolveNomination:
		return gameplay.ResolveNominationCmd{SenderID: msg.PlayerID}, nil
	case MsgFinalizeDay:
		return gameplay.FinalizeDayCmd{SenderID: msg.PlayerID}, nil
	case MsgExecutePlayer:
		return gameplay.ExecutePlayerCmd{SenderID: msg.PlayerID, PlayerID: msg.targetPlayerID()}, nil
	case MsgUseSlayerAbility:
		return gameplay.UseSlayerAbilityCmd{SenderID: msg.PlayerID, TargetPlayerID: msg.TargetPlayerID}, nil
	case MsgKillPlayer:
		return gameplay.KillPlayerCmd{SenderID: msg.PlayerID, PlayerID: msg.targetPlayerID(), Cause: msg.Cause.DeathCause()}, nil
	case MsgSubmitNightAction:
		return gameplay.SubmitNightActionCmd{SenderID: msg.PlayerID, ActionType: msg.ActionType, TargetIDs: msg.TargetIDs, Result: msg.Result}, nil
	case MsgConfirmNightAction:
		return gameplay.ConfirmNightActionCmd{SenderID: msg.PlayerID, TargetIDs: msg.TargetIDs, Result: msg.Result}, nil
	case MsgAcknowledgeNightAction:
		return gameplay.AcknowledgeNightActionCmd{SenderID: msg.PlayerID}, nil
	case MsgSkipNightAction:
		return gameplay.SkipNightActionCmd{SenderID: msg.PlayerID}, nil
	case MsgPrepareDawn:
		return gameplay.PrepareDawnCmd{SenderID: msg.PlayerID}, nil
	case MsgConfirmDawn:
		return gameplay.ConfirmDawnCmd{SenderID: msg.PlayerID, DeathPlayerIDs: msg.TargetIDs}, nil
	case MsgResolveNight:
		return nil, errors.New("prepare and confirm dawn before entering day")
	case MsgEndGame:
		return gameplay.EndGameCmd{SenderID: msg.PlayerID, Winner: msg.Winner.Team(), Reason: game.WinReasonStorytellerDecision, Description: msg.Description}, nil
	case MsgPublishGrimoire:
		return gameplay.PublishGrimoireCmd{SenderID: msg.PlayerID}, nil
	}
	return nil, errors.New("unsupported command")
}

func decorateRoomState(state *RoomState, metadata session.RoomMetadata) {
	if state == nil {
		return
	}
	state.RoomID = metadata.RoomID
	state.CreatorID = metadata.CreatorID
	state.MaxPlayers = metadata.MaxPlayers
	if state.ScriptID == "" {
		state.ScriptID = metadata.ScriptID
	}
}

func fingerprint(msg ClientMessage) string {
	copy := msg
	copy.ResumeCredential = ""
	data, _ := json.Marshal(copy)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func (h *Hub) sendV2Error(conn Connection, err error, nextClientSequence uint64) {
	code := ProtocolErrorInternal
	switch {
	case errors.Is(err, session.ErrRoomNotFound):
		code = ProtocolErrorRoomNotFound
	case errors.Is(err, session.ErrInvalidCredential):
		code = ProtocolErrorInvalidCredential
	case errors.Is(err, session.ErrStaleConnection):
		code = ProtocolErrorStaleConnection
	case errors.Is(err, session.ErrForbidden):
		code = ProtocolErrorForbidden
	case errors.Is(err, session.ErrParticipantSetFrozen):
		code = ProtocolErrorParticipantSetFrozen
	case errors.Is(err, session.ErrUnexpectedSequence):
		code = ProtocolErrorUnexpectedSequence
	case errors.Is(err, session.ErrSequenceConflict):
		code = ProtocolErrorSequenceConflict
	case errors.Is(err, session.ErrIdempotencyConflict):
		code = ProtocolErrorIdempotencyConflict
	case errors.Is(err, session.ErrPersistenceUnavailable):
		code = ProtocolErrorPersistenceUnavailable
	case errors.Is(err, session.ErrPersistenceConflict):
		code = ProtocolErrorPersistenceConflict
	case errors.Is(err, session.ErrRoomFull):
		code = ProtocolErrorRoomFull
	case errors.Is(err, session.ErrInvalidCommand):
		code = ProtocolErrorInvalidCommand
	}
	h.outbound.send(conn, ServerMessage{Type: ServerMsgError, Code: code, Error: err.Error(), NextClientSequence: nextClientSequence})
}
