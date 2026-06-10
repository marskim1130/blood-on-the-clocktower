export const meta = {
  name: 'vote-engine',
  description: 'Implement vote engine for Blood on the Clocktower MVP',
  phases: [
    { title: 'Implement', detail: 'Implement vote engine with nomination and voting logic' },
    { title: 'Test', detail: 'Write unit tests for vote engine' }
  ]
}

phase('Implement')

const voteEngine = await agent(`Implement a vote engine for Blood on the Clocktower.

The game needs a voting system where:
1. During day phase, any alive player can nominate another alive player
2. All alive players vote thumbs up or down
3. Dead players get 1 ghost vote for the rest of the game
4. Majority threshold = ceil(alivePlayers / 2)
5. If majority votes yes, the nominee is executed (dies)
6. If tie or no majority, no execution

Create the file packages/core/src/vote-engine/index.ts with:

1. Interfaces:
   - VoteState: tracks current nominee, votes cast, vote result
   - Nomination: who nominated whom
   - VoteResult: whether execution happened

2. Functions:
   - canNominate(state, nominatorId, nomineeId): check if nomination is valid
   - nominate(state, nominatorId, nomineeId): create nomination
   - castVote(state, voterId, vote): record a vote
   - tallyVotes(state, alivePlayers, deadPlayers): count votes and determine result
   - resolveExecution(state): apply execution if majority

3. Export a createVoteEngine function that returns vote state management

Follow the patterns in packages/core/src/state-machine/index.ts:
- Use TypeScript with branded types (PlayerId)
- Use immutable state updates (spread operator)
- Export types and functions, not classes
- Use JSDoc comments

The file should be self-contained and not import from other internal modules except types.`, { label: 'implement-vote-engine', phase: 'Implement' })

phase('Test')

const voteTests = await agent(`Write unit tests for the vote engine.

Create the file packages/core/src/vote-engine/__tests__/vote-engine.test.ts

Test cases should cover:

1. Nomination rules:
   - Valid nomination (alive nominator, alive nominee)
   - Cannot nominate self
   - Cannot nominate dead player
   - Cannot nominate if nominator is dead
   - Cannot nominate during non-day phase

2. Voting rules:
   - Each alive player can vote once
   - Dead players get 1 ghost vote
   - Cannot vote twice
   - Cannot vote if not in voting phase

3. Vote tallying:
   - Majority threshold calculation
   - Execution on majority
   - No execution on tie
   - No execution below threshold

4. Ghost votes:
   - Dead player has 1 ghost vote
   - Ghost vote counts toward total
   - Cannot use ghost vote twice

5. Edge cases:
   - All alive players vote yes
   - All alive players vote no
   - Exactly half vote yes
   - Single player game (not valid but handle gracefully)

Use vitest for testing. Follow patterns in packages/core/src/state-machine/__tests__/

Import from '../index.js' (the vote engine module).`, { label: 'write-vote-tests', phase: 'Test' })

return { voteEngine, voteTests }
