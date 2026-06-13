// Auto-generated from proto/game.proto
// DO NOT EDIT - regenerate with: pnpm proto:generate

export const enum Team {
  UNSPECIFIED = 0,
  GOOD = 1,
  EVIL = 2,
}

export const enum GamePhase {
  UNSPECIFIED = 0,
  SETUP = 1,
  DAY = 2,
  NIGHT = 3,
  VOTING = 4,
  FINISHED = 5,
}

export const enum DeathCause {
  UNSPECIFIED = 0,
  EXECUTION = 1,
  NIGHT_KILL = 2,
  ABILITY = 3,
}

export const enum NightActionType {
  UNSPECIFIED = 0,
  POISON = 1,
  PROTECT = 2,
  KILL = 3,
  LEARN_TOWNSFOLK = 4,
  LEARN_OUTSIDER = 5,
  LEARN_MINION = 6,
  LEARN_EVIL_PAIRS = 7,
  LEARN_EVIL_NEIGHBORS = 8,
  CHECK_DEMON = 9,
  LEARN_EXECUTED = 10,
  LEARN_DIED = 11,
  LEARN_MASTER = 12,
  LEARN_DEMON = 13,
  CHOOSE_PLAYER = 14,
  NONE = 15,
}

export const enum WinReason {
  UNSPECIFIED = 0,
  IMP_EXECUTED = 1,
  MAYOR_ENDGAME = 2,
  EVIL_MAJORITY = 3,
  SAINT_EXECUTED = 4,
  IMP_STARPASS = 5,
  STORYTELLER_DECISION = 6,
}

export interface ProtoPlayer {
  readonly id: string;
  readonly name: string;
  readonly character?: ProtoCharacter;
  readonly isAlive: boolean;
  readonly votes: number;
  readonly poisonedUntil?: number;
}

export interface ProtoCharacter {
  readonly id: string;
  readonly name: string;
  readonly team: Team;
  readonly ability: string;
}

export interface ProtoGameState {
  readonly id: string;
  readonly phase: GamePhase;
  readonly players: readonly ProtoPlayer[];
  readonly dayNumber: number;
  readonly storytellerId?: string;
  readonly votes: Readonly<Record<string, string>>;
  readonly deaths: readonly ProtoDeathRecord[];
  readonly nomination?: ProtoNomination;
  readonly nightActions: readonly ProtoNightAction[];
  readonly winner?: Team;
}

export interface ProtoDeathRecord {
  readonly playerId: string;
  readonly cause: DeathCause;
  readonly dayNumber: number;
  readonly killedBy?: string;
}

export interface ProtoNomination {
  readonly nominatorId: string;
  readonly nomineeId: string;
  readonly votes: Readonly<Record<string, boolean>>;
  readonly resolved: boolean;
}

export interface ProtoNightAction {
  readonly actorId: string;
  readonly actionType: NightActionType;
  readonly targetIds: readonly string[];
  readonly result?: string;
}

export type ProtoGameEvent =
  | { readonly playerJoined: { readonly player: ProtoPlayer } }
  | { readonly playerLeft: { readonly playerId: string } }
  | { readonly phaseChanged: { readonly phase: GamePhase } }
  | { readonly voteCast: { readonly voterId: string; readonly targetId?: string; readonly decision?: boolean } }
  | { readonly characterAssigned: { readonly playerId: string; readonly character: ProtoCharacter } }
  | { readonly playerDied: { readonly playerId: string; readonly cause: DeathCause; readonly dayNumber: number } }
  | { readonly nominationStarted: { readonly nominatorId: string; readonly nomineeId: string } }
  | { readonly nominationResolved: { readonly nomineeId: string; readonly executed: boolean; readonly yesVotes: number; readonly noVotes: number; readonly requiredVotes: number } }
  | { readonly nightActionSubmitted: { readonly actorId: string; readonly actionType: NightActionType; readonly targetIds: readonly string[]; readonly result?: string } }
  | { readonly gameEnded: { readonly winner: Team; readonly reason: WinReason; readonly description: string } };
