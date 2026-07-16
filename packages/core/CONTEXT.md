# @clocktower/core — Core Package Context

## Purpose

This package contains shared pure rule modules, Trouble Brewing setup data, the protocol v2 WebSocket client, and generated TypeScript protocol types. The backend remains authoritative for committed room and game state.

## Key Concepts

### Shared Rule Modules

The vote, death, win-condition, night-phase, and script-setup modules are pure TypeScript rule helpers. They do not own committed room state and are not a client-side replica of the backend Gameplay Session.

### Authoritative Projection

The backend Authoritative Game Session owns committed state. The Frontend stores the latest accepted recipient-specific `RoomState`; Core does not maintain a parallel `GameState` or event reducer.

### Branded Types

We use TypeScript branded types for type safety:
- `PlayerId` — unique player identifier

This prevents accidentally passing a regular string where a player identifier is expected inside the pure rule modules.

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
- A projection gate that strips duplicate, stale, or gapped room projections before application handlers run
- At most one automatic full-state request while a revision resynchronization is in flight
- Client and server envelope types generated from `proto/game.proto`

The projection gate owns Room Revision filtering. Downstream consumers receive only accepted complete projections and must not implement a second revision gate or reconstruct state from incremental domain events.

The Core client keeps identity in memory. Durable Resume Credential storage belongs to the Frontend adapter.

## Business Rules

1. Shared rule modules are pure and do not mutate committed room state
2. Character assignments are immutable once confirmed
3. Protocol projections are accepted only through the WebSocket client's revision gate
4. The Frontend must not reconstruct committed state from incremental events

## Dependencies

Core has no runtime package dependencies. Protocol generation tooling is owned by the workspace root.

## Consumers

- `@clocktower/frontend` — uses script setup, generated types, and the WebSocket client
- Backend protocol generation — shares `proto/game.proto` as the canonical contract source
