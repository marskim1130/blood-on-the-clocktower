// @vitest-environment jsdom
import React from 'react';
import { createRequire } from 'node:module';
import { afterEach, expect, it, vi } from 'vitest';
import { RoomRecoveryPanel } from './room-recovery-panel';

const require = createRequire(import.meta.url);
interface TestButton { readonly textContent: string | null; readonly disabled: boolean; click(): void }
interface TestContainer { readonly textContent: string | null; querySelectorAll(selector: string): ArrayLike<TestButton> }
const testDocument = (globalThis as unknown as { document: { createElement(tag: string): TestContainer } }).document;
const { createRoot } = require('react-dom/client') as { createRoot(node: TestContainer): { render(node: React.ReactNode): void; unmount(): void } };
const { act } = require('react-dom/test-utils') as { act(callback: () => void): void };
const fixture = vi.hoisted(() => ({ state: {} as Record<string, unknown> }));
vi.mock('@tarojs/components', () => ({ View: 'div', Text: 'span', Button: 'button', Input: 'input' }));
vi.mock('@tarojs/taro', () => ({ default: {}, useDidHide: vi.fn() }));
vi.mock('../lib/room-session-store', () => ({ useRoomSession: (select: (state: Record<string, unknown>) => unknown) => select(fixture.state) }));

afterEach(() => { vi.unstubAllGlobals(); });

it('发送失败时只解锁同申请重试，等待及断线时不重复提交', () => {
  vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT', true);
  const retryRecovery = vi.fn();
  fixture.state = { identity: null, status: 'connected', pendingCommand: 'resume', recoveryStatus: 'failed', canRetryRecovery: true, retryRecovery };
  const container = testDocument.createElement('div');
  const root = createRoot(container);
  try {
    act(() => root.render(<RoomRecoveryPanel />));
    const retry = Array.from(container.querySelectorAll('button')).find(button => button.textContent === '重试恢复申请');
    expect(retry).toBeDefined();
    expect(retry?.disabled).toBe(false);
    expect(container.textContent).not.toContain('申请未获批准');
    act(() => retry?.click());
    expect(retryRecovery).toHaveBeenCalledTimes(1);
    fixture.state = { ...fixture.state, status: 'disconnected' };
    act(() => root.render(<RoomRecoveryPanel />));
    expect(Array.from(container.querySelectorAll('button')).find(button => button.textContent === '重试恢复申请')?.disabled).toBe(true);
    fixture.state = { ...fixture.state, status: 'connected', recoveryStatus: 'pending', canRetryRecovery: false };
    act(() => root.render(<RoomRecoveryPanel />));
    expect(container.textContent).not.toContain('重试恢复申请');
    expect(container.textContent).toContain('等待');
  } finally {
    act(() => root.unmount());
  }
});
