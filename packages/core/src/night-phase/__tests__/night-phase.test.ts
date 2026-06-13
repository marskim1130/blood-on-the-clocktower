import { describe, expect, it } from 'vitest';
import { getNextToWake, getWakeOrder, INITIAL_NIGHT_STATE } from '../index.js';
import type { Character, PlayerId } from '../../types/index.js';

function playerId(id: string): PlayerId {
  return id as PlayerId;
}

describe('night phase wake order', () => {
  it('matches the canonical Trouble Brewing Fortune Teller id', () => {
    const firstNight = getWakeOrder(1);
    const subsequentNight = getWakeOrder(2);

    expect(firstNight.some((entry) => entry.characterId === 'fortuneteller')).toBe(true);
    expect(subsequentNight.some((entry) => entry.characterId === 'fortuneteller')).toBe(true);
    expect(firstNight.some((entry) => entry.characterId === 'fortune_teller')).toBe(false);
    expect(subsequentNight.some((entry) => entry.characterId === 'fortune_teller')).toBe(false);
  });

  it('does not wake the Imp to kill on the first night', () => {
    const firstNight = getWakeOrder(1);

    expect(firstNight.some((entry) => entry.characterId === 'imp')).toBe(false);
    expect(firstNight.some((entry) => entry.actionType === 'kill')).toBe(false);
  });

  it('can find Fortune Teller as the next living character to wake', () => {
    const aliveCharacters = new Map<PlayerId, Character>([
      [
        playerId('alice'),
        {
          id: 'fortuneteller',
          name: 'Fortune Teller',
          team: 'good',
          ability: 'Each night, choose 2 players: you learn if either is a Demon.',
        },
      ],
    ]);

    expect(getNextToWake(INITIAL_NIGHT_STATE, 1, aliveCharacters)).toEqual({
      characterId: 'fortuneteller',
      order: 9,
      actionType: 'check_demon',
    });
  });
});
