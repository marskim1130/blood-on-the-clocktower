// Messages matching Go server protocol
export interface ClientMessage {
  readonly type:
    | 'CREATE_ROOM'
    | 'JOIN_ROOM'
    | 'LEAVE_ROOM'
    | 'SET_STORYTELLER'
    | 'ASSIGN_CHARACTERS'
    | 'SUBMIT_EVENT';
  readonly roomId?: string;
  readonly playerName?: string;
  readonly playerId?: string;
  readonly targetPlayerId?: string;
  readonly maxPlayers?: number;
  readonly assignments?: Record<string, string>;
  readonly event?: Record<string, unknown>;
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
  readonly storytellerId?: string;
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
  | { readonly voteCast: { readonly voterId: string; readonly targetId?: string } }
  | { readonly characterAssigned: { readonly playerId: string; readonly character: GameCharacter } };

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

export class GameWebSocketClient {
  private transport: WebSocketTransport | null = null;
  private readonly options: ResolvedWebSocketClientOptions;
  private readonly handlers: Set<MessageHandler> = new Set();
  private readonly statusHandlers: Set<StatusHandler> = new Set();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private intentionalDisconnect = false;
  private _status: ConnectionStatus = 'disconnected';

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
    });

    transport.onMessage((data) => {
      try {
        const msg: ServerMessage = JSON.parse(data);
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

  createRoom(playerId: string, playerName: string, maxPlayers: number): void {
    this.send({ type: 'CREATE_ROOM', playerId, playerName, maxPlayers });
  }

  joinRoom(roomId: string, playerId: string, playerName: string): void {
    this.send({ type: 'JOIN_ROOM', roomId, playerId, playerName });
  }

  leaveRoom(): void {
    this.send({ type: 'LEAVE_ROOM' });
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
}
