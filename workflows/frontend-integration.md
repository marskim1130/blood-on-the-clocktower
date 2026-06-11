export const meta = {
  name: 'frontend-integration',
  description: 'Integrate MVP core logic into Taro frontend',
  phases: [
    { title: 'State', detail: 'Update state machine and WebSocket messages' },
    { title: 'UI', detail: 'Update game UI components' },
    { title: 'Test', detail: 'Build and verify' }
  ]
}

phase('State')

const stateUpdate = await agent(`Update the core state machine and WebSocket messages for MVP.

1. Update packages/core/src/state-machine/index.ts:
   - Add new event types to GameEvent union:
     - PLAYER_DIED with playerId, cause, dayNumber
     - NOMINATION_STARTED with nominatorId, nomineeId
     - VOTE_CAST (already exists, verify fields)
     - NOMINATION_RESOLVED with nomineeId, executed, yesVotes, noVotes
     - NIGHT_ACTION with actorId, actionType, targetIds, result
     - GAME_OVER with winner, reason, description
   - Update gameReducer to handle new events:
     - PLAYER_DIED: mark player as dead, track ghost votes
     - NOMINATION_STARTED: store current nomination
     - VOTE_CAST: record vote
     - NOMINATION_RESOLVED: clear nomination, apply execution if needed
     - NIGHT_ACTION: record night action
     - GAME_OVER: set phase to 'finished', store winner

2. Update packages/core/src/websocket/index.ts:
   - Add new message types to ClientMessage:
     - START_GAME
     - CHANGE_PHASE with phase
     - NOMINATE with nomineeId
     - CAST_VOTE with decision (boolean)
     - RESOLVE_NOMINATION
     - EXECUTE_PLAYER with playerId
     - SUBMIT_NIGHT_ACTION with actionType, targetIds
     - RESOLVE_NIGHT
   - Add corresponding methods to GameWebSocketClient

3. Update types if needed to match backend.

Follow existing patterns in the codebase. Use branded types (PlayerId, GameId).`, { label: 'update-state', phase: 'State' })

phase('UI')

const uiUpdate = await agent(`Update the frontend UI for MVP game flow.

Read and update packages/frontend/src/pages/index/index.tsx to add:

1. Game phase display:
   - Show current phase (Setup, Day, Voting, Night, Finished)
   - Show day number
   - Show whose turn it is (during night)

2. Voting UI:
   - Nomination button (during Day phase)
   - Vote buttons (thumbs up/down during Voting phase)
   - Vote result display (executed/spared)
   - Ghost vote indicator for dead players

3. Night phase UI (Storyteller only):
   - Night action form (select action type, select targets)
   - Submit night action button
   - Resolve night button
   - Night results display

4. Death tracking:
   - Mark dead players with visual indicator
   - Show death announcements
   - Track ghost votes remaining

5. Win condition display:
   - Game over screen with winner and reason
   - Post-game reveal (show all characters)
   - Back to lobby button

6. Storyteller controls:
   - Start game button (after characters assigned)
   - Change phase buttons
   - Execute player button
   - Resolve nomination button

Follow existing Taro/React patterns in the file. Use the WebSocket client for all actions.`, { label: 'update-ui', phase: 'UI' })

phase('Test')

const buildTest = await agent(`Build and test the frontend.

1. Run TypeScript check:
   cd packages/frontend && pnpm tsc --noEmit

2. Run frontend build:
   cd packages/frontend && pnpm build

3. Fix any compilation errors.

4. Run core tests to ensure nothing is broken:
   cd packages/core && pnpm test

5. Output a summary of:
   - Build status
   - Any errors found and fixed
   - Tests passing`, { label: 'build-test', phase: 'Test' })

return { stateUpdate, uiUpdate, buildTest }
