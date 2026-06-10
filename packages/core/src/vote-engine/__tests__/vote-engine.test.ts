import { describe, it, expect } from 'vitest';
import {
  voteReducer,
  computeMajorityThreshold,
  countVotes,
  validateNomination,
  validateVote,
  INITIAL_VOTE_STATE,
  type VoteState,
  type VoteDecision,
} from '../index.js';
import { createPlayerId, type PlayerId } from '../../types/index.js';

// ─── Helpers ─────────────────────────────────────────────────────

function p(id: string): PlayerId {
  return createPlayerId(id);
}

/** Build a set of alive player IDs for isAlive lookups. */
function makeAliveSet(...ids: string[]): Set<PlayerId> {
  return new Set(ids.map(p));
}

function isAliveFactory(alive: Set<PlayerId>) {
  return (id: PlayerId) => alive.has(id);
}

// ─── computeMajorityThreshold ────────────────────────────────────

describe('computeMajorityThreshold', () => {
  it('returns ceil(n/2) for odd counts', () => {
    expect(computeMajorityThreshold(5)).toBe(3);
    expect(computeMajorityThreshold(7)).toBe(4);
    expect(computeMajorityThreshold(1)).toBe(1);
  });

  it('returns n/2 for even counts', () => {
    expect(computeMajorityThreshold(6)).toBe(3);
    expect(computeMajorityThreshold(12)).toBe(6);
    expect(computeMajorityThreshold(2)).toBe(1);
  });

  it('returns 0 for 0 players', () => {
    expect(computeMajorityThreshold(0)).toBe(0);
  });
});

// ─── countVotes ──────────────────────────────────────────────────

describe('countVotes', () => {
  it('counts yes and no votes', () => {
    const votes = new Map<PlayerId, VoteDecision>([
      [p('a'), true],
      [p('b'), false],
      [p('c'), true],
    ]);
    expect(countVotes(votes)).toEqual({ yes: 2, no: 1 });
  });

  it('returns 0/0 for empty map', () => {
    expect(countVotes(new Map())).toEqual({ yes: 0, no: 0 });
  });
});

// ─── validateNomination ──────────────────────────────────────────

describe('validateNomination', () => {
  const alive = makeAliveSet('alice', 'bob');
  const isAlive = isAliveFactory(alive);

  it('returns null for valid nomination', () => {
    expect(validateNomination(INITIAL_VOTE_STATE, p('alice'), p('bob'), isAlive)).toBeNull();
  });

  it('rejects when a nomination is already in progress', () => {
    const state: VoteState = { ...INITIAL_VOTE_STATE, phase: 'voting' };
    const err = validateNomination(state, p('alice'), p('bob'), isAlive);
    expect(err?.code).toBe('NOMINATION_IN_PROGRESS');
  });

  it('rejects dead nominator', () => {
    const err = validateNomination(INITIAL_VOTE_STATE, p('deadguy'), p('bob'), isAlive);
    expect(err?.code).toBe('NOMINATOR_DEAD');
  });

  it('rejects dead nominee', () => {
    const err = validateNomination(INITIAL_VOTE_STATE, p('alice'), p('deadguy'), isAlive);
    expect(err?.code).toBe('NOMINEE_DEAD');
  });

  it('rejects self-nomination', () => {
    const err = validateNomination(INITIAL_VOTE_STATE, p('alice'), p('alice'), isAlive);
    expect(err?.code).toBe('SELF_NOMINATION');
  });
});

// ─── validateVote ────────────────────────────────────────────────

describe('validateVote', () => {
  const alive = makeAliveSet('alice', 'bob', 'ghost_dead');
  const isAlive = isAliveFactory(alive);

  const votingState: VoteState = {
    ...INITIAL_VOTE_STATE,
    phase: 'voting',
    nominatorId: p('alice'),
    nomineeId: p('bob'),
    alivePlayerCount: 3,
  };

  it('returns null for valid vote from alive player', () => {
    expect(validateVote(votingState, p('alice'), isAlive)).toBeNull();
  });

  it('rejects vote when not in voting phase', () => {
    const err = validateVote(INITIAL_VOTE_STATE, p('alice'), isAlive);
    expect(err?.code).toBe('NOT_VOTING_PHASE');
  });

  it('rejects duplicate vote', () => {
    const state: VoteState = {
      ...votingState,
      votes: new Map([[p('alice'), true]]),
    };
    const err = validateVote(state, p('alice'), isAlive);
    expect(err?.code).toBe('ALREADY_VOTED');
  });

  it('rejects dead player with spent ghost vote', () => {
    const deadAlive = makeAliveSet('alice', 'bob'); // ghost_dead is NOT alive
    const state: VoteState = {
      ...votingState,
      ghostVotesUsed: new Set([p('ghost_dead')]),
    };
    const err = validateVote(state, p('ghost_dead'), isAliveFactory(deadAlive));
    expect(err?.code).toBe('GHOST_VOTE_SPENT');
  });
});

// ─── voteReducer: NOMINATED ─────────────────────────────────────

describe('voteReducer: NOMINATED', () => {
  const alive = makeAliveSet('alice', 'bob', 'charlie');
  const isAlive = isAliveFactory(alive);

  it('transitions from idle to voting phase on valid nomination', () => {
    const next = voteReducer(
      INITIAL_VOTE_STATE,
      { type: 'NOMINATED', nominatorId: p('alice'), nomineeId: p('bob') },
      { isAlive, alivePlayerCount: 3 },
    );

    expect(next.phase).toBe('voting');
    expect(next.nominatorId).toBe(p('alice'));
    expect(next.nomineeId).toBe(p('bob'));
    expect(next.alivePlayerCount).toBe(3);
    expect(next.votes.size).toBe(0);
    expect(next.result).toBeNull();
  });

  it('ignores invalid nomination (self-nomination)', () => {
    const next = voteReducer(
      INITIAL_VOTE_STATE,
      { type: 'NOMINATED', nominatorId: p('alice'), nomineeId: p('alice') },
      { isAlive, alivePlayerCount: 3 },
    );

    expect(next).toBe(INITIAL_VOTE_STATE);
  });

  it('ignores nomination when already voting', () => {
    const votingState: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      nominatorId: p('alice'),
      nomineeId: p('bob'),
    };

    const next = voteReducer(
      votingState,
      { type: 'NOMINATED', nominatorId: p('charlie'), nomineeId: p('alice') },
      { isAlive, alivePlayerCount: 3 },
    );

    expect(next).toBe(votingState);
  });

  it('ignores nomination from dead player', () => {
    const deadAlive = makeAliveSet('bob', 'charlie');
    const next = voteReducer(
      INITIAL_VOTE_STATE,
      { type: 'NOMINATED', nominatorId: p('alice'), nomineeId: p('bob') },
      { isAlive: isAliveFactory(deadAlive), alivePlayerCount: 2 },
    );

    expect(next).toBe(INITIAL_VOTE_STATE);
  });
});

// ─── voteReducer: VOTE_CAST ──────────────────────────────────────

describe('voteReducer: VOTE_CAST', () => {
  const alive = makeAliveSet('alice', 'bob', 'charlie', 'deadguy');
  const isAlive = isAliveFactory(alive);

  const votingState: VoteState = {
    ...INITIAL_VOTE_STATE,
    phase: 'voting',
    nominatorId: p('alice'),
    nomineeId: p('bob'),
    alivePlayerCount: 3,
  };

  it('records a thumbs-up (guilty) vote', () => {
    const next = voteReducer(votingState, { type: 'VOTE_CAST', voterId: p('alice'), decision: true }, { isAlive, alivePlayerCount: 4 });
    expect(next.votes.get(p('alice'))).toBe(true);
  });

  it('records a thumbs-down (innocent) vote', () => {
    const next = voteReducer(votingState, { type: 'VOTE_CAST', voterId: p('bob'), decision: false }, { isAlive, alivePlayerCount: 4 });
    expect(next.votes.get(p('bob'))).toBe(false);
  });

  it('ignores duplicate vote from same player', () => {
    const state: VoteState = { ...votingState, votes: new Map([[p('alice'), true]]) };
    const next = voteReducer(state, { type: 'VOTE_CAST', voterId: p('alice'), decision: false }, { isAlive, alivePlayerCount: 4 });
    // Should remain unchanged
    expect(next.votes.get(p('alice'))).toBe(true);
  });

  it('ignores vote when not in voting phase', () => {
    const next = voteReducer(INITIAL_VOTE_STATE, { type: 'VOTE_CAST', voterId: p('alice'), decision: true }, { isAlive, alivePlayerCount: 4 });
    expect(next).toBe(INITIAL_VOTE_STATE);
  });

  it('records ghost vote from dead player with unused ghost vote', () => {
    // deadguy is in the alive set but we want to test a dead player scenario
    const deadAlive = makeAliveSet('alice', 'bob', 'charlie'); // deadguy is dead
    const deadIsAlive = isAliveFactory(deadAlive);
    const state: VoteState = { ...votingState, ghostVotesUsed: new Set<PlayerId>() };

    const next = voteReducer(state, { type: 'VOTE_CAST', voterId: p('deadguy'), decision: true }, { isAlive: deadIsAlive, alivePlayerCount: 3 });

    expect(next.votes.get(p('deadguy'))).toBe(true);
    expect(next.ghostVotesUsed.has(p('deadguy'))).toBe(true);
  });

  it('rejects ghost vote from dead player whose ghost vote is already spent', () => {
    const deadAlive = makeAliveSet('alice', 'bob', 'charlie');
    const deadIsAlive = isAliveFactory(deadAlive);
    const state: VoteState = {
      ...votingState,
      ghostVotesUsed: new Set([p('deadguy')]),
    };

    const next = voteReducer(state, { type: 'VOTE_CAST', voterId: p('deadguy'), decision: false }, { isAlive: deadIsAlive, alivePlayerCount: 3 });

    // Should be no-op
    expect(next.votes.has(p('deadguy'))).toBe(false);
  });
});

// ─── voteReducer: VOTING_RESOLVED ────────────────────────────────

describe('voteReducer: VOTING_RESOLVED', () => {
  it('executes when yes votes >= majority threshold (5 alive, need 3)', () => {
    const state: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      nominatorId: p('a'),
      nomineeId: p('b'),
      alivePlayerCount: 5,
      votes: new Map<PlayerId, VoteDecision>([
        [p('1'), true],
        [p('2'), true],
        [p('3'), true],
        [p('4'), false],
      ]),
    };

    const next = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next.phase).toBe('resolved');
    expect(next.result).toBe('executed');
  });

  it('spares when yes votes < majority threshold (5 alive, need 3)', () => {
    const state: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      nominatorId: p('a'),
      nomineeId: p('b'),
      alivePlayerCount: 5,
      votes: new Map<PlayerId, VoteDecision>([
        [p('1'), true],
        [p('2'), true],
        [p('3'), false],
        [p('4'), false],
      ]),
    };

    const next = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next.phase).toBe('resolved');
    expect(next.result).toBe('spared');
  });

  it('spares on exact tie (6 alive, threshold 3, 3 yes / 3 no)', () => {
    const state: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      nominatorId: p('a'),
      nomineeId: p('b'),
      alivePlayerCount: 6,
      votes: new Map<PlayerId, VoteDecision>([
        [p('1'), true],
        [p('2'), true],
        [p('3'), true],
        [p('4'), false],
        [p('5'), false],
        [p('6'), false],
      ]),
    };

    // 3 yes >= ceil(6/2)=3 -> executed (exact majority counts)
    const next = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next.result).toBe('executed');
  });

  it('spares when no votes at all', () => {
    const state: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      nominatorId: p('a'),
      nomineeId: p('b'),
      alivePlayerCount: 5,
      votes: new Map(),
    };

    const next = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next.result).toBe('spared');
  });

  it('ignores VOTING_RESOLVED when not in voting phase', () => {
    const next = voteReducer(INITIAL_VOTE_STATE, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next).toBe(INITIAL_VOTE_STATE);
  });

  it('handles 1 alive player (threshold = 1)', () => {
    const state: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      alivePlayerCount: 1,
      votes: new Map<PlayerId, VoteDecision>([[p('sole'), true]]),
    };

    const next = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next.result).toBe('executed');
  });
});

// ─── voteReducer: VOTE_RESET ─────────────────────────────────────

describe('voteReducer: VOTE_RESET', () => {
  it('resets to idle but preserves ghostVotesUsed', () => {
    const state: VoteState = {
      phase: 'resolved',
      nominatorId: p('a'),
      nomineeId: p('b'),
      votes: new Map([[p('1'), true]]),
      ghostVotesUsed: new Set([p('deadguy')]),
      alivePlayerCount: 5,
      result: 'executed',
    };

    const next = voteReducer(state, { type: 'VOTE_RESET' }, { isAlive: () => true, alivePlayerCount: 5 });

    expect(next.phase).toBe('idle');
    expect(next.nominatorId).toBeNull();
    expect(next.nomineeId).toBeNull();
    expect(next.votes.size).toBe(0);
    expect(next.result).toBeNull();
    // Ghost votes persist across the whole game
    expect(next.ghostVotesUsed.has(p('deadguy'))).toBe(true);
  });
});

// ─── Full Nomination Round Integration ───────────────────────────

describe('full nomination round', () => {
  const alive = makeAliveSet('storyteller', 'alice', 'bob', 'charlie', 'dave');
  const isAlive = isAliveFactory(alive);

  it('completes a full round: nominate -> vote -> resolve -> reset', () => {
    // 1. Start in idle
    let state: VoteState = { ...INITIAL_VOTE_STATE };

    // 2. storyteller nominates alice
    state = voteReducer(
      state,
      { type: 'NOMINATED', nominatorId: p('storyteller'), nomineeId: p('alice') },
      { isAlive, alivePlayerCount: 5 },
    );
    expect(state.phase).toBe('voting');
    expect(state.nomineeId).toBe(p('alice'));

    // 3. Everyone votes: 3 guilty, 1 innocent
    for (const [voter, decision] of [
      ['storyteller', true],
      ['bob', true],
      ['charlie', true],
      ['dave', false],
    ] as const) {
      state = voteReducer(state, { type: 'VOTE_CAST', voterId: p(voter), decision }, { isAlive, alivePlayerCount: 4 });
    }

    expect(state.votes.size).toBe(4);

    // 4. Resolve voting
    state = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(state.phase).toBe('resolved');
    expect(state.result).toBe('executed'); // 3 >= ceil(5/2)=3

    // 5. Reset for next nomination
    state = voteReducer(state, { type: 'VOTE_RESET' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(state.phase).toBe('idle');
    expect(state.result).toBeNull();

    // 6. New nomination: bob nominates charlie
    state = voteReducer(
      state,
      { type: 'NOMINATED', nominatorId: p('bob'), nomineeId: p('charlie') },
      { isAlive, alivePlayerCount: 5 },
    );
    expect(state.phase).toBe('voting');
    expect(state.nomineeId).toBe(p('charlie'));
  });

  it('handles ghost vote: dead player uses their one ghost vote across rounds', () => {
    const deadGuy = p('deadguy');
    // deadguy is dead, others alive
    const aliveWithDead = makeAliveSet('alice', 'bob', 'charlie');
    const isAliveWithDead = isAliveFactory(aliveWithDead);

    let state: VoteState = {
      ...INITIAL_VOTE_STATE,
      ghostVotesUsed: new Set<PlayerId>(), // deadguy has NOT used ghost vote yet
    };

    // Round 1: nomination
    state = voteReducer(
      state,
      { type: 'NOMINATED', nominatorId: p('alice'), nomineeId: p('bob') },
      { isAlive: isAliveWithDead, alivePlayerCount: 3 },
    );

    // deadguy casts ghost vote (thumbs up)
    state = voteReducer(state, { type: 'VOTE_CAST', voterId: deadGuy, decision: true }, { isAlive: isAliveWithDead, alivePlayerCount: 3 });
    expect(state.votes.get(deadGuy)).toBe(true);
    expect(state.ghostVotesUsed.has(deadGuy)).toBe(true);

    // Resolve and reset
    state = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    state = voteReducer(state, { type: 'VOTE_RESET' }, { isAlive: () => true, alivePlayerCount: 5 });

    // Round 2: deadguy tries to vote again
    state = voteReducer(
      state,
      { type: 'NOMINATED', nominatorId: p('charlie'), nomineeId: p('alice') },
      { isAlive: isAliveWithDead, alivePlayerCount: 3 },
    );

    // Ghost vote is spent from round 1, should be rejected
    state = voteReducer(state, { type: 'VOTE_CAST', voterId: deadGuy, decision: false }, { isAlive: isAliveWithDead, alivePlayerCount: 3 });
    expect(state.votes.has(deadGuy)).toBe(false); // rejected
  });
});

// ─── Edge Cases ──────────────────────────────────────────────────

describe('edge cases', () => {
  it('handles single alive player executing themselves via nomination', () => {
    const alive = makeAliveSet('lone');
    const isAlive = isAliveFactory(alive);

    // lone cannot self-nominate
    const err = validateNomination(INITIAL_VOTE_STATE, p('lone'), p('lone'), isAlive);
    expect(err?.code).toBe('SELF_NOMINATION');
  });

  it('all players vote innocent -> spared', () => {
    const state: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      alivePlayerCount: 5,
      votes: new Map<PlayerId, VoteDecision>([
        [p('1'), false],
        [p('2'), false],
        [p('3'), false],
        [p('4'), false],
        [p('5'), false],
      ]),
    };

    const next = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next.result).toBe('spared');
  });

  it('all players vote guilty -> executed', () => {
    const state: VoteState = {
      ...INITIAL_VOTE_STATE,
      phase: 'voting',
      alivePlayerCount: 5,
      votes: new Map<PlayerId, VoteDecision>([
        [p('1'), true],
        [p('2'), true],
        [p('3'), true],
        [p('4'), true],
        [p('5'), true],
      ]),
    };

    const next = voteReducer(state, { type: 'VOTING_RESOLVED' }, { isAlive: () => true, alivePlayerCount: 5 });
    expect(next.result).toBe('executed');
  });

  it('unknown event type is a no-op', () => {
    // @ts-expect-error -- testing unknown event type
    const next = voteReducer(INITIAL_VOTE_STATE, { type: 'UNKNOWN' });
    expect(next).toBe(INITIAL_VOTE_STATE);
  });
});
