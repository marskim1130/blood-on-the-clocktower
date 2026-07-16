export type PlayerId = string & { readonly __brand: 'PlayerId' };

export type Character = {
  readonly id: string;
  readonly name: string;
  readonly team: Team;
  readonly ability: string;
};

export type Team = 'good' | 'evil';

/** How a player died; mirrors death-system/DeathCause. */
export type DeathCause = 'execution' | 'night_kill' | 'ability';

export function createPlayerId(id: string): PlayerId {
  return id as PlayerId;
}
