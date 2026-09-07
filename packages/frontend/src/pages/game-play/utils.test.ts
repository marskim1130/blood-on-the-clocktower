import { describe, expect, it } from 'vitest';
import { currentVoterId, executionProposalText } from './utils';

describe('game play execution proposal', () => {
  it('describes a unique candidate as pending storyteller confirmation', () => {
    expect(executionProposalText(
      [{ id: 'p1', name: '阿青' }],
      'p1',
      3,
      false,
    )).toBe('阿青当前上台（3 票）；确认日终后才会被处决。');
  });
});

describe('clockwise ballot', () => {
  it('returns only the player at the authoritative current voter index', () => {
    expect(currentVoterId({
      voterOrder: ['p2', 'p3', 'p1'],
      currentVoterIndex: 1,
    })).toBe('p3');
    expect(currentVoterId({
      voterOrder: ['p2', 'p3', 'p1'],
      currentVoterIndex: 3,
    })).toBeUndefined();
  });
});
