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
