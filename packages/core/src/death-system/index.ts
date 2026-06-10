import type { PlayerId, Character } from '../types/index.js';

// ─── Death System Types ──────────────────────────────────────────

/** How a player died. */
export type DeathCause = 'execution' | 'night_kill' | 'ability';

/** Record of a single player's death. */
export interface DeathRecord {
  readonly playerId: PlayerId;
  readonly cause: DeathCause;
  /** The game day number when the death occurred. */
  readonly dayNumber: number;
  /** The player who caused this death, if applicable (e.g. the demon for a night kill). */
  readonly killedBy: PlayerId | null;
}

/** Broadcast-ready announcement sent to all players when someone dies. */
export interface DeathAnnouncement {
  readonly playerId: PlayerId;
  readonly cause: DeathCause;
  readonly dayNumber: number;
}

/** Special character ability triggers that fire on death. */
export interface DeathTrigger {
  readonly type: 'saint_execution' | 'scarlet_woman' | 'ravenkeeper' | 'undertaker';
  /** The player whose death caused this trigger. */
  readonly sourcePlayerId: PlayerId;
  /** Additional context: the character of the dead player (for undertaker/ravenkeeper). */
  readonly deadPlayerCharacter?: Character;
  /** Additional context: the player whose character should be revealed. */
  readonly targetPlayerId?: PlayerId;
}

// ─── Death State ─────────────────────────────────────────────────

export interface DeathState {
  /** All death records, keyed by the dead player's ID. */
  readonly deaths: ReadonlyMap<PlayerId, DeathRecord>;
  /** Dead players who still have their one ghost vote available. */
  readonly ghostVotesRemaining: ReadonlySet<PlayerId>;
}

export const INITIAL_DEATH_STATE: DeathState = {
  deaths: new Map(),
  ghostVotesRemaining: new Set(),
};

// ─── Death Events ────────────────────────────────────────────────

export type DeathEvent =
  | {
      readonly type: 'PLAYER_DIED';
      readonly playerId: PlayerId;
      readonly cause: DeathCause;
      readonly dayNumber: number;
      readonly killedBy?: PlayerId;
    }
  | {
      readonly type: 'GHOST_VOTE_CAST';
      readonly playerId: PlayerId;
    };

// ─── Validation ──────────────────────────────────────────────────

export interface DeathValidationError {
  readonly code: string;
  readonly message: string;
}

/**
 * Validate a PLAYER_DIED event.
 *
 * Rules:
 *  1. Player must not already be dead (no duplicate death records).
 *  2. dayNumber must be a positive integer.
 */
export function validateDeath(
  state: DeathState,
  playerId: PlayerId,
  dayNumber: number,
): DeathValidationError | null {
  if (state.deaths.has(playerId)) {
    return { code: 'ALREADY_DEAD', message: 'This player is already dead.' };
  }
  if (!Number.isInteger(dayNumber) || dayNumber < 1) {
    return { code: 'INVALID_DAY', message: 'Day number must be a positive integer.' };
  }
  return null;
}

/**
 * Validate a GHOST_VOTE_CAST event.
 *
 * Rules:
 *  1. Player must be dead (must exist in deaths map).
 *  2. Player must still have an unused ghost vote.
 */
export function validateGhostVoteCast(
  state: DeathState,
  playerId: PlayerId,
): DeathValidationError | null {
  if (!state.deaths.has(playerId)) {
    return { code: 'NOT_DEAD', message: 'Only dead players can use ghost votes.' };
  }
  if (!state.ghostVotesRemaining.has(playerId)) {
    return { code: 'GHOST_VOTE_SPENT', message: 'This player has already used their ghost vote.' };
  }
  return null;
}

// ─── Reducer ─────────────────────────────────────────────────────

/**
 * Pure reducer: applies a DeathEvent to the current DeathState and returns the next state.
 *
 * This follows the same pattern as `voteReducer` in vote-engine/index.ts.
 */
export function deathReducer(
  state: DeathState,
  event: DeathEvent,
): DeathState {
  switch (event.type) {
    case 'PLAYER_DIED': {
      const error = validateDeath(state, event.playerId, event.dayNumber);
      if (error) return state; // invalid: no-op

      const newDeaths = new Map(state.deaths);
      newDeaths.set(event.playerId, {
        playerId: event.playerId,
        cause: event.cause,
        dayNumber: event.dayNumber,
        killedBy: event.killedBy ?? null,
      });

      // Dead players receive exactly 1 ghost vote for the rest of the game
      const newGhostVotes = new Set(state.ghostVotesRemaining);
      newGhostVotes.add(event.playerId);

      return {
        deaths: newDeaths,
        ghostVotesRemaining: newGhostVotes,
      };
    }

    case 'GHOST_VOTE_CAST': {
      const error = validateGhostVoteCast(state, event.playerId);
      if (error) return state; // invalid: no-op

      const newGhostVotes = new Set(state.ghostVotesRemaining);
      newGhostVotes.delete(event.playerId);

      return {
        ...state,
        ghostVotesRemaining: newGhostVotes,
      };
    }

    default:
      return state;
  }
}

// ─── Death Trigger Computation ───────────────────────────────────

export interface DeathTriggerContext {
  /** The character of the player who died (if known). */
  readonly deadPlayerCharacter?: Character;
  /** Total number of players still alive AFTER this death. */
  readonly alivePlayerCountAfterDeath: number;
  /** Characters of all players in the game, for Scarlet Woman lookup. */
  readonly characters?: ReadonlyMap<PlayerId, Character>;
}

/**
 * Compute death triggers that fire when a player dies.
 *
 * This is a pure function that examines the death context and returns
 * any special ability triggers that should be resolved by the game engine.
 *
 * Triggers:
 *  - **saint_execution**: If a Saint is executed, evil wins immediately.
 *  - **scarlet_woman**: If the Imp dies with 5+ players alive, a Scarlet Woman
 *    (if present) becomes the new Imp.
 *  - **ravenkeeper**: If a Ravenkeeper dies at night, they may learn the character
 *    of one player.
 *  - **undertaker**: If a player is executed during the day, the Undertaker
 *    (if alive) learns that player's character.
 */
export function computeDeathTriggers(
  record: DeathRecord,
  ctx: DeathTriggerContext,
): readonly DeathTrigger[] {
  const triggers: DeathTrigger[] = [];

  // Saint execution: evil wins immediately
  if (
    record.cause === 'execution' &&
    ctx.deadPlayerCharacter?.id === 'saint'
  ) {
    triggers.push({
      type: 'saint_execution',
      sourcePlayerId: record.playerId,
      deadPlayerCharacter: ctx.deadPlayerCharacter,
    });
  }

  // Scarlet Woman: if Imp dies with 5+ players alive
  if (
    ctx.deadPlayerCharacter?.id === 'imp' &&
    ctx.alivePlayerCountAfterDeath >= 5 &&
    ctx.characters
  ) {
    // Find a Scarlet Woman player
    for (const [playerId, character] of ctx.characters) {
      if (character.id === 'scarlet_woman') {
        triggers.push({
          type: 'scarlet_woman',
          sourcePlayerId: record.playerId,
          targetPlayerId: playerId,
        });
        break; // Only one Scarlet Woman trigger
      }
    }
  }

  // Ravenkeeper: dies at night, can learn a character
  if (
    record.cause === 'night_kill' &&
    ctx.deadPlayerCharacter?.id === 'ravenkeeper'
  ) {
    triggers.push({
      type: 'ravenkeeper',
      sourcePlayerId: record.playerId,
      deadPlayerCharacter: ctx.deadPlayerCharacter,
    });
  }

  // Undertaker: learns character of player executed during the day
  if (record.cause === 'execution') {
    triggers.push({
      type: 'undertaker',
      sourcePlayerId: record.playerId,
      ...(ctx.deadPlayerCharacter ? { deadPlayerCharacter: ctx.deadPlayerCharacter } : {}),
    });
  }

  return triggers;
}

// ─── Convenience Helpers ─────────────────────────────────────────

/** Check if a player is dead. */
export function isDead(state: DeathState, playerId: PlayerId): boolean {
  return state.deaths.has(playerId);
}

/** Get the death record for a player, or undefined if alive. */
export function getDeathRecord(
  state: DeathState,
  playerId: PlayerId,
): DeathRecord | undefined {
  return state.deaths.get(playerId);
}

/** Check if a dead player still has their ghost vote available. */
export function hasGhostVote(state: DeathState, playerId: PlayerId): boolean {
  return state.ghostVotesRemaining.has(playerId);
}

/** Get total number of deaths. */
export function getDeathCount(state: DeathState): number {
  return state.deaths.size;
}

/** Get all death records matching a specific cause. */
export function getDeathsByCause(
  state: DeathState,
  cause: DeathCause,
): readonly DeathRecord[] {
  return Array.from(state.deaths.values()).filter((d) => d.cause === cause);
}

/** Get all deaths that occurred on a specific day. */
export function getDeathsOnDay(
  state: DeathState,
  dayNumber: number,
): readonly DeathRecord[] {
  return Array.from(state.deaths.values()).filter((d) => d.dayNumber === dayNumber);
}

/**
 * Create a death announcement from a death record.
 * This is the broadcast-ready message sent to all players.
 */
export function createDeathAnnouncement(record: DeathRecord): DeathAnnouncement {
  return {
    playerId: record.playerId,
    cause: record.cause,
    dayNumber: record.dayNumber,
  };
}
