import { describe, it, expect } from 'vitest';
import {
  deathReducer,
  computeDeathTriggers,
  validateDeath,
  validateGhostVoteCast,
  isDead,
  getDeathRecord,
  hasGhostVote,
  getDeathCount,
  getDeathsByCause,
  getDeathsOnDay,
  createDeathAnnouncement,
  INITIAL_DEATH_STATE,
  type DeathState,
  type DeathRecord,
  type DeathTriggerContext,
} from '../index.js';
import { createPlayerId, type PlayerId, type Character } from '../../types/index.js';

// ─── Helpers ─────────────────────────────────────────────────────

function p(id: string): PlayerId {
  return createPlayerId(id);
}

function makeCharacter(id: string, team: 'good' | 'evil' = 'good'): Character {
  return { id, name: id, team, ability: `${id}-ability` };
}

// ─── INITIAL_DEATH_STATE ─────────────────────────────────────────

describe('INITIAL_DEATH_STATE', () => {
  it('starts with empty deaths and ghost votes', () => {
    expect(INITIAL_DEATH_STATE.deaths.size).toBe(0);
    expect(INITIAL_DEATH_STATE.ghostVotesRemaining.size).toBe(0);
  });
});

// ─── validateDeath ───────────────────────────────────────────────

describe('validateDeath', () => {
  it('returns null for a valid death', () => {
    expect(validateDeath(INITIAL_DEATH_STATE, p('alice'), 1)).toBeNull();
  });

  it('rejects duplicate death (player already dead)', () => {
    const state: DeathState = {
      deaths: new Map([
        [p('alice'), { playerId: p('alice'), cause: 'execution', dayNumber: 1, killedBy: null }],
      ]),
      ghostVotesRemaining: new Set([p('alice')]),
    };
    const err = validateDeath(state, p('alice'), 2);
    expect(err?.code).toBe('ALREADY_DEAD');
  });

  it('rejects invalid day number (0)', () => {
    const err = validateDeath(INITIAL_DEATH_STATE, p('alice'), 0);
    expect(err?.code).toBe('INVALID_DAY');
  });

  it('rejects negative day number', () => {
    const err = validateDeath(INITIAL_DEATH_STATE, p('alice'), -1);
    expect(err?.code).toBe('INVALID_DAY');
  });

  it('rejects fractional day number', () => {
    const err = validateDeath(INITIAL_DEATH_STATE, p('alice'), 1.5);
    expect(err?.code).toBe('INVALID_DAY');
  });
});

// ─── validateGhostVoteCast ───────────────────────────────────────

describe('validateGhostVoteCast', () => {
  const stateWithDead: DeathState = {
    deaths: new Map([
      [p('deadguy'), { playerId: p('deadguy'), cause: 'execution', dayNumber: 1, killedBy: null }],
    ]),
    ghostVotesRemaining: new Set([p('deadguy')]),
  };

  it('returns null for valid ghost vote cast', () => {
    expect(validateGhostVoteCast(stateWithDead, p('deadguy'))).toBeNull();
  });

  it('rejects ghost vote from alive player', () => {
    const err = validateGhostVoteCast(stateWithDead, p('alice'));
    expect(err?.code).toBe('NOT_DEAD');
  });

  it('rejects ghost vote from dead player with spent ghost vote', () => {
    const spent: DeathState = {
      ...stateWithDead,
      ghostVotesRemaining: new Set(), // ghost vote already used
    };
    const err = validateGhostVoteCast(spent, p('deadguy'));
    expect(err?.code).toBe('GHOST_VOTE_SPENT');
  });
});

// ─── deathReducer: PLAYER_DIED ───────────────────────────────────

describe('deathReducer: PLAYER_DIED', () => {
  it('records an execution death', () => {
    const next = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });

    expect(next.deaths.size).toBe(1);
    const record = next.deaths.get(p('alice'))!;
    expect(record.playerId).toBe(p('alice'));
    expect(record.cause).toBe('execution');
    expect(record.dayNumber).toBe(1);
    expect(record.killedBy).toBeNull();
  });

  it('records a night kill with killer info', () => {
    const next = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('bob'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    });

    const record = next.deaths.get(p('bob'))!;
    expect(record.cause).toBe('night_kill');
    expect(record.dayNumber).toBe(2);
    expect(record.killedBy).toBe(p('imp'));
  });

  it('records an ability death', () => {
    const next = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('charlie'),
      cause: 'ability',
      dayNumber: 3,
    });

    const record = next.deaths.get(p('charlie'))!;
    expect(record.cause).toBe('ability');
  });

  it('grants a ghost vote to the dead player', () => {
    const next = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });

    expect(next.ghostVotesRemaining.has(p('alice'))).toBe(true);
  });

  it('ignores death of already-dead player', () => {
    const first = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });

    const second = deathReducer(first, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'night_kill',
      dayNumber: 2,
    });

    // Should be unchanged (no-op)
    expect(second).toBe(first);
    expect(second.deaths.size).toBe(1);
    expect(second.deaths.get(p('alice'))!.cause).toBe('execution');
  });

  it('ignores death with invalid day number', () => {
    const next = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 0,
    });

    expect(next).toBe(INITIAL_DEATH_STATE);
  });

  it('tracks multiple deaths independently', () => {
    let state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('bob'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    });
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('charlie'),
      cause: 'ability',
      dayNumber: 2,
    });

    expect(state.deaths.size).toBe(3);
    expect(state.ghostVotesRemaining.size).toBe(3);
    expect(state.deaths.get(p('alice'))!.cause).toBe('execution');
    expect(state.deaths.get(p('bob'))!.cause).toBe('night_kill');
    expect(state.deaths.get(p('charlie'))!.cause).toBe('ability');
  });
});

// ─── deathReducer: GHOST_VOTE_CAST ──────────────────────────────

describe('deathReducer: GHOST_VOTE_CAST', () => {
  const stateWithDead: DeathState = {
    deaths: new Map([
      [p('deadguy'), { playerId: p('deadguy'), cause: 'execution', dayNumber: 1, killedBy: null }],
    ]),
    ghostVotesRemaining: new Set([p('deadguy')]),
  };

  it('consumes the ghost vote', () => {
    const next = deathReducer(stateWithDead, {
      type: 'GHOST_VOTE_CAST',
      playerId: p('deadguy'),
    });

    expect(next.ghostVotesRemaining.has(p('deadguy'))).toBe(false);
    // Death record is preserved
    expect(next.deaths.has(p('deadguy'))).toBe(true);
  });

  it('rejects ghost vote from alive player', () => {
    const next = deathReducer(stateWithDead, {
      type: 'GHOST_VOTE_CAST',
      playerId: p('alice'),
    });

    expect(next).toBe(stateWithDead);
  });

  it('rejects ghost vote from dead player with spent ghost vote', () => {
    const spent = deathReducer(stateWithDead, {
      type: 'GHOST_VOTE_CAST',
      playerId: p('deadguy'),
    });

    // Try to vote again
    const next = deathReducer(spent, {
      type: 'GHOST_VOTE_CAST',
      playerId: p('deadguy'),
    });

    expect(next).toBe(spent);
    expect(next.ghostVotesRemaining.has(p('deadguy'))).toBe(false);
  });
});

// ─── deathReducer: default case ──────────────────────────────────

describe('deathReducer: unknown event', () => {
  it('returns state unchanged for unknown event type', () => {
    // @ts-expect-error -- testing unknown event type
    const next = deathReducer(INITIAL_DEATH_STATE, { type: 'UNKNOWN' });
    expect(next).toBe(INITIAL_DEATH_STATE);
  });
});

// ─── computeDeathTriggers ────────────────────────────────────────

describe('computeDeathTriggers', () => {
  const executionRecord: DeathRecord = {
    playerId: p('alice'),
    cause: 'execution',
    dayNumber: 1,
    killedBy: null,
  };

  const nightKillRecord: DeathRecord = {
    playerId: p('bob'),
    cause: 'night_kill',
    dayNumber: 2,
    killedBy: p('imp'),
  };

  it('triggers saint_execution when a Saint is executed', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('saint'),
      alivePlayerCountAfterDeath: 6,
    };

    const triggers = computeDeathTriggers(executionRecord, ctx);
    // Saint execution triggers both saint_execution and undertaker (since it's an execution)
    expect(triggers).toHaveLength(2);
    const saintTrigger = triggers.find((t) => t.type === 'saint_execution');
    expect(saintTrigger).toBeDefined();
    expect(saintTrigger!.sourcePlayerId).toBe(p('alice'));
  });

  it('does NOT trigger saint_execution for non-Saint execution', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('chef'),
      alivePlayerCountAfterDeath: 6,
    };

    const triggers = computeDeathTriggers(executionRecord, ctx);
    expect(triggers.find((t) => t.type === 'saint_execution')).toBeUndefined();
  });

  it('does NOT trigger saint_execution for night kill (only execution)', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('saint'),
      alivePlayerCountAfterDeath: 6,
    };

    const triggers = computeDeathTriggers(nightKillRecord, ctx);
    expect(triggers.find((t) => t.type === 'saint_execution')).toBeUndefined();
  });

  it('triggers scarlet_woman when Imp dies with 5+ alive', () => {
    const characters = new Map<PlayerId, Character>([
      [p('imp_player'), makeCharacter('imp', 'evil')],
      [p('scarlet'), makeCharacter('scarlet_woman', 'evil')],
    ]);

    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('imp', 'evil'),
      alivePlayerCountAfterDeath: 5,
      characters,
    };

    const impRecord: DeathRecord = {
      playerId: p('imp_player'),
      cause: 'execution',
      dayNumber: 3,
      killedBy: null,
    };

    const triggers = computeDeathTriggers(impRecord, ctx);
    const sw = triggers.find((t) => t.type === 'scarlet_woman');
    expect(sw).toBeDefined();
    expect(sw!.sourcePlayerId).toBe(p('imp_player'));
    expect(sw!.targetPlayerId).toBe(p('scarlet'));
  });

  it('does NOT trigger scarlet_woman when Imp dies with fewer than 5 alive', () => {
    const characters = new Map<PlayerId, Character>([
      [p('imp_player'), makeCharacter('imp', 'evil')],
      [p('scarlet'), makeCharacter('scarlet_woman', 'evil')],
    ]);

    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('imp', 'evil'),
      alivePlayerCountAfterDeath: 4,
      characters,
    };

    const triggers = computeDeathTriggers(executionRecord, ctx);
    expect(triggers.find((t) => t.type === 'scarlet_woman')).toBeUndefined();
  });

  it('does NOT trigger scarlet_woman when no Scarlet Woman in game', () => {
    const characters = new Map<PlayerId, Character>([
      [p('imp_player'), makeCharacter('imp', 'evil')],
    ]);

    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('imp', 'evil'),
      alivePlayerCountAfterDeath: 5,
      characters,
    };

    const triggers = computeDeathTriggers(executionRecord, ctx);
    expect(triggers.find((t) => t.type === 'scarlet_woman')).toBeUndefined();
  });

  it('triggers ravenkeeper when Ravenkeeper dies at night', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('ravenkeeper'),
      alivePlayerCountAfterDeath: 5,
    };

    const triggers = computeDeathTriggers(nightKillRecord, ctx);
    expect(triggers).toHaveLength(1);
    expect(triggers[0]!.type).toBe('ravenkeeper');
    expect(triggers[0]!.sourcePlayerId).toBe(p('bob'));
  });

  it('does NOT trigger ravenkeeper when Ravenkeeper is executed (not night kill)', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('ravenkeeper'),
      alivePlayerCountAfterDeath: 5,
    };

    const triggers = computeDeathTriggers(executionRecord, ctx);
    expect(triggers.find((t) => t.type === 'ravenkeeper')).toBeUndefined();
  });

  it('triggers undertaker when a player is executed', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('chef'),
      alivePlayerCountAfterDeath: 5,
    };

    const triggers = computeDeathTriggers(executionRecord, ctx);
    const ut = triggers.find((t) => t.type === 'undertaker');
    expect(ut).toBeDefined();
    expect(ut!.sourcePlayerId).toBe(p('alice'));
    expect(ut!.deadPlayerCharacter?.id).toBe('chef');
  });

  it('does NOT trigger undertaker for night kills', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('chef'),
      alivePlayerCountAfterDeath: 5,
    };

    const triggers = computeDeathTriggers(nightKillRecord, ctx);
    expect(triggers.find((t) => t.type === 'undertaker')).toBeUndefined();
  });

  it('triggers both saint_execution and undertaker for Saint execution', () => {
    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('saint'),
      alivePlayerCountAfterDeath: 5,
    };

    const triggers = computeDeathTriggers(executionRecord, ctx);
    expect(triggers).toHaveLength(2);
    expect(triggers.map((t) => t.type)).toContain('saint_execution');
    expect(triggers.map((t) => t.type)).toContain('undertaker');
  });

  it('returns empty array for ability death with no special character', () => {
    const abilityRecord: DeathRecord = {
      playerId: p('charlie'),
      cause: 'ability',
      dayNumber: 2,
      killedBy: null,
    };

    const ctx: DeathTriggerContext = {
      deadPlayerCharacter: makeCharacter('chef'),
      alivePlayerCountAfterDeath: 5,
    };

    const triggers = computeDeathTriggers(abilityRecord, ctx);
    expect(triggers).toHaveLength(0);
  });
});

// ─── Helper functions ────────────────────────────────────────────

describe('isDead', () => {
  it('returns false for alive player', () => {
    expect(isDead(INITIAL_DEATH_STATE, p('alice'))).toBe(false);
  });

  it('returns true for dead player', () => {
    const state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    expect(isDead(state, p('alice'))).toBe(true);
  });
});

describe('getDeathRecord', () => {
  it('returns undefined for alive player', () => {
    expect(getDeathRecord(INITIAL_DEATH_STATE, p('alice'))).toBeUndefined();
  });

  it('returns death record for dead player', () => {
    const state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    const record = getDeathRecord(state, p('alice'));
    expect(record).toBeDefined();
    expect(record!.playerId).toBe(p('alice'));
    expect(record!.cause).toBe('execution');
    expect(record!.dayNumber).toBe(1);
  });
});

describe('hasGhostVote', () => {
  it('returns false for alive player', () => {
    expect(hasGhostVote(INITIAL_DEATH_STATE, p('alice'))).toBe(false);
  });

  it('returns true for dead player with unused ghost vote', () => {
    const state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    expect(hasGhostVote(state, p('alice'))).toBe(true);
  });

  it('returns false for dead player with spent ghost vote', () => {
    let state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    state = deathReducer(state, {
      type: 'GHOST_VOTE_CAST',
      playerId: p('alice'),
    });
    expect(hasGhostVote(state, p('alice'))).toBe(false);
  });
});

describe('getDeathCount', () => {
  it('returns 0 for no deaths', () => {
    expect(getDeathCount(INITIAL_DEATH_STATE)).toBe(0);
  });

  it('returns correct count after multiple deaths', () => {
    let state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('bob'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    });
    expect(getDeathCount(state)).toBe(2);
  });
});

describe('getDeathsByCause', () => {
  it('returns empty array when no deaths match cause', () => {
    expect(getDeathsByCause(INITIAL_DEATH_STATE, 'execution')).toHaveLength(0);
  });

  it('filters deaths by cause', () => {
    let state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('bob'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    });
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('charlie'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    });

    expect(getDeathsByCause(state, 'execution')).toHaveLength(1);
    expect(getDeathsByCause(state, 'night_kill')).toHaveLength(2);
    expect(getDeathsByCause(state, 'ability')).toHaveLength(0);
  });
});

describe('getDeathsOnDay', () => {
  it('returns empty array for day with no deaths', () => {
    expect(getDeathsOnDay(INITIAL_DEATH_STATE, 1)).toHaveLength(0);
  });

  it('returns deaths for a specific day', () => {
    let state = deathReducer(INITIAL_DEATH_STATE, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('bob'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    });
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('charlie'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    });

    expect(getDeathsOnDay(state, 1)).toHaveLength(1);
    expect(getDeathsOnDay(state, 2)).toHaveLength(2);
    expect(getDeathsOnDay(state, 3)).toHaveLength(0);
  });
});

describe('createDeathAnnouncement', () => {
  it('creates a broadcast-ready announcement from a death record', () => {
    const record: DeathRecord = {
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 3,
      killedBy: null,
    };

    const announcement = createDeathAnnouncement(record);
    expect(announcement).toEqual({
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 3,
    });
    // Announcement should NOT include killedBy (private info)
    expect(announcement).not.toHaveProperty('killedBy');
  });
});

// ─── Full Game Integration ───────────────────────────────────────

describe('full game scenario', () => {
  it('simulates a multi-day game with executions, night kills, and ghost votes', () => {
    let state: DeathState = { ...INITIAL_DEATH_STATE };

    // Day 1: Alice is executed
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('alice'),
      cause: 'execution',
      dayNumber: 1,
    });
    expect(getDeathCount(state)).toBe(1);
    expect(hasGhostVote(state, p('alice'))).toBe(true);

    // Night 1: Bob is killed by the Imp
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('bob'),
      cause: 'night_kill',
      dayNumber: 1,
      killedBy: p('imp'),
    });
    expect(getDeathCount(state)).toBe(2);
    expect(getDeathsOnDay(state, 1)).toHaveLength(2);

    // Day 2: Charlie is executed
    state = deathReducer(state, {
      type: 'PLAYER_DIED',
      playerId: p('charlie'),
      cause: 'execution',
      dayNumber: 2,
    });
    expect(getDeathCount(state)).toBe(3);

    // Alice uses her ghost vote on a nomination
    state = deathReducer(state, {
      type: 'GHOST_VOTE_CAST',
      playerId: p('alice'),
    });
    expect(hasGhostVote(state, p('alice'))).toBe(false);

    // Bob still has his ghost vote
    expect(hasGhostVote(state, p('bob'))).toBe(true);

    // Alice tries to use ghost vote again - rejected
    state = deathReducer(state, {
      type: 'GHOST_VOTE_CAST',
      playerId: p('alice'),
    });
    expect(hasGhostVote(state, p('alice'))).toBe(false); // still no vote

    // Check death causes
    expect(getDeathsByCause(state, 'execution')).toHaveLength(2);
    expect(getDeathsByCause(state, 'night_kill')).toHaveLength(1);
  });

  it('handles Saint execution trigger correctly', () => {
    // Saint is executed on day 2
    const record: DeathRecord = {
      playerId: p('saint_player'),
      cause: 'execution',
      dayNumber: 2,
      killedBy: null,
    };

    const triggers = computeDeathTriggers(record, {
      deadPlayerCharacter: makeCharacter('saint'),
      alivePlayerCountAfterDeath: 5,
    });

    // Saint execution should trigger evil win
    expect(triggers.some((t) => t.type === 'saint_execution')).toBe(true);
    // Undertaker should also trigger
    expect(triggers.some((t) => t.type === 'undertaker')).toBe(true);
  });

  it('handles Scarlet Woman promotion scenario', () => {
    const characters = new Map<PlayerId, Character>([
      [p('imp_p'), makeCharacter('imp', 'evil')],
      [p('sw_p'), makeCharacter('scarlet_woman', 'evil')],
      [p('alice'), makeCharacter('chef', 'good')],
      [p('bob'), makeCharacter('empath', 'good')],
      [p('charlie'), makeCharacter('investigator', 'good')],
      [p('dave'), makeCharacter('monk', 'good')],
    ]);

    const record: DeathRecord = {
      playerId: p('imp_p'),
      cause: 'execution',
      dayNumber: 3,
      killedBy: null,
    };

    const triggers = computeDeathTriggers(record, {
      deadPlayerCharacter: makeCharacter('imp', 'evil'),
      alivePlayerCountAfterDeath: 5,
      characters,
    });

    const sw = triggers.find((t) => t.type === 'scarlet_woman');
    expect(sw).toBeDefined();
    expect(sw!.targetPlayerId).toBe(p('sw_p'));
  });

  it('handles Ravenkeeper night death scenario', () => {
    const record: DeathRecord = {
      playerId: p('raven'),
      cause: 'night_kill',
      dayNumber: 2,
      killedBy: p('imp'),
    };

    const triggers = computeDeathTriggers(record, {
      deadPlayerCharacter: makeCharacter('ravenkeeper'),
      alivePlayerCountAfterDeath: 4,
    });

    expect(triggers).toHaveLength(1);
    expect(triggers[0]!.type).toBe('ravenkeeper');
    expect(triggers[0]!.sourcePlayerId).toBe(p('raven'));
  });
});
