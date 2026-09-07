package ws

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/gameplay"
)

const maxUndoSteps = 30
const maxOperationLogs = 100

type historyCommand struct{ redo, confirmPhaseChange bool }
type historyEntry struct {
	Before    []byte         `json:"before"`
	After     []byte         `json:"after"`
	FromPhase game.GamePhase `json:"fromPhase"`
	ToPhase   game.GamePhase `json:"toPhase"`
}
type operationLogEntry = OperationLogEntry
type engineHistory struct {
	Undo   []historyEntry      `json:"undo,omitempty"`
	Redo   []historyEntry      `json:"redo,omitempty"`
	Logs   []operationLogEntry `json:"logs,omitempty"`
	NextID uint64              `json:"nextId"`
}

func (h engineHistory) clone() engineHistory {
	h.Undo = slices.Clone(h.Undo)
	h.Redo = slices.Clone(h.Redo)
	h.Logs = slices.Clone(h.Logs)
	return h
}

func (e *sessionGameEngine) appendOperation(actorID, action string, from, to game.GamePhase) {
	e.history.NextID++
	e.history.Logs = append(e.history.Logs, operationLogEntry{ID: e.history.NextID, ActorID: actorID, Action: action, CreatedAtUnixMs: time.Now().UnixMilli(), FromPhase: from, ToPhase: to, DisclosureWarning: true})
	if len(e.history.Logs) > maxOperationLogs {
		e.history.Logs = slices.Clone(e.history.Logs[len(e.history.Logs)-maxOperationLogs:])
	}
}

func (e *sessionGameEngine) executeRecorded(actorID string, command gameplay.Command) (bool, error) {
	before, err := restorableGameSnapshot(e.session)
	if err != nil {
		return false, err
	}
	from := e.session.Phase()
	result, err := e.session.Apply(command)
	if err != nil || !result.Updated {
		return result.Updated, err
	}
	after, err := restorableGameSnapshot(e.session)
	if err != nil {
		return false, err
	}
	e.history.Redo = nil
	switch command.(type) {
	case gameplay.SetStorytellerCmd, gameplay.SetSeatOrderCmd, gameplay.UpdateRoomSettingsCmd:
		e.history.Undo = nil
	default:
		e.history.Undo = append(e.history.Undo, historyEntry{Before: before, After: after, FromPhase: from, ToPhase: e.session.Phase()})
		if len(e.history.Undo) > maxUndoSteps {
			e.history.Undo = slices.Clone(e.history.Undo[len(e.history.Undo)-maxUndoSteps:])
		}
	}
	action := strings.TrimSuffix(strings.TrimPrefix(fmt.Sprintf("%T", command), "gameplay."), "Cmd")
	e.appendOperation(actorID, action, from, e.session.Phase())
	return true, nil
}

func restorableGameSnapshot(gs *gameplay.GameSession) ([]byte, error) {
	raw, err := gs.MarshalSnapshot()
	if err != nil {
		return nil, err
	}
	var data map[string]json.RawMessage
	if err = json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if value := data["nomination"]; len(value) > 0 && string(value) != "null" {
		var nomination game.Nomination
		if err = json.Unmarshal(value, &nomination); err != nil {
			return nil, err
		}
		if !nomination.Paused && nomination.DeadlineUnixMs > 0 {
			nomination.RemainingMs = max(int64(0), nomination.DeadlineUnixMs-time.Now().UnixMilli())
			nomination.DeadlineUnixMs = 0
			nomination.Paused = true
			data["nomination"], err = json.Marshal(nomination)
			if err != nil {
				return nil, err
			}
		}
	}
	return json.Marshal(data)
}

func (e *sessionGameEngine) executeHistory(actorID string, command historyCommand) (bool, error) {
	if actorID == "" || actorID != e.session.StorytellerID() {
		return false, fmt.Errorf("only storyteller can undo or redo game actions")
	}
	source := e.history.Undo
	if command.redo {
		source = e.history.Redo
	}
	if len(source) == 0 {
		return false, fmt.Errorf("no game action available to restore")
	}
	entry := source[len(source)-1]
	if entry.FromPhase != entry.ToPhase && !command.confirmPhaseChange {
		return false, fmt.Errorf("confirm phase change; disclosed information cannot be recalled")
	}
	snapshot := entry.Before
	if command.redo {
		snapshot = entry.After
	}
	restored, err := gameplay.LoadGameSessionSnapshot(snapshot)
	if err != nil {
		return false, err
	}
	from := e.session.Phase()
	e.session = restored
	action := "UndoGame"
	if command.redo {
		e.history.Redo = source[:len(source)-1]
		e.history.Undo = append(e.history.Undo, entry)
		action = "RedoGame"
	} else {
		e.history.Undo = source[:len(source)-1]
		e.history.Redo = append(e.history.Redo, entry)
	}
	e.appendOperation(actorID, action, from, e.session.Phase())
	return true, nil
}

func (e *sessionGameEngine) marshalWithHistory() ([]byte, error) {
	snapshot, err := e.session.MarshalSnapshot()
	if err != nil {
		return nil, err
	}
	var data map[string]json.RawMessage
	if err = json.Unmarshal(snapshot, &data); err != nil {
		return nil, err
	}
	history, err := json.Marshal(e.history)
	if err != nil {
		return nil, err
	}
	data["operationHistory"] = history
	return json.Marshal(data)
}
