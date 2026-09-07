import { useStore } from 'zustand';
import { createStore, type StoreApi } from 'zustand/vanilla';
import {
  GameWebSocketClient,
  TROUBLE_BREWING_SCRIPT,
  type ClientRoomIdentity,
  type ConnectionStatus,
  type ServerMessage,
} from '@clocktower/core';
import { createTaroWebSocketTransport } from './taro-websocket-transport';
import {
  applyServerMessage,
  clearRoomIdentityState,
  createRoomExperienceState,
  selectGamePhase,
  type RoomExperienceState,
} from './room-experience';
import {
  getOrCreatePlayerId,
  getStoredRoomIdentity,
  getStoredString,
  persistRoomIdentity,
  persistString,
  removeStoredValue,
  type StoredRoomIdentity,
} from './utils';
import { protocolErrorMessage } from './protocol-feedback';
import type { RecoveryIdentity } from './room-recovery';

export const STORAGE_KEYS = {
  playerName: 'clocktower.playerName',
  endpoint: 'clocktower.wsUrl',
  roomIdentity: 'clocktower.roomIdentity.v2',
  legacyRoomId: 'clocktower.lastRoomId',
  maxPlayers: 'clocktower.maxPlayers',
  recoveryTicket: 'clocktower.recoveryTicket.v1',
} as const;

export const DEFAULT_SCRIPT_ID = TROUBLE_BREWING_SCRIPT.id;
const DEVELOPMENT_WS_URL = 'ws://localhost:8080/ws';

type PendingCommand =
  | 'transfer-ownership'
  | 'restart-game'
  | 'create'
  | 'join'
  | 'resume'
  | 'rejoin'
  | 'leave'
  | 'close'
  | 'settings'
  | 'storyteller'
  | 'ready'
  | 'confirm-character'
  | 'assign'
  | 'start'
  | 'nominate'
  | 'vote'
  | 'record-vote'
  | 'resolve-vote'
  | 'phase'
  | 'finalize-day'
  | 'slayer'
  | 'death'
  | 'night-action'
  | 'confirm-night-action'
  | 'acknowledge-night-action'
  | 'skip-night-action'
  | 'prepare-dawn'
  | 'confirm-dawn'
  | 'resolve-night'
  | 'end-game'
  | 'publish-grimoire';

type SessionClient = Pick<
  GameWebSocketClient,
  | 'status'
  | 'identity'
  | 'hasPendingRequest'
  | 'connect'
  | 'disconnect'
  | 'onMessage'
  | 'onStatusChange'
  | 'createRoom'
  | 'joinRoom'
  | 'resumeRoom'
  | 'requestRecovery'
  | 'reviewRecovery'
  | 'getRecoveryRequests'
  | 'undoGame'
  | 'redoGame'
  | 'rejoinRoom'
  | 'getRoomState'
  | 'closeRoom'
  | 'transferOwnership'
  | 'restartGame'
  | 'leaveRoom'
  | 'kickPlayer'
  | 'updateRoomSettings'
  | 'setStoryteller'
  | 'setSeatOrder'
  | 'setReady'
  | 'confirmCharacter'
  | 'assignCharacters'
  | 'startGame'
  | 'changePhase'
  | 'finalizeDay'
  | 'nominate'
  | 'castVote'
  | 'recordVote'
  | 'advanceNominationStage'
  | 'controlNominationTimer'
  | 'expireNominationTimer'
  | 'resolveNomination'
  | 'executePlayer'
  | 'useSlayerAbility'
  | 'killPlayer'
  | 'submitNightAction'
  | 'confirmNightAction'
  | 'acknowledgeNightAction'
  | 'skipNightAction'
  | 'prepareDawn'
  | 'confirmDawn'
  | 'resolveNight'
  | 'endGame'
  | 'publishGrimoire'
>;

export interface RoomSessionDependencies {
  readonly createClient: (endpoint: string) => SessionClient;
  readonly configuredEndpoint: string;
  readonly development: boolean;
  readonly getPlayerId: () => string;
  readonly getString: typeof getStoredString;
  readonly getIdentity: typeof getStoredRoomIdentity;
  readonly persistString: typeof persistString;
  readonly persistIdentity: typeof persistRoomIdentity;
  readonly removeValue: typeof removeStoredValue;
}

export interface RoomSessionState {
  readonly initialized: boolean;
  readonly status: ConnectionStatus;
  readonly experience: RoomExperienceState;
  readonly identity: StoredRoomIdentity | null;
  readonly playerId: string;
  readonly playerName: string;
  readonly endpoint: string;
  readonly maxPlayers: number;
  readonly inviteRoomId: string;
  readonly pendingCommand: PendingCommand | null;
  readonly errorMessage: string;
  readonly logs: readonly string[];
  readonly recoveryStatus: string | null;
  readonly recoveryRequests: NonNullable<ServerMessage['recoveryRequests']>;
  initialize(): void;
  connect(): void;
  disconnect(): void;
  setInviteRoomId(roomId: string): void;
  setPlayerName(name: string): void;
  setEndpoint(endpoint: string): void;
  setMaxPlayers(maxPlayers: number): void;
  clearError(): void;
  forgetRoomIdentity(): void;
  importRoomIdentity(identity: RecoveryIdentity): void;
  getRecoveryRequests(): void;
  reviewRecovery(requestId: string, decision: boolean): void;
  undoGame(confirmPhaseChange?: boolean): void;
  redoGame(confirmPhaseChange?: boolean): void;
  createRoom(): void;
  joinRoom(roomId: string): void;
  resumeLastRoom(): void;
  rejoinRoom(): void;
  leaveRoom(): void;
  closeRoom(): void;
  transferOwnership(playerId: string): void;
  restartGame(): void;
  kickPlayer(playerId: string): void;
  updateRoomSettings(maxPlayers: number): void;
  setStoryteller(playerId: string): void;
  setSeatOrder(seatOrder: readonly string[]): void;
  setReady(ready: boolean): void;
  confirmCharacter(): void;
  assignCharacters(
    assignments: Record<string, string>,
    shownCharacters?: Record<string, string>,
    fortuneTellerRedHerringId?: string,
    demonBluffCharacterIds?: readonly string[],
  ): void;
  startGame(): void;
  changePhase(phase: string): void;
  finalizeDay(): void;
  nominate(playerId: string): void;
  castVote(decision: boolean): void;
  recordVote(playerId: string, decision: boolean): void;
  advanceNominationStage(): void;
  controlNominationTimer(action: 'pause' | 'resume' | 'restart'): void;
  expireNominationTimer(deadlineUnixMs: number): void;
  resolveNomination(): void;
  executePlayer(playerId: string): void;
  useSlayerAbility(playerId: string): void;
  killPlayer(playerId: string, cause: string): void;
  submitNightAction(actionType: string, targetIds: readonly string[], result?: string): void;
  confirmNightAction(targetIds: readonly string[], result?: string): void;
  acknowledgeNightAction(): void;
  skipNightAction(): void;
  prepareDawn(): void;
  confirmDawn(deathPlayerIds: readonly string[]): void;
  resolveNight(): void;
  endGame(winner: 'good' | 'evil', description?: string): void;
  publishGrimoire(): void;
}

const configuredEndpoint = typeof __CLOCKTOWER_WS_URL__ === 'string' ? __CLOCKTOWER_WS_URL__ : '';
const developmentBuild = typeof __CLOCKTOWER_DEV__ === 'boolean' ? __CLOCKTOWER_DEV__ : true;

const defaultDependencies: RoomSessionDependencies = {
  createClient: (endpoint) => new GameWebSocketClient({
    url: endpoint,
    reconnectInterval: 2000,
    maxReconnectAttempts: 8,
    transportFactory: createTaroWebSocketTransport,
  }),
  configuredEndpoint,
  development: developmentBuild,
  getPlayerId: getOrCreatePlayerId,
  getString: getStoredString,
  getIdentity: getStoredRoomIdentity,
  persistString,
  persistIdentity: persistRoomIdentity,
  removeValue: removeStoredValue,
};

function toStoredIdentity(identity: ClientRoomIdentity): StoredRoomIdentity {
  return { version: 2, ...identity };
}

export function createRoomSessionStore(
  dependencies: RoomSessionDependencies = defaultDependencies,
): StoreApi<RoomSessionState> {
  let client: SessionClient | null = null;
  let initialResumeRequested = false;

  return createStore<RoomSessionState>((set, get) => {
    const appendLog = (message: string): void => {
      set((state) => ({ logs: [message, ...state.logs].slice(0, 30) }));
    };

    const clearIdentity = (): void => {
      dependencies.removeValue(STORAGE_KEYS.roomIdentity);
      dependencies.removeValue(STORAGE_KEYS.legacyRoomId);
      set({ identity: null, recoveryRequests: [] });
    };

    const persistClientIdentity = (): void => {
      const current = client?.identity;
      if (!current) return;
      const identity = toStoredIdentity(current);
      dependencies.persistIdentity(STORAGE_KEYS.roomIdentity, identity);
      dependencies.persistString(STORAGE_KEYS.legacyRoomId, identity.roomId);
      set({ identity, playerId: identity.playerId });
    };

    const handleMessage = (message: ServerMessage): void => {
      const previous = get();
      const previousPhase = selectGamePhase(previous.experience);
      let experience = applyServerMessage(previous.experience, message);
      let errorMessage = '';
      if (message.type === 'RECOVERY_REQUESTS') set({ recoveryRequests: message.recoveryRequests ?? [] });
      if (message.type === 'RECOVERY_STATUS') {
        set({ recoveryStatus: message.recoveryStatus ?? null });
        if (message.recoveryStatus === 'approved' || message.recoveryStatus === 'rejected' || message.recoveryStatus === 'expired') dependencies.removeValue(STORAGE_KEYS.recoveryTicket);
        if (message.recoveryStatus === 'approved') persistClientIdentity();
      }
      if (message.type === 'SESSION_REPLACED') {
        clearIdentity();
        client?.disconnect();
        initialResumeRequested = false;
        set({ experience: clearRoomIdentityState(), pendingCommand: null, recoveryStatus: null, errorMessage: '身份已在另一设备经批准恢复，此设备已退出。' });
        return;
      }

      if (
        message.type === 'CREATE_ROOM_RESULT' ||
        message.type === 'JOIN_ROOM_RESULT' ||
        message.type === 'RESUME_ROOM_RESULT'
      ) {
        persistClientIdentity();
        try { client?.getRecoveryRequests(); } catch { /* The next explicit refresh can retry. */ }
      }

      if (message.type === 'ERROR') {
        if (previous.recoveryStatus === 'pending') set({ recoveryStatus: 'rejected' });
        errorMessage = protocolErrorMessage(message.error, message.code);
        if (message.code === 'INVALID_CREDENTIAL' || message.code === 'ROOM_NOT_FOUND') {
          clearIdentity();
        }
        if (message.code === 'UNEXPECTED_SEQUENCE' || message.code === 'SEQUENCE_CONFLICT') {
          try {
            client?.getRoomState();
          } catch {
            errorMessage = '状态同步失败，请等待连接恢复';
          }
        }
      }

      if (message.type === 'KICKED' || message.type === 'ROOM_CLOSED') {
        clearIdentity();
        errorMessage = message.type === 'KICKED' ? '你已被房主移出房间' : '房间已关闭或因长时间不活跃而过期';
      }

      if (
        message.type === 'COMMAND_RESULT' &&
        previous.pendingCommand === 'leave' &&
        previousPhase === 'finished'
      ) {
        clearIdentity();
        experience = clearRoomIdentityState();
      }

      set({
        experience,
        errorMessage,
        pendingCommand: client?.hasPendingRequest ? previous.pendingCommand : null,
      });
      appendLog(message.type);
    };

    const requireClient = (): SessionClient | null => {
      if (!client || client.status !== 'connected') {
        set({ errorMessage: '正在连接服务器，请稍后重试' });
        return null;
      }
      return client;
    };

    const send = (pendingCommand: PendingCommand, operation: (activeClient: SessionClient) => void): void => {
      const activeClient = requireClient();
      if (!activeClient) return;
      set({ pendingCommand, errorMessage: '' });
      try {
        operation(activeClient);
      } catch (error) {
        set({
          pendingCommand: null,
          errorMessage: error instanceof Error ? error.message : '操作发送失败',
        });
      }
    };

    const connect = (): void => {
      const endpoint = get().endpoint.trim();
      if (!endpoint) {
        set({ status: 'error', errorMessage: '生产服务地址尚未配置' });
        return;
      }
      if (client) {
        client.connect();
        return;
      }

      client = dependencies.createClient(endpoint);
      client.onMessage(handleMessage);
      client.onStatusChange((status) => {
        set({ status });
      });
      const identity = get().identity;
      if (identity && !initialResumeRequested) {
        initialResumeRequested = true;
        set({ pendingCommand: 'resume' });
        client.resumeRoom(identity.roomId, identity.playerId, identity.resumeCredential);
      }
      client.connect();
    };

    return {
      initialized: false,
      status: 'disconnected',
      experience: createRoomExperienceState(),
      identity: null,
      playerId: '',
      playerName: '',
      endpoint: '',
      maxPlayers: 5,
      inviteRoomId: '',
      pendingCommand: null,
      errorMessage: '',
      logs: [],
      recoveryStatus: null,
      recoveryRequests: [],

      initialize(): void {
        if (get().initialized) return;
        const generatedPlayerId = dependencies.getPlayerId();
        const identity = dependencies.getIdentity(STORAGE_KEYS.roomIdentity);
        const playerId = identity?.playerId ?? generatedPlayerId;
        const playerName = dependencies.getString(
          STORAGE_KEYS.playerName,
          `玩家${playerId.slice(-4)}`,
        );
        const storedEndpoint = dependencies.development
          ? dependencies.getString(STORAGE_KEYS.endpoint, DEVELOPMENT_WS_URL)
          : '';
        const endpoint = dependencies.configuredEndpoint.trim() || storedEndpoint;
        const storedMaxPlayers = Number.parseInt(
          dependencies.getString(STORAGE_KEYS.maxPlayers, '5'),
          10,
        );
        if (!identity && dependencies.getString(STORAGE_KEYS.legacyRoomId)) {
          dependencies.removeValue(STORAGE_KEYS.legacyRoomId);
        }
        set({
          initialized: true,
          identity,
          playerId,
          playerName,
          endpoint,
          maxPlayers: Number.isInteger(storedMaxPlayers) ? storedMaxPlayers : 5,
        });
        connect();
      },
      connect,
      disconnect(): void {
        client?.disconnect();
        client = null;
        initialResumeRequested = false;
        set({ status: 'disconnected', pendingCommand: null });
      },
      setInviteRoomId(roomId): void {
        set({ inviteRoomId: roomId.trim() });
      },
      setPlayerName(name): void {
        set({ playerName: name });
        dependencies.persistString(STORAGE_KEYS.playerName, name);
      },
      setEndpoint(endpoint): void {
        if (!dependencies.development) return;
        dependencies.persistString(STORAGE_KEYS.endpoint, endpoint);
        client?.disconnect();
        client = null;
        initialResumeRequested = false;
        set({ endpoint, status: 'disconnected' });
      },
      setMaxPlayers(maxPlayers): void {
        const value = Math.min(15, Math.max(5, Math.trunc(maxPlayers)));
        dependencies.persistString(STORAGE_KEYS.maxPlayers, String(value));
        set({ maxPlayers: value });
      },
      clearError(): void {
        set({ errorMessage: '' });
      },
      forgetRoomIdentity(): void {
        const identityStatus = get().experience.identityStatus;
        const nonRecoverableRetainedIdentity = identityStatus?.status === 'retained' && !identityStatus.canRejoin;
        if (!dependencies.development && !nonRecoverableRetainedIdentity) {
          set({ errorMessage: '生产环境不允许强制丢弃房间身份' });
          return;
        }
        clearIdentity();
        client?.disconnect();
        client = null;
        initialResumeRequested = false;
        set({ experience: clearRoomIdentityState(), pendingCommand: null });
        connect();
      },
      createRoom(): void {
        if (get().identity) {
          set({ errorMessage: '请先离开或放弃当前房间身份' });
          return;
        }
        const { playerId, playerName, maxPlayers } = get();
        send('create', (activeClient) => {
          activeClient.createRoom(playerId, playerName.trim() || `玩家${playerId.slice(-4)}`, maxPlayers, DEFAULT_SCRIPT_ID);
        });
      },
      importRoomIdentity(identity): void {
        if (get().identity) {
          set({ errorMessage: '请先离开当前房间，再导入恢复码' });
          return;
        }
        const activeClient = requireClient();
        if (!activeClient) return;
        set({ recoveryStatus: 'pending', pendingCommand: 'resume', errorMessage: '' });
        let previousRequestId: string | undefined;
        try {
          const ticket = JSON.parse(dependencies.getString(STORAGE_KEYS.recoveryTicket, '{}')) as { requestId?: unknown; roomId?: unknown; playerId?: unknown };
          if (ticket.roomId === identity.roomId && ticket.playerId === identity.playerId && typeof ticket.requestId === 'string' && ticket.requestId.length <= 128) previousRequestId = ticket.requestId;
        } catch { /* A damaged ticket is replaced without logging user data. */ }
        const requestId = previousRequestId
          ? activeClient.requestRecovery(identity.roomId, identity.playerId, identity.recoveryCredential, previousRequestId)
          : activeClient.requestRecovery(identity.roomId, identity.playerId, identity.recoveryCredential);
        dependencies.persistString(STORAGE_KEYS.recoveryTicket, JSON.stringify({ requestId, roomId: identity.roomId, playerId: identity.playerId }));
      },
      getRecoveryRequests(): void { const active = requireClient(); if (active && get().identity) active.getRecoveryRequests(); },
      reviewRecovery(requestId, decision): void { send('settings', active => active.reviewRecovery(requestId, decision)); },
      undoGame(confirmPhaseChange = false): void { send('settings', active => active.undoGame(confirmPhaseChange)); },
      redoGame(confirmPhaseChange = false): void { send('settings', active => active.redoGame(confirmPhaseChange)); },
      joinRoom(roomId): void {
        const targetRoomId = roomId.trim();
        if (!targetRoomId) {
          set({ errorMessage: '请输入房间号' });
          return;
        }
        const existingIdentity = get().identity;
        if (existingIdentity) {
          if (existingIdentity.roomId === targetRoomId) {
            get().resumeLastRoom();
          } else {
            set({ errorMessage: '当前设备保留了其他房间身份，请先放弃后再加入' });
          }
          return;
        }
        const { playerId, playerName } = get();
        send('join', (activeClient) => {
          activeClient.joinRoom(targetRoomId, playerId, playerName.trim() || `玩家${playerId.slice(-4)}`);
        });
      },
      resumeLastRoom(): void {
        const identity = get().identity;
        if (!identity) {
          set({ errorMessage: '没有可恢复的房间身份' });
          return;
        }
        initialResumeRequested = true;
        send('resume', (activeClient) => {
          activeClient.resumeRoom(identity.roomId, identity.playerId, identity.resumeCredential);
        });
      },
      rejoinRoom(): void {
        send('rejoin', (activeClient) => activeClient.rejoinRoom());
      },
      leaveRoom(): void {
        send('leave', (activeClient) => activeClient.leaveRoom());
      },
      closeRoom(): void {
        send('close', (activeClient) => activeClient.closeRoom());
      },
      transferOwnership(playerId): void {
        send('transfer-ownership', (activeClient) => activeClient.transferOwnership(playerId));
      },
      restartGame(): void {
        send('restart-game', (activeClient) => activeClient.restartGame());
      },
      kickPlayer(playerId): void {
        send('settings', (activeClient) => activeClient.kickPlayer(playerId));
      },
      updateRoomSettings(maxPlayers): void {
        get().setMaxPlayers(maxPlayers);
        send('settings', (activeClient) => activeClient.updateRoomSettings(maxPlayers));
      },
      setStoryteller(playerId): void {
        send('storyteller', (activeClient) => activeClient.setStoryteller(playerId));
      },
      setSeatOrder(seatOrder): void {
        send('settings', (activeClient) => activeClient.setSeatOrder(seatOrder));
      },
      setReady(ready): void {
        send('ready', (activeClient) => activeClient.setReady(ready));
      },
      confirmCharacter(): void {
        send('confirm-character', (activeClient) => activeClient.confirmCharacter());
      },
      assignCharacters(assignments, shownCharacters, fortuneTellerRedHerringId, demonBluffCharacterIds): void {
        send('assign', (activeClient) => {
          activeClient.assignCharacters(assignments, shownCharacters, fortuneTellerRedHerringId, demonBluffCharacterIds);
        });
      },
      startGame(): void {
        send('start', (activeClient) => activeClient.startGame());
      },
      changePhase(phase): void {
        send('phase', (activeClient) => activeClient.changePhase(phase));
      },
      finalizeDay(): void {
        send('finalize-day', (activeClient) => activeClient.finalizeDay());
      },
      nominate(playerId): void {
        send('nominate', (activeClient) => activeClient.nominate(playerId));
      },
      castVote(decision): void {
        send('vote', (activeClient) => activeClient.castVote(decision));
      },
      recordVote(playerId, decision): void {
        send('record-vote', (activeClient) => activeClient.recordVote(playerId, decision));
      },
      advanceNominationStage(): void { send('vote', client => client.advanceNominationStage()); },
      controlNominationTimer(action): void { send('vote', client => client.controlNominationTimer(action)); },
      expireNominationTimer(deadline): void { send('vote', client => client.expireNominationTimer(deadline)); },
      resolveNomination(): void {
        send('resolve-vote', (activeClient) => activeClient.resolveNomination());
      },
      executePlayer(playerId): void {
        send('resolve-vote', (activeClient) => activeClient.executePlayer(playerId));
      },
      useSlayerAbility(playerId): void {
        send('slayer', (activeClient) => activeClient.useSlayerAbility(playerId));
      },
      killPlayer(playerId, cause): void {
        send('death', (activeClient) => activeClient.killPlayer(playerId, cause));
      },
      submitNightAction(actionType, targetIds, result): void {
        send('night-action', (activeClient) => activeClient.submitNightAction(actionType, targetIds, result));
      },
      confirmNightAction(targetIds, result): void {
        send('confirm-night-action', (activeClient) => activeClient.confirmNightAction(targetIds, result));
      },
      acknowledgeNightAction(): void {
        send('acknowledge-night-action', (activeClient) => activeClient.acknowledgeNightAction());
      },
      skipNightAction(): void {
        send('skip-night-action', (activeClient) => activeClient.skipNightAction());
      },
      prepareDawn(): void {
        send('prepare-dawn', (activeClient) => activeClient.prepareDawn());
      },
      confirmDawn(deathPlayerIds): void {
        send('confirm-dawn', (activeClient) => activeClient.confirmDawn(deathPlayerIds));
      },
      resolveNight(): void {
        send('resolve-night', (activeClient) => activeClient.resolveNight());
      },
      endGame(winner, description): void {
        send('end-game', (activeClient) => activeClient.endGame(winner, description));
      },
      publishGrimoire(): void {
        send('publish-grimoire', (activeClient) => activeClient.publishGrimoire());
      },
    };
  });
}

export const roomSessionStore = createRoomSessionStore();

export function useRoomSession<T>(selector: (state: RoomSessionState) => T): T {
  return useStore(roomSessionStore, selector);
}

export function getRoomSessionState(): RoomSessionState {
  return roomSessionStore.getState();
}

export const isDevelopmentBuild = defaultDependencies.development;
