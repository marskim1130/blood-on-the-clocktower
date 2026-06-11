export const meta = {
  name: 'backend-integration',
  description: 'Integrate MVP core logic into Go backend',
  phases: [
    { title: 'Domain', detail: 'Update game domain types and events' },
    { title: 'Session', detail: 'Update game session with new commands' },
    { title: 'Hub', detail: 'Update hub message handling' },
    { title: 'Test', detail: 'Run tests and fix issues' }
  ]
}

phase('Domain')

const domainTypes = await agent(`Update the Go backend domain types for MVP game flow.

Read and update packages/backend/internal/game/game.go to add:

1. New event types to GameEvent struct:
   - PlayerDied *PlayerDiedEvent - for death tracking
   - NominationStarted *NominationStartedEvent - for voting flow
   - VoteCast *VoteCastEvent (already exists but enhance)
   - NominationResolved *NominationResolvedEvent - vote result
   - NightActionSubmitted *NightActionEvent - for night actions
   - GameEnded *GameEndedEvent - for win conditions

2. New event structs:
   - PlayerDiedEvent: PlayerID string, Cause DeathCause, DayNumber int32
   - NominationStartedEvent: NominatorID string, NomineeID string
   - VoteCastEvent: VoterID string, TargetID string, Decision bool (enhance existing)
   - NominationResolvedEvent: NomineeID string, Executed bool, YesVotes int, NoVotes int
   - NightActionEvent: ActorID string, ActionType NightActionType, TargetIDs []string, Result string
   - GameEndedEvent: Winner Team, Reason WinReason, Description string

3. New enums/types:
   - DeathCause: "execution", "night_kill", "ability"
   - NightActionType: "poison", "protect", "kill", "learn_townsfolk", etc.
   - WinReason: "imp_executed", "mayor_endgame", "evil_majority", "saint_executed", "imp_starpass"

4. Update GameState to include:
   - Deaths []DeathRecord
   - Nomination *Nomination (current active nomination)
   - NightActions []NightAction
   - Winner *Team

5. Add Nomination struct:
   - NominatorID string
   - NomineeID string
   - Votes map[string]bool (voterID -> decision)
   - Resolved bool

6. Add DeathRecord struct:
   - PlayerID string
   - Cause DeathCause
   - DayNumber int32
   - KilledBy string (optional)

Follow existing code patterns in the file. Use pointer types for optional fields in events.`, { label: 'update-domain-types', phase: 'Domain' })

phase('Session')

const gameSession = await agent(`Update the game session with new commands for MVP.

Read and update packages/backend/internal/ws/game_session.go to add:

1. New command types (implement Command interface):
   - StartGameCmd: SenderID string - transitions from Setup to Day/Night
   - ChangePhaseCmd: SenderID string, Phase GamePhase - storyteller changes phase
   - NominateCmd: SenderID string, NomineeID string - player nominates another
   - CastVoteCmd: SenderID string, Decision bool - player votes
   - ResolveNominationCmd: SenderID string - storyteller resolves vote
   - ExecutePlayerCmd: SenderID string, PlayerID string - storyteller executes
   - SubmitNightActionCmd: SenderID string, ActionType string, TargetIDs []string
   - ResolveNightCmd: SenderID string - storyteller resolves night
   - KillPlayerCmd: SenderID string, PlayerID string, Cause DeathCause

2. Add apply functions for each command:
   - applyStartGame: validate phase=Setup, storyteller set, characters assigned, transition to Night (first night)
   - applyChangePhase: validate storyteller, validate transition (Day->Voting->Day->Night->Day cycle)
   - applyNominate: validate phase=Day, nominator alive, nominee alive, not self-nomination
   - applyCastVote: validate phase=Voting, voter alive or has ghost vote, not duplicate vote
   - applyResolveNomination: count votes, check majority, transition back to Day
   - applyExecutePlayer: set player IsAlive=false, record death, check win conditions
   - applySubmitNightAction: validate phase=Night, validate character ability, record action
   - applyResolveNight: process night actions, determine deaths, transition to Day

3. Update GameSession struct:
   - Add phase GamePhase field
   - Add dayNumber int32 field
   - Add nomination *Nomination field
   - Add nightActions []NightAction field
   - Add deaths []DeathRecord field
   - Add ghostVotesUsed map[string]bool field

4. Add win condition checking:
   - checkWinConditions(players, deaths, phase) -> *GameEndedEvent
   - Check after each death: demon dead, saint executed, mayor endgame, evil majority

5. Update Apply() method to handle new commands.

Follow existing command pattern in the file. Return ApplyResult with events to broadcast.`, { label: 'update-game-session', phase: 'Session' })

phase('Hub')

const hubUpdate = await agent(`Update the hub to handle new message types.

Read and update packages/backend/internal/ws/hub.go to:

1. Add new message type constants (if not already in message.go):
   - START_GAME
   - CHANGE_PHASE
   - NOMINATE
   - CAST_VOTE
   - RESOLVE_NOMINATION
   - EXECUTE_PLAYER
   - SUBMIT_NIGHT_ACTION
   - RESOLVE_NIGHT

2. Update handleMessage switch to route new types:
   - START_GAME -> StartGameCmd
   - CHANGE_PHASE -> ChangePhaseCmd
   - NOMINATE -> NominateCmd
   - CAST_VOTE -> CastVoteCmd
   - RESOLVE_NOMINATION -> ResolveNominationCmd
   - EXECUTE_PLAYER -> ExecutePlayerCmd
   - SUBMIT_NIGHT_ACTION -> SubmitNightActionCmd
   - RESOLVE_NIGHT -> ResolveNightCmd

3. Update buildRoomState to include new fields:
   - Phase
   - DayNumber
   - Nomination (current active)
   - Deaths (list of dead player IDs)
   - Winner (if game ended)

4. Ensure storyteller-only commands are validated:
   - CHANGE_PHASE, RESOLVE_NOMINATION, EXECUTE_PLAYER, RESOLVE_NIGHT

Follow existing patterns in the file for message routing and command creation.`, { label: 'update-hub', phase: 'Hub' })

phase('Test')

const testUpdate = await agent(`Update tests and verify the integration.

1. Update existing tests in packages/backend/internal/ws/game_session_test.go:
   - Add tests for new commands
   - Test phase transitions
   - Test voting flow
   - Test night action flow
   - Test win conditions

2. Update existing tests in packages/backend/internal/ws/hub_test.go:
   - Test new message types
   - Test storyteller validation
   - Test game flow end-to-end

3. Run all backend tests:
   cd packages/backend && go test ./...

4. Fix any compilation errors or test failures.

Output a summary of:
- Tests added
- Tests passing
- Any issues found and fixed`, { label: 'run-tests', phase: 'Test' })

return { domainTypes, gameSession, hubUpdate, testUpdate }
