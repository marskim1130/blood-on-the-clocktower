import Taro from '@tarojs/taro';
import type { WebSocketTransport } from '@clocktower/core';

const READY_STATE_CONNECTING = 0;
const READY_STATE_OPEN = 1;
const READY_STATE_CLOSED = 3;

interface TaroSocketMessage {
  readonly data: string | ArrayBuffer;
}

interface TaroSocketTask {
  readonly readyState?: number;
  readonly OPEN?: number;
  send(option: { readonly data: string }): Promise<unknown> | void;
  close(option?: { readonly code?: number; readonly reason?: string }): Promise<unknown> | void;
  onOpen(handler: () => void): void;
  onMessage(handler: (message: TaroSocketMessage) => void): void;
  onClose(handler: () => void): void;
  onError(handler: () => void): void;
}

interface TaroConnectSocketOption {
  readonly url: string;
}

type TaroConnectSocket = (option: TaroConnectSocketOption) => Promise<TaroSocketTask>;

export class TaroWebSocketTransport implements WebSocketTransport {
  private task: TaroSocketTask | null = null;
  private state = READY_STATE_CLOSED;
  private openHandler: (() => void) | null = null;
  private messageHandler: ((data: string) => void) | null = null;
  private closeHandler: (() => void) | null = null;
  private errorHandler: (() => void) | null = null;
  private connectGeneration = 0;

  constructor(private readonly url: string) {}

  get readyState(): number {
    return this.state;
  }

  connect(): void {
    if (this.state === READY_STATE_CONNECTING || this.state === READY_STATE_OPEN) {
      return;
    }

    this.state = READY_STATE_CONNECTING;
    const generation = ++this.connectGeneration;
    const connectSocket = Taro.connectSocket as unknown as TaroConnectSocket;
    connectSocket({ url: this.url })
      .then((task) => {
        if (generation !== this.connectGeneration || this.state !== READY_STATE_CONNECTING) {
          void task.close({ code: 1000, reason: 'client disconnect' });
          return;
        }

        this.task = task;
        task.onOpen(() => {
          this.markOpen();
        });
        task.onMessage((message) => {
          const data = typeof message.data === 'string' ? message.data : '';
          this.messageHandler?.(data);
        });
        task.onClose(() => {
          this.state = READY_STATE_CLOSED;
          this.task = null;
          this.closeHandler?.();
        });
        task.onError(() => {
          this.state = READY_STATE_CLOSED;
          this.errorHandler?.();
        });
        if (this.isTaskOpen(task)) {
          this.markOpen();
        }
      })
      .catch(() => {
        this.state = READY_STATE_CLOSED;
        this.errorHandler?.();
      });
  }

  send(data: string): void {
    if (this.state !== READY_STATE_OPEN || !this.task) {
      throw new Error('WebSocket is not connected');
    }
    void this.task.send({ data });
  }

  close(): void {
    this.state = READY_STATE_CLOSED;
    if (this.task) {
      void this.task.close({ code: 1000, reason: 'client disconnect' });
      this.task = null;
    }
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

  private markOpen(): void {
    if (this.state === READY_STATE_OPEN) return;
    this.state = READY_STATE_OPEN;
    this.openHandler?.();
  }

  private isTaskOpen(task: TaroSocketTask): boolean {
    const openState = task.OPEN ?? READY_STATE_OPEN;
    return task.readyState === openState;
  }
}

export function createTaroWebSocketTransport(url: string): WebSocketTransport {
  return new TaroWebSocketTransport(url);
}
