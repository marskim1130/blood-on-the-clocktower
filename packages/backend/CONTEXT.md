# @clocktower/backend — Backend Package Context

## Purpose

This package contains the Go backend server that manages game sessions, validates game rules, and facilitates real-time communication between players.

## Key Concepts

### Game Server

The backend:
- Manages game sessions (create, join, leave)
- Validates all game events against rules
- Broadcasts state changes to connected players
- Persists game state (database integration TBD)

### Communication

- **gRPC**: For type-safe API calls between frontend and backend
- **WebSocket**: For real-time game events (alternative to gRPC streaming)
- **ProtoBuf**: Single source of truth for data structures

### Type Synchronization

Go structs are generated from the same ProtoBuf definitions as TypeScript types. This ensures:
- Identical data structures on both sides
- Compile-time verification of type compatibility
- Automatic code generation from `proto/game.proto`

## Business Rules

1. Server is authoritative for all game state
2. All game events must be validated server-side before applying
3. Invalid events are rejected with descriptive error messages
4. Game state is persisted after each valid event

## Persistence

Production deployments should set `CLOCKTOWER_REDIS_URL` to enable Redis-backed snapshot persistence. The optional `CLOCKTOWER_REDIS_KEY` overrides the default snapshot key.

Local or single-node deployments can set `CLOCKTOWER_SNAPSHOT_PATH` to persist the same snapshot data to a JSON file. Redis takes precedence when both are set.

## Dependencies

- `google.golang.org/grpc` — gRPC framework
- `google.golang.org/protobuf` — ProtoBuf runtime

## Consumers

- Frontend clients via gRPC/WebSocket
- Potentially other services (analytics, matchmaking)

## API Design

### gRPC Services

```protobuf
service GameService {
  rpc CreateGame(CreateGameRequest) returns (GameState);
  rpc JoinGame(JoinGameRequest) returns (GameState);
  rpc SubmitEvent(GameEvent) returns (GameState);
  rpc StreamEvents(StreamEventsRequest) returns (stream GameEvent);
}
```

### WebSocket Protocol

Messages are JSON-encoded `GameEvent` objects. Connection lifecycle:
1. Client connects with game ID and player ID
2. Server sends current game state
3. Bidirectional event streaming begins
