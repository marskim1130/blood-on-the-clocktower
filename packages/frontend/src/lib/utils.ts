import Taro from '@tarojs/taro';
export { mapProtocolPhase, normalizeWinner, type GamePhase } from './room-experience';

export type DeathCause = 'execution' | 'night_kill' | 'ability';

export interface InputEvent {
  readonly detail: {
    readonly value: string;
  };
}

export interface StoredRoomIdentity {
  readonly version: 2;
  readonly roomId: string;
  readonly playerId: string;
  readonly resumeCredential: string;
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

export function removeStoredValue(key: string): void {
  Taro.removeStorageSync(key);
}

export function getStoredRoomIdentity(key: string): StoredRoomIdentity | null {
  const stored = Taro.getStorageSync<unknown>(key);
  if (!stored || typeof stored !== 'object') return null;

  const identity = stored as Partial<StoredRoomIdentity>;
  if (
    identity.version !== 2 ||
    typeof identity.roomId !== 'string' || !identity.roomId.trim() ||
    typeof identity.playerId !== 'string' || !identity.playerId.trim() ||
    typeof identity.resumeCredential !== 'string' || !identity.resumeCredential.trim()
  ) {
    return null;
  }

  return {
    version: 2,
    roomId: identity.roomId,
    playerId: identity.playerId,
    resumeCredential: identity.resumeCredential,
  };
}

export function persistRoomIdentity(key: string, identity: StoredRoomIdentity): void {
  Taro.setStorageSync(key, identity);
}

export function getOrCreatePlayerId(): string {
  const stored = Taro.getStorageSync<string>('clocktower.playerId');
  if (stored) return stored;

  const generated = `player_${Math.random().toString(36).slice(2, 10)}`;
  Taro.setStorageSync('clocktower.playerId', generated);
  return generated;
}
