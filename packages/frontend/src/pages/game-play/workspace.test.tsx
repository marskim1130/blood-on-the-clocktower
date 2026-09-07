import React from 'react';
import { createRequire } from 'node:module';
import { beforeEach, expect, it, vi } from 'vitest';
import GamePlayPage from './index';

const { renderToStaticMarkup } = createRequire(import.meta.url)('react-dom/server') as {
  renderToStaticMarkup(node: React.ReactNode): string;
};
const fixture = vi.hoisted(() => ({ state: {} as Record<string, unknown>, character: null as unknown }));
vi.mock('@tarojs/components', () => ({
  View: ({ ariaLabel, hoverClass: _hoverClass, ...props }: Record<string, unknown>) => React.createElement('div', { ...props, 'aria-label': ariaLabel }),
  Text: 'span', Button: 'button', Picker: 'div', Textarea: 'textarea',
}));
vi.mock('@tarojs/taro', () => ({ default: {}, useDidHide: () => {} }));
vi.mock('../../lib/use-session-route', () => ({ useSessionRoute: () => {} }));
vi.mock('../../lib/room-session-store', () => ({ useRoomSession: (selector: (state: Record<string, unknown>) => unknown) => selector(fixture.state) }));
vi.mock('../../components/session-shell', () => ({
  LoadingState: () => null,
  SessionShell: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));
vi.mock('../../lib/room-experience', () => ({
  selectGamePhase: () => 'night',
  selectDeathRecords: () => ({}),
  selectGhostVotes: () => new Set(),
  selectVisibleCharacter: () => fixture.character,
  selectActingPlayerId: () => 'p1',
}));

beforeEach(() => {
  fixture.character = null;
  fixture.state = {
    playerId: 'st', pendingCommand: null,
    experience: { roomState: {
      roomId: 'room', storytellerId: 'st', creatorId: 'st', dayNumber: 1, nightNumber: 1,
      players: Array.from({ length: 15 }, (_, index) => ({ id: `p${index + 1}`, name: `座位玩家${index + 1}`, isAlive: true })),
      nightWakeSteps: [], currentNightWakeIndex: 0,
    } },
  };
});

it('opens on the current action instead of the fifteen-player roster or management tools', () => {
  const html = renderToStaticMarkup(<GamePlayPage />);
  expect(html).toContain('当前夜间行动');
  expect(html).toContain('生成黎明死亡建议');
  expect(html).toContain('pageTabActive');
  expect(html).not.toContain('seatCard');
  expect(html).not.toContain('座位玩家15');
  expect(html).not.toContain('关闭房间');
});

it('keeps the private character masked in the player action workspace', () => {
  fixture.state.playerId = 'p1';
  fixture.character = { id: 'imp', name: '私密测试恶魔', ability: '私密测试能力', team: 2 };
  const html = renderToStaticMarkup(<GamePlayPage />);
  expect(html).toContain('夜晚降临');
  expect(html).toContain('你的身份');
  expect(html).not.toContain('私密测试恶魔');
  expect(html).not.toContain('私密测试能力');
  expect(html).not.toContain('小恶魔');
  expect(html).not.toContain('说书人</button>');
});
