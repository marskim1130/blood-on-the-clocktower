import type { NightActionType } from '../night-phase/index.js';
import type { Team } from '../types/index.js';

export type CharacterType = 'townsfolk' | 'outsider' | 'minion' | 'demon';

export interface RoleCount {
  readonly townsfolk: number;
  readonly outsiders: number;
  readonly minions: number;
  readonly demons: number;
}

export interface ScriptCharacterDefinition {
  readonly id: string;
  readonly name: string;
  readonly type: CharacterType;
  readonly team: Team;
  readonly ability: string;
}

export interface NightWakeStep {
  readonly characterId: string;
  readonly order: number;
  readonly actionType: NightActionType;
  readonly prompt: string;
  readonly minTargets: number;
  readonly maxTargets: number;
}

export interface ScriptDefinition {
  readonly id: string;
  readonly name: string;
  readonly characters: readonly ScriptCharacterDefinition[];
}

export type ScriptAssignmentValidation =
  | {
      readonly ok: true;
      readonly expected: RoleCount;
      readonly actual: RoleCount;
    }
  | {
      readonly ok: false;
      readonly code:
        | 'UNKNOWN_SCRIPT'
        | 'UNSUPPORTED_PLAYER_COUNT'
        | 'ASSIGNMENT_COUNT_MISMATCH'
        | 'UNKNOWN_CHARACTER'
        | 'DUPLICATE_CHARACTER'
        | 'ROLE_COUNT_MISMATCH';
      readonly message: string;
      readonly expected?: RoleCount;
      readonly actual?: RoleCount;
    };

const TROUBLE_BREWING_ID = 'trouble_brewing';

export const TROUBLE_BREWING_SCRIPT: ScriptDefinition = {
  id: TROUBLE_BREWING_ID,
  name: 'Trouble Brewing',
  characters: [
    {
      id: 'washerwoman',
      name: 'Washerwoman',
      type: 'townsfolk',
      team: 'good',
      ability: 'You start knowing that one of two players is a particular Townsfolk.',
    },
    {
      id: 'librarian',
      name: 'Librarian',
      type: 'townsfolk',
      team: 'good',
      ability: 'You start knowing that one of two players is a particular Outsider.',
    },
    {
      id: 'investigator',
      name: 'Investigator',
      type: 'townsfolk',
      team: 'good',
      ability: 'You start knowing that one of two players is a particular Minion.',
    },
    {
      id: 'chef',
      name: 'Chef',
      type: 'townsfolk',
      team: 'good',
      ability: 'You start knowing how many pairs of evil players there are.',
    },
    {
      id: 'empath',
      name: 'Empath',
      type: 'townsfolk',
      team: 'good',
      ability: 'Each night, you learn how many of your alive neighbours are evil.',
    },
    {
      id: 'fortuneteller',
      name: 'Fortune Teller',
      type: 'townsfolk',
      team: 'good',
      ability: 'Each night, choose 2 players: you learn if either is a Demon.',
    },
    {
      id: 'undertaker',
      name: 'Undertaker',
      type: 'townsfolk',
      team: 'good',
      ability: 'Each night, you learn which character died by execution today.',
    },
    {
      id: 'monk',
      name: 'Monk',
      type: 'townsfolk',
      team: 'good',
      ability: 'Each night, choose a player: they are safe from the Demon tonight.',
    },
    {
      id: 'ravenkeeper',
      name: 'Ravenkeeper',
      type: 'townsfolk',
      team: 'good',
      ability: 'If you die at night, you are woken to choose a player: you learn their character.',
    },
    {
      id: 'virgin',
      name: 'Virgin',
      type: 'townsfolk',
      team: 'good',
      ability: 'The first time you are nominated, if the nominator is a Townsfolk, they are executed immediately.',
    },
    {
      id: 'slayer',
      name: 'Slayer',
      type: 'townsfolk',
      team: 'good',
      ability: 'Once per game, during the day, choose a player: if they are the Demon, they die.',
    },
    {
      id: 'soldier',
      name: 'Soldier',
      type: 'townsfolk',
      team: 'good',
      ability: 'You are safe from the Demon at night.',
    },
    {
      id: 'mayor',
      name: 'Mayor',
      type: 'townsfolk',
      team: 'good',
      ability: 'If only 3 players live and no execution occurs, your team wins. If you die at night, another player might die instead.',
    },
    {
      id: 'butler',
      name: 'Butler',
      type: 'outsider',
      team: 'good',
      ability: 'Each night, choose a player: you may only vote if they vote.',
    },
    {
      id: 'drunk',
      name: 'Drunk',
      type: 'outsider',
      team: 'good',
      ability: 'You do not know you are the Drunk. You think you are a Townsfolk, but your ability malfunctions.',
    },
    {
      id: 'recluse',
      name: 'Recluse',
      type: 'outsider',
      team: 'good',
      ability: 'You might register as evil, even if you are good.',
    },
    {
      id: 'saint',
      name: 'Saint',
      type: 'outsider',
      team: 'good',
      ability: 'If you are executed, your team loses.',
    },
    {
      id: 'poisoner',
      name: 'Poisoner',
      type: 'minion',
      team: 'evil',
      ability: 'Each night, choose a player: they are poisoned until next dusk.',
    },
    {
      id: 'spy',
      name: 'Spy',
      type: 'minion',
      team: 'evil',
      ability: 'Each night, you see the Grimoire. You might register as good or as a Townsfolk or Outsider.',
    },
    {
      id: 'baron',
      name: 'Baron',
      type: 'minion',
      team: 'evil',
      ability: 'There are extra Outsiders in play.',
    },
    {
      id: 'scarletwoman',
      name: 'Scarlet Woman',
      type: 'minion',
      team: 'evil',
      ability: 'If there are 5 or more players alive and the Demon dies, you become the Demon.',
    },
    {
      id: 'imp',
      name: 'Imp',
      type: 'demon',
      team: 'evil',
      ability: 'Each night, choose a player: they die. If you kill yourself this way, a Minion becomes the Imp.',
    },
  ],
};

export const SCRIPT_CATALOG: readonly ScriptDefinition[] = [TROUBLE_BREWING_SCRIPT];

export const BASE_ROLE_COUNTS: Readonly<Record<number, RoleCount>> = {
  5: { townsfolk: 3, outsiders: 0, minions: 1, demons: 1 },
  6: { townsfolk: 3, outsiders: 1, minions: 1, demons: 1 },
  7: { townsfolk: 5, outsiders: 0, minions: 1, demons: 1 },
  8: { townsfolk: 5, outsiders: 1, minions: 1, demons: 1 },
  9: { townsfolk: 5, outsiders: 2, minions: 1, demons: 1 },
  10: { townsfolk: 7, outsiders: 0, minions: 2, demons: 1 },
  11: { townsfolk: 7, outsiders: 1, minions: 2, demons: 1 },
  12: { townsfolk: 7, outsiders: 2, minions: 2, demons: 1 },
  13: { townsfolk: 9, outsiders: 0, minions: 3, demons: 1 },
  14: { townsfolk: 9, outsiders: 1, minions: 3, demons: 1 },
  15: { townsfolk: 9, outsiders: 2, minions: 3, demons: 1 },
};

export const TROUBLE_BREWING_FIRST_NIGHT_ORDER: readonly NightWakeStep[] = [
  {
    characterId: 'poisoner',
    order: 1,
    actionType: 'poison',
    prompt: 'Poisoner chooses one player to poison until dusk.',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'washerwoman',
    order: 2,
    actionType: 'learn_townsfolk',
    prompt: 'Washerwoman learns that one of two players is a specific Townsfolk.',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'librarian',
    order: 3,
    actionType: 'learn_outsider',
    prompt: 'Librarian learns that one of two players is a specific Outsider, or that none are in play.',
    minTargets: 0,
    maxTargets: 2,
  },
  {
    characterId: 'investigator',
    order: 4,
    actionType: 'learn_minion',
    prompt: 'Investigator learns that one of two players is a specific Minion.',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'chef',
    order: 5,
    actionType: 'learn_evil_pairs',
    prompt: 'Chef learns the number of adjacent evil pairs.',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'empath',
    order: 6,
    actionType: 'learn_evil_neighbors',
    prompt: 'Empath learns how many alive neighbours are evil.',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'fortuneteller',
    order: 7,
    actionType: 'check_demon',
    prompt: 'Fortune Teller chooses two players and learns if either registers as the Demon.',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'butler',
    order: 8,
    actionType: 'learn_master',
    prompt: 'Butler chooses their master for tomorrow.',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'imp',
    order: 9,
    actionType: 'kill',
    prompt: 'Imp chooses one player to die.',
    minTargets: 1,
    maxTargets: 1,
  },
];

export const TROUBLE_BREWING_SUBSEQUENT_NIGHT_ORDER: readonly NightWakeStep[] = [
  {
    characterId: 'poisoner',
    order: 1,
    actionType: 'poison',
    prompt: 'Poisoner chooses one player to poison until dusk.',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'monk',
    order: 2,
    actionType: 'protect',
    prompt: 'Monk chooses one player other than themself to protect from the Demon.',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'imp',
    order: 3,
    actionType: 'kill',
    prompt: 'Imp chooses one player to die.',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'empath',
    order: 4,
    actionType: 'learn_evil_neighbors',
    prompt: 'Empath learns how many alive neighbours are evil.',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'fortuneteller',
    order: 5,
    actionType: 'check_demon',
    prompt: 'Fortune Teller chooses two players and learns if either registers as the Demon.',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'undertaker',
    order: 6,
    actionType: 'learn_executed',
    prompt: 'Undertaker learns which character died by execution today.',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'butler',
    order: 7,
    actionType: 'learn_master',
    prompt: 'Butler chooses their master for tomorrow.',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'ravenkeeper',
    order: 8,
    actionType: 'learn_died',
    prompt: 'If Ravenkeeper died tonight, they choose one player and learn their character.',
    minTargets: 1,
    maxTargets: 1,
  },
];

const DEFAULT_CHARACTER_PRIORITY: Record<CharacterType, readonly string[]> = {
  townsfolk: [
    'washerwoman',
    'librarian',
    'investigator',
    'chef',
    'empath',
    'fortuneteller',
    'undertaker',
    'monk',
    'ravenkeeper',
  ],
  outsider: ['butler', 'drunk'],
  minion: ['poisoner', 'spy', 'scarletwoman', 'baron'],
  demon: ['imp'],
};

export function getScriptById(scriptId: string): ScriptDefinition | null {
  return SCRIPT_CATALOG.find((script) => script.id === scriptId) ?? null;
}

export function getScriptCharacterById(
  characterId: string,
  scriptId = TROUBLE_BREWING_ID,
): ScriptCharacterDefinition | null {
  const script = getScriptById(scriptId);
  return script?.characters.find((character) => character.id === characterId) ?? null;
}

export function getScriptCharactersByType(
  type: CharacterType,
  scriptId = TROUBLE_BREWING_ID,
): readonly ScriptCharacterDefinition[] {
  const script = getScriptById(scriptId);
  return script?.characters.filter((character) => character.type === type) ?? [];
}

export function getBaseRoleCount(playerCount: number): RoleCount | null {
  return BASE_ROLE_COUNTS[playerCount] ?? null;
}

export function getExpectedRoleCount(
  playerCount: number,
  characterIds: readonly string[] = [],
): RoleCount | null {
  const base = getBaseRoleCount(playerCount);
  if (!base) return null;

  if (!characterIds.includes('baron')) {
    return base;
  }

  return {
    townsfolk: Math.max(0, base.townsfolk - 2),
    outsiders: base.outsiders + 2,
    minions: base.minions,
    demons: base.demons,
  };
}

export function countCharacterTypes(
  characterIds: readonly string[],
  scriptId = TROUBLE_BREWING_ID,
): RoleCount | null {
  const script = getScriptById(scriptId);
  if (!script) return null;

  const counts: Record<CharacterType, number> = {
    townsfolk: 0,
    outsider: 0,
    minion: 0,
    demon: 0,
  };

  for (const characterId of characterIds) {
    const definition = script.characters.find((character) => character.id === characterId);
    if (!definition) return null;
    counts[definition.type]++;
  }

  return {
    townsfolk: counts.townsfolk,
    outsiders: counts.outsider,
    minions: counts.minion,
    demons: counts.demon,
  };
}

export function validateScriptAssignment(
  assignments: Readonly<Record<string, string>>,
  playerCount: number,
  scriptId = TROUBLE_BREWING_ID,
): ScriptAssignmentValidation {
  const script = getScriptById(scriptId);
  if (!script) {
    return { ok: false, code: 'UNKNOWN_SCRIPT', message: `Unknown script: ${scriptId}` };
  }

  const characterIds = Object.values(assignments);
  const expected = getExpectedRoleCount(playerCount, characterIds);
  if (!expected) {
    return {
      ok: false,
      code: 'UNSUPPORTED_PLAYER_COUNT',
      message: `Unsupported player count: ${playerCount}`,
    };
  }

  if (characterIds.length !== playerCount) {
    return {
      ok: false,
      code: 'ASSIGNMENT_COUNT_MISMATCH',
      message: `Expected ${playerCount} assignments, got ${characterIds.length}`,
      expected,
    };
  }

  const seen = new Set<string>();
  for (const characterId of characterIds) {
    if (!script.characters.some((character) => character.id === characterId)) {
      return {
        ok: false,
        code: 'UNKNOWN_CHARACTER',
        message: `Unknown character: ${characterId}`,
        expected,
      };
    }
    if (seen.has(characterId)) {
      return {
        ok: false,
        code: 'DUPLICATE_CHARACTER',
        message: `Duplicate character: ${characterId}`,
        expected,
      };
    }
    seen.add(characterId);
  }

  const actual = countCharacterTypes(characterIds, scriptId);
  if (!actual) {
    return {
      ok: false,
      code: 'UNKNOWN_CHARACTER',
      message: 'Assignment contains an unknown character',
      expected,
    };
  }

  if (!roleCountsEqual(actual, expected)) {
    return {
      ok: false,
      code: 'ROLE_COUNT_MISMATCH',
      message: 'Character type counts do not match the script setup rules',
      expected,
      actual,
    };
  }

  return { ok: true, expected, actual };
}

export function buildDefaultScriptAssignments(
  players: readonly { readonly id: string }[],
  scriptId = TROUBLE_BREWING_ID,
): Record<string, string> {
  if (scriptId !== TROUBLE_BREWING_ID) {
    throw new Error(`Unsupported script: ${scriptId}`);
  }

  const roleCount = getBaseRoleCount(players.length);
  if (!roleCount) {
    throw new Error(`Unsupported player count: ${players.length}`);
  }

  const characterIds = [
    ...DEFAULT_CHARACTER_PRIORITY.townsfolk.slice(0, roleCount.townsfolk),
    ...DEFAULT_CHARACTER_PRIORITY.outsider.slice(0, roleCount.outsiders),
    ...DEFAULT_CHARACTER_PRIORITY.minion.slice(0, roleCount.minions),
    ...DEFAULT_CHARACTER_PRIORITY.demon.slice(0, roleCount.demons),
  ];

  if (characterIds.length !== players.length) {
    throw new Error(`Expected ${players.length} characters, got ${characterIds.length}`);
  }

  return players.reduce<Record<string, string>>((assignments, player, index) => {
    const characterId = characterIds[index];
    if (!characterId) return assignments;
    assignments[player.id] = characterId;
    return assignments;
  }, {});
}

export function getScriptWakeOrder(
  scriptId: string,
  nightNumber: number,
): readonly NightWakeStep[] {
  if (scriptId !== TROUBLE_BREWING_ID) return [];
  return nightNumber <= 1 ? TROUBLE_BREWING_FIRST_NIGHT_ORDER : TROUBLE_BREWING_SUBSEQUENT_NIGHT_ORDER;
}

function roleCountsEqual(left: RoleCount, right: RoleCount): boolean {
  return (
    left.townsfolk === right.townsfolk &&
    left.outsiders === right.outsiders &&
    left.minions === right.minions &&
    left.demons === right.demons
  );
}
