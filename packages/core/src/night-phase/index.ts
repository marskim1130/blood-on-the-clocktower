import type { PlayerId, Character } from '../types/index.js';

// ─── Night Phase Types ──────────────────────────────────────────

/** The type of night action a character performs. */
export type NightActionType =
  | 'learn_townsfolk' // Washerwoman
  | 'learn_outsider' // Librarian
  | 'learn_minion' // Investigator
  | 'learn_evil_pairs' // Chef
  | 'learn_evil_neighbors' // Empath
  | 'check_demon' // Fortune Teller
  | 'learn_executed' // Undertaker
  | 'protect' // Monk
  | 'learn_died' // Ravenkeeper
  | 'learn_master' // Butler
  | 'poison' // Poisoner
  | 'kill'; // Imp

/** A night action recorded by the storyteller. */
export interface NightAction {
  readonly playerId: PlayerId;
  readonly actionType: NightActionType;
  /** The target(s) of the action (if applicable). */
  readonly targets: readonly PlayerId[];
  /** The result of the action (learned information, etc.). */
  readonly result: string | null;
}

/** State of the night phase. */
export interface NightState {
  readonly nightNumber: number;
  readonly actions: readonly NightAction[];
  readonly currentWakeIndex: number;
  readonly isComplete: boolean;
}

/** Wake order entry: character acts at a specific order number. */
export interface WakeOrderEntry {
  readonly characterId: string;
  readonly order: number;
  readonly actionType: NightActionType;
}

// ─── Wake Order Definitions ─────────────────────────────────────

/**
 * First night wake order for Trouble Brewing.
 */
export const FIRST_NIGHT_ORDER: readonly WakeOrderEntry[] = [
  { characterId: 'poisoner', order: 1, actionType: 'poison' },
  { characterId: 'washerwoman', order: 2, actionType: 'learn_townsfolk' },
  { characterId: 'librarian', order: 3, actionType: 'learn_outsider' },
  { characterId: 'investigator', order: 4, actionType: 'learn_minion' },
  { characterId: 'chef', order: 5, actionType: 'learn_evil_pairs' },
  { characterId: 'empath', order: 6, actionType: 'learn_evil_neighbors' },
  { characterId: 'fortune_teller', order: 7, actionType: 'check_demon' },
  { characterId: 'butler', order: 8, actionType: 'learn_master' },
  { characterId: 'imp', order: 9, actionType: 'kill' },
];

/**
 * Subsequent night wake order for Trouble Brewing.
 */
export const SUBSEQUENT_NIGHT_ORDER: readonly WakeOrderEntry[] = [
  { characterId: 'poisoner', order: 1, actionType: 'poison' },
  { characterId: 'monk', order: 2, actionType: 'protect' },
  { characterId: 'imp', order: 3, actionType: 'kill' },
  { characterId: 'empath', order: 4, actionType: 'learn_evil_neighbors' },
  { characterId: 'fortune_teller', order: 5, actionType: 'check_demon' },
  { characterId: 'undertaker', order: 6, actionType: 'learn_executed' },
  { characterId: 'butler', order: 7, actionType: 'learn_master' },
  { characterId: 'ravenkeeper', order: 8, actionType: 'learn_died' },
];

// ─── Initial State ───────────────────────────────────────────────

export const INITIAL_NIGHT_STATE: NightState = {
  nightNumber: 0,
  actions: [],
  currentWakeIndex: 0,
  isComplete: false,
};

// ─── Night Phase Functions ───────────────────────────────────────

/**
 * Get the wake order for a specific night number.
 */
export function getWakeOrder(nightNumber: number): readonly WakeOrderEntry[] {
  if (nightNumber <= 1) {
    return FIRST_NIGHT_ORDER;
  }
  return SUBSEQUENT_NIGHT_ORDER;
}

/**
 * Get the next character to wake during the night.
 */
export function getNextToWake(
  state: NightState,
  nightNumber: number,
  aliveCharacters: ReadonlyMap<PlayerId, Character>,
): WakeOrderEntry | null {
  const wakeOrder = getWakeOrder(nightNumber);

  for (let i = state.currentWakeIndex; i < wakeOrder.length; i++) {
    const entry = wakeOrder[i]!;

    for (const [, character] of aliveCharacters) {
      if (character.id === entry.characterId) {
        return entry;
      }
    }
  }

  return null;
}

/**
 * Record a night action.
 */
export function recordNightAction(
  state: NightState,
  action: NightAction,
): NightState {
  const wakeOrder = getWakeOrder(state.nightNumber);
  const expectedEntry = wakeOrder[state.currentWakeIndex];

  if (!expectedEntry || expectedEntry.actionType !== action.actionType) {
    return state; // Invalid action: no-op
  }

  const newActions = [...state.actions, action];
  const newWakeIndex = state.currentWakeIndex + 1;
  const isComplete = newWakeIndex >= wakeOrder.length;

  return {
    ...state,
    actions: newActions,
    currentWakeIndex: newWakeIndex,
    isComplete,
  };
}

/**
 * Resolve all night actions and determine results.
 */
export function resolveNightActions(
  state: NightState,
  players: ReadonlyMap<PlayerId, Character>,
  deaths: ReadonlySet<PlayerId>,
): DawnResults {
  const results: NightActionResult[] = [];

  for (const action of state.actions) {
    const player = players.get(action.playerId);
    if (!player) continue;

    const result: NightActionResult = {
      playerId: action.playerId,
      characterId: player.id,
      actionType: action.actionType,
      targets: action.targets,
      result: action.result,
    };

    results.push(result);
  }

  return {
    nightNumber: state.nightNumber,
    actions: results,
    deaths: Array.from(deaths),
    isComplete: state.isComplete,
  };
}

/**
 * Start a new night phase.
 */
export function startNight(nightNumber: number): NightState {
  return {
    nightNumber,
    actions: [],
    currentWakeIndex: 0,
    isComplete: false,
  };
}

/**
 * Check if the night phase is complete.
 */
export function isNightComplete(state: NightState): boolean {
  return state.isComplete;
}

/**
 * Get all actions recorded during the night.
 */
export function getNightActions(state: NightState): readonly NightAction[] {
  return state.actions;
}

/**
 * Get the result of a specific night action.
 */
export function getNightActionResult(
  state: NightState,
  playerId: PlayerId,
): NightAction | undefined {
  return state.actions.find((a) => a.playerId === playerId);
}

// ─── Dawn Results ────────────────────────────────────────────────

/** Result of a single night action to be announced at dawn. */
export interface NightActionResult {
  readonly playerId: PlayerId;
  readonly characterId: string;
  readonly actionType: NightActionType;
  readonly targets: readonly PlayerId[];
  readonly result: string | null;
}

/** Dawn results to be announced to all players. */
export interface DawnResults {
  readonly nightNumber: number;
  readonly actions: readonly NightActionResult[];
  readonly deaths: readonly PlayerId[];
  readonly isComplete: boolean;
}

// ─── Validation ──────────────────────────────────────────────────

export interface NightPhaseValidationError {
  readonly code: string;
  readonly message: string;
}

/**
 * Validate a night action request.
 */
export function validateNightAction(
  state: NightState,
  action: NightAction,
  aliveCharacters: ReadonlyMap<PlayerId, Character>,
): NightPhaseValidationError | null {
  if (state.isComplete) {
    return { code: 'NIGHT_COMPLETE', message: 'Night phase is already complete.' };
  }

  const wakeOrder = getWakeOrder(state.nightNumber);
  const expectedEntry = wakeOrder[state.currentWakeIndex];

  if (!expectedEntry) {
    return { code: 'NO_MORE_ACTIONS', message: 'No more actions expected tonight.' };
  }

  if (expectedEntry.actionType !== action.actionType) {
    return {
      code: 'WRONG_ACTION_TYPE',
      message: `Expected ${expectedEntry.actionType}, got ${action.actionType}.`,
    };
  }

  if (!aliveCharacters.has(action.playerId)) {
    return { code: 'PLAYER_DEAD', message: 'Dead players cannot perform night actions.' };
  }

  const playerCharacter = aliveCharacters.get(action.playerId);
  if (playerCharacter && playerCharacter.id !== expectedEntry.characterId) {
    return {
      code: 'WRONG_CHARACTER',
      message: `Expected ${expectedEntry.characterId} to act, got ${playerCharacter.id}.`,
    };
  }

  return null;
}

/**
 * Get the expected action type for the current wake position.
 */
export function getExpectedActionType(state: NightState): NightActionType | null {
  if (state.isComplete) {
    return null;
  }

  const wakeOrder = getWakeOrder(state.nightNumber);
  const expectedEntry = wakeOrder[state.currentWakeIndex];

  return expectedEntry?.actionType ?? null;
}

/**
 * Get the expected character for the current wake position.
 */
export function getExpectedCharacter(state: NightState): string | null {
  if (state.isComplete) {
    return null;
  }

  const wakeOrder = getWakeOrder(state.nightNumber);
  const expectedEntry = wakeOrder[state.currentWakeIndex];

  return expectedEntry?.characterId ?? null;
}

/**
 * Get the number of actions remaining in the night.
 */
export function getActionsRemaining(state: NightState): number {
  const wakeOrder = getWakeOrder(state.nightNumber);
  return Math.max(0, wakeOrder.length - state.currentWakeIndex);
}

/**
 * Get the total number of actions expected for this night.
 */
export function getTotalActionsExpected(nightNumber: number): number {
  return getWakeOrder(nightNumber).length;
}

/**
 * Get the progress of the night phase as a percentage.
 */
export function getNightProgress(state: NightState): number {
  const total = getTotalActionsExpected(state.nightNumber);
  if (total === 0) return 100;

  return Math.round((state.currentWakeIndex / total) * 100);
}
