import { describe, expect, it, vi } from 'vitest';

vi.mock('@tarojs/taro', () => ({
  default: {
    connectSocket: vi.fn(),
    getStorageSync: vi.fn(),
    setStorageSync: vi.fn(),
    removeStorageSync: vi.fn(),
  },
}));
import type {
  ClientRoomIdentity,
  ConnectionStatus,
  GameWebSocketClient,
  ServerMessage,
} from '@clocktower/core';
import {
  createRoomSessionStore,
  STORAGE_KEYS,
  type RoomSessionDependencies,
} from './room-session-store';

class FakeClient {
  status: ConnectionStatus = 'disconnected';
  identity: ClientRoomIdentity | null = null;
  hasPendingRequest = false;
  readonly calls: Array<{ readonly method: string; readonly args: readonly unknown[] }> = [];
  private readonly messageHandlers = new Set<(message: ServerMessage) => void>();
  private readonly statusHandlers = new Set<(status: ConnectionStatus) => void>();

  connect(): void {
    this.calls.push({ method: 'connect', args: [] });
    this.status = 'connected';
    this.statusHandlers.forEach((handler) => handler(this.status));
  }

  disconnect(): void {
    this.status = 'disconnected';
    this.statusHandlers.forEach((handler) => handler(this.status));
  }

  onMessage(handler: (message: ServerMessage) => void): () => void {
    this.messageHandlers.add(handler);
    return () => this.messageHandlers.delete(handler);
  }

  onStatusChange(handler: (status: ConnectionStatus) => void): () => void {
    this.statusHandlers.add(handler);
    return () => this.statusHandlers.delete(handler);
  }

  emit(message: ServerMessage): void {
    this.messageHandlers.forEach((handler) => handler(message));
  }

  createRoom(...args: readonly unknown[]): void { this.record('createRoom', args); }
  joinRoom(...args: readonly unknown[]): void { this.record('joinRoom', args); }
  requestRecovery(...args: readonly unknown[]): string { this.record('requestRecovery', args); return typeof args[3] === 'string' ? args[3] : 'recovery-ticket-1'; }
  getRecoveryRequests(...args: readonly unknown[]): void { this.record('getRecoveryRequests', args); }
  resumeRoom(roomId: string, playerId: string, resumeCredential: string): void {
    this.identity = { roomId, playerId, resumeCredential };
    this.record('resumeRoom', [roomId, playerId, resumeCredential]);
  }
  rejoinRoom(...args: readonly unknown[]): void { this.record('rejoinRoom', args); }
  getRoomState(...args: readonly unknown[]): void { this.record('getRoomState', args); }
  closeRoom(...args: readonly unknown[]): void { this.record('closeRoom', args); }
  transferOwnership(...args: readonly unknown[]): void { this.record('transferOwnership', args); }
  restartGame(...args: readonly unknown[]): void { this.record('restartGame', args); }
  leaveRoom(...args: readonly unknown[]): void { this.record('leaveRoom', args); }
  kickPlayer(...args: readonly unknown[]): void { this.record('kickPlayer', args); }
  updateRoomSettings(...args: readonly unknown[]): void { this.record('updateRoomSettings', args); }
  setStoryteller(...args: readonly unknown[]): void { this.record('setStoryteller', args); }
  setSeatOrder(...args: readonly unknown[]): void { this.record('setSeatOrder', args); }
  setReady(...args: readonly unknown[]): void { this.record('setReady', args); }
  confirmCharacter(...args: readonly unknown[]): void { this.record('confirmCharacter', args); }
  assignCharacters(...args: readonly unknown[]): void { this.record('assignCharacters', args); }
  startGame(...args: readonly unknown[]): void { this.record('startGame', args); }
  changePhase(...args: readonly unknown[]): void { this.record('changePhase', args); }
  finalizeDay(...args: readonly unknown[]): void { this.record('finalizeDay', args); }
  nominate(...args: readonly unknown[]): void { this.record('nominate', args); }
  castVote(...args: readonly unknown[]): void { this.record('castVote', args); }
  recordVote(...args: readonly unknown[]): void { this.record('recordVote', args); }
  resolveNomination(...args: readonly unknown[]): void { this.record('resolveNomination', args); }
  executePlayer(...args: readonly unknown[]): void { this.record('executePlayer', args); }
  useSlayerAbility(...args: readonly unknown[]): void { this.record('useSlayerAbility', args); }
  killPlayer(...args: readonly unknown[]): void { this.record('killPlayer', args); }
  submitNightAction(...args: readonly unknown[]): void { this.record('submitNightAction', args); }
  confirmNightAction(...args: readonly unknown[]): void { this.record('confirmNightAction', args); }
  acknowledgeNightAction(...args: readonly unknown[]): void { this.record('acknowledgeNightAction', args); }
  skipNightAction(...args: readonly unknown[]): void { this.record('skipNightAction', args); }
  prepareDawn(...args: readonly unknown[]): void { this.record('prepareDawn', args); }
  confirmDawn(...args: readonly unknown[]): void { this.record('confirmDawn', args); }
  resolveNight(...args: readonly unknown[]): void { this.record('resolveNight', args); }
  endGame(...args: readonly unknown[]): void { this.record('endGame', args); }
  publishGrimoire(...args: readonly unknown[]): void { this.record('publishGrimoire', args); }

  private record(method: string, args: readonly unknown[]): void {
    this.calls.push({ method, args });
  }
}

function fixture(storedIdentity: ClientRoomIdentity | null = null) {
  const client = new FakeClient();
  const storage = new Map<string, unknown>();
  if (storedIdentity) {
    storage.set(STORAGE_KEYS.roomIdentity, { version: 2, ...storedIdentity });
  }
  const dependencies: RoomSessionDependencies = {
    createClient: () => client as unknown as GameWebSocketClient,
    configuredEndpoint: 'ws://test/ws',
    development: false,
    getPlayerId: () => 'generated-player',
    getString: (key, fallback = '') => {
      const value = storage.get(key);
      return typeof value === 'string' && value.trim() ? value : fallback;
    },
    getIdentity: (key) => {
      const value = storage.get(key);
      return value && typeof value === 'object' ? value as never : null;
    },
    persistString: (key, value) => { storage.set(key, value); },
    persistIdentity: (key, value) => { storage.set(key, value); },
    removeValue: (key) => { storage.delete(key); },
  };
  return { client, storage, dependencies, store: createRoomSessionStore(dependencies) };
}

describe('room session store', () => {
  it('reuses a non-secret recovery ticket after refreshing and repasting the code', () => {
    const { client, storage, dependencies, store } = fixture();
    store.getState().initialize();
    const recovery = { roomId: 'room', playerId: 'player', recoveryCredential: 'NEVER-PERSIST' };
    store.getState().importRoomIdentity(recovery);
    expect([...storage.values()].join(' ')).not.toContain('NEVER-PERSIST');
    const refreshed = createRoomSessionStore(dependencies);
    refreshed.getState().initialize();
    refreshed.getState().importRoomIdentity(recovery);
    expect(client.calls.filter(call => call.method === 'requestRecovery').at(-1)?.args).toEqual(['room', 'player', 'NEVER-PERSIST', 'recovery-ticket-1']);
  });
  it('requests approval without persisting or resuming the imported identity', () => {
    const { client, storage, store } = fixture();
    store.getState().initialize();
    const identity = { roomId: 'ROOM2', playerId: 'old-player', recoveryCredential: 'recovery-secret' };
    store.getState().importRoomIdentity(identity);
    expect(store.getState().identity).toBeNull();
    expect(storage.get(STORAGE_KEYS.roomIdentity)).toBeUndefined();
    expect(client.calls.filter(call => call.method === 'resumeRoom')).toEqual([]);
    expect(client.calls).toContainEqual({ method: 'requestRecovery', args: ['ROOM2', 'old-player', 'recovery-secret'] });
    expect(client.calls.some(call => call.method === 'joinRoom')).toBe(false);
  });
  it('initializes one client and resumes the stored identity once', () => {
    const identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };
    const { client, store } = fixture(identity);

    store.getState().initialize();
    store.getState().initialize();

    expect(client.calls.filter((call) => call.method === 'connect')).toHaveLength(1);
    expect(client.calls.filter((call) => call.method === 'resumeRoom')).toEqual([
      { method: 'resumeRoom', args: ['ROOM1', 'p1', 'secret'] },
    ]);
    expect(store.getState().playerId).toBe('p1');
  });

  it('blocks joining a different room while a recoverable identity exists', () => {
    const { client, store } = fixture({ roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' });
    store.getState().initialize();

    store.getState().joinRoom('ROOM2');

    expect(client.calls.some((call) => call.method === 'joinRoom')).toBe(false);
    expect(store.getState().errorMessage).toContain('其他房间身份');
  });

  it('persists a newly created identity and authoritative projection', () => {
    const { client, storage, store } = fixture();
    store.getState().initialize();
    client.identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };

    client.emit({
      type: 'CREATE_ROOM_RESULT',
      roomId: 'ROOM1',
      resumeCredential: 'secret',
      roomRevision: 1,
      state: {
        roomId: 'ROOM1',
        players: [],
        maxPlayers: 5,
        scriptId: 'trouble_brewing',
        scriptName: '暗流涌动',
        phase: 1,
        dayNumber: 0,
        nightNumber: 0,
      },
    });

    expect(store.getState().experience.roomState?.roomId).toBe('ROOM1');
    expect(storage.get(STORAGE_KEYS.roomIdentity)).toEqual({
      version: 2,
      roomId: 'ROOM1',
      playerId: 'p1',
      resumeCredential: 'secret',
    });
  });

  it('keeps the terminal reason while clearing an invalid identity', () => {
    const { storage, client, store } = fixture({ roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' });
    store.getState().initialize();

    client.emit({ type: 'ERROR', code: 'INVALID_CREDENTIAL', error: 'invalid credential' });

    expect(store.getState().identity).toBeNull();
    expect(store.getState().experience.terminalReason).toBe('invalid-credential');
    expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
  });

  it('does not discard a recoverable identity in production', () => {
    const identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };
    const { storage, store } = fixture(identity);

    store.getState().initialize();
    store.getState().forgetRoomIdentity();

    expect(store.getState().identity).toMatchObject(identity);
    expect(store.getState().errorMessage).toContain('生产环境');
    expect(storage.get(STORAGE_KEYS.roomIdentity)).toMatchObject(identity);
  });

  it('allows a retained identity to be cleared after rejoin becomes impossible', () => {
    const identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };
    const { client, storage, store } = fixture(identity);
    store.getState().initialize();
    client.emit({
      type: 'RESUME_ROOM_RESULT',
      roomId: 'ROOM1',
      identityStatus: {
        status: 'retained',
        canRejoin: false,
        participantSetFrozen: true,
      },
      roomRevision: 4,
      nextClientSequence: 2,
    });

    store.getState().forgetRoomIdentity();

    expect(store.getState().identity).toBeNull();
    expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
  });

  it('clears an obsolete projection when the room no longer exists', () => {
    const identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };
    const { client, store } = fixture(identity);
    store.getState().initialize();
    client.emit({
      type: 'ROOM_STATE',
      roomRevision: 3,
      state: {
        roomId: 'ROOM1',
        players: [],
        maxPlayers: 5,
        scriptId: 'trouble_brewing',
        scriptName: '暗流涌动',
        phase: 3,
        dayNumber: 0,
        nightNumber: 1,
      },
    });

    client.emit({ type: 'ERROR', code: 'ROOM_NOT_FOUND', error: 'room not found' });

    expect(store.getState().identity).toBeNull();
    expect(store.getState().experience.roomState).toBeNull();
    expect(store.getState().experience.terminalReason).toBe('closed');
  });

  it('keeps the authoritative projection while reconnecting', () => {
    const identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };
    const { client, store } = fixture(identity);
    store.getState().initialize();
    client.emit({
      type: 'ROOM_STATE',
      roomRevision: 3,
      state: {
        roomId: 'ROOM1',
        players: [],
        maxPlayers: 5,
        scriptId: 'trouble_brewing',
        scriptName: '暗流涌动',
        phase: 3,
        dayNumber: 0,
        nightNumber: 1,
      },
    });

    client.disconnect();
    store.getState().connect();

    expect(store.getState().experience.roomRevision).toBe(3);
    expect(client.calls.filter((call) => call.method === 'connect')).toHaveLength(2);
  });

  it('keeps a replayed vote pending until the core client confirms it', () => {
    const identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };
    const { client, store } = fixture(identity);
    store.getState().initialize();
    store.getState().castVote(true);
    client.hasPendingRequest = true;

    client.emit({
      type: 'RESUME_ROOM_RESULT',
      roomId: 'ROOM1',
      roomRevision: 8,
      nextClientSequence: 4,
      state: {
        roomId: 'ROOM1',
        players: [],
        maxPlayers: 5,
        scriptId: 'trouble_brewing',
        scriptName: '暗流涌动',
        phase: 4,
        dayNumber: 1,
        nightNumber: 1,
      },
    });
    expect(store.getState().pendingCommand).toBe('vote');

    client.hasPendingRequest = false;
    client.emit({
      type: 'COMMAND_RESULT',
      acceptedSequence: 4,
      nextClientSequence: 5,
      roomRevision: 9,
    });
    expect(store.getState().pendingCommand).toBeNull();
  });
  it('persists approved credentials only when the core has accepted the grant', () => {
    const { client, storage, store } = fixture();
    store.getState().initialize();
    store.getState().importRoomIdentity({ roomId: 'ROOM2', playerId: 'old-player', recoveryCredential: 'old-recovery' });
    client.identity = { roomId: 'ROOM2', playerId: 'old-player', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' };
    client.emit({ type: 'RECOVERY_STATUS', recoveryStatus: 'approved' });
    expect(storage.get(STORAGE_KEYS.roomIdentity)).toEqual({ version: 2, ...client.identity });
    expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
    expect(store.getState().experience.roomState).toBeNull();
  });

  it('clears only the pending ticket when recovery is rejected or expires', () => {
    for (const recoveryStatus of ['rejected', 'expired']) {
      const { client, storage, store } = fixture();
      store.getState().initialize();
      store.getState().importRoomIdentity({ roomId: 'room', playerId: 'player', recoveryCredential: 'secret' });
      expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(true);
      client.emit({ type: 'RECOVERY_STATUS', recoveryStatus });
      expect(storage.has(STORAGE_KEYS.recoveryTicket)).toBe(false);
      expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
    }
  });

  it('removes local identity and private state when a new device replaces this session', () => {
    const { client, storage, store } = fixture({ roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'old' });
    store.getState().initialize();
    client.emit({ type: 'SESSION_REPLACED' });
    expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
    expect(store.getState().identity).toBeNull();
    expect(store.getState().experience.roomState).toBeNull();
    expect(store.getState().status).toBe('disconnected');
  });

  it('forwards the complete seat order without changing the authoritative projection locally', () => {
    const { client, store } = fixture();
    store.getState().initialize();
    const seatOrder = ['p3', 'p1', 'p2'];

    store.getState().setSeatOrder(seatOrder);

    expect(client.calls.find((call) => call.method === 'setSeatOrder')).toEqual({
      method: 'setSeatOrder',
      args: [seatOrder],
    });
    expect(store.getState().pendingCommand).toBe('settings');
    expect(store.getState().experience.roomState).toBeNull();
  });

  it('forwards readiness without changing the authoritative projection locally', () => {
    const { client, store } = fixture();
    store.getState().initialize();
    client.emit({
      type: 'ROOM_STATE',
      roomRevision: 1,
      state: {
        roomId: 'ROOM1',
        players: [{
          id: 'generated-player',
          name: 'P1',
          isAlive: true,
          votes: 0,
          isReady: false,
          hasConfirmedCharacter: false,
        }],
        maxPlayers: 5,
        scriptId: 'trouble_brewing',
        scriptName: '暗流涌动',
        phase: 1,
        dayNumber: 0,
        nightNumber: 0,
      },
    });

    store.getState().setReady(true);

    expect(client.calls.find((call) => call.method === 'setReady')).toEqual({
      method: 'setReady',
      args: [true],
    });
    expect(store.getState().pendingCommand).toBe('ready');
    expect(store.getState().experience.roomState?.players[0]?.isReady).toBe(false);
  });

  it('forwards character confirmation without changing the authoritative projection locally', () => {
    const { client, store } = fixture();
    store.getState().initialize();
    client.emit({
      type: 'ROOM_STATE',
      roomRevision: 1,
      state: {
        roomId: 'ROOM1',
        players: [{
          id: 'generated-player',
          name: 'P1',
          character: { id: 'slayer', name: '猎手', team: 1, ability: '能力' },
          isAlive: true,
          votes: 0,
          isReady: true,
          hasConfirmedCharacter: false,
        }],
        maxPlayers: 5,
        scriptId: 'trouble_brewing',
        scriptName: '暗流涌动',
        phase: 1,
        dayNumber: 0,
        nightNumber: 0,
      },
    });

    store.getState().confirmCharacter();

    expect(client.calls.find((call) => call.method === 'confirmCharacter')).toEqual({
      method: 'confirmCharacter',
      args: [],
    });
    expect(store.getState().pendingCommand).toBe('confirm-character');
    expect(store.getState().experience.roomState?.players[0]?.hasConfirmedCharacter).toBe(false);
  });

  it('forwards ownership transfer and restart through the server-authoritative store', () => {
    const transfer = fixture();
    transfer.store.getState().initialize();
    transfer.store.getState().transferOwnership('member');
    expect(transfer.client.calls.at(-1)).toEqual({ method: 'transferOwnership', args: ['member'] });
    expect(transfer.store.getState().pendingCommand).toBe('transfer-ownership');
    const restart = fixture();
    restart.store.getState().initialize();
    restart.store.getState().restartGame();
    expect(restart.client.calls.at(-1)).toEqual({ method: 'restartGame', args: [] });
    expect(restart.store.getState().pendingCommand).toBe('restart-game');
  });

  it('forwards day finalization as its own pending command', () => {
    const { client, store } = fixture();
    store.getState().initialize();

    store.getState().finalizeDay();

    expect(client.calls.find((call) => call.method === 'finalizeDay')).toEqual({
      method: 'finalizeDay',
      args: [],
    });
    expect(store.getState().pendingCommand).toBe('finalize-day');
  });

  it('forwards a storyteller proxy decision as its own pending command', () => {
    const { client, store } = fixture();
    store.getState().initialize();

    store.getState().recordVote('p4', false);

    expect(client.calls.find((call) => call.method === 'recordVote')).toEqual({
      method: 'recordVote',
      args: ['p4', false],
    });
    expect(store.getState().pendingCommand).toBe('record-vote');
  });

  it('forwards the three night review lifecycle actions', () => {
    const { client, store } = fixture();
    store.getState().initialize();

    store.getState().confirmNightAction(['p2'], '确认结果');
    store.getState().acknowledgeNightAction();
    store.getState().skipNightAction();

    expect(client.calls.filter((call) => [
      'confirmNightAction',
      'acknowledgeNightAction',
      'skipNightAction',
    ].includes(call.method))).toEqual([
      { method: 'confirmNightAction', args: [['p2'], '确认结果'] },
      { method: 'acknowledgeNightAction', args: [] },
      { method: 'skipNightAction', args: [] },
    ]);
    expect(store.getState().pendingCommand).toBe('skip-night-action');
  });

  it('forwards Demon bluffs with the character setup', () => {
    const { client, store } = fixture();
    store.getState().initialize();

    store.getState().assignCharacters(
      { p1: 'washerwoman', p2: 'imp' },
      {},
      undefined,
      ['chef', 'empath', 'fortuneteller'],
    );

    expect(client.calls.find((call) => call.method === 'assignCharacters')).toEqual({
      method: 'assignCharacters',
      args: [
        { p1: 'washerwoman', p2: 'imp' },
        {},
        undefined,
        ['chef', 'empath', 'fortuneteller'],
      ],
    });
  });

  it('forwards grimoire publication as its own pending command', () => {
    const { client, store } = fixture();
    store.getState().initialize();

    store.getState().publishGrimoire();

    expect(client.calls.find((call) => call.method === 'publishGrimoire')).toEqual({
      method: 'publishGrimoire',
      args: [],
    });
    expect(store.getState().pendingCommand).toBe('publish-grimoire');
  });

  it('forwards dawn preparation and edited death confirmation', () => {
    const { client, store } = fixture();
    store.getState().initialize();

    store.getState().prepareDawn();
    store.getState().confirmDawn(['p2', 'p4']);

    expect(client.calls.filter((call) => ['prepareDawn', 'confirmDawn'].includes(call.method))).toEqual([
      { method: 'prepareDawn', args: [] },
      { method: 'confirmDawn', args: [['p2', 'p4']] },
    ]);
    expect(store.getState().pendingCommand).toBe('confirm-dawn');
  });

  it('clears local identity only after a finished leave is confirmed', () => {
    const identity = { roomId: 'ROOM1', playerId: 'p1', resumeCredential: 'secret' };
    const { client, storage, store } = fixture(identity);
    store.getState().initialize();
    client.emit({
      type: 'ROOM_STATE',
      roomRevision: 10,
      state: {
        roomId: 'ROOM1',
        players: [],
        maxPlayers: 5,
        scriptId: 'trouble_brewing',
        scriptName: '暗流涌动',
        phase: 5,
        dayNumber: 2,
        nightNumber: 2,
        winner: { winner: 1, reason: 'imp_executed', description: '善良获胜' },
      },
    });
    store.getState().leaveRoom();
    client.hasPendingRequest = false;

    client.emit({
      type: 'COMMAND_RESULT',
      acceptedSequence: 2,
      nextClientSequence: 3,
      roomRevision: 11,
    });

    expect(store.getState().identity).toBeNull();
    expect(store.getState().experience.roomState).toBeNull();
    expect(storage.has(STORAGE_KEYS.roomIdentity)).toBe(false);
  });
});
