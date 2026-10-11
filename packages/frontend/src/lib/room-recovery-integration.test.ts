import { afterEach, expect, it, vi } from 'vitest';

vi.mock('@tarojs/taro', () => ({
  default: { connectSocket: vi.fn(), getStorageSync: vi.fn(), setStorageSync: vi.fn(), removeStorageSync: vi.fn() },
}));

import { GameWebSocketClient, type ClientMessage, type WebSocketTransport } from '@clocktower/core';
import { createRoomSessionStore, STORAGE_KEYS, type RoomSessionDependencies } from './room-session-store';
import type { StoredRoomIdentity } from './utils';

class RecoveryTransport implements WebSocketTransport {
  readyState = 3;
  readonly sent: ClientMessage[] = [];
  failSend = false;
  onSend: ((message: ClientMessage) => void) | undefined;
  sendCompletion: (() => Promise<void>) | undefined;
  private opened = (): void => {};
  private received = (_data: string): void => {};
  private closed = (): void => {};

  connect(): void { this.readyState = 1; this.opened(); }
  close(): void { this.readyState = 3; this.closed(); }
  send(data: string): void | Promise<void> {
    this.sent.push(JSON.parse(data) as ClientMessage);
    this.onSend?.(this.sent.at(-1)!);
    if (this.failSend) throw new Error('transport failed');
    return this.sendCompletion?.();
  }
  onOpen(handler: () => void): void { this.opened = handler; }
  onMessage(handler: (data: string) => void): void { this.received = handler; }
  onClose(handler: () => void): void { this.closed = handler; }
  onError(_handler: () => void): void {}
  receive(message: Record<string, unknown>): void { this.received(JSON.stringify(message)); }
}

const cleanup: Array<() => void> = [];
afterEach(() => { cleanup.splice(0).forEach(dispose => dispose()); });

function fixture(storage = new Map<string, unknown>(), requestNumber = 0, storageFails = false) {
  const transport = new RecoveryTransport();
  const client = new GameWebSocketClient({
    url: 'ws://test/ws', transportFactory: () => transport,
    requestIdFactory: () => `request-${++requestNumber}`,
  });
  const dependencies: RoomSessionDependencies = {
    createClient: () => client, configuredEndpoint: 'ws://test/ws', development: false,
    getPlayerId: () => 'temporary-player',
    getString: (key, fallback = '') => typeof storage.get(key) === 'string' ? storage.get(key) as string : fallback,
    getIdentity: key => storage.get(key) as StoredRoomIdentity | undefined ?? null,
    persistString: (key, value) => { if (storageFails) throw new Error('storage unavailable'); storage.set(key, value); },
    persistIdentity: (key, value) => { if (storageFails) throw new Error('storage unavailable'); storage.set(key, value); },
    removeValue: key => { storage.delete(key); },
  };
  const store = createRoomSessionStore(dependencies);
  store.getState().initialize();
  cleanup.push(() => store.getState().disconnect());
  return { transport, client, store, storage, dependencies };
}

it('旧恢复申请的终态不能清除当前申请票据或改变等待状态', () => {
  const { transport, client, store, storage } = fixture();
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  const ticket = storage.get(STORAGE_KEYS.recoveryTicket);
  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'old-request', roomId: 'room', playerId: 'player', recoveryStatus: 'expired' });
  expect(store.getState().recoveryStatus).toBe('pending');
  expect(storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  expect(store.getState().pendingCommand).toBe('resume');
  expect(client.hasPendingRequest).toBe(true);
  expect(store.getState().identity).toBeNull();
});

it('发送异常保留非秘密票据，用户手动重试同一申请后继续等待', () => {
  const { transport, store, storage } = fixture();
  transport.failSend = true;
  expect(() => store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' })).not.toThrow();
  expect(store.getState().recoveryStatus).toBe('failed');
  expect(store.getState().canRetryRecovery).toBe(true);
  expect(store.getState().errorMessage).toContain('审批结果未知');
  const ticket = storage.get(STORAGE_KEYS.recoveryTicket);
  expect(JSON.parse(ticket as string)).toEqual({ requestId: 'request-1', roomId: 'room', playerId: 'player' });
  expect(JSON.stringify([...storage.values()])).not.toContain('private-recovery');
  expect(store.getState().identity).toBeNull();
  expect(transport.sent).toHaveLength(1);

  transport.failSend = false;
  store.getState().retryRecovery();
  expect(transport.sent).toHaveLength(2);
  expect(transport.sent[1]).toEqual(transport.sent[0]);
  expect(storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  expect(store.getState().recoveryStatus).toBe('pending');
  expect(store.getState().canRetryRecovery).toBe(false);
  expect(store.getState().errorMessage).toBe('');
});

it('不完整批准或未知恢复状态不能结束当前申请', () => {
  const { transport, client, store, storage } = fixture();
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  const ticket = storage.get(STORAGE_KEYS.recoveryTicket);
  for (const invalid of [
    { recoveryStatus: 'approved', resumeCredential: 'new-device' },
    { recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: '' },
    { recoveryStatus: 'approved', resumeCredential: 123, recoveryCredential: 'new-recovery' },
    { recoveryStatus: 'approved', resumeCredential: '  ', recoveryCredential: 'new-recovery' },
    { recoveryStatus: 'unexpected' },
  ]) {
    transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', ...invalid });
    expect(store.getState().recoveryStatus).toBe('pending');
    expect(storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
    expect(client.identity).toBeNull();
    expect(client.hasPendingRequest).toBe(true);
    expect(transport.sent.map(message => message.type)).toEqual(['REQUEST_RECOVERY']);
  }
});

it('普通错误保留原申请供手动重试，迟到等待消息不能重新锁住重试，迟到批准仍有效', () => {
  const { transport, client, store, storage } = fixture();
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  const ticket = storage.get(STORAGE_KEYS.recoveryTicket);
  transport.receive({ type: 'ERROR', code: 'INTERNAL' });
  expect(store.getState().recoveryStatus).toBe('failed');
  expect(store.getState().canRetryRecovery).toBe(true);
  expect(store.getState().errorMessage).not.toContain('未获批准');
  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'pending' });
  expect(store.getState().recoveryStatus).toBe('failed');
  expect(store.getState().canRetryRecovery).toBe(true);
  client.disconnect();
  client.connect();
  expect(transport.sent.map(message => message.type)).toEqual(['REQUEST_RECOVERY']);
  expect(storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  transport.receive({ type: 'ROOM_STATE', roomRevision: 7, state: { secret: 'role' } });
  expect(store.getState().experience.roomState).toBeNull();
  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(store.getState().recoveryStatus).toBe('approved');
  expect(store.getState().identity?.resumeCredential).toBe('new-device');
  expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
  expect(transport.sent.at(-1)?.type).toBe('RESUME_ROOM');
});

it('批准双凭据必须在入房请求前保存，紧接的无效身份回复不能被外层批准覆盖', () => {
  const { transport, client, store, storage } = fixture();
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  let savedBeforeResume: unknown;
  transport.onSend = message => {
    if (message.type !== 'RESUME_ROOM') return;
    savedBeforeResume = storage.get(STORAGE_KEYS.roomIdentity);
    transport.receive({ type: 'ERROR', code: 'INVALID_CREDENTIAL' });
  };
  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(savedBeforeResume).toEqual({ version: 2, roomId: 'room', playerId: 'player', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(store.getState().identity).toBeNull();
  expect(client.identity).toBeNull();
  expect(store.getState().recoveryStatus).toBeNull();
  expect(store.getState().canRetryRecovery).toBe(false);
  expect(store.getState().experience.terminalReason).toBe('invalid-credential');
  expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
  expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
});

it('发送失败的票据在刷新后仍复用，恢复码只在重新粘贴后进入内存', () => {
  const first = fixture();
  first.transport.failSend = true;
  const imported = { roomId: 'room', playerId: 'player', recoveryCredential: 'NEVER-PERSIST' };
  first.store.getState().importRoomIdentity(imported);
  const ticket = first.storage.get(STORAGE_KEYS.recoveryTicket);
  first.store.getState().disconnect();
  const refreshed = fixture(first.storage, 100);
  expect(refreshed.transport.sent).toEqual([]);
  expect(refreshed.store.getState().identity).toBeNull();
  refreshed.store.getState().importRoomIdentity(imported);
  expect(refreshed.transport.sent).toEqual([first.transport.sent[0]]);
  expect(refreshed.storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  expect(JSON.stringify([...refreshed.storage.values()])).not.toContain('NEVER-PERSIST');
  expect(refreshed.storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
});

it.each(['rejected', 'expired'])('服务端终态%s清除票据并允许新的申请，不建立身份', recoveryStatus => {
  const { transport, client, store, storage } = fixture();
  const imported = { roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' };
  store.getState().importRoomIdentity(imported);
  expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(true);
  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus });
  expect(store.getState().recoveryStatus).toBe(recoveryStatus);
  expect(client.hasPendingRequest).toBe(false);
  expect(store.getState().pendingCommand).toBeNull();
  expect(store.getState().canRetryRecovery).toBe(false);
  expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
  expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
  store.getState().importRoomIdentity(imported);
  expect(transport.sent.at(-1)?.requestId).toBe('request-2');
});

it('恢复结果未知时拒绝另一份身份操作，避免迟到批准覆盖新身份', () => {
  const { transport, client, store } = fixture();
  transport.failSend = true;
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  transport.failSend = false;
  expect(() => client.createRoom('new-player', '新玩家', 5)).toThrow('恢复申请尚未结束');
  expect(() => client.joinRoom('another-room', 'new-player', '新玩家')).toThrow('恢复申请尚未结束');
  expect(() => client.requestRecovery('other-room', 'other-player', 'other-secret')).toThrow('恢复申请尚未结束');
  expect(transport.sent).toHaveLength(1);
  expect(client.recovery?.requestId).toBe('request-1');
});

it('异步发送失败同样允许手动重试原申请', async () => {
  const { transport, client, store } = fixture();
  let rejectSend: (reason?: unknown) => void = () => {};
  const completion = new Promise<void>((_, reject) => { rejectSend = reject; });
  void completion.catch(() => {});
  transport.sendCompletion = () => completion;
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  rejectSend(new Error('async send failed'));
  await Promise.resolve();
  expect(store.getState().recoveryStatus).toBe('failed');
  expect(store.getState().canRetryRecovery).toBe(true);
  transport.sendCompletion = undefined;
  store.getState().retryRecovery();
  expect(transport.sent[1]).toEqual(transport.sent[0]);
  expect(client.recovery?.status).toBe('pending');
});

it('批准后入房发送失败可重连，后续普通错误不能丢掉已批准的恢复凭据', () => {
  const { transport, client, store, storage } = fixture();
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  transport.failSend = true;
  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  const approvedIdentity = storage.get(STORAGE_KEYS.roomIdentity);
  expect(store.getState().recoveryStatus).toBe('approved');
  expect(store.getState().status).toBe('error');
  expect(store.getState().canRetryRecovery).toBe(false);
  expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
  transport.failSend = false;
  store.getState().connect();
  expect(transport.sent.map(message => message.type)).toEqual(['REQUEST_RECOVERY', 'RESUME_ROOM', 'RESUME_ROOM']);
  expect(transport.sent.at(-1)?.resumeCredential).toBe('new-device');
  transport.receive({ type: 'RESUME_ROOM_RESULT', roomId: 'room', nextClientSequence: 1 });
  store.getState().resumeLastRoom();
  transport.receive({ type: 'ERROR', code: 'INTERNAL' });
  expect(client.identity?.recoveryCredential).toBe('new-recovery');
  expect(storage.get(STORAGE_KEYS.roomIdentity)).toEqual(approvedIdentity);
  expect(store.getState().recoveryStatus).toBe('approved');
});

it('丢弃客户端时清理内存恢复视图但保留刷新可用的非秘密票据', () => {
  const { transport, store, storage } = fixture();
  transport.failSend = true;
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  const ticket = storage.get(STORAGE_KEYS.recoveryTicket);
  store.getState().disconnect();
  expect(store.getState().recoveryStatus).toBeNull();
  expect(store.getState().canRetryRecovery).toBe(false);
  expect(store.getState().pendingCommand).toBeNull();
  expect(storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
});

it.each(['retry', 'reconnect', 'approved'] as const)('旧发送Promise迟到失败不能回退%s之后的恢复状态', async scenario => {
  const { transport, client, store } = fixture();
  let rejectOld: (reason?: unknown) => void = () => {};
  const completion = new Promise<void>((_, reject) => { rejectOld = reject; });
  transport.sendCompletion = () => completion;
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  transport.sendCompletion = undefined;
  if (scenario === 'retry') {
    transport.receive({ type: 'ERROR', code: 'INTERNAL' });
    store.getState().retryRecovery();
  } else if (scenario === 'reconnect') {
    client.disconnect();
    client.connect();
  } else {
    transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  }
  rejectOld(new Error('old attempt failed'));
  await Promise.resolve();
  expect(store.getState().recoveryStatus).toBe(scenario === 'approved' ? 'approved' : 'pending');
  expect(store.getState().canRetryRecovery).toBe(false);
  expect(store.getState().errorMessage).toBe('');
});

it('批准后附带查询异步失败仍可通过重连继续使用已批准身份', async () => {
  const { transport, store } = fixture();
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  transport.sendCompletion = () => Promise.reject(new Error('query send failed'));
  transport.receive({ type: 'RESUME_ROOM_RESULT', roomId: 'room', nextClientSequence: 1 });
  await Promise.resolve();
  expect(store.getState().status).toBe('error');
  const beforeReconnect = transport.sent.length;
  transport.sendCompletion = undefined;
  store.getState().connect();
  expect(transport.sent.length).toBe(beforeReconnect + 1);
  expect(transport.sent.at(-1)?.type).toBe('RESUME_ROOM');
  expect(store.getState().identity?.recoveryCredential).toBe('new-recovery');
});

it('设备存储失败不阻断申请与批准通知，当前身份可继续使用并明确提示未保存', () => {
  const { transport, store, storage } = fixture(new Map(), 0, true);
  expect(() => store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' })).not.toThrow();
  expect(transport.sent.at(-1)?.type).toBe('REQUEST_RECOVERY');
  expect(store.getState().errorMessage).toContain('保存失败');
  expect(() => transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' })).not.toThrow();
  expect(store.getState().identity?.resumeCredential).toBe('new-device');
  expect(store.getState().errorMessage).toContain('保存失败');
  expect(transport.sent.at(-1)?.type).toBe('RESUME_ROOM');
  expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
  transport.receive({ type: 'RESUME_ROOM_RESULT', roomId: 'room', nextClientSequence: 1 });
  expect(store.getState().errorMessage).toContain('保存失败');
});

it('批准身份首次保存失败时保留票据，入房回复补存成功后清理票据', () => {
  const { transport, store, storage, dependencies } = fixture();
  store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  const ticket = storage.get(STORAGE_KEYS.recoveryTicket);
  vi.spyOn(dependencies, 'persistIdentity').mockImplementationOnce(() => { throw new Error('temporary storage failure'); });

  transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
  expect(storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  expect(store.getState().errorMessage).toContain('保存失败');

  transport.receive({ type: 'RESUME_ROOM_RESULT', roomId: 'room', nextClientSequence: 1 });
  expect(storage.get(STORAGE_KEYS.roomIdentity)).toEqual({ version: 2, roomId: 'room', playerId: 'player', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
  expect(store.getState().errorMessage).toBe('');
  expect(store.getState().recoveryStatus).toBe('approved');
});

it.each(['legacy-room', 'ticket-removal'] as const)('批准后%s存储失败，刷新并恢复同一身份后补清票据', failure => {
  const first = fixture();
  first.store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'private-recovery' });
  const ticket = first.storage.get(STORAGE_KEYS.recoveryTicket);
  if (failure === 'legacy-room') {
    vi.spyOn(first.dependencies, 'persistString').mockImplementation((key, value) => {
      if (key === STORAGE_KEYS.legacyRoomId) throw new Error('temporary storage failure');
      first.storage.set(key, value);
    });
  } else {
    vi.spyOn(first.dependencies, 'removeValue').mockImplementation(() => { throw new Error('temporary storage failure'); });
  }
  first.transport.receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', roomId: 'room', playerId: 'player', recoveryStatus: 'approved', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  const savedIdentity = first.storage.get(STORAGE_KEYS.roomIdentity);
  expect(savedIdentity).toEqual({ version: 2, roomId: 'room', playerId: 'player', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(first.storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  expect(first.store.getState().errorMessage).toContain('失败');

  first.store.getState().disconnect();
  const refreshed = fixture(first.storage, 100);
  expect(refreshed.client.recovery).toBeNull();
  expect(refreshed.transport.sent.at(-1)?.type).toBe('RESUME_ROOM');
  expect(refreshed.storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  refreshed.transport.receive({ type: 'RESUME_ROOM_RESULT', roomId: 'room', recoveryCredential: 'new-recovery', nextClientSequence: 1 });
  expect(refreshed.storage.get(STORAGE_KEYS.roomIdentity)).toEqual(savedIdentity);
  expect(refreshed.storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
  expect(refreshed.store.getState().errorMessage).toBe('');
});

it.each([
  { roomId: 'other-room', playerId: 'player' },
  { roomId: 'room', playerId: 'other-player' },
])('恢复身份后保留属于其它房间或玩家的票据：%j', ticketIdentity => {
  const storage = new Map<string, unknown>();
  storage.set(STORAGE_KEYS.roomIdentity, { version: 2, roomId: 'room', playerId: 'player', resumeCredential: 'device', recoveryCredential: 'recovery' });
  const ticket = JSON.stringify({ requestId: 'other-request', ...ticketIdentity });
  storage.set(STORAGE_KEYS.recoveryTicket, ticket);
  const { transport, store } = fixture(storage);

  transport.receive({ type: 'RESUME_ROOM_RESULT', roomId: 'room', nextClientSequence: 1 });
  expect(storage.get(STORAGE_KEYS.recoveryTicket)).toBe(ticket);
  expect(store.getState().errorMessage).toBe('');
});
