import { describe, expect, it } from 'vitest';
import { createRoomExperienceState, type RoomExperienceState } from './room-experience';
import { routeForExperience, SESSION_ROUTES } from './session-routing';

function experience(phase: number): RoomExperienceState {
  return {
    ...createRoomExperienceState(),
    roomState: {
      roomId: 'ROOM1',
      players: [],
      maxPlayers: 5,
      scriptId: 'trouble_brewing',
      scriptName: '暗流涌动',
      phase,
      dayNumber: 0,
      nightNumber: 0,
    },
  };
}

describe('session routing', () => {
  it('waits for initialization and identity recovery', () => {
    expect(routeForExperience(createRoomExperienceState(), false, false)).toBeNull();
    expect(routeForExperience(createRoomExperienceState(), true, true)).toBeNull();
  });

  it('routes authoritative phases to their owning pages', () => {
    expect(routeForExperience(experience(1), true, true)).toBe(SESSION_ROUTES.setup);
    expect(routeForExperience(experience(2), true, true)).toBe(SESSION_ROUTES.play);
    expect(routeForExperience(experience(3), true, true)).toBe(SESSION_ROUTES.play);
    expect(routeForExperience(experience(4), true, true)).toBe(SESSION_ROUTES.play);
    expect(routeForExperience(experience(5), true, true)).toBe(SESSION_ROUTES.over);
  });

  it('returns to lobby only after identity is gone', () => {
    expect(routeForExperience(createRoomExperienceState(), false, true)).toBe(SESSION_ROUTES.lobby);
  });

  it('returns retained identities to the lobby with or without rejoin permission', () => {
    for (const canRejoin of [true, false]) {
      expect(routeForExperience({
        ...createRoomExperienceState(),
        identityStatus: {
          status: 'retained',
          canRejoin,
          participantSetFrozen: !canRejoin,
        },
      }, true, true)).toBe(SESSION_ROUTES.lobby);
    }
  });
});
