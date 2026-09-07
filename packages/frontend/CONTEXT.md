# @clocktower/frontend — Frontend Package Context

## 2026-09-07 手机面杀版本补充

本轮实际交付与构建目标为 H5 和微信小程序，React Native 未验收。详见 [手机面杀验收清单](../../docs/playtest-checklist.md)：新增恢复审批、长按私密信息、说书人操作日志/撤销、成员管理与同房再开；无震动无音效。下面旧约束若仍写不支持转房主、重赛、历史或恢复码，以当前清单和用户新需求为准。界面必须等待服务器权威投影，不得乐观暴露私密身份。

## Purpose

This package contains the Taro 4.x frontend application that runs on both WeChat Mini Program and React Native platforms.

## Key Concepts

### Cross-Platform

Uses Taro's component system to write once, run on:
- WeChat Mini Program (primary)
- React Native (secondary)

### State Management

Uses an application-level Zustand Room Session Store. The frontend:
- Owns durable room identity and the single WebSocket client instance
- Replaces Room Experience state from accepted authoritative projections
- Sends user commands through `@clocktower/core` without maintaining a parallel game reducer

### Room Experience

`src/lib/room-experience.ts` owns the frontend Room Experience projection state. It atomically replaces the complete authoritative room projection, derives phase-specific view data, and clears all projection-derived data when identity is retained, kicked, invalid, or the room closes.

The page owns transport side effects and rendering only. It must not reconstruct committed room state from incremental events, maintain parallel nomination/death/night/winner stores, or repeat the Core Room Revision gate.

### Component Architecture

- **Pages**: Route-level components (lobby, game, settings)
- **Components**: Reusable UI components (player list, voting panel, chat)
- **Hooks**: Custom hooks for game state, WebSocket connection

## Business Rules

1. UI must reflect game state changes in real-time via WebSocket
2. User actions must be validated client-side before sending to server
3. Authoritative room mutations are applied only from committed server projections; local validation may improve UX but must not invent committed state
4. Room identity is stored as versioned `clocktower.roomIdentity.v2` data containing `roomId`, `playerId`, and Resume Credential
5. Legacy room storage without a Resume Credential is cleared and requires an explicit new join
6. Retained Identity can explicitly rejoin; kicked or closed identities are cleared
7. A Room Revision gap or regression triggers `GET_ROOM_STATE`, and the inconsistent incremental payload is not applied
8. Clearing or replacing a room projection must also clear stale nomination, death, night-action, phase, and winner UI state
9. Night Action history is rendered only from the Storyteller's complete authoritative projection; players never receive it

## Dependencies

- `@clocktower/core` — game logic and types
- `@tarojs/*` — Taro framework
- `react` — UI library
- `zustand` — state management

## Platform Considerations

### WeChat Mini Program
- Limited WebSocket support (single connection)
- No `WebSocket` constructor, use `Taro.connectSocket`
- File size limits apply

### React Native
- Standard WebSocket support
- Different navigation system
- Platform-specific styling may be needed
