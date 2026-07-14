// Code generated from proto/game.proto (b836f9c2b639b731). DO NOT EDIT.
// Run: pnpm proto:generate

export type ClientMessageType = 'CREATE_ROOM' | 'JOIN_ROOM' | 'RESUME_ROOM' | 'REJOIN_ROOM' | 'GET_ROOM_STATE' | 'CLOSE_ROOM' | 'LEAVE_ROOM' | 'KICK_PLAYER' | 'UPDATE_ROOM_SETTINGS' | 'SET_STORYTELLER' | 'ASSIGN_CHARACTERS' | 'SUBMIT_EVENT' | 'START_GAME' | 'CHANGE_PHASE' | 'NOMINATE' | 'CAST_VOTE' | 'RESOLVE_NOMINATION' | 'EXECUTE_PLAYER' | 'USE_SLAYER_ABILITY' | 'KILL_PLAYER' | 'SUBMIT_NIGHT_ACTION' | 'RESOLVE_NIGHT' | 'END_GAME';
export type ServerMessageType = 'CREATE_ROOM_RESULT' | 'JOIN_ROOM_RESULT' | 'RESUME_ROOM_RESULT' | 'COMMAND_RESULT' | 'ROOM_STATE' | 'ROOM_STATE_CHANGED' | 'KICKED' | 'ROOM_CLOSED' | 'ERROR';
export type ProtocolErrorCode = 'INVALID_MESSAGE' | 'UNSUPPORTED_PROTOCOL' | 'ROOM_NOT_FOUND' | 'INVALID_CREDENTIAL' | 'STALE_CONNECTION' | 'FORBIDDEN' | 'PARTICIPANT_SET_FROZEN' | 'UNEXPECTED_SEQUENCE' | 'SEQUENCE_CONFLICT' | 'IDEMPOTENCY_CONFLICT' | 'PERSISTENCE_UNAVAILABLE' | 'PERSISTENCE_CONFLICT' | 'INTERNAL' | 'ROOM_FULL' | 'INVALID_COMMAND';
export type IdentityState = 'member' | 'retained';

export interface GameCharacter {
  readonly id: string;
  readonly name: string;
  readonly team: number;
  readonly ability: string;
}

export interface RoomPlayer {
  readonly id: string;
  readonly name: string;
  readonly character?: GameCharacter;
  readonly isAlive: boolean;
  readonly votes: number;
  readonly poisonedUntil?: number;
  readonly shownCharacter?: GameCharacter;
}

export interface RoomNomination {
  readonly nominatorId: string;
  readonly nomineeId: string;
  readonly votes: Readonly<Record<string, boolean>>;
  readonly resolved: boolean;
}

export interface RoomDeathRecord {
  readonly playerId: string;
  readonly cause: string;
  readonly dayNumber: number;
  readonly killedBy?: string;
}

export interface RoomNightWakeStep {
  readonly characterId: string;
  readonly characterType?: string;
  readonly order: number;
  readonly actionType: string;
  readonly prompt: string;
  readonly minTargets: number;
  readonly maxTargets: number;
}

export interface RoomNightAction {
  readonly actorId: string;
  readonly actionType: string;
  readonly targetIds: readonly string[];
  readonly result?: string;
}

export interface GameEndedPayload {
  readonly winner: number | string;
  readonly reason: string;
  readonly description: string;
}

export interface IdentityStatus {
  readonly status: IdentityState;
  readonly canRejoin: boolean;
  readonly participantSetFrozen: boolean;
  readonly nextClientSequence?: number;
}

export interface RoomState {
  readonly roomId: string;
  readonly players: readonly RoomPlayer[];
  readonly maxPlayers: number;
  readonly scriptId: string;
  readonly scriptName: string;
  readonly creatorId?: string;
  readonly storytellerId?: string;
  readonly phase: number;
  readonly dayNumber: number;
  readonly nomination?: RoomNomination;
  readonly deaths?: readonly RoomDeathRecord[];
  readonly ghostVotesRemaining?: readonly string[];
  readonly nightWakeSteps?: readonly RoomNightWakeStep[];
  readonly currentNightWakeIndex?: number;
  readonly currentNightWakeStep?: RoomNightWakeStep;
  readonly winner?: GameEndedPayload;
  readonly nightActions?: readonly RoomNightAction[];
  readonly storytellerName?: string;
  readonly nightNumber: number;
  readonly fortuneTellerRedHerringId?: string;
}

export interface ClientMessage {
  readonly protocolVersion: 2;
  readonly type: ClientMessageType;
  readonly requestId?: string;
  readonly joinRequestId?: string;
  readonly resumeCredential?: string;
  readonly clientSequence?: number;
  readonly roomId?: string;
  readonly playerName?: string;
  readonly playerId?: string;
  readonly targetPlayerId?: string;
  readonly maxPlayers?: number;
  readonly scriptId?: string;
  readonly assignments?: Readonly<Record<string, string>>;
  readonly shownCharacters?: Readonly<Record<string, string>>;
  readonly fortuneTellerRedHerringId?: string;
  readonly event?: Readonly<Record<string, unknown>>;
  readonly nomineeId?: string;
  readonly decision?: boolean;
  readonly phase?: string | number;
  readonly winner?: 'good' | 'evil' | 1 | 2;
  readonly reason?: string;
  readonly description?: string;
  readonly cause?: string | number;
  readonly actionType?: string;
  readonly targetIds?: readonly string[];
  readonly result?: string;
}

export interface ServerMessage {
  readonly type: ServerMessageType;
  readonly code?: ProtocolErrorCode;
  readonly roomId?: string;
  readonly state?: RoomState;
  readonly identityStatus?: IdentityStatus;
  readonly error?: string;
  readonly resumeCredential?: string;
  readonly roomRevision?: number;
  readonly acceptedSequence?: number;
  readonly nextClientSequence?: number;
}
