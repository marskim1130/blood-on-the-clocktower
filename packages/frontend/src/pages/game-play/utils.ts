export function executionProposalText(
  players: readonly { readonly id: string; readonly name: string }[],
  candidateId: string | undefined,
  highestVotes: number | undefined,
  tied: boolean | undefined,
): string {
  const votes = highestVotes ?? 0;
  if (tied) {
    return `最高票并列（${votes} 票）；确认日终后今天无人处决。`;
  }
  if (!candidateId) {
    return '当前无人上台；确认日终后今天无人处决。';
  }
  const index = players.findIndex((player) => player.id === candidateId);
  const candidateName = players[index]?.name || (index >= 0 ? `座位 ${index + 1}` : '未知玩家');
  return `${candidateName}当前上台（${votes} 票）；确认日终后才会被处决。`;
}

export function currentVoterId(nomination: {
  readonly voterOrder: readonly string[];
  readonly currentVoterIndex: number;
}): string | undefined {
  if (!Number.isInteger(nomination.currentVoterIndex) || nomination.currentVoterIndex < 0) {
    return undefined;
  }
  return nomination.voterOrder[nomination.currentVoterIndex];
}
