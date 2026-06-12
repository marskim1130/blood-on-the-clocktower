import { describe, it, expect, beforeAll, beforeEach, afterAll } from 'vitest';
import { WebSocketServer, WebSocket as WsWebSocket } from 'ws';
import { GameWebSocketClient } from '../index.js';
import type { ServerMessage } from '../index.js';

// Use ws package as mock server since vitest runs in Node
const originalWebSocket = globalThis.WebSocket;
let receivedMessages: Array<Record<string, unknown>> = [];

function createMockServer(port: number) {
  const wss = new WebSocketServer({ port });

  const rooms = new Map<string, Set<WsWebSocket>>();

  wss.on('connection', (ws) => {
    let currentRoom: string | null = null;

    ws.on('message', (data) => {
      const msg = JSON.parse(data.toString());
      receivedMessages.push(msg);

      if (msg.type === 'CREATE_ROOM') {
        currentRoom = `${msg.playerId}-room`;
        if (!rooms.has(currentRoom)) rooms.set(currentRoom, new Set());
        rooms.get(currentRoom)!.add(ws);

        const state: ServerMessage = {
          type: 'ROOM_STATE',
          roomId: currentRoom,
          state: { roomId: currentRoom, players: [] },
        };
        ws.send(JSON.stringify(state));
      }

      if (msg.type === 'JOIN_ROOM') {
        currentRoom = msg.roomId;
        if (!rooms.has(msg.roomID)) rooms.set(msg.roomId, new Set());
        rooms.get(msg.roomId)!.add(ws);

        // Send ROOM_STATE
        const state: ServerMessage = {
          type: 'ROOM_STATE',
          roomId: msg.roomId,
          state: { roomId: msg.roomId, players: [] },
        };
        ws.send(JSON.stringify(state));
      }
    });

    ws.on('close', () => {
      if (currentRoom) rooms.get(currentRoom)?.delete(ws);
    });
  });

  return wss;
}

describe('GameWebSocketClient', () => {
  let server: WebSocketServer;
  const port = 18765;

  beforeAll(async () => {
    // Patch global WebSocket for Node environment
    (globalThis as Record<string, unknown>).WebSocket = WsWebSocket;
    server = createMockServer(port);
  });

  beforeEach(() => {
    receivedMessages = [];
  });

  afterAll(() => {
    server.close();
    (globalThis as Record<string, unknown>).WebSocket = originalWebSocket;
  });

  it('connects to server and reports status changes', async () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      maxReconnectAttempts: 0,
    });

    const statuses: string[] = [];
    client.onStatusChange((s) => statuses.push(s));

    client.connect();

    await new Promise<void>((resolve) => {
      client.onStatusChange((s) => {
        if (s === 'connected') resolve();
      });
    });

    expect(statuses).toContain('connecting');
    expect(statuses).toContain('connected');
    expect(client.status).toBe('connected');

    client.disconnect();
    expect(client.status).toBe('disconnected');
  });

  it('joins a room and receives ROOM_STATE', async () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      maxReconnectAttempts: 0,
    });

    client.connect();
    await new Promise<void>((resolve) => {
      client.onStatusChange((s) => {
        if (s === 'connected') resolve();
      });
    });

    const messages: ServerMessage[] = [];
    client.onMessage((msg) => messages.push(msg));

    client.joinRoom('room-1', 'p1', 'Alice');

    // Wait for ROOM_STATE response
    await new Promise((r) => setTimeout(r, 100));

    expect(messages).toHaveLength(1);
    expect(messages[0]!.type).toBe('ROOM_STATE');
    expect(messages[0]!.roomId).toBe('room-1');

    client.disconnect();
  });

  it('rejoins the created room after reconnecting', async () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      reconnectInterval: 10,
      maxReconnectAttempts: 2,
    });

    client.connect();
    await waitFor(() => client.status === 'connected');

    const messages: ServerMessage[] = [];
    client.onMessage((msg) => messages.push(msg));

    client.createRoom('creator', 'Alice', 5);
    await waitFor(() => messages.some((msg) => msg.type === 'ROOM_STATE' && msg.roomId === 'creator-room'));

    server.clients.forEach((socket) => socket.close());

    await waitFor(() =>
      receivedMessages.some(
        (msg) =>
          msg.type === 'JOIN_ROOM' &&
          msg.roomId === 'creator-room' &&
          msg.playerId === 'creator' &&
          msg.playerName === 'Alice'
      )
    );

    client.disconnect();
  });

  it('sends script id when creating a room', async () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      maxReconnectAttempts: 0,
    });

    client.connect();
    await waitFor(() => client.status === 'connected');

    client.createRoom('creator', 'Alice', 5, 'trouble_brewing');
    await waitFor(() =>
      receivedMessages.some(
        (msg) =>
          msg.type === 'CREATE_ROOM' &&
          msg.playerId === 'creator' &&
          msg.scriptId === 'trouble_brewing'
      )
    );

    client.disconnect();
  });

  it('sends manual end game command', async () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      maxReconnectAttempts: 0,
    });

    client.connect();
    await waitFor(() => client.status === 'connected');

    client.endGame('evil', 'The Storyteller called the game.');
    await waitFor(() =>
      receivedMessages.some(
        (msg) =>
          msg.type === 'END_GAME' &&
          msg.winner === 'evil' &&
          msg.reason === 'storyteller_decision' &&
          msg.description === 'The Storyteller called the game.'
      )
    );

    client.disconnect();
  });

  it('sends kick player command', async () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      maxReconnectAttempts: 0,
    });

    client.connect();
    await waitFor(() => client.status === 'connected');

    client.kickPlayer('p2');
    await waitFor(() =>
      receivedMessages.some(
        (msg) =>
          msg.type === 'KICK_PLAYER' &&
          msg.targetPlayerId === 'p2'
      )
    );

    client.disconnect();
  });

  it('does not resume a room after receiving kicked error', async () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      reconnectInterval: 10,
      maxReconnectAttempts: 1,
    });

    client.connect();
    await waitFor(() => client.status === 'connected');

    const messages: ServerMessage[] = [];
    client.onMessage((msg) => messages.push(msg));
    client.joinRoom('room-kick', 'p1', 'Alice');
    await waitFor(() => messages.some((msg) => msg.type === 'ROOM_STATE' && msg.roomId === 'room-kick'));

    receivedMessages = [];
    server.clients.forEach((socket) => {
      socket.send(JSON.stringify({ type: 'ERROR', error: 'kicked from room' }));
      socket.close();
    });
    await new Promise((resolve) => setTimeout(resolve, 200));

    expect(receivedMessages.some((msg) => msg.type === 'JOIN_ROOM')).toBe(false);

    client.disconnect();
  });

  it('throws when sending on disconnected client', () => {
    const client = new GameWebSocketClient({
      url: `ws://localhost:${port}`,
      maxReconnectAttempts: 0,
    });

    expect(() => client.send({ type: 'JOIN_ROOM', roomId: 'r1' })).toThrow(
      'WebSocket is not connected'
    );
  });
});

async function waitFor(predicate: () => boolean): Promise<void> {
  const deadline = Date.now() + 2000;
  while (Date.now() < deadline) {
    if (predicate()) return;
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  throw new Error('condition was not met before timeout');
}
