export const meta = {
  name: 'win-conditions',
  description: 'Implement win condition checking for Blood on the Clocktower MVP',
  phases: [
    { title: 'Implement', detail: 'Implement win condition checker' },
    { title: 'Test', detail: 'Write unit tests for win conditions' }
  ]
}

phase('Implement')

const winConditions = await agent(`Implement win condition checking for Blood on the Clocktower.

The game needs to detect when the game ends:
1. Good wins if:
   - Imp is executed (and no Scarlet Woman trigger)
   - Mayor endgame: 3 players alive, no execution during day

2. Evil wins if:
   - Only 2 players alive (evil wins by majority)
   - Saint is executed (evil wins immediately)
   - Imp kills themselves at night (if no Scarlet Woman)

Create the file packages/core/src/win-conditions/index.ts with:

1. Interfaces:
   - WinResult: winner ('good' | 'evil'), reason, trigger (optional)
   - GameState: minimal state needed for checking (players, phase, deaths, votes)

2. Functions:
   - checkWinConditions(state): evaluate all win conditions
   - checkGoodWin(state): check if good team wins
   - checkEvilWin(state): check if evil team wins
   - checkMayorEndgame(state): check mayor special condition
   - checkSaintExecution(state): check saint execution
   - checkFinalTwo(state): check if only 2 players alive
   - checkImpDeath(state): check if imp died

3. Win conditions:
   - Good wins: imp executed, mayor endgame
   - Evil wins: final two, saint executed, imp suicide

4. Game over flow:
   - Set phase to 'finished'
   - Return WinResult with winner and reason

Export a createWinConditionChecker function.

Follow the patterns in packages/core/src/state-machine/index.ts:
- Use TypeScript with branded types (PlayerId, GameId)
- Use immutable state updates
- Export types and functions, not classes
- Use JSDoc comments`, { label: 'implement-win-conditions', phase: 'Implement' })

phase('Test')

const winTests = await agent(`Write unit tests for win condition checking.

Create the file packages/core/src/win-conditions/__tests__/win-conditions.test.ts

Test cases should cover:

1. Good win conditions:
   - Imp executed → good wins
   - Mayor endgame (3 alive, no execution) → good wins
   - Imp dies at night with Scarlet Woman → no good win

2. Evil win conditions:
   - Only 2 players alive → evil wins
   - Saint executed → evil wins
   - Imp kills themselves → evil wins

3. Edge cases:
   - Multiple win conditions triggered (should pick first)
   - Game not started yet
   - Game already finished
   - Invalid game state

4. Special cases:
   - Scarlet Woman trigger prevents imp death win
   - Mayor endgame requires no execution
   - Final two check counts alive players correctly

5. Integration with death system:
   - Win condition check after execution
   - Win condition check after night kill

Use vitest for testing. Follow patterns in packages/core/src/state-machine/__tests__/

Import from '../index.js' (the win conditions module).`, { label: 'write-win-tests', phase: 'Test' })

return { winConditions, winTests }
