package ws

import (
	"fmt"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/gameplay"
	"github.com/marskim1130/blood-on-the-clocktower/internal/session"
)

type sessionGameEngine struct {
	roomID  string
	session *gameplay.GameSession
}

func newSessionGameEngine(roomID, scriptID string) *sessionGameEngine {
	return &sessionGameEngine{roomID: roomID, session: gameplay.NewGameSession(scriptID)}
}

func (e *sessionGameEngine) Clone() session.Engine {
	return &sessionGameEngine{roomID: e.roomID, session: e.session.Clone()}
}

func (e *sessionGameEngine) SetRoomID(roomID string) { e.roomID = roomID }

func (e *sessionGameEngine) AddPlayer(playerID, name string) error {
	e.session.AddPlayer(game.Player{ID: playerID, Name: name, IsAlive: true})
	return nil
}

func (e *sessionGameEngine) RemovePlayer(playerID string) {
	e.session.RemovePlayer(playerID)
}

func (e *sessionGameEngine) Finished() bool { return e.session.Finished() }
func (e *sessionGameEngine) Restart(memberIDs []string) error {
	return e.session.Restart(memberIDs)
}

func (e *sessionGameEngine) StorytellerID() string { return e.session.StorytellerID() }
func (e *sessionGameEngine) ClearGameHistory()     { e.session.ClearGameHistory() }
func (e *sessionGameEngine) ParticipantLockState() bool {
	return e.session.ParticipantLockState()
}

func (e *sessionGameEngine) Execute(actorID string, payload any) (bool, error) {
	// Bind game-history requests to the authenticated room actor.
	switch command := payload.(type) {
	case gameplay.UndoGameCmd:
		command.SenderID = actorID
		payload = command
	case gameplay.RedoGameCmd:
		command.SenderID = actorID
		payload = command
	}
	command, ok := payload.(gameplay.Command)
	if !ok {
		return false, fmt.Errorf("invalid game command payload")
	}
	result, err := e.session.Apply(command)
	return result.Updated, err
}

func (e *sessionGameEngine) Project(recipientID string) any {
	return roomStateFromProjection(e.roomID, e.session.ProjectionFor(recipientID))
}

func (e *sessionGameEngine) Marshal() ([]byte, error) { return e.session.MarshalSnapshot() }

func loadSessionGameEngine(roomID string, data []byte) (session.Engine, error) {
	gameSession, err := gameplay.LoadGameSessionSnapshot(data)
	if err != nil {
		return nil, err
	}
	return &sessionGameEngine{roomID: roomID, session: gameSession}, nil
}

func roomStateFromProjection(roomID string, projection *gameplay.Projection) *RoomState {
	if projection == nil {
		return nil
	}
	var operationLog []OperationLogEntry
	for _, entry := range projection.OperationLog {
		operationLog = append(operationLog, OperationLogEntry(entry))
	}
	return &RoomState{
		OperationLog:              operationLog,
		CanUndo:                   projection.CanUndo,
		CanRedo:                   projection.CanRedo,
		UndoCrossesPhase:          projection.UndoCrossesPhase,
		RedoCrossesPhase:          projection.RedoCrossesPhase,
		RoomID:                    roomID,
		Players:                   projection.Players,
		ScriptID:                  projection.ScriptID,
		ScriptName:                projection.ScriptName,
		StorytellerID:             projection.StorytellerID,
		StorytellerName:           projection.StorytellerName,
		Phase:                     projection.Phase,
		DayNumber:                 projection.DayNumber,
		NightNumber:               projection.NightNumber,
		Nomination:                projection.Nomination,
		NominationResults:         projection.NominationResults,
		ExecutionCandidateID:      projection.ExecutionCandidateID,
		ExecutionCandidateVotes:   projection.ExecutionCandidateVotes,
		ExecutionTied:             projection.ExecutionTied,
		Deaths:                    projection.Deaths,
		GhostVotesRemaining:       projection.GhostVotesRemaining,
		NightWakeSteps:            projection.NightWakeSteps,
		CurrentNightWakeIndex:     projection.CurrentNightWakeIndex,
		CurrentNightWakeStep:      projection.CurrentNightWakeStep,
		Winner:                    projection.Winner,
		NightActions:              projection.NightActions,
		NightTurnStatus:           projection.NightTurnStatus,
		PendingNightAction:        projection.PendingNightAction,
		ConfirmedNightAction:      projection.ConfirmedNightAction,
		DemonBluffCharacterIDs:    projection.DemonBluffCharacterIDs,
		GrimoireRevealed:          projection.GrimoireRevealed,
		DawnReviewPending:         projection.DawnReviewPending,
		PendingDawnDeathIDs:       projection.PendingDawnDeathIDs,
		FortuneTellerRedHerringID: projection.FortuneTellerRedHerringID,
	}
}
