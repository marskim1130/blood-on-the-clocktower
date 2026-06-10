import { describe, it, expect } from 'vitest';
import {
  Team,
  GamePhase,
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

  describe('ProtoPlayer interface', () => {
    it('can construct a valid player object', () => {
      const player: ProtoPlayer = {
        id: 'player-1',
        name: 'Alice',
        isAlive: true,
        votes: 0,
      };
      expect(player.id).toBe('player-1');
      expect(player.isAlive).toBe(true);
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
      };
      expect(state.phase).toBe(GamePhase.SETUP);
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
  });
});
