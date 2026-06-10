export const meta = {
  name: 'night-phase',
  description: 'Implement night phase framework for Blood on the Clocktower MVP',
  phases: [
    { title: 'Implement', detail: 'Implement night phase with character abilities' },
    { title: 'Test', detail: 'Write unit tests for night phase' }
  ]
}

phase('Implement')

const nightPhase = await agent(`Implement a night phase framework for Blood on the Clocktower.

The game needs a night phase where:
1. Storyteller wakes players in order
2. Each player performs their night action
3. Storyteller records results
4. Dawn: resolve night actions and announce deaths

Create the file packages/core/src/night-phase/index.ts with:

1. Interfaces:
   - NightState: current night number, actions recorded, wake order
   - NightAction: playerId, action type, target(s), result
   - WakeOrder: character -> order number

2. Constants:
   - FIRST_NIGHT_ORDER: wake order for first night
   - SUBSEQUENT_NIGHT_ORDER: wake order for other nights

3. Functions:
   - getWakeOrder(nightNumber): get wake order for this night
   - getNextToWake(state, nightNumber): get next character to wake
   - recordNightAction(state, action): record a night action
   - resolveNightActions(state): resolve all night actions
   - getNightResults(state): get results to announce at dawn

4. Character abilities (Trouble Brewing):
   - Washerwoman: learns one of two players is a specific townsfolk
   - Librarian: learns one of two players is a specific outsider (or no outsiders)
   - Investigator: learns one of two players is a specific minion
   - Chef: learns number of evil pairs
   - Empath: learns number of evil neighbors
   - Fortune Teller: picks 2 players, learns if either is demon
   - Undertaker: learns character of executed player
   - Monk: protects a player from Imp kill
   - Ravenkeeper: learns character of player who died
   - Butler: learns who their master is
   - Drunk: thinks they are a different character (handled by storyteller)
   - Poisoner: poisons a player (ability malfunctions)
   - Imp: kills a player

5. Night resolution:
   - Collect all actions in wake order
   - Apply effects (poison, protection, kills)
   - Determine who dies
   - Return dawn results

Export a createNightPhaseManager function.

Follow the patterns in packages/core/src/state-machine/index.ts:
- Use TypeScript with branded types (PlayerId)
- Use immutable state updates
- Export types and functions, not classes
- Use JSDoc comments`, { label: 'implement-night-phase', phase: 'Implement' })

phase('Test')

const nightTests = await agent(`Write unit tests for the night phase.

Create the file packages/core/src/night-phase/__tests__/night-phase.test.ts

Test cases should cover:

1. Wake order:
   - First night has correct order
   - Subsequent nights have correct order
   - Get next character to wake
   - Skip dead characters

2. Night actions:
   - Record night action
   - Cannot record action for dead player
   - Cannot record action out of order
   - Multiple actions for same character (Poisoner + Imp)

3. Character abilities:
   - Washerwoman learns townsfolk
   - Librarian learns outsider
   - Investigator learns minion
   - Chef learns evil pairs
   - Empath learns evil neighbors
   - Fortune Teller checks demon
   - Monk protects from Imp kill
   - Poisoner poisons player
   - Imp kills player

4. Night resolution:
   - Resolve all actions in order
   - Apply protection (Monk saves from Imp)
   - Apply poisoning (ability malfunctions)
   - Determine night deaths
   - Handle multiple deaths

5. Edge cases:
   - All players dead (should not happen)
   - No night actions recorded
   - Duplicate actions
   - Invalid targets

Use vitest for testing. Follow patterns in packages/core/src/state-machine/__tests__/

Import from '../index.js' (the night phase module).`, { label: 'write-night-tests', phase: 'Test' })

return { nightPhase, nightTests }
