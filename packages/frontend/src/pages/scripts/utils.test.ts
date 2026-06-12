import { describe, expect, it } from 'vitest';
import { TROUBLE_BREWING_SCRIPT } from '@clocktower/core';
import {
  charactersByType,
  roleCountText,
  getFirstNightOrder,
  getLaterNightOrder,
  CHARACTER_TYPE_ORDER,
  DEFAULT_SCRIPT_ID,
} from './utils';

describe('charactersByType', () => {
  it('filters characters by type', () => {
    const townsfolk = charactersByType('townsfolk');
    expect(townsfolk.length).toBeGreaterThan(0);
    for (const character of townsfolk) {
      expect(character.type).toBe('townsfolk');
    }
  });

  it('returns all character types from the script', () => {
    for (const type of CHARACTER_TYPE_ORDER) {
      const characters = charactersByType(type);
      expect(characters.length).toBeGreaterThan(0);
    }
  });

  it('covers all characters in the script', () => {
    const allFiltered = CHARACTER_TYPE_ORDER.flatMap((type) => charactersByType(type));
    expect(allFiltered.length).toBe(TROUBLE_BREWING_SCRIPT.characters.length);
  });
});

describe('roleCountText', () => {
  it('returns formatted role count for valid player counts', () => {
    const text = roleCountText(5);
    expect(text).toBe('5人：镇民3 / 外来者0 / 爪牙1 / 恶魔1');
  });

  it('returns correct count for 7 players', () => {
    const text = roleCountText(7);
    expect(text).toBe('7人：镇民5 / 外来者0 / 爪牙1 / 恶魔1');
  });

  it('returns empty string for invalid player count', () => {
    expect(roleCountText(3)).toBe('');
    expect(roleCountText(20)).toBe('');
  });
});

describe('night order', () => {
  it('first night has ordered wake steps', () => {
    const steps = getFirstNightOrder();
    expect(steps.length).toBeGreaterThan(0);
    for (let i = 1; i < steps.length; i++) {
      expect(steps[i]!.order).toBeGreaterThanOrEqual(steps[i - 1]!.order);
    }
  });

  it('later nights have ordered wake steps', () => {
    const steps = getLaterNightOrder();
    expect(steps.length).toBeGreaterThan(0);
    for (let i = 1; i < steps.length; i++) {
      expect(steps[i]!.order).toBeGreaterThanOrEqual(steps[i - 1]!.order);
    }
  });

  it('first night has more steps than later nights', () => {
    expect(getFirstNightOrder().length).toBeGreaterThanOrEqual(getLaterNightOrder().length);
  });

  it('all wake steps reference valid script characters', () => {
    const characterIds = new Set(TROUBLE_BREWING_SCRIPT.characters.map((c) => c.id));
    for (const step of getFirstNightOrder()) {
      expect(characterIds.has(step.characterId)).toBe(true);
    }
  });

  it('DEFAULT_SCRIPT_ID matches the Trouble Brewing script', () => {
    expect(DEFAULT_SCRIPT_ID).toBe(TROUBLE_BREWING_SCRIPT.id);
  });
});
