import { describe, expect, it } from 'vitest';
import {
  PRIVATE_CHARACTER_REVEAL_MS,
  concealPrivateCharacter,
  createPrivateCharacterVisibility,
  revealPrivateCharacter,
} from './private-character-visibility';

describe('private character visibility', () => {
  it('starts masked and records a thirty second reveal deadline when pressed', () => {
    const initial = createPrivateCharacterVisibility();

    expect(initial).toEqual({ revealed: false, hasRevealed: false, concealAt: null });
    expect(revealPrivateCharacter(initial, 125)).toEqual({
      revealed: true,
      hasRevealed: true,
      concealAt: 125 + PRIVATE_CHARACTER_REVEAL_MS,
    });
  });

  it('conceals immediately without forgetting that the character was viewed', () => {
    expect(concealPrivateCharacter({
      revealed: true,
      hasRevealed: true,
      concealAt: 30_000,
    })).toEqual({
      revealed: false,
      hasRevealed: true,
      concealAt: null,
    });
  });
});
