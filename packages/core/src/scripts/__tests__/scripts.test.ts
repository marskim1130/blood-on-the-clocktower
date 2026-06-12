import { describe, expect, it } from 'vitest';
import {
  buildDefaultScriptAssignments,
  countCharacterTypes,
  getExpectedRoleCount,
  getScriptById,
  getScriptWakeOrder,
  TROUBLE_BREWING_SCRIPT,
  validateScriptAssignment,
} from '../index.js';

function players(count: number): readonly { readonly id: string }[] {
  return Array.from({ length: count }, (_, index) => ({ id: `p${index + 1}` }));
}

describe('script catalog', () => {
  it('contains the complete Trouble Brewing role set', () => {
    const counts = countCharacterTypes(
      TROUBLE_BREWING_SCRIPT.characters.map((character) => character.id),
    );

    expect(getScriptById('trouble_brewing')).toBe(TROUBLE_BREWING_SCRIPT);
    expect(TROUBLE_BREWING_SCRIPT.characters).toHaveLength(22);
    expect(counts).toEqual({
      townsfolk: 13,
      outsiders: 4,
      minions: 4,
      demons: 1,
    });
  });

  it('builds a valid default assignment without duplicate characters', () => {
    const assignments = buildDefaultScriptAssignments(players(15));
    const characterIds = Object.values(assignments);

    expect(characterIds).toHaveLength(15);
    expect(new Set(characterIds).size).toBe(15);
    expect(validateScriptAssignment(assignments, 15)).toEqual({
      ok: true,
      expected: { townsfolk: 9, outsiders: 2, minions: 3, demons: 1 },
      actual: { townsfolk: 9, outsiders: 2, minions: 3, demons: 1 },
    });
  });

  it('applies the Baron setup modifier when validating assignments', () => {
    const validWithBaron = {
      p1: 'washerwoman',
      p2: 'librarian',
      p3: 'investigator',
      p4: 'butler',
      p5: 'saint',
      p6: 'baron',
      p7: 'imp',
    };

    const invalidBaseCountWithBaron = {
      p1: 'washerwoman',
      p2: 'librarian',
      p3: 'investigator',
      p4: 'chef',
      p5: 'empath',
      p6: 'baron',
      p7: 'imp',
    };

    expect(getExpectedRoleCount(7, Object.values(validWithBaron))).toEqual({
      townsfolk: 3,
      outsiders: 2,
      minions: 1,
      demons: 1,
    });
    expect(validateScriptAssignment(validWithBaron, 7).ok).toBe(true);
    expect(validateScriptAssignment(invalidBaseCountWithBaron, 7)).toMatchObject({
      ok: false,
      code: 'ROLE_COUNT_MISMATCH',
    });
  });

  it('rejects duplicate characters even when type counts would otherwise fit', () => {
    const duplicateTownsfolk = {
      p1: 'washerwoman',
      p2: 'washerwoman',
      p3: 'investigator',
      p4: 'poisoner',
      p5: 'imp',
    };

    expect(validateScriptAssignment(duplicateTownsfolk, 5)).toMatchObject({
      ok: false,
      code: 'DUPLICATE_CHARACTER',
    });
  });

  it('uses canonical character ids in the night wake order', () => {
    const firstNightIds = getScriptWakeOrder('trouble_brewing', 1).map((step) => step.characterId);
    const subsequentNightIds = getScriptWakeOrder('trouble_brewing', 2).map((step) => step.characterId);

    expect(firstNightIds).toContain('fortuneteller');
    expect(subsequentNightIds).toContain('fortuneteller');
    expect(firstNightIds).not.toContain('fortune_teller');
    expect(subsequentNightIds).not.toContain('fortune_teller');
  });
});
