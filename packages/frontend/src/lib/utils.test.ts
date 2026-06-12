import { describe, expect, it, vi, beforeEach } from 'vitest';

vi.mock('@tarojs/taro', () => ({
  default: {
    getStorageSync: vi.fn(),
    setStorageSync: vi.fn(),
  },
}));

import Taro from '@tarojs/taro';
import {
  mapProtocolPhase,
  normalizeWinner,
  eventValue,
  getStoredString,
  persistString,
  getOrCreatePlayerId,
} from './utils';

const mockedTaro = vi.mocked(Taro);

beforeEach(() => {
  vi.clearAllMocks();
});

describe('mapProtocolPhase', () => {
  it('maps protocol phase numbers to game phases', () => {
    expect(mapProtocolPhase(0)).toBe('setup');
    expect(mapProtocolPhase(1)).toBe('setup');
    expect(mapProtocolPhase(2)).toBe('day');
    expect(mapProtocolPhase(3)).toBe('night');
    expect(mapProtocolPhase(4)).toBe('voting');
    expect(mapProtocolPhase(5)).toBe('finished');
  });

  it('returns null for undefined', () => {
    expect(mapProtocolPhase(undefined)).toBeNull();
  });

  it('returns null for unknown phase numbers', () => {
    expect(mapProtocolPhase(99)).toBeNull();
    expect(mapProtocolPhase(-1)).toBeNull();
  });
});

describe('normalizeWinner', () => {
  it('normalizes numeric winner values', () => {
    expect(normalizeWinner(1)).toBe('good');
    expect(normalizeWinner(2)).toBe('evil');
  });

  it('normalizes string winner values', () => {
    expect(normalizeWinner('good')).toBe('good');
    expect(normalizeWinner('evil')).toBe('evil');
  });

  it('converts unknown values to string', () => {
    expect(normalizeWinner(3)).toBe('3');
    expect(normalizeWinner('draw')).toBe('draw');
  });
});

describe('eventValue', () => {
  it('extracts value from input event', () => {
    expect(eventValue({ detail: { value: 'hello' } })).toBe('hello');
  });

  it('returns empty string for empty value', () => {
    expect(eventValue({ detail: { value: '' } })).toBe('');
  });
});

describe('getStoredString', () => {
  it('returns stored value when present and non-empty', () => {
    mockedTaro.getStorageSync.mockReturnValue('stored-value');
    expect(getStoredString('key')).toBe('stored-value');
  });

  it('returns fallback when stored value is empty', () => {
    mockedTaro.getStorageSync.mockReturnValue('');
    expect(getStoredString('key', 'fallback')).toBe('fallback');
  });

  it('returns fallback when stored value is whitespace only', () => {
    mockedTaro.getStorageSync.mockReturnValue('   ');
    expect(getStoredString('key', 'fallback')).toBe('fallback');
  });

  it('returns empty string as default fallback', () => {
    mockedTaro.getStorageSync.mockReturnValue('');
    expect(getStoredString('key')).toBe('');
  });

  it('returns fallback when stored value is not a string', () => {
    mockedTaro.getStorageSync.mockReturnValue(42);
    expect(getStoredString('key', 'fallback')).toBe('fallback');
  });
});

describe('persistString', () => {
  it('calls setStorageSync with key and value', () => {
    persistString('myKey', 'myValue');
    expect(mockedTaro.setStorageSync).toHaveBeenCalledWith('myKey', 'myValue');
  });
});

describe('getOrCreatePlayerId', () => {
  it('returns existing player ID from storage', () => {
    mockedTaro.getStorageSync.mockReturnValue('existing-id');
    expect(getOrCreatePlayerId()).toBe('existing-id');
    expect(mockedTaro.setStorageSync).not.toHaveBeenCalled();
  });

  it('generates and stores a new player ID when none exists', () => {
    mockedTaro.getStorageSync.mockReturnValue('');
    const id = getOrCreatePlayerId();
    expect(id).toMatch(/^player_[a-z0-9]{8}$/);
    expect(mockedTaro.setStorageSync).toHaveBeenCalledWith('clocktower.playerId', id);
  });
});
