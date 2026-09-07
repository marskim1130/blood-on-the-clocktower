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
  readonly characterType?: CharacterType;
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

export interface ScriptSetup {
  readonly assignments: Readonly<Record<string, string>>;
  readonly shownCharacters: Readonly<Record<string, string>>;
  readonly fortuneTellerRedHerringId: string | null;
  readonly demonBluffCharacterIds: readonly string[];
}

export type ScriptSetupValidation =
  | Extract<ScriptAssignmentValidation, { readonly ok: true }>
  | Extract<ScriptAssignmentValidation, { readonly ok: false }>
  | {
      readonly ok: false;
      readonly code:
        | 'PLAYER_ASSIGNMENT_MISMATCH'
        | 'INVALID_DRUNK_SHOWN_CHARACTER'
        | 'INVALID_FORTUNE_TELLER_RED_HERRING'
        | 'INVALID_DEMON_BLUFFS';
      readonly message: string;
    };

export type RandomSource = () => number;

const TROUBLE_BREWING_ID = 'trouble_brewing';

export const TROUBLE_BREWING_SCRIPT: ScriptDefinition = {
  id: TROUBLE_BREWING_ID,
  name: '暗流涌动',
  characters: [
    {
      id: 'washerwoman',
      name: '洗衣妇',
      type: 'townsfolk',
      team: 'good',
      ability: '游戏开始时，你得知两名玩家之一是某个镇民。',
    },
    {
      id: 'librarian',
      name: '图书管理员',
      type: 'townsfolk',
      team: 'good',
      ability: '游戏开始时，你得知两名玩家之一是某个外来者；若没有外来者在场，你会得知这一点。',
    },
    {
      id: 'investigator',
      name: '调查员',
      type: 'townsfolk',
      team: 'good',
      ability: '游戏开始时，你得知两名玩家之一是某个爪牙。',
    },
    {
      id: 'chef',
      name: '厨师',
      type: 'townsfolk',
      team: 'good',
      ability: '游戏开始时，你得知有多少对相邻的邪恶玩家。',
    },
    {
      id: 'empath',
      name: '共情者',
      type: 'townsfolk',
      team: 'good',
      ability: '每个夜晚，你得知与你相邻的存活玩家中有多少名是邪恶阵营。',
    },
    {
      id: 'fortuneteller',
      name: '占卜师',
      type: 'townsfolk',
      team: 'good',
      ability: '每个夜晚，选择 2 名玩家：你得知其中是否有恶魔。会有一名善良玩家被登记为恶魔。',
    },
    {
      id: 'undertaker',
      name: '入殓师',
      type: 'townsfolk',
      team: 'good',
      ability: '每个夜晚，你得知今天被处决的玩家是什么角色。',
    },
    {
      id: 'monk',
      name: '僧侣',
      type: 'townsfolk',
      team: 'good',
      ability: '每个夜晚，选择除自己以外的一名玩家：今晚他不会被恶魔杀死。',
    },
    {
      id: 'ravenkeeper',
      name: '守鸦人',
      type: 'townsfolk',
      team: 'good',
      ability: '如果你在夜晚死亡，你会被唤醒并选择一名玩家：你得知他的角色。',
    },
    {
      id: 'virgin',
      name: '圣女',
      type: 'townsfolk',
      team: 'good',
      ability: '第一次有人提名你时，如果提名者是镇民，他立刻被处决。',
    },
    {
      id: 'slayer',
      name: '杀手',
      type: 'townsfolk',
      team: 'good',
      ability: '每局一次，在白天选择一名玩家：如果他是恶魔，他死亡。',
    },
    {
      id: 'soldier',
      name: '士兵',
      type: 'townsfolk',
      team: 'good',
      ability: '你在夜晚不会被恶魔杀死。',
    },
    {
      id: 'mayor',
      name: '镇长',
      type: 'townsfolk',
      team: 'good',
      ability: '如果只剩 3 名玩家存活且当天无人被处决，你的阵营获胜。如果你在夜晚死亡，另一名玩家可能会代替你死亡。',
    },
    {
      id: 'butler',
      name: '管家',
      type: 'outsider',
      team: 'good',
      ability: '每个夜晚，选择一名玩家作为主人：明天只有当主人投票时，你才能投票。',
    },
    {
      id: 'drunk',
      name: '酒鬼',
      type: 'outsider',
      team: 'good',
      ability: '你不知道自己是酒鬼。你以为自己是镇民，但你的能力失效。',
    },
    {
      id: 'recluse',
      name: '隐士',
      type: 'outsider',
      team: 'good',
      ability: '即使你是善良阵营，也可能被登记为邪恶。',
    },
    {
      id: 'saint',
      name: '圣徒',
      type: 'outsider',
      team: 'good',
      ability: '如果你被处决，你的阵营落败。',
    },
    {
      id: 'poisoner',
      name: '投毒者',
      type: 'minion',
      team: 'evil',
      ability: '每个夜晚，选择一名玩家：他中毒直到下一个黄昏。',
    },
    {
      id: 'spy',
      name: '间谍',
      type: 'minion',
      team: 'evil',
      ability: '每个夜晚，你可以查看魔典。你可能被登记为善良、镇民或外来者。',
    },
    {
      id: 'baron',
      name: '男爵',
      type: 'minion',
      team: 'evil',
      ability: '场上有额外的外来者。',
    },
    {
      id: 'scarletwoman',
      name: '红唇女郎',
      type: 'minion',
      team: 'evil',
      ability: '如果有 5 名或更多玩家存活且恶魔死亡，你变成恶魔。',
    },
    {
      id: 'imp',
      name: '小恶魔',
      type: 'demon',
      team: 'evil',
      ability: '除首夜外的每个夜晚，选择一名玩家：他死亡。如果你用这种方式杀死自己，一名爪牙变成小恶魔。',
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
    characterId: '',
    characterType: 'minion',
    order: 1,
    actionType: 'learn_demon',
    prompt: '爪牙得知哪名玩家是恶魔。',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: '',
    characterType: 'demon',
    order: 2,
    actionType: 'learn_minion',
    prompt: '恶魔得知哪些玩家是爪牙。',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'poisoner',
    order: 3,
    actionType: 'poison',
    prompt: '投毒者选择一名玩家，使其中毒直到黄昏。',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'washerwoman',
    order: 4,
    actionType: 'learn_townsfolk',
    prompt: '洗衣妇得知两名玩家之一是某个镇民。',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'librarian',
    order: 5,
    actionType: 'learn_outsider',
    prompt: '图书管理员得知两名玩家之一是某个外来者，或得知没有外来者在场。',
    minTargets: 0,
    maxTargets: 2,
  },
  {
    characterId: 'investigator',
    order: 6,
    actionType: 'learn_minion',
    prompt: '调查员得知两名玩家之一是某个爪牙。',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'chef',
    order: 7,
    actionType: 'learn_evil_pairs',
    prompt: '厨师得知相邻邪恶玩家的对数。',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'empath',
    order: 8,
    actionType: 'learn_evil_neighbors',
    prompt: '共情者得知相邻存活玩家中有多少名是邪恶阵营。',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'fortuneteller',
    order: 9,
    actionType: 'check_demon',
    prompt: '占卜师选择两名玩家，并得知其中是否有人被登记为恶魔。',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'butler',
    order: 10,
    actionType: 'learn_master',
    prompt: '管家选择明天的主人。',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'spy', order: 11, actionType: 'show_grimoire',
    prompt: '间谍查看说书人确认的魔典，包括全部角色与状态。', minTargets: 0, maxTargets: 0,
  },
];

export const TROUBLE_BREWING_SUBSEQUENT_NIGHT_ORDER: readonly NightWakeStep[] = [
  {
    characterId: 'poisoner',
    order: 1,
    actionType: 'poison',
    prompt: '投毒者选择一名玩家，使其中毒直到黄昏。',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'monk',
    order: 2,
    actionType: 'protect',
    prompt: '僧侣选择除自己以外的一名玩家，使其今晚不被恶魔杀死。',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'imp',
    order: 3,
    actionType: 'kill',
    prompt: '小恶魔选择一名玩家死亡。',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'empath',
    order: 4,
    actionType: 'learn_evil_neighbors',
    prompt: '共情者得知相邻存活玩家中有多少名是邪恶阵营。',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'fortuneteller',
    order: 5,
    actionType: 'check_demon',
    prompt: '占卜师选择两名玩家，并得知其中是否有人被登记为恶魔。',
    minTargets: 2,
    maxTargets: 2,
  },
  {
    characterId: 'undertaker',
    order: 6,
    actionType: 'learn_executed',
    prompt: '入殓师得知今天被处决的玩家是什么角色。',
    minTargets: 0,
    maxTargets: 0,
  },
  {
    characterId: 'butler',
    order: 7,
    actionType: 'learn_master',
    prompt: '管家选择明天的主人。',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'ravenkeeper',
    order: 8,
    actionType: 'learn_died',
    prompt: '如果守鸦人今晚死亡，他选择一名玩家并得知其角色。',
    minTargets: 1,
    maxTargets: 1,
  },
  {
    characterId: 'spy', order: 9, actionType: 'show_grimoire',
    prompt: '间谍查看说书人确认的魔典，包括全部角色与状态。', minTargets: 0, maxTargets: 0,
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

export function validateScriptSetup(
  setup: ScriptSetup,
  players: readonly { readonly id: string }[],
  scriptId = TROUBLE_BREWING_ID,
): ScriptSetupValidation {
  const assignmentValidation = validateScriptAssignment(
    setup.assignments,
    players.length,
    scriptId,
  );
  if (!assignmentValidation.ok) return assignmentValidation;

  const playerIds = players.map((player) => player.id);
  const assignedPlayerIds = Object.keys(setup.assignments);
  const playerIdSet = new Set(playerIds);
  if (
    playerIdSet.size !== playerIds.length ||
    assignedPlayerIds.some((playerId) => !playerIdSet.has(playerId)) ||
    playerIds.some((playerId) => !hasOwn(setup.assignments, playerId))
  ) {
    return {
      ok: false,
      code: 'PLAYER_ASSIGNMENT_MISMATCH',
      message: 'Assignments must contain each player exactly once',
    };
  }

  const characterIds = Object.values(setup.assignments);
  const drunkPlayerId = assignedPlayerIds.find(
    (playerId) => setup.assignments[playerId] === 'drunk',
  );
  const shownEntries = Object.entries(setup.shownCharacters);
  if (!drunkPlayerId) {
    if (shownEntries.length > 0) {
      return {
        ok: false,
        code: 'INVALID_DRUNK_SHOWN_CHARACTER',
        message: 'Shown characters are only valid when the Drunk is in play',
      };
    }
  } else {
    const shownCharacterId = setup.shownCharacters[drunkPlayerId];
    const shownDefinition = shownCharacterId
      ? getScriptCharacterById(shownCharacterId, scriptId)
      : null;
    if (
      shownEntries.length !== 1 ||
      !shownCharacterId ||
      !shownDefinition ||
      shownDefinition.type !== 'townsfolk' ||
      characterIds.includes(shownCharacterId)
    ) {
      return {
        ok: false,
        code: 'INVALID_DRUNK_SHOWN_CHARACTER',
        message: 'The Drunk must be shown an out-of-play Townsfolk character',
      };
    }
  }

  const fortuneTellerPlayerId = assignedPlayerIds.find(
    (playerId) => setup.assignments[playerId] === 'fortuneteller',
  );
  if (!fortuneTellerPlayerId) {
    if (setup.fortuneTellerRedHerringId !== null) {
      return {
        ok: false,
        code: 'INVALID_FORTUNE_TELLER_RED_HERRING',
        message: 'A red herring is only valid when the Fortune Teller is in play',
      };
    }
  } else {
    const redHerringId = setup.fortuneTellerRedHerringId;
    const redHerringCharacterId = redHerringId ? setup.assignments[redHerringId] : null;
    const redHerringDefinition = redHerringCharacterId
      ? getScriptCharacterById(redHerringCharacterId, scriptId)
      : null;
    if (
      !redHerringId ||
      redHerringId === fortuneTellerPlayerId ||
      !redHerringDefinition ||
      redHerringDefinition.team !== 'good'
    ) {
      return {
        ok: false,
        code: 'INVALID_FORTUNE_TELLER_RED_HERRING',
        message: 'The Fortune Teller red herring must be another good player',
      };
    }
  }

  const bluffIds = setup.demonBluffCharacterIds;
  const unavailableBluffs = new Set([...characterIds, ...Object.values(setup.shownCharacters)]);
  if (
    bluffIds.length !== 3 ||
    new Set(bluffIds).size !== 3 ||
    bluffIds.some((characterId) => {
      const character = getScriptCharacterById(characterId, scriptId);
      return !character || character.team !== 'good' || unavailableBluffs.has(characterId);
    })
  ) {
    return {
      ok: false,
      code: 'INVALID_DEMON_BLUFFS',
      message: 'Demon bluffs must be three unique out-of-play good characters',
    };
  }

  return assignmentValidation;
}

export function randomizeScriptAssignments(
  players: readonly { readonly id: string }[],
  scriptId = TROUBLE_BREWING_ID,
  random: RandomSource = Math.random,
): ScriptSetup {
  const script = getScriptById(scriptId);
  if (!script) throw new Error(`Unknown script: ${scriptId}`);

  const baseRoleCount = getBaseRoleCount(players.length);
  if (!baseRoleCount) {
    throw new Error(`Unsupported player count: ${players.length}`);
  }

  const playerIds = players.map((player) => player.id);
  if (new Set(playerIds).size !== playerIds.length) {
    throw new Error('Player ids must be unique');
  }

  const demons = selectRandomCharacterIds(
    getScriptCharactersByType('demon', scriptId),
    baseRoleCount.demons,
    random,
  );
  const minions = selectRandomCharacterIds(
    getScriptCharactersByType('minion', scriptId),
    baseRoleCount.minions,
    random,
  );
  const expected = getExpectedRoleCount(players.length, [...demons, ...minions]);
  if (!expected) {
    throw new Error(`Unsupported player count: ${players.length}`);
  }

  const townsfolk = selectRandomCharacterIds(
    getScriptCharactersByType('townsfolk', scriptId),
    expected.townsfolk,
    random,
  );
  const outsiders = selectRandomCharacterIds(
    getScriptCharactersByType('outsider', scriptId),
    expected.outsiders,
    random,
  );
  const characterIds = shuffle([...townsfolk, ...outsiders, ...minions, ...demons], random);

  const assignments = playerIds.reduce<Record<string, string>>((result, playerId, index) => {
    const characterId = characterIds[index];
    if (!characterId) throw new Error(`Missing character for player: ${playerId}`);
    result[playerId] = characterId;
    return result;
  }, {});

  const shownCharacters: Record<string, string> = {};
  const drunkPlayerId = playerIds.find((playerId) => assignments[playerId] === 'drunk');
  if (drunkPlayerId) {
    const outOfPlayTownsfolk = getScriptCharactersByType('townsfolk', scriptId).filter(
      (character) => !characterIds.includes(character.id),
    );
    const shownCharacter = selectRandomCharacterIds(outOfPlayTownsfolk, 1, random)[0];
    if (!shownCharacter) throw new Error('No out-of-play Townsfolk is available for the Drunk');
    shownCharacters[drunkPlayerId] = shownCharacter;
  }

  const fortuneTellerPlayerId = playerIds.find(
    (playerId) => assignments[playerId] === 'fortuneteller',
  );
  let fortuneTellerRedHerringId: string | null = null;
  if (fortuneTellerPlayerId) {
    const candidates = playerIds.filter((playerId) => {
      if (playerId === fortuneTellerPlayerId) return false;
      const characterId = assignments[playerId];
      return characterId
        ? getScriptCharacterById(characterId, scriptId)?.team === 'good'
        : false;
    });
    fortuneTellerRedHerringId = shuffle(candidates, random)[0] ?? null;
    if (!fortuneTellerRedHerringId) {
      throw new Error('No good player is available as the Fortune Teller red herring');
    }
  }

  const bluffCandidates = script.characters.filter(
    (character) => character.team === 'good' &&
      !characterIds.includes(character.id) &&
      !Object.values(shownCharacters).includes(character.id),
  );
  const demonBluffCharacterIds = selectRandomCharacterIds(bluffCandidates, 3, random);

  const setup: ScriptSetup = {
    assignments,
    shownCharacters,
    fortuneTellerRedHerringId,
    demonBluffCharacterIds,
  };
  const validation = validateScriptSetup(setup, players, scriptId);
  if (!validation.ok) {
    throw new Error(`Generated invalid script setup: ${validation.message}`);
  }
  return setup;
}

export function swapScriptAssignments(
  assignments: Readonly<Record<string, string>>,
  firstPlayerId: string,
  secondPlayerId: string,
): Record<string, string> {
  if (!hasOwn(assignments, firstPlayerId)) {
    throw new Error(`Unknown assigned player: ${firstPlayerId}`);
  }
  if (!hasOwn(assignments, secondPlayerId)) {
    throw new Error(`Unknown assigned player: ${secondPlayerId}`);
  }

  return {
    ...assignments,
    [firstPlayerId]: assignments[secondPlayerId]!,
    [secondPlayerId]: assignments[firstPlayerId]!,
  };
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

function selectRandomCharacterIds(
  characters: readonly ScriptCharacterDefinition[],
  count: number,
  random: RandomSource,
): string[] {
  if (characters.length < count) {
    throw new Error(`Expected at least ${count} characters, got ${characters.length}`);
  }
  return shuffle(
    characters.map((character) => character.id),
    random,
  ).slice(0, count);
}

function shuffle<T>(values: readonly T[], random: RandomSource): T[] {
  const result = [...values];
  for (let index = result.length - 1; index > 0; index--) {
    const value = random();
    if (!Number.isFinite(value) || value < 0 || value >= 1) {
      throw new Error('Random source must return a finite number in [0, 1)');
    }
    const swapIndex = Math.floor(value * (index + 1));
    [result[index], result[swapIndex]] = [result[swapIndex]!, result[index]!];
  }
  return result;
}

function hasOwn(record: Readonly<Record<string, string>>, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(record, key);
}
