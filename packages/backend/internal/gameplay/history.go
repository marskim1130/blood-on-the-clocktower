package gameplay

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

const maxUndoSteps = 30
const maxOperationLogs = 100

// ClearGameHistory invalidates restores after a roster change, retaining the log.
// Finished-room departures use this without removing the final game player.
func (gs *GameSession) ClearGameHistory() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.history.Undo, gs.history.Redo = nil, nil
}

// ParticipantLockState reports whether the current game requires a fixed roster.
// The authoritative room session persists this value with a restored game.
func (gs *GameSession) ParticipantLockState() bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if gs.phase != game.GamePhaseSetup {
		return true
	}
	for _, player := range gs.players {
		if player.Character != nil {
			return true
		}
	}
	return false
}

type UndoGameCmd struct {
	SenderID           string
	ConfirmPhaseChange bool
}

type RedoGameCmd struct {
	SenderID           string
	ConfirmPhaseChange bool
}

func (c UndoGameCmd) actorID() string { return c.SenderID }
func (c RedoGameCmd) actorID() string { return c.SenderID }

func (gs *GameSession) applyHistoryLocked(actorID string, redo, confirmPhaseChange bool) (ApplyResult, error) {
	if actorID == "" || actorID != gs.storytellerID {
		return ApplyResult{}, fmt.Errorf("only storyteller can undo or redo game actions")
	}
	source := gs.history.Undo
	if redo {
		source = gs.history.Redo
	}
	if len(source) == 0 {
		return ApplyResult{}, fmt.Errorf("no game action available to restore")
	}
	entry := source[len(source)-1]
	if entry.FromPhase != entry.ToPhase && !confirmPhaseChange {
		return ApplyResult{}, fmt.Errorf("confirm phase change; disclosed information cannot be recalled")
	}
	raw := entry.Before
	if redo {
		raw = entry.After
	}
	var restored gameSessionSnapshot
	if err := json.Unmarshal(raw, &restored); err != nil {
		return ApplyResult{}, err
	}
	from := gs.phase
	gs.restoreSnapshotLocked(restored)
	action := "UndoGame"
	if redo {
		gs.history.Redo = source[:len(source)-1]
		gs.history.Undo = append(gs.history.Undo, entry)
		action = "RedoGame"
	} else {
		gs.history.Undo = source[:len(source)-1]
		gs.history.Redo = append(gs.history.Redo, entry)
	}
	gs.appendOperationLocked(actorID, action, from, gs.phase)
	return ApplyResult{Updated: true}, nil
}

func (gs *GameSession) applyRecordedLocked(command Command) (ApplyResult, error) {
	before, err := gs.restorableSnapshotLocked()
	if err != nil {
		return ApplyResult{}, err
	}
	from := gs.phase
	result, err := gs.applyLocked(command)
	if err != nil || !result.Updated {
		return result, err
	}
	after, err := gs.restorableSnapshotLocked()
	if err != nil {
		return ApplyResult{}, err
	}
	gs.history.Redo = nil
	switch command.(type) {
	case SetStorytellerCmd, SetSeatOrderCmd, UpdateRoomSettingsCmd, KickPlayerCmd:
		gs.history.Undo = nil
	default:
		gs.history.Undo = append(gs.history.Undo, historyEntry{Before: before, After: after, FromPhase: from, ToPhase: gs.phase})
		if len(gs.history.Undo) > maxUndoSteps {
			gs.history.Undo = slices.Clone(gs.history.Undo[len(gs.history.Undo)-maxUndoSteps:])
		}
	}
	action := strings.TrimSuffix(strings.TrimPrefix(fmt.Sprintf("%T", command), "gameplay."), "Cmd")
	gs.appendOperationLocked(command.actorID(), action, from, gs.phase)
	return result, nil
}

func (gs *GameSession) restorableSnapshotLocked() ([]byte, error) {
	snapshot := gs.snapshotLocked()
	if n := snapshot.Nomination; n != nil && !n.Paused && n.DeadlineUnixMs > 0 {
		n.RemainingMs = max(int64(0), n.DeadlineUnixMs-time.Now().UnixMilli())
		n.DeadlineUnixMs = 0
		n.Paused = true
	}
	return json.Marshal(snapshot)
}

func (gs *GameSession) appendOperationLocked(actorID, action string, from, to game.GamePhase) {
	gs.history.NextID++
	gs.history.Logs = append(gs.history.Logs, OperationLogEntry{ID: gs.history.NextID, ActorID: actorID, Action: action, CreatedAtUnixMs: time.Now().UnixMilli(), FromPhase: from, ToPhase: to, DisclosureWarning: true})
	if len(gs.history.Logs) > maxOperationLogs {
		gs.history.Logs = slices.Clone(gs.history.Logs[len(gs.history.Logs)-maxOperationLogs:])
	}
}

// OperationLogEntry is a storyteller-only record of a game action.
type OperationLogEntry struct {
	ID                uint64         `json:"id"`
	Action            string         `json:"action"`
	ActorID           string         `json:"actorId"`
	CreatedAtUnixMs   int64          `json:"createdAtUnixMs"`
	FromPhase         game.GamePhase `json:"fromPhase"`
	ToPhase           game.GamePhase `json:"toPhase"`
	DisclosureWarning bool           `json:"disclosureWarning"`
}

type historyEntry struct {
	Before    []byte         `json:"before"`
	After     []byte         `json:"after"`
	FromPhase game.GamePhase `json:"fromPhase"`
	ToPhase   game.GamePhase `json:"toPhase"`
}

type gameHistory struct {
	Undo   []historyEntry      `json:"undo,omitempty"`
	Redo   []historyEntry      `json:"redo,omitempty"`
	Logs   []OperationLogEntry `json:"logs,omitempty"`
	NextID uint64              `json:"nextId"`
}

func (h gameHistory) clone() gameHistory {
	h.Undo = slices.Clone(h.Undo)
	h.Redo = slices.Clone(h.Redo)
	h.Logs = slices.Clone(h.Logs)
	// History payloads are immutable after creation; only the stacks are mutated.
	return h
}

func (gs *GameSession) projectHistoryLocked(projection *Projection) {
	projection.OperationLog = slices.Clone(gs.history.Logs)
	projection.CanUndo, projection.CanRedo = len(gs.history.Undo) > 0, len(gs.history.Redo) > 0
	if projection.CanUndo {
		last := gs.history.Undo[len(gs.history.Undo)-1]
		projection.UndoCrossesPhase = last.FromPhase != last.ToPhase
	}
	if projection.CanRedo {
		last := gs.history.Redo[len(gs.history.Redo)-1]
		projection.RedoCrossesPhase = last.FromPhase != last.ToPhase
	}
}
