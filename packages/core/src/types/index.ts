// Game types - these will be auto-generated from ProtoBuf
export type PlayerId = string & { readonly __brand: 'PlayerId' };
export type GameId = string & { readonly __brand: 'GameId' };

export interface Player {
  readonly id: PlayerId;
  readonly name: string;
  readonly character: Character | null;
  readonly isAlive: boolean;
  readonly votes: number;
}

export type Character = {
  readonly id: string;
  readonly name: string;
  readonly team: Team;
  readonly ability: string;
};

export type Team = 'good' | 'evil';

export type GamePhase = 'setup' | 'day' | 'night' | 'voting' | 'finished';

export interface GameState {
  readonly id: GameId;
  readonly phase: GamePhase;
  readonly players: readonly Player[];
  readonly dayNumber: number;
  readonly storyteller: PlayerId | null;
  readonly votes: ReadonlyMap<PlayerId, PlayerId | null>;
}

export type GameEvent =
  | { readonly type: 'PLAYER_JOINED'; readonly player: Player }
  | { readonly type: 'PLAYER_LEFT'; readonly playerId: PlayerId }
  | { readonly type: 'PHASE_CHANGED'; readonly phase: GamePhase }
  | { readonly type: 'VOTE_CAST'; readonly voterId: PlayerId; readonly targetId: PlayerId | null }
  | { readonly type: 'CHARACTER_ASSIGNED'; readonly playerId: PlayerId; readonly character: Character }
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
