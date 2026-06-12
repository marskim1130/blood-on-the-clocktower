import Taro from '@tarojs/taro';

export type GamePhase = 'setup' | 'day' | 'voting' | 'night' | 'finished';

export type DeathCause = 'execution' | 'night_kill' | 'ability';

export interface InputEvent {
  readonly detail: {
    readonly value: string;
  };
}

export function mapProtocolPhase(phase: number | undefined): GamePhase | null {
  if (phase === undefined) return null;
  const phaseMap: Record<number, GamePhase> = {
    0: 'setup',
    1: 'setup',
    2: 'day',
    3: 'night',
    4: 'voting',
    5: 'finished',
  };
  return phaseMap[phase] ?? null;
}

export function normalizeWinner(winner: number | string): string {
  if (winner === 1 || winner === 'good') return 'good';
  if (winner === 2 || winner === 'evil') return 'evil';
  return String(winner);
}

export function eventValue(event: InputEvent): string {
  return event.detail.value;
}

export function getStoredString(key: string, fallback = ''): string {
  const stored = Taro.getStorageSync<string>(key);
  return typeof stored === 'string' && stored.trim() ? stored : fallback;
}

export function persistString(key: string, value: string): void {
  Taro.setStorageSync(key, value);
}

export function getOrCreatePlayerId(): string {
  const stored = Taro.getStorageSync<string>('clocktower.playerId');
  if (stored) return stored;

  const generated = `player_${Math.random().toString(36).slice(2, 10)}`;
  Taro.setStorageSync('clocktower.playerId', generated);
  return generated;
}
