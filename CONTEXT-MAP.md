# Context Map

This document maps the multi-context domain documentation for the Blood on the Clocktower project.

## Project Structure

This is a monorepo with the following sub-projects:

| Package | Context | ADRs | Description |
|---------|---------|------|-------------|
| `@clocktower/core` | [packages/core/CONTEXT.md](./packages/core/CONTEXT.md) | [packages/core/docs/adr/](./packages/core/docs/adr/) | Game state machine, WebSocket client, shared types |
| `@clocktower/frontend` | [packages/frontend/CONTEXT.md](./packages/frontend/CONTEXT.md) | [packages/frontend/docs/adr/](./packages/frontend/docs/adr/) | Taro 4.x (React) frontend for WeApp and RN |
| `@clocktower/backend` | [packages/backend/CONTEXT.md](./packages/backend/CONTEXT.md) | [packages/backend/docs/adr/](./packages/backend/docs/adr/) | Go backend server |

## Cross-Cutting Concerns

- **Type Safety**: End-to-end type safety between Go backend and TypeScript frontend via ProtoBuf
- **State Management**: Zustand for frontend state, game state machine in core package
- **Communication**: WebSocket for real-time game updates

## Reading Order

When working on a specific area:
1. Start with this file to understand the overall structure
2. Read the relevant sub-project's `CONTEXT.md` for domain language
3. Check `docs/adr/` for past architectural decisions
4. Cross-reference other sub-projects if changes span multiple packages
