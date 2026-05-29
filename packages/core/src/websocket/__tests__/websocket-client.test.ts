import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { WebSocketServer, WebSocket as WsWebSocket } from 'ws';
import { GameWebSocketClient } from '../index.js';
import type { ServerMessage } from '../index.js';

// Use ws package as mock server since vitest runs in Node
const originalWebSocket = globalThis.WebSocket;

function createMockServer(port: number) {
  const wss = new WebSocketServer({ port });

  const rooms = new Map<string, Set<WsWebSocket>>();

  wss.on('connection', (ws) => {
    let currentRoom: string | null = null;

    ws.on('message', (data) => {
      const msg = JSON.parse(data.toString());

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
    expect(messages[0].type).toBe('ROOM_STATE');
    expect(messages[0].roomId).toBe('room-1');

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
