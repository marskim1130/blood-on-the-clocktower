package ws

import (
	"encoding/json"
	"fmt"

	"github.com/your-org/blood-on-the-clocktower/internal/game"
	"github.com/your-org/blood-on-the-clocktower/internal/session"
)

type sessionGameEngine struct {
	roomID  string
	session *GameSession
}

func newSessionGameEngine(roomID, scriptID string) *sessionGameEngine {
	return &sessionGameEngine{roomID: roomID, session: NewGameSession(scriptID)}
}

func (e *sessionGameEngine) Clone() session.Engine {
	return &sessionGameEngine{roomID: e.roomID, session: newGameSessionFromSnapshot(e.session.snapshot())}
}

func (e *sessionGameEngine) SetRoomID(roomID string) { e.roomID = roomID }

func (e *sessionGameEngine) AddPlayer(playerID, name string) error {
	e.session.AddPlayer(game.Player{ID: playerID, Name: name, IsAlive: true})
	return nil
}

func (e *sessionGameEngine) RemovePlayer(playerID string) { e.session.RemovePlayer(playerID) }

func (e *sessionGameEngine) Execute(actorID string, payload any) (bool, error) {
	command, ok := payload.(Command)
	if !ok {
		return false, fmt.Errorf("invalid game command payload")
	}
	result, err := e.session.Apply(command)
	return result.Updated, err
}

func (e *sessionGameEngine) Project(recipientID string) any {
	return e.session.StateForRoomForRecipient(e.roomID, recipientID)
}

func (e *sessionGameEngine) Marshal() ([]byte, error) { return json.Marshal(e.session.snapshot()) }

func loadSessionGameEngine(roomID string, data []byte) (session.Engine, error) {
	var snapshot gameSessionSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &sessionGameEngine{roomID: roomID, session: newGameSessionFromSnapshot(snapshot)}, nil
}
