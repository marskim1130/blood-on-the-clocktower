import { createStore, type StoreApi } from 'zustand/vanilla';
import type { GameState, GameEvent, GameId } from '../types/index.js';

interface GameStore {
  state: GameState;
  dispatch: (event: GameEvent) => void;
}

const defaultState: GameState = {
  id: '' as GameId,
  phase: 'setup',
  players: [],
  dayNumber: 0,
  storyteller: null,
  votes: new Map(),
  currentNomination: null,
  nightActions: [],
  deathRecords: new Map(),
  ghostVotesRemaining: new Set(),
  winner: null,
  winReason: null,
  winDescription: null,
};

function gameReducer(state: GameState, event: GameEvent): GameState {
  switch (event.type) {
    case 'PLAYER_JOINED':
      return {
        ...state,
        players: [...state.players, event.player],
      };

    case 'PLAYER_LEFT':
      return {
        ...state,
        players: state.players.filter((p) => p.id !== event.playerId),
      };

    case 'PHASE_CHANGED':
      return {
        ...state,
        phase: event.phase,
        dayNumber: event.phase === 'day' ? state.dayNumber + 1 : state.dayNumber,
        // Reset night actions when entering a new night
        nightActions: event.phase === 'night' ? [] : state.nightActions,
      };

    case 'VOTE_CAST': {
      const newVotes = new Map(state.votes);
      newVotes.set(event.voterId, event.targetId);
      return {
        ...state,
        votes: newVotes,
      };
    }

    case 'CHARACTER_ASSIGNED':
      return {
        ...state,
        players: state.players.map((p) =>
          p.id === event.playerId ? { ...p, character: event.character } : p
        ),
      };

    case 'PLAYER_DIED': {
      // Mark player as dead
      const newPlayers = state.players.map((p) =>
        p.id === event.playerId ? { ...p, isAlive: false } : p
      );

      // Track death record
      const newDeathRecords = new Map(state.deathRecords);
      newDeathRecords.set(event.playerId, {
        cause: event.cause,
        dayNumber: event.dayNumber,
      });

      // Grant ghost vote to newly dead player
      const newGhostVotes = new Set(state.ghostVotesRemaining);
      newGhostVotes.add(event.playerId);

      return {
        ...state,
        players: newPlayers,
        deathRecords: newDeathRecords,
        ghostVotesRemaining: newGhostVotes,
      };
    }

    case 'NOMINATION_STARTED': {
      return {
        ...state,
        currentNomination: {
          nominatorId: event.nominatorId,
          nomineeId: event.nomineeId,
          votes: new Map(),
        },
      };
    }

    case 'NOMINATION_RESOLVED': {
      if (!state.currentNomination) return state;

      // If the nominee was executed, mark them dead
      const newPlayers = event.executed
        ? state.players.map((p) =>
            p.id === event.nomineeId ? { ...p, isAlive: false } : p
          )
        : state.players;

      const newDeathRecords = event.executed
        ? new Map(state.deathRecords).set(event.nomineeId, {
            cause: 'execution' as const,
            dayNumber: state.dayNumber,
          })
        : state.deathRecords;

      const newGhostVotes = event.executed
        ? new Set(state.ghostVotesRemaining).add(event.nomineeId)
        : state.ghostVotesRemaining;

      return {
        ...state,
        currentNomination: null,
        players: newPlayers,
        deathRecords: newDeathRecords,
        ghostVotesRemaining: newGhostVotes,
      };
    }

    case 'NIGHT_ACTION': {
      return {
        ...state,
        nightActions: [
          ...state.nightActions,
          {
            actorId: event.actorId,
            actionType: event.actionType,
            targetIds: event.targetIds,
            result: event.result,
          },
        ],
      };
    }

    case 'GAME_OVER': {
      return {
        ...state,
        phase: 'finished',
        winner: event.winner,
        winReason: event.reason,
        winDescription: event.description,
        players: state.players.map((p) => {
          const revealed = event.revealedPlayers.find((r) => r.playerId === p.id);
          return revealed ? { ...p, character: revealed.character } : p;
        }),
      };
    }

    default:
      return state;
  }
}

export function createGameStore(initialState?: Partial<GameState>): StoreApi<GameStore> {
  return createStore<GameStore>()((set) => ({
    state: { ...defaultState, ...initialState } as GameState,
    dispatch: (event: GameEvent) =>
      set((store) => ({ state: gameReducer(store.state, event) })),
  }));
}

export type { GameStore };
