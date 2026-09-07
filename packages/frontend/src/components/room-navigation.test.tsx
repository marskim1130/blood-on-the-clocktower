import React from 'react';
import { createRequire } from 'node:module';
import { expect, it, vi } from 'vitest';
import { RoomNavigation } from './room-navigation';
const { renderToStaticMarkup } = createRequire(import.meta.url)('react-dom/server') as { renderToStaticMarkup(node: React.ReactNode): string };
vi.mock('@tarojs/components', () => ({ View: 'div', Text: 'span', Button: 'button' }));
it('offers four distinct table-app destinations and identifies the selected destination', () => {
  const html = renderToStaticMarkup(React.createElement(RoomNavigation, { active: 'main', hasRoom: true, approvalCount: 2, onSelect: () => {} }));
  expect(html).toContain('主桌');
  expect(html).toContain('图鉴');
  expect(html).toContain('记录');
  expect(html).toContain('更多');
  expect(html.match(/aria-pressed="true"/g)).toHaveLength(1);
  expect(html).toContain('2');
});
