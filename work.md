# Work Log

## 2026-05-30 15:10 — 前端状态机第一阶段初始化
### 问题
需要引导用户学习前端的 Branded Types 模式以及 Zustand 状态管理设计。
### 解决方案
1. 切换工作区至第一个 Git 提交（哈希 `17fe71b9031fc051176aa78775ca5b8aab8e8536`）。
2. 在根目录下重新建立教学审计日志，开启 TypeScript 前端核心库的第一阶段教学。
### 修改文件
- `work.md` (重建)
### 撤回方式
```bash
git checkout main
```

---

## 2026-05-30 15:11 — 修复 state-machine/index.ts 的 GameId 漏导 Bug
### 问题
Zustand 初始化状态机时对 id 字段使用了 `'' as GameId` 类型转换，但第 2 行的 import 语句漏导了 `GameId` 类型，导致 TypeScript 编译器报错 `Cannot find name 'GameId'`。
### 解决方案
在 `packages/core/src/state-machine/index.ts` 导入声明中加上 `GameId` 类型。
### 修改文件
- `packages/core/src/state-machine/index.ts`
### 撤回方式
```bash
git checkout -- packages/core/src/state-machine/index.ts
```

---

## 2026-05-30 15:56 — 前端状态机 TDD 红（RED）阶段初始化
### 问题
需要实战测试 TypeScript 类型与 Zustand 状态机，根据 TDD 规范，需先建立失败断言。
### 解决方案
在 `packages/core/src/state-machine/__tests__` 下创建了 `state_syntax.test.ts`，并写入一个必定断言失败的测试用例。
### 修改文件
- `packages/core/src/state-machine/__tests__/state_syntax.test.ts`
### 撤回方式
```bash
rm packages/core/src/state-machine/__tests__/state_syntax.test.ts
```

---

## 2026-05-30 16:15 — 前端状态机 TDD 绿（GREEN）阶段测试通过
### 问题
第一阶段的测试用例已完成编写，需要验证 Branded Types 的编译防护与 Zustand 状态机的状态转移正确性，并完成第一阶段的测试闭环。
### 解决方案
1. 用户在本地执行 `pnpm --filter @clocktower/core test`。
2. 单元测试全部通过，验证了 `PLAYER_JOINED`、`VOTE_CAST`、`PHASE_CHANGED` 事件的状态转移及 Branded Types 的静态类型安全。
### 修改文件
- `packages/core/src/state-machine/__tests__/state_syntax.test.ts` (绿阶段已充实并确认通过)
### 撤回方式
```bash
git checkout -- packages/core/src/state-machine/__tests__/state_syntax.test.ts
```

---

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

---

## 2026-05-29 17:30 — 架构重构：分解 Hub 单体 + 提取 Connection 接口

### 问题
`hub.go` 是 426 行的单体模块，混合了 WebSocket 传输、房间生命周期和游戏逻辑。测试需要真实的 HTTP + WebSocket 连接。无法独立测试房间逻辑。

### 解决方案
按架构审查建议，将 Hub 分解为三个深度模块：

1. **Connection 接口**（`conn.go`）— 传输层接缝，两个适配器：wsConn（生产）+ FakeConnection（测试）
2. **Broadcaster 接口**（`broadcaster.go`）— 广播职责从 Room 移到 Hub
3. **RoomManager**（`room_manager.go`）— 房间生命周期、客户端准入，无 GameState
4. **GameSession**（`game_session.go`）— 游戏状态、命令处理，无房间/连接知识
5. **Hub**（`hub.go`）— 薄路由器，实现 Broadcaster

### 关键设计决策
- `Client` 结构体（`Conn Connection` + `PlayerID string`），支持断线重连
- `storytellerID` 仅由 GameSession 持有，避免数据不一致
- `GameSession.Apply(cmd)` 使用 `sync.Mutex` 保证并发安全
- 命令模式：`SetStorytellerCmd`、`AssignCharactersCmd`、`SubmitEventCmd`

### 修改文件

**新建：**
- `internal/ws/conn.go` — Connection 接口 + Client + wsConn + FakeConnection
- `internal/ws/broadcaster.go` — Broadcaster 接口
- `internal/ws/room_manager.go` — Room + RoomManager
- `internal/ws/room_manager_test.go` — 10 个单元测试（FakeConnection，无 HTTP）
- `internal/ws/game_session.go` — GameSession + Commands
- `internal/ws/game_session_test.go` — 10 个单元测试（纯领域逻辑）

**修改：**
- `internal/ws/hub.go` — 从 426 行瘦身到 ~310 行薄路由器

### 测试结果
- Go ws: 37 个测试通过（17 集成 + 10 RoomManager + 10 GameSession）
- Go game: 6 个测试通过
- TypeScript: 11 个测试通过

### 撤回方式
```bash
git revert HEAD
```

## 2026-06-09 — Hub 重构代码审查 + Issue 发布

### 问题
对 commit `d6dafe3`（refactor: decompose Hub monolith into three deep modules）进行全量代码审查，识别潜在缺陷并发布为可独立领取的 Issues。

### 解决方案
1. 启动 9 个独立搜索代理（逐行扫描、移除行为审计、跨文件追踪、Go 语言陷阱、包装器检查、复用机会、简化机会、效率问题、架构深度检查）
2. 汇总 72 个候选发现，去重为 13 个独立缺陷
3. 启动 3 个验证代理，12 个 CONFIRMED、1 个 REFUTED
4. 缺口扫描发现 6 个额外缺陷
5. 整合为 7 个垂直切片 Issues 发布到 GitHub

### 发现的关键缺陷
- 🔴 **数据竞争**: `h.connToRoom` 在 3 个 handler 中无锁读取，Go 并发 map 读写触发 fatal panic
- 🔴 **授权绕过**: 信任客户端 `msg.PlayerID` 而非从连接推导身份
- 🟠 **wsConn 缺少写互斥锁**: gorilla/websocket 要求单写入者
- 🟠 **generateRoomID TOCTOU**: 唯一性检查与插入分离
- 🟠 **AddPlayer 无去重**: 重连场景产生重复玩家条目
- 🟡 **双重玩家列表非原子更新**: room.clients 与 gs.players 可能漂移

### 发布的 Issues
- #23: connToRoom 数据竞争修复
- #24: 授权绕过修复（从连接推导身份）
- #25: wsConn 写互斥锁
- #26: generateRoomID TOCTOU 修复
- #27: AddPlayer/JoinRoom/connToRoom 修复
- #28: 接口完善 + 错误处理
- #29: playerID 校验（依赖 #24）

### 修改文件
- 无代码修改（纯审查 + Issue 发布）

### 撤回方式
```bash
# 关闭所有发布的 Issues
gh issue close 23 24 25 26 27 28 29
```

## 2026-06-09 — TDD 修复 Issue #23: connToRoom 数据竞争

### 问题
`handleSetStoryteller`、`handleAssignCharacters`、`handleSubmitEvent` 读取 `h.connToRoom[conn]` 未持有 `h.mu`，与并发写入构成 fatal 并发 map 访问。

### 解决方案（TDD 流程）
1. **RED**: 编写 `TestConnToRoomRaceSetStoryteller` — 并发 join + storyteller 调用，触发 `fatal error: concurrent map read and map write` at hub.go:164
2. **GREEN**: 为 3 个 handler 的 `h.connToRoom[conn]` 读取添加 `h.mu.RLock()` 保护；合并 `handleAssignCharacters` 和 `handleSubmitEvent` 中原本分离的两次 RLock 为一次
3. **验证**: 10 次连续运行全部通过，全量测试无回归

### 修改文件
- `internal/ws/hub.go` — 3 处 connToRoom 读取添加 RLock
- `internal/ws/hub_race_test.go` — 新增 3 个竞态测试

### 撤回方式
```bash
git revert HEAD
```

## 2026-06-09 — TDD 修复全部 7 个 Code Review Issues (#23-#29)

### 问题
代码审查发现的 15 个缺陷，整合为 7 个 Issues，使用 TDD 流程逐个修复。

### 修复汇总（TDD: RED → GREEN → 重构）

| Issue | 修复内容 | 关键变更 |
|-------|---------|---------|
| #23 | connToRoom 数据竞争 | 3 个 handler 添加 RLock 保护 |
| #24 | 授权绕过 | 从连接推导身份，删除死代码 |
| #25 | wsConn 写互斥锁 | 添加 sync.Mutex 到 wsConn |
| #26 | generateRoomID TOCTOU | 唯一性检查移入 Lock 临界区 |
| #27 | AddPlayer 去重 + connToRoom 泄漏 | AddPlayer 更新逻辑 + handleDisconnect 重构 |
| #28 | 接口完善 + 错误处理 | Connection.ReadMessage + Broadcast 错误日志 |
| #29 | playerID 校验 | applyAssignCharacters 验证 playerID 存在 |

### 新增测试（13 个）
- `hub_race_test.go`: 3 个竞态测试（connToRoom 并发读写）
- `hub_auth_test.go`: 2 个授权测试（伪造 PlayerID）
- `conn_test.go`: 1 个并发写入测试
- `room_manager_race_test.go`: 1 个并发唯一性测试
- `player_dedup_test.go`: 3 个去重/泄漏测试
- `game_session_test.go`: 1 个伪造 playerID 测试
- `conn.go`: FakeConnection.ReadMessage + ClearMessages

### 修改文件
- `internal/ws/hub.go` — 6 处修改
- `internal/ws/conn.go` — 接口扩展 + wsConn mutex
- `internal/ws/room_manager.go` — GetPlayerByConn + generateRoomIDUnlocked
- `internal/ws/game_session.go` — AddPlayer 去重 + playerID 校验

### 测试结果
- 44 个测试全部通过（含 13 个新增）

### 撤回方式
```bash
git revert HEAD
```

## 2026-06-09 16:07:37 --- Code Review 缺口缺陷二次 TDD 修复 --- 使用垂直切片 RED→GREEN→Refactor 修复 --- 修改 backend ws 模块与测试

### 问题
在 Issue #23-#29 修复后的二次代码审查中，又发现 9 个缺口问题：
- `handleLeaveRoom` 仍信任客户端 `msg.PlayerID`，攻击者可伪造 `PlayerID` 踢出其他玩家
- `FakeConnection.ReadMessage` 在 `incomingCh` 路径下无法被 `Close()` 唤醒，测试可能挂死
- `FakeConnection.ReadMessage` 使用 busy-wait goroutine 轮询 `closed`，存在 goroutine 泄漏风险
- `FakeConnection.incomingCh` 读取缺少互斥保护，存在测试 data race 风险
- `handleSetStoryteller` 对未知连接返回误导性的 `no game session`
- `handleSetStoryteller` 传入 `SetStorytellerCmd` 的 `SenderID` 仍使用可伪造的 `msg.PlayerID`
- `handleAssignCharacters` / `handleSubmitEvent` 对未知连接处理不一致
- `generateRoomIDUnlocked` 在 6 位房间号耗尽时无限循环
- `handleCreateRoom` 直接写 `room.clients`，绕过房间成员添加语义

### 解决方案
1. **RED → GREEN: LeaveRoom 授权绕过**
   - 新增 `TestLeaveRoomIgnoresForgedPlayerID`
   - `handleLeaveRoom` 改为从 `connToRoom` + `GetPlayerByConn` 派生真实 `playerID`
   - 删除 `connToRoom`、移除 `GameSession.players`、广播 `PlayerLeft` 全部使用派生身份
2. **RED → GREEN: FakeConnection 阻塞/泄漏修复**
   - 新增 `TestFakeConnectionReadMessageClosePropagation`
   - 新增 `TestFakeConnectionReadMessageReturnsData`
   - `FakeConnection` 增加 `closeCh`
   - `Close()` 幂等关闭 `closeCh`
   - `ReadMessage()` 使用 `select` 同时监听 `incomingCh` 与 `closeCh`，移除 busy-wait goroutine 与 `time.Sleep`
3. **RED → GREEN: 未知连接与 SetStoryteller 身份一致性**
   - 新增 `TestSetStorytellerRejectsUnknownConnection`
   - 新增 `TestSetStorytellerUsesConnectionIdentity`
   - 新增 `TestAssignCharactersRejectsUnknownConnection`
   - 新增 `TestSubmitEventRejectsUnknownConnection`
   - 3 个 handler 对 `roomID == ""` / 派生 `senderID == ""` 返回 `not in any room`
   - `SetStorytellerCmd.SenderID` 改用连接派生的 `senderID`
   - `handleSetStoryteller` 合并 `connToRoom` 与 `sessions` 的 RLock 读取，缩小 TOCTOU 窗口
4. **RED → GREEN: RoomID 耗尽保护**
   - 新增 `TestGenerateRoomIDExhaustionPanics`
   - `generateRoomIDUnlocked` 增加 `maxAttempts`，房间号空间耗尽时 fail-fast panic，避免永久循环
5. **Refactor: 创建房间成员添加语义统一**
   - 新增 `TestCreateRoomRegistersCreatorByConnection` 作为特征测试 [Characterization Test]
   - 提取 `Room.addClient` 私有 helper
   - `handleCreateRoom` 与 `RoomManager.JoinRoom` 共享成员添加逻辑

### 修改文件
- `packages/backend/internal/ws/hub.go`
  - `handleLeaveRoom` 从连接推导身份
  - `handleSetStoryteller` 未知连接检查、合并锁读取、使用派生 `senderID`
  - `handleAssignCharacters` / `handleSubmitEvent` 增加未知连接检查
  - `handleCreateRoom` 改用 `room.addClient`
- `packages/backend/internal/ws/conn.go`
  - `FakeConnection` 增加 `closeCh`
  - `Close()` 幂等关闭
  - `ReadMessage()` 改为 `select`，移除轮询 goroutine
- `packages/backend/internal/ws/room_manager.go`
  - 新增 `Room.addClient`
  - `JoinRoom` 复用 `addClient`
  - `generateRoomIDUnlocked` 增加耗尽保护
- `packages/backend/internal/ws/conn_test.go`
  - 新增 FakeConnection 关闭传播与读消息测试
- `packages/backend/internal/ws/hub_auth_test.go`
  - 新增 LeaveRoom 伪造身份、未知连接、SetStoryteller 身份一致性测试
- `packages/backend/internal/ws/room_manager_race_test.go`
  - 新增房间号耗尽 fail-fast 测试
- `packages/backend/internal/ws/player_dedup_test.go`
  - 新增创建房间注册创建者连接测试

### 测试结果
- `go test ./internal/ws/ -count=1 -timeout 30s` ✅ 通过
- `go test ./... -count=1 -timeout 60s` ✅ 通过
- `git diff --check` ✅ 无 whitespace error；仅 Windows CRLF 提示
- `CGO_ENABLED=1 go test -race ./internal/ws/ -count=1 -timeout 60s` ✅ 通过
- `CGO_ENABLED=1 go test -race ./... -count=1 -timeout 90s` ✅ 通过

### 撤回方式
```bash
# 若本次修改作为单独提交：
git revert HEAD

# 若仍在工作区未提交：
git checkout -- packages/backend/internal/ws/conn.go \
  packages/backend/internal/ws/hub.go \
  packages/backend/internal/ws/room_manager.go \
  packages/backend/internal/ws/conn_test.go \
  packages/backend/internal/ws/hub_auth_test.go \
  packages/backend/internal/ws/room_manager_race_test.go \
  packages/backend/internal/ws/player_dedup_test.go \
  work.md
```
