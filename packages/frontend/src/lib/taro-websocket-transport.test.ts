import { beforeEach, describe, expect, it, vi } from 'vitest';

const taroMock = vi.hoisted(() => ({
  connectSocket: vi.fn(),
}));

vi.mock('@tarojs/taro', () => ({
  default: {
    connectSocket: taroMock.connectSocket,
  },
}));

import { TaroWebSocketTransport } from './taro-websocket-transport';

class MockSocketTask {
  private openHandler: (() => void) | null = null;
  readonly OPEN = 1;
  readyState = 0;

  readonly send = vi.fn();
  readonly close = vi.fn();

  onOpen(handler: () => void): void {
    this.openHandler = handler;
  }

  onMessage(): void {}

  onClose(): void {}

  onError(): void {}

  open(): void {
    this.readyState = this.OPEN;
    this.openHandler?.();
  }
}

describe('TaroWebSocketTransport', () => {
  beforeEach(() => {
    taroMock.connectSocket.mockReset();
  });

  it('cancels a pending socket task when closed before connect resolves', async () => {
    let resolveTask: ((task: MockSocketTask) => void) | null = null;
    taroMock.connectSocket.mockImplementation(
      () =>
        new Promise<MockSocketTask>((resolve) => {
          resolveTask = resolve;
        })
    );

    const transport = new TaroWebSocketTransport('ws://example.test/ws');
    const opened = vi.fn();
    transport.onOpen(opened);

    transport.connect();
    transport.close();

    if (!resolveTask) {
      throw new Error('expected connectSocket to be called');
    }

    const task = new MockSocketTask();
    (resolveTask as (task: MockSocketTask) => void)(task);
    await Promise.resolve();

    expect(task.close).toHaveBeenCalledTimes(1);

    task.open();
    expect(opened).not.toHaveBeenCalled();
    expect(transport.readyState).toBe(3);
  });

  it('ignores an older pending socket when reconnect starts before it resolves', async () => {
    const resolvers: Array<(task: MockSocketTask) => void> = [];
    taroMock.connectSocket.mockImplementation(
      () =>
        new Promise<MockSocketTask>((resolve) => {
          resolvers.push(resolve);
        })
    );

    const transport = new TaroWebSocketTransport('ws://example.test/ws');
    const opened = vi.fn();
    transport.onOpen(opened);

    transport.connect();
    transport.close();
    transport.connect();

    const resolveOldTask = resolvers[0];
    const resolveCurrentTask = resolvers[1];
    if (!resolveOldTask || !resolveCurrentTask) {
      throw new Error('expected two connectSocket calls');
    }

    const oldTask = new MockSocketTask();
    resolveOldTask(oldTask);
    await Promise.resolve();

    expect(oldTask.close).toHaveBeenCalledTimes(1);

    oldTask.open();
    expect(opened).not.toHaveBeenCalled();

    const currentTask = new MockSocketTask();
    resolveCurrentTask(currentTask);
    await Promise.resolve();

    currentTask.open();
    expect(opened).toHaveBeenCalledTimes(1);
  });

  it('marks the transport open when the socket task is already open before handlers are registered', async () => {
    const task = new MockSocketTask();
    task.open();
    taroMock.connectSocket.mockResolvedValue(task);

    const transport = new TaroWebSocketTransport('ws://example.test/ws');
    const opened = vi.fn();
    transport.onOpen(opened);

    transport.connect();
    await Promise.resolve();

    expect(opened).toHaveBeenCalledTimes(1);
    expect(transport.readyState).toBe(1);
  });
});
