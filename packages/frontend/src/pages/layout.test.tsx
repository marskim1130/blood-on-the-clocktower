import React from 'react';
import { createRequire } from 'node:module';
import { beforeEach, expect, it, vi } from 'vitest';
import GameSetupPage from './game-setup';
import GameOverPage from './game-over';
import ScriptsPage from './scripts';
import LobbyPage from './index';

const { renderToStaticMarkup } = createRequire(import.meta.url)('react-dom/server') as { renderToStaticMarkup(node: React.ReactNode): string };
const fixture = vi.hoisted(() => ({ state: {} as Record<string, unknown>, params: {} as { roomId?: string } }));
vi.mock('@tarojs/components', () => ({ View: 'div', Text: 'span', Button: 'button', Picker: 'div', Input: 'input' }));
vi.mock('@tarojs/taro', () => ({ default: {}, useShareAppMessage: vi.fn(), useDidHide: vi.fn(), useRouter: () => ({ params: fixture.params }) }));
vi.mock('../components/room-recovery-panel', () => ({ RoomRecoveryPanel: () => null }));
vi.mock('../lib/room-session-store', () => ({ isDevelopmentBuild: true, useRoomSession: (selector: (state: unknown) => unknown) => selector(fixture.state) }));
vi.mock('../lib/use-session-route', () => ({ useSessionRoute: vi.fn() }));
vi.mock('../components/session-shell', () => ({
  SessionShell: ({ title, children, actions }: { title: string; children: React.ReactNode; actions?: React.ReactNode }) => React.createElement('main', null, title, actions, children),
  LoadingState: () => React.createElement('span', null, 'loading'),
}));

beforeEach(() => {
  vi.stubGlobal('React', React);
  fixture.params = {};
  fixture.state = {
    initialized: true, status: 'connected', playerName: '朋友甲', endpoint: '', maxPlayers: 5,
    playerId: 'owner', pendingCommand: null,
    experience: { roomState: { roomId: 'room', creatorId: 'owner', storytellerId: '', maxPlayers: 5, players: [{ id: 'owner', name: '朋友甲', isAlive: true, isReady: false }, { id: 'friend', name: '朋友乙', isAlive: true, isReady: false }] }, identityStatus: { participantSetFrozen: false } },
  };
});

it('lobby uses one shared nickname and only the selected entry form', () => {
  const html = renderToStaticMarkup(React.createElement(LobbyPage));
  expect(html.match(/输入桌上使用的昵称/g)).toHaveLength(1);
  expect(html).toContain('实际玩家上限');
  expect(html).not.toContain('输入邀请中的房间号');
  expect(html).toContain('lobbyEntryCard');
  expect(html).not.toContain('developerBand');
});

it('an invitation selects the join form without submitting or duplicating the nickname', () => {
  fixture.params = { roomId: 'invited-room' };
  const joinRoom = vi.fn();
  fixture.state.joinRoom = joinRoom;
  const html = renderToStaticMarkup(React.createElement(LobbyPage));
  expect(html.match(/输入桌上使用的昵称/g)).toHaveLength(1);
  expect(html).toContain('输入邀请中的房间号');
  expect(html).toContain('invited-room');
  expect(html).not.toContain('实际玩家上限');
  expect(joinRoom).not.toHaveBeenCalled();
});

it('setup defaults to a seat board, keeps immediate storyteller selection, and hides management controls', () => {
  const html = renderToStaticMarkup(React.createElement(GameSetupPage));
  expect(html).toContain('setupProgress');
  expect(html).toContain('seatGrid');
  expect(html).toContain('pageTabActive');
  expect(html).toContain('指定说书人');
  expect(html).not.toContain('移出朋友乙');
  expect(html).not.toContain('调整顺时针座次');
  expect(html).not.toContain('确认并发放身份');
});

it('the catalog starts with searchable role cards instead of expanded night orders and counts', () => {
  const html = renderToStaticMarkup(React.createElement(ScriptsPage));
  expect(html).toContain('搜索角色或能力关键词');
  expect(html).toContain('characterCardGrid');
  expect(html).toContain('洗衣妇');
  expect(html).not.toContain('wakeRow');
  expect(html).not.toContain('roleCountGrid');
});

it('game over keeps the grimoire private and the timelines collapsed initially', () => {
  const html = renderToStaticMarkup(React.createElement(GameOverPage));
  expect(html).toContain('resultHero');
  expect(html).toContain('等待说书人公开魔典');
  expect(html).toContain('展开对局时间线');
  expect(html).not.toContain('revealCard');
  expect(html).not.toContain('resultTimelines');
});
