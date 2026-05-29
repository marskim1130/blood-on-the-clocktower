# Work Log

## 2026-05-28 14:30 — 项目初始化

### 问题
需要搭建一个支持多端（微信小程序 + React Native）的 Monorepo 仓库，实现端到端类型安全。

### 解决方案
使用 pnpm workspaces + ProtoBuf 实现 Go 后端与 TypeScript 前端的类型同步。

### 修改文件

**根目录配置：**
- `pnpm-workspace.yaml` — pnpm 工作区配置
- `package.json` — 根项目配置
- `tsconfig.base.json` — 基础 TypeScript 配置（Matt Pocock 风格严格模式）
- `tsconfig.json` — 项目引用配置
- `.gitignore` — Git 忽略规则
- `CLAUDE.md` — Agent Skills 配置
- `CONTEXT-MAP.md` — 多上下文文档索引

**Agent Skills 配置：**
- `docs/agents/issue-tracker.md` — GitHub Issues 配置
- `docs/agents/triage-labels.md` — Triage 标签映射
- `docs/agents/domain.md` — 领域文档布局

**@clocktower/core 包：**
- `packages/core/package.json` — 包配置
- `packages/core/tsconfig.json` — TypeScript 配置
- `packages/core/src/index.ts` — 入口文件
- `packages/core/src/types/index.ts` — 游戏类型定义（含 Branded Types）
- `packages/core/src/state-machine/index.ts` — Zustand 游戏状态机
- `packages/core/src/websocket/index.ts` — WebSocket 客户端
- `packages/core/CONTEXT.md` — 领域上下文文档

**@clocktower/frontend 包：**
- `packages/frontend/package.json` — Taro 4.x 配置
- `packages/frontend/tsconfig.json` — TypeScript 配置
- `packages/frontend/CONTEXT.md` — 领域上下文文档

**@clocktower/backend 包：**
- `packages/backend/go.mod` — Go 模块配置
- `packages/backend/cmd/server/main.go` — gRPC 服务器入口
- `packages/backend/internal/game/game.go` — Go 游戏类型定义
- `packages/backend/CONTEXT.md` — 领域上下文文档

**ProtoBuf 定义：**
- `proto/game.proto` — 数据结构定义（单一事实来源）
- `proto/README.md` — 使用说明
- `scripts/generate-types.mjs` — TypeScript 类型生成脚本

### 撤回方式
```bash
git revert 17fe71b
```

---

## 2026-05-28 14:35 — 修复 Taro React Native 依赖问题

### 问题
`pnpm install` 报错：`@tarojs/plugin-platform-rn` 返回 404 Not Found。

### 根因
Taro 4.x 的 React Native 支持不使用 `@tarojs/plugin-platform-rn` 包。正确的包名是：
- `@tarojs/taro-rn` — 核心框架
- `@tarojs/runtime-rn` — 运行时
- `@tarojs/router-rn` — 路由

### 解决方案
更新 `packages/frontend/package.json`，移除不存在的包，添加正确的 RN 包。

### 修改文件
- `packages/frontend/package.json` — 替换 React Native 依赖

### 撤回方式
```bash
git revert HEAD
```

---

## 2026-05-28 14:45 — 创建 MVP PRD

### 问题
需要明确 MVP 范围和实现计划，为开发提供清晰指引。

### 解决方案
基于需求访谈，创建了 Human Storyteller 模式的 MVP PRD，包含：
- 35 个用户故事
- 9 个核心模块划分
- 数据模型和 API 契约
- 测试策略

### 关键决策
- 游戏模式：混合模式（人类 Storyteller 优先）
- 自动化级别：最大自动化
- 角色版本：Trouble Brewing only
- 玩家规模：5-15 人
- 并发目标：20-30 局（2c2g 服务器）
- 持久化：Redis
- 主平台：微信小程序
- 用户系统：匿名房间制
- 测试范围：Core Game Logic + Backend Services

### 修改文件
- `docs/prd/mvp-human-storyteller.md` — MVP PRD 文档

### 撤回方式
```bash
git revert HEAD
```

---

## 2026-05-28 15:00 — 创建 GitHub Issue

### 操作
使用 gh CLI 创建 PRD issue 并设置 triage labels。

### 结果
- Issue #1: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/1
- Labels: `ready-for-agent`

---

## 2026-05-29 10:00 — 拆解 PRD 为垂直切片

### 操作
将 PRD #1 拆解为 9 个垂直切片 issues，覆盖 35 个用户故事。

### Issues 创建

| Issue | 标题 | 阻塞 |
|-------|------|------|
| #2 | 基础设施：ProtoBuf + WebSocket 基础 | 无 |
| #3 | 房间管理：创建/加入/销毁 | #2 |
| #4 | Storyteller 指定 | #3 |
| #5 | 角色分配：Trouble Brewing 角色集 | #4 |
| #6 | 白天阶段：玩家列表 + 讨论 | #5 |
| #7 | 提名系统 | #6 |
| #8 | 投票系统 | #7 |
| #9 | 处决 + 死亡宣告 | #8 |
| #10 | 夜间阶段 + 胜负判定 | #9 |

### 链接
- Issue #2: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/2
- Issue #3: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/3
- Issue #4: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/4
- Issue #5: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/5
- Issue #6: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/6
- Issue #7: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/7
- Issue #8: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/8
- Issue #9: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/9
- Issue #10: https://github.com/CodeApeKQ/blood-on-the-clocktower/issues/10

---

## 2026-05-29 14:30 — 实现 Issue #2：ProtoBuf + WebSocket 基础

### 问题
需要建立项目的基础设施：ProtoBuf 类型同步流水线和 Go ↔ TypeScript WebSocket 通信。

### 解决方案
TDD 方式实现，先写测试再写代码。

### 修改文件

**依赖更新：**
- `package.json` — 添加 protobufjs、protobufjs-cli、ws、@types/ws 依赖

**Go WebSocket 服务器：**
- `packages/backend/internal/ws/message.go` — 消息类型定义（ClientMessage、ServerMessage、RoomState）
- `packages/backend/internal/ws/hub.go` — WebSocket Hub（连接管理、房间路由、消息广播）
- `packages/backend/internal/ws/ws_test.go` — 5 个测试：连接、加入房间、广播、房间隔离、端到端事件流
- `packages/backend/cmd/server/main.go` — 改用 HTTP + WebSocket（替代 gRPC 骨架）

**TypeScript WebSocket 客户端：**
- `packages/core/src/websocket/index.ts` — 重写客户端，匹配服务器协议（ClientMessage/ServerMessage）
- `packages/core/src/websocket/__tests__/websocket-client.test.ts` — 3 个测试：连接状态、加入房间、断开发送

**ProtoBuf 生成验证：**
- `packages/core/src/types/generated/__tests__/proto-types.test.ts` — 8 个测试：Team/GamePhase 枚举、Player/GameState/GameEvent 接口

### 测试结果
- TypeScript: 11 个测试通过（2 个测试文件）
- Go: 5 个测试通过（1 个测试文件）

### 撤回方式
```bash
git revert HEAD
```

---

## 2026-05-29 14:50 — 实现 Issue #3：房间管理

### 问题
需要实现房间的完整生命周期：创建、加入、离开、销毁，以及人数限制。

### 解决方案
TDD 方式，在 Issue #2 的 Hub 基础上扩展房间管理功能。

### 修改文件

**Go WebSocket 服务器：**
- `packages/backend/internal/ws/message.go` — 添加 MaxPlayers、Error 字段
- `packages/backend/internal/ws/hub.go` — 添加 CREATE_ROOM、LEAVE_ROOM 处理器，房间验证，自动销毁
- `packages/backend/internal/ws/ws_test.go` — 10 个测试（5 个原有 + 5 个新增）

**新增测试：**
- TestCreateRoomGeneratesUniqueID — 创建房间生成 6 位唯一 ID
- TestJoinNonExistentRoomReturnsError — 加入不存在房间返回错误
- TestJoinFullRoomReturnsError — 房间满员拒绝加入
- TestLeaveRoomRemovesPlayer — 玩家离开广播 PLAYER_LEFT
- TestRoomAutoDestroysWhenEmpty — 最后玩家离开后房间销毁

### 测试结果
- Go: 10 个测试通过
- TypeScript: 11 个测试通过

### 撤回方式
```bash
git revert HEAD
```

---

## 2026-05-29 16:00 — 实现 Issue #4：Storyteller 指定

### 问题
需要实现 Storyteller 角色的指定和权限验证。

### 解决方案
TDD 方式，在 Room 中添加 creatorID 和 storytellerID 字段，实现 SET_STORYTELLER 处理器。

### 修改文件

**Go WebSocket 服务器：**
- `packages/backend/internal/ws/message.go` — 添加 TargetPlayerID、StorytellerID 字段
- `packages/backend/internal/ws/hub.go` — 添加 SET_STORYTELLER 处理器（权限验证、移除玩家列表、广播）
- `packages/backend/internal/ws/ws_test.go` — 14 个测试（10 个原有 + 4 个新增）

**新增测试：**
- TestSetStoryteller — 指定 Storyteller 并广播
- TestStorytellerRemovedFromPlayerList — Storyteller 从玩家列表移除
- TestOnlyOneStoryteller — 只能有一个 Storyteller
- TestNonStorytellerCannotPerformStorytellerActions — 非创建者无法指定

### 测试结果
- Go: 14 个测试通过
- TypeScript: 11 个测试通过

### 撤回方式
```bash
git revert HEAD
```

---

## 2026-05-29 16:15 — 实现 Issue #5：角色分配

### 问题
需要实现 Trouble Brewing 角色集和 Storyteller 角色分配功能。

### 解决方案
TDD 方式，先定义角色数据和验证逻辑，再实现 WebSocket 处理器。

### 修改文件

**角色系统：**
- `packages/backend/internal/game/characters.go` — 22 个 Trouble Brewing 角色定义、角色分配验证
- `packages/backend/internal/game/characters_test.go` — 6 个测试

**WebSocket 处理器：**
- `packages/backend/internal/ws/message.go` — 添加 Assignments 字段
- `packages/backend/internal/ws/hub.go` — 添加 ASSIGN_CHARACTERS 处理器
- `packages/backend/internal/ws/ws_test.go` — 17 个测试（14 个原有 + 3 个新增）

**新增测试：**
- TestStorytellerCanAssignCharacters — Storyteller 分配角色并广播
- TestInvalidAssignmentRejected — 非法分配被拒绝
- TestNonStorytellerCannotAssignCharacters — 非 Storyteller 无法分配

### 测试结果
- Go: 23 个测试通过（6 game + 17 ws）
- TypeScript: 11 个测试通过

### 撤回方式
```bash
git revert HEAD
```
