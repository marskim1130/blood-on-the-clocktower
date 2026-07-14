import type { RoomExperienceState } from './room-experience';

export const SESSION_ROUTES = {
  lobby: '/pages/index/index',
  setup: '/pages/game-setup/index',
  play: '/pages/game-play/index',
  over: '/pages/game-over/index',
} as const;

export type SessionRoute = (typeof SESSION_ROUTES)[keyof typeof SESSION_ROUTES];

export function routeForExperience(
  experience: RoomExperienceState,
  hasIdentity: boolean,
  initialized: boolean,
): SessionRoute | null {
  if (!initialized) return null;
  if (experience.identityStatus?.status === 'retained') return SESSION_ROUTES.lobby;
  const room = experience.roomState;
  if (!room) return hasIdentity ? null : SESSION_ROUTES.lobby;
  if (room.phase === 5) return SESSION_ROUTES.over;
  if (room.phase === 1 || room.phase === 0) return SESSION_ROUTES.setup;
  return SESSION_ROUTES.play;
}
