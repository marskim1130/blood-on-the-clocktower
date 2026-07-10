# @clocktower/core — Core Package Context

## Purpose

This package contains the shared game logic that is used by both the frontend (Taro) and potentially validated by the backend (Go). It is the single source of truth for game rules and state management on the client side.

## Key Concepts

### Game State Machine

The game progresses through phases:
- **Setup**: Players join, characters are assigned
- **Day**: Players discuss and nominate
- **Voting**: Players vote on nominations
- **Night**: Characters use abilities
- **Finished**: Game ends with a winner

### Branded Types

We use TypeScript branded types for type safety:
- `PlayerId` — unique player identifier
- `GameId` — unique game identifier

These prevent accidentally passing a regular string where a specific ID is expected.

### State Management

Uses Zustand with vanilla store for framework-agnostic state management. The store:
- Holds the complete `GameState`
- Provides a `dispatch` method for `GameEvent`s
- Uses immutable updates for predictable state changes

### WebSocket Client

The `GameWebSocketClient` provides:
- Automatic reconnection with configurable attempts
- Type-safe event handling
- Connection status tracking
- Protocol v2 message envelopes and stable room identity metadata
- Automatic `RESUME_ROOM` after transport reconnection
- One in-flight sequenced state command at a time
- Client sequence advancement only after authoritative acknowledgement
- Room revision tracking for full-state resynchronization

The Core client keeps identity in memory. Durable Resume Credential storage belongs to the Frontend adapter.

## Business Rules

1. A player can only vote once per voting phase
2. Dead players cannot vote
3. Character assignments are immutable once confirmed
4. Phase transitions follow the sequence: setup -> day -> voting -> night -> day...

## Dependencies

- `zustand` — state management

## Consumers

- `@clocktower/frontend` — uses state machine and WebSocket client
- Backend validation — should match the same rules (enforced via ProtoBuf types)
