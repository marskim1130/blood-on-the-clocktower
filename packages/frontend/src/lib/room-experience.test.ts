import { describe, expect, it } from 'vitest';
import type { RoomState, ServerMessage } from '@clocktower/core';
import {
  applyServerMessage,
  createRoomExperienceState,
  selectDeathAnnouncements,
  selectDeathRecords,
  selectActingPlayerId,
  selectGameOver,
  selectGamePhase,
  selectGhostVotes,
} from './room-experience';

function roomState(overrides: Partial<RoomState> = {}): RoomState {
  return {
    roomId: 'room-1',
    players: [{ id: 'p1', name: 'Alice', isAlive: true, votes: 0, isReady: false, hasConfirmedCharacter: false }],
    maxPlayers: 5,
    scriptId: 'trouble_brewing',
    scriptName: 'Trouble Brewing',
    phase: 1,
    dayNumber: 0,
    nightNumber: 0,
    ...overrides,
  };
}

function apply(message: ServerMessage) {
  return applyServerMessage(createRoomExperienceState(), message);
}

describe('room experience projection', () => {
  it('derives the complete view from one authoritative projection', () => {
    const state = apply({
      type: 'ROOM_STATE',
      roomRevision: 7,
      state: roomState({
        phase: 5,
        dayNumber: 3,
        players: [{ id: 'p1', name: 'Alice', isAlive: false, votes: 0, isReady: false, hasConfirmedCharacter: false }],
        deaths: [{ playerId: 'p1', cause: 'execution', dayNumber: 3 }],
        ghostVotesRemaining: ['p1'],
        winner: { winner: 1, reason: 'imp_executed', description: 'Good wins' },
      }),
    });

    expect(state.roomRevision).toBe(7);
    expect(state.identityStatus?.status).toBe('member');
    expect(selectGamePhase(state)).toBe('finished');
    expect(selectDeathRecords(state)).toEqual({ p1: { cause: 'execution', dayNumber: 3 } });
    expect(selectGhostVotes(state).has('p1')).toBe(true);
    expect(selectGameOver(state)?.winner).toBe('good');
    expect(selectDeathAnnouncements(state)).toEqual(['第 3 天：Alice 死亡']);
  });

  it('replaces optional state instead of retaining stale UI facts', () => {
    const stale = apply({
      type: 'ROOM_STATE',
      roomRevision: 2,
      state: roomState({
        nomination: {
          nominatorId: 'p1',
          nomineeId: 'p2',
          votes: {},
          resolved: false,
          voterOrder: ['p2', 'p1'],
          currentVoterIndex: 0,
        },
        deaths: [{ playerId: 'p1', cause: 'ability', dayNumber: 1 }],
        winner: { winner: 2, reason: 'evil_majority', description: 'Evil wins' },
      }),
    });
    const replaced = applyServerMessage(stale, {
      type: 'ROOM_STATE',
      roomRevision: 3,
      state: roomState(),
    });

    expect(replaced.roomState?.nomination).toBeUndefined();
    expect(selectDeathRecords(replaced)).toEqual({});
    expect(selectGameOver(replaced)).toBeNull();
  });

  it('preserves the authoritative participant freeze status', () => {
    const state = apply({
      type: 'COMMAND_RESULT',
      state: roomState(),
      identityStatus: {
        status: 'member',
        canRejoin: false,
        participantSetFrozen: true,
      },
    });

    expect(state.identityStatus?.participantSetFrozen).toBe(true);
  });

  it('clears the projection for retained and terminal identities', () => {
    const member = apply({ type: 'ROOM_STATE', roomRevision: 2, state: roomState() });
    const retained = applyServerMessage(member, {
      type: 'COMMAND_RESULT',
      roomRevision: 3,
      identityStatus: { status: 'retained', canRejoin: true, participantSetFrozen: false },
    });
    expect(retained.roomState).toBeNull();
    expect(retained.identityStatus?.status).toBe('retained');

    const kicked = applyServerMessage(member, { type: 'KICKED', roomId: 'room-1' });
    expect(kicked).toEqual({ ...createRoomExperienceState(), terminalReason: 'kicked' });

    const missing = applyServerMessage(member, { type: 'ERROR', code: 'ROOM_NOT_FOUND', error: 'room not found' });
    expect(missing).toEqual({ ...createRoomExperienceState(), terminalReason: 'closed' });
  });

  it('matches night actors by actual or shown character', () => {
    const players = [
      { id: 'drunk', character: { id: 'drunk' }, shownCharacter: { id: 'monk' } },
      { id: 'butler', character: { id: 'butler' } },
    ];

    expect(selectActingPlayerId(players, 'monk')).toBe('drunk');
    expect(selectActingPlayerId(players, 'butler')).toBe('butler');
    expect(selectActingPlayerId(players, 'imp')).toBeUndefined();
  });
});
