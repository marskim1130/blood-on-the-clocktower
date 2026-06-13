import { describe, it, expect } from 'vitest';
import {
  Team,
  GamePhase,
  DeathCause,
  NightActionType,
  WinReason,
  type ProtoPlayer,
  type ProtoGameState,
  type ProtoGameEvent,
} from '../index.js';

describe('ProtoBuf generated types', () => {
  describe('Team enum', () => {
    it('has UNSPECIFIED, GOOD, EVIL values', () => {
      expect(Team.UNSPECIFIED).toBe(0);
      expect(Team.GOOD).toBe(1);
      expect(Team.EVIL).toBe(2);
    });
  });

  describe('GamePhase enum', () => {
    it('has all game phases', () => {
      expect(GamePhase.UNSPECIFIED).toBe(0);
      expect(GamePhase.SETUP).toBe(1);
      expect(GamePhase.DAY).toBe(2);
      expect(GamePhase.NIGHT).toBe(3);
      expect(GamePhase.VOTING).toBe(4);
      expect(GamePhase.FINISHED).toBe(5);
    });
  });

  describe('MVP game flow enums', () => {
    it('has death causes, night actions, and win reasons', () => {
      expect(DeathCause.EXECUTION).toBe(1);
      expect(NightActionType.KILL).toBe(3);
      expect(WinReason.STORYTELLER_DECISION).toBe(6);
    });
  });

  describe('ProtoPlayer interface', () => {
    it('can construct a valid player object', () => {
      const player: ProtoPlayer = {
        id: 'player-1',
        name: 'Alice',
        isAlive: true,
        votes: 0,
        poisonedUntil: 2,
      };
      expect(player.id).toBe('player-1');
      expect(player.isAlive).toBe(true);
      expect(player.poisonedUntil).toBe(2);
    });

    it('supports optional character', () => {
      const player: ProtoPlayer = {
        id: 'player-1',
        name: 'Alice',
        character: {
          id: 'washerwoman',
          name: 'Washerwoman',
          team: Team.GOOD,
          ability: 'You start knowing that one of two players is a particular Townsfolk.',
        },
        isAlive: true,
        votes: 0,
      };
      expect(player.character?.name).toBe('Washerwoman');
    });
  });

  describe('ProtoGameState interface', () => {
    it('can construct a valid game state', () => {
      const state: ProtoGameState = {
        id: 'game-1',
        phase: GamePhase.SETUP,
        players: [],
        dayNumber: 0,
        votes: {},
        deaths: [],
        nightActions: [],
      };
      expect(state.phase).toBe(GamePhase.SETUP);
    });

    it('supports nomination, deaths, night actions, and winner', () => {
      const state: ProtoGameState = {
        id: 'game-1',
        phase: GamePhase.FINISHED,
        players: [],
        dayNumber: 2,
        votes: {},
        deaths: [{ playerId: 'p2', cause: DeathCause.EXECUTION, dayNumber: 2 }],
        nomination: {
          nominatorId: 'p1',
          nomineeId: 'p2',
          votes: { p1: true },
          resolved: true,
        },
        nightActions: [
          {
            actorId: 'storyteller',
            actionType: NightActionType.KILL,
            targetIds: ['p3'],
          },
        ],
        winner: Team.GOOD,
      };
      expect(state.deaths[0]?.cause).toBe(DeathCause.EXECUTION);
      expect(state.nightActions[0]?.targetIds).toEqual(['p3']);
      expect(state.winner).toBe(Team.GOOD);
    });
  });

  describe('ProtoGameEvent type', () => {
    it('supports playerJoined event', () => {
      const event: ProtoGameEvent = {
        playerJoined: {
          player: { id: 'p1', name: 'Alice', isAlive: true, votes: 0 },
        },
      };
      expect(event.playerJoined).toBeDefined();
    });

    it('supports phaseChanged event', () => {
      const event: ProtoGameEvent = {
        phaseChanged: { phase: GamePhase.DAY },
      };
      expect(event.phaseChanged?.phase).toBe(GamePhase.DAY);
    });

    it('supports voteCast event', () => {
      const event: ProtoGameEvent = {
        voteCast: { voterId: 'p1', targetId: 'p2' },
      };
      expect(event.voteCast?.voterId).toBe('p1');
    });

    it('supports MVP game flow events', () => {
      const events: readonly ProtoGameEvent[] = [
        { playerDied: { playerId: 'p2', cause: DeathCause.EXECUTION, dayNumber: 1 } },
        { nominationStarted: { nominatorId: 'p1', nomineeId: 'p2' } },
        {
          nominationResolved: {
            nomineeId: 'p2',
            executed: true,
            yesVotes: 3,
            noVotes: 1,
            requiredVotes: 3,
          },
        },
        {
          nightActionSubmitted: {
            actorId: 'storyteller',
            actionType: NightActionType.KILL,
            targetIds: ['p3'],
          },
        },
        {
          gameEnded: {
            winner: Team.GOOD,
            reason: WinReason.IMP_EXECUTED,
            description: 'The Demon is dead.',
          },
        },
      ];

      expect(events).toHaveLength(5);
    });
  });
});
