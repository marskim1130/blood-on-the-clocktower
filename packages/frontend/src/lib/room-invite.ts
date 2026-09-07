import { SESSION_ROUTES } from './session-routing';

export function getBrowserOrigin(): string | undefined {
  const runtime = globalThis as typeof globalThis & { location?: { origin?: string } };
  const origin = runtime.location?.origin;
  return origin && /^https?:\/\//.test(origin) ? origin : undefined;
}

export function createRoomInvite(roomId: string, origin?: string): string {
  const path = `${SESSION_ROUTES.lobby}?roomId=${encodeURIComponent(roomId)}`;
  return origin ? `${origin.replace(/\/$/, '')}/#${path}` : path;
}
