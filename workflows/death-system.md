export const meta = {
  name: 'death-system',
  description: 'Implement death system for Blood on the Clocktower MVP',
  phases: [
    { title: 'Implement', detail: 'Implement death tracking and ghost votes' },
    { title: 'Test', detail: 'Write unit tests for death system' }
  ]
}

phase('Implement')

const deathSystem = await agent(`Implement a death system for Blood on the Clocktower.

The game needs death tracking with:
1. Record cause of death (execution, night kill, ability)
2. Record day number of death
3. Track who killed whom (for abilities like Undertaker)
4. Dead players get exactly 1 ghost vote for the rest of the game
5. Death triggers: Saint execution (evil wins), Scarlet Woman (Imp dies with 5+ players)

Create the file packages/core/src/death-system/index.ts with:

1. Interfaces:
   - DeathRecord: playerId, cause, dayNumber, killerId (optional)
   - DeathState: tracks all deaths, ghost votes remaining
   - GhostVote: playerId, remaining votes (always 1)

2. Functions:
   - recordDeath(state, playerId, cause, dayNumber, killerId?): add death record
   - canUseGhostVote(state, playerId): check if dead player has ghost vote
   - useGhostVote(state, playerId): use ghost vote (decrement to 0)
   - getGhostVotesRemaining(state, playerId): get remaining ghost votes
   - isPlayerDead(state, playerId): check if player is dead
   - getDeathsByDay(state, dayNumber): get deaths on specific day
   - getDeathsByCause(state, cause): get deaths by cause
   - checkDeathTriggers(state, deadPlayerId): check for special death effects

3. Death causes enum:
   - EXECUTION: voted out during day
   - NIGHT_KILL: killed by Imp at night
   - ABILITY: killed by character ability

4. Death triggers:
   - Saint execution: return { type: 'EVIL_WINS', reason: 'Saint executed' }
   - Scarlet Woman: return { type: 'SCARLET_WOMAN_TRIGGER', newImpId: ... }

Export a createDeathSystem function that returns death state management.

Follow the patterns in packages/core/src/state-machine/index.ts:
- Use TypeScript with branded types (PlayerId)
- Use immutable state updates (spread operator)
- Export types and functions, not classes
- Use JSDoc comments`, { label: 'implement-death-system', phase: 'Implement' })

phase('Test')

const deathTests = await agent(`Write unit tests for the death system.

Create the file packages/core/src/death-system/__tests__/death-system.test.ts

Test cases should cover:

1. Death recording:
   - Record execution death
   - Record night kill death
   - Record ability death
   - Track day number
   - Track killer (optional)

2. Ghost votes:
   - Dead player has 1 ghost vote
   - Use ghost vote decrements to 0
   - Cannot use ghost vote twice
   - Cannot use ghost vote if alive

3. Death queries:
   - Check if player is dead
   - Get deaths by day
   - Get deaths by cause
   - Get death record for player

4. Death triggers:
   - Saint execution detection
   - Scarlet Woman trigger detection
   - No trigger for normal deaths

5. Edge cases:
   - Player dies twice (should not happen but handle)
   - Unknown player death
   - Death on day 0 (first night)

Use vitest for testing. Follow patterns in packages/core/src/state-machine/__tests__/

Import from '../index.js' (the death system module).`, { label: 'write-death-tests', phase: 'Test' })

return { deathSystem, deathTests }
