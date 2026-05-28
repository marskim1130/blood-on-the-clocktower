# ProtoBuf Definitions

This directory contains Protocol Buffer definitions that serve as the single source of truth for data structures shared between the Go backend and TypeScript frontend.

## Usage

### Generate TypeScript types

```bash
pnpm proto:generate
```

This will:
1. Parse `game.proto`
2. Generate TypeScript interfaces in `packages/core/src/types/generated/`
3. Create both the raw protobufjs bundle and clean TypeScript interfaces

### Generate Go types

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/game.proto
```

## Type Safety

The generated TypeScript types use `readonly` modifiers and strict typing to match the Go structs exactly. Any changes to the proto definitions require regenerating types on both sides.

## Adding New Types

1. Add the message/enum to `game.proto`
2. Run `pnpm proto:generate`
3. Update the Go code to use the new types
4. Update the TypeScript code to use the new generated types
