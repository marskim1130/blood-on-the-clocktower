// Code generated from proto/game.proto (a2bd378896954fea). DO NOT EDIT.
// Run: pnpm proto:generate

package ws

import (
	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
	"github.com/marskim1130/blood-on-the-clocktower/internal/session"
)

const (
	MsgCreateRoom             = "CREATE_ROOM"
	MsgJoinRoom               = "JOIN_ROOM"
	MsgResumeRoom             = "RESUME_ROOM"
	MsgRejoinRoom             = "REJOIN_ROOM"
	MsgGetRoomState           = "GET_ROOM_STATE"
	MsgCloseRoom              = "CLOSE_ROOM"
	MsgLeaveRoom              = "LEAVE_ROOM"
	MsgKickPlayer             = "KICK_PLAYER"
	MsgUpdateRoomSettings     = "UPDATE_ROOM_SETTINGS"
	MsgSetStoryteller         = "SET_STORYTELLER"
	MsgAssignCharacters       = "ASSIGN_CHARACTERS"
	MsgSubmitEvent            = "SUBMIT_EVENT"
	MsgStartGame              = "START_GAME"
	MsgChangePhase            = "CHANGE_PHASE"
	MsgNominate               = "NOMINATE"
	MsgCastVote               = "CAST_VOTE"
	MsgResolveNomination      = "RESOLVE_NOMINATION"
	MsgExecutePlayer          = "EXECUTE_PLAYER"
	MsgUseSlayerAbility       = "USE_SLAYER_ABILITY"
	MsgKillPlayer             = "KILL_PLAYER"
	MsgSubmitNightAction      = "SUBMIT_NIGHT_ACTION"
	MsgResolveNight           = "RESOLVE_NIGHT"
	MsgEndGame                = "END_GAME"
	MsgSetSeatOrder           = "SET_SEAT_ORDER"
	MsgSetReady               = "SET_READY"
	MsgConfirmCharacter       = "CONFIRM_CHARACTER"
	MsgFinalizeDay            = "FINALIZE_DAY"
	MsgRecordVote             = "RECORD_VOTE"
	MsgConfirmNightAction     = "CONFIRM_NIGHT_ACTION"
	MsgAcknowledgeNightAction = "ACKNOWLEDGE_NIGHT_ACTION"
	MsgSkipNightAction        = "SKIP_NIGHT_ACTION"
	MsgPublishGrimoire        = "PUBLISH_GRIMOIRE"
	MsgPrepareDawn            = "PREPARE_DAWN"
	MsgConfirmDawn            = "CONFIRM_DAWN"
	MsgAdvanceNominationStage = "ADVANCE_NOMINATION_STAGE"
	MsgControlNominationTimer = "CONTROL_NOMINATION_TIMER"
	MsgExpireNominationTimer  = "EXPIRE_NOMINATION_TIMER"
	MsgTransferOwnership      = "TRANSFER_OWNERSHIP"
	MsgRestartGame            = "RESTART_GAME"
	MsgUndoGame               = "UNDO_GAME"
	MsgRedoGame               = "REDO_GAME"
	MsgRequestRecovery        = "REQUEST_RECOVERY"
	MsgReviewRecovery         = "REVIEW_RECOVERY"
	MsgGetRecoveryRequests    = "GET_RECOVERY_REQUESTS"

	ServerMsgCreateRoomResult = "CREATE_ROOM_RESULT"
	ServerMsgJoinRoomResult   = "JOIN_ROOM_RESULT"
	ServerMsgResumeRoomResult = "RESUME_ROOM_RESULT"
	ServerMsgCommandResult    = "COMMAND_RESULT"
	ServerMsgRoomState        = "ROOM_STATE"
	ServerMsgRoomStateChanged = "ROOM_STATE_CHANGED"
	ServerMsgKicked           = "KICKED"
	ServerMsgRoomClosed       = "ROOM_CLOSED"
	ServerMsgError            = "ERROR"
	ServerMsgRecoveryStatus   = "RECOVERY_STATUS"
	ServerMsgRecoveryRequests = "RECOVERY_REQUESTS"
	ServerMsgSessionReplaced  = "SESSION_REPLACED"

	ProtocolErrorInvalidMessage         = "INVALID_MESSAGE"
	ProtocolErrorUnsupportedProtocol    = "UNSUPPORTED_PROTOCOL"
	ProtocolErrorRoomNotFound           = "ROOM_NOT_FOUND"
	ProtocolErrorInvalidCredential      = "INVALID_CREDENTIAL"
	ProtocolErrorStaleConnection        = "STALE_CONNECTION"
	ProtocolErrorForbidden              = "FORBIDDEN"
	ProtocolErrorParticipantSetFrozen   = "PARTICIPANT_SET_FROZEN"
	ProtocolErrorUnexpectedSequence     = "UNEXPECTED_SEQUENCE"
	ProtocolErrorSequenceConflict       = "SEQUENCE_CONFLICT"
	ProtocolErrorIdempotencyConflict    = "IDEMPOTENCY_CONFLICT"
	ProtocolErrorPersistenceUnavailable = "PERSISTENCE_UNAVAILABLE"
	ProtocolErrorPersistenceConflict    = "PERSISTENCE_CONFLICT"
	ProtocolErrorInternal               = "INTERNAL"
	ProtocolErrorRoomFull               = "ROOM_FULL"
	ProtocolErrorInvalidCommand         = "INVALID_COMMAND"
)

type ClientMessage struct {
	ProtocolVersion           int               `json:"protocolVersion"`
	Type                      string            `json:"type"`
	RequestID                 string            `json:"requestId,omitempty"`
	JoinRequestID             string            `json:"joinRequestId,omitempty"`
	ResumeCredential          string            `json:"resumeCredential,omitempty"`
	ClientSequence            uint64            `json:"clientSequence,omitempty"`
	RoomID                    string            `json:"roomId,omitempty"`
	PlayerName                string            `json:"playerName,omitempty"`
	PlayerID                  string            `json:"playerId,omitempty"`
	TargetPlayerID            string            `json:"targetPlayerId,omitempty"`
	MaxPlayers                int               `json:"maxPlayers,omitempty"`
	ScriptID                  string            `json:"scriptId,omitempty"`
	Assignments               map[string]string `json:"assignments,omitempty"`
	ShownCharacters           map[string]string `json:"shownCharacters,omitempty"`
	FortuneTellerRedHerringID string            `json:"fortuneTellerRedHerringId,omitempty"`
	Event                     *game.GameEvent   `json:"event,omitempty"`
	NomineeID                 string            `json:"nomineeId,omitempty"`
	Decision                  *bool             `json:"decision,omitempty"`
	Phase                     ClientGamePhase   `json:"phase,omitempty"`
	Winner                    ClientTeam        `json:"winner,omitempty"`
	Reason                    string            `json:"reason,omitempty"`
	Description               string            `json:"description,omitempty"`
	Cause                     ClientDeathCause  `json:"cause,omitempty"`
	ActionType                string            `json:"actionType,omitempty"`
	TargetIDs                 []string          `json:"targetIds,omitempty"`
	Result                    string            `json:"result,omitempty"`
	SeatOrder                 []string          `json:"seatOrder,omitempty"`
	Ready                     *bool             `json:"ready,omitempty"`
	DemonBluffCharacterIDs    []string          `json:"demonBluffCharacterIds,omitempty"`
	TimerAction               string            `json:"timerAction,omitempty"`
	DeadlineUnixMs            int64             `json:"deadlineUnixMs,omitempty"`
	ConfirmPhaseChange        bool              `json:"confirmPhaseChange,omitempty"`
	RecoveryCredential        string            `json:"recoveryCredential,omitempty"`
	RecoveryRequestID         string            `json:"recoveryRequestId,omitempty"`
}

type ServerMessage struct {
	Type               string                  `json:"type"`
	Code               string                  `json:"code,omitempty"`
	RoomID             string                  `json:"roomId,omitempty"`
	State              *RoomState              `json:"state,omitempty"`
	IdentityStatus     *session.IdentityStatus `json:"identityStatus,omitempty"`
	Error              string                  `json:"error,omitempty"`
	ResumeCredential   string                  `json:"resumeCredential,omitempty"`
	RoomRevision       uint64                  `json:"roomRevision,omitempty"`
	AcceptedSequence   uint64                  `json:"acceptedSequence,omitempty"`
	NextClientSequence uint64                  `json:"nextClientSequence,omitempty"`
	PlayerID           string                  `json:"playerId,omitempty"`
	RecoveryRequestID  string                  `json:"recoveryRequestId,omitempty"`
	RecoveryStatus     string                  `json:"recoveryStatus,omitempty"`
	RecoveryCredential string                  `json:"recoveryCredential,omitempty"`
	RecoveryRequests   []RecoveryRequest       `json:"recoveryRequests,omitempty"`
}

type RoomState struct {
	RoomID                    string                  `json:"roomId"`
	Players                   []game.Player           `json:"players"`
	MaxPlayers                int                     `json:"maxPlayers"`
	ScriptID                  string                  `json:"scriptId"`
	ScriptName                string                  `json:"scriptName"`
	CreatorID                 string                  `json:"creatorId,omitempty"`
	StorytellerID             string                  `json:"storytellerId,omitempty"`
	Phase                     game.GamePhase          `json:"phase"`
	DayNumber                 int32                   `json:"dayNumber"`
	Nomination                *game.Nomination        `json:"nomination,omitempty"`
	Deaths                    []game.DeathRecord      `json:"deaths,omitempty"`
	GhostVotesRemaining       []string                `json:"ghostVotesRemaining,omitempty"`
	NightWakeSteps            []game.NightWakeStep    `json:"nightWakeSteps,omitempty"`
	CurrentNightWakeIndex     int                     `json:"currentNightWakeIndex,omitempty"`
	CurrentNightWakeStep      *game.NightWakeStep     `json:"currentNightWakeStep,omitempty"`
	Winner                    *game.GameEndedEvent    `json:"winner,omitempty"`
	NightActions              []game.NightAction      `json:"nightActions,omitempty"`
	StorytellerName           string                  `json:"storytellerName,omitempty"`
	NightNumber               int32                   `json:"nightNumber"`
	FortuneTellerRedHerringID string                  `json:"fortuneTellerRedHerringId,omitempty"`
	NominationResults         []game.NominationResult `json:"nominationResults,omitempty"`
	ExecutionCandidateID      string                  `json:"executionCandidateId,omitempty"`
	ExecutionCandidateVotes   int                     `json:"executionCandidateVotes,omitempty"`
	ExecutionTied             bool                    `json:"executionTied,omitempty"`
	NightTurnStatus           string                  `json:"nightTurnStatus,omitempty"`
	PendingNightAction        *game.NightAction       `json:"pendingNightAction,omitempty"`
	ConfirmedNightAction      *game.NightAction       `json:"confirmedNightAction,omitempty"`
	DemonBluffCharacterIDs    []string                `json:"demonBluffCharacterIds,omitempty"`
	GrimoireRevealed          bool                    `json:"grimoireRevealed,omitempty"`
	DawnReviewPending         bool                    `json:"dawnReviewPending,omitempty"`
	PendingDawnDeathIDs       []string                `json:"pendingDawnDeathIds,omitempty"`
	OperationLog              []OperationLogEntry     `json:"operationLog,omitempty"`
	CanUndo                   bool                    `json:"canUndo,omitempty"`
	CanRedo                   bool                    `json:"canRedo,omitempty"`
	UndoCrossesPhase          bool                    `json:"undoCrossesPhase,omitempty"`
	RedoCrossesPhase          bool                    `json:"redoCrossesPhase,omitempty"`
}

type OperationLogEntry struct {
	ID                uint64         `json:"id"`
	Action            string         `json:"action"`
	ActorID           string         `json:"actorId"`
	CreatedAtUnixMs   int64          `json:"createdAtUnixMs"`
	FromPhase         game.GamePhase `json:"fromPhase"`
	ToPhase           game.GamePhase `json:"toPhase"`
	DisclosureWarning bool           `json:"disclosureWarning"`
}

type RecoveryRequest struct {
	RequestID         string `json:"requestId"`
	PlayerID          string `json:"playerId"`
	PlayerName        string `json:"playerName"`
	RequestedAtUnixMs int64  `json:"requestedAtUnixMs"`
}
