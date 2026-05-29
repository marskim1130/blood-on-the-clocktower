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

export interface ProtoPlayer {
  readonly id: string;
  readonly name: string;
  readonly character?: ProtoCharacter;
  readonly isAlive: boolean;
  readonly votes: number;
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
}

export type ProtoGameEvent =
  | { readonly playerJoined: { readonly player: ProtoPlayer } }
  | { readonly playerLeft: { readonly playerId: string } }
  | { readonly phaseChanged: { readonly phase: GamePhase } }
  | { readonly voteCast: { readonly voterId: string; readonly targetId?: string } }
  | { readonly characterAssigned: { readonly playerId: string; readonly character: ProtoCharacter } };
