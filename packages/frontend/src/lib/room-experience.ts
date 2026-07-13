import type { GameCharacter, IdentityStatus, RoomState, ServerMessage } from '@clocktower/core';

export type GamePhase = 'setup' | 'day' | 'voting' | 'night' | 'finished';

export interface RoomExperienceState {
  readonly roomState: RoomState | null;
  readonly identityStatus: IdentityStatus | null;
  readonly roomRevision: number | null;
  readonly terminalReason: 'kicked' | 'closed' | 'invalid-credential' | null;
}

export interface DeathRecordView {
  readonly cause: string;
  readonly dayNumber: number;
}

export interface GameOverView {
  readonly winner: string;
  readonly reason: string;
  readonly description: string;
}

export function createRoomExperienceState(): RoomExperienceState {
  return {
    roomState: null,
    identityStatus: null,
    roomRevision: null,
    terminalReason: null,
  };
}

export function applyServerMessage(
  current: RoomExperienceState,
  message: ServerMessage,
): RoomExperienceState {
  if (message.type === 'KICKED') {
    return { ...createRoomExperienceState(), terminalReason: 'kicked' };
  }
  if (message.type === 'ROOM_CLOSED') {
    return { ...createRoomExperienceState(), terminalReason: 'closed' };
  }
  if (message.type === 'ERROR' && message.code === 'INVALID_CREDENTIAL') {
    return { ...createRoomExperienceState(), terminalReason: 'invalid-credential' };
  }

  const identityStatus = message.identityStatus ?? current.identityStatus;
  const retained = identityStatus?.status === 'retained';
  const memberResult = message.type === 'CREATE_ROOM_RESULT' || message.type === 'JOIN_ROOM_RESULT';
  const memberProjection = Boolean(message.state);

  return {
    roomState: retained ? null : (message.state ?? current.roomState),
    identityStatus: memberResult || memberProjection
      ? { status: 'member', canRejoin: false, participantSetFrozen: false }
      : identityStatus,
    roomRevision: message.roomRevision ?? current.roomRevision,
    terminalReason: null,
  };
}

export function clearRoomProjection(current: RoomExperienceState): RoomExperienceState {
  return { ...current, roomState: null, roomRevision: null, terminalReason: null };
}

export function clearRoomIdentityState(): RoomExperienceState {
  return createRoomExperienceState();
}

export function mapProtocolPhase(phase: number | undefined): GamePhase | null {
  if (phase === undefined) return null;
  const phases: Readonly<Record<number, GamePhase>> = {
    0: 'setup',
    1: 'setup',
    2: 'day',
    3: 'night',
    4: 'voting',
    5: 'finished',
  };
  return phases[phase] ?? null;
}

export function normalizeWinner(winner: number | string): string {
  if (winner === 1 || winner === 'good') return 'good';
  if (winner === 2 || winner === 'evil') return 'evil';
  return String(winner);
}

export function selectGamePhase(state: RoomExperienceState): GamePhase {
  return mapProtocolPhase(state.roomState?.phase) ?? 'setup';
}

export function selectVisibleCharacter(
  state: RoomExperienceState,
  playerId: string,
): GameCharacter | null {
  return state.roomState?.players.find((player) => player.id === playerId)?.character ?? null;
}

export function selectDeathRecords(
  state: RoomExperienceState,
): Readonly<Record<string, DeathRecordView>> {
  return Object.fromEntries(
    (state.roomState?.deaths ?? []).map((death) => [
      death.playerId,
      { cause: death.cause, dayNumber: death.dayNumber },
    ]),
  );
}

export function selectGhostVotes(state: RoomExperienceState): ReadonlySet<string> {
  return new Set(state.roomState?.ghostVotesRemaining ?? []);
}

export function selectGameOver(state: RoomExperienceState): GameOverView | null {
  const winner = state.roomState?.winner;
  if (!winner) return null;
  return {
    winner: normalizeWinner(winner.winner),
    reason: winner.reason,
    description: winner.description,
  };
}

export function selectDeathAnnouncements(state: RoomExperienceState): readonly string[] {
  const room = state.roomState;
  if (!room) return [];
  const names = new Map(room.players.map((player) => [player.id, player.name || player.id]));
  return [...(room.deaths ?? [])]
    .reverse()
    .map((death) => `第 ${death.dayNumber} 天：${names.get(death.playerId) ?? death.playerId} 死亡`);
}
