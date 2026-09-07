import { describe, expect, it } from 'vitest';
import {
  buildDefaultScriptAssignments,
  countCharacterTypes,
  getExpectedRoleCount,
  getScriptById,
  getScriptWakeOrder,
  randomizeScriptAssignments,
  swapScriptAssignments,
  TROUBLE_BREWING_SCRIPT,
  validateScriptAssignment,
  validateScriptSetup,
} from '../index.js';

function players(count: number): readonly { readonly id: string }[] {
  return Array.from({ length: count }, (_, index) => ({ id: `p${index + 1}` }));
}

function seededRandom(seed: number): () => number {
  return () => {
    seed = (Math.imul(seed, 1_664_525) + 1_013_904_223) >>> 0;
    return seed / 0x1_0000_0000;
  };
}

describe('script catalog', () => {
  it('wakes the Spy to view the grimoire after other actions every night', () => {
    for (const night of [1, 2]) {
      const steps = getScriptWakeOrder('trouble_brewing', night);
      expect(steps.at(-1)).toMatchObject({ characterId: 'spy', actionType: 'show_grimoire', minTargets: 0, maxTargets: 0 });
    }
  });
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

  it.each([5, 15])('randomizes a valid setup for %i players', (playerCount) => {
    const setup = randomizeScriptAssignments(
      players(playerCount),
      'trouble_brewing',
      seededRandom(playerCount),
    );

    expect(Object.keys(setup.assignments)).toHaveLength(playerCount);
    expect(new Set(Object.values(setup.assignments))).toHaveLength(playerCount);
    expect(validateScriptSetup(setup, players(playerCount))).toMatchObject({ ok: true });
  });

  it('generates three unique out-of-play good characters as Demon bluffs', () => {
    const setup = randomizeScriptAssignments(players(5), 'trouble_brewing', seededRandom(23));
    const assigned = new Set(Object.values(setup.assignments));

    expect(setup.demonBluffCharacterIds).toHaveLength(3);
    expect(new Set(setup.demonBluffCharacterIds)).toHaveLength(3);
    for (const characterId of setup.demonBluffCharacterIds) {
      const character = TROUBLE_BREWING_SCRIPT.characters.find((candidate) => candidate.id === characterId);
      expect(assigned.has(characterId)).toBe(false);
      expect(character?.team).toBe('good');
    }
  });

  it('selects outsiders using the Baron-adjusted role count', () => {
    const values = [0.99, 0, 0.99];
    const setup = randomizeScriptAssignments(players(5), 'trouble_brewing', () => {
      return values.shift() ?? 0.99;
    });

    expect(Object.values(setup.assignments)).toContain('baron');
    expect(countCharacterTypes(Object.values(setup.assignments))).toEqual({
      townsfolk: 1,
      outsiders: 2,
      minions: 1,
      demons: 1,
    });
    expect(Object.values(setup.shownCharacters)).toHaveLength(1);
    expect(validateScriptSetup(setup, players(5))).toMatchObject({ ok: true });
  });

  it('validates the Drunk shown character and Fortune Teller red herring', () => {
    const drunkSetup = {
      assignments: {
        p1: 'washerwoman',
        p2: 'butler',
        p3: 'drunk',
        p4: 'baron',
        p5: 'imp',
      },
      shownCharacters: { p3: 'chef' },
      fortuneTellerRedHerringId: null,
      demonBluffCharacterIds: ['librarian', 'investigator', 'empath'],
    };
    const fortuneTellerSetup = {
      assignments: {
        p1: 'fortuneteller',
        p2: 'chef',
        p3: 'empath',
        p4: 'poisoner',
        p5: 'imp',
      },
      shownCharacters: {},
      fortuneTellerRedHerringId: 'p2',
      demonBluffCharacterIds: ['washerwoman', 'librarian', 'investigator'],
    };

    expect(validateScriptSetup(drunkSetup, players(5))).toMatchObject({ ok: true });
    expect(validateScriptSetup(fortuneTellerSetup, players(5))).toMatchObject({ ok: true });
    expect(
      validateScriptSetup(
        { ...drunkSetup, shownCharacters: { p3: 'washerwoman' } },
        players(5),
      ),
    ).toMatchObject({ ok: false, code: 'INVALID_DRUNK_SHOWN_CHARACTER' });
    expect(
      validateScriptSetup(
        { ...fortuneTellerSetup, fortuneTellerRedHerringId: 'p4' },
        players(5),
      ),
    ).toMatchObject({ ok: false, code: 'INVALID_FORTUNE_TELLER_RED_HERRING' });
  });

  it('swaps two assignments without mutating or losing characters', () => {
    const original = buildDefaultScriptAssignments(players(5));
    const swapped = swapScriptAssignments(original, 'p1', 'p5');

    expect(swapped).not.toBe(original);
    expect(swapped.p1).toBe(original.p5);
    expect(swapped.p5).toBe(original.p1);
    expect(Object.values(swapped).sort()).toEqual(Object.values(original).sort());
    expect(new Set(Object.values(swapped))).toHaveLength(5);
  });

  it('rejects invalid setup and randomization inputs', () => {
    const setup = randomizeScriptAssignments(players(5), 'trouble_brewing', seededRandom(5));

    expect(
      validateScriptSetup(setup, [...players(4), { id: 'different-player' }]),
    ).toMatchObject({ ok: false, code: 'PLAYER_ASSIGNMENT_MISMATCH' });
    expect(() => randomizeScriptAssignments(players(4))).toThrow('Unsupported player count');
    expect(() => randomizeScriptAssignments([{ id: 'same' }, { id: 'same' }, ...players(3)])).toThrow(
      'Player ids must be unique',
    );
    expect(() => randomizeScriptAssignments(players(5), 'trouble_brewing', () => 1)).toThrow(
      'Random source must return a finite number in [0, 1)',
    );
    expect(() => swapScriptAssignments(setup.assignments, 'p1', 'missing')).toThrow(
      'Unknown assigned player: missing',
    );
  });
});
