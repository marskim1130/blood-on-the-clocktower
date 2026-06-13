import { Button, Input, ScrollView, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  buildDefaultScriptAssignments,
  GameWebSocketClient,
  getScriptWakeOrder,
  TROUBLE_BREWING_SCRIPT,
} from '@clocktower/core';
import type { ConnectionStatus, GameCharacter, RoomState, ServerMessage } from '@clocktower/core';
import { createTaroWebSocketTransport } from '../../lib/taro-websocket-transport';
import {
  type GamePhase,
  type DeathCause,
  type InputEvent,
  mapProtocolPhase,
  normalizeWinner,
  eventValue,
  getStoredString,
  persistString,
  getOrCreatePlayerId,
} from '../../lib/utils';
import './index.css';

const PLAYER_ID_STORAGE_KEY = 'clocktower.playerId';
const PLAYER_NAME_STORAGE_KEY = 'clocktower.playerName';
const WS_URL_STORAGE_KEY = 'clocktower.wsUrl';
const LAST_ROOM_ID_STORAGE_KEY = 'clocktower.lastRoomId';
const MAX_PLAYERS_STORAGE_KEY = 'clocktower.maxPlayers';
const DEFAULT_WS_URL = 'ws://localhost:8080/ws';
const DEFAULT_SCRIPT_ID = TROUBLE_BREWING_SCRIPT.id;
const DEFAULT_SCRIPT_NAME = TROUBLE_BREWING_SCRIPT.name;

// ─── Phase & Role Constants ──────────────────────────────────────

const PHASE_LABELS: Record<GamePhase, string> = {
  setup: '准备',
  day: '白天',
  voting: '投票',
  night: '夜晚',
  finished: '游戏结束',
};

const STATUS_LABELS: Record<ConnectionStatus, string> = {
  disconnected: '未连接',
  connecting: '连接中',
  connected: '已连接',
  error: '连接异常',
};

const DEATH_CAUSE_LABELS: Record<DeathCause, string> = {
  execution: '处决',
  night_kill: '夜晚击杀',
  ability: '能力致死',
};

const DEATH_CAUSE_OPTIONS: ReadonlyArray<{ readonly id: DeathCause; readonly label: string }> = [
  { id: 'execution', label: '处决' },
  { id: 'night_kill', label: '夜杀' },
  { id: 'ability', label: '能力' },
];

// ─── Night Action Options ────────────────────────────────────────

const NIGHT_ACTION_TYPES = [
  { id: 'kill', label: '击杀', needsTarget: true },
  { id: 'poison', label: '下毒', needsTarget: true },
  { id: 'protect', label: '保护', needsTarget: true },
  { id: 'learn_demon', label: '确认恶魔信息', needsTarget: false },
  { id: 'learn_townsfolk', label: '确认镇民信息', needsTarget: false },
  { id: 'learn_outsider', label: '确认外来者信息', needsTarget: false },
  { id: 'learn_minion', label: '确认爪牙信息', needsTarget: false },
  { id: 'learn_evil_pairs', label: '邪恶相邻数量', needsTarget: false },
  { id: 'learn_evil_neighbors', label: '邪恶邻居数量', needsTarget: false },
  { id: 'check_demon', label: '查验恶魔', needsTarget: true },
  { id: 'learn_executed', label: '确认处决角色', needsTarget: false },
  { id: 'learn_died', label: '确认死亡触发', needsTarget: true },
  { id: 'learn_master', label: '选择主人', needsTarget: false },
] as const;

const CHARACTER_TYPE_LABELS: Record<string, string> = {
  townsfolk: '镇民',
  outsider: '外来者',
  minion: '爪牙',
  demon: '恶魔',
};

// ─── Local Game State Types ──────────────────────────────────────

interface NominationInfo {
  readonly nominatorId: string;
  readonly nomineeId: string;
  readonly votes: Record<string, boolean>;
}

interface DeathRecord {
  readonly cause: DeathCause;
  readonly dayNumber: number;
}

interface NightActionRecord {
  readonly actorId: string;
  readonly actionType: string;
  readonly targetIds: readonly string[];
  readonly result: string | null;
}

interface GameOverInfo {
  readonly winner: string;
  readonly reason: string;
  readonly description: string;
}

const SERVER_MESSAGE_LABELS: Readonly<Record<string, string>> = {
  ERROR: '错误',
  ROOM_STATE: '房间状态',
  EVENT: '游戏事件',
};

const SERVER_ERROR_LABELS: Readonly<Record<string, string>> = {
  'room not found': '房间不存在',
  'player was kicked from room': '你已被房主移出房间',
  'kicked from room': '已被房主移出房间',
  'room is full': '房间人数已满',
  'unknown command type': '未知操作类型',
  'storyteller already set': '说书人已设置',
  'target player not found': '目标玩家不存在',
  'only storyteller can assign characters': '只有说书人可以分配角色',
  'invalid character assignment for player count': '角色分配与当前玩家人数不匹配',
  'raw event submission is disabled; use explicit game commands': '不能直接提交原始事件，请使用明确的游戏操作',
  'players can only be kicked during setup phase': '只能在准备阶段踢出玩家',
  'target player is required': '请选择目标玩家',
  'room creator cannot kick themselves': '房主不能踢出自己',
  'room settings can only be updated during setup phase': '只能在准备阶段修改房间设置',
  'at least one room setting is required': '至少需要提供一项房间设置',
  'maxPlayers must be between 5 and 15': '实际玩家数必须在 5 到 15 之间',
  'maxPlayers cannot be less than current player count': '实际玩家数不能小于当前房间人数',
  'unsupported script': '暂不支持该剧本',
  'script cannot be changed after characters are assigned': '角色分配后不能更换剧本',
  'storyteller must be set before starting the game': '开始游戏前必须设置说书人',
  'only the storyteller can start the game': '只有说书人可以开始游戏',
  'only the storyteller can change phase': '只有说书人可以切换阶段',
  'nominations can only happen during the day phase': '只能在白天阶段发起提名',
  'cannot nominate yourself': '不能提名自己',
  'dead players cannot nominate': '死亡玩家不能提名',
  'cannot nominate a dead player': '不能提名死亡玩家',
  'no active nomination to vote on': '当前没有可投票的提名',
  'ghost vote already used': '幽灵票已经使用',
  'only the storyteller can resolve a nomination': '只有说书人可以结算提名',
  'no active nomination to resolve': '当前没有可结算的提名',
  'only the storyteller can execute a player': '只有说书人可以处决玩家',
  'Slayer ability can only be used during the day phase': '杀手能力只能在白天阶段使用',
  'dead players cannot use the Slayer ability': '死亡玩家不能使用杀手能力',
  'only the Slayer can use this ability': '只有杀手可以使用该能力',
  'Slayer ability already used': '杀手能力已经使用过',
  'cannot target a dead player': '不能选择死亡玩家作为目标',
  'night actions can only be submitted during the night phase': '只能在夜晚阶段提交夜间行动',
  'dead players cannot submit night actions': '死亡玩家不能提交夜间行动',
  'no remaining night wake steps': '今夜没有剩余唤醒步骤',
  'only the storyteller can resolve the night': '只有说书人可以结束夜晚',
  'night can only be resolved during the night phase': '只能在夜晚阶段结束夜晚',
  'only the storyteller can end the game': '只有说书人可以结束游戏',
  'game cannot be ended before it starts': '游戏开始前不能结束游戏',
  'game is already finished': '游戏已经结束',
  'winner must be good or evil': '胜利阵营必须是善良或邪恶',
  'only the storyteller can kill players': '只有说书人可以宣告玩家死亡',
  'game cannot kill players before it starts': '游戏开始前不能宣告死亡',
  'death cause is required': '请选择死因',
};

function serverMessageLabel(type: string): string {
  return SERVER_MESSAGE_LABELS[type] ?? '服务器消息';
}

function serverErrorMessage(error: string | undefined): string {
  if (!error) return '未知错误';
  const exact = SERVER_ERROR_LABELS[error];
  if (exact) return exact;
  const playerNotFound = error.match(/^player (.+) not found(?: in game)?$/);
  if (playerNotFound) return `玩家不存在：${playerNotFound[1]}`;
  const playerAlreadyVoted = error.match(/^player (.+) has already voted$/);
  if (playerAlreadyVoted) return `玩家已投票：${playerAlreadyVoted[1]}`;
  const playerAlreadyDead = error.match(/^player (.+) is already dead$/);
  if (playerAlreadyDead) return `玩家已死亡：${playerAlreadyDead[1]}`;
  const characterMissing = error.match(/^player (.+) has no character assigned$/);
  if (characterMissing) return `玩家尚未分配角色：${characterMissing[1]}`;
  const nightTargetMissing = error.match(/^night action target (.+) not found$/);
  if (nightTargetMissing) return `夜间行动目标不存在：${nightTargetMissing[1]}`;
  const duplicateNightTarget = error.match(/^duplicate night action target (.+)$/);
  if (duplicateNightTarget) return `夜间行动目标重复：${duplicateNightTarget[1]}`;
  const nightActionMinimum = error.match(/^night action (.+) requires at least (.+) target\(s\)$/);
  if (nightActionMinimum) return `夜间行动${nightActionLabel(nightActionMinimum[1] ?? '')}至少需要 ${nightActionMinimum[2]} 个目标`;
  const nightActionMaximum = error.match(/^night action (.+) allows at most (.+) target\(s\)$/);
  if (nightActionMaximum) return `夜间行动${nightActionLabel(nightActionMaximum[1] ?? '')}最多允许 ${nightActionMaximum[2]} 个目标`;
  const expectedNightAction = error.match(/^expected night action (.+), got (.+)$/);
  if (expectedNightAction) {
    return `当前步骤需要${nightActionLabel(expectedNightAction[1] ?? '')}，收到的是${nightActionLabel(expectedNightAction[2] ?? '')}`;
  }
  const remainingWakeSteps = error.match(/^cannot resolve night with (.+) wake step\(s\) remaining$/);
  if (remainingWakeSteps) return `还有 ${remainingWakeSteps[1]} 个唤醒步骤未处理，不能结束夜晚`;
  const unsupportedDeathCause = error.match(/^unsupported death cause (.+)$/);
  if (unsupportedDeathCause) return `不支持的死因：${unsupportedDeathCause[1]}`;
  const invalidPhaseTransition = error.match(/^invalid phase transition from (.+) to (.+)$/);
  if (invalidPhaseTransition) return '当前阶段不能这样切换';
  const cannotChangePhase = error.match(/^cannot change phase from (.+)$/);
  if (cannotChangePhase) return '当前阶段不能切换';
  const cannotExecute = error.match(/^cannot execute player in phase (.+)$/);
  if (cannotExecute) return '当前阶段不能处决玩家';
  return error;
}

function nightActionLabel(actionType: string): string {
  return NIGHT_ACTION_TYPES.find((action) => action.id === actionType)?.label ?? actionType;
}

function buildSampleAssignments(players: RoomState['players']): Record<string, string> {
  return buildDefaultScriptAssignments(players, DEFAULT_SCRIPT_ID);
}

function chooseFortuneTellerRedHerring(assignments: Record<string, string>): string | undefined {
  if (!Object.values(assignments).includes('fortuneteller')) {
    return undefined;
  }
  const charactersByID = new Map(TROUBLE_BREWING_SCRIPT.characters.map((character) => [character.id, character]));
  for (const [playerID, characterID] of Object.entries(assignments)) {
    const character = charactersByID.get(characterID);
    if (characterID !== 'fortuneteller' && character?.team === 'good') {
      return playerID;
    }
  }
  return undefined;
}

export default function IndexPage() {
  const clientRef = useRef<GameWebSocketClient | null>(null);
  const [wsUrl, setWsUrl] = useState(DEFAULT_WS_URL);
  const [playerId, setPlayerId] = useState('');
  const [playerName, setPlayerName] = useState('');
  const [roomIdInput, setRoomIdInput] = useState('');
  const [maxPlayersInput, setMaxPlayersInput] = useState('5');
  const [status, setStatus] = useState<ConnectionStatus>('disconnected');
  const [roomState, setRoomState] = useState<RoomState | null>(null);
  const [myCharacter, setMyCharacter] = useState<GameCharacter | null>(null);
  const [errorMessage, setErrorMessage] = useState('');
  const [logs, setLogs] = useState<readonly string[]>([]);

  // ─── Game Phase State ────────────────────────────────────────
  const [gamePhase, setGamePhase] = useState<GamePhase>('setup');
  const [dayNumber, setDayNumber] = useState(0);

  // ─── Voting State ────────────────────────────────────────────
  const [currentNomination, setCurrentNomination] = useState<NominationInfo | null>(null);
  const [nomineeIdInput, setNomineeIdInput] = useState('');
  const [lastNominationResult, setLastNominationResult] = useState<{
    readonly nomineeId: string;
    readonly executed: boolean;
    readonly yesVotes: number;
    readonly noVotes: number;
    readonly requiredVotes: number | null;
  } | null>(null);

  // ─── Death State ─────────────────────────────────────────────
  const [deathRecords, setDeathRecords] = useState<Record<string, DeathRecord>>({});
  const [ghostVotesRemaining, setGhostVotesRemaining] = useState<Set<string>>(new Set());
  const [deathAnnouncements, setDeathAnnouncements] = useState<readonly string[]>([]);
  const [manualDeathTargetId, setManualDeathTargetId] = useState('');
  const [manualDeathCause, setManualDeathCause] = useState<DeathCause>('execution');

  // ─── Night State ─────────────────────────────────────────────
  const [nightActions, setNightActions] = useState<readonly NightActionRecord[]>([]);
  const [nightActionType, setNightActionType] = useState<string>(NIGHT_ACTION_TYPES[0].id);
  const [nightTargetIds, setNightTargetIds] = useState<readonly string[]>([]);
  const [nightResultInput, setNightResultInput] = useState('');

  // ─── Win Condition State ─────────────────────────────────────
  const [gameOver, setGameOver] = useState<GameOverInfo | null>(null);
  const [endGameDescriptionInput, setEndGameDescriptionInput] = useState('');

  useEffect(() => {
    const id = getOrCreatePlayerId();
    const defaultName = `玩家${id.slice(-4)}`;
    setPlayerId(id);
    setPlayerName(getStoredString(PLAYER_NAME_STORAGE_KEY, defaultName));
    setWsUrl(getStoredString(WS_URL_STORAGE_KEY, DEFAULT_WS_URL));
    setRoomIdInput(getStoredString(LAST_ROOM_ID_STORAGE_KEY));
    setMaxPlayersInput(getStoredString(MAX_PLAYERS_STORAGE_KEY, '5'));

    return () => {
      clientRef.current?.disconnect();
      clientRef.current = null;
    };
  }, []);

  const isStoryteller = roomState?.storytellerId === playerId;
  const isRoomCreator = roomState?.creatorId === playerId;
  const selfPlayer = useMemo(
    () => roomState?.players.find((player) => player.id === playerId) ?? null,
    [playerId, roomState],
  );
  const visibleCharacter = useMemo(() => {
    if (myCharacter) return myCharacter;
    return selfPlayer?.character ?? null;
  }, [myCharacter, selfPlayer]);

  const alivePlayers = useMemo(
    () => (roomState?.players ?? []).filter((player) => player.isAlive),
    [roomState],
  );

  const deadPlayers = useMemo(
    () => (roomState?.players ?? []).filter((player) => !player.isAlive),
    [roomState],
  );

  const selectedNightAction = useMemo(
    () => NIGHT_ACTION_TYPES.find((a) => a.id === nightActionType) ?? NIGHT_ACTION_TYPES[0],
    [nightActionType],
  );
  const localScriptWakeSteps = useMemo(
    () => getScriptWakeOrder(DEFAULT_SCRIPT_ID, Math.max(1, dayNumber)),
    [dayNumber],
  );
  const nightWakeSteps = useMemo(
    () => (roomState?.nightWakeSteps && roomState.nightWakeSteps.length > 0 ? roomState.nightWakeSteps : localScriptWakeSteps),
    [localScriptWakeSteps, roomState],
  );
  const currentNightWakeStep = roomState?.currentNightWakeStep ?? null;
  const currentNightWakeIndex = roomState?.currentNightWakeIndex ?? 0;
  const selectedNightMinTargets = currentNightWakeStep?.actionType === nightActionType ? currentNightWakeStep.minTargets : 0;
  const selectedNightMaxTargets = currentNightWakeStep?.actionType === nightActionType
    ? currentNightWakeStep.maxTargets
    : selectedNightAction.needsTarget
      ? 1
      : 0;
  const selectedNightNeedsTarget = selectedNightMaxTargets > 0;
  const assignedCharacterIds = useMemo(
    () =>
      new Set(
        (roomState?.players ?? [])
          .map((player) => player.character?.id)
          .filter((characterId): characterId is string => Boolean(characterId)),
      ),
    [roomState],
  );
  const assignedCharacterTypes = useMemo(
    () =>
      new Set<string>(
        (roomState?.players ?? [])
          .map((player) => {
            const characterId = player.character?.id;
            return TROUBLE_BREWING_SCRIPT.characters.find((item) => item.id === characterId)?.type;
          })
          .filter((characterType): characterType is NonNullable<typeof characterType> => Boolean(characterType)),
      ),
    [roomState],
  );
  const currentRoomId = roomState?.roomId ?? roomIdInput.trim();
  const currentScriptName = roomState?.scriptName ?? DEFAULT_SCRIPT_NAME;
  const currentExecutionThreshold = Math.ceil(alivePlayers.length / 2);
  const canUseSlayerAbility = gamePhase === 'day' && selfPlayer?.isAlive === true && visibleCharacter?.id === 'slayer';
  const players = roomState?.players ?? [];
  const storytellerName = roomState?.storytellerId
    ? players.find((player) => player.id === roomState.storytellerId)?.name ?? roomState.storytellerId
    : '未设置';
  const roomCapacity = roomState?.maxPlayers ?? (Number.parseInt(maxPlayersInput, 10) || 5);
  const roomOccupancy = roomState ? `${players.length}/${roomCapacity}` : '未加入';
  const roomCodeLabel = currentRoomId || '未加入';
  const roleLabel = isStoryteller ? '说书人' : visibleCharacter?.name ?? '未分配';
  const aliveSummary = roomState ? `${alivePlayers.length} 存活 / ${deadPlayers.length} 死亡` : '等待加入房间';
  const latestLog = logs[0] ?? '暂无操作记录';
  const statusIcon = status === 'connected' ? '●' : status === 'connecting' ? '◐' : status === 'error' ? '!' : '○';

  function toggleNightTarget(targetId: string): void {
    setNightTargetIds((current) =>
      current.includes(targetId)
        ? current.filter((id) => id !== targetId)
        : [...current, targetId],
    );
  }

  function appendLog(message: string): void {
    setLogs((previous) => [message, ...previous].slice(0, 30));
  }

  function updateWsUrl(value: string): void {
    setWsUrl(value);
    persistString(WS_URL_STORAGE_KEY, value);
  }

  function updatePlayerName(value: string): void {
    setPlayerName(value);
    persistString(PLAYER_NAME_STORAGE_KEY, value);
  }

  function updateRoomIdInput(value: string): void {
    setRoomIdInput(value);
    persistString(LAST_ROOM_ID_STORAGE_KEY, value.trim());
  }

  function updateMaxPlayersInput(value: string): void {
    setMaxPlayersInput(value);
    persistString(MAX_PLAYERS_STORAGE_KEY, value);
  }

  function syncRoomSnapshot(state: RoomState): void {
    setRoomState(state);
    if (typeof state.maxPlayers === 'number') {
      setMaxPlayersInput(String(state.maxPlayers));
      persistString(MAX_PLAYERS_STORAGE_KEY, String(state.maxPlayers));
    }

    const snapshotPhase = mapProtocolPhase(state.phase);
    if (snapshotPhase) {
      setGamePhase(snapshotPhase);
    }
    if (typeof state.dayNumber === 'number') {
      setDayNumber(state.dayNumber);
    }

    if (state.nomination) {
      setCurrentNomination({
        nominatorId: state.nomination.nominatorId,
        nomineeId: state.nomination.nomineeId,
        votes: state.nomination.votes ?? {},
      });
    } else if (snapshotPhase !== 'voting') {
      setCurrentNomination(null);
    }

    const nextDeathRecords = (state.deaths ?? []).reduce<Record<string, DeathRecord>>((records, death) => {
      records[death.playerId] = {
        cause: death.cause as DeathCause,
        dayNumber: death.dayNumber,
      };
      return records;
    }, {});
    setDeathRecords(nextDeathRecords);
    setGhostVotesRemaining(new Set(state.ghostVotesRemaining ?? []));
    if (state.currentNightWakeStep) {
      setNightActionType(state.currentNightWakeStep.actionType);
      setNightTargetIds([]);
    }

    const self = state.players.find((player) => player.id === playerId);
    setMyCharacter(self?.character ?? null);

    if (state.winner) {
      setGameOver({
        winner: normalizeWinner(state.winner.winner),
        reason: state.winner.reason,
        description: state.winner.description,
      });
      setGamePhase('finished');
    } else if (snapshotPhase !== 'finished') {
      setGameOver(null);
    }
  }

  function applyMessage(message: ServerMessage): void {
    appendLog(`收到${serverMessageLabel(message.type)}`);

    if (message.type === 'ERROR') {
      setErrorMessage(serverErrorMessage(message.error));
      if (message.error === 'kicked from room') {
        setRoomState(null);
        setMyCharacter(null);
        updateRoomIdInput('');
        appendLog('已被房主移出房间');
      }
      return;
    }

    if (message.type === 'ROOM_STATE') {
      if (message.roomId) updateRoomIdInput(message.roomId);
      if (message.state) syncRoomSnapshot(message.state);
      return;
    }

    if (!message.event) return;

    const event = message.event;

    // ─── Player Joined ───────────────────────────────────────
    if ('playerJoined' in event) {
      const joinedPlayer = (event as { readonly playerJoined: { readonly player: { readonly id: string; readonly name: string; readonly isAlive: boolean } } }).playerJoined.player;
      setRoomState((current) => {
        if (!current || current.players.some((player) => player.id === joinedPlayer.id)) return current;
        return { ...current, players: [...current.players, joinedPlayer] };
      });
      return;
    }

    // ─── Player Left ─────────────────────────────────────────
    if ('playerLeft' in event) {
      const leftPlayerId = (event as { readonly playerLeft: { readonly playerId: string } }).playerLeft.playerId;
      setRoomState((current) => {
        if (!current) return current;
        return { ...current, players: current.players.filter((player) => player.id !== leftPlayerId) };
      });
      return;
    }

    // ─── Character Assigned ──────────────────────────────────
    if ('characterAssigned' in event) {
      const assignment = (event as {
        readonly characterAssigned: {
          readonly playerId: string;
          readonly character: GameCharacter;
          readonly shownCharacter?: GameCharacter | null;
        };
      }).characterAssigned;
      if (assignment.playerId === playerId) {
        setMyCharacter(assignment.character);
      }
      // Also update the roomState player character for storyteller view
      setRoomState((current) => {
        if (!current) return current;
        return {
          ...current,
          players: current.players.map((player) =>
            player.id === assignment.playerId
              ? { ...player, character: assignment.character, shownCharacter: assignment.shownCharacter ?? null }
              : player,
          ),
        };
      });
      return;
    }

    // ─── Phase Changed ───────────────────────────────────────
    if ('phaseChanged' in event) {
      const phaseValue = (event as { readonly phaseChanged: { readonly phase: number } }).phaseChanged.phase;
      const mapped = mapProtocolPhase(phaseValue);
      if (mapped) {
        setGamePhase(mapped);
        if (mapped === 'day') {
          setDayNumber((current) => current + 1);
        }
        if (mapped === 'night') {
          setNightActions([]);
          setNightTargetIds([]);
        }
        if (mapped !== 'voting') {
          setCurrentNomination(null);
        }
      }
      appendLog(`阶段切换为${mapped ? PHASE_LABELS[mapped] : phaseValue}`);
      return;
    }

    // ─── Player Died ─────────────────────────────────────────
    if ('playerDied' in event) {
      const deathEvent = (event as { readonly playerDied: { readonly playerId: string; readonly cause: string; readonly dayNumber: number } }).playerDied;
      const deadPlayerId = deathEvent.playerId;
      const cause = deathEvent.cause;
      const deathDay = deathEvent.dayNumber;
      setDeathRecords((current) => ({
        ...current,
        [deadPlayerId]: { cause: cause as DeathCause, dayNumber: deathDay },
      }));
      setGhostVotesRemaining((current) => {
        const next = new Set(current);
        next.add(deadPlayerId);
        return next;
      });
      setDeathAnnouncements((current) => {
        const label = DEATH_CAUSE_LABELS[cause as DeathCause] ?? cause;
        const player = roomState?.players.find((player) => player.id === deadPlayerId);
        const name = player?.name ?? deadPlayerId;
        return [`第 ${deathDay} 天：${name} 死亡（${label}）`, ...current].slice(0, 20);
      });
      // Update roomState isAlive
      setRoomState((current) => {
        if (!current) return current;
        return {
          ...current,
          players: current.players.map((player) =>
            player.id === deadPlayerId ? { ...player, isAlive: false } : player,
          ),
        };
      });
      appendLog(`玩家死亡：${deadPlayerId}（${DEATH_CAUSE_LABELS[cause as DeathCause] ?? cause}）`);
      return;
    }

    // ─── Nomination Started ──────────────────────────────────
    if ('nominationStarted' in event) {
      const nomination = (event as { readonly nominationStarted: { readonly nominatorId: string; readonly nomineeId: string } }).nominationStarted;
      setCurrentNomination({ nominatorId: nomination.nominatorId, nomineeId: nomination.nomineeId, votes: {} });
      setLastNominationResult(null);
      setGamePhase('voting');
      appendLog(`提名：${nomination.nominatorId} 提名 ${nomination.nomineeId}`);
      return;
    }

    // ─── Vote Cast ───────────────────────────────────────────
    if ('voteCast' in event) {
      const voteData = (event as { readonly voteCast: { readonly voterId: string; readonly targetId?: string; readonly decision?: boolean } }).voteCast;
      const decision = typeof voteData.decision === 'boolean'
        ? voteData.decision
        : voteData.targetId !== undefined && voteData.targetId !== '';
      setCurrentNomination((current) => {
        if (!current) return current;
        return {
          ...current,
          votes: { ...current.votes, [voteData.voterId]: decision },
        };
      });
      return;
    }

    // ─── Nomination Resolved ─────────────────────────────────
    if ('nominationResolved' in event) {
      const resolved = (event as { readonly nominationResolved: { readonly nomineeId: string; readonly executed: boolean; readonly yesVotes: number; readonly noVotes: number; readonly requiredVotes?: number } }).nominationResolved;
      setLastNominationResult({
        nomineeId: resolved.nomineeId,
        executed: resolved.executed,
        yesVotes: resolved.yesVotes,
        noVotes: resolved.noVotes,
        requiredVotes: resolved.requiredVotes ?? null,
      });
      setCurrentNomination(null);
      const name = roomState?.players.find((player) => player.id === resolved.nomineeId)?.name ?? resolved.nomineeId;
      appendLog(`投票结果: ${name} ${resolved.executed ? '被处决' : '幸存'} (${resolved.yesVotes}/${resolved.noVotes})`);
      return;
    }

    // ─── Night Action ────────────────────────────────────────
    if ('nightAction' in event || 'nightActionSubmitted' in event) {
      const action = 'nightAction' in event
        ? (event as { readonly nightAction: { readonly actorId: string; readonly actionType: string; readonly targetIds: readonly string[]; readonly result: string | null } }).nightAction
        : (event as { readonly nightActionSubmitted: { readonly actorId: string; readonly actionType: string; readonly targetIds: readonly string[]; readonly result?: string | null } }).nightActionSubmitted;
      setNightActions((current) => [...current, { actorId: action.actorId, actionType: action.actionType, targetIds: action.targetIds, result: action.result ?? null }]);
      appendLog(`夜间行动: ${action.actionType} (${action.actorId})`);
      return;
    }

    // ─── Game Over ───────────────────────────────────────────
    if ('gameOver' in event || 'gameEnded' in event) {
      const gameOverEvent = 'gameOver' in event
        ? (event as { readonly gameOver: { readonly winner: number | string; readonly reason: string; readonly description: string } }).gameOver
        : (event as { readonly gameEnded: { readonly winner: number | string; readonly reason: string; readonly description: string } }).gameEnded;
      setGameOver({ winner: normalizeWinner(gameOverEvent.winner), reason: gameOverEvent.reason, description: gameOverEvent.description });
      setGamePhase('finished');
      appendLog(`游戏结束: ${gameOverEvent.winner} 胜利 - ${gameOverEvent.description}`);
      return;
    }
  }

  function connect(onConnected?: (client: GameWebSocketClient) => void): void {
    setErrorMessage('');
    clientRef.current?.disconnect();
    persistString(WS_URL_STORAGE_KEY, wsUrl.trim());

    const client = new GameWebSocketClient({
      url: wsUrl.trim(),
      reconnectInterval: 2000,
      maxReconnectAttempts: 8,
      transportFactory: createTaroWebSocketTransport,
    });
    let handledConnected = false;
    client.onStatusChange((nextStatus) => {
      setStatus(nextStatus);
      if (nextStatus === 'connected' && onConnected && !handledConnected) {
        handledConnected = true;
        onConnected(client);
      }
    });
    client.onMessage(applyMessage);
    client.connect();
    clientRef.current = client;
    appendLog(`正在连接服务：${wsUrl.trim()}`);
  }

  function disconnect(): void {
    clientRef.current?.disconnect();
    clientRef.current = null;
    setStatus('disconnected');
    appendLog('已断开连接');
  }

  function requireClient(): GameWebSocketClient | null {
    const client = clientRef.current;
    if (!client || client.status !== 'connected') {
      setErrorMessage('请先连接服务');
      return null;
    }
    return client;
  }

  function createRoom(): void {
    const client = requireClient();
    if (!client) return;

    const maxPlayers = Number.parseInt(maxPlayersInput, 10);
    const displayName = playerName.trim() || playerId;
    persistString(PLAYER_NAME_STORAGE_KEY, displayName);
    persistString(MAX_PLAYERS_STORAGE_KEY, maxPlayersInput);
    client.createRoom(playerId, displayName, Number.isNaN(maxPlayers) ? 5 : maxPlayers, DEFAULT_SCRIPT_ID);
    setMyCharacter(null);
    appendLog('已发送创建房间请求');
  }

  function joinRoom(): void {
    const client = requireClient();
    if (!client) return;

    const roomId = roomIdInput.trim();
    if (!roomId) {
      setErrorMessage('请输入房间号');
      return;
    }

    const displayName = playerName.trim() || playerId;
    persistString(PLAYER_NAME_STORAGE_KEY, displayName);
    updateRoomIdInput(roomId);
    client.joinRoom(roomId, playerId, displayName);
    setMyCharacter(null);
    appendLog(`已发送加入房间请求：${roomId}`);
  }

  function resumeLastRoom(): void {
    const roomId = roomIdInput.trim() || getStoredString(LAST_ROOM_ID_STORAGE_KEY);
    if (!roomId) {
      setErrorMessage('没有可恢复的房间号');
      return;
    }

    const displayName = playerName.trim() || playerId;
    updateRoomIdInput(roomId);
    persistString(PLAYER_NAME_STORAGE_KEY, displayName);

    const join = (client: GameWebSocketClient): void => {
      client.joinRoom(roomId, playerId, displayName);
      setMyCharacter(null);
      appendLog(`已恢复并加入房间：${roomId}`);
    };

    const client = clientRef.current;
    if (client?.status === 'connected') {
      join(client);
      return;
    }

    connect(join);
  }

  function copyInviteText(): void {
    const roomId = currentRoomId.trim();
    if (!roomId) {
      setErrorMessage('没有可复制的房间号');
      return;
    }

    const inviteText = `血染钟楼房间号：${roomId}\n昵称：${playerName.trim() || playerId}\n打开小程序后输入房间号加入。`;
    void Taro.setClipboardData({ data: inviteText })
      .then(() => {
        setErrorMessage('');
        void Taro.showToast({ title: '邀请信息已复制', icon: 'success' });
        appendLog(`已复制邀请信息 ${roomId}`);
      })
      .catch(() => {
        setErrorMessage('复制邀请信息失败');
      });
  }

  function openScriptPage(): void {
    void Taro.navigateTo({ url: '/pages/scripts/index' });
  }

  function leaveRoom(): void {
    const client = requireClient();
    if (!client) return;

    client.leaveRoom();
    setRoomState(null);
    setMyCharacter(null);
    updateRoomIdInput('');
    appendLog('已发送离开房间请求');
  }

  function kickPlayer(targetPlayerId: string): void {
    const client = requireClient();
    if (!client) return;
    client.kickPlayer(targetPlayerId);
    appendLog(`已发送踢出玩家请求：${targetPlayerId}`);
  }

  function updateRoomSettings(): void {
    const client = requireClient();
    if (!client) return;

    const maxPlayers = Number.parseInt(maxPlayersInput, 10);
    if (!Number.isInteger(maxPlayers) || maxPlayers < 5 || maxPlayers > 15) {
      setErrorMessage('实际玩家数必须在 5-15 之间');
      return;
    }

    client.updateRoomSettings(maxPlayers);
    appendLog(`已发送保存设置请求：${maxPlayers} 人`);
  }

  function setStoryteller(targetPlayerId: string): void {
    const client = requireClient();
    if (!client) return;

    client.setStoryteller(targetPlayerId);
    appendLog(`已发送设置说书人请求：${targetPlayerId}`);
  }

  function assignSampleCharacters(): void {
    const client = requireClient();
    if (!client) return;
    if (!roomState?.storytellerId) {
      setErrorMessage('请先设置说书人');
      return;
    }

    try {
      const assignments = buildSampleAssignments(roomState.players);
      const fortuneTellerRedHerringId = chooseFortuneTellerRedHerring(assignments);
      client.assignCharacters(assignments, undefined, fortuneTellerRedHerringId);
      appendLog(`已发送角色分配请求，共 ${Object.keys(assignments).length} 人`);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '生成示例分配失败');
    }
  }

  function resetIdentity(): void {
    const nextId = `player_${Math.random().toString(36).slice(2, 10)}`;
    Taro.setStorageSync(PLAYER_ID_STORAGE_KEY, nextId);
    setPlayerId(nextId);
    setPlayerName(`玩家${nextId.slice(-4)}`);
    persistString(PLAYER_NAME_STORAGE_KEY, `玩家${nextId.slice(-4)}`);
    updateRoomIdInput('');
    setRoomState(null);
    setMyCharacter(null);
    appendLog(`已重置身份 ${nextId}`);
  }

  // ─── Game Flow Actions ────────────────────────────────────────

  function startGame(): void {
    const client = requireClient();
    if (!client) return;
    client.startGame();
    setGameOver(null);
    appendLog('已发送开始游戏请求');
  }

  function changePhase(phase: string): void {
    const client = requireClient();
    if (!client) return;
    client.changePhase(phase);
    appendLog(`已发送阶段切换请求：${PHASE_LABELS[phase as GamePhase] ?? phase}`);
  }

  function nominatePlayer(): void {
    const client = requireClient();
    if (!client) return;
    const nomineeId = nomineeIdInput.trim();
    if (!nomineeId) {
      setErrorMessage('请选择被提名的玩家');
      return;
    }
    client.nominate(nomineeId);
    setNomineeIdInput('');
    appendLog(`已发送提名请求：${nomineeId}`);
  }

  function castVote(decision: boolean): void {
    const client = requireClient();
    if (!client) return;
    client.castVote(decision);

    // If I'm dead, consume my ghost vote locally
    if (!roomState?.players.find((player) => player.id === playerId)?.isAlive) {
      setGhostVotesRemaining((current) => {
        const next = new Set(current);
        next.delete(playerId);
        return next;
      });
    }
    appendLog(`已投票: ${decision ? '赞成处决' : '反对处决'}`);
  }

  function resolveNomination(): void {
    const client = requireClient();
    if (!client) return;
    client.resolveNomination();
    appendLog('已发送结算提名请求');
  }

  function executePlayer(targetPlayerId: string): void {
    const client = requireClient();
    if (!client) return;
    client.executePlayer(targetPlayerId);
    appendLog(`已发送处决请求：${targetPlayerId}`);
  }

  function useSlayerAbility(targetPlayerId: string): void {
    const client = requireClient();
    if (!client) return;
    client.useSlayerAbility(targetPlayerId);
    appendLog(`已发送杀手能力请求：${targetPlayerId}`);
  }

  function declarePlayerDeath(): void {
    const client = requireClient();
    if (!client) return;
    if (!manualDeathTargetId) {
      setErrorMessage('请选择要宣告死亡的玩家');
      return;
    }

    client.killPlayer(manualDeathTargetId, manualDeathCause);
    appendLog(`已发送死亡宣告请求：${manualDeathTargetId}（${DEATH_CAUSE_LABELS[manualDeathCause]}）`);
    setManualDeathTargetId('');
  }

  function submitNightAction(): void {
    const client = requireClient();
    if (!client) return;
    const action = NIGHT_ACTION_TYPES.find((a) => a.id === nightActionType);
    if (!action) return;
    if (currentNightWakeStep && nightActionType !== currentNightWakeStep.actionType) {
      const expectedActionLabel = NIGHT_ACTION_TYPES.find((item) => item.id === currentNightWakeStep.actionType)?.label ?? '指定行动';
      setErrorMessage(`当前步骤需要${expectedActionLabel}`);
      return;
    }
    if (nightTargetIds.length < selectedNightMinTargets) {
      setErrorMessage(`此行动至少需要选择 ${selectedNightMinTargets} 个目标`);
      return;
    }
    if (nightTargetIds.length > selectedNightMaxTargets) {
      setErrorMessage(`此行动最多选择 ${selectedNightMaxTargets} 个目标`);
      return;
    }
    const result = nightResultInput.trim() || null;
    client.submitNightAction(nightActionType, nightTargetIds, result ?? undefined);
    setNightTargetIds([]);
    setNightResultInput('');
    appendLog(`已发送夜间行动: ${action.label}`);
  }

  function resolveNight(): void {
    const client = requireClient();
    if (!client) return;
    client.resolveNight();
    appendLog('已发送结束夜晚请求');
  }

  function endGame(winner: 'good' | 'evil'): void {
    const client = requireClient();
    if (!client) return;
    client.endGame(winner, endGameDescriptionInput);
    appendLog(`已发送结束游戏请求：${winner === 'good' ? '善良阵营' : '邪恶阵营'}`);
  }

  function returnToLobby(): void {
    setGamePhase('setup');
    setDayNumber(0);
    setRoomState(null);
    setMyCharacter(null);
    setCurrentNomination(null);
    setLastNominationResult(null);
    setDeathRecords({});
    setGhostVotesRemaining(new Set());
    setDeathAnnouncements([]);
    setNightActions([]);
    setNightTargetIds([]);
    setGameOver(null);
    setEndGameDescriptionInput('');
    updateRoomIdInput('');
    appendLog('已返回大厅');
  }

  // ─── Game Over Screen ───────────────────────────────────────
  if (gameOver) {
    const winnerLabel = gameOver.winner === 'good' ? '善良阵营' : '邪恶阵营';
    return (
      <ScrollView className='page' scrollY>
        <View className='hero'>
          <View>
            <Text className='eyebrow'>游戏结束</Text>
            <Text className='title'>游戏结束</Text>
            <Text className='subtitle'>{winnerLabel}</Text>
          </View>
        </View>

        <View className='card'>
          <View className='gameOverBox'>
            <Text className='gameOverWinner'>
              {gameOver.winner === 'good' ? '正义获胜' : '邪恶获胜'}
            </Text>
            <Text className='gameOverTeam'>阵营：{winnerLabel}</Text>
            <Text className='gameOverReason'>{gameOver.description}</Text>
          </View>
        </View>

        <View className='card'>
          <Text className='sectionTitle'>◆ 角色揭示</Text>
          {(roomState?.players ?? []).map((player) => {
            const isDead = !player.isAlive;
            const team = player.character?.team === 2 ? 'evil' : 'good';
            return (
              <View className={`player ${isDead ? 'playerDead' : ''}`} key={player.id}>
                <View className='playerInfo'>
                  <Text className='playerName'>
                    {player.name || player.id}
                    {isDead ? '（死亡）' : ''}
                  </Text>
                  <Text className='hint'>
                    角色：{player.character?.name ?? '未知'}
                    {player.character ? ` | 阵营：${team === 'good' ? '善良' : '邪恶'}` : ''}
                  </Text>
                  {player.character && (
                    <Text className='ability'>{player.character.ability}</Text>
                  )}
                </View>
              </View>
            );
          })}
        </View>

        <Button className='button primary' onClick={returnToLobby}>← 返回大厅</Button>
      </ScrollView>
    );
  }

  return (
    <ScrollView className='page' scrollY>
      <View className='hero'>
        <View>
          <Text className='eyebrow'>血染钟楼 H5</Text>
          <Text className='title'>血染钟楼线上房间</Text>
          <Text className='subtitle'>剧本：{currentScriptName}</Text>
        </View>
        <View className={`statusPill status-${status}`}>
          <Text className='statusIcon'>{statusIcon}</Text>
          <Text>{STATUS_LABELS[status]}</Text>
        </View>
      </View>

      <View className='summaryGrid'>
        <View className='summaryItem'>
          <Text className='summaryIcon'>#</Text>
          <Text className='summaryLabel'>房间</Text>
          <Text className='summaryValue'>{roomCodeLabel}</Text>
        </View>
        <View className='summaryItem'>
          <Text className='summaryIcon'>@</Text>
          <Text className='summaryLabel'>身份</Text>
          <Text className='summaryValue'>{roleLabel}</Text>
        </View>
        <View className='summaryItem'>
          <Text className='summaryIcon'>◆</Text>
          <Text className='summaryLabel'>阶段</Text>
          <Text className='summaryValue'>{PHASE_LABELS[gamePhase]}</Text>
        </View>
        <View className='summaryItem'>
          <Text className='summaryIcon'>●</Text>
          <Text className='summaryLabel'>玩家</Text>
          <Text className='summaryValue'>{roomOccupancy}</Text>
        </View>
      </View>

      <View className='card phaseBar'>
        <View>
          <Text className='phaseLabel'>{PHASE_LABELS[gamePhase]}</Text>
          {dayNumber > 0 && <Text className='dayNumber'>第 {dayNumber} 天</Text>}
        </View>
        <Text className='phaseHint'>
          {gamePhase === 'night'
            ? '说书人正在处理夜间行动'
            : gamePhase === 'day'
              ? '白天阶段可提名、投票与执行公开行动'
              : gamePhase === 'voting'
                ? '当前提名正在投票'
                : gamePhase === 'setup'
                  ? '连接后创建或加入房间'
                  : '游戏已结束'}
        </Text>
      </View>

      <View className='card'>
        <Text className='sectionTitle'>↔ 连接</Text>
        <Text className={`status status-${status}`}>状态：{STATUS_LABELS[status]}</Text>
        <Input className='input' value={wsUrl} placeholder='服务地址' onInput={(event: InputEvent) => updateWsUrl(eventValue(event))} />
        <View className='row'>
          <Button className='button primary' onClick={() => connect()}>↔ 连接</Button>
          <Button className='button' onClick={disconnect}>× 断开</Button>
        </View>
      </View>

      <View className='card'>
        <Text className='sectionTitle'>@ 匿名身份</Text>
        <Text className='mono'>玩家编号：{playerId}</Text>
        <Input className='input' value={playerName} placeholder='昵称' onInput={(event: InputEvent) => updatePlayerName(eventValue(event))} />
        <Button className='button warn' onClick={resetIdentity}>↻ 重置匿名身份</Button>
      </View>

      <View className='card'>
        <Text className='sectionTitle'># 房间</Text>
        <Input className='input' value={maxPlayersInput} type='number' placeholder='实际玩家数，默认 5' onInput={(event: InputEvent) => updateMaxPlayersInput(eventValue(event))} />
        <View className='row'>
          <Button className='button primary' onClick={createRoom}>+ 创建房间</Button>
          {isRoomCreator && gamePhase === 'setup' && (
            <Button className='button' onClick={updateRoomSettings}>✓ 保存设置</Button>
          )}
        </View>
        <Input className='input' value={roomIdInput} placeholder='房间号' onInput={(event: InputEvent) => updateRoomIdInput(eventValue(event))} />
        <View className='row'>
          <Button className='button primary' onClick={joinRoom}>→ 加入房间</Button>
          <Button className='button' onClick={resumeLastRoom}>↻ 恢复最近</Button>
          <Button className='button' onClick={leaveRoom}>← 离开房间</Button>
        </View>
        <View className='inviteBox'>
          <Text className='hint'>当前房间</Text>
          <Text className='inviteCode'>{currentRoomId || '未加入'}</Text>
          <Text className='hint'>剧本：{currentScriptName}</Text>
          <Button className='button' onClick={copyInviteText}>⧉ 复制邀请信息</Button>
        </View>
        <Button className='button' onClick={openScriptPage}>≡ 查看剧本与夜晚顺序</Button>
        <Text className='hint'>提示：5 人局需要 1 名说书人和 5 名实际玩家身份。</Text>
      </View>

      <View className='card'>
        <Text className='sectionTitle'>● 玩家列表</Text>
        <View className='metaLine'>
          <Text>说书人：{storytellerName}</Text>
          <Text>{aliveSummary}</Text>
        </View>
        {players.length === 0 && (
          <Text className='emptyState'>连接并创建或加入房间后，玩家会显示在这里。</Text>
        )}
        {players.map((player) => {
          const isDead = !player.isAlive;
          const hasGhostVote = ghostVotesRemaining.has(player.id);
          const poisonedUntil = isStoryteller && typeof player.poisonedUntil === 'number'
            ? player.poisonedUntil
            : null;
          const shownCharacterText = isStoryteller && player.shownCharacter
            ? `（显示为 ${player.shownCharacter.name}）`
            : '';
          return (
            <View className={`player ${isDead ? 'playerDead' : ''}`} key={player.id}>
              <View className='playerInfo'>
                <Text className='playerName'>
                  {player.name || player.id}
                  {isDead ? '（死亡）' : ''}
                </Text>
                <Text className='mono'>{player.id}</Text>
                <Text className='hint'>
                  角色：{player.character?.name ?? '隐藏/未分配'}{shownCharacterText}
                  {isDead && hasGhostVote ? ' | 幽灵票可用' : ''}
                  {isDead && !hasGhostVote ? ' | 幽灵票已用' : ''}
                </Text>
                {poisonedUntil !== null && (
                  <Text className='poisonStatus'>中毒至第 {poisonedUntil} 天黄昏</Text>
                )}
              </View>
              {roomState && !roomState.storytellerId && (
                <Button className='miniButton' onClick={() => setStoryteller(player.id)}>☆ 设为说书人</Button>
              )}
              {isRoomCreator && gamePhase === 'setup' && player.id !== playerId && (
                <Button className='miniButton danger' onClick={() => kickPlayer(player.id)}>× 踢出</Button>
              )}
              {isStoryteller && gamePhase === 'day' && !isDead && (
                <Button className='miniButton danger' onClick={() => executePlayer(player.id)}>! 处决</Button>
              )}
            </View>
          );
        })}
        {roomState && !roomState.storytellerId && playerId && (
          <Button className='button' onClick={() => setStoryteller(playerId)}>☆ 设自己为说书人</Button>
        )}
      </View>

      <View className='card'>
        <Text className='sectionTitle'>◆ 角色</Text>
        {visibleCharacter ? (
          <View className='character'>
            <Text className='characterName'>{visibleCharacter.name}</Text>
            <Text className='hint'>阵营：{visibleCharacter.team === 2 ? '邪恶阵营' : '善良阵营'}</Text>
            <Text className='ability'>{visibleCharacter.ability}</Text>
          </View>
        ) : (
          <Text className='hint'>你还没有收到角色，或你是说书人。</Text>
        )}
        {isStoryteller && (
          <Button className='button primary' onClick={assignSampleCharacters}>✓ 一键示例分配角色</Button>
        )}
        {!isStoryteller && <Text className='hint'>只有说书人可以分配角色。</Text>}
      </View>

      {canUseSlayerAbility && (
        <View className='card'>
          <Text className='sectionTitle'>! 杀手能力</Text>
          <View className='targetSelector'>
            {alivePlayers.map((player) => (
              <Button
                className='miniButton'
                key={player.id}
                onClick={() => useSlayerAbility(player.id)}
              >
                {player.name || player.id}
              </Button>
            ))}
          </View>
        </View>
      )}

      {/* ─── Storyteller Controls ──────────────────────────── */}
      {isStoryteller && (
        <View className='card'>
          <Text className='sectionTitle'>☆ 说书人控制</Text>

          {gamePhase === 'setup' && (
            <Button className='button primary' onClick={startGame}>▶ 开始游戏</Button>
          )}

          {gamePhase !== 'setup' && gamePhase !== 'finished' && (
            <View className='row'>
              <Button className='button' onClick={() => changePhase('day')}>日 进入白天</Button>
              <Button className='button' onClick={() => changePhase('night')}>夜 进入夜晚</Button>
            </View>
          )}

          {gamePhase !== 'setup' && gamePhase !== 'finished' && (
            <View className='endGameControls'>
              <Text className='sectionTitle'>! 宣告死亡</Text>
              <View className='actionTypeList'>
                {DEATH_CAUSE_OPTIONS.map((cause) => (
                  <Button
                    className={`miniButton ${manualDeathCause === cause.id ? 'selected' : ''}`}
                    key={cause.id}
                    onClick={() => setManualDeathCause(cause.id)}
                  >
                    {cause.label}
                  </Button>
                ))}
              </View>
              <View className='targetSelector'>
                {alivePlayers.map((player) => (
                  <Button
                    className={`miniButton ${manualDeathTargetId === player.id ? 'selected' : ''}`}
                    key={player.id}
                    onClick={() => setManualDeathTargetId(player.id)}
                  >
                    {player.name || player.id}
                  </Button>
                ))}
              </View>
              <Button className='button danger' onClick={declarePlayerDeath}>! 宣告死亡</Button>
            </View>
          )}

          {gamePhase !== 'setup' && gamePhase !== 'finished' && (
            <View className='endGameControls'>
              <Input
                className='input'
                value={endGameDescriptionInput}
                placeholder='结局说明（可选）'
                onInput={(event: InputEvent) => setEndGameDescriptionInput(eventValue(event))}
              />
              <View className='row'>
                <Button className='button primary' onClick={() => endGame('good')}>✓ 善良胜利</Button>
                <Button className='button danger' onClick={() => endGame('evil')}>! 邪恶胜利</Button>
              </View>
            </View>
          )}
        </View>
      )}

      {/* ─── Nomination UI (Day Phase) ─────────────────────── */}
      {gamePhase === 'day' && roomState && (
        <View className='card'>
          <Text className='sectionTitle'>→ 提名</Text>
          <Text className='hint'>在白天阶段，任何活着的玩家可以提名其他玩家。</Text>
          <Input
            className='input'
            value={nomineeIdInput}
            placeholder='被提名玩家编号'
            onInput={(event: InputEvent) => setNomineeIdInput(eventValue(event))}
          />
          <View className='row'>
            {alivePlayers.map((player) => (
              <Button
                className='miniButton'
                key={player.id}
                onClick={() => setNomineeIdInput(player.id)}
              >
                {player.name || player.id}
              </Button>
            ))}
          </View>
          <Button className='button primary' onClick={nominatePlayer}>→ 发起提名</Button>
        </View>
      )}

      {/* ─── Voting UI ─────────────────────────────────────── */}
      {gamePhase === 'voting' && currentNomination && (
        <View className='card'>
          <Text className='sectionTitle'>✓ 投票</Text>
          <View className='nominationBanner'>
            <Text className='nominationText'>
              {roomState?.players.find((player) => player.id === currentNomination.nominatorId)?.name ?? currentNomination.nominatorId}
              {' 提名 '}
              {roomState?.players.find((player) => player.id === currentNomination.nomineeId)?.name ?? currentNomination.nomineeId}
            </Text>
            <Text className='hint'>处决阈值：{currentExecutionThreshold} 张赞成票</Text>
          </View>

          <View className='voteButtons'>
            <Button className='button voteYes' onClick={() => castVote(true)}>✓ 赞成处决</Button>
            <Button className='button voteNo' onClick={() => castVote(false)}>× 反对处决</Button>
          </View>

          {/* Ghost vote indicator for dead players */}
          {(() => {
            const self = roomState?.players.find((player) => player.id === playerId);
            const isSelfDead = self ? !self.isAlive : false;
            const hasGhost = ghostVotesRemaining.has(playerId);
            if (isSelfDead && hasGhost) {
              return <Text className='ghostVoteHint'>你已死亡，但可以使用幽灵票投一次票</Text>;
            }
            if (isSelfDead && !hasGhost) {
              return <Text className='ghostVoteSpent'>你已用完幽灵票</Text>;
            }
            return null;
          })()}

          {/* Vote tally display */}
          <View className='voteTally'>
            <Text className='sectionTitle'>≡ 投票记录</Text>
            {Object.entries(currentNomination.votes).map(([voterId, decision]) => {
              const voterName = roomState?.players.find((player) => player.id === voterId)?.name ?? voterId;
              return (
                <Text className='voteRecord' key={voterId}>
                  {voterName}: {decision ? '赞成' : '反对'}
                </Text>
              );
            })}
          </View>

          {isStoryteller && (
            <Button className='button primary' onClick={resolveNomination}>✓ 结算投票</Button>
          )}
        </View>
      )}

      {/* ─── Last Nomination Result ────────────────────────── */}
      {lastNominationResult && (
        <View className='card'>
          <Text className='sectionTitle'>= 投票结果</Text>
          <View className={lastNominationResult.executed ? 'resultExecuted' : 'resultSpared'}>
            <Text className='resultText'>
              {roomState?.players.find((player) => player.id === lastNominationResult.nomineeId)?.name ?? lastNominationResult.nomineeId}
              {lastNominationResult.executed ? ' 被处决了' : ' 幸存'}
            </Text>
            <Text className='hint'>
              赞成: {lastNominationResult.yesVotes} | 反对: {lastNominationResult.noVotes}
              {lastNominationResult.requiredVotes !== null ? ` | 处决阈值: ${lastNominationResult.requiredVotes}` : ''}
            </Text>
          </View>
        </View>
      )}

      {/* ─── Death Tracking ────────────────────────────────── */}
      {deadPlayers.length > 0 && (
        <View className='card'>
          <Text className='sectionTitle'>! 死亡记录</Text>
          {deadPlayers.map((player) => {
            const record = deathRecords[player.id];
            const hasGhost = ghostVotesRemaining.has(player.id);
            return (
              <View className='player playerDead' key={player.id}>
                <View className='playerInfo'>
                  <Text className='playerName'>{player.name || player.id}（死亡）</Text>
                  <Text className='hint'>
                    {record ? `死因: ${DEATH_CAUSE_LABELS[record.cause] ?? record.cause} | 第 ${record.dayNumber} 天` : '已死亡'}
                    {hasGhost ? ' | 幽灵票可用' : ' | 幽灵票已用'}
                  </Text>
                </View>
              </View>
            );
          })}
        </View>
      )}

      {/* ─── Death Announcements ───────────────────────────── */}
      {deathAnnouncements.length > 0 && (
        <View className='card'>
          <Text className='sectionTitle'>! 死亡公告</Text>
          {deathAnnouncements.map((announcement, index) => (
            <Text className='deathAnnouncement' key={`${announcement}-${index}`}>{announcement}</Text>
          ))}
        </View>
      )}

      {/* ─── Night Phase UI (Storyteller only) ─────────────── */}
      {gamePhase === 'night' && isStoryteller && (
        <View className='card'>
          <Text className='sectionTitle'>N 夜间行动</Text>

          <View className='wakeOrderList'>
            <Text className='sectionTitle'>≡ 唤醒顺序</Text>
            {nightWakeSteps.map((step, index) => {
              const character = TROUBLE_BREWING_SCRIPT.characters.find((item) => item.id === step.characterId);
              const groupLabel = step.characterType ? CHARACTER_TYPE_LABELS[step.characterType] ?? step.characterType : '';
              const stepLabel = character?.name || groupLabel || step.characterId;
              const inPlay = step.characterType
                ? assignedCharacterTypes.has(step.characterType)
                : assignedCharacterIds.has(step.characterId);
              const isCurrent = currentNightWakeStep?.characterId === step.characterId && currentNightWakeStep.order === step.order;
              const isCompleted = index < currentNightWakeIndex;
              return (
                <View
                  className={`wakeStep ${inPlay ? 'wakeStepActive' : 'wakeStepInactive'} ${isCurrent ? 'wakeStepCurrent' : ''} ${isCompleted ? 'wakeStepDone' : ''}`}
                  key={`${step.order}-${step.characterId || step.characterType || step.actionType}`}
                >
                  <View className='wakeStepHeader'>
                    <Text className='wakeStepOrder'>{step.order}</Text>
                    <Text className='wakeStepName'>
                      {stepLabel}
                      {isCurrent ? ' · 当前' : isCompleted ? ' · 已完成' : inPlay ? ' · 待处理' : ' · 未在场'}
                    </Text>
                  </View>
                  <Text className='hint'>{step.prompt}</Text>
                  <Text className='hint'>目标数：{step.minTargets} - {step.maxTargets}</Text>
                </View>
              );
            })}
          </View>

          {/* Action type selector */}
          <Text className='hint'>选择行动类型:</Text>
          <View className='actionTypeList'>
            {NIGHT_ACTION_TYPES.map((action) => (
              <Button
                key={action.id}
                className={`miniButton ${nightActionType === action.id ? 'selected' : ''}`}
                onClick={() => {
                  setNightActionType(action.id);
                  setNightTargetIds([]);
                }}
              >
                {action.label}
              </Button>
            ))}
          </View>

          {/* Target selector */}
          {selectedNightNeedsTarget && (
            <View className='targetSelector'>
              <Text className='hint'>选择目标：{selectedNightMinTargets} - {selectedNightMaxTargets} 个</Text>
              {alivePlayers.map((player) => (
                <Button
                  key={player.id}
                  className={`miniButton ${nightTargetIds.includes(player.id) ? 'selected' : ''}`}
                  onClick={() => toggleNightTarget(player.id)}
                >
                  {player.name || player.id}
                </Button>
              ))}
            </View>
          )}

          {/* Result input (for storyteller to enter learned info) */}
          <Input
            className='input'
            value={nightResultInput}
            placeholder='行动结果（可选，用于记录信息）'
            onInput={(event: InputEvent) => setNightResultInput(eventValue(event))}
          />

          <View className='row'>
            <Button className='button primary' onClick={submitNightAction}>✓ 提交行动</Button>
            <Button className='button warn' onClick={resolveNight}>→ 结束夜晚</Button>
          </View>

          {/* Night actions recorded this night */}
          {nightActions.length > 0 && (
            <View className='nightActionLog'>
              <Text className='sectionTitle'>≡ 今夜行动记录</Text>
              {nightActions.map((action, index) => {
                const actionLabel = NIGHT_ACTION_TYPES.find((a) => a.id === action.actionType)?.label ?? action.actionType;
                const targetNames = action.targetIds.map(
                  (tid) => roomState?.players.find((player) => player.id === tid)?.name ?? tid,
                );
                return (
                  <View className='nightActionEntry' key={index}>
                    <Text className='hint'>
                      {actionLabel}
                      {targetNames.length > 0 ? `：${targetNames.join('、')}` : ''}
                    </Text>
                    {action.result && <Text className='nightActionResult'>结果: {action.result}</Text>}
                  </View>
                );
              })}
            </View>
          )}
        </View>
      )}

      {/* ─── Night Phase (non-storyteller) ─────────────────── */}
      {gamePhase === 'night' && !isStoryteller && (
        <View className='card'>
          <Text className='sectionTitle'>N 夜晚</Text>
          <Text className='hint'>夜晚降临... 请闭上眼睛。</Text>
          <Text className='hint'>说书人正在处理夜间行动，请耐心等待。</Text>
        </View>
      )}

      {errorMessage && (
        <View className='error'>
          <Text>{errorMessage}</Text>
        </View>
      )}

      <View className='card'>
        <Text className='sectionTitle'>≡ 日志</Text>
        <Text className='hint'>最近：{latestLog}</Text>
        {logs.length === 0 && <Text className='emptyState'>操作和服务器消息会显示在这里。</Text>}
        {logs.map((log, index) => (
          <Text className='log' key={`${log}-${index}`}>{log}</Text>
        ))}
      </View>
    </ScrollView>
  );
}
