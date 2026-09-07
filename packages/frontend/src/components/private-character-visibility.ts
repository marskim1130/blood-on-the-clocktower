export const PRIVATE_CHARACTER_REVEAL_MS = 30_000;

export interface PrivateCharacterVisibility {
  readonly revealed: boolean;
  readonly hasRevealed: boolean;
  readonly concealAt: number | null;
}

export function createPrivateCharacterVisibility(): PrivateCharacterVisibility {
  return { revealed: false, hasRevealed: false, concealAt: null };
}

export function revealPrivateCharacter(
  state: PrivateCharacterVisibility,
  now: number,
): PrivateCharacterVisibility {
  return {
    ...state,
    revealed: true,
    hasRevealed: true,
    concealAt: now + PRIVATE_CHARACTER_REVEAL_MS,
  };
}

export function concealPrivateCharacter(
  state: PrivateCharacterVisibility,
): PrivateCharacterVisibility {
  return {
    ...state,
    revealed: false,
    concealAt: null,
  };
}
