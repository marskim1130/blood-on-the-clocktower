import React from 'react';
import { createRequire } from 'node:module';
import { expect, it, vi } from 'vitest';
import type { RoomState } from '@clocktower/core';
import { GameHistoryPanel } from './game-history-panel';

const { renderToStaticMarkup } = createRequire(import.meta.url)('react-dom/server') as { renderToStaticMarkup(node: React.ReactNode): string };
vi.mock('@tarojs/components', () => ({ View: 'div', Text: 'span', Button: 'button', ScrollView: 'div' }));
vi.mock('@tarojs/taro', () => ({ default: { showModal: vi.fn() } }));

it('only offers private history and rollback controls to the storyteller', () => {
  const room = { roomId: 'room', storytellerId: 'st', players: [], phase: 3, canUndo: true, operationLog: [{ id: 1, action: 'ConfirmNightAction', actorId: 'st', createdAtUnixMs: 1, fromPhase: 3, toPhase: 3, disclosureWarning: true }] } as unknown as RoomState;
  const render = (actorId: string) => renderToStaticMarkup(React.createElement(GameHistoryPanel, { room, actorId, busy: false, onUndo: () => {}, onRedo: () => {} }));
  expect(render('player')).toBe('');
  expect(render('st')).toContain('撤销上一步');
  expect(render('st')).toContain('确认夜间信息');
});
