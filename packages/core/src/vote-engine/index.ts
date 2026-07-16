import type { PlayerId } from '../types/index.js';

// ─── Vote Engine Types ───────────────────────────────────────────

export type VotePhase = 'idle' | 'voting' | 'resolved';

export type VoteDecision = true | false; // true = guilty (execute), false = innocent (spare)

export interface VoteState {
  readonly phase: VotePhase;
  readonly nominatorId: PlayerId | null;
  readonly nomineeId: PlayerId | null;
  /** Votes cast in current round: voterId -> decision */
  readonly votes: ReadonlyMap<PlayerId, VoteDecision>;
  /** Players who have used their ghost vote this game (across all nomination rounds) */
  readonly ghostVotesUsed: ReadonlySet<PlayerId>;
  /** Number of alive players at time of nomination (for threshold calculation) */
  readonly alivePlayerCount: number;
  /** Computed result after voting completes; null while voting is in progress */
  readonly result: 'executed' | 'spared' | null;
}

export type VoteEvent =
  | { readonly type: 'NOMINATED'; readonly nominatorId: PlayerId; readonly nomineeId: PlayerId }
  | { readonly type: 'VOTE_CAST'; readonly voterId: PlayerId; readonly decision: VoteDecision }
  | { readonly type: 'VOTING_RESOLVED' }
  | { readonly type: 'VOTE_RESET' };

// ─── Initial State ───────────────────────────────────────────────

export const INITIAL_VOTE_STATE: VoteState = {
  phase: 'idle',
  nominatorId: null,
  nomineeId: null,
  votes: new Map(),
  ghostVotesUsed: new Set(),
  alivePlayerCount: 0,
  result: null,
};

// ─── Threshold Calculation ───────────────────────────────────────

/**
 * Blood on the Clocktower majority threshold:
 * ceil(alivePlayers / 2)
 *
 * Examples:
 *  - 5 alive  -> ceil(5/2) = 3
 *  - 6 alive  -> ceil(6/2) = 3
 *  - 7 alive  -> ceil(7/2) = 4
 *  - 12 alive -> ceil(12/2) = 6
 */
export function computeMajorityThreshold(alivePlayerCount: number): number {
  return Math.ceil(alivePlayerCount / 2);
}

// ─── Validation Helpers ──────────────────────────────────────────

export interface VoteValidationError {
  readonly code: string;
  readonly message: string;
}

/**
 * Validate a nomination request against current state.
 *
 * Rules (Blood on the Clocktower):
 *  1. Must be in 'idle' phase (no active nomination in progress).
 *  2. Nominator must be alive.
 *  3. Nominee must be alive.
 *  4. Nominator cannot nominate themselves.
 */
export function validateNomination(
  state: VoteState,
  nominatorId: PlayerId,
  nomineeId: PlayerId,
  isAlive: (id: PlayerId) => boolean,
): VoteValidationError | null {
  if (state.phase !== 'idle') {
    return { code: 'NOMINATION_IN_PROGRESS', message: 'A nomination is already in progress.' };
  }
  if (!isAlive(nominatorId)) {
    return { code: 'NOMINATOR_DEAD', message: 'Dead players cannot nominate.' };
  }
  if (!isAlive(nomineeId)) {
    return { code: 'NOMINEE_DEAD', message: 'Cannot nominate a dead player.' };
  }
  if (nominatorId === nomineeId) {
    return { code: 'SELF_NOMINATION', message: 'A player cannot nominate themselves.' };
  }
  return null;
}

/**
 * Validate a vote cast against current state.
 *
 * Rules:
 *  1. Must be in 'voting' phase.
 *  2. Voter must be alive, OR be a dead player with an unused ghost vote.
 *  3. Each player can vote at most once per nomination round.
 */
export function validateVote(
  state: VoteState,
  voterId: PlayerId,
  isAlive: (id: PlayerId) => boolean,
): VoteValidationError | null {
  if (state.phase !== 'voting') {
    return { code: 'NOT_VOTING_PHASE', message: 'Voting is not currently active.' };
  }
  if (state.votes.has(voterId)) {
    return { code: 'ALREADY_VOTED', message: 'This player has already voted in this round.' };
  }

  const alive = isAlive(voterId);
  if (!alive && state.ghostVotesUsed.has(voterId)) {
    return { code: 'GHOST_VOTE_SPENT', message: 'This dead player has already used their ghost vote.' };
  }

  return null;
}

// ─── Reducer ─────────────────────────────────────────────────────

/**
 * Pure reducer: applies a VoteEvent to the current VoteState and returns the next state.
 *
 * Invalid events leave the current state unchanged.
 */
export function voteReducer(
  state: VoteState,
  event: VoteEvent,
  ctx: {
    isAlive: (id: PlayerId) => boolean;
    alivePlayerCount: number;
  },
): VoteState {
  switch (event.type) {
    case 'NOMINATED': {
      const { isAlive, alivePlayerCount: aliveCount } = ctx;
      const error = validateNomination(state, event.nominatorId, event.nomineeId, isAlive);
      if (error) return state; // invalid nomination: no-op

      return {
        ...state,
        phase: 'voting',
        nominatorId: event.nominatorId,
        nomineeId: event.nomineeId,
        votes: new Map(),
        alivePlayerCount: aliveCount,
        result: null,
      };
    }

    case 'VOTE_CAST': {
      const { isAlive } = ctx;
      const error = validateVote(state, event.voterId, isAlive);
      if (error) return state; // invalid vote: no-op

      const newVotes = new Map(state.votes);
      newVotes.set(event.voterId, event.decision);

      // Track ghost vote usage for dead players
      const newGhostVotesUsed = new Set(state.ghostVotesUsed);
      if (!isAlive(event.voterId)) {
        newGhostVotesUsed.add(event.voterId);
      }

      return {
        ...state,
        votes: newVotes,
        ghostVotesUsed: newGhostVotesUsed,
      };
    }

    case 'VOTING_RESOLVED': {
      if (state.phase !== 'voting') return state;

      const yesVotes = Array.from(state.votes.values()).filter((d) => d === true).length;
      const threshold = computeMajorityThreshold(state.alivePlayerCount);
      const executed = yesVotes >= threshold;

      return {
        ...state,
        phase: 'resolved',
        result: executed ? 'executed' : 'spared',
      };
    }

    case 'VOTE_RESET': {
      return {
        ...INITIAL_VOTE_STATE,
        // Preserve ghost votes across the entire game
        ghostVotesUsed: state.ghostVotesUsed,
      };
    }

    default:
      return state;
  }
}

// ─── Convenience: count votes ────────────────────────────────────

export function countVotes(votes: ReadonlyMap<PlayerId, VoteDecision>): {
  readonly yes: number;
  readonly no: number;
} {
  let yes = 0;
  let no = 0;
  for (const decision of votes.values()) {
    if (decision) yes++;
    else no++;
  }
  return { yes, no };
}
