import { expect, it } from 'vitest';
import { GameWebSocketClient, type WebSocketTransport } from '../index';

function harness() {
  const sent: Array<Record<string, unknown>> = [];
  let receive: (data: string) => void = () => {};
  let opened = () => {};
  const transport: WebSocketTransport = {
    readyState: 1, connect: () => opened(), close: () => {},
    send: data => { sent.push(JSON.parse(data)); },
    onOpen: handler => { opened = handler; }, onMessage: handler => { receive = handler; },
    onClose: () => {}, onError: () => {},
  };
  const client = new GameWebSocketClient({ url: 'ws://test', transportFactory: () => transport });
  client.connect();
  return { client, sent, receive: (message: Record<string, unknown>) => receive(JSON.stringify(message)) };
}

it('审批前恢复请求不建立身份，批准后才用新凭据恢复', () => {
  const { client, sent, receive } = harness();
  client.requestRecovery('room', 'player', 'recovery-secret', 'request-1');
  expect(client.identity).toBeNull();
  expect(sent.at(-1)).toMatchObject({ type: 'REQUEST_RECOVERY', requestId: 'request-1', recoveryCredential: 'recovery-secret' });
  receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-1', recoveryStatus: 'approved', roomId: 'room', playerId: 'player', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(client.identity).toEqual({ roomId: 'room', playerId: 'player', resumeCredential: 'new-device', recoveryCredential: 'new-recovery' });
  expect(sent.at(-1)).toMatchObject({ type: 'RESUME_ROOM', resumeCredential: 'new-device' });
});

it('等待审批期间即使收到投影也不交给页面', () => {
  const { client, receive } = harness();
  const messages: unknown[] = [];
  client.onMessage(message => messages.push(message));
  client.requestRecovery('room', 'player', 'secret', 'request-2');
  receive({ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-2', recoveryStatus: 'pending', state: { secret: 'role' }, roomRevision: 4 });
  expect(messages).toEqual([{ type: 'RECOVERY_STATUS', recoveryRequestId: 'request-2', recoveryStatus: 'pending' }]);
});

it('重连重发同一恢复申请而不会尝试直接恢复', () => {
  const { client, sent } = harness();
  client.requestRecovery('room', 'player', 'secret', 'stable-request');
  client.disconnect();
  client.connect();
  expect(sent.map(message => message.type)).toEqual(['REQUEST_RECOVERY', 'REQUEST_RECOVERY']);
  expect(sent[0]).toEqual(sent[1]);
});

it('设备被替换后清除身份并停止恢复', () => {
  const { client, receive } = harness();
  client.resumeRoom('room', 'player', 'old-device');
  receive({ type: 'SESSION_REPLACED' });
  expect(client.identity).toBeNull();
  expect(client.status).toBe('disconnected');
  expect(client.hasPendingRequest).toBe(false);
});
