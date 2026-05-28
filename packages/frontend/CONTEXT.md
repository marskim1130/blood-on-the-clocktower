# @clocktower/frontend — Frontend Package Context

## Purpose

This package contains the Taro 4.x frontend application that runs on both WeChat Mini Program and React Native platforms.

## Key Concepts

### Cross-Platform

Uses Taro's component system to write once, run on:
- WeChat Mini Program (primary)
- React Native (secondary)

### State Management

Uses Zustand with the core package's state machine. The frontend:
- Creates the game store from `@clocktower/core`
- Connects to the WebSocket server
- Dispatches user actions as `GameEvent`s

### Component Architecture

- **Pages**: Route-level components (lobby, game, settings)
- **Components**: Reusable UI components (player list, voting panel, chat)
- **Hooks**: Custom hooks for game state, WebSocket connection

## Business Rules

1. UI must reflect game state changes in real-time via WebSocket
2. User actions must be validated client-side before sending to server
3. Optimistic updates for better UX, with rollback on server rejection

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
