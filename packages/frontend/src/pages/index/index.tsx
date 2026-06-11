import { Button, Input, ScrollView, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect, useMemo, useRef, useState } from 'react';
import { GameWebSocketClient } from '@clocktower/core';
import type { ConnectionStatus, GameCharacter, RoomState, ServerMessage } from '@clocktower/core';
import { createTaroWebSocketTransport } from '../../lib/taro-websocket-transport';
import './index.css';

const PLAYER_ID_STORAGE_KEY = 'clocktower.playerId';
const DEFAULT_WS_URL = 'ws://localhost:8080/ws';

// ─── Phase & Role Constants ──────────────────────────────────────

type GamePhase = 'setup' | 'day' | 'voting' | 'night' | 'finished';

const PHASE_LABELS: Record<GamePhase, string> = {
  setup: '准备 Setup',
  day: '白天 Day',
  voting: '投票 Voting',
  night: '夜晚 Night',
  finished: '游戏结束 Finished',
};

type DeathCause = 'execution' | 'night_kill' | 'ability';

const DEATH_CAUSE_LABELS: Record<DeathCause, string> = {
  execution: '处决',
  night_kill: '夜晚击杀',
  ability: '能力致死',
};

// ─── Night Action Options ────────────────────────────────────────

const NIGHT_ACTION_TYPES = [
  { id: 'kill', label: '击杀 (Imp)', needsTarget: true },
  { id: 'poison', label: '下毒 (Poisoner)', needsTarget: true },
  { id: 'protect', label: '保护 (Monk)', needsTarget: true },
  { id: 'learn_townsfolk', label: '鉴镇民 (Washerwoman)', needsTarget: false },
  { id: 'learn_outsider', label: '鉴外来者 (Librarian)', needsTarget: false },
  { id: 'learn_minion', label: '鉴爪牙 (Investigator)', needsTarget: false },
  { id: 'learn_evil_pairs', label: '邪恶相邻 (Chef)', needsTarget: false },
  { id: 'learn_evil_neighbors', label: '邪恶邻居 (Empath)', needsTarget: false },
  { id: 'check_demon', label: '查验恶魔 (Fortune Teller)', needsTarget: true },
  { id: 'learn_executed', label: '鉴处决 (Undertaker)', needsTarget: false },
  { id: 'learn_died', label: '鉴死者 (Ravenkeeper)', needsTarget: true },
  { id: 'learn_master', label: '鉴主人 (Butler)', needsTarget: false },
] as const;

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

const ROLE_COUNTS: Record<number, { readonly townsfolk: number; readonly outsiders: number; readonly minions: number; readonly demons: number }> = {
  4: { townsfolk: 3, outsiders: 0, minions: 0, demons: 1 },
  5: { townsfolk: 3, outsiders: 0, minions: 1, demons: 1 },
  6: { townsfolk: 3, outsiders: 1, minions: 1, demons: 1 },
  7: { townsfolk: 5, outsiders: 0, minions: 1, demons: 1 },
  8: { townsfolk: 5, outsiders: 1, minions: 1, demons: 1 },
  9: { townsfolk: 5, outsiders: 2, minions: 1, demons: 1 },
  10: { townsfolk: 7, outsiders: 0, minions: 2, demons: 1 },
  11: { townsfolk: 7, outsiders: 1, minions: 2, demons: 1 },
  12: { townsfolk: 7, outsiders: 2, minions: 2, demons: 1 },
  13: { townsfolk: 9, outsiders: 0, minions: 3, demons: 1 },
  14: { townsfolk: 9, outsiders: 1, minions: 3, demons: 1 },
  15: { townsfolk: 9, outsiders: 2, minions: 3, demons: 1 },
};

const SAMPLE_CHARACTERS = {
  townsfolk: [
    'washerwoman',
    'librarian',
    'investigator',
    'chef',
    'empath',
    'fortuneteller',
    'undertaker',
    'monk',
    'ravenkeeper',
  ],
  outsiders: ['butler', 'drunk'],
  minions: ['poisoner', 'spy', 'baron'],
  demons: ['imp'],
} as const;

type InputEvent = {
  readonly detail: {
    readonly value: string;
  };
};

function getOrCreatePlayerId(): string {
  const stored = Taro.getStorageSync<string>(PLAYER_ID_STORAGE_KEY);
  if (stored) return stored;

  const generated = `player_${Math.random().toString(36).slice(2, 10)}`;
  Taro.setStorageSync(PLAYER_ID_STORAGE_KEY, generated);
  return generated;
}

function buildSampleAssignments(players: RoomState['players']): Record<string, string> {
  const counts = ROLE_COUNTS[players.length];
  if (!counts) {
    throw new Error(`当前玩家数 ${players.length} 没有合法角色分布`);
  }

  const characterIds = [
    ...SAMPLE_CHARACTERS.townsfolk.slice(0, counts.townsfolk),
    ...SAMPLE_CHARACTERS.outsiders.slice(0, counts.outsiders),
    ...SAMPLE_CHARACTERS.minions.slice(0, counts.minions),
    ...SAMPLE_CHARACTERS.demons.slice(0, counts.demons),
  ];

  if (characterIds.length !== players.length) {
    throw new Error(`角色数量 ${characterIds.length} 与玩家数量 ${players.length} 不一致`);
  }

  return players.reduce<Record<string, string>>((assignments, player, index) => {
    const characterId = characterIds[index];
    if (!characterId) return assignments;
    assignments[player.id] = characterId;
    return assignments;
  }, {});
}

function eventValue(event: InputEvent): string {
  return event.detail.value;
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
  } | null>(null);

  // ─── Death State ─────────────────────────────────────────────
  const [deathRecords, setDeathRecords] = useState<Record<string, DeathRecord>>({});
  const [ghostVotesRemaining, setGhostVotesRemaining] = useState<Set<string>>(new Set());
  const [deathAnnouncements, setDeathAnnouncements] = useState<readonly string[]>([]);

  // ─── Night State ─────────────────────────────────────────────
  const [nightActions, setNightActions] = useState<readonly NightActionRecord[]>([]);
  const [nightActionType, setNightActionType] = useState<string>(NIGHT_ACTION_TYPES[0].id);
  const [nightTargetIds, setNightTargetIds] = useState<readonly string[]>([]);
  const [nightResultInput, setNightResultInput] = useState('');

  // ─── Win Condition State ─────────────────────────────────────
  const [gameOver, setGameOver] = useState<GameOverInfo | null>(null);

  useEffect(() => {
    const id = getOrCreatePlayerId();
    setPlayerId(id);
    setPlayerName(`玩家${id.slice(-4)}`);

    return () => {
      clientRef.current?.disconnect();
      clientRef.current = null;
    };
  }, []);

  const isStoryteller = roomState?.storytellerId === playerId;
  const visibleCharacter = useMemo(() => {
    if (myCharacter) return myCharacter;
    const self = roomState?.players.find((player) => player.id === playerId);
    return self?.character ?? null;
  }, [myCharacter, playerId, roomState]);

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

  function applyMessage(message: ServerMessage): void {
    appendLog(`收到 ${message.type}`);

    if (message.type === 'ERROR') {
      setErrorMessage(message.error ?? '未知错误');
      return;
    }

    if (message.type === 'ROOM_STATE') {
      if (message.roomId) setRoomIdInput(message.roomId);
      if (message.state) setRoomState(message.state);
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
      const assignment = (event as { readonly characterAssigned: { readonly playerId: string; readonly character: GameCharacter } }).characterAssigned;
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
              ? { ...player, character: assignment.character }
              : player,
          ),
        };
      });
      return;
    }

    // ─── Phase Changed ───────────────────────────────────────
    if ('phaseChanged' in event) {
      const phaseMap: Record<number, GamePhase> = {
        0: 'setup',
        1: 'day',
        2: 'voting',
        3: 'night',
        4: 'finished',
      };
      const phaseValue = (event as { readonly phaseChanged: { readonly phase: number } }).phaseChanged.phase;
      const mapped = phaseMap[phaseValue];
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
      appendLog(`阶段切换 -> ${mapped ?? phaseValue}`);
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
      appendLog(`玩家死亡: ${deadPlayerId} (${cause})`);
      return;
    }

    // ─── Nomination Started ──────────────────────────────────
    if ('nominationStarted' in event) {
      const nomination = (event as { readonly nominationStarted: { readonly nominatorId: string; readonly nomineeId: string } }).nominationStarted;
      setCurrentNomination({ nominatorId: nomination.nominatorId, nomineeId: nomination.nomineeId, votes: {} });
      setLastNominationResult(null);
      setGamePhase('voting');
      appendLog(`提名: ${nomination.nominatorId} -> ${nomination.nomineeId}`);
      return;
    }

    // ─── Vote Cast ───────────────────────────────────────────
    if ('voteCast' in event) {
      const voteData = (event as { readonly voteCast: { readonly voterId: string; readonly targetId?: string } }).voteCast;
      setCurrentNomination((current) => {
        if (!current) return current;
        return {
          ...current,
          votes: { ...current.votes, [voteData.voterId]: voteData.targetId !== undefined && voteData.targetId !== '' },
        };
      });
      return;
    }

    // ─── Nomination Resolved ─────────────────────────────────
    if ('nominationResolved' in event) {
      const resolved = (event as { readonly nominationResolved: { readonly nomineeId: string; readonly executed: boolean; readonly yesVotes: number; readonly noVotes: number } }).nominationResolved;
      setLastNominationResult({ nomineeId: resolved.nomineeId, executed: resolved.executed, yesVotes: resolved.yesVotes, noVotes: resolved.noVotes });
      setCurrentNomination(null);
      const name = roomState?.players.find((player) => player.id === resolved.nomineeId)?.name ?? resolved.nomineeId;
      appendLog(`投票结果: ${name} ${resolved.executed ? '被处决' : '幸存'} (${resolved.yesVotes}/${resolved.noVotes})`);
      return;
    }

    // ─── Night Action ────────────────────────────────────────
    if ('nightAction' in event) {
      const action = (event as { readonly nightAction: { readonly actorId: string; readonly actionType: string; readonly targetIds: readonly string[]; readonly result: string | null } }).nightAction;
      setNightActions((current) => [...current, { actorId: action.actorId, actionType: action.actionType, targetIds: action.targetIds, result: action.result }]);
      appendLog(`夜间行动: ${action.actionType} (${action.actorId})`);
      return;
    }

    // ─── Game Over ───────────────────────────────────────────
    if ('gameOver' in event) {
      const gameOverEvent = (event as { readonly gameOver: { readonly winner: string; readonly reason: string; readonly description: string } }).gameOver;
      setGameOver({ winner: gameOverEvent.winner, reason: gameOverEvent.reason, description: gameOverEvent.description });
      setGamePhase('finished');
      appendLog(`游戏结束: ${gameOverEvent.winner} 胜利 - ${gameOverEvent.description}`);
      return;
    }
  }

  function connect(): void {
    setErrorMessage('');
    clientRef.current?.disconnect();

    const client = new GameWebSocketClient({
      url: wsUrl.trim(),
      maxReconnectAttempts: 0,
      transportFactory: createTaroWebSocketTransport,
    });
    client.onStatusChange(setStatus);
    client.onMessage(applyMessage);
    client.connect();
    clientRef.current = client;
    appendLog(`连接 ${wsUrl.trim()}`);
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
      setErrorMessage('请先连接 WebSocket');
      return null;
    }
    return client;
  }

  function createRoom(): void {
    const client = requireClient();
    if (!client) return;

    const maxPlayers = Number.parseInt(maxPlayersInput, 10);
    client.createRoom(playerId, playerName.trim() || playerId, Number.isNaN(maxPlayers) ? 5 : maxPlayers);
    setMyCharacter(null);
    appendLog('已发送 CREATE_ROOM');
  }

  function joinRoom(): void {
    const client = requireClient();
    if (!client) return;

    const roomId = roomIdInput.trim();
    if (!roomId) {
      setErrorMessage('请输入房间号');
      return;
    }

    client.joinRoom(roomId, playerId, playerName.trim() || playerId);
    setMyCharacter(null);
    appendLog(`已发送 JOIN_ROOM ${roomId}`);
  }

  function leaveRoom(): void {
    const client = requireClient();
    if (!client) return;

    client.leaveRoom();
    setRoomState(null);
    setMyCharacter(null);
    appendLog('已发送 LEAVE_ROOM');
  }

  function setStoryteller(targetPlayerId: string): void {
    const client = requireClient();
    if (!client) return;

    client.setStoryteller(targetPlayerId);
    appendLog(`已发送 SET_STORYTELLER ${targetPlayerId}`);
  }

  function assignSampleCharacters(): void {
    const client = requireClient();
    if (!client) return;
    if (!roomState?.storytellerId) {
      setErrorMessage('请先设置 Storyteller');
      return;
    }

    try {
      const assignments = buildSampleAssignments(roomState.players);
      client.assignCharacters(assignments);
      appendLog(`已发送 ASSIGN_CHARACTERS，共 ${Object.keys(assignments).length} 人`);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : '生成示例分配失败');
    }
  }

  function resetIdentity(): void {
    const nextId = `player_${Math.random().toString(36).slice(2, 10)}`;
    Taro.setStorageSync(PLAYER_ID_STORAGE_KEY, nextId);
    setPlayerId(nextId);
    setPlayerName(`玩家${nextId.slice(-4)}`);
    setRoomState(null);
    setMyCharacter(null);
    appendLog(`已重置身份 ${nextId}`);
  }

  // ─── Game Flow Actions ────────────────────────────────────────

  function startGame(): void {
    const client = requireClient();
    if (!client) return;
    client.startGame();
    setGamePhase('day');
    setDayNumber(1);
    setGameOver(null);
    appendLog('已发送 START_GAME');
  }

  function changePhase(phase: string): void {
    const client = requireClient();
    if (!client) return;
    client.changePhase(phase);
    appendLog(`已发送 CHANGE_PHASE -> ${phase}`);
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
    appendLog(`已发送 NOMINATE -> ${nomineeId}`);
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
    appendLog('已发送 RESOLVE_NOMINATION');
  }

  function executePlayer(targetPlayerId: string): void {
    const client = requireClient();
    if (!client) return;
    client.executePlayer(targetPlayerId);
    appendLog(`已发送 EXECUTE_PLAYER -> ${targetPlayerId}`);
  }

  function submitNightAction(): void {
    const client = requireClient();
    if (!client) return;
    const action = NIGHT_ACTION_TYPES.find((a) => a.id === nightActionType);
    if (!action) return;
    if (action.needsTarget && nightTargetIds.length === 0) {
      setErrorMessage('此行动需要选择目标');
      return;
    }
    const result = nightResultInput.trim() || null;
    client.submitNightAction(nightActionType, nightTargetIds);
    if (result) {
      // Store result locally for display
      setNightActions((current) => [
        ...current,
        {
          actorId: playerId,
          actionType: nightActionType,
          targetIds: nightTargetIds,
          result,
        },
      ]);
    }
    setNightTargetIds([]);
    setNightResultInput('');
    appendLog(`已发送夜间行动: ${action.label}`);
  }

  function resolveNight(): void {
    const client = requireClient();
    if (!client) return;
    client.resolveNight();
    appendLog('已发送 RESOLVE_NIGHT');
  }

  function returnToLobby(): void {
    setGamePhase('setup');
    setDayNumber(0);
    setCurrentNomination(null);
    setLastNominationResult(null);
    setDeathRecords({});
    setGhostVotesRemaining(new Set());
    setDeathAnnouncements([]);
    setNightActions([]);
    setGameOver(null);
    appendLog('已返回大厅');
  }

  // ─── Game Over Screen ───────────────────────────────────────
  if (gameOver) {
    const winnerLabel = gameOver.winner === 'good' ? '善良阵营 Good' : '邪恶阵营 Evil';
    return (
      <ScrollView className='page' scrollY>
        <View className='hero'>
          <Text className='title'>游戏结束 Game Over</Text>
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
          <Text className='sectionTitle'>角色揭示 [Character Reveal]</Text>
          {(roomState?.players ?? []).map((player) => {
            const isDead = !player.isAlive;
            const team = player.character?.team === 2 ? 'evil' : 'good';
            return (
              <View className={`player ${isDead ? 'playerDead' : ''}`} key={player.id}>
                <View className='playerInfo'>
                  <Text className='playerName'>
                    {player.name || player.id}
                    {isDead ? ' [死亡]' : ''}
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

        {isStoryteller && (
          <Button className='button primary' onClick={returnToLobby}>返回大厅</Button>
        )}
      </ScrollView>
    );
  }

  return (
    <ScrollView className='page' scrollY>
      <View className='hero'>
        <Text className='title'>血染钟楼 MVP 调试台</Text>
        <Text className='subtitle'>创建房间、加入房间、指定 Storyteller、一键分配角色</Text>
      </View>

      {/* ─── Phase Display ─────────────────────────────────── */}
      {gamePhase !== 'setup' && (
        <View className='card phaseBar'>
          <Text className='phaseLabel'>{PHASE_LABELS[gamePhase]}</Text>
          {dayNumber > 0 && <Text className='dayNumber'>第 {dayNumber} 天</Text>}
          {gamePhase === 'night' && (
            <Text className='hint'>Storyteller 正在处理夜间行动...</Text>
          )}
        </View>
      )}

      <View className='card'>
        <Text className='sectionTitle'>连接 [Connection]</Text>
        <Text className={`status status-${status}`}>状态：{status}</Text>
        <Input className='input' value={wsUrl} placeholder='WebSocket 地址' onInput={(event: InputEvent) => setWsUrl(eventValue(event))} />
        <View className='row'>
          <Button className='button primary' onClick={connect}>连接</Button>
          <Button className='button' onClick={disconnect}>断开</Button>
        </View>
      </View>

      <View className='card'>
        <Text className='sectionTitle'>匿名身份 [Anonymous Identity]</Text>
        <Text className='mono'>playerId: {playerId}</Text>
        <Input className='input' value={playerName} placeholder='昵称' onInput={(event: InputEvent) => setPlayerName(eventValue(event))} />
        <Button className='button warn' onClick={resetIdentity}>重置匿名身份</Button>
      </View>

      <View className='card'>
        <Text className='sectionTitle'>房间 [Room]</Text>
        <Input className='input' value={maxPlayersInput} type='number' placeholder='实际玩家数，默认 5' onInput={(event: InputEvent) => setMaxPlayersInput(eventValue(event))} />
        <Button className='button primary' onClick={createRoom}>创建房间</Button>
        <Input className='input' value={roomIdInput} placeholder='房间号' onInput={(event: InputEvent) => setRoomIdInput(eventValue(event))} />
        <View className='row'>
          <Button className='button primary' onClick={joinRoom}>加入房间</Button>
          <Button className='button' onClick={leaveRoom}>离开房间</Button>
        </View>
        <Text className='mono'>当前房间：{roomState?.roomId ?? (roomIdInput || '未加入')}</Text>
        <Text className='hint'>提示：5 人局需要 1 个 Storyteller + 5 个实际玩家身份。</Text>
      </View>

      <View className='card'>
        <Text className='sectionTitle'>玩家列表 [Players]</Text>
        <Text className='hint'>Storyteller: {roomState?.storytellerId ?? '未设置'}</Text>
        {(roomState?.players ?? []).map((player) => {
          const isDead = !player.isAlive;
          const hasGhostVote = ghostVotesRemaining.has(player.id);
          return (
            <View className={`player ${isDead ? 'playerDead' : ''}`} key={player.id}>
              <View className='playerInfo'>
                <Text className='playerName'>
                  {player.name || player.id}
                  {isDead ? ' [死亡]' : ''}
                </Text>
                <Text className='mono'>{player.id}</Text>
                <Text className='hint'>
                  角色：{player.character?.name ?? '隐藏/未分配'}
                  {isDead && hasGhostVote ? ' | 幽灵票可用' : ''}
                  {isDead && !hasGhostVote ? ' | 幽灵票已用' : ''}
                </Text>
              </View>
              {roomState && !roomState.storytellerId && (
                <Button className='miniButton' onClick={() => setStoryteller(player.id)}>设为 ST</Button>
              )}
              {isStoryteller && gamePhase === 'day' && !isDead && (
                <Button className='miniButton danger' onClick={() => executePlayer(player.id)}>处决</Button>
              )}
            </View>
          );
        })}
        {!roomState?.storytellerId && playerId && (
          <Button className='button' onClick={() => setStoryteller(playerId)}>设自己为 Storyteller</Button>
        )}
      </View>

      <View className='card'>
        <Text className='sectionTitle'>角色 [Character]</Text>
        {visibleCharacter ? (
          <View className='character'>
            <Text className='characterName'>{visibleCharacter.name}</Text>
            <Text className='hint'>阵营：{visibleCharacter.team === 2 ? '邪恶 Evil' : '善良 Good'}</Text>
            <Text className='ability'>{visibleCharacter.ability}</Text>
          </View>
        ) : (
          <Text className='hint'>你还没有收到角色，或你是 Storyteller。</Text>
        )}
        {isStoryteller && (
          <Button className='button primary' onClick={assignSampleCharacters}>一键示例分配角色</Button>
        )}
        {!isStoryteller && <Text className='hint'>只有 Storyteller 可以分配角色。</Text>}
      </View>

      {/* ─── Storyteller Controls ──────────────────────────── */}
      {isStoryteller && (
        <View className='card'>
          <Text className='sectionTitle'>Storyteller 控制 [Storyteller Controls]</Text>

          {gamePhase === 'setup' && (
            <Button className='button primary' onClick={startGame}>开始游戏</Button>
          )}

          {gamePhase !== 'setup' && gamePhase !== 'finished' && (
            <View className='row'>
              <Button className='button' onClick={() => changePhase('day')}>进入白天</Button>
              <Button className='button' onClick={() => changePhase('night')}>进入夜晚</Button>
            </View>
          )}
        </View>
      )}

      {/* ─── Nomination UI (Day Phase) ─────────────────────── */}
      {gamePhase === 'day' && roomState && (
        <View className='card'>
          <Text className='sectionTitle'>提名 [Nomination]</Text>
          <Text className='hint'>在白天阶段，任何活着的玩家可以提名其他玩家。</Text>
          <Input
            className='input'
            value={nomineeIdInput}
            placeholder='被提名玩家 ID'
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
          <Button className='button primary' onClick={nominatePlayer}>发起提名</Button>
        </View>
      )}

      {/* ─── Voting UI ─────────────────────────────────────── */}
      {gamePhase === 'voting' && currentNomination && (
        <View className='card'>
          <Text className='sectionTitle'>投票 [Voting]</Text>
          <View className='nominationBanner'>
            <Text className='nominationText'>
              {roomState?.players.find((player) => player.id === currentNomination.nominatorId)?.name ?? currentNomination.nominatorId}
              {' -> '}
              {roomState?.players.find((player) => player.id === currentNomination.nomineeId)?.name ?? currentNomination.nomineeId}
            </Text>
          </View>

          <View className='voteButtons'>
            <Button className='button voteYes' onClick={() => castVote(true)}>赞成处决</Button>
            <Button className='button voteNo' onClick={() => castVote(false)}>反对处决</Button>
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
            <Text className='sectionTitle'>投票记录</Text>
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
            <Button className='button primary' onClick={resolveNomination}>结算投票</Button>
          )}
        </View>
      )}

      {/* ─── Last Nomination Result ────────────────────────── */}
      {lastNominationResult && (
        <View className='card'>
          <Text className='sectionTitle'>投票结果 [Vote Result]</Text>
          <View className={lastNominationResult.executed ? 'resultExecuted' : 'resultSpared'}>
            <Text className='resultText'>
              {roomState?.players.find((player) => player.id === lastNominationResult.nomineeId)?.name ?? lastNominationResult.nomineeId}
              {lastNominationResult.executed ? ' 被处决了' : ' 幸存'}
            </Text>
            <Text className='hint'>
              赞成: {lastNominationResult.yesVotes} | 反对: {lastNominationResult.noVotes}
            </Text>
          </View>
        </View>
      )}

      {/* ─── Death Tracking ────────────────────────────────── */}
      {deadPlayers.length > 0 && (
        <View className='card'>
          <Text className='sectionTitle'>死亡记录 [Death Records]</Text>
          {deadPlayers.map((player) => {
            const record = deathRecords[player.id];
            const hasGhost = ghostVotesRemaining.has(player.id);
            return (
              <View className='player playerDead' key={player.id}>
                <View className='playerInfo'>
                  <Text className='playerName'>{player.name || player.id} [死亡]</Text>
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
          <Text className='sectionTitle'>死亡公告 [Death Announcements]</Text>
          {deathAnnouncements.map((announcement, index) => (
            <Text className='deathAnnouncement' key={`${announcement}-${index}`}>{announcement}</Text>
          ))}
        </View>
      )}

      {/* ─── Night Phase UI (Storyteller only) ─────────────── */}
      {gamePhase === 'night' && isStoryteller && (
        <View className='card'>
          <Text className='sectionTitle'>夜间行动 [Night Actions]</Text>

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
          {selectedNightAction.needsTarget && (
            <View className='targetSelector'>
              <Text className='hint'>选择目标:</Text>
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
            <Button className='button primary' onClick={submitNightAction}>提交行动</Button>
            <Button className='button warn' onClick={resolveNight}>结束夜晚</Button>
          </View>

          {/* Night actions recorded this night */}
          {nightActions.length > 0 && (
            <View className='nightActionLog'>
              <Text className='sectionTitle'>今夜行动记录</Text>
              {nightActions.map((action, index) => {
                const actionLabel = NIGHT_ACTION_TYPES.find((a) => a.id === action.actionType)?.label ?? action.actionType;
                const targetNames = action.targetIds.map(
                  (tid) => roomState?.players.find((player) => player.id === tid)?.name ?? tid,
                );
                return (
                  <View className='nightActionEntry' key={index}>
                    <Text className='hint'>
                      {actionLabel}
                      {targetNames.length > 0 ? ` -> ${targetNames.join(', ')}` : ''}
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
          <Text className='sectionTitle'>夜晚 [Night]</Text>
          <Text className='hint'>夜晚降临... 请闭上眼睛。</Text>
          <Text className='hint'>Storyteller 正在处理夜间行动，请耐心等待。</Text>
        </View>
      )}

      {errorMessage && (
        <View className='error'>
          <Text>{errorMessage}</Text>
        </View>
      )}

      <View className='card'>
        <Text className='sectionTitle'>日志 [Logs]</Text>
        {logs.map((log, index) => (
          <Text className='log' key={`${log}-${index}`}>{log}</Text>
        ))}
      </View>
    </ScrollView>
  );
}
