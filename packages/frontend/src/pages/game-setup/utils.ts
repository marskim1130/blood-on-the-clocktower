export function membershipKey(players: readonly { readonly id: string }[]): string {
  return players.map((player) => player.id).sort().join('|');
}

export function moveSeatClockwise(seatOrder: readonly string[], playerId: string): string[] {
  const currentIndex = seatOrder.indexOf(playerId);
  if (currentIndex < 0 || seatOrder.length < 2) return [...seatOrder];

  const nextOrder = [...seatOrder];
  const movedPlayer = nextOrder.splice(currentIndex, 1)[0];
  if (movedPlayer === undefined) return [...seatOrder];

  const nextIndex = currentIndex === nextOrder.length ? 0 : currentIndex + 1;
  nextOrder.splice(nextIndex, 0, movedPlayer);
  return nextOrder;
}

export function summarizeReadiness(
  players: readonly { readonly id: string; readonly isReady: boolean }[],
  currentPlayerId: string,
): {
  readonly readyCount: number;
  readonly totalCount: number;
  readonly allReady: boolean;
  readonly selfReady: boolean | null;
} {
  const currentPlayer = players.find((player) => player.id === currentPlayerId);
  return {
    readyCount: players.filter((player) => player.isReady).length,
    totalCount: players.length,
    allReady: players.length > 0 && players.every((player) => player.isReady),
    selfReady: currentPlayer?.isReady ?? null,
  };
}

export function canAssignCharacters(configurationIsValid: boolean, allPlayersReady: boolean): boolean {
  return configurationIsValid && allPlayersReady;
}

export function summarizeCharacterConfirmation(
  players: readonly { readonly id: string; readonly hasConfirmedCharacter: boolean }[],
  currentPlayerId: string,
): {
  readonly confirmedCount: number;
  readonly totalCount: number;
  readonly allConfirmed: boolean;
  readonly selfConfirmed: boolean | null;
} {
  const currentPlayer = players.find((player) => player.id === currentPlayerId);
  return {
    confirmedCount: players.filter((player) => player.hasConfirmedCharacter).length,
    totalCount: players.length,
    allConfirmed: players.length > 0 && players.every((player) => player.hasConfirmedCharacter),
    selfConfirmed: currentPlayer?.hasConfirmedCharacter ?? null,
  };
}

export function normalizeDemonBluffs(
  eligibleCharacterIds: readonly string[],
  currentCharacterIds: readonly string[],
): string[] {
  const eligible = new Set(eligibleCharacterIds);
  const result: string[] = [];
  for (const characterId of currentCharacterIds) {
    if (eligible.has(characterId) && !result.includes(characterId)) result.push(characterId);
    if (result.length === 3) return result;
  }
  for (const characterId of eligibleCharacterIds) {
    if (!result.includes(characterId)) result.push(characterId);
    if (result.length === 3) break;
  }
  return result;
}
