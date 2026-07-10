package ws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
	"github.com/your-org/blood-on-the-clocktower/internal/session"
)

func (h *Hub) handleMessageV2(conn Connection, msg ClientMessage) {
	switch msg.Type {
	case MsgCreateRoom:
		h.handleCreateRoomV2(conn, msg)
	case MsgJoinRoom:
		h.handleJoinRoomV2(conn, msg)
	case MsgResumeRoom:
		h.handleResumeRoomV2(conn, msg)
	case MsgRejoinRoom, MsgLeaveRoom, MsgKickPlayer, MsgUpdateRoomSettings, MsgSetStoryteller,
		MsgAssignCharacters, MsgSubmitEvent, MsgStartGame, MsgChangePhase, MsgNominate, MsgCastVote,
		MsgResolveNomination, MsgExecutePlayer, MsgUseSlayerAbility, MsgKillPlayer, MsgSubmitNightAction,
		MsgResolveNight, MsgEndGame, MsgCloseRoom:
		h.handleCommandV2(conn, msg)
	case MsgGetRoomState:
		h.handleGetRoomStateV2(conn, msg)
	default:
		h.sendV2Error(conn, errors.New("unsupported message type"), 0)
	}
}

func (h *Hub) handleCreateRoomV2(conn Connection, msg ClientMessage) {
	if msg.RequestID == "" || msg.PlayerID == "" {
		h.sendV2Error(conn, errors.New("requestId and playerId are required"), 0)
		return
	}
	engine := newSessionGameEngine("", msg.ScriptID)
	result, err := h.registry.Create(context.Background(), session.CreateInput{RequestID: msg.RequestID, PlayerID: msg.PlayerID, PlayerName: msg.PlayerName, MaxPlayers: msg.MaxPlayers, ScriptID: msg.ScriptID, Fingerprint: fingerprint(msg), Engine: engine})
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	generation, previous := h.active.Takeover(result.RoomID, result.PlayerID, conn)
	_ = generation
	if previous != nil {
		_ = previous.Close()
	}
	state, _ := result.State.(*RoomState)
	if s, ok := h.registry.Get(result.RoomID); ok {
		decorateRoomState(state, s.Metadata())
	}
	_ = conn.SendJSON(ServerMessage{Type: "CREATE_ROOM_RESULT", RoomID: result.RoomID, State: state, ResumeCredential: result.ResumeCredential, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence})
}

func (h *Hub) handleJoinRoomV2(conn Connection, msg ClientMessage) {
	result, err := h.registry.Join(context.Background(), session.JoinInput{RequestID: msg.JoinRequestID, RoomID: msg.RoomID, PlayerID: msg.PlayerID, PlayerName: msg.PlayerName, Fingerprint: fingerprint(msg)})
	if err != nil {
		h.sendV2Error(conn, err, 0)
		return
	}
	_, previous := h.active.Takeover(msg.RoomID, msg.PlayerID, conn)
	if previous != nil {
		_ = previous.Close()
	}
	state, _ := result.State.(*RoomState)
	if s, ok := h.registry.Get(msg.RoomID); ok {
		decorateRoomState(state, s.Metadata())
	}
	_ = conn.SendJSON(ServerMessage{Type: "JOIN_ROOM_RESULT", RoomID: msg.RoomID, State: state, ResumeCredential: result.ResumeCredential, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence})
	h.broadcastSessionState(msg.RoomID)
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
	_, previous := h.active.Takeover(msg.RoomID, msg.PlayerID, conn)
	if previous != nil {
		_ = previous.Close()
	}
	state, _ := result.Room.(*RoomState)
	decorateRoomState(state, s.Metadata())
	_ = conn.SendJSON(ServerMessage{Type: "RESUME_ROOM_RESULT", RoomID: msg.RoomID, State: state, IdentityStatus: result.Identity, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence})
}

func (h *Hub) handleGetRoomStateV2(conn Connection, msg ClientMessage) {
	if !h.currentConnection(conn, msg.RoomID, msg.PlayerID) {
		h.sendV2Error(conn, session.ErrInvalidCredential, 0)
		return
	}
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
	decorateRoomState(state, s.Metadata())
	_ = conn.SendJSON(ServerMessage{Type: "ROOM_STATE", RoomID: msg.RoomID, State: state, IdentityStatus: result.Identity, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence})
}

func (h *Hub) handleCommandV2(conn Connection, msg ClientMessage) {
	detached := msg.Type == MsgCloseRoom
	if !detached && !h.currentConnection(conn, msg.RoomID, msg.PlayerID) {
		h.sendV2Error(conn, session.ErrStaleConnection, 0)
		return
	}
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
	var result session.CommandResult
	execute := func() error {
		var executeErr error
		result, executeErr = s.Execute(context.Background(), session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential, ClientSequence: msg.ClientSequence, Detached: detached}, command)
		return executeErr
	}
	if detached {
		err = execute()
	} else {
		err = h.active.WithCurrent(msg.RoomID, msg.PlayerID, conn, execute)
		if errors.Is(err, errStaleActiveConnection) {
			err = session.ErrStaleConnection
		}
	}
	if err != nil {
		nextClientSequence := uint64(0)
		if queryResult, queryErr := s.Query(session.Actor{PlayerID: msg.PlayerID, Credential: msg.ResumeCredential}); queryErr == nil {
			nextClientSequence = queryResult.NextClientSequence
		}
		h.sendV2Error(conn, err, nextClientSequence)
		return
	}
	if msg.Type == MsgCloseRoom {
		h.registry.Remove(msg.RoomID)
	}
	state, _ := result.DirectResponse.(*RoomState)
	var identityStatus any
	if status, ok := result.DirectResponse.(session.IdentityStatus); ok {
		identityStatus = status
	}
	decorateRoomState(state, s.Metadata())
	_ = conn.SendJSON(ServerMessage{Type: "COMMAND_RESULT", RoomID: msg.RoomID, State: state, IdentityStatus: identityStatus, RoomRevision: result.RoomRevision, AcceptedSequence: result.AcceptedSequence, NextClientSequence: result.NextClientSequence})
	if !result.Duplicate {
		h.broadcastCommittedResult(msg.RoomID, result, s.Metadata())
	}
	for _, effect := range result.ConnectionEffects {
		switch effect {
		case session.CloseSender:
			_ = conn.Close()
			h.active.Remove(conn)
		case session.CloseTarget:
			if target, _, ok := h.active.Get(msg.RoomID, result.TargetPlayerID); ok {
				_ = target.SendJSON(ServerMessage{Type: "KICKED", RoomID: msg.RoomID})
				_ = target.Close()
				h.active.Remove(target)
			}
		case session.CloseAll:
			for _, target := range h.active.RemoveRoom(msg.RoomID) {
				_ = target.SendJSON(ServerMessage{Type: "ROOM_CLOSED", RoomID: msg.RoomID})
				_ = target.Close()
			}
		}
	}
}

func (h *Hub) broadcastCommittedResult(roomID string, result session.CommandResult, metadata session.RoomMetadata) {
	for _, delivery := range result.Deliveries {
		connection, _, ok := h.active.Get(roomID, delivery.PlayerID)
		if !ok {
			continue
		}
		state, _ := delivery.Payload.(*RoomState)
		decorateRoomState(state, metadata)
		if err := connection.SendJSON(ServerMessage{Type: "ROOM_STATE_CHANGED", RoomID: roomID, State: state, RoomRevision: result.RoomRevision}); err != nil {
			_ = connection.Close()
			h.active.Remove(connection)
		}
	}
}

func toSessionCommand(msg ClientMessage) (session.Command, error) {
	command := session.Command{Fingerprint: fingerprint(msg), TargetID: msg.targetPlayerID(), MaxPlayers: msg.MaxPlayers, ScriptID: msg.ScriptID, FreezeParticipants: msg.Type == MsgStartGame}
	switch msg.Type {
	case MsgRejoinRoom:
		command.Kind = session.CommandRejoin
	case MsgLeaveRoom:
		command.Kind = session.CommandLeave
	case MsgKickPlayer:
		command.Kind = session.CommandKick
	case MsgCloseRoom:
		command.Kind = session.CommandClose
	case MsgUpdateRoomSettings:
		command.Kind = session.CommandUpdateSettings
		command.Payload = UpdateRoomSettingsCmd{SenderID: msg.PlayerID, MaxPlayers: msg.MaxPlayers, ScriptID: msg.ScriptID}
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

func toGameCommand(msg ClientMessage) (Command, error) {
	switch msg.Type {
	case MsgSetStoryteller:
		return SetStorytellerCmd{SenderID: msg.PlayerID, TargetPlayerID: msg.TargetPlayerID}, nil
	case MsgAssignCharacters:
		return AssignCharactersCmd{SenderID: msg.PlayerID, Assignments: msg.Assignments, ShownCharacters: msg.ShownCharacters, FortuneTellerRedHerringID: msg.FortuneTellerRedHerringID}, nil
	case MsgSubmitEvent:
		if msg.Event == nil {
			return nil, errors.New("event is required")
		}
		return SubmitEventCmd{SenderID: msg.PlayerID, Event: *msg.Event}, nil
	case MsgStartGame:
		return StartGameCmd{SenderID: msg.PlayerID}, nil
	case MsgChangePhase:
		return ChangePhaseCmd{SenderID: msg.PlayerID, Phase: msg.Phase.GamePhase()}, nil
	case MsgNominate:
		return NominateCmd{SenderID: msg.PlayerID, NomineeID: msg.NomineeID}, nil
	case MsgCastVote:
		if msg.Decision == nil {
			return nil, errors.New("decision is required")
		}
		return CastVoteCmd{SenderID: msg.PlayerID, Decision: *msg.Decision}, nil
	case MsgResolveNomination:
		return ResolveNominationCmd{SenderID: msg.PlayerID}, nil
	case MsgExecutePlayer:
		return ExecutePlayerCmd{SenderID: msg.PlayerID, PlayerID: msg.targetPlayerID()}, nil
	case MsgUseSlayerAbility:
		return UseSlayerAbilityCmd{SenderID: msg.PlayerID, TargetPlayerID: msg.TargetPlayerID}, nil
	case MsgKillPlayer:
		return KillPlayerCmd{SenderID: msg.PlayerID, PlayerID: msg.targetPlayerID(), Cause: msg.Cause.DeathCause()}, nil
	case MsgSubmitNightAction:
		return SubmitNightActionCmd{SenderID: msg.PlayerID, ActionType: msg.ActionType, TargetIDs: msg.TargetIDs, Result: msg.Result}, nil
	case MsgResolveNight:
		return ResolveNightCmd{SenderID: msg.PlayerID}, nil
	case MsgEndGame:
		return EndGameCmd{SenderID: msg.PlayerID, Winner: msg.Winner.Team(), Reason: game.WinReasonStorytellerDecision, Description: msg.Description}, nil
	}
	return nil, errors.New("unsupported command")
}

func (h *Hub) currentConnection(conn Connection, roomID, playerID string) bool {
	current, generation, ok := h.active.Get(roomID, playerID)
	return ok && current == conn && h.active.IsCurrent(roomID, playerID, conn, generation)
}
func (h *Hub) broadcastSessionState(roomID string) {
	s, ok := h.registry.Get(roomID)
	if !ok {
		return
	}
	for playerID, conn := range h.active.Room(roomID) {
		result, err := s.Query(session.Actor{PlayerID: playerID, Credential: h.credentialFor(roomID, playerID)})
		if err != nil {
			continue
		}
		state, _ := result.Room.(*RoomState)
		decorateRoomState(state, s.Metadata())
		if err := conn.SendJSON(ServerMessage{Type: "ROOM_STATE_CHANGED", RoomID: roomID, State: state, RoomRevision: result.RoomRevision, NextClientSequence: result.NextClientSequence}); err != nil {
			_ = conn.Close()
			h.active.Remove(conn)
		}
	}
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

func (h *Hub) credentialFor(roomID, playerID string) string {
	s, ok := h.registry.Get(roomID)
	if !ok {
		return ""
	}
	credential, _ := s.CredentialFor(playerID)
	return credential
}
func fingerprint(msg ClientMessage) string {
	copy := msg
	copy.ResumeCredential = ""
	data, _ := json.Marshal(copy)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func (h *Hub) sendV2Error(conn Connection, err error, nextClientSequence uint64) {
	code := "INTERNAL"
	switch {
	case errors.Is(err, session.ErrRoomNotFound):
		code = "ROOM_NOT_FOUND"
	case errors.Is(err, session.ErrInvalidCredential):
		code = "INVALID_CREDENTIAL"
	case errors.Is(err, session.ErrStaleConnection):
		code = "STALE_CONNECTION"
	case errors.Is(err, session.ErrForbidden):
		code = "FORBIDDEN"
	case errors.Is(err, session.ErrParticipantSetFrozen):
		code = "PARTICIPANT_SET_FROZEN"
	case errors.Is(err, session.ErrUnexpectedSequence):
		code = "UNEXPECTED_SEQUENCE"
	case errors.Is(err, session.ErrSequenceConflict):
		code = "SEQUENCE_CONFLICT"
	case errors.Is(err, session.ErrIdempotencyConflict):
		code = "IDEMPOTENCY_CONFLICT"
	case errors.Is(err, session.ErrPersistenceUnavailable):
		code = "PERSISTENCE_UNAVAILABLE"
	case errors.Is(err, session.ErrPersistenceConflict):
		code = "PERSISTENCE_CONFLICT"
	}
	_ = conn.SendJSON(ServerMessage{Type: "ERROR", Code: code, Error: err.Error(), NextClientSequence: nextClientSequence})
}
