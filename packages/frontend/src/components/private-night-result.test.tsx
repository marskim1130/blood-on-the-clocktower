import React from 'react';
import { createRequire } from 'node:module';
import { expect, it, vi } from 'vitest';
import { PrivateNightResult } from './private-night-result';

const { renderToStaticMarkup } = createRequire(import.meta.url)('react-dom/server') as {
  renderToStaticMarkup(node: React.ReactNode): string;
};

vi.mock('@tarojs/components', () => ({ View: 'div', Text: 'span', Button: 'button' }));
vi.mock('@tarojs/taro', () => ({ useDidHide: vi.fn() }));

it('does not render private night information before the player holds to reveal it', () => {
  const html = renderToStaticMarkup(React.createElement(PrivateNightResult, { resultKey: 'night-1', roomId: 'room', playerName: '玩家', result: '秘密：恶魔是小明', targets: '小明、小红', onAcknowledge: () => {} }));
  expect(html).not.toContain('恶魔是小明');
  expect(html).not.toContain('小明、小红');
  expect(html).toContain('按住查看夜间信息');
  expect(html).toContain('disabled');
});
