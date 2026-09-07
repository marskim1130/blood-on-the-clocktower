package ws

import (
	"encoding/json"
	"fmt"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/gameplay"
	"github.com/marskim1130/blood-on-the-clocktower/internal/session"
)

type sessionGameEngine struct {
	roomID  string
	session *gameplay.GameSession
	history engineHistory
}

func newSessionGameEngine(roomID, scriptID string) *sessionGameEngine {
	return &sessionGameEngine{roomID: roomID, session: gameplay.NewGameSession(scriptID)}
}

func (e *sessionGameEngine) Clone() session.Engine {
	return &sessionGameEngine{roomID: e.roomID, session: e.session.Clone(), history: e.history.clone()}
}

func (e *sessionGameEngine) SetRoomID(roomID string) { e.roomID = roomID }

func (e *sessionGameEngine) AddPlayer(playerID, name string) error {
	e.session.AddPlayer(game.Player{ID: playerID, Name: name, IsAlive: true})
	e.history.Undo, e.history.Redo = nil, nil
	return nil
}

func (e *sessionGameEngine) RemovePlayer(playerID string) {
	e.session.RemovePlayer(playerID)
	e.history.Undo, e.history.Redo = nil, nil
}

func (e *sessionGameEngine) Finished() bool { return e.session.Finished() }
func (e *sessionGameEngine) Restart(memberIDs []string) error {
	if err := e.session.Restart(memberIDs); err != nil {
		return err
	}
	e.history = engineHistory{}
	return nil
}

func (e *sessionGameEngine) StorytellerID() string { return e.session.StorytellerID() }
func (e *sessionGameEngine) ClearGameHistory()     { e.history.Undo, e.history.Redo = nil, nil }
func (e *sessionGameEngine) ParticipantLockState() bool {
	if e.session.Phase() != game.GamePhaseSetup {
		return true
	}
	for _, player := range e.session.Players() {
		if player.Character != nil {
			return true
		}
	}
	return false
}

func (e *sessionGameEngine) Execute(actorID string, payload any) (bool, error) {
	if command, ok := payload.(historyCommand); ok {
		return e.executeHistory(actorID, command)
	}
	command, ok := payload.(gameplay.Command)
	if !ok {
		return false, fmt.Errorf("invalid game command payload")
	}
	return e.executeRecorded(actorID, command)
}

func (e *sessionGameEngine) Project(recipientID string) any {
	state := roomStateFromProjection(e.roomID, e.session.ProjectionFor(recipientID))
	if state != nil && recipientID != "" && recipientID == e.session.StorytellerID() {
		for _, entry := range e.history.Logs {
			state.OperationLog = append(state.OperationLog, entry)
		}
		state.CanUndo, state.CanRedo = len(e.history.Undo) > 0, len(e.history.Redo) > 0
		if state.CanUndo {
			last := e.history.Undo[len(e.history.Undo)-1]
			state.UndoCrossesPhase = last.FromPhase != last.ToPhase
		}
		if state.CanRedo {
			last := e.history.Redo[len(e.history.Redo)-1]
			state.RedoCrossesPhase = last.FromPhase != last.ToPhase
		}
	}
	return state
}

func (e *sessionGameEngine) Marshal() ([]byte, error) { return e.marshalWithHistory() }

func loadSessionGameEngine(roomID string, data []byte) (session.Engine, error) {
	gameSession, err := gameplay.LoadGameSessionSnapshot(data)
	if err != nil {
		return nil, err
	}
	var saved struct {
		History engineHistory `json:"operationHistory"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}
	return &sessionGameEngine{roomID: roomID, session: gameSession, history: saved.History}, nil
}

func roomStateFromProjection(roomID string, projection *gameplay.Projection) *RoomState {
	if projection == nil {
		return nil
	}
	return &RoomState{
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
