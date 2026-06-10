import { describe, it, expect } from 'vitest';
import {
  checkWinAfterExecution,
  checkWinAfterNightDeath,
  checkWinAtEndOfDay,
  checkWinCondition,
  createGameOverEvent,
  getAliveCount,
  getAlivePlayers,
  getDeadPlayers,
  findPlayerByCharacterId,
  isCharacterAlive,
  isTeamAlive,
  getPlayersByTeam,
  INITIAL_WIN_CHECK_CONTEXT,
  type WinCheckContext,
  type WinCheckPlayer,
  type WinResult,
} from '../index.js';
import { createPlayerId, type PlayerId, type Character } from '../../types/index.js';

// ─── Helpers ─────────────────────────────────────────────────────

function p(id: string): PlayerId {
  return createPlayerId(id);
}

function makeCharacter(id: string, team: 'good' | 'evil' = 'good', name?: string): Character {
  return { id, name: name ?? id, team, ability: `${id}-ability` };
}

function makePlayer(
  id: string,
  character: Character | null = null,
  isAlive: boolean = true,
): WinCheckPlayer {
  return { id: p(id), character, isAlive };
}

function makeContext(overrides: Partial<WinCheckContext> = {}): WinCheckContext {
  return { ...INITIAL_WIN_CHECK_CONTEXT, ...overrides };
}

// ─── Validation ──────────────────────────────────────────────────

describe('validateWinCheckContext', () => {
  it('rejects empty players', () => {
    const ctx = makeContext({ players: [], dayNumber: 1 });
    const result = checkWinCondition(ctx);
    expect(result).toBeNull();
  });

  it('rejects day 0', () => {
    const ctx = makeContext({
      players: [makePlayer('alice')],
      dayNumber: 0,
    });
    const result = checkWinCondition(ctx);
    expect(result).toBeNull();
  });

  it('rejects negative day', () => {
    const ctx = makeContext({
      players: [makePlayer('alice')],
      dayNumber: -1,
    });
    const result = checkWinCondition(ctx);
    expect(result).toBeNull();
  });
});

// ─── Good Win Conditions ─────────────────────────────────────────

describe('good win conditions', () => {
  it('good wins when Imp is executed (no Scarlet Woman)', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp_player', makeCharacter('imp', 'evil')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
      ],
      dayNumber: 2,
      executedToday: true,
      executedPlayerId: p('imp_player'),
      scarletWomanTriggered: false,
    });

    const result = checkWinAfterExecution(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('good');
    expect(result!.reason).toBe('imp_executed');
  });

  it('good does NOT win when Imp is executed but Scarlet Woman triggered', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp_player', makeCharacter('imp', 'evil')),
        makePlayer('scarlet', makeCharacter('scarlet_woman', 'evil')),
        makePlayer('alice', makeCharacter('chef', 'good')),
      ],
      dayNumber: 2,
      executedToday: true,
      executedPlayerId: p('imp_player'),
      scarletWomanTriggered: true,
    });

    const result = checkWinAfterExecution(ctx);
    expect(result).toBeNull();
  });

  it('good wins with Mayor endgame (3 alive, no execution)', () => {
    const ctx = makeContext({
      players: [
        makePlayer('mayor', makeCharacter('mayor', 'good')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('imp', makeCharacter('imp', 'evil')),
      ],
      dayNumber: 3,
      executedToday: false,
    });

    const result = checkWinAtEndOfDay(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('good');
    expect(result!.reason).toBe('mayor_endgame');
  });

  it('good does NOT win with Mayor endgame if no Mayor in game', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
        makePlayer('imp', makeCharacter('imp', 'evil')),
      ],
      dayNumber: 3,
      executedToday: false,
    });

    const result = checkWinAtEndOfDay(ctx);
    expect(result).toBeNull();
  });

  it('good does NOT win with Mayor endgame if Mayor is dead', () => {
    const ctx = makeContext({
      players: [
        makePlayer('mayor', makeCharacter('mayor', 'good'), false),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
        makePlayer('imp', makeCharacter('imp', 'evil')),
      ],
      dayNumber: 3,
      executedToday: false,
    });

    const result = checkWinAtEndOfDay(ctx);
    expect(result).toBeNull();
  });

  it('good does NOT win with Mayor endgame if execution occurred', () => {
    const ctx = makeContext({
      players: [
        makePlayer('mayor', makeCharacter('mayor', 'good')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('imp', makeCharacter('imp', 'evil')),
      ],
      dayNumber: 3,
      executedToday: true,
      executedPlayerId: p('alice'),
    });

    const result = checkWinAtEndOfDay(ctx);
    expect(result).toBeNull();
  });
});

// ─── Evil Win Conditions ─────────────────────────────────────────

describe('evil win conditions', () => {
  it('evil wins when only 2 players alive', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp', makeCharacter('imp', 'evil')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good'), false),
      ],
      dayNumber: 3,
      executedToday: false,
    });

    const result = checkWinAtEndOfDay(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('evil');
    expect(result!.reason).toBe('evil_majority');
  });

  it('evil wins when Saint is executed', () => {
    const ctx = makeContext({
      players: [
        makePlayer('saint', makeCharacter('saint', 'good')),
        makePlayer('imp', makeCharacter('imp', 'evil')),
        makePlayer('alice', makeCharacter('chef', 'good')),
      ],
      dayNumber: 2,
      executedToday: true,
      executedPlayerId: p('saint'),
    });

    const result = checkWinAfterExecution(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('evil');
    expect(result!.reason).toBe('saint_executed');
  });

  it('evil wins when Imp kills themselves at night (no Scarlet Woman)', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp', makeCharacter('imp', 'evil'), false),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
        makePlayer('charlie', makeCharacter('monk', 'good')),
      ],
      dayNumber: 2,
      lastDeathCause: 'night_kill',
      lastDeadPlayerId: p('imp'),
      scarletWomanTriggered: false,
    });

    const result = checkWinAfterNightDeath(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('evil');
    expect(result!.reason).toBe('imp_starpass');
  });

  it('evil does NOT win when Imp kills themselves but Scarlet Woman triggered', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp', makeCharacter('imp', 'evil'), false),
        makePlayer('scarlet', makeCharacter('scarlet_woman', 'evil')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
      ],
      dayNumber: 2,
      lastDeathCause: 'night_kill',
      lastDeadPlayerId: p('imp'),
      scarletWomanTriggered: true,
    });

    const result = checkWinAfterNightDeath(ctx);
    expect(result).toBeNull();
  });
});

// ─── Unified Win Condition Check ─────────────────────────────────

describe('checkWinCondition', () => {
  it('returns null when no win condition met', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp', makeCharacter('imp', 'evil')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
        makePlayer('charlie', makeCharacter('monk', 'good')),
      ],
      dayNumber: 1,
      executedToday: false,
    });

    const result = checkWinCondition(ctx);
    expect(result).toBeNull();
  });

  it('checks execution win conditions first', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp', makeCharacter('imp', 'evil')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
      ],
      dayNumber: 2,
      executedToday: true,
      executedPlayerId: p('imp'),
      scarletWomanTriggered: false,
    });

    const result = checkWinCondition(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('good');
    expect(result!.reason).toBe('imp_executed');
  });

  it('checks night death win conditions', () => {
    const ctx = makeContext({
      players: [
        makePlayer('imp', makeCharacter('imp', 'evil'), false),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('empath', 'good')),
        makePlayer('charlie', makeCharacter('monk', 'good')),
      ],
      dayNumber: 2,
      lastDeathCause: 'night_kill',
      lastDeadPlayerId: p('imp'),
      scarletWomanTriggered: false,
    });

    const result = checkWinCondition(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('evil');
    expect(result!.reason).toBe('imp_starpass');
  });

  it('checks end of day win conditions', () => {
    const ctx = makeContext({
      players: [
        makePlayer('mayor', makeCharacter('mayor', 'good')),
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('imp', makeCharacter('imp', 'evil')),
      ],
      dayNumber: 3,
      executedToday: false,
    });

    const result = checkWinCondition(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('good');
    expect(result!.reason).toBe('mayor_endgame');
  });
});

// ─── Helper Functions ────────────────────────────────────────────

describe('getAliveCount', () => {
  it('counts alive players', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', null, true),
        makePlayer('bob', null, true),
        makePlayer('charlie', null, false),
      ],
    });

    expect(getAliveCount(ctx)).toBe(2);
  });

  it('returns 0 for all dead', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', null, false),
        makePlayer('bob', null, false),
      ],
    });

    expect(getAliveCount(ctx)).toBe(0);
  });
});

describe('getAlivePlayers', () => {
  it('returns only alive players', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', null, true),
        makePlayer('bob', null, false),
        makePlayer('charlie', null, true),
      ],
    });

    const alive = getAlivePlayers(ctx);
    expect(alive).toHaveLength(2);
    expect(alive.map((p) => p.id)).toContain(p('alice'));
    expect(alive.map((p) => p.id)).toContain(p('charlie'));
  });
});

describe('getDeadPlayers', () => {
  it('returns only dead players', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', null, true),
        makePlayer('bob', null, false),
        makePlayer('charlie', null, true),
      ],
    });

    const dead = getDeadPlayers(ctx);
    expect(dead).toHaveLength(1);
    expect(dead[0]!.id).toBe(p('bob'));
  });
});

describe('findPlayerByCharacterId', () => {
  it('finds player by character ID', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good')),
        makePlayer('bob', makeCharacter('imp', 'evil')),
      ],
    });

    const player = findPlayerByCharacterId(ctx, 'imp');
    expect(player).toBeDefined();
    expect(player!.id).toBe(p('bob'));
  });

  it('returns undefined if not found', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good')),
      ],
    });

    const player = findPlayerByCharacterId(ctx, 'imp');
    expect(player).toBeUndefined();
  });
});

describe('isCharacterAlive', () => {
  it('returns true if character is alive', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good'), true),
        makePlayer('bob', makeCharacter('imp', 'evil'), false),
      ],
    });

    expect(isCharacterAlive(ctx, 'chef')).toBe(true);
  });

  it('returns false if character is dead', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good'), true),
        makePlayer('bob', makeCharacter('imp', 'evil'), false),
      ],
    });

    expect(isCharacterAlive(ctx, 'imp')).toBe(false);
  });

  it('returns false if character not in game', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good'), true),
      ],
    });

    expect(isCharacterAlive(ctx, 'imp')).toBe(false);
  });
});

describe('isTeamAlive', () => {
  it('returns true if team has living members', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good'), true),
        makePlayer('bob', makeCharacter('imp', 'evil'), false),
      ],
    });

    expect(isTeamAlive(ctx, 'good')).toBe(true);
    expect(isTeamAlive(ctx, 'evil')).toBe(false);
  });
});

describe('getPlayersByTeam', () => {
  it('returns all players on team (dead or alive)', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good'), true),
        makePlayer('bob', makeCharacter('empath', 'good'), false),
        makePlayer('charlie', makeCharacter('imp', 'evil'), true),
      ],
    });

    const goodPlayers = getPlayersByTeam(ctx, 'good');
    expect(goodPlayers).toHaveLength(2);
    expect(goodPlayers.map((p) => p.id)).toContain(p('alice'));
    expect(goodPlayers.map((p) => p.id)).toContain(p('bob'));
  });
});

// ─── Game Over Event ─────────────────────────────────────────────

describe('createGameOverEvent', () => {
  it('creates game over event with revealed players', () => {
    const result: WinResult = {
      winner: 'good',
      reason: 'imp_executed',
      description: 'The Imp was executed. Good wins!',
    };

    const players = [
      makePlayer('alice', makeCharacter('chef', 'good')),
      makePlayer('bob', makeCharacter('imp', 'evil')),
    ];

    const event = createGameOverEvent(result, players);

    expect(event.type).toBe('GAME_OVER');
    expect(event.winner).toBe('good');
    expect(event.reason).toBe('imp_executed');
    expect(event.description).toBe('The Imp was executed. Good wins!');
    expect(event.revealedPlayers).toHaveLength(2);
    expect(event.revealedPlayers[0]!.character?.id).toBe('chef');
    expect(event.revealedPlayers[1]!.character?.id).toBe('imp');
  });
});

// ─── Edge Cases ──────────────────────────────────────────────────

describe('edge cases', () => {
  it('handles execution that leaves only 2 alive', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good'), false),
        makePlayer('bob', makeCharacter('empath', 'good'), true),
        makePlayer('imp', makeCharacter('imp', 'evil'), true),
      ],
      dayNumber: 3,
      executedToday: true,
      executedPlayerId: p('alice'),
      scarletWomanTriggered: false,
    });

    // Should detect evil majority after execution
    const result = checkWinCondition(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('evil');
    expect(result!.reason).toBe('evil_majority');
  });

  it('handles night death that leaves only 2 alive', () => {
    const ctx = makeContext({
      players: [
        makePlayer('alice', makeCharacter('chef', 'good'), false),
        makePlayer('bob', makeCharacter('empath', 'good'), true),
        makePlayer('imp', makeCharacter('imp', 'evil'), true),
      ],
      dayNumber: 2,
      lastDeathCause: 'night_kill',
      lastDeadPlayerId: p('alice'),
      scarletWomanTriggered: false,
    });

    const result = checkWinAfterNightDeath(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('evil');
    expect(result!.reason).toBe('evil_majority');
  });

  it('handles multiple win conditions (execution + majority)', () => {
    const ctx = makeContext({
      players: [
        makePlayer('saint', makeCharacter('saint', 'good'), false),
        makePlayer('imp', makeCharacter('imp', 'evil'), true),
        makePlayer('alice', makeCharacter('chef', 'good'), true),
      ],
      dayNumber: 2,
      executedToday: true,
      executedPlayerId: p('saint'),
      scarletWomanTriggered: false,
    });

    // Saint execution should trigger first
    const result = checkWinCondition(ctx);
    expect(result).not.toBeNull();
    expect(result!.winner).toBe('evil');
    expect(result!.reason).toBe('saint_executed');
  });
});
