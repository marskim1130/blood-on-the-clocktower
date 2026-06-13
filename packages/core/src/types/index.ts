// Game types - these will be auto-generated from ProtoBuf
export type PlayerId = string & { readonly __brand: 'PlayerId' };
export type GameId = string & { readonly __brand: 'GameId' };

export interface Player {
  readonly id: PlayerId;
  readonly name: string;
  readonly character: Character | null;
  readonly shownCharacter?: Character | null;
  readonly isAlive: boolean;
  readonly votes: number;
  readonly poisonedUntil?: number;
}

export type Character = {
  readonly id: string;
  readonly name: string;
  readonly team: Team;
  readonly ability: string;
};

export type Team = 'good' | 'evil';

export type GamePhase = 'setup' | 'day' | 'night' | 'voting' | 'finished';

/** How a player died — canonical source, mirrors death-system/DeathCause. */
export type DeathCause = 'execution' | 'night_kill' | 'ability';

/** Snapshot of an in-progress nomination. */
export interface NominationState {
  readonly nominatorId: PlayerId;
  readonly nomineeId: PlayerId;
  /** Votes cast so far: voterId -> true (guilty) / false (innocent). */
  readonly votes: ReadonlyMap<PlayerId, boolean>;
}

/** A single night action recorded by the storyteller. */
export interface NightActionRecord {
  readonly actorId: PlayerId;
  readonly actionType: string;
  readonly targetIds: readonly PlayerId[];
  readonly result: string | null;
}

export interface GameState {
  readonly id: GameId;
  readonly phase: GamePhase;
  readonly players: readonly Player[];
  readonly dayNumber: number;
  readonly storyteller: PlayerId | null;
  readonly votes: ReadonlyMap<PlayerId, PlayerId | null>;
  /** Active nomination (null when no nomination is in progress). */
  readonly currentNomination: NominationState | null;
  /** All night actions recorded during the current night phase. */
  readonly nightActions: readonly NightActionRecord[];
  /** Death records keyed by dead player id. */
  readonly deathRecords: ReadonlyMap<PlayerId, { readonly cause: DeathCause; readonly dayNumber: number }>;
  /** Dead players who still have their one ghost vote available. */
  readonly ghostVotesRemaining: ReadonlySet<PlayerId>;
  /** Populated when the game ends. */
  readonly winner: Team | null;
  readonly winReason: string | null;
  readonly winDescription: string | null;
}

export type GameEvent =
  | { readonly type: 'PLAYER_JOINED'; readonly player: Player }
  | { readonly type: 'PLAYER_LEFT'; readonly playerId: PlayerId }
  | { readonly type: 'PHASE_CHANGED'; readonly phase: GamePhase }
  | { readonly type: 'VOTE_CAST'; readonly voterId: PlayerId; readonly targetId: PlayerId | null }
  | {
      readonly type: 'CHARACTER_ASSIGNED';
      readonly playerId: PlayerId;
      readonly character: Character;
      readonly shownCharacter?: Character | null;
    }
  | {
      readonly type: 'PLAYER_DIED';
      readonly playerId: PlayerId;
      readonly cause: DeathCause;
      readonly dayNumber: number;
    }
  | {
      readonly type: 'NOMINATION_STARTED';
      readonly nominatorId: PlayerId;
      readonly nomineeId: PlayerId;
    }
  | {
      readonly type: 'NOMINATION_RESOLVED';
      readonly nomineeId: PlayerId;
      readonly executed: boolean;
      readonly yesVotes: number;
      readonly noVotes: number;
    }
  | {
      readonly type: 'NIGHT_ACTION';
      readonly actorId: PlayerId;
      readonly actionType: string;
      readonly targetIds: readonly PlayerId[];
      readonly result: string | null;
    }
  | {
      readonly type: 'GAME_OVER';
      readonly winner: Team;
      readonly reason: string;
      readonly description: string;
      readonly revealedPlayers: readonly { readonly playerId: PlayerId; readonly character: Character | null }[];
    };

// Type-safe branded constructor
export function createPlayerId(id: string): PlayerId {
  return id as PlayerId;
}

export function createGameId(id: string): GameId {
  return id as GameId;
}
