import { Button, Input, ScrollView, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect, useMemo, useRef, useState } from 'react';
import { GameWebSocketClient } from '@clocktower/core';
import type { ConnectionStatus, GameCharacter, RoomState, ServerMessage } from '@clocktower/core';
import { createTaroWebSocketTransport } from '../../lib/taro-websocket-transport';
import './index.css';

const PLAYER_ID_STORAGE_KEY = 'clocktower.playerId';
const DEFAULT_WS_URL = 'ws://localhost:8080/ws';

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

    if (message.event && 'playerJoined' in message.event) {
      const joinedPlayer = message.event.playerJoined.player;
      setRoomState((current) => {
        if (!current || current.players.some((player) => player.id === joinedPlayer.id)) return current;
        return { ...current, players: [...current.players, joinedPlayer] };
      });
      return;
    }

    if (message.event && 'playerLeft' in message.event) {
      const leftPlayerId = message.event.playerLeft.playerId;
      setRoomState((current) => {
        if (!current) return current;
        return { ...current, players: current.players.filter((player) => player.id !== leftPlayerId) };
      });
      return;
    }

    if (message.event && 'characterAssigned' in message.event) {
      const assignment = message.event.characterAssigned;
      if (assignment.playerId === playerId) {
        setMyCharacter(assignment.character);
      }
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

  return (
    <ScrollView className='page' scrollY>
      <View className='hero'>
        <Text className='title'>血染钟楼 MVP 调试台</Text>
        <Text className='subtitle'>创建房间、加入房间、指定 Storyteller、一键分配角色</Text>
      </View>

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
        {roomState?.players.map((player) => (
          <View className='player' key={player.id}>
            <View className='playerInfo'>
              <Text className='playerName'>{player.name || player.id}</Text>
              <Text className='mono'>{player.id}</Text>
              <Text className='hint'>角色：{player.character?.name ?? '隐藏/未分配'}</Text>
            </View>
            {!roomState.storytellerId && (
              <Button className='miniButton' onClick={() => setStoryteller(player.id)}>设为 ST</Button>
            )}
          </View>
        ))}
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
