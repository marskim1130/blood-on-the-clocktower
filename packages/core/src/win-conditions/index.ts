import type { PlayerId, Team, Character, DeathCause } from '../types/index.js';

// ─── Win Condition Types ─────────────────────────────────────────

/** The specific reason the game ended. */
export type WinReason =
  | 'imp_executed'
  | 'mayor_endgame'
  | 'evil_majority'
  | 'saint_executed'
  | 'imp_starpass';

/** The result of a win condition check. Null means the game continues. */
export interface WinResult {
  readonly winner: Team;
  readonly reason: WinReason;
  /** Human-readable description of how the game ended. */
  readonly description: string;
}

/** Snapshot of a single player used for win condition evaluation. */
export interface WinCheckPlayer {
  readonly id: PlayerId;
  readonly character: Character | null;
  readonly isAlive: boolean;
}

/**
 * Context passed to win condition evaluation functions.
 *
 * This bundles every piece of state the checker needs, keeping the
 * functions pure and testable without coupling to the full GameState.
 */
export interface WinCheckContext {
  /** All players in the game. */
  readonly players: readonly WinCheckPlayer[];
  /** Current day number. */
  readonly dayNumber: number;
  /** Whether any player was executed today. */
  readonly executedToday: boolean;
  /** The player who was executed today, if any. */
  readonly executedPlayerId: PlayerId | null;
  /** Whether the Scarlet Woman promotion has already fired this round. */
  readonly scarletWomanTriggered: boolean;
  /** The cause of the most recent death, if checking after a death. */
  readonly lastDeathCause: DeathCause | null;
  /** The player who just died, if checking after a death. */
  readonly lastDeadPlayerId: PlayerId | null;
}

// ─── Initial Context ─────────────────────────────────────────────

export const INITIAL_WIN_CHECK_CONTEXT: WinCheckContext = {
  players: [],
  dayNumber: 0,
  executedToday: false,
  executedPlayerId: null,
  scarletWomanTriggered: false,
  lastDeathCause: null,
  lastDeadPlayerId: null,
};

// ─── Validation ──────────────────────────────────────────────────

export interface WinCheckValidationError {
  readonly code: string;
  readonly message: string;
}

/**
 * Validate that the win check context has enough data to evaluate.
 *
 * Rules:
 *  1. There must be at least 1 player.
 *  2. dayNumber must be a positive integer (game must have started).
 */
export function validateWinCheckContext(
  ctx: WinCheckContext,
): WinCheckValidationError | null {
  if (ctx.players.length === 0) {
    return { code: 'NO_PLAYERS', message: 'Cannot check win conditions with no players.' };
  }
  if (!Number.isInteger(ctx.dayNumber) || ctx.dayNumber < 1) {
    return { code: 'INVALID_DAY', message: 'Day number must be a positive integer.' };
  }
  return null;
}

// ─── Pure Evaluation Functions ────────────────────────────────────

/**
 * Check win conditions after a player is executed.
 *
 * Called immediately after a player dies from execution.
 *
 * Checks:
 *  - If the executed player was the **Imp** and no Scarlet Woman triggered -> good wins.
 *  - If the executed player was the **Saint** -> evil wins immediately.
 *
 * Returns `null` if no win condition is met (game continues).
 */
export function checkWinAfterExecution(ctx: WinCheckContext): WinResult | null {
  const error = validateWinCheckContext(ctx);
  if (error) return null;

  if (!ctx.executedPlayerId) return null;

  const executedPlayer = ctx.players.find((p) => p.id === ctx.executedPlayerId);
  if (!executedPlayer) return null;

  const characterId = executedPlayer.character?.id;

  // Saint executed -> evil wins immediately
  if (characterId === 'saint') {
    return {
      winner: 'evil',
      reason: 'saint_executed',
      description: `${executedPlayer.character?.name ?? 'Saint'} was executed. Evil wins!`,
    };
  }

  // Imp executed -> good wins (only if Scarlet Woman did NOT trigger)
  if (characterId === 'imp' && !ctx.scarletWomanTriggered) {
    return {
      winner: 'good',
      reason: 'imp_executed',
      description: 'The Imp was executed. Good wins!',
    };
  }

  return null;
}

/**
 * Check win conditions after a player dies at night.
 *
 * Called immediately after a night death is recorded.
 *
 * Checks:
 *  - If the dead player was the **Imp** (starpass / self-kill) and no Scarlet Woman
 *    triggered -> evil wins (Imp chose to die, passing the role, but with no
 *    Scarlet Woman the chain breaks).
 *  - If only 2 players remain alive after this death -> evil wins by majority.
 *
 * Returns `null` if no win condition is met (game continues).
 */
export function checkWinAfterNightDeath(ctx: WinCheckContext): WinResult | null {
  const error = validateWinCheckContext(ctx);
  if (error) return null;

  // Imp self-kill (starpass) with no Scarlet Woman - check first for more specific reason
  if (ctx.lastDeadPlayerId) {
    const deadPlayer = ctx.players.find((p) => p.id === ctx.lastDeadPlayerId);
    if (
      deadPlayer?.character?.id === 'imp' &&
      ctx.lastDeathCause === 'night_kill' &&
      !ctx.scarletWomanTriggered
    ) {
      return {
        winner: 'evil',
        reason: 'imp_starpass',
        description: 'The Imp killed themselves at night with no Scarlet Woman. Evil wins!',
      };
    }
  }

  // Evil wins by majority: only 2 alive
  const aliveCount = getAliveCount(ctx);
  if (aliveCount <= 2) {
    return {
      winner: 'evil',
      reason: 'evil_majority',
      description: `Only ${aliveCount} player(s) remain. Evil wins by majority!`,
    };
  }

  return null;
}

/**
 * Check win conditions at the end of the day phase when no execution occurred.
 *
 * Called when the day ends without any player being executed.
 *
 * Checks:
 *  - **Mayor endgame**: If exactly 3 players are alive and one of them is the
 *    Mayor, good wins. (The Mayor's ability: if no one is executed and only 3
 *    players live, the good team wins.)
 *  - **Evil majority**: If only 2 players are alive, evil wins.
 *
 * Returns `null` if no win condition is met (game continues to night).
 */
export function checkWinAtEndOfDay(ctx: WinCheckContext): WinResult | null {
  const error = validateWinCheckContext(ctx);
  if (error) return null;

  const aliveCount = getAliveCount(ctx);

  // Evil wins by majority: only 2 alive
  if (aliveCount <= 2) {
    return {
      winner: 'evil',
      reason: 'evil_majority',
      description: `Only ${aliveCount} player(s) remain. Evil wins by majority!`,
    };
  }

  // Mayor endgame: 3 alive, no execution, Mayor is among the living
  if (aliveCount === 3 && !ctx.executedToday) {
    const mayorAlive = ctx.players.some(
      (p) => p.isAlive && p.character?.id === 'mayor',
    );
    if (mayorAlive) {
      return {
        winner: 'good',
        reason: 'mayor_endgame',
        description:
          'No one was executed and the Mayor is alive with 3 players remaining. Good wins!',
      };
    }
  }

  return null;
}

/**
 * Unified entry point: given context, determines which check to run.
 *
 * This is a convenience dispatcher that delegates to the appropriate
 * specific check function based on the current game state.
 *
 * Priority order:
 *  1. If a death just occurred from execution -> `checkWinAfterExecution`
 *  2. If a death just occurred at night -> `checkWinAfterNightDeath`
 *  3. If end of day with no execution -> `checkWinAtEndOfDay`
 *  4. Otherwise -> no win condition met
 */
export function checkWinCondition(ctx: WinCheckContext): WinResult | null {
  // After an execution
  if (ctx.executedToday && ctx.executedPlayerId) {
    const result = checkWinAfterExecution(ctx);
    if (result) return result;

    // Even if the execution itself didn't end the game, check if
    // the resulting deaths left only 2 alive.
    if (getAliveCount(ctx) <= 2) {
      return {
        winner: 'evil',
        reason: 'evil_majority',
        description: `Only ${getAliveCount(ctx)} player(s) remain. Evil wins by majority!`,
      };
    }
  }

  // After a night death
  if (ctx.lastDeathCause === 'night_kill' && ctx.lastDeadPlayerId) {
    return checkWinAfterNightDeath(ctx);
  }

  // End of day with no execution
  if (!ctx.executedToday) {
    return checkWinAtEndOfDay(ctx);
  }

  return null;
}

// ─── Game Over Event Helpers ─────────────────────────────────────

/**
 * Create a GAME_OVER event payload from a WinResult.
 *
 * This is used by the game engine to broadcast the result to all clients.
 * The game reducer should set `phase: 'finished'` when processing this event.
 */
export interface GameOverEvent {
  readonly type: 'GAME_OVER';
  readonly winner: Team;
  readonly reason: WinReason;
  readonly description: string;
  /** All players with their characters revealed (post-game). */
  readonly revealedPlayers: readonly {
    readonly playerId: PlayerId;
    readonly character: Character | null;
  }[];
}

export function createGameOverEvent(
  result: WinResult,
  players: readonly WinCheckPlayer[],
): GameOverEvent {
  return {
    type: 'GAME_OVER',
    winner: result.winner,
    reason: result.reason,
    description: result.description,
    revealedPlayers: players.map((p) => ({
      playerId: p.id,
      character: p.character,
    })),
  };
}

// ─── Convenience Helpers ─────────────────────────────────────────

/** Count alive players in the context. */
export function getAliveCount(ctx: WinCheckContext): number {
  return ctx.players.filter((p) => p.isAlive).length;
}

/** Get all alive players. */
export function getAlivePlayers(ctx: WinCheckContext): readonly WinCheckPlayer[] {
  return ctx.players.filter((p) => p.isAlive);
}

/** Get all dead players. */
export function getDeadPlayers(ctx: WinCheckContext): readonly WinCheckPlayer[] {
  return ctx.players.filter((p) => !p.isAlive);
}

/** Find a player by character ID. Returns undefined if not found. */
export function findPlayerByCharacterId(
  ctx: WinCheckContext,
  characterId: string,
): WinCheckPlayer | undefined {
  return ctx.players.find((p) => p.character?.id === characterId);
}

/** Check if a specific character is alive in the game. */
export function isCharacterAlive(ctx: WinCheckContext, characterId: string): boolean {
  return ctx.players.some((p) => p.isAlive && p.character?.id === characterId);
}

/** Check if a specific team has any living members. */
export function isTeamAlive(ctx: WinCheckContext, team: Team): boolean {
  return ctx.players.some((p) => p.isAlive && p.character?.team === team);
}

/**
 * Get all players on a specific team (dead or alive).
 * Useful for post-game reveal.
 */
export function getPlayersByTeam(ctx: WinCheckContext, team: Team): readonly WinCheckPlayer[] {
  return ctx.players.filter((p) => p.character?.team === team);
}
