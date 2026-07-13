// Messages matching Go server protocol
export interface ClientMessage {
  readonly protocolVersion: 2;
  readonly type:
    | 'CREATE_ROOM'
    | 'JOIN_ROOM'
    | 'RESUME_ROOM'
    | 'REJOIN_ROOM'
    | 'GET_ROOM_STATE'
    | 'CLOSE_ROOM'
    | 'LEAVE_ROOM'
    | 'KICK_PLAYER'
    | 'UPDATE_ROOM_SETTINGS'
    | 'SET_STORYTELLER'
    | 'ASSIGN_CHARACTERS'
    | 'SUBMIT_EVENT'
    | 'START_GAME'
    | 'CHANGE_PHASE'
    | 'NOMINATE'
    | 'CAST_VOTE'
    | 'RESOLVE_NOMINATION'
    | 'EXECUTE_PLAYER'
    | 'USE_SLAYER_ABILITY'
    | 'KILL_PLAYER'
    | 'SUBMIT_NIGHT_ACTION'
    | 'RESOLVE_NIGHT'
    | 'END_GAME';
  readonly roomId?: string;
  readonly requestId?: string;
  readonly joinRequestId?: string;
  readonly resumeCredential?: string;
  readonly clientSequence?: number;
  readonly playerName?: string;
  readonly playerId?: string;
  readonly targetPlayerId?: string;
  readonly maxPlayers?: number;
  readonly scriptId?: string;
  readonly assignments?: Record<string, string>;
  readonly shownCharacters?: Record<string, string>;
  readonly fortuneTellerRedHerringId?: string;
  readonly event?: Record<string, unknown>;
  /** Phase name for CHANGE_PHASE (e.g. 'day', 'night', 'voting'). */
  readonly phase?: string;
  /** Nominee player id for NOMINATE. */
  readonly nomineeId?: string;
  /** Boolean vote decision for CAST_VOTE: true = guilty, false = innocent. */
  readonly decision?: boolean;
  /** Legacy alias for EXECUTE_PLAYER; prefer targetPlayerId. */
  readonly executePlayerId?: string;
  /** Death cause for KILL_PLAYER (e.g. 'execution', 'night_kill', 'ability'). */
  readonly cause?: string;
  /** Night action type for SUBMIT_NIGHT_ACTION (e.g. 'kill', 'poison'). */
  readonly actionType?: string;
  /** Target player ids for SUBMIT_NIGHT_ACTION. */
  readonly targetIds?: readonly string[];
  /** Optional Storyteller-adjudicated result for SUBMIT_NIGHT_ACTION. */
  readonly result?: string;
  /** Winning team for END_GAME. */
  readonly winner?: 'good' | 'evil';
  /** Optional machine-readable reason for END_GAME. */
  readonly reason?: string;
  /** Optional human-readable explanation for END_GAME. */
  readonly description?: string;
}

export interface ServerMessage {
  readonly type:
    | 'CREATE_ROOM_RESULT'
    | 'JOIN_ROOM_RESULT'
    | 'RESUME_ROOM_RESULT'
    | 'COMMAND_RESULT'
    | 'ROOM_STATE'
    | 'ROOM_STATE_CHANGED'
    | 'IDENTITY_STATUS'
    | 'KICKED'
    | 'ROOM_CLOSED'
    | 'EVENT_BROADCAST'
    | 'ERROR';
  readonly roomId?: string;
  readonly state?: RoomState;
  readonly event?: GameServerEvent;
  readonly error?: string;
  readonly code?: string;
  readonly resumeCredential?: string;
  readonly acceptedSequence?: number;
  readonly nextClientSequence?: number;
  readonly roomRevision?: number;
  readonly identityStatus?: IdentityStatus;
}

export interface IdentityStatus {
  readonly status: 'member' | 'retained';
  readonly canRejoin: boolean;
  readonly participantSetFrozen: boolean;
}

export interface RoomState {
  readonly roomId: string;
  readonly players: ReadonlyArray<{
    readonly id: string;
    readonly name: string;
    readonly character?: GameCharacter | null;
    readonly shownCharacter?: GameCharacter | null;
    readonly isAlive: boolean;
    readonly votes?: number;
    readonly poisonedUntil?: number;
  }>;
  readonly maxPlayers?: number;
  readonly scriptId?: string;
  readonly scriptName?: string;
  readonly creatorId?: string;
  readonly storytellerId?: string;
  readonly phase?: number;
  readonly dayNumber?: number;
  readonly nomination?: {
    readonly nominatorId: string;
    readonly nomineeId: string;
    readonly votes?: Record<string, boolean>;
  } | null;
  readonly deaths?: ReadonlyArray<{
    readonly playerId: string;
    readonly cause: string;
    readonly dayNumber: number;
    readonly killedBy?: string;
  }>;
  readonly ghostVotesRemaining?: readonly string[];
  readonly nightWakeSteps?: readonly RoomNightWakeStep[];
  readonly currentNightWakeIndex?: number;
  readonly currentNightWakeStep?: RoomNightWakeStep | null;
  readonly winner?: GameEndedPayload | null;
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

export interface GameCharacter {
  readonly id: string;
  readonly name: string;
  readonly team: number;
  readonly ability: string;
}

export type GameServerEvent =
  | { readonly playerJoined: { readonly player: { readonly id: string; readonly name: string; readonly isAlive: boolean } } }
  | { readonly playerLeft: { readonly playerId: string } }
  | { readonly phaseChanged: { readonly phase: number } }
  | { readonly voteCast: { readonly voterId: string; readonly targetId?: string; readonly decision?: boolean } }
  | { readonly characterAssigned: { readonly playerId: string; readonly character: GameCharacter; readonly shownCharacter?: GameCharacter | null } }
  | { readonly playerDied: { readonly playerId: string; readonly cause: string; readonly dayNumber: number } }
  | { readonly nominationStarted: { readonly nominatorId: string; readonly nomineeId: string } }
  | { readonly nominationResolved: { readonly nomineeId: string; readonly executed: boolean; readonly yesVotes: number; readonly noVotes: number; readonly requiredVotes?: number } }
  | { readonly nightAction: { readonly actorId: string; readonly actionType: string; readonly targetIds: readonly string[]; readonly result: string | null } }
  | { readonly nightActionSubmitted: { readonly actorId: string; readonly actionType: string; readonly targetIds: readonly string[]; readonly result?: string | null } }
  | { readonly gameOver: GameEndedPayload }
  | { readonly gameEnded: GameEndedPayload };

export interface GameEndedPayload {
  readonly winner: number | string;
  readonly reason: string;
  readonly description: string;
}

type MessageHandler = (msg: ServerMessage) => void;
type StatusHandler = (status: ConnectionStatus) => void;
export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

const READY_STATE_CONNECTING = 0;
const READY_STATE_OPEN = 1;
const READY_STATE_CLOSED = 3;

export interface WebSocketTransport {
  readonly readyState: number;
  connect(): void;
  send(data: string): void;
  close(): void;
  onOpen(handler: () => void): void;
  onMessage(handler: (data: string) => void): void;
  onClose(handler: () => void): void;
  onError(handler: () => void): void;
}

export type WebSocketTransportFactory = (url: string) => WebSocketTransport;

interface BrowserLikeWebSocket {
  readonly readyState: number;
  send(data: string): void;
  close(): void;
  onopen: ((event: unknown) => void) | null;
  onmessage: ((event: { readonly data: unknown }) => void) | null;
  onclose: ((event: unknown) => void) | null;
  onerror: ((event: unknown) => void) | null;
}

type BrowserWebSocketConstructor = new (url: string) => BrowserLikeWebSocket;

class BrowserWebSocketTransport implements WebSocketTransport {
  private socket: BrowserLikeWebSocket | null = null;
  private openHandler: (() => void) | null = null;
  private messageHandler: ((data: string) => void) | null = null;
  private closeHandler: (() => void) | null = null;
  private errorHandler: (() => void) | null = null;

  constructor(private readonly url: string) {}

  get readyState(): number {
    return this.socket?.readyState ?? READY_STATE_CLOSED;
  }

  connect(): void {
    if (this.socket?.readyState === READY_STATE_CONNECTING || this.socket?.readyState === READY_STATE_OPEN) {
      return;
    }

    const WebSocketCtor = (globalThis as { readonly WebSocket?: BrowserWebSocketConstructor }).WebSocket;
    if (!WebSocketCtor) {
      throw new Error('WebSocket is not available in this environment');
    }

    this.socket = new WebSocketCtor(this.url);
    this.socket.onopen = () => this.openHandler?.();
    this.socket.onmessage = (event) => this.messageHandler?.(String(event.data));
    this.socket.onclose = () => this.closeHandler?.();
    this.socket.onerror = () => this.errorHandler?.();
  }

  send(data: string): void {
    if (this.socket?.readyState !== READY_STATE_OPEN) {
      throw new Error('WebSocket is not connected');
    }
    this.socket.send(data);
  }

  close(): void {
    this.socket?.close();
    this.socket = null;
  }

  onOpen(handler: () => void): void {
    this.openHandler = handler;
  }

  onMessage(handler: (data: string) => void): void {
    this.messageHandler = handler;
  }

  onClose(handler: () => void): void {
    this.closeHandler = handler;
  }

  onError(handler: () => void): void {
    this.errorHandler = handler;
  }
}

export function createBrowserWebSocketTransport(url: string): WebSocketTransport {
  return new BrowserWebSocketTransport(url);
}

export interface WebSocketClientOptions {
  readonly url: string;
  readonly reconnectInterval?: number;
  readonly maxReconnectAttempts?: number;
  readonly transportFactory?: WebSocketTransportFactory;
  readonly requestIdFactory?: () => string;
}

interface ResolvedWebSocketClientOptions {
  readonly url: string;
  readonly reconnectInterval: number;
  readonly maxReconnectAttempts: number;
  readonly transportFactory: WebSocketTransportFactory;
  readonly requestIdFactory: () => string;
}

export interface ClientRoomIdentity {
  readonly roomId: string;
  readonly playerId: string;
  readonly resumeCredential: string;
}

interface PendingIdentity {
  readonly playerId: string;
  readonly requestKind: 'create' | 'join';
}

type OutboundClientMessage = Omit<ClientMessage, 'protocolVersion'> & { readonly protocolVersion?: 2 };
type SequencedClientMessage = Omit<OutboundClientMessage, 'clientSequence'>;

function createRequestId(): string {
  const cryptoApi = globalThis.crypto;
  if (cryptoApi?.randomUUID) return cryptoApi.randomUUID();
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

export class GameWebSocketClient {
  private transport: WebSocketTransport | null = null;
  private readonly options: ResolvedWebSocketClientOptions;
  private readonly handlers: Set<MessageHandler> = new Set();
  private readonly statusHandlers: Set<StatusHandler> = new Set();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private intentionalDisconnect = false;
  private _status: ConnectionStatus = 'disconnected';
  private resumeSession: ClientRoomIdentity | null = null;
  private pendingIdentity: PendingIdentity | null = null;
  private nextClientSequence: number | null = null;
  private roomRevision: number | null = null;
  private resyncInFlight = false;
  private pendingSequencedCommand: { readonly sequence: number; readonly message: ClientMessage } | null = null;
  private queuedSequencedCommands: SequencedClientMessage[] = [];

  constructor(options: WebSocketClientOptions) {
    this.options = {
      url: options.url,
      reconnectInterval: options.reconnectInterval ?? 3000,
      maxReconnectAttempts: options.maxReconnectAttempts ?? 5,
      transportFactory: options.transportFactory ?? createBrowserWebSocketTransport,
      requestIdFactory: options.requestIdFactory ?? createRequestId,
    };
  }

  get status(): ConnectionStatus {
    return this._status;
  }

  get identity(): ClientRoomIdentity | null {
    return this.resumeSession;
  }

  get nextSequence(): number | null {
    return this.nextClientSequence;
  }

  get currentRoomRevision(): number | null {
    return this.roomRevision;
  }

  connect(): void {
    if (
      this.transport?.readyState === READY_STATE_OPEN ||
      this.transport?.readyState === READY_STATE_CONNECTING
    ) {
      return;
    }

    this.intentionalDisconnect = false;
    this.setStatus('connecting');

    const transport = this.options.transportFactory(this.options.url);
    this.transport = transport;

    transport.onOpen(() => {
      this.setStatus('connected');
      this.reconnectAttempts = 0;
      this.resumeRoomIfNeeded();
    });

    transport.onMessage((data) => {
      try {
        const msg: ServerMessage = JSON.parse(data);
        const acceptedMessage = this.acceptProjection(msg);
        this.captureProtocolState(acceptedMessage);
        this.handlers.forEach((handler) => handler(acceptedMessage));
      } catch {
        console.warn('Failed to parse WebSocket message');
      }
    });

    transport.onClose(() => {
      this.transport = null;
      this.setStatus('disconnected');
      if (!this.intentionalDisconnect) {
        this.tryReconnect();
      }
    });

    transport.onError(() => {
      this.setStatus('error');
    });

    try {
      transport.connect();
    } catch {
      this.transport = null;
      this.setStatus('error');
      this.tryReconnect();
    }
  }

  disconnect(): void {
    this.intentionalDisconnect = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.transport?.close();
    this.transport = null;
    this.setStatus('disconnected');
  }

  createRoom(playerId: string, playerName: string, maxPlayers: number, scriptId?: string, requestId = this.options.requestIdFactory()): void {
    this.pendingIdentity = { playerId, requestKind: 'create' };
    this.send({
      type: 'CREATE_ROOM',
      requestId,
      playerId,
      playerName,
      maxPlayers,
      ...(scriptId ? { scriptId } : {}),
    });
  }

  joinRoom(roomId: string, playerId: string, playerName: string, joinRequestId = this.options.requestIdFactory()): void {
    this.pendingIdentity = { playerId, requestKind: 'join' };
    this.send({ type: 'JOIN_ROOM', joinRequestId, roomId, playerId, playerName });
  }

  resumeRoom(roomId: string, playerId: string, resumeCredential: string): void {
    this.resumeSession = { roomId, playerId, resumeCredential };
    this.send({ type: 'RESUME_ROOM', roomId, playerId, resumeCredential });
  }

  rejoinRoom(): void {
    this.sendSequenced({ type: 'REJOIN_ROOM' });
  }

  getRoomState(): void {
    this.sendAuthenticated({ type: 'GET_ROOM_STATE' });
  }

  closeRoom(): void {
    this.sendSequenced({ type: 'CLOSE_ROOM' });
  }

  leaveRoom(): void {
    this.sendSequenced({ type: 'LEAVE_ROOM' });
  }

  kickPlayer(targetPlayerId: string): void {
    this.sendSequenced({ type: 'KICK_PLAYER', targetPlayerId });
  }

  updateRoomSettings(maxPlayers?: number, scriptId?: string): void {
    this.sendSequenced({
      type: 'UPDATE_ROOM_SETTINGS',
      ...(typeof maxPlayers === 'number' ? { maxPlayers } : {}),
      ...(scriptId ? { scriptId } : {}),
    });
  }

  setStoryteller(targetPlayerId: string): void {
    this.sendSequenced({ type: 'SET_STORYTELLER', targetPlayerId });
  }

  assignCharacters(
    assignments: Record<string, string>,
    shownCharacters?: Record<string, string>,
    fortuneTellerRedHerringId?: string,
  ): void {
    this.sendSequenced({
      type: 'ASSIGN_CHARACTERS',
      assignments,
      ...(shownCharacters && Object.keys(shownCharacters).length > 0 ? { shownCharacters } : {}),
      ...(fortuneTellerRedHerringId ? { fortuneTellerRedHerringId } : {}),
    });
  }

  submitEvent(event: Record<string, unknown>): void {
    this.sendSequenced({ type: 'SUBMIT_EVENT', event });
  }

  startGame(): void {
    this.sendSequenced({ type: 'START_GAME' });
  }

  changePhase(phase: string): void {
    this.sendSequenced({ type: 'CHANGE_PHASE', phase });
  }

  nominate(nomineeId: string): void {
    this.sendSequenced({ type: 'NOMINATE', nomineeId });
  }

  castVote(decision: boolean): void {
    this.sendSequenced({ type: 'CAST_VOTE', decision });
  }

  resolveNomination(): void {
    this.sendSequenced({ type: 'RESOLVE_NOMINATION' });
  }

  executePlayer(playerId: string): void {
    this.sendSequenced({ type: 'EXECUTE_PLAYER', targetPlayerId: playerId });
  }

  useSlayerAbility(targetPlayerId: string): void {
    this.sendSequenced({ type: 'USE_SLAYER_ABILITY', targetPlayerId });
  }

  killPlayer(targetPlayerId: string, cause: string): void {
    this.sendSequenced({ type: 'KILL_PLAYER', targetPlayerId, cause });
  }

  submitNightAction(actionType: string, targetIds: readonly string[], result?: string): void {
    this.sendSequenced({
      type: 'SUBMIT_NIGHT_ACTION',
      actionType,
      targetIds,
      ...(result?.trim() ? { result: result.trim() } : {}),
    });
  }

  resolveNight(): void {
    this.sendSequenced({ type: 'RESOLVE_NIGHT' });
  }

  endGame(winner: 'good' | 'evil', description?: string): void {
    this.sendSequenced({
      type: 'END_GAME',
      winner,
      reason: 'storyteller_decision',
      ...(description?.trim() ? { description: description.trim() } : {}),
    });
  }

  send(msg: OutboundClientMessage): void {
    if (this.transport?.readyState !== READY_STATE_OPEN) {
      throw new Error('WebSocket is not connected');
    }
    this.transport.send(JSON.stringify({ ...msg, protocolVersion: 2 } satisfies ClientMessage));
  }

  onMessage(handler: MessageHandler): () => void {
    this.handlers.add(handler);
    return () => this.handlers.delete(handler);
  }

  onStatusChange(handler: StatusHandler): () => void {
    this.statusHandlers.add(handler);
    return () => this.statusHandlers.delete(handler);
  }

  private setStatus(status: ConnectionStatus): void {
    this._status = status;
    this.statusHandlers.forEach((handler) => handler(status));
  }

  private tryReconnect(): void {
    if (this.intentionalDisconnect || this.reconnectAttempts >= this.options.maxReconnectAttempts) return;

    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, this.options.reconnectInterval);
  }

  private captureProtocolState(msg: ServerMessage): void {
    if (typeof msg.roomRevision === 'number') {
      this.roomRevision = msg.roomRevision;
    }

    if (msg.type === 'KICKED' || msg.type === 'ROOM_CLOSED' || (msg.type === 'ERROR' && msg.code === 'INVALID_CREDENTIAL')) {
      this.resumeSession = null;
      this.pendingIdentity = null;
      this.nextClientSequence = null;
      this.pendingSequencedCommand = null;
      this.queuedSequencedCommands = [];
      return;
    }

    if (
      (msg.type === 'CREATE_ROOM_RESULT' || msg.type === 'JOIN_ROOM_RESULT') &&
      msg.roomId &&
      msg.resumeCredential &&
      this.pendingIdentity
    ) {
      this.resumeSession = {
        roomId: msg.roomId,
        playerId: this.pendingIdentity.playerId,
        resumeCredential: msg.resumeCredential,
      };
      this.pendingIdentity = null;
    }

    if (
      typeof msg.nextClientSequence === 'number' &&
      (msg.type !== 'ERROR' || msg.code === 'UNEXPECTED_SEQUENCE' || msg.code === 'SEQUENCE_CONFLICT')
    ) {
      this.nextClientSequence = msg.nextClientSequence;
    }

    if (
      msg.type === 'COMMAND_RESULT' &&
      this.pendingSequencedCommand &&
      msg.acceptedSequence === this.pendingSequencedCommand.sequence &&
      typeof msg.nextClientSequence === 'number'
    ) {
      this.pendingSequencedCommand = null;
      this.flushSequencedCommand();
    }

    if (
      msg.type === 'RESUME_ROOM_RESULT' &&
      this.pendingSequencedCommand &&
      typeof msg.nextClientSequence === 'number'
    ) {
      if (msg.nextClientSequence === this.pendingSequencedCommand.sequence) {
        this.send(this.pendingSequencedCommand.message);
      } else {
        this.pendingSequencedCommand = null;
        this.flushSequencedCommand();
      }
    }

    if (msg.type === 'ERROR' && this.pendingSequencedCommand) {
      this.pendingSequencedCommand = null;
      this.flushSequencedCommand();
    }
  }

  private acceptProjection(msg: ServerMessage): ServerMessage {
    if (!msg.state || typeof msg.roomRevision !== 'number') return msg;

    const currentRevision = this.roomRevision;
    const isFullProjection =
      msg.type === 'CREATE_ROOM_RESULT' ||
      msg.type === 'JOIN_ROOM_RESULT' ||
      msg.type === 'RESUME_ROOM_RESULT' ||
      msg.type === 'ROOM_STATE';

    if (currentRevision !== null && msg.roomRevision <= currentRevision) {
      const { state: _state, roomRevision: _roomRevision, ...messageWithoutProjection } = msg;
      return messageWithoutProjection;
    }

    if (!isFullProjection && currentRevision !== null && msg.roomRevision > currentRevision + 1) {
      if (!this.resyncInFlight) {
        this.resyncInFlight = true;
        try {
          this.getRoomState();
        } catch {
          this.resyncInFlight = false;
        }
      }
      const { state: _state, roomRevision: _roomRevision, ...messageWithoutProjection } = msg;
      return messageWithoutProjection;
    }

    if (isFullProjection) this.resyncInFlight = false;
    return msg;
  }

  private sendAuthenticated(msg: OutboundClientMessage): void {
    const identity = this.requireIdentity();
    this.send({
      ...msg,
      roomId: identity.roomId,
      playerId: identity.playerId,
      resumeCredential: identity.resumeCredential,
    });
  }

  private sendSequenced(msg: SequencedClientMessage): void {
    this.requireIdentity();
    if (this.nextClientSequence === null) {
      throw new Error('Next client sequence is not available');
    }
    this.queuedSequencedCommands.push(msg);
    this.flushSequencedCommand();
  }

  private flushSequencedCommand(): void {
    if (this.pendingSequencedCommand || this.queuedSequencedCommands.length === 0) return;
    const sequence = this.nextClientSequence;
    if (sequence === null) return;

    const queued = this.queuedSequencedCommands.shift()!;
    const identity = this.requireIdentity();
    const message: ClientMessage = {
      ...queued,
      protocolVersion: 2,
      roomId: identity.roomId,
      playerId: identity.playerId,
      resumeCredential: identity.resumeCredential,
      clientSequence: sequence,
    };
    this.pendingSequencedCommand = { sequence, message };
    this.send(message);
  }

  private requireIdentity(): ClientRoomIdentity {
    if (!this.resumeSession) {
      throw new Error('Room identity is not available');
    }
    return this.resumeSession;
  }

  private resumeRoomIfNeeded(): void {
    if (!this.resumeSession) return;

    try {
      this.send({
        type: 'RESUME_ROOM',
        roomId: this.resumeSession.roomId,
        playerId: this.resumeSession.playerId,
        resumeCredential: this.resumeSession.resumeCredential,
      });
    } catch {
      this.setStatus('error');
    }
  }
}
