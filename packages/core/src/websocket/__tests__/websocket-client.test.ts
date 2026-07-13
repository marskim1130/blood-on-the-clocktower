import { afterAll, beforeAll, beforeEach, describe, expect, it } from 'vitest';
import { WebSocket as WsWebSocket, WebSocketServer } from 'ws';
import { GameWebSocketClient } from '../index.js';
import type { ServerMessage } from '../index.js';

const originalWebSocket = globalThis.WebSocket;
const port = 18765;
let receivedMessages: Array<Record<string, unknown>> = [];
let server: WebSocketServer;
let rejectNextSequencedCommand = false;

function send(socket: WsWebSocket, message: ServerMessage): void {
  socket.send(JSON.stringify(message));
}

function createMockServer(): WebSocketServer {
  const mockServer = new WebSocketServer({ port });
  mockServer.on('connection', (socket) => {
    socket.on('message', (data) => {
      const message = JSON.parse(data.toString()) as Record<string, unknown>;
      receivedMessages.push(message);

      if (message.type === 'CREATE_ROOM') {
        send(socket, {
          type: 'CREATE_ROOM_RESULT',
          roomId: `${message.playerId}-room`,
          resumeCredential: 'creator-credential',
          nextClientSequence: 1,
          roomRevision: 1,
        });
      } else if (message.type === 'JOIN_ROOM') {
        send(socket, {
          type: 'JOIN_ROOM_RESULT',
          roomId: String(message.roomId),
          resumeCredential: 'member-credential',
          nextClientSequence: 1,
          roomRevision: 2,
        });
      } else if (message.type === 'RESUME_ROOM') {
        send(socket, {
          type: 'RESUME_ROOM_RESULT',
          roomId: String(message.roomId),
          nextClientSequence: 1,
          roomRevision: 2,
        });
      } else if (
        typeof message.clientSequence === 'number' &&
        message.type !== 'FAIL_COMMAND'
      ) {
        if (rejectNextSequencedCommand) {
          rejectNextSequencedCommand = false;
          send(socket, {
            type: 'ERROR',
            code: 'INTERNAL',
            error: 'rejected',
            nextClientSequence: Number(message.clientSequence) + 1,
          });
          return;
        }
        send(socket, {
          type: 'COMMAND_RESULT',
          acceptedSequence: message.clientSequence,
          nextClientSequence: message.clientSequence + 1,
          roomRevision: Number(message.clientSequence) + 2,
        });
      }
    });
  });
  return mockServer;
}

function createClient(options: Partial<ConstructorParameters<typeof GameWebSocketClient>[0]> = {}): GameWebSocketClient {
  return new GameWebSocketClient({
    url: `ws://localhost:${port}`,
    maxReconnectAttempts: 0,
    requestIdFactory: () => 'generated-request-id',
    ...options,
  });
}

async function connect(client: GameWebSocketClient): Promise<void> {
  client.connect();
  await waitFor(() => client.status === 'connected');
}

async function createIdentity(client: GameWebSocketClient): Promise<void> {
  client.createRoom('creator', 'Alice', 5, 'trouble_brewing', 'create-request');
  await waitFor(() => client.identity !== null);
}

describe('GameWebSocketClient protocol v2', () => {
  beforeAll(() => {
    (globalThis as Record<string, unknown>).WebSocket = WsWebSocket;
    server = createMockServer();
  });

  beforeEach(() => {
    receivedMessages = [];
    rejectNextSequencedCommand = false;
  });

  afterAll(() => {
    server.close();
    (globalThis as Record<string, unknown>).WebSocket = originalWebSocket;
  });

  it('adds protocolVersion and create request id, then stores identity metadata', async () => {
    const client = createClient();
    await connect(client);

    client.createRoom('creator', 'Alice', 5, 'trouble_brewing', 'create-1');
    await waitFor(() => client.identity !== null);

    expect(receivedMessages[0]).toMatchObject({
      protocolVersion: 2,
      type: 'CREATE_ROOM',
      requestId: 'create-1',
      playerId: 'creator',
      scriptId: 'trouble_brewing',
    });
    expect(client.identity).toEqual({
      roomId: 'creator-room',
      playerId: 'creator',
      resumeCredential: 'creator-credential',
    });
    expect(client.nextSequence).toBe(1);
    expect(client.currentRoomRevision).toBe(1);
    client.disconnect();
  });

  it('adds joinRequestId and stores the joined identity', async () => {
    const client = createClient();
    await connect(client);

    client.joinRoom('room-1', 'p1', 'Alice', 'join-1');
    await waitFor(() => client.identity !== null);

    expect(receivedMessages[0]).toMatchObject({
      protocolVersion: 2,
      type: 'JOIN_ROOM',
      joinRequestId: 'join-1',
      roomId: 'room-1',
      playerId: 'p1',
    });
    expect(client.identity?.resumeCredential).toBe('member-credential');
    client.disconnect();
  });

  it('uses RESUME_ROOM with the saved credential after reconnecting', async () => {
    const client = createClient({ reconnectInterval: 10, maxReconnectAttempts: 2 });
    await connect(client);
    await createIdentity(client);
    receivedMessages = [];

    server.clients.forEach((socket) => socket.close());
    await waitFor(() => receivedMessages.some((message) => message.type === 'RESUME_ROOM'));

    expect(receivedMessages.find((message) => message.type === 'RESUME_ROOM')).toMatchObject({
      protocolVersion: 2,
      roomId: 'creator-room',
      playerId: 'creator',
      resumeCredential: 'creator-credential',
    });
    client.disconnect();
  });

  it('does not replay an already committed command after resume advances the sequence', async () => {
    const client = createClient();
    await connect(client);
    await createIdentity(client);
    receivedMessages = [];

    client.startGame();
    await waitFor(() => client.nextSequence === 2);
    const commandCount = receivedMessages.filter((message) => message.type === 'START_GAME').length;
    server.clients.forEach((socket) => send(socket, {
      type: 'RESUME_ROOM_RESULT',
      roomId: 'creator-room',
      nextClientSequence: 2,
      roomRevision: 3,
    }));
    await new Promise((resolve) => setTimeout(resolve, 20));

    expect(receivedMessages.filter((message) => message.type === 'START_GAME')).toHaveLength(commandCount);
    client.disconnect();
  });

  it('attaches one sequence at a time and advances only after success', async () => {
    const client = createClient();
    await connect(client);
    await createIdentity(client);
    receivedMessages = [];

    client.updateRoomSettings(7, 'trouble_brewing');
    client.startGame();

    await waitFor(() => receivedMessages.filter((message) => message.clientSequence !== undefined).length === 2);
    const commands = receivedMessages.filter((message) => message.clientSequence !== undefined);
    expect(commands[0]).toMatchObject({
      type: 'UPDATE_ROOM_SETTINGS',
      clientSequence: 1,
      resumeCredential: 'creator-credential',
    });
    expect(commands[1]).toMatchObject({ type: 'START_GAME', clientSequence: 2 });
    await waitFor(() => client.nextSequence === 3);
    expect(client.currentRoomRevision).toBe(4);
    client.disconnect();
  });

  it('does not advance the sequence for an error response', async () => {
    const client = createClient();
    await connect(client);
    await createIdentity(client);

    rejectNextSequencedCommand = true;
    client.startGame();
    await waitFor(() => receivedMessages.some((message) => message.type === 'START_GAME'));
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(client.nextSequence).toBe(1);
    client.disconnect();
  });

  it('sends new lifecycle and query commands with the correct sequencing rules', async () => {
    const client = createClient();
    await connect(client);
    await createIdentity(client);
    receivedMessages = [];

    client.getRoomState();
    client.rejoinRoom();
    client.closeRoom();

    await waitFor(() => receivedMessages.some((message) => message.type === 'CLOSE_ROOM'));
    expect(receivedMessages.find((message) => message.type === 'GET_ROOM_STATE')).toMatchObject({
      protocolVersion: 2,
      resumeCredential: 'creator-credential',
    });
    expect(receivedMessages.find((message) => message.type === 'REJOIN_ROOM')).toMatchObject({ clientSequence: 1 });
    expect(receivedMessages.find((message) => message.type === 'CLOSE_ROOM')).toMatchObject({ clientSequence: 2 });
    client.disconnect();
  });

  it('clears saved identity after KICKED', async () => {
    const client = createClient();
    await connect(client);
    await createIdentity(client);

    server.clients.forEach((socket) => send(socket, { type: 'KICKED', roomId: 'creator-room' }));
    await waitFor(() => client.identity === null);

    expect(client.nextSequence).toBeNull();
    expect(() => client.getRoomState()).toThrow('Room identity is not available');
    client.disconnect();
  });

  it('preserves the transport seam for raw protocol messages', async () => {
    const client = createClient();
    await connect(client);

    client.send({ type: 'GET_ROOM_STATE', roomId: 'r1', playerId: 'p1', resumeCredential: 'credential' });
    await waitFor(() => receivedMessages.length === 1);

    expect(receivedMessages[0]?.protocolVersion).toBe(2);
    client.disconnect();
  });

  it('filters duplicate, gapped, and stale room projections before handlers', async () => {
    const client = createClient();
    const handled: ServerMessage[] = [];
    client.onMessage((message) => handled.push(message));
    await connect(client);
    await createIdentity(client);
    receivedMessages = [];

    const emptyState = {
      roomId: 'creator-room',
      players: [],
      maxPlayers: 5,
      scriptId: 'trouble_brewing',
      scriptName: 'Trouble Brewing',
      phase: 1,
      dayNumber: 0,
    };

    server.clients.forEach((socket) => {
      send(socket, { type: 'ROOM_STATE_CHANGED', roomRevision: 1, state: emptyState });
      send(socket, { type: 'ROOM_STATE_CHANGED', roomRevision: 3, state: emptyState });
      send(socket, { type: 'ROOM_STATE', roomRevision: 2, state: emptyState });
      send(socket, { type: 'ROOM_STATE', roomRevision: 4, state: emptyState });
    });

    await waitFor(() => handled.filter((message) => message.type === 'ROOM_STATE').length === 2);
    const duplicate = handled.find((message) => message.type === 'ROOM_STATE_CHANGED' && message.roomRevision === undefined);
    const gapped = handled.filter((message) => message.type === 'ROOM_STATE_CHANGED' && message.state === undefined);
    expect(duplicate?.state).toBeUndefined();
    expect(gapped).toHaveLength(2);
    expect(client.currentRoomRevision).toBe(4);
    expect(receivedMessages.filter((message) => message.type === 'GET_ROOM_STATE')).toHaveLength(1);
    client.disconnect();
  });

  it('throws when sending on a disconnected client', () => {
    const client = createClient();
    expect(() => client.send({ type: 'GET_ROOM_STATE' })).toThrow('WebSocket is not connected');
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
