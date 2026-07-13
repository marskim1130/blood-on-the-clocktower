# @clocktower/backend — Backend Package Context

## Purpose

This package contains the Go backend server that manages game sessions, validates game rules, and facilitates real-time communication between players.

## Key Concepts

### Session Ownership

**Authoritative Game Session**:
The sole owner of room membership, storyteller assignment, room settings, reconnect eligibility, and complete game state.
_Avoid_: Room state split across Hub, RoomManager, and GameSession

**Gameplay Session**:
The deep in-memory game rules module composed behind the Authoritative Game Session. It owns phases, character abilities, nominations, voting, night actions, deaths, win conditions, snapshots, and recipient-specific Gameplay Projections. It has no knowledge of credentials, Room Records, connections, or WebSocket messages.
_Avoid_: WebSocket handler, room lifecycle owner, one module per character

**Gameplay Projection**:
A privacy-safe game-only view for one recipient. The WebSocket adapter combines it with Authoritative Game Session metadata to produce the generated protocol `RoomState`.
_Avoid_: Room Record, protocol envelope, incremental event stream

**Room Member**:
A player identity retained by the Authoritative Game Session independently of network connectivity. Disconnecting does not remove membership.
_Avoid_: Connected client

**Active Connection**:
The single current network connection attached to a Room Member. A new connection using the same `playerId` replaces and invalidates the previous connection.
_Avoid_: Membership, session ownership

**Closed Room**:
An explicitly terminated Authoritative Game Session removed atomically from the session registry and persisted snapshot. A finished game is not a Closed Room.
_Avoid_: Finished game, empty room

**Room Creator**:
The persistent player identity that created the room and exclusively owns Close Room authority. The Room Creator may leave membership but cannot be kicked or have ownership transferred.
_Avoid_: Storyteller, currently connected host

**Resume Credential**:
A server-issued, unguessable secret proving control of a Room Member identity. Reconnection and Active Connection takeover require both the public `playerId` and its Resume Credential.
_Avoid_: Player ID, room code, display name

**Retained Identity**:
A former Room Member identity preserved after voluntary leave so the same `playerId` and Resume Credential can rejoin the room.
_Avoid_: Active member, disconnected member

**Room Ban**:
A room-scoped revocation created by Kick Player that prevents the same `playerId` from joining again. It lasts until the room is closed.
_Avoid_: Disconnect, voluntary leave

**Participant Set**:
The Room Members present when the game starts. It becomes permanently frozen for the remaining room lifecycle.
_Avoid_: Active connections, lobby members after start

**Detached Management Command**:
A credential-authenticated room command that does not require current membership or an Active Connection. Close Room is the only Detached Management Command.
_Avoid_: Game command, member command

**Room Record**:
The independently persisted representation of one Authoritative Game Session. It can be created, replaced, or deleted without rewriting another room's state.
_Avoid_: Global snapshot

**Create Request**:
An idempotent Create Room attempt identified by a client-generated, unguessable `requestId`. Retrying the same request returns the original room and creator credentials.
_Avoid_: Room ID, connection ID

**Client Sequence**:
A monotonically increasing number scoped to one Room Member identity and used to order and deduplicate its commands. The server persists the last accepted sequence and command identity.
_Avoid_: Event index, room revision, connection counter

**Join Request**:
An idempotent first-time Join Room attempt identified by a client-generated, unguessable `joinRequestId`. It creates one new player identity and returns its Resume Credential and initial Client Sequence.
_Avoid_: Reconnect, retained identity rejoin

**Room ID**:
A long, unguessable persistent identifier for one room lifecycle. It is never intentionally reused after Close Room.
_Avoid_: Short invite code, display name

**Room Revision**:
A monotonically increasing number for successfully committed Room Record replacements. It detects stale writes and supports future coordination but does not itself provide multi-instance safety.
_Avoid_: Client Sequence, event count

### Game Server

The backend:
- Manages game sessions (create, join, leave)
- Validates all game events against rules
- Broadcasts state changes to connected players
- Persists game state (database integration TBD)

### Communication

- **WebSocket v2**: The only supported room transport protocol; Hub is a thin transport adapter over Authoritative Game Session
- **JSON Wire Format**: Runtime WebSocket messages use JSON encoding
- **ProtoBuf Contract**: `proto/game.proto` generates the Go and TypeScript message contracts

### Type Synchronization

WebSocket contract structs and constants are generated from the same ProtoBuf definitions as TypeScript types. This ensures:
- Identical data structures on both sides
- Compile-time verification of type compatibility
- Automatic code generation from `proto/game.proto`

## Business Rules

1. Server is authoritative for all game state
2. All game events must be validated server-side before applying
3. Invalid events are rejected with descriptive error messages
4. Game state is persisted after each valid event
5. Disconnecting an Active Connection does not remove its Room Member
6. A Room Member can reconnect using the same `playerId`
7. Each Room Member has at most one Active Connection; a new connection takes over and closes the previous connection
8. Only an explicit leave or kick removes membership or reconnect eligibility
9. A command succeeds only after its resulting Authoritative Game Session state is persisted
10. Command application, snapshot persistence, and in-memory publication form a transactional commit: persistence failure restores the previous in-memory state
11. State changes are broadcast only after persistence succeeds; clients must never observe unpersisted state
12. Broadcast failure after a successful commit does not fail or roll back the command
13. A connection that fails during broadcast is invalidated and closed without removing its Room Member
14. Reconnection restores the latest full persisted state; missed events are not replayed
15. Commands for the same Authoritative Game Session are processed serially in server receive order
16. A command reads the state produced by the previous successfully persisted command; persistence failure leaves the next command the previous committed state
17. Different Authoritative Game Sessions may process commands concurrently
18. Broadcast uses an immutable payload produced by the committed state and does not block the next command
19. Join, leave, kick, room settings, storyteller assignment, and all game commands share the same per-session serial command sequence
20. Active Connection takeover is a transport-only atomic operation: it replaces the connection without changing or persisting Authoritative Game Session state
21. Finishing a game does not close or delete its Authoritative Game Session; members may reconnect and inspect the final state
22. A room is deleted only by an explicit Close Room command, which atomically removes it from the session registry and persisted snapshot
23. An empty room is retained; leaving the last member does not implicitly close it
24. Automatic time-to-live cleanup is outside the current room lifecycle
25. Only the Room Creator may execute Close Room; Storyteller assignment does not grant close authority
26. The Room Creator cannot be kicked and retains close authority after leaving or disconnecting
27. Room Creator ownership cannot be transferred, and administrator recovery is outside the current lifecycle
28. The server issues a Resume Credential when a Room Member identity is first created
29. Reconnection and Active Connection takeover require a matching `playerId` and Resume Credential
30. Resume Credentials are persisted with the Authoritative Game Session, returned only to their owning client, and never broadcast to other members
31. Losing a Resume Credential means losing control of that Room Member identity; credential recovery is outside the current lifecycle
32. A Resume Credential remains stable for the lifetime of its Room Member identity and is not automatically rotated on reconnect or connection takeover
33. Explicit credential rotation and revocation are outside the current lifecycle
34. If a Join Room request uses an existing `playerId`, it is accepted only as reconnect or connection takeover with the matching Resume Credential
35. A missing or incorrect Resume Credential for an existing `playerId` is rejected with the same generic credential error and must not reveal whether the member exists
36. The server does not replace or generate a different `playerId` for a rejected Join Room request; a new member must present an unused `playerId`
37. Voluntary Leave Room removes membership but preserves a Retained Identity; the same `playerId` and Resume Credential may rejoin
38. Kick Player removes membership, revokes the Resume Credential, and creates a Room Ban for that `playerId`
39. A Room Ban cannot be removed during the current room lifecycle and is deleted only when the room is closed
40. Closing a room deletes all Room Members, Retained Identities, Resume Credentials, and Room Bans
41. Starting the game freezes the Participant Set for the remaining room lifecycle, including after the game finishes
42. After the Participant Set is frozen, new members and Retained Identities cannot join, and Leave Room or Kick Player cannot remove participants
43. Existing participants may disconnect, reconnect, or perform Active Connection takeover without changing the Participant Set
44. A Room Creator who has voluntarily left may execute Close Room as a Detached Management Command using `roomId`, `playerId`, and the matching Resume Credential
45. Detached Close Room does not restore membership and is the only command available to a Room Creator who is not a current Room Member
46. Detached Close Room enters the room's serial command sequence and transactionally removes the room from persistence
47. Each Authoritative Game Session is persisted as an independent Room Record
48. Create Room creates one Room Record, each successful command atomically replaces that record, and Close Room deletes it
49. Server startup enumerates and restores all Room Records
50. Persistence of one Room Record must not require rewriting another room's state; different rooms may persist concurrently
51. Redis uses a distinct key per Room Record; a File adapter may share a physical file only if it preserves atomic per-room update semantics
52. `CLOCKTOWER_SNAPSHOT_PATH` identifies a directory containing one `<roomId>.json` file per Room Record
53. The File adapter replaces a Room Record by writing and flushing a temporary file in the same directory, then atomically renaming it over the room file
54. Close Room deletes its room file, and startup restores rooms by scanning the snapshot directory
55. A corrupt room file is logged and skipped without preventing other Room Records from loading
56. Automatic migration from the previous single-file snapshot format is outside the current persistence lifecycle
57. Every Room Record contains an explicit `schemaVersion`, beginning with version 1
58. The server loads only supported schema versions; an unknown newer version is logged and skipped without inferred conversion
59. Older supported versions are upgraded through explicit stepwise migration functions
60. Migration occurs in memory, and the latest schema version is persisted on the room's next successful command
61. The current implementation establishes the versioning framework but does not require a legacy Room Record migration
62. Create Room requires a client-generated, unguessable `requestId` and is idempotent for that identifier
63. A successful Create Request persists its parameters and resulting `roomId`, Room Creator identity, and Resume Credential before responding
64. Retrying the same `requestId` with identical parameters returns the original result without creating another room
65. Reusing a `requestId` with different parameters is rejected
66. The Create Request record is deleted when its room is closed
67. Every Room Member command carries a `clientSequence` scoped to that member identity
68. The next new command must use the server's last accepted sequence plus one
69. Repeating the last accepted sequence with the same command content does not execute or broadcast again and returns the current committed state
70. Reusing an accepted sequence with different command content is rejected as a sequence conflict
71. Older or skipped sequences are rejected with the next sequence expected by the server
72. The last accepted Client Sequence and command identity are persisted in the Room Record, and reconnect responses include the next expected sequence
73. Client Sequence belongs to the persistent player identity, not its current membership or Active Connection
74. Voluntary leave preserves the last accepted Client Sequence in the Retained Identity, and rejoining continues with the next expected sequence
75. Detached Close Room uses the Room Creator identity's next Client Sequence
76. Reconnect and Active Connection takeover do not consume a Client Sequence because they do not change Authoritative Game Session state
77. A newly created player identity receives its initial next Client Sequence from the server
78. First-time Join Room requires a client-generated, unguessable `joinRequestId` and is idempotent for that identifier
79. A successful Join Request persists its parameters, player identity, Resume Credential, and initial next Client Sequence before responding
80. Retrying the same `joinRequestId` with identical parameters returns the original result without creating another identity
81. Reusing a `joinRequestId` with different parameters is rejected
82. After identity creation, reconnect and Retained Identity rejoin use the player identity rules rather than the Join Request
83. Join Request records are deleted when the room is closed
84. Retained Identity rejoin is a state-changing member command carrying `playerId`, Resume Credential, and the identity's next Client Sequence
85. A successful Retained Identity rejoin restores Room Member status and persists the accepted Client Sequence in the same transactional commit
86. Retrying the same accepted sequence with the same rejoin command returns the current committed state without another broadcast
87. Network reconnect for an existing Room Member remains a transport operation and does not consume a Client Sequence
88. Leave Room is a state-changing member command that consumes the identity's next Client Sequence
89. After Leave Room commits, the member becomes a Retained Identity and its Active Connection is closed
90. Authentication after voluntary leave returns Retained Identity status and the next expected Client Sequence without restoring membership
91. Membership is restored only by an explicit Retained Identity rejoin command
92. Retrying an accepted Leave Room sequence returns the committed left state without executing again
93. Kick Player transactionally removes the target membership, revokes its Resume Credential, creates its Room Ban, and advances the sender's Client Sequence before any notification
94. After Kick Player commits, the server best-effort sends a `KICKED` notification, then closes and removes the target Active Connection regardless of notification success
95. A kicked identity presenting its former credential receives the generic credential error without disclosure of Room Ban status
96. Broadcast state after a kick contains no Resume Credential or Room Ban details
97. Close Room validates the Room Creator credential and next Client Sequence within the room's serial command sequence
98. Close Room succeeds only after the Room Record and its Create Request record are durably deleted; deletion failure leaves the room open and unchanged
99. After durable deletion, the room is removed from the in-memory session registry
100. Close Room uses a pre-removal immutable connection list to best-effort send `ROOM_CLOSED`, then closes all Active Connections
101. Requests made after closure return a generic room-not-found error without historical membership disclosure
102. Room IDs are long, unguessable identifiers and are never intentionally reused after Close Room
103. Create Room conditionally creates its Room Record only if the generated Room ID is absent; collision generates a new identifier and retries
104. Closed Room tombstones are not retained because Room ID entropy makes accidental reuse negligible
105. A future human-friendly invite code must be a separate rotatable alias and must not serve as the persistent Room ID
106. The current deployment model supports exactly one active backend instance
107. Redis provides Room Record persistence only and does not provide distributed command ordering, locking, or failover coordination
108. Every successful Room Record replacement increments its Room Revision for stale-write detection and future coordination
109. Multi-instance deployment, room sharding, distributed leases, and active failover are outside the current architecture
110. Documentation and startup logging must state the Single Active Backend Instance restriction
111. Every full room state, successful command acknowledgement, and committed broadcast includes the resulting Room Revision
112. A client that observes a Room Revision gap or regression requests a full room state instead of inferring missed events
113. A deduplicated command response returns the current latest Room Revision
114. Room Revision represents persisted commit order only and is distinct from game phase, event count, and Client Sequence
115. Backend protocol support for Room Revision is in scope; automatic frontend resynchronization may be delivered as a later vertical slice
116. A Real-time Game Session Projection is a privacy-safe, recipient-specific complete room state generated from exactly one committed view
117. Projection Delivery begins only after persistence and atomic publication; delivery failure never changes the committed command result
118. Join and command commits generate all recipient projections while holding the room command sequence, so a projection batch cannot mix revisions
119. Projection metadata, game state, Room Revision, and next Client Sequence must come from the same committed view; Hub must not reconstruct metadata from a later view
120. Every active connection has an Ordered Outbound Queue; messages for one connection preserve enqueue order while different connections drain independently
121. A slow or failed consumer is disconnected without removing Room Membership; a bounded queue overflow follows the same rule
122. Connection takeover discards the old connection's pending outbound queue before closing it
123. Spy visibility and finished-game character reveal do not grant storyteller-only poisoning or night-management visibility
124. During Night, the full Night Action history is included only in the Storyteller projection and is omitted from every player projection
125. Hub accepts only protocol v2 and must not own a parallel room model, legacy protocol handler, or global snapshot lifecycle
126. Gameplay Session lives in `internal/gameplay`; `internal/ws` must not own game rules
127. Gameplay depends only on game domain types and must not import `internal/ws` or `internal/session`
128. WebSocket session engine is the adapter from Gameplay Projection to the generated protocol Room State
129. Role interactions remain inside one Gameplay Session module; character-specific public interfaces are not introduced without a real second adapter

## Persistence

Production startup requires a persistent `CLOCKTOWER_CREDENTIAL_KEY` of at least 32 bytes. Redis persistence uses `CLOCKTOWER_REDIS_URL` and treats `CLOCKTOWER_REDIS_KEY` as a namespace prefix. File persistence treats `CLOCKTOWER_SNAPSHOT_PATH` as a directory containing one `<roomId>.json` Room Record. Redis takes precedence when both are configured; without either adapter, room state is in memory only.

The backend supports one active instance. Legacy global Snapshot files are not migrated automatically. A Room Record revision conflict makes `/health` return `503`. See `packages/backend/README.md`.

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

Messages use the breaking JSON WebSocket protocol v2. The Hub is an adapter for protocol validation, active connection generations, session calls, and delivery. Membership, credentials, sequences, revisions, and persistence live in `AuthoritativeGameSession`; game rules and recipient-specific game views live in `GameplaySession`. See `docs/protocol/websocket-v2.md`.
