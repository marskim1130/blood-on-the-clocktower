import type { PlayerId } from '../types/index.js';

// Messages matching Go server protocol
export interface ClientMessage {
  readonly type: 'JOIN_ROOM' | 'SUBMIT_EVENT';
  readonly roomId?: string;
  readonly playerName?: string;
  readonly playerId?: string;
  readonly event?: Record<string, unknown>;
}

export interface ServerMessage {
  readonly type: 'ROOM_STATE' | 'EVENT_BROADCAST';
  readonly roomId?: string;
  readonly state?: RoomState;
  readonly event?: GameServerEvent;
}

export interface RoomState {
  readonly roomId: string;
  readonly players: ReadonlyArray<{
    readonly id: string;
    readonly name: string;
    readonly isAlive: boolean;
  }>;
}

export type GameServerEvent =
  | { readonly playerJoined: { readonly player: { readonly id: string; readonly name: string; readonly isAlive: boolean } } }
  | { readonly playerLeft: { readonly playerId: string } }
  | { readonly phaseChanged: { readonly phase: number } }
  | { readonly voteCast: { readonly voterId: string; readonly targetId?: string } }
  | { readonly characterAssigned: { readonly playerId: string; readonly character: { readonly id: string; readonly name: string; readonly team: number; readonly ability: string } } };

type MessageHandler = (msg: ServerMessage) => void;
type StatusHandler = (status: ConnectionStatus) => void;
export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface WebSocketClientOptions {
  readonly url: string;
  readonly reconnectInterval?: number;
  readonly maxReconnectAttempts?: number;
}

export class GameWebSocketClient {
  private ws: WebSocket | null = null;
  private options: Required<WebSocketClientOptions>;
  private handlers: Set<MessageHandler> = new Set();
  private statusHandlers: Set<StatusHandler> = new Set();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private _status: ConnectionStatus = 'disconnected';

  constructor(options: WebSocketClientOptions) {
    this.options = {
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      ...options,
    };
  }

  get status(): ConnectionStatus {
    return this._status;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    this.setStatus('connecting');
    this.ws = new WebSocket(this.options.url);

    this.ws.onopen = () => {
      this.setStatus('connected');
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const msg: ServerMessage = JSON.parse(event.data as string);
        this.handlers.forEach((handler) => handler(msg));
      } catch {
        console.error('Failed to parse WebSocket message');
      }
    };

    this.ws.onclose = () => {
      this.setStatus('disconnected');
      this.tryReconnect();
    };

    this.ws.onerror = () => {
      this.setStatus('error');
    };
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.ws?.close();
    this.ws = null;
    this.setStatus('disconnected');
  }

  joinRoom(roomId: string, playerId: string, playerName: string): void {
    this.send({ type: 'JOIN_ROOM', roomId, playerId, playerName });
  }

  send(msg: ClientMessage): void {
    if (this.ws?.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket is not connected');
    }
    this.ws.send(JSON.stringify(msg));
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
    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) return;

    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, this.options.reconnectInterval);
  }
}
