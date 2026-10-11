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

### 恢复状态的展示与持久化

房间会话存储通过现有消息订阅读取 Core 的恢复状态与可重试标记，不能从原始恢复通知或普通错误自行推断批准、拒绝、过期。页面只在 Core 允许时显示“重试恢复申请”，连接可用后由用户触发。

- 首次发送前保存仅含申请编号、房间及玩家标识的非秘密票据；恢复码只留在内存。刷新后重新粘贴恢复码并复用票据。
- 批准后立即保存 Core 已认可的完整身份，再移除同房间、同玩家的票据；刷新后成功恢复并保存身份时也补清遗留票据，不清理其它身份的票据。批准后的入房失败不退回待审批。拒绝、过期或身份失效清理相应票据。
- 丢弃客户端或切换开发地址时清理内存恢复视图，保留非秘密票据供重新粘贴。
- 设备存储异常不能阻断消息处理；界面明确提示未保存，当前页面继续使用已认可的内存身份。批准身份未能保存时保留原票据，不声称持久化成功。
- 恢复行为通过真实 Core 与房间会话存储的整合测试验证，传输和存储使用测试适配器；页面测试负责按钮与文案的交互合同。

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
