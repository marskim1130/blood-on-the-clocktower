import { describe, expect, it } from 'vitest';
import {
  canAssignCharacters,
  membershipKey,
  moveSeatClockwise,
  normalizeDemonBluffs,
  summarizeCharacterConfirmation,
  summarizeReadiness,
} from './utils';

describe('game setup membership', () => {
  it('does not treat a clockwise seat reorder as a membership change', () => {
    const before = membershipKey([{ id: 'p1' }, { id: 'p2' }, { id: 'p3' }]);
    const after = membershipKey([{ id: 'p3' }, { id: 'p1' }, { id: 'p2' }]);

    expect(after).toBe(before);
  });

  it('moves the last seated player clockwise into the first seat', () => {
    expect(moveSeatClockwise(['p1', 'p2', 'p3'], 'p3')).toEqual(['p3', 'p1', 'p2']);
  });

  it('summarizes authoritative readiness for the current player and the room', () => {
    expect(summarizeReadiness([
      { id: 'p1', isReady: true },
      { id: 'p2', isReady: false },
    ], 'p1')).toEqual({
      readyCount: 1,
      totalCount: 2,
      allReady: false,
      selfReady: true,
    });
  });

  it('does not allow assignment while any seated player is unready', () => {
    expect(canAssignCharacters(true, false)).toBe(false);
  });

  it('summarizes authoritative character confirmations before the first night', () => {
    expect(summarizeCharacterConfirmation([
      { id: 'p1', hasConfirmedCharacter: true },
      { id: 'p2', hasConfirmedCharacter: false },
    ], 'p2')).toEqual({
      confirmedCount: 1,
      totalCount: 2,
      allConfirmed: false,
      selfConfirmed: false,
    });
  });

  it('keeps valid Demon bluffs and fills invalidated slots from eligible characters', () => {
    expect(normalizeDemonBluffs(
      ['chef', 'empath', 'fortuneteller', 'monk'],
      ['chef', 'in-play', 'chef'],
    )).toEqual(['chef', 'empath', 'fortuneteller']);
  });
});
