# Blood on the Clocktower

Digital implementation of the Blood on the Clocktower social deduction game.

## Tech Stack

- **Frontend**: Taro 4.x (React) — WeChat Mini Program & React Native
- **State Management**: Zustand
- **Backend**: Go
- **Type Safety**: ProtoBuf for end-to-end type synchronization
- **Package Manager**: pnpm workspaces

## Project Structure

This is a monorepo with the following packages:

- `@clocktower/core` — Shared rules, WebSocket client, generated protocol types
- `@clocktower/frontend` — Taro 4.x frontend application
- `@clocktower/backend` — Go backend server

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues. See `docs/agents/issue-tracker.md`.

### Triage labels

Uses canonical labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Multi-context layout with per-package CONTEXT.md files. See `docs/agents/domain.md` and `CONTEXT-MAP.md`.
