export const meta = {
  name: 'mvp-development',
  description: 'Develop Blood on the Clocktower to minimum playable version',
  phases: [
    { title: 'Design', detail: 'Design game loop architecture and data models' },
    { title: 'Core Logic', detail: 'Implement vote engine, death system, win conditions' },
    { title: 'Night Phase', detail: 'Implement night phase framework and character abilities' },
    { title: 'Backend', detail: 'Update backend to support new game events and state transitions' },
    { title: 'Frontend', detail: 'Update frontend UI for complete game flow' },
    { title: 'Integration', detail: 'Integration testing and bug fixes' }
  ]
}

// Phase 1: Design the game loop architecture
phase('Design')

const gameLoopDesign = await agent(`Design the game loop architecture for Blood on the Clocktower MVP.

Current implementation:
- State machine with phases: setup, day, night, voting, finished
- Basic event types: PLAYER_JOINED, PLAYER_LEFT, PHASE_CHANGED, VOTE_CAST, CHARACTER_ASSIGNED
- Character system with 22 Trouble Brewing characters
- WebSocket-based communication

Missing for MVP:
1. Vote Engine (nomination, tallying, execution)
2. Night Phase Framework (wake order, character abilities, death resolution)
3. Day Phase Game Loop (discussion, nominations, execution)
4. Win Condition Checking (demon-dead, final-two, mayor, saint)
5. Player Death System (death tracking, ghost votes, death triggers)
6. Game Start/End Flow (start game, game over, post-game reveal)

Design requirements:
- Extend existing state machine and event system
- Support human Storyteller controlling night actions
- Maintain backward compatibility with existing code
- Follow existing code patterns and naming conventions

Output a detailed design document with:
1. New event types needed
2. State machine transitions
3. Vote engine logic
4. Night phase framework
5. Win condition rules
6. Death system mechanics
7. Implementation priority order`, { label: 'design-game-loop', phase: 'Design' })

// Phase 2: Implement core game logic
phase('Core Logic')

// 2.1 Vote Engine
const voteEngine = await agent(`Implement the vote engine for Blood on the Clocktower.

Based on the design, create a vote engine that handles:
1. Nomination system:
   - Any alive player can nominate another alive player during day phase
   - Nomination requires nominator and nominee to be alive
   - Nominations happen sequentially during day phase

2. Voting mechanics:
   - All alive players vote (thumbs up/down)
   - Dead players get 1 ghost vote for the rest of the game
   - Majority threshold = ceil(alivePlayers / 2)
   - If majority: nominee is executed (dies)
   - If tie or no majority: no execution

3. Vote state tracking:
   - Current nominee
   - Votes cast (player -> vote)
   - Vote result (executed/not executed)

4. State transitions:
   - After nomination: move to voting phase
   - After voting completes: check for execution, then back to day or night

Create the implementation in packages/core/src/vote-engine/index.ts with:
- VoteEngine class or functions
- VoteState interface
- VoteEvent types
- Unit tests in packages/core/src/vote-engine/__tests__/vote-engine.test.ts

Follow existing code patterns from state-machine/index.ts`, { label: 'implement-vote-engine', phase: 'Core Logic' })

// 2.2 Death System
const deathSystem = await agent(`Implement the player death system for Blood on the Clocktower.

Create a death system that handles:
1. Death tracking:
   - Record cause of death (execution, night kill, ability)
   - Record day number of death
   - Track who killed whom (for abilities like Undertaker)

2. Ghost votes:
   - Dead players get exactly 1 ghost vote for the rest of the game
   - Ghost votes are tracked separately from alive player votes
   - Ghost votes can be used on any nomination

3. Death triggers:
   - Saint execution: evil wins immediately
   - Scarlet Woman: if Imp dies with 5+ players alive, Scarlet Woman becomes Imp
   - Ravenkeeper: can learn character of player who died at night
   - Undertaker: learns character of player executed during day

4. Death events:
   - PLAYER_DIED event with cause and day number
   - DEATH_ANNOUNCEMENT broadcast to all players

Create the implementation in packages/core/src/death-system/index.ts with:
- DeathSystem class or functions
- DeathState interface
- DeathEvent types
- Unit tests in packages/core/src/death-system/__tests__/death-system.test.ts

Follow existing code patterns from state-machine/index.ts`, { label: 'implement-death-system', phase: 'Core Logic' })

// 2.3 Win Conditions
const winConditions = await agent(`Implement win condition checking for Blood on the Clocktower.

Create a win condition checker that evaluates:
1. Good wins conditions:
   - Imp is executed (and no Scarlet Woman trigger)
   - Mayor endgame: 3 players alive, no execution during day

2. Evil wins conditions:
   - Only 2 players alive (evil wins by majority)
   - Saint is executed (evil wins immediately)
   - Imp kills themselves at night (if no Scarlet Woman)

3. Win condition evaluation:
   - Check after each execution
   - Check after each night death
   - Check at end of day phase if no execution

4. Game over flow:
   - Set phase to 'finished'
   - Broadcast GAME_OVER event with winner team
   - Reveal all characters (post-game)

Create the implementation in packages/core/src/win-conditions/index.ts with:
- WinConditionChecker class or functions
- WinResult interface
- Unit tests in packages/core/src/win-conditions/__tests__/win-conditions.test.ts

Follow existing code patterns from state-machine/index.ts`, { label: 'implement-win-conditions', phase: 'Core Logic' })

// Phase 3: Night Phase Framework
phase('Night Phase')

const nightPhase = await agent(`Implement the night phase framework for Blood on the Clocktower.

Create a night phase system that handles:
1. Wake order system:
   - Define wake order for Trouble Brewing characters
   - First night order (different from subsequent nights)
   - Subsequent night order

2. Character abilities:
   - Empath: learns number of evil neighbors
   - Fortune Teller: picks 2 players, learns if either is demon
   - Poisoner: poisons a player (ability malfunctions)
   - Imp: kills a player
   - Washerwoman/Librarian/Investigator: learn about character tokens
   - Chef: learns number of evil pairs
   - Monk: protects a player from Imp kill
   - Soldier: cannot be killed by Imp
   - Butler: learns who their master is
   - Drunk: thinks they are a different character

3. Storyteller night actions:
   - Wake each player in order
   - Record their action/choice
   - Resolve ability effects
   - Track poisoning/drunk status

4. Night resolution:
   - Collect all night actions
   - Apply effects in correct order
   - Determine night deaths
   - Broadcast dawn results

Create the implementation in packages/core/src/night-phase/index.ts with:
- NightPhaseManager class or functions
- NightAction interface
- WakeOrder definitions
- Unit tests in packages/core/src/night-phase/__tests__/night-phase.test.ts

Follow existing code patterns from state-machine/index.ts`, { label: 'implement-night-phase', phase: 'Night Phase' })

// Phase 4: Backend Updates
phase('Backend')

const backendUpdates = await agent(`Update the Go backend to support new game events and state transitions.

Based on the core logic implementations, update:

1. Game events (packages/backend/internal/game/game.go):
   - Add new event types: PLAYER_DIED, GAME_OVER, VOTE_STARTED, VOTE_ENDED
   - Update GameState with death tracking, vote state, night actions

2. Game session (packages/backend/internal/ws/game_session.go):
   - Add commands for: StartGame, NominatePlayer, CastVote, NightAction, KillPlayer
   - Update Apply() method to handle new commands
   - Implement state transitions for game loop

3. WebSocket messages (packages/backend/internal/ws/message.go):
   - Add new message types for game flow
   - Update RoomState with game state

4. Hub message handling (packages/backend/internal/ws/hub.go):
   - Handle new message types from frontend
   - Broadcast game events to room

5. Character abilities (packages/backend/internal/game/characters.go):
   - Add ability execution logic
   - Handle special character interactions

Update existing tests and add new tests for:
- Game flow integration
- Night phase resolution
- Vote counting
- Win condition detection`, { label: 'update-backend', phase: 'Backend' })

// Phase 5: Frontend Updates
phase('Frontend')

const frontendUpdates = await agent(`Update the Taro frontend to support complete game flow.

Based on the backend updates, update:

1. Game state management (packages/core/src/state-machine/index.ts):
   - Add new event types to reducer
   - Update GameState interface with new fields

2. WebSocket messages (packages/core/src/websocket/index.ts):
   - Add new message types
   - Update client methods for game actions

3. Main page UI (packages/frontend/src/pages/index/index.tsx):
   - Add game phase display
   - Add voting UI (nomination list, vote buttons)
   - Add night phase UI (storyteller controls)
   - Add death announcements
   - Add win condition display

4. New components:
   - VotePanel component for voting UI
   - NightPanel component for storyteller night controls
   - GameStatus component for phase/day display
   - PlayerList component with death indicators

5. Styling:
   - Update CSS for game flow UI
   - Add animations for phase transitions

Follow existing Taro/React patterns in the codebase`, { label: 'update-frontend', phase: 'Frontend' })

// Phase 6: Integration Testing
phase('Integration')

const integrationTest = await agent(`Perform integration testing and bug fixes.

1. Test complete game flow:
   - Create room -> join players -> set storyteller -> assign characters
   - Start game -> first night -> day phase -> nominations -> voting
   - Execution -> night phase -> repeat until win condition

2. Test edge cases:
   - Saint execution (evil wins)
   - Mayor endgame (3 alive, no execution)
   - Scarlet Woman trigger (Imp dies with 5+ players)
   - Ghost votes from dead players
   - Reconnection during game

3. Test storyteller tools:
   - Night phase control
   - Death announcements
   - Win condition detection

4. Fix any bugs found during testing

5. Update documentation:
   - Update PRD with implementation details
   - Add API documentation for new messages
   - Update README with setup instructions

Output a test report with:
- Tests passed
- Bugs found and fixed
- Remaining issues`, { label: 'integration-test', phase: 'Integration' })

return {
  design: gameLoopDesign,
  voteEngine,
  deathSystem,
  winConditions,
  nightPhase,
  backendUpdates,
  frontendUpdates,
  integrationTest
}
