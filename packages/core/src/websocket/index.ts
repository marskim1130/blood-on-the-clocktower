// Messages matching Go server protocol
export interface ClientMessage {
  readonly type:
    | 'CREATE_ROOM'
    | 'JOIN_ROOM'
    | 'LEAVE_ROOM'
    | 'KICK_PLAYER'
    | 'SET_STORYTELLER'
    | 'ASSIGN_CHARACTERS'
    | 'SUBMIT_EVENT'
    | 'START_GAME'
    | 'CHANGE_PHASE'
    | 'NOMINATE'
    | 'CAST_VOTE'
    | 'RESOLVE_NOMINATION'
    | 'EXECUTE_PLAYER'
    | 'SUBMIT_NIGHT_ACTION'
    | 'RESOLVE_NIGHT'
    | 'END_GAME';
  readonly roomId?: string;
  readonly playerName?: string;
  readonly playerId?: string;
  readonly targetPlayerId?: string;
  readonly maxPlayers?: number;
  readonly scriptId?: string;
  readonly assignments?: Record<string, string>;
  readonly event?: Record<string, unknown>;
  /** Phase name for CHANGE_PHASE (e.g. 'day', 'night', 'voting'). */
  readonly phase?: string;
  /** Nominee player id for NOMINATE. */
  readonly nomineeId?: string;
  /** Boolean vote decision for CAST_VOTE: true = guilty, false = innocent. */
  readonly decision?: boolean;
  /** Legacy alias for EXECUTE_PLAYER; prefer targetPlayerId. */
  readonly executePlayerId?: string;
  /** Night action type for SUBMIT_NIGHT_ACTION (e.g. 'kill', 'poison'). */
  readonly actionType?: string;
  /** Target player ids for SUBMIT_NIGHT_ACTION. */
  readonly targetIds?: readonly string[];
  /** Winning team for END_GAME. */
  readonly winner?: 'good' | 'evil';
  /** Optional machine-readable reason for END_GAME. */
  readonly reason?: string;
  /** Optional human-readable explanation for END_GAME. */
  readonly description?: string;
}

export interface ServerMessage {
  readonly type: 'ROOM_STATE' | 'EVENT_BROADCAST' | 'ERROR';
  readonly roomId?: string;
  readonly state?: RoomState;
  readonly event?: GameServerEvent;
  readonly error?: string;
}

export interface RoomState {
  readonly roomId: string;
  readonly players: ReadonlyArray<{
    readonly id: string;
    readonly name: string;
    readonly character?: GameCharacter | null;
    readonly isAlive: boolean;
    readonly votes?: number;
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
  | { readonly characterAssigned: { readonly playerId: string; readonly character: GameCharacter } }
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
}

interface ResolvedWebSocketClientOptions {
  readonly url: string;
  readonly reconnectInterval: number;
  readonly maxReconnectAttempts: number;
  readonly transportFactory: WebSocketTransportFactory;
}

interface RoomResumeSession {
  readonly roomId: string;
  readonly playerId: string;
  readonly playerName: string;
}

interface PendingCreatedRoomIdentity {
  readonly playerId: string;
  readonly playerName: string;
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
  private resumeSession: RoomResumeSession | null = null;
  private pendingCreatedRoomIdentity: PendingCreatedRoomIdentity | null = null;

  constructor(options: WebSocketClientOptions) {
    this.options = {
      url: options.url,
      reconnectInterval: options.reconnectInterval ?? 3000,
      maxReconnectAttempts: options.maxReconnectAttempts ?? 5,
      transportFactory: options.transportFactory ?? createBrowserWebSocketTransport,
    };
  }

  get status(): ConnectionStatus {
    return this._status;
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
        this.captureResumeSession(msg);
        this.handlers.forEach((handler) => handler(msg));
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

  createRoom(playerId: string, playerName: string, maxPlayers: number, scriptId?: string): void {
    this.pendingCreatedRoomIdentity = { playerId, playerName };
    this.send({
      type: 'CREATE_ROOM',
      playerId,
      playerName,
      maxPlayers,
      ...(scriptId ? { scriptId } : {}),
    });
  }

  joinRoom(roomId: string, playerId: string, playerName: string): void {
    this.resumeSession = { roomId, playerId, playerName };
    this.pendingCreatedRoomIdentity = null;
    this.send({ type: 'JOIN_ROOM', roomId, playerId, playerName });
  }

  leaveRoom(): void {
    this.resumeSession = null;
    this.pendingCreatedRoomIdentity = null;
    this.send({ type: 'LEAVE_ROOM' });
  }

  kickPlayer(targetPlayerId: string): void {
    this.send({ type: 'KICK_PLAYER', targetPlayerId });
  }

  setStoryteller(targetPlayerId: string): void {
    this.send({ type: 'SET_STORYTELLER', targetPlayerId });
  }

  assignCharacters(assignments: Record<string, string>): void {
    this.send({ type: 'ASSIGN_CHARACTERS', assignments });
  }

  submitEvent(event: Record<string, unknown>): void {
    this.send({ type: 'SUBMIT_EVENT', event });
  }

  startGame(): void {
    this.send({ type: 'START_GAME' });
  }

  changePhase(phase: string): void {
    this.send({ type: 'CHANGE_PHASE', phase });
  }

  nominate(nomineeId: string): void {
    this.send({ type: 'NOMINATE', nomineeId });
  }

  castVote(decision: boolean): void {
    this.send({ type: 'CAST_VOTE', decision });
  }

  resolveNomination(): void {
    this.send({ type: 'RESOLVE_NOMINATION' });
  }

  executePlayer(playerId: string): void {
    this.send({ type: 'EXECUTE_PLAYER', targetPlayerId: playerId });
  }

  submitNightAction(actionType: string, targetIds: readonly string[]): void {
    this.send({ type: 'SUBMIT_NIGHT_ACTION', actionType, targetIds });
  }

  resolveNight(): void {
    this.send({ type: 'RESOLVE_NIGHT' });
  }

  endGame(winner: 'good' | 'evil', description?: string): void {
    this.send({
      type: 'END_GAME',
      winner,
      reason: 'storyteller_decision',
      ...(description?.trim() ? { description: description.trim() } : {}),
    });
  }

  send(msg: ClientMessage): void {
    if (this.transport?.readyState !== READY_STATE_OPEN) {
      throw new Error('WebSocket is not connected');
    }
    this.transport.send(JSON.stringify(msg));
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

  private captureResumeSession(msg: ServerMessage): void {
    if (msg.type === 'ERROR' && msg.error === 'kicked from room') {
      this.resumeSession = null;
      this.pendingCreatedRoomIdentity = null;
      return;
    }

    if (msg.type !== 'ROOM_STATE' || !msg.roomId || !this.pendingCreatedRoomIdentity) return;

    this.resumeSession = {
      roomId: msg.roomId,
      playerId: this.pendingCreatedRoomIdentity.playerId,
      playerName: this.pendingCreatedRoomIdentity.playerName,
    };
    this.pendingCreatedRoomIdentity = null;
  }

  private resumeRoomIfNeeded(): void {
    if (!this.resumeSession) return;

    try {
      this.send({
        type: 'JOIN_ROOM',
        roomId: this.resumeSession.roomId,
        playerId: this.resumeSession.playerId,
        playerName: this.resumeSession.playerName,
      });
    } catch {
      this.setStatus('error');
    }
  }
}
