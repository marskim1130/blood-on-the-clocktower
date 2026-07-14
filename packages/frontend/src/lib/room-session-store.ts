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

export const STORAGE_KEYS = {
  playerName: 'clocktower.playerName',
  endpoint: 'clocktower.wsUrl',
  roomIdentity: 'clocktower.roomIdentity.v2',
  legacyRoomId: 'clocktower.lastRoomId',
  maxPlayers: 'clocktower.maxPlayers',
} as const;

export const DEFAULT_SCRIPT_ID = TROUBLE_BREWING_SCRIPT.id;
const DEVELOPMENT_WS_URL = 'ws://localhost:8080/ws';

type PendingCommand =
  | 'create'
  | 'join'
  | 'resume'
  | 'rejoin'
  | 'leave'
  | 'close'
  | 'settings'
  | 'storyteller'
  | 'assign'
  | 'start'
  | 'nominate'
  | 'vote'
  | 'resolve-vote'
  | 'phase'
  | 'slayer'
  | 'death'
  | 'night-action'
  | 'resolve-night'
  | 'end-game';

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
  | 'rejoinRoom'
  | 'getRoomState'
  | 'closeRoom'
  | 'leaveRoom'
  | 'kickPlayer'
  | 'updateRoomSettings'
  | 'setStoryteller'
  | 'assignCharacters'
  | 'startGame'
  | 'changePhase'
  | 'nominate'
  | 'castVote'
  | 'resolveNomination'
  | 'executePlayer'
  | 'useSlayerAbility'
  | 'killPlayer'
  | 'submitNightAction'
  | 'resolveNight'
  | 'endGame'
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
  initialize(): void;
  connect(): void;
  disconnect(): void;
  setInviteRoomId(roomId: string): void;
  setPlayerName(name: string): void;
  setEndpoint(endpoint: string): void;
  setMaxPlayers(maxPlayers: number): void;
  clearError(): void;
  forgetRoomIdentity(): void;
  createRoom(): void;
  joinRoom(roomId: string): void;
  resumeLastRoom(): void;
  rejoinRoom(): void;
  leaveRoom(): void;
  closeRoom(): void;
  kickPlayer(playerId: string): void;
  updateRoomSettings(maxPlayers: number): void;
  setStoryteller(playerId: string): void;
  assignCharacters(
    assignments: Record<string, string>,
    shownCharacters?: Record<string, string>,
    fortuneTellerRedHerringId?: string,
  ): void;
  startGame(): void;
  changePhase(phase: string): void;
  nominate(playerId: string): void;
  castVote(decision: boolean): void;
  resolveNomination(): void;
  executePlayer(playerId: string): void;
  useSlayerAbility(playerId: string): void;
  killPlayer(playerId: string, cause: string): void;
  submitNightAction(actionType: string, targetIds: readonly string[], result?: string): void;
  resolveNight(): void;
  endGame(winner: 'good' | 'evil', description?: string): void;
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
      set({ identity: null });
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

      if (
        message.type === 'CREATE_ROOM_RESULT' ||
        message.type === 'JOIN_ROOM_RESULT' ||
        message.type === 'RESUME_ROOM_RESULT'
      ) {
        persistClientIdentity();
      }

      if (message.type === 'ERROR') {
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
        errorMessage = message.type === 'KICKED' ? '你已被房主移出房间' : '房间已由房主关闭';
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
      assignCharacters(assignments, shownCharacters, fortuneTellerRedHerringId): void {
        send('assign', (activeClient) => {
          activeClient.assignCharacters(assignments, shownCharacters, fortuneTellerRedHerringId);
        });
      },
      startGame(): void {
        send('start', (activeClient) => activeClient.startGame());
      },
      changePhase(phase): void {
        send('phase', (activeClient) => activeClient.changePhase(phase));
      },
      nominate(playerId): void {
        send('nominate', (activeClient) => activeClient.nominate(playerId));
      },
      castVote(decision): void {
        send('vote', (activeClient) => activeClient.castVote(decision));
      },
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
      resolveNight(): void {
        send('resolve-night', (activeClient) => activeClient.resolveNight());
      },
      endGame(winner, description): void {
        send('end-game', (activeClient) => activeClient.endGame(winner, description));
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
