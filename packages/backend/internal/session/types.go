package session

import (
	"errors"

	"github.com/your-org/blood-on-the-clocktower/internal/sessionstore"
)

const SchemaVersion = 1

var (
	ErrRoomNotFound           = errors.New("room not found")
	ErrInvalidCredential      = errors.New("invalid credential")
	ErrStaleConnection        = errors.New("stale connection")
	ErrForbidden              = errors.New("forbidden")
	ErrParticipantSetFrozen   = errors.New("participant set frozen")
	ErrUnexpectedSequence     = errors.New("unexpected client sequence")
	ErrSequenceConflict       = errors.New("client sequence conflict")
	ErrIdempotencyConflict    = errors.New("idempotency conflict")
	ErrPersistenceConflict    = errors.New("persistence conflict")
	ErrPersistenceUnavailable = errors.New("persistence unavailable")
)

type StoreRecord = sessionstore.Record
type Store = sessionstore.RoomRecordStore

type CredentialCodec interface {
	Issue(roomID, playerID string) (nonce string, credential string, err error)
	Encode(roomID, playerID, nonce string) string
	Verify(roomID, playerID, nonce, credential string) bool
}

type IDGenerator interface {
	RoomID() (string, error)
}

type IdentityState string

const (
	IdentityMember   IdentityState = "member"
	IdentityRetained IdentityState = "retained"
)

type Identity struct {
	PlayerID        string        `json:"playerId"`
	Name            string        `json:"name"`
	CredentialNonce string        `json:"credentialNonce"`
	State           IdentityState `json:"state"`
	LastSequence    uint64        `json:"lastSequence"`
	LastFingerprint string        `json:"lastFingerprint,omitempty"`
	LastResult      ResultSummary `json:"lastResult,omitempty"`
}

type ResultSummary struct {
	RoomRevision       uint64 `json:"roomRevision"`
	AcceptedSequence   uint64 `json:"acceptedSequence"`
	NextClientSequence uint64 `json:"nextClientSequence"`
}

type IdempotencyRecord struct {
	Fingerprint string `json:"fingerprint"`
	PlayerID    string `json:"playerId"`
}

type RoomRecord struct {
	SchemaVersion        int                          `json:"schemaVersion"`
	RoomID               string                       `json:"roomId"`
	RoomRevision         uint64                       `json:"roomRevision"`
	CreatorID            string                       `json:"creatorId"`
	MaxPlayers           int                          `json:"maxPlayers"`
	ScriptID             string                       `json:"scriptId"`
	ParticipantSetFrozen bool                         `json:"participantSetFrozen"`
	Members              map[string]Identity          `json:"members"`
	Retained             map[string]Identity          `json:"retained"`
	Bans                 map[string]bool              `json:"bans"`
	CreateRequests       map[string]IdempotencyRecord `json:"createRequests"`
	JoinRequests         map[string]IdempotencyRecord `json:"joinRequests"`
	Game                 []byte                       `json:"game"`
}

type Actor struct {
	PlayerID       string
	Credential     string
	ClientSequence uint64
	Detached       bool
}

type CommandKind string

const (
	CommandRejoin         CommandKind = "REJOIN_ROOM"
	CommandLeave          CommandKind = "LEAVE_ROOM"
	CommandKick           CommandKind = "KICK_PLAYER"
	CommandClose          CommandKind = "CLOSE_ROOM"
	CommandUpdateSettings CommandKind = "UPDATE_ROOM_SETTINGS"
	CommandSetStoryteller CommandKind = "SET_STORYTELLER"
	CommandGame           CommandKind = "GAME_COMMAND"
)

type Command struct {
	Kind               CommandKind
	Fingerprint        string
	TargetID           string
	MaxPlayers         int
	ScriptID           string
	Payload            any
	FreezeParticipants bool
}

type Delivery struct {
	PlayerID string
	Payload  any
}

type ConnectionEffect string

const (
	CloseSender ConnectionEffect = "close_sender"
	CloseTarget ConnectionEffect = "close_target"
	CloseAll    ConnectionEffect = "close_all"
)

type CommandResult struct {
	AcceptedSequence   uint64
	NextClientSequence uint64
	RoomRevision       uint64
	Duplicate          bool
	DirectResponse     any
	Deliveries         []Delivery
	ConnectionEffects  []ConnectionEffect
	TargetPlayerID     string
	Metadata           RoomMetadata
}

type CommitObserver func(CommandResult)

type IdentityStatus struct {
	Status               IdentityState `json:"status"`
	CanRejoin            bool          `json:"canRejoin"`
	NextClientSequence   uint64        `json:"nextClientSequence"`
	ParticipantSetFrozen bool          `json:"participantSetFrozen"`
}

type QueryResult struct {
	RoomRevision       uint64
	NextClientSequence uint64
	Room               any
	Identity           *IdentityStatus
	Metadata           RoomMetadata
}

type JoinObserver func(JoinResult)

type Engine interface {
	Clone() Engine
	SetRoomID(roomID string)
	AddPlayer(playerID, name string) error
	RemovePlayer(playerID string)
	Execute(actorID string, payload any) (updated bool, err error)
	Project(recipientID string) any
	Marshal() ([]byte, error)
}

type RoomMetadata struct {
	RoomID               string
	CreatorID            string
	MaxPlayers           int
	ScriptID             string
	ParticipantSetFrozen bool
}

type EngineLoader func([]byte) (Engine, error)
