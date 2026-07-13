# Work Log

## 2026-07-11 17:53:50 +08:00 --- 启动开发服务器 [Dev Server] 时后端 [Backend] 因缺少至少 32 字节的环境变量凭证密钥 [Credential Signing Key] 报错退出 --- 创建本地环境变量配置文件 [Environment Configuration File] 并配置凭证密钥，以便在启动时注入 --- 修改了 `c:\Users\Qilia\Desktop\blood-on-the-clocktower\.env`、`c:\Users\Qilia\Desktop\blood-on-the-clocktower\work.md`

### 撤回方式 [Rollback Strategy]
删除 `c:\Users\Qilia\Desktop\blood-on-the-clocktower\.env` 文件，并回滚对 `c:\Users\Qilia\Desktop\blood-on-the-clocktower\work.md` 的修改。

## 2026-06-13 16:00:42 +08:00 --- 发现 txt 中指定的 vless/tuic 节点尚未加入目标 yaml --- 通过追加两个代理节点 [Proxy Nodes] 并更新策略组 [Proxy Groups] 引用解决 --- 修改了 `C:\Users\Qilia\Desktop\2625_updated_manual_select_no_old_racknerd(1).yaml`、`C:\Users\Qilia\Desktop\blood-on-the-clocktower\work.md`

### 撤回方式 [Rollback Strategy]
从目标 yaml 删除 `vless-reality-vision-ecs-Slbca` 和 `tuic5-ecs-Slbca` 两个 `proxies` 块，并删除它们在 `自动选择` 锚点列表 [Anchor List] 与 `🌍选择代理节点` 列表中的引用；同时删除本日志条目。

## 2026-06-11 09:41 — MVP 前后端集成完成

### 问题
核心逻辑模块（投票引擎、死亡系统、胜利条件、夜间阶段）已完成，需要集成到前后端以支持完整游戏流程。

### 解决方案

#### 后端集成 (Go)
1. **更新 `packages/backend/internal/game/game.go`**:
   - 新增事件类型：PlayerDied、NominationStarted、NominationResolved、NightAction、GameEnded
   - 新增枚举：DeathCause、NightActionType、WinReason
   - 新增结构：Nomination、DeathRecord、NightAction

2. **更新 `packages/backend/internal/ws/game_session.go`** (+616 行):
   - 新增命令：StartGame、ChangePhase、Nominate、CastVote、ResolveNomination、ExecutePlayer、SubmitNightAction、ResolveNight
   - 实现阶段转换逻辑
   - 实现胜利条件检查
   - 实现夜间行动处理

3. **更新 `packages/backend/internal/ws/message.go`**:
   - 新增消息类型
   - 更新 RoomState 包含完整游戏状态

#### 前端集成 (TypeScript)
1. **更新 `packages/core/src/types/index.ts`**:
   - 新增类型：DeathCause、NominationState、NightActionRecord
   - 扩展 GameState 和 GameEvent

2. **更新 `packages/core/src/state-machine/index.ts`**:
   - 新增事件处理：PLAYER_DIED、NOMINATION_STARTED、NOMINATION_RESOLVED、NIGHT_ACTION、GAME_OVER

3. **更新 `packages/core/src/websocket/index.ts`**:
   - 新增 8 个消息类型和对应的客户端方法

4. **更新 `packages/frontend/src/pages/index/index.tsx`**:
   - 完整游戏 UI：阶段显示、投票界面、夜间控制、死亡追踪、胜利条件

5. **更新 `packages/frontend/src/pages/index/index.css`**:
   - 新增游戏 UI 样式

### 测试结果
- ✅ 后端 Go 测试通过
- ✅ Core TypeScript 测试通过 (130 个测试)
- ✅ 前端 TypeScript 类型检查通过
- ✅ 前端构建成功

### 修改文件
- `packages/backend/internal/game/game.go`
- `packages/backend/internal/ws/game_session.go`
- `packages/backend/internal/ws/message.go`
- `packages/core/src/types/index.ts`
- `packages/core/src/state-machine/index.ts`
- `packages/core/src/websocket/index.ts`
- `packages/core/src/death-system/index.ts`
- `packages/core/src/win-conditions/index.ts`
- `packages/frontend/src/pages/index/index.tsx`
- `packages/frontend/src/pages/index/index.css`

### 撤回方式
```bash
git checkout -- packages/backend/
git checkout -- packages/core/src/
git checkout -- packages/frontend/src/
```

---

## 2026-06-10 18:24 — 代码审查修复

### 问题
代码审查发现以下问题需要修复：

**CRITICAL:**
1. `package.json` 缺少 `./night-phase` 子路径导出
2. `death-system/index.ts` 使用类型断言绕过 `readonly` 约束

**HIGH:**
3. `night-phase` 的 `firstNightOnly` 字段从未使用（死代码）
4. `voteReducer` 的 `ctx` 参数不安全的默认值
5. `VotePhase` 包含未使用的 `'nominating'` 状态
6. `checkWinAfterNightDeath` 中胜利条件检查顺序错误

### 解决方案

1. **添加 `./night-phase` 子路径导出** — `package.json`
2. **修复 readonly 突变** — 使用展开运算符替代类型断言
3. **移除 `firstNightOnly` 字段** — 从 `WakeOrderEntry` 接口和所有常量中移除
4. **使 `ctx` 参数必需** — 移除默认值，强制调用者提供上下文
5. **移除 `'nominating'` 状态** — 从 `VotePhase` 类型中移除
6. **修复胜利条件顺序** — 先检查 Imp 自杀（更具体的原因），再检查多数派

### 修改文件
- `packages/core/package.json` — 添加 night-phase 子路径导出
- `packages/core/src/death-system/index.ts` — 修复 readonly 突变
- `packages/core/src/night-phase/index.ts` — 移除 firstNightOnly 字段
- `packages/core/src/vote-engine/index.ts` — 使 ctx 必需，移除 nominating 状态
- `packages/core/src/vote-engine/__tests__/vote-engine.test.ts` — 更新测试以传递必需的 ctx
- `packages/core/src/win-conditions/index.ts` — 修复胜利条件检查顺序

### 测试结果
- ✅ 全部 130 个测试通过
- ✅ TypeScript 类型检查通过
- ✅ 后端 Go 测试通过

### 撤回方式
```bash
git checkout -- packages/core/package.json
git checkout -- packages/core/src/death-system/index.ts
git checkout -- packages/core/src/night-phase/index.ts
git checkout -- packages/core/src/vote-engine/index.ts
git checkout -- packages/core/src/vote-engine/__tests__/vote-engine.test.ts
git checkout -- packages/core/src/win-conditions/index.ts
```

---

## 2026-06-10 18:09 — MVP 核心逻辑模块完成

### 问题
项目缺少完整的游戏循环逻辑，无法从头到尾运行一局 Blood on the Clocktower 游戏。需要实现投票引擎、死亡系统、胜利条件、夜间阶段四个核心模块。

### 解决方案
通过并行工作流同时开发四个独立模块，采用统一的架构模式（纯函数 reducer + 不可变状态 + 验证器）：

#### 1. 投票引擎（Vote Engine）— 已有实现
- 文件：`packages/core/src/vote-engine/index.ts`
- 功能：提名系统、投票机制、多数决阈值、幽灵票、处决判定
- 测试：37 个单元测试

#### 2. 死亡系统（Death System）— 已有实现
- 文件：`packages/core/src/death-system/index.ts`
- 功能：死亡记录、幽灵票、死亡触发器（Saint、Scarlet Woman、Ravenkeeper、Undertaker）
- 测试：50 个单元测试

#### 3. 胜利条件（Win Conditions）— 已有实现 + 新增测试
- 文件：`packages/core/src/win-conditions/index.ts`
- 功能：善良获胜（Imp 处决、市长结局）、邪恶获胜（2 人存活、Saint 处决、Imp 自杀）
- 测试：32 个单元测试（本次新增）

#### 4. 夜间阶段（Night Phase）— 本次新建
- 文件：`packages/core/src/night-phase/index.ts`
- 功能：唤醒顺序系统、角色能力执行、Storyteller 夜间操作、黎明结果解析
- 特性：
  - 第一夜唤醒顺序（9 个角色）
  - 后续夜唤醒顺序（8 个角色）
  - 12 种夜间行动类型
  - 行动验证和状态追踪
  - 黎明结果生成

### 架构设计
所有模块遵循统一模式：
- **纯函数 Reducer**：`(state, event) => newState`
- **不可变状态**：使用 `ReadonlyMap`、`ReadonlySet`、`readonly` 属性
- **验证器**：独立的验证函数，返回错误码或 `null`
- **事件驱动**：通过事件类型触发状态转换

### 修改文件
- `packages/core/src/night-phase/index.ts`（新建）— 夜间阶段实现
- `packages/core/src/win-conditions/__tests__/win-conditions.test.ts`（新建）— 32 个单元测试
- `packages/core/src/death-system/index.ts` — 修复 TypeScript 类型错误

### 测试结果
- ✅ 全部 130 个测试通过（5 个测试文件）
- ✅ TypeScript 类型检查通过
- ✅ 后端 Go 测试通过

### 撤回方式
```bash
rm packages/core/src/night-phase/index.ts
rm packages/core/src/win-conditions/__tests__/win-conditions.test.ts
git checkout -- packages/core/src/death-system/index.ts
```

---

## 2026-06-10 16:47 — 实现死亡系统（Death System）

### 问题
游戏缺少玩家死亡系统。Blood on the Clocktower 需要：死亡追踪（死因、死亡日、击杀者）、幽灵票机制（死人获得 1 次幽灵票）、死亡触发器（Saint、Scarlet Woman、Ravenkeeper、Undertaker）、死亡事件广播。

### 解决方案
采用与 vote-engine 相同的架构模式（纯函数 reducer + 不可变状态 + 验证器），实现死亡系统：

1. **`DeathState` 接口**：跟踪死亡记录（`deaths: Map`）和剩余幽灵票（`ghostVotesRemaining: Set`）
2. **`DeathCause` 类型**：`'execution' | 'night_kill' | 'ability'` 三种死因
3. **`DeathRecord` 接口**：记录 playerId、cause、dayNumber、killedBy
4. **`DeathEvent` 联合类型**：PLAYER_DIED / GHOST_VOTE_CAST 两种事件
5. **`deathReducer` 纯函数**：事件驱动的状态转换，含完整验证（防止重复死亡、无效日数、重复使用幽灵票）
6. **`computeDeathTriggers` 纯函数**：根据死亡上下文计算触发器
   - `saint_execution`：Saint 被处决时邪恶阵营立即获胜
   - `scarlet_woman`：Imp 死亡且存活 5+ 人时，Scarlet Woman 变为 Imp
   - `ravenkeeper`：Ravenkeeper 夜间死亡时可查看一名玩家角色
   - `undertaker`：处决发生时 Undertaker 学习被处决者角色
7. **6 个便捷辅助函数**：isDead、getDeathRecord、hasGhostVote、getDeathCount、getDeathsByCause、getDeathsOnDay、createDeathAnnouncement
8. **50 个单元测试**：覆盖死亡记录、幽灵票、所有触发器场景、完整游戏集成场景

### 设计决策
- **触发器与 Reducer 分离**：`deathReducer` 只负责状态转换，`computeDeathTriggers` 独立计算触发器。游戏引擎负责消费触发器并执行副作用，保持 reducer 纯函数特性。
- **幽灵票双重追踪**：`DeathState.ghostVotesRemaining` 记录剩余幽灵票资格，`VoteState.ghostVotesUsed`（已有）记录投票引擎中的使用情况。两个模块各司其职。
- **DeathAnnouncement 不含 killedBy**：广播给所有玩家的公告只包含 playerId、cause、dayNumber，击杀者信息为私有上下文。

### 修改文件
- `packages/core/src/death-system/index.ts`（新建）— 死亡系统实现
- `packages/core/src/death-system/__tests__/death-system.test.ts`（新建）— 50 个单元测试
- `packages/core/src/index.ts` — 添加 death-system 导出
- `packages/core/package.json` — 添加 `./death-system` 子路径导出

### 测试结果
- 全部 98 个测试通过（50 新 + 48 已有）

### 撤回方式
```bash
rm packages/core/src/death-system/index.ts
rm packages/core/src/death-system/__tests__/death-system.test.ts
git checkout -- packages/core/src/index.ts
git checkout -- packages/core/package.json
```

---

## 2026-06-10 16:36 — 实现投票引擎（Vote Engine）

### 问题
游戏缺少投票系统实现。Blood on the Clocktower 需要：提名系统（活人提名活人）、投票机制（活人投票+死人幽灵票）、多数决阈值计算、投票状态追踪和状态转换。

### 解决方案
采用与 state-machine 相同的架构模式（纯函数 reducer + 不可变状态），实现投票引擎：

1. **`VoteState` 接口**：跟踪投票阶段、提名人、被提名人、已投票记录、幽灵票使用情况、存活玩家数、投票结果
2. **`VoteEvent` 联合类型**：NOMINATED / VOTE_CAST / VOTING_RESOLVED / VOTE_RESET 四种事件
3. **`voteReducer` 纯函数**：事件驱动的状态转换，包含完整的参数验证（self-nomination、dead nominator/nominee、duplicate vote、spent ghost vote）
4. **`computeMajorityThreshold`**：`ceil(alivePlayers / 2)` 多数决计算
5. **`countVotes`**：统计赞成/反对票数的工具函数
6. **37 个单元测试**：覆盖提名验证、投票验证、幽灵票、完整投票轮次、边界情况

### 修改文件
- `packages/core/src/vote-engine/index.ts`（新建）— 投票引擎实现
- `packages/core/src/vote-engine/__tests__/vote-engine.test.ts`（新建）— 37 个单元测试
- `packages/core/src/index.ts` — 添加 vote-engine 导出
- `packages/core/package.json` — 添加 `./vote-engine` 子路径导出

### 测试结果
- ✅ 全部 48 个测试通过（37 新 + 11 已有）

### 撤回方式
```bash
rm packages/core/src/vote-engine/index.ts
rm packages/core/src/vote-engine/__tests__/vote-engine.test.ts
git checkout -- packages/core/src/index.ts
git checkout -- packages/core/package.json
```

---

## 2026-06-10 14:27 — 修复 TypeScript 类型检查错误和 ESLint 配置

### 问题
1. TypeScript 类型检查有多个错误：
   - 未使用的导入：`GamePhase`, `Player`, `PlayerId`
   - 未使用的变量：`initialState`, `ProtoCharacter`
   - 类型错误：`GameId` 未定义
   - 可选链警告：`Object is possibly 'undefined'`
   - 联合类型访问错误：`GameServerEvent` 属性访问问题

2. ESLint 配置缺失

### 解决方案
1. **修复 `packages/core/src/state-machine/index.ts`**：
   - 移除未使用的导入：`GamePhase`, `Player`, `PlayerId`
   - 添加缺失的导入：`GameId`
   - 将 `initialState` 重命名为 `defaultState` 并使用它

2. **修复 `packages/core/src/types/generated/__tests__/proto-types.test.ts`**：
   - 移除未使用的导入：`ProtoCharacter`

3. **修复 `packages/core/src/websocket/__tests__/websocket-client.test.ts`**：
   - 使用非空断言 `!` 修复可选链警告

4. **修复 `packages/frontend/src/lib/taro-websocket-transport.test.ts`**：
   - 使用类型断言修复 `resolveTask` 调用问题

5. **修复 `packages/frontend/src/pages/index/index.tsx`**：
   - 使用 `in` 操作符进行类型守卫，正确访问 `GameServerEvent` 联合类型的属性

6. **修复 `packages/core/tsconfig.json`**：
   - 添加 `"composite": true` 以支持项目引用

7. **创建 `.eslintrc.json` 配置文件**：
   - 配置 TypeScript ESLint 规则
   - 设置 `no-console` 为 `off`

8. **修复 `packages/core/src/websocket/index.ts`**：
   - 将 `console.error` 改为 `console.warn`

### 修改文件
- `packages/core/src/state-machine/index.ts`
- `packages/core/src/types/generated/__tests__/proto-types.test.ts`
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`
- `packages/frontend/src/lib/taro-websocket-transport.test.ts`
- `packages/frontend/src/pages/index/index.tsx`
- `packages/core/tsconfig.json`
- `.eslintrc.json`
- `packages/core/src/websocket/index.ts`

### 测试结果
- ✅ TypeScript 类型检查：通过
- ✅ ESLint 检查：通过（0 错误，0 警告）
- ✅ 单元测试：13 个测试全部通过
- ✅ 后端 Go 测试：通过
- ✅ 完整构建：成功

### 撤回方式
```bash
git checkout -- packages/core/src/state-machine/index.ts
git checkout -- packages/core/src/types/generated/__tests__/proto-types.test.ts
git checkout -- packages/core/src/websocket/__tests__/websocket-client.test.ts
git checkout -- packages/frontend/src/lib/taro-websocket-transport.test.ts
git checkout -- packages/frontend/src/pages/index/index.tsx
git checkout -- packages/core/tsconfig.json
rm .eslintrc.json
git checkout -- packages/core/src/websocket/index.ts
```

---

## 2026-06-10 10:53 — 启动项目服务

### 操作
1. 启动后端 Go 服务器（端口 8080）
2. 启动前端 Taro 开发服务器（watch 模式）

### 服务状态
- 后端服务器：✅ 运行中 (`:8080`)
- 前端开发服务器：✅ 运行中 (watch 模式)

### 验证
- 健康检查：`curl http://localhost:8080/health` → 返回 `ok`
- 前端编译：成功（2.50s）

---

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
2. 测试通过，进入重构阶段。
### 修改文件
- 无（测试通过，无需修改）
### 撤回方式
- 无需撤回

---

2026-06-11 10:41:19 +08:00 --- 发现 WebSocket 游戏流程新增处理器存在 START_GAME 权限缺失、CHANGE_PHASE 字符串阶段无法反序列化、EXECUTE_PLAYER 字段名不兼容、SUBMIT_NIGHT_ACTION 泄露夜间行动目标，以及前端阶段枚举映射错误 --- 使用服务端消息兼容解析、会话层权限校验、夜间行动定向发送、客户端字段统一、前端阶段映射修正和回归测试解决 --- 修改了 packages/backend/internal/ws/message.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/core/src/websocket/index.ts、packages/frontend/src/pages/index/index.tsx

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/message.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/core/src/websocket/index.ts packages/frontend/src/pages/index/index.tsx work.md`，并删除 `packages/backend/internal/ws/hub_game_flow_test.go`。

---

2026-06-11 11:02:29 +08:00 --- 发现真实玩家 MVP 的断线重连 [Reconnect] 会被当成主动离开，导致玩家从 GameSession [游戏会话] 移除，重连后可能丢失角色、死亡状态、主持人身份、当前阶段快照 [Snapshot]、幽灵票状态，并且前端未兼容后端的 gameEnded/nightActionSubmitted/voteCast.decision 协议字段 --- 使用连接生命周期 [Connection Lifecycle] 与房间生命周期 [Room Lifecycle] 解耦、会话层 AddOrReconnectPlayer 保留身份状态、断线不广播 PlayerLeft、主动离开才按参与者数量销毁房间、RoomState 增加 ghostVotesRemaining 快照、前端从 ROOM_STATE 恢复阶段/天数/提名/死亡/胜负/幽灵票并兼容后端事件字段解决 --- 修改了 packages/backend/internal/ws/room_manager.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/message.go、packages/backend/internal/ws/room_manager_test.go、packages/backend/internal/ws/ws_test.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/core/src/websocket/index.ts、packages/frontend/src/pages/index/index.tsx

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/room_manager.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/message.go packages/backend/internal/ws/room_manager_test.go packages/backend/internal/ws/ws_test.go packages/backend/internal/ws/hub_game_flow_test.go packages/core/src/websocket/index.ts packages/frontend/src/pages/index/index.tsx work.md`。

---

2026-06-11 11:07:39 +08:00 --- 发现真实玩家关键路径 [Critical Path] 仍会在首夜卡住：前端只允许 Storyteller [主持人] 提交夜间行动，但后端要求 SUBMIT_NIGHT_ACTION 发送者必须是玩家列表中的活人，导致主持人夜晚裁决被拒绝；同时 RESOLVE_NIGHT 依赖前端未发送的 Result 字段才会造成夜杀；后端胜负条件 [Win Condition] 还把邪恶人数不少于善良人数误判为邪恶胜利，导致 5 人局首夜死 1 人后提前结束 --- 使用主持人裁决型夜间行动 [Adjudicated Night Action] 允许 Storyteller 提交 kill 并只让主持人的 kill 在夜晚结算时造成死亡，玩家提交的夜间行动仅作为私密选择通知主持人；将邪恶胜利修正为仅剩 2 名存活玩家；新增完整多人 happy path 测试覆盖创建房间、分配角色、开始首夜、主持人夜杀、白天提名、幽灵票投票、处决 Imp、善良胜利 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub_game_flow_test.go

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub_game_flow_test.go work.md`。

---

2026-06-11 11:11:52 +08:00 --- 发现完整局流程虽然已有 Hub 级内存测试，但缺少通过真实 WebSocket 连接 [WebSocket Connection] 驱动的多客户端联调证据，无法证明 1 个 Storyteller [主持人] + 5 个玩家在实际传输层 [Transport Layer] 能完成创建房间、加入、角色分配、夜晚、提名投票和胜负结算 --- 使用 httptest WebSocket 服务器新增 `TestWebSocketCompleteMVPGameFlow`，通过实际 `WriteJSON`/`ReadMessage` 建立 6 个连接并跑完整 MVP happy path；新增稳定读取目标快照 [Snapshot] 的测试辅助函数，避免广播顺序影响断言 --- 修改了 packages/backend/internal/ws/ws_test.go

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/ws_test.go work.md`。

---

2026-06-11 11:15:35 +08:00 --- 发现后端已经支持断线重连 [Reconnect]，但 core `GameWebSocketClient` 只会重开 WebSocket，不会在重连后重新加入房间，且前端页面把 `maxReconnectAttempts` 设为 0，导致真实玩家移动网络抖动后仍停留在未入房连接状态 --- 使用客户端会话恢复 [Session Resume] 记录最后一次房间身份：`JOIN_ROOM` 后保存 roomId/playerId/playerName，`CREATE_ROOM` 收到 `ROOM_STATE` 后用服务端返回房间号保存恢复会话；每次 socket 打开后自动发送 `JOIN_ROOM` 恢复房间；前端改为 2 秒间隔、最多 8 次有界自动重连；新增 core WebSocket 客户端测试验证创建房间后断线重连会自动回到原房间 --- 修改了 packages/core/src/websocket/index.ts、packages/core/src/websocket/__tests__/websocket-client.test.ts、packages/frontend/src/pages/index/index.tsx

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/core/src/websocket/index.ts packages/core/src/websocket/__tests__/websocket-client.test.ts packages/frontend/src/pages/index/index.tsx work.md`。

---

2026-06-11 11:19:59 +08:00 --- 发现小程序页面 [Mini Program Page] 只持久化 playerId，服务器地址、昵称、最近房间号和玩家数都只存在内存里，玩家重开页面后无法一键恢复最近房间，即使 core 客户端支持 Session Resume [会话恢复] 也缺少页面级入口 --- 使用 Taro 本地存储 [Local Storage] 持久化 wsUrl、playerName、lastRoomId、maxPlayers；在收到 ROOM_STATE、创建/加入/离开房间、重置身份时同步最近房间状态；新增“恢复最近房间”动作，未连接时先连接再 JOIN_ROOM，已连接时直接 JOIN_ROOM --- 修改了 packages/frontend/src/pages/index/index.tsx

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/src/pages/index/index.tsx work.md`。

---

2026-06-11 11:25:04 +08:00 --- 发现创建房间后邀请朋友仍需要手动读取/转述房间号，缺少真实玩家邀请入口 [Invite Flow]，容易阻塞首局开局 --- 使用 Taro 剪贴板 API [Clipboard API] 增加复制邀请信息动作，并在房间卡片中突出显示当前房间号 [Room Code]；复制内容包含房间号、昵称和加入提示，复制成功后显示 Toast 反馈 [User Feedback] --- 修改了 packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/index/index.css

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/index/index.css work.md`。

2026-06-11 15:44:24 +08:00 --- 发现产品版角色/剧本支持 [Script Support] 仍散落在前端硬编码 [Hard-coded] 常量和后端 Trouble Brewing 校验中：核心包缺少统一角色目录 [Character Catalog]，夜晚顺序 [Wake Order] 使用了错误的 `fortune_teller` ID，后端角色分配 [Character Assignment] 没有拒绝重复角色，也没有处理男爵 [Baron] 带来的外来者数量修正，真实 WebSocket 测试 [Transport Test] 还会在投票快照未到达时提前结算 --- 使用核心脚本目录统一 Trouble Brewing 角色、角色分布、默认分配和夜晚唤醒步骤；夜晚阶段复用脚本目录并修正 `fortuneteller` ID；后端新增脚本定义、重复角色拒绝和男爵设置修正；前端改用核心默认分配并为 Storyteller [主持人] 展示夜晚唤醒清单；WebSocket 流程测试在结算前等待 3 张投票进入服务器快照 --- 修改了 packages/core/src/scripts/index.ts、packages/core/src/scripts/__tests__/scripts.test.ts、packages/core/src/night-phase/index.ts、packages/core/src/night-phase/__tests__/night-phase.test.ts、packages/core/src/index.ts、packages/core/package.json、packages/backend/internal/game/characters.go、packages/backend/internal/game/characters_test.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/ws_test.go、packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/index/index.css、packages/core/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/core/src/night-phase/index.ts packages/core/src/index.ts packages/core/package.json packages/backend/internal/game/characters.go packages/backend/internal/game/characters_test.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/ws_test.go packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/index/index.css packages/core/tsconfig.tsbuildinfo work.md`，并删除 `packages/core/src/scripts/index.ts`、`packages/core/src/scripts/__tests__/scripts.test.ts`、`packages/core/src/night-phase/__tests__/night-phase.test.ts`。

---

2026-06-11 15:51:20 +08:00 --- 发现剧本支持 [Script Support] 只存在核心目录，尚未进入房间协议 [Room Protocol] 和前端路由 [Frontend Routing]：创建房间无法携带 `scriptId`，`ROOM_STATE` 缺少剧本元数据，未知剧本不会被协议层明确拒绝，前端仍只有单页，无法独立查看角色目录与夜晚顺序 --- 使用房间元数据 [Room Metadata] 贯穿后端 RoomManager、GameSession、Hub 与 RoomState；创建房间时校验未知剧本并返回 `unsupported script`；核心 WebSocket 客户端支持发送 `scriptId`；新增 Taro 剧本页面 [Script Page] 展示 Trouble Brewing 角色目录、合法玩家分布和首夜/后续夜唤醒顺序；首页创建房间传入默认剧本并显示服务器返回的剧本名 --- 修改了 packages/backend/internal/ws/message.go、packages/backend/internal/ws/room_manager.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/room_manager_test.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/core/src/websocket/index.ts、packages/core/src/websocket/__tests__/websocket-client.test.ts、packages/frontend/src/app.config.ts、packages/frontend/src/pages/index/index.config.ts、packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/scripts/index.config.ts、packages/frontend/src/pages/scripts/index.tsx、packages/frontend/src/pages/scripts/index.css、packages/core/tsconfig.tsbuildinfo、packages/frontend/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/message.go packages/backend/internal/ws/room_manager.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/room_manager_test.go packages/backend/internal/ws/hub_game_flow_test.go packages/core/src/websocket/index.ts packages/core/src/websocket/__tests__/websocket-client.test.ts packages/frontend/src/app.config.ts packages/frontend/src/pages/index/index.config.ts packages/frontend/src/pages/index/index.tsx packages/core/tsconfig.tsbuildinfo packages/frontend/tsconfig.tsbuildinfo work.md`，并删除 `packages/frontend/src/pages/scripts/index.config.ts`、`packages/frontend/src/pages/scripts/index.tsx`、`packages/frontend/src/pages/scripts/index.css`。

---

2026-06-11 16:00:09 +08:00 --- 发现夜晚信息流 [Night Information Flow] 仍由前端本地推测：后端 `ROOM_STATE` 不包含当前夜晚唤醒步骤 [Wake Step]，Storyteller [主持人] 可以任意提交 `kill` 并直接结算夜晚，缺少动作顺序、目标数量和未完成步骤的防误操 [Misclick Prevention] --- 使用后端领域层 [Domain Layer] 定义 Trouble Brewing 首夜/后续夜唤醒顺序与目标数规则；GameSession 进入夜晚时初始化 `nightNumber/nightWakeIndex`，`ROOM_STATE` 返回 `nightWakeSteps/currentNightWakeStep/currentNightWakeIndex`；主持人提交夜晚行动时必须匹配当前步骤并满足 `minTargets/maxTargets`，未完成所有在场步骤时拒绝 `RESOLVE_NIGHT`；玩家私密夜晚提交仍只通知本人和主持人；前端夜晚面板改为消费服务端当前步骤、自动选中当前动作、展示已完成/当前/待处理状态并按目标数做本地校验 --- 修改了 packages/backend/internal/game/game.go、packages/backend/internal/game/characters.go、packages/backend/internal/game/characters_test.go、packages/backend/internal/ws/message.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/backend/internal/ws/ws_test.go、packages/core/src/websocket/index.ts、packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/index/index.css、packages/core/tsconfig.tsbuildinfo、packages/frontend/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/game/game.go packages/backend/internal/game/characters.go packages/backend/internal/game/characters_test.go packages/backend/internal/ws/message.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub_game_flow_test.go packages/backend/internal/ws/ws_test.go packages/core/src/websocket/index.ts packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/index/index.css packages/core/tsconfig.tsbuildinfo packages/frontend/tsconfig.tsbuildinfo work.md`。

---

2026-06-11 16:05:58 +08:00 --- 发现后端提名结算 [Nomination Resolution] 使用 `yes > no` 判定处决，和核心投票引擎 [Vote Engine] 的 Blood on the Clocktower 多数票阈值 [Majority Threshold] `ceil(alive/2)` 不一致，导致 5 人局首夜死 1 人后 1 张赞成票也能处决；同时成功处决后仍停留白天，允许同一天继续提名/处决的误操作 --- 使用后端存活玩家数计算处决阈值，`NominationResolvedEvent` 增加 `requiredVotes`，赞成票数达到阈值才处决；未达阈值返回白天，达阈值且未触发胜负时直接进入夜晚并初始化夜晚唤醒步骤；前端投票面板显示服务端处决阈值和结果阈值；新增回归测试覆盖低于阈值不处决、达到阈值处决后进入夜晚 --- 修改了 packages/backend/internal/game/game.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/core/src/websocket/index.ts、packages/frontend/src/pages/index/index.tsx、packages/core/tsconfig.tsbuildinfo、packages/frontend/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/game/game.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub_game_flow_test.go packages/core/src/websocket/index.ts packages/frontend/src/pages/index/index.tsx packages/core/tsconfig.tsbuildinfo packages/frontend/tsconfig.tsbuildinfo work.md`。

---

2026-06-12 09:44:33 +08:00 --- 发现后端仍保留 `SUBMIT_EVENT` 原始事件直通口，已入房客户端可以伪造任意 `GameEvent` 并广播阶段切换、死亡或胜负事件，破坏服务器权威 [Server Authoritative] 与游戏边界 [Game Boundary] --- 使用显式命令 [Explicit Commands] 取代原始事件广播：`GameSession.applySubmitEvent` 统一拒绝原始事件提交并返回明确错误；将旧端到端广播测试改为拒绝与不广播的回归测试 [Regression Test]；更新会话层单元测试验证拒绝后不产生状态变更 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/ws_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/ws_test.go work.md`。

---

2026-06-12 09:52:57 +08:00 --- 发现 MVP 仍缺少服务器重启恢复 [Restart Recovery]：房间元数据 [Room Metadata] 与游戏会话 [Game Session] 只在内存中，后端进程重启后玩家即使保留房间号也无法恢复进行中游戏；同时 `go test -race ./internal/ws` 发现 `ROOM_STATE` 广播复用内部 `Nomination` 指针，JSON 编码 [JSON Encoding] 与提名结算 [Nomination Resolution] 并发时存在数据竞争 [Data Race] --- 使用可插拔快照存储 [Snapshot Store] 增加文件快照实现 [File Snapshot Store]，持久化房间、玩家、主持人、角色、阶段、夜晚进度、提名、死亡、幽灵票和胜负状态；后端入口支持通过 `CLOCKTOWER_SNAPSHOT_PATH` 启用快照恢复；所有有效状态变更后写入快照，断线清理不写入连接状态；`StateForRoom` 对可变字段做深拷贝 [Deep Copy]，并新增重启恢复、房间销毁后快照移除、不可变房间快照和 race 回归验证 --- 修改了 packages/backend/cmd/server/main.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/hub_game_flow_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/cmd/server/main.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/hub_game_flow_test.go work.md`，并删除 `packages/backend/internal/ws/persistence.go`。


---

2026-06-13 16:06:54 +08:00 --- 发现开始游戏 [Start Game] 命令只校验“现有玩家都有角色”，没有在进入夜晚前重新校验玩家人数 [Player Count] 与角色分布 [Role Distribution]；若后端状态绕过正式分配 API [Assignment API]，可能出现 0 人局或非法角色组合仍能开局 --- 在 `applyStartGame` 中从当前玩家角色重建分配表 [Assignment Map]，调用 `ValidateScriptAssignment` 重新校验剧本人数规则 [Script Setup Rules]；新增回归测试 [Regression Tests] 覆盖无实际玩家与非法手工角色分布都不能开局 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go`，并从 `work.md` 删除本条记录；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-12 10:00:34 +08:00 --- 发现文件快照 [File Snapshot] 只能覆盖单机/本地恢复，生产部署 [Production Deployment] 或多实例 [Multi-instance] 场景缺少集中式快照后端；同时 `github.com/redis/go-redis/v9@latest` 会提升 `go` 指令 [Go Directive] 到 1.24，不符合当前后端 `go 1.22` 约束 --- 使用 Context7 查询 go-redis 官方用法，固定 `github.com/redis/go-redis/v9 v9.17.3`；新增 Redis 快照存储 [Redis Snapshot Store]，通过 `CLOCKTOWER_REDIS_URL` 与 `CLOCKTOWER_REDIS_KEY` 配置，缺失键返回空快照，读写使用超时上下文 [Timeout Context]，Redis 优先于文件快照；补充 fake Redis 单元测试 [Unit Tests] 覆盖缺失键、保存/读取、默认 key、关闭客户端和非法 URL；更新后端上下文文档 [Context Documentation]，并通过 `go test ./...`、`go test -race ./internal/ws`、`pnpm test`、`pnpm typecheck`、`pnpm build:frontend`、`pnpm build:core` --- 修改了 packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/persistence_test.go、packages/backend/cmd/server/main.go、packages/backend/go.mod、packages/backend/go.sum、packages/backend/CONTEXT.md、work.md

撤回方式 [Rollback Strategy]：若只撤回 Redis 支持，执行 `git checkout -- packages/backend/CONTEXT.md packages/backend/cmd/server/main.go packages/backend/go.mod packages/backend/go.sum work.md`，删除 `packages/backend/internal/ws/persistence_test.go`，并从 `packages/backend/internal/ws/persistence.go` 移除 `RedisSnapshotStore`、`NewRedisSnapshotStore`、`NewHubWithRedisSnapshot`、`redisClientAdapter` 与 go-redis 相关导入；若要撤回整个快照持久化，则执行上一条 09:52:57 记录的完整回滚。

---

2026-06-12 10:06:37 +08:00 --- 发现人类主持模式 [Human Storyteller Mode] 仍缺少手动判定胜负 [Manual Winner Adjudication]：PRD 要求 Storyteller 能判定游戏胜负，但当前只能依赖自动胜利条件 [Automatic Win Conditions]，实局中遇到主持裁定、玩家认输或规则外结算时无法从前端正式结束游戏 --- 使用显式 `END_GAME` 命令 [Explicit Command] 贯通后端协议、GameSession、Hub、core WebSocket 客户端和 Taro 页面；后端新增 `storyteller_decision` 胜利原因 [Win Reason]，只允许 Storyteller 在已开始且未结束的游戏中结束，赢家 [Winner] 必须是善良或邪恶阵营；客户端发送 `winner/reason/description`，前端 Storyteller 控制区增加结局说明输入和善良/邪恶胜利按钮；补充会话层、Hub 消息路径、赢家解析和 core 发送测试，并通过 `go test ./internal/ws`、`pnpm --filter @clocktower/core test -- websocket-client.test.ts`、`go test ./...`、`pnpm test`、`pnpm typecheck`、`go test -race ./internal/ws`、`pnpm build:core`、`pnpm build:frontend` --- 修改了 packages/backend/internal/game/game.go、packages/backend/internal/ws/message.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/core/src/websocket/index.ts、packages/core/src/websocket/__tests__/websocket-client.test.ts、packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/index/index.css、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/game/game.go packages/backend/internal/ws/message.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/hub_game_flow_test.go packages/core/src/websocket/index.ts packages/core/src/websocket/__tests__/websocket-client.test.ts packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/index/index.css work.md`；若只撤回前端入口，可保留后端协议并仅回滚 `packages/frontend/src/pages/index/index.tsx` 与 `packages/frontend/src/pages/index/index.css`。

---

2026-06-12 10:14:31 +08:00 --- 发现 PRD 房间管理 [Room Management] 中“房主踢出玩家 [Kick Player]”仍未实现：房主只能等待玩家主动离开，无法在准备阶段移除错误身份、重复身份或不守规矩玩家；同时断线重连 [Session Resume] 会让被移除玩家自动回到房间，生产快照 [Production Snapshot] 也缺少踢出名单 --- 使用显式 `KICK_PLAYER` 命令 [Explicit Command] 贯通后端 RoomManager、GameSession、Hub、core WebSocket 客户端和 Taro 页面；只允许房主 [Room Creator] 在准备阶段 [Setup Phase] 踢出其他玩家，服务端从连接身份 [Connection Identity] 推导权限并拒绝伪造 `playerId`；RoomManager 记录被踢玩家并拒绝同一 `playerId` 重进，快照持久化 `kickedPlayerIds`；被踢客户端收到 `kicked from room` 后清空会话恢复 [Session Resume] 和页面房间状态；前端玩家列表在准备阶段向房主显示“踢出”按钮；补充 RoomManager、GameSession、Hub 权限/流程/持久化和 core 客户端测试，并通过 `go test ./internal/ws`、`pnpm --filter @clocktower/core test -- websocket-client.test.ts`、`pnpm typecheck`、`go test ./...`、`pnpm test`、`go test -race ./internal/ws`、`pnpm build:core`、`pnpm build:frontend`、`git diff --check` --- 修改了 packages/backend/internal/ws/message.go、packages/backend/internal/ws/room_manager.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/room_manager_test.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/hub_auth_test.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/backend/internal/ws/persistence_test.go、packages/core/src/websocket/index.ts、packages/core/src/websocket/__tests__/websocket-client.test.ts、packages/frontend/src/pages/index/index.tsx、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/message.go packages/backend/internal/ws/room_manager.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/room_manager_test.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/hub_auth_test.go packages/backend/internal/ws/hub_game_flow_test.go packages/core/src/websocket/index.ts packages/core/src/websocket/__tests__/websocket-client.test.ts packages/frontend/src/pages/index/index.tsx work.md`，并从 `packages/backend/internal/ws/persistence.go` 与 `packages/backend/internal/ws/persistence_test.go` 移除 `KickedPlayerIDs/kickedPlayerIds` 相关字段、快照读写和断言；若要撤回整个快照持久化，则按 09:52:57 与 10:00:34 记录执行完整回滚。

---

2026-06-12 11:57 +08:00 --- 发现以上所有未提交改动累积为 3151 行新增、296 行删除，横跨 28 个文件和 8+ 个功能模块，一次性提交难以追溯和回滚 --- 使用按功能层级拆分提交 [Split Commit] 将改动分为 5 个 commit：(1) `4537e7e` core 脚本目录 + 夜晚阶段模块、(2) `55afa91` 后端快照持久化（文件 + Redis）、(3) `fc72b98` 后端游戏功能（脚本协议/夜晚顺序/提名阈值/踢人/手动结算）+ core WebSocket 客户端、(4) `949eb55` 前端全部功能（断线重连/本地存储/剧本页面/夜晚 UI/邀请/踢人/结算）、(5) `5244acf` 文档与审计日志；并通过 `go test ./...`、`pnpm test`、`pnpm typecheck` 验证每个 commit 后代码均可用 --- 未修改代码内容，仅重新组织提交结构

撤回方式 [Rollback Strategy]：执行 `git reset --soft 5acc32a` 将以上 5 个 commit 合并回工作区。

---

2026-06-12 14:44:27 +08:00 --- 发现前端页面 [Frontend Pages] 中仍有一批页面内工具函数 [Inline Utilities] 难以单独测试，剧本页 [Script Page] 的角色分组和夜晚顺序展示也缺少独立回归测试；同时根仓库缺少一键同时启动后端和 Taro 小程序构建监听的开发入口 [Dev Entrypoint] --- 使用前端工具模块 [Utility Module] 提取协议阶段映射、赢家归一化、本地存储和输入事件读取逻辑，并为首页与剧本页工具函数补充 Vitest 单元测试 [Unit Tests]；根 `package.json` 增加 `dev/dev:backend/dev:frontend` 脚本并引入 `concurrently`，前端包显式提供 `test/test:watch`；通过 `pnpm test`、`pnpm typecheck`、`go test ./...`、`pnpm build:core`、`pnpm build:frontend`、`pnpm lint` 和真实 WebSocket MVP 冒烟测试 [Smoke Test] --- 修改了 package.json、packages/frontend/package.json、pnpm-lock.yaml、packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/scripts/index.tsx、packages/frontend/src/lib/utils.ts、packages/frontend/src/lib/utils.test.ts、packages/frontend/src/pages/scripts/utils.ts、packages/frontend/src/pages/scripts/utils.test.ts、packages/frontend/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- package.json packages/frontend/package.json pnpm-lock.yaml packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/scripts/index.tsx packages/frontend/tsconfig.tsbuildinfo work.md`，并删除 `packages/frontend/src/lib/utils.ts`、`packages/frontend/src/lib/utils.test.ts`、`packages/frontend/src/pages/scripts/utils.ts`、`packages/frontend/src/pages/scripts/utils.test.ts`。

---

2026-06-12 15:20:12 +08:00 --- 发现小程序端 WebSocket 适配器 [WebSocket Adapter] 只在 `SocketTask.onOpen` 回调里推进连接状态；当 Taro/微信运行时 [Runtime] 在 Promise 解析前已经把 `SocketTask.readyState` 置为 OPEN 时，前端会错过打开事件并一直停在 connecting，表现为小程序端连接不上 WebSocket --- 使用回归测试 [Regression Test] 复现已打开 SocketTask 的时序边界，并在注册事件处理器后检查 `readyState === OPEN`，通过 `markOpen` 做幂等 [Idempotent] 状态推进，避免重复触发 --- 修改了 packages/frontend/src/lib/taro-websocket-transport.ts、packages/frontend/src/lib/taro-websocket-transport.test.ts、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/src/lib/taro-websocket-transport.ts packages/frontend/src/lib/taro-websocket-transport.test.ts work.md`。

---

2026-06-12 15:38:38 +08:00 --- 发现后端 Go 领域模型 [Domain Model] 已支持死亡 [Death]、提名 [Nomination]、夜晚行动 [Night Action]、胜负 [Game End] 等 MVP 事件，但 `proto/game.proto` 和生成的 TypeScript 协议类型 [Generated TypeScript Protocol Types] 仍停留在早期基础事件，破坏 ProtoBuf 单一事实源 [Single Source of Truth] 和端到端类型同步 [End-to-End Type Sync] --- 使用协议扩展 [Protocol Extension] 补齐 DeathCause、NightActionType、WinReason 枚举，补齐 GameState 的 deaths、nomination、nightActions、winner 字段，补齐 GameEvent 的 playerDied、nominationStarted、nominationResolved、nightActionSubmitted、gameEnded 事件，并更新生成脚本 [Codegen Script] 与类型测试；运行 `pnpm proto:generate` 重新生成类型，并通过 `go test ./...`、`pnpm --filter @clocktower/core test -- proto-types.test.ts`、`pnpm typecheck` --- 修改了 proto/game.proto、scripts/generate-types.mjs、packages/core/src/types/generated/index.ts、packages/core/src/types/generated/__tests__/proto-types.test.ts、packages/core/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- proto/game.proto scripts/generate-types.mjs packages/core/src/types/generated/index.ts packages/core/src/types/generated/__tests__/proto-types.test.ts packages/core/tsconfig.tsbuildinfo work.md`。

---

2026-06-12 15:41:06 +08:00 --- 发现 PRD 要求游戏结束后自动销毁房间 [Auto Destroy Room]，但后端在自动胜利或主持人手动结束 [Manual End Game] 后只把会话置为 Finished，仍保留 RoomManager 房间、GameSession 和连接映射 [Connection Mapping]，服务器资源不会自动释放 --- 使用统一提交函数 [Commit Function] 收敛 Hub 中所有状态更新后的持久化 [Persistence] 与房间状态广播 [Room State Broadcast]；当房间阶段变为 Finished 时，先向当前客户端广播最终 `ROOM_STATE`，再删除房间、会话和连接映射，最后持久化清理后的快照；补充自动胜利和手动结束两条回归测试 [Regression Tests]，确认客户端能收到最终状态且后端资源被释放；通过 `go test ./...` 和 `git diff --check` --- 修改了 packages/backend/internal/ws/hub.go、packages/backend/internal/ws/hub_game_flow_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/hub.go packages/backend/internal/ws/hub_game_flow_test.go work.md`。

---

2026-06-12 15:52:26 +08:00 --- 发现 PRD 要求房主在开局前设置游戏参数 [Room Settings]，但后端只支持创建房间时传入 `maxPlayers` / `scriptId`，创建后无法调整；同时缺少对应显式命令 [Explicit Command]、房主权限校验 [Creator Authorization] 和 setup 阶段约束 [Setup Phase Constraint] --- 使用 `UPDATE_ROOM_SETTINGS` 消息贯通 Hub、GameSession 和 RoomManager：Hub 从连接身份 [Connection Identity] 推导房主权限，不信任客户端伪造 `playerId`；GameSession 只允许 setup 阶段更新，校验 `maxPlayers` 必须在 5-15 且不能低于当前实际玩家数 [Current Player Count]，校验脚本必须受支持；RoomManager 保存新的房间元数据 [Room Metadata] 并广播/持久化最新 `ROOM_STATE`；补充 GameSession、RoomManager、Hub 权限/流程回归测试，并通过 `go test ./...`、`git diff --check` --- 修改了 packages/backend/internal/ws/message.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/room_manager.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/room_manager_test.go、packages/backend/internal/ws/hub_game_flow_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/message.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/room_manager.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/room_manager_test.go packages/backend/internal/ws/hub_game_flow_test.go work.md`。

---

2026-06-12 16:22:05 +08:00 --- 发现后端已定义 Scarlet Woman 接恶魔 [Imp Starpass] 的胜负原因 [Win Reason]，但恶魔死亡后 `checkWinConditions` 会直接判好人胜利，没有在 5 名及以上玩家存活且 Scarlet Woman 存活时把 Scarlet Woman 转换为 Imp --- 使用胜负判断前置接棒逻辑 [Starpass Resolution]：当没有存活恶魔 [Alive Demon]、存在死亡恶魔 [Dead Demon]、存活人数不少于 5 且 Scarlet Woman 存活时，将 Scarlet Woman 的角色替换为 Imp，并让游戏继续；不广播公开角色分配事件 [Public Assignment Event]，依赖现有按接收者裁剪的 `ROOM_STATE` 展示；补充 Scarlet Woman 成功接棒和人数不足无法接棒的回归测试，并通过 `go test ./...`、`git diff --check` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go work.md`。

---

2026-06-12 16:24:15 +08:00 --- 发现后端夜晚结算 [Night Resolution] 对 `protect` 行动和 Soldier 免疫只记录动作但不生效，Imp 夜杀 [Imp Night Kill] 仍会杀死 Monk 保护目标或 Soldier --- 使用夜晚防护结算 [Night Protection Resolution]：在 `ResolveNight` 中先收集 Storyteller 裁决的 Monk 保护目标 [Protected Targets]，再处理 kill 行动；若目标被保护或角色是 Soldier，则跳过死亡记录 [Death Record] 和 `PlayerDied` 事件；补充 Monk 保护目标不死亡、Soldier 被夜杀不死亡的回归测试，并通过 `go test ./...`、`git diff --check` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go work.md`。

---

2026-06-12 16:30:37 +08:00 --- 发现后端尚未实现 Slayer 白天能力 [Slayer Ability]：玩家无法在白天发起一次性射击，服务端也没有持久化能力使用状态 [Ability Usage State]，重启后可能重置 --- 使用显式 `USE_SLAYER_ABILITY` 命令 [Explicit Command] 贯通 Hub 与 GameSession：Hub 从连接身份 [Connection Identity] 推导能力使用者，不信任客户端伪造的 `playerId`；GameSession 限制只有存活 Slayer 能在白天使用一次，命中恶魔则以能力死亡 [Ability Death] 结算并触发胜负判断，未命中只消耗使用次数；快照持久化 `slayerUsed`，补充会话层命中/未命中测试、Hub 伪造身份测试和 Redis 快照恢复断言，并通过 `go test ./...`、`git diff --check` --- 修改了 packages/backend/internal/ws/message.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/persistence_test.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/hub_game_flow_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/message.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/persistence.go packages/backend/internal/ws/persistence_test.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/hub_game_flow_test.go work.md`。

---

2026-06-12 16:35:25 +08:00 --- 发现 PRD 要求 Storyteller 宣告玩家死亡 [Manual Death Declaration]，但后端只在 GameSession 内部保留 `KillPlayerCmd`，WebSocket 没有显式入口，且命令本身缺少 Storyteller 权限 [Authorization]、阶段 [Phase] 和死因 [Death Cause] 校验 --- 使用 `KILL_PLAYER` 显式命令 [Explicit Command] 贯通消息解析、Hub 和 GameSession：客户端必须提供 `targetPlayerId` 与 `cause`，死因支持字符串/协议数字解析；Hub 从连接身份 [Connection Identity] 推导发送者，不信任伪造 `playerId`；GameSession 只允许 Storyteller 在已开局且未结束时宣告死亡，并记录死亡事件 [Death Event] 与死亡记录 [Death Record] 后走统一胜负判断和房间状态广播；补充死因解析、会话层权限/死因测试、Hub 正向路径和伪造身份拒绝测试，并通过 `go test ./...`、`git diff --check` --- 修改了 packages/backend/internal/ws/message.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/hub_game_flow_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/message.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/hub_game_flow_test.go work.md`。

---

2026-06-12 16:39:02 +08:00 --- 发现后端领域模型 [Domain Model] 已有 `NightAction.Result` 与 `NightActionEvent.Result`，但 `SUBMIT_NIGHT_ACTION` 客户端消息没有 `result` 字段，GameSession 也不会保存或回传 Storyteller 裁决结果 [Adjudicated Result]，导致信息类夜晚行动只能记录目标，不能记录结果 --- 使用 `result` 字段贯通 ClientMessage、Hub、SubmitNightActionCmd 与 GameSession：提交夜晚行动时修剪并保存结果文本，事件只在有结果时携带 `result` 指针；保留现有夜晚行动隐私广播 [Privacy Broadcast]，结果只发送给行动者和 Storyteller；补充会话层保存/事件测试、Hub 隐私测试和 Redis 快照恢复断言；初次测试失败暴露复用了已完成夜晚步骤的夹具 [Fixture]，将测试中的 `nightWakeIndex` 调整到 Imp 当前步骤后通过 `go test ./...`、`git diff --check` --- 修改了 packages/backend/internal/ws/message.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/backend/internal/ws/persistence_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/message.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/hub_game_flow_test.go packages/backend/internal/ws/persistence_test.go work.md`。

---

2026-06-12 16:57:44 +08:00 --- 发现后端已经支持房间设置更新 [Room Settings Update]、Slayer 白天能力 [Slayer Ability]、Storyteller 手动宣告死亡 [Manual Death Declaration] 与夜晚裁决结果 [Night Action Result]，但 core WebSocket 客户端 [WebSocket Client] 与 Taro 首页 [Taro Page] 还没有完整发送入口：前端只能创建时设置房间人数，Slayer 无法在页面发起能力，Storyteller 只能走旧的 `EXECUTE_PLAYER` 快捷处决，夜晚 `result` 只在本地显示不进入服务器状态 --- 使用端到端接线 [End-to-End Wiring] 补齐 `UPDATE_ROOM_SETTINGS`、`USE_SLAYER_ABILITY`、`KILL_PLAYER` 和 `SUBMIT_NIGHT_ACTION.result` 的 core 客户端类型、方法与发送测试；前端复用现有房间人数输入增加房主保存设置入口，给存活 Slayer 的白天页面增加目标按钮，给 Storyteller 控制区增加死因与目标选择后提交 `KILL_PLAYER`，并将夜晚结果随 `SUBMIT_NIGHT_ACTION` 发给后端而不是本地乐观追加；通过 `pnpm --filter @clocktower/core test -- websocket-client.test.ts`、`pnpm typecheck`、`pnpm test`、`go test ./...`、`pnpm build:frontend`、`pnpm build:core`、`git diff --check` --- 修改了 packages/core/src/websocket/index.ts、packages/core/src/websocket/__tests__/websocket-client.test.ts、packages/frontend/src/pages/index/index.tsx、packages/core/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/core/src/websocket/index.ts packages/core/src/websocket/__tests__/websocket-client.test.ts packages/frontend/src/pages/index/index.tsx packages/core/tsconfig.tsbuildinfo`，并从 `work.md` 删除本条 2026-06-12 16:57:44 记录。

---

2026-06-12 17:07:42 +08:00 --- 审查未提交代码 [Uncommitted Code Review] 时发现前端保存房间设置 [Room Settings Save] 会总是发送 `DEFAULT_SCRIPT_ID`，而后端在角色已分配 [Characters Assigned] 后会拒绝任何 `scriptId` 字段，导致房主只是保存玩家人数也可能被误判为脚本变更 [Script Change] --- 将 Taro 首页的保存设置动作改为只发送 `maxPlayers`，保留脚本切换给未来明确 UI；重新通过 `pnpm test`、`pnpm typecheck`、`go test ./...`、`pnpm build:frontend`、`pnpm --filter @clocktower/frontend exec taro build --type h5`、`pnpm build:core` --- 修改了 packages/frontend/src/pages/index/index.tsx、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/src/pages/index/index.tsx`，并从 `work.md` 删除本条 2026-06-12 17:07:42 记录。

---

2026-06-12 17:26:33 +08:00 --- 发现 Taro H5 监听构建 [H5 Watch Build] 在 Node.js 24 下启动后崩溃：`@tarojs/webpack5-prebundle` 调用 `webpack-virtual-modules` 的 `_writeVirtualFile`，但当前 webpack 输入文件系统 [Input File System] 不再暴露该方法，导致 `dev:h5` 无法进入联调 --- 使用 Taro webpack5 编译器对象配置 [Compiler Object Config] 显式禁用预构建 [Prebundle]：将 `compiler: 'webpack5'` 改为 `compiler.type = 'webpack5'` 且 `compiler.prebundle.enable = false`，避开该虚拟模块写入路径，保留 webpack5 构建链路 --- 修改了 packages/frontend/config/index.ts、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/config/index.ts`，并从 `work.md` 删除本条 2026-06-12 17:26:33 记录。

---

2026-06-12 17:47:01 +08:00 --- 发现 H5 dev server [H5 Development Server] 虽然能编译并提供 `/runtime.js`、`/app.js`，但根路径返回目录列表 [Directory Listing]，`/index.html` 为 404，导致浏览器看不到页面；进一步检查 `taro inspect --type h5 plugins` 发现项目缺少 `src/index.html` 时 Taro 4 不会注入 `HtmlWebpackPlugin`，同时 `@tarojs/plugin-platform-h5` 使用浮动版本范围可能把 H5 依赖漂移到 4.2.x --- 新增 Taro H5 HTML 模板 [HTML Template] `packages/frontend/src/index.html`，包含 `#app` 挂载点 [Mount Point] 和 `htmlWebpackPlugin.options.script`；将 `@tarojs/plugin-platform-h5` 精确锁定到 `4.0.0`，与 CLI、webpack runner、runtime 保持同版本线；重启 `dev:h5` 后确认根路径返回标题为“血染钟楼”的 HTML，并在应用内浏览器看到首页内容且无控制台错误 --- 修改了 packages/frontend/src/index.html、packages/frontend/package.json、pnpm-lock.yaml、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/package.json pnpm-lock.yaml work.md` 并删除 `packages/frontend/src/index.html`。

---

2026-06-12 18:02:09 +08:00 --- 发现 H5 首页 [H5 Home Page] 虽已接通完整 WebSocket 游戏功能，但信息和按钮按功能直铺，缺少首页摘要 [Summary]、清晰操作区 [Operation Panels]、空状态 [Empty State] 与基础图标 [Icons]，单设备多窗口联调时不易快速判断连接、房间、阶段、身份和玩家状态 --- 保留现有 WebSocket 行为与命令入口，重排首页为简单控制台 [Control Console]：增加顶部状态摘要、房间/身份/阶段/玩家统计，按连接、身份、房间、玩家、角色、Storyteller、提名、投票、死亡、夜晚和日志分区；给主要按钮增加文本图标 [Text Icons]；统一 CSS 为简洁工作台样式 [Workbench Style]，减少装饰并修复未加入房间时可误点“设自己为 Storyteller”的入口 --- 修改了 packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/index/index.css、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/index/index.css work.md`。

---

2026-06-12 18:20:57 +08:00 --- 发现 H5 页面和剧本页仍有英文/中英混排的可见文案 [UI Copy]，并且共享脚本数据 [Shared Script Data] 中的暗流涌动角色名、角色能力 [Ability Text] 与夜晚提示 [Night Prompt] 仍是英文，导致玩家看到的提示和角色描述不一致 --- 将 `TROUBLE_BREWING_SCRIPT` 的剧本名、22 个角色名、能力描述和首夜/后续夜晚唤醒提示全部改为中文；将 H5 首页阶段、连接状态、说书人、阵营、夜间行动、错误提示、操作日志和各区块标题中文化；将剧本页区块标题与角色分类标签改为中文，并增加常见服务端错误的中文显示映射 [Error Message Mapping]；运行类型检查 [Typecheck] 后同步已跟踪的增量编译文件 [Incremental Build File] --- 修改了 packages/core/src/scripts/index.ts、packages/core/tsconfig.tsbuildinfo、packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/scripts/index.tsx、packages/frontend/src/pages/scripts/utils.ts、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/core/src/scripts/index.ts packages/core/tsconfig.tsbuildinfo packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/scripts/index.tsx packages/frontend/src/pages/scripts/utils.ts work.md`。

---

2026-06-13 09:34:55 +08:00 --- 发现代码审查 [Code Review] 中指出 3 个终局相关问题：红唇女郎接魔 [Scarlet Woman Starpass] 使用恶魔死亡后的存活人数导致 5 人存活时误判善良胜利；终局角色揭示 [Role Reveal] 仍沿用普通玩家隐私快照导致非说书人看不到全员角色；普通玩家终局页没有返回大厅路径 [Return to Lobby] --- 在死亡路径中记录恶魔死亡前存活人数 [Pre-death Alive Count] 并传入胜利条件检查 [Win-condition Check]，修正接魔阈值判断；让已结束游戏或已有胜者的房间快照对所有接收者揭示角色；终局页所有玩家均显示返回大厅按钮，并清理本地房间状态；补充 5 人接魔、4 人不接魔与终局揭示测试；通过 `go test ./...`、`npm test`、`npm run typecheck` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、packages/frontend/src/pages/index/index.tsx、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go packages/frontend/src/pages/index/index.tsx work.md`。

---

2026-06-13 10:00:26 +08:00 --- 发现后端白天提名流程 [Day Nomination Flow] 仍缺少桌游规则中的每日提名限制 [Daily Nomination Limits]：同一天内同一玩家可以多次提名，且同一玩家可以被多次提名；同时该限制若只放内存会在服务器重启后丢失 --- 在 `GameSession` 中新增 `nominatorsToday` 与 `nomineesToday` 玩家标记映射 [Player Flag Maps]，提名成功时记录提名者与被提名者，进入夜晚或新白天时重置；将两个映射纳入快照保存/恢复 [Snapshot Save/Restore]；补充重复提名者、重复被提名者、新一天重置、快照恢复与 Redis 快照字段测试；通过 `go test ./...` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/persistence_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/persistence.go packages/backend/internal/ws/persistence_test.go work.md`。

---

2026-06-13 10:10:52 +08:00 --- 发现后端 Virgin 首次被提名能力 [Virgin First Nomination Ability] 尚未实现：Virgin 被镇民 [Townsfolk] 首次提名时不会立即处决提名者，且能力使用状态 [Ability Usage State] 在服务器重启后会丢失 --- 在提名流程 [Nomination Flow] 中记录 Virgin 能力使用状态，首次被提名时若提名者为镇民则立即以处决 [Execution] 方式杀死提名者并结束白天进入夜晚；非镇民首次提名只消耗能力并进入正常投票；将 `virginAbilityUsed` 纳入快照保存/恢复 [Snapshot Save/Restore]；补充 Virgin 镇民触发、非镇民消耗、快照恢复测试；通过 `go test ./...`、`git diff --check` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/persistence_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/persistence.go packages/backend/internal/ws/persistence_test.go work.md`。


---

2026-06-13 10:55:47 +08:00 --- 发现后端角色能力系统 [Character Ability System] 缺少核心地基：Poisoner 中毒状态 [Poisoned State] 的完整生命周期管理，包括状态设置、过期清理、隐私过滤和持久化支持；当前 Washerwoman/Librarian/Investigator/Chef/Empath/Fortune Teller 等信息类能力尚未实现结算逻辑 [Resolution Logic]，无法验证中毒效果 --- 使用最小垂直切片 [Minimal Vertical Slice] 在 `game.Player` 中新增 `PoisonedUntil *int32` 字段标记中毒到第几天黄昏；在 `applySubmitNightAction` 中检测 `NightActionPoison` 并立即设置目标玩家的 `PoisonedUntil = dayNumber + 1`；在 `applyChangePhase` 的 Day→Night 转换前调用 `clearExpiredPoisonLocked` 清理过期中毒；在 `stateForRoom` 中对所有非说书人接收者隐藏 `PoisonedUntil` 字段（包括被中毒者自己）；新增 6 个测试覆盖中毒设置、过期清理、说书人/玩家隐私、重复中毒和 Redis 快照持久化；通过 `go test ./internal/ws` 与 `go test ./internal/game` --- 修改了 packages/backend/internal/game/game.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/poison_test.go、packages/backend/internal/ws/persistence_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/game/game.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/persistence_test.go work.md`，并删除 `packages/backend/internal/ws/poison_test.go`。
---

2026-06-13 11:41:36 +08:00 --- 发现 Poisoner 中毒状态 [Poisoned State] 已写入后端 `game.Player`，但公开协议 [Public Protocol] 与核心 TypeScript 类型 [Core TypeScript Types] 尚未同步，导致说书人客户端 [Storyteller Client] 收到 `poisonedUntil` 后缺少类型契约 [Type Contract]；同时 `work.md` 文件末尾存在空白行导致 `git diff --check` 失败 --- 在 `proto/game.proto` 的 `Player` 消息中新增 `optional int32 poisoned_until = 6`，更新生成脚本 [Generation Script] 的干净接口 [Clean Interfaces]、核心手写 Player/RoomState 类型与 ProtoPlayer 类型测试，并运行 `pnpm proto:generate` 刷新生成类型 [Generated Types]；移除后端字段上的过期 TODO，清理 `work.md` 末尾空白 --- 修改了 proto/game.proto、scripts/generate-types.mjs、packages/core/src/types/generated/index.ts、packages/core/src/types/generated/__tests__/proto-types.test.ts、packages/core/src/types/index.ts、packages/core/src/websocket/index.ts、packages/backend/internal/game/game.go、work.md

撤回方式 [Rollback Strategy]：提交前执行 `git checkout -- proto/game.proto scripts/generate-types.mjs packages/core/src/types/generated/index.ts packages/core/src/types/generated/__tests__/proto-types.test.ts packages/core/src/types/index.ts packages/core/src/websocket/index.ts packages/backend/internal/game/game.go work.md`；若本次提交已成为最新提交，执行 `git revert HEAD`。


---

2026-06-13 11:16:32 +08:00 --- 发现后端信息类能力 [Information Abilities] 仍需要说书人手动输入结果 [Manual Result Input]，缺少自动计算 [Auto-computation] 支持；Empath 夜间信息是最简单的示例（统计存活邻居中的邪恶玩家数），可以作为能力自动结算 [Ability Resolution] 的第一个垂直切片 [Vertical Slice] --- 使用自动结算地基 [Auto-resolution Foundation] 在 `applySubmitNightAction` 中检测 `NightActionLearnEvilNeighbors` 且说书人未提供结果时，调用 `computeEmpathResultLocked` 自动计算；新增 `currentWakeCharacterIDLocked` 获取当前唤醒角色、`isPlayerEvilLocked` 判断玩家阵营；Empath 计算逻辑：找到 Empath 玩家索引，检查是否中毒（中毒时返回空结果由说书人提供假信息），计算循环邻居列表 [Circular Neighbor List] 中存活且邪恶的玩家数；新增 5 个测试覆盖正常计算、中毒不计算、手动覆盖、双邻居邪恶和死亡邻居不计数；通过 `go test ./internal/ws` 与 `go test ./internal/game` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/empath_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/empath_test.go`。

---

2026-06-13 12:19:14 +08:00 --- 发现 Poisoner 中毒状态 [Poisoned State] 已经通过后端房间快照 [Room Snapshot] 暴露给说书人客户端 [Storyteller Client]，但 Taro 首页 [Taro Page] 的玩家列表 [Player List] 没有消费 `poisonedUntil`，导致说书人无法在界面确认当前中毒持续到哪一天黄昏 --- 在主玩家列表中仅当当前用户是说书人 [Storyteller] 且玩家快照包含 `poisonedUntil` 时显示“中毒至第 N 天黄昏”状态；新增克制的中毒状态标记样式 [Status Badge]，不改变普通玩家隐私模型 [Privacy Model] --- 修改了 packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/index/index.css、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/index/index.css work.md`。

---

2026-06-13 14:06:47 +08:00 --- 发现后端信息类能力 [Information Abilities] 中 Empath 已支持自动计算 [Auto-computation]，但 Chef 的“邪恶相邻对数”[Adjacent Evil Pair Count] 仍需要说书人 [Storyteller] 手动输入，导致 Trouble Brewing 首夜信息结算 [First Night Information Resolution] 不完整 --- 使用同一自动结算路径 [Auto-resolution Path] 为 `NightActionLearnEvilPairs` 增加 Chef 结果计算：找到存活且未中毒 [Poisoned] 的 Chef，按座位环 [Seating Circle] 统计相邻邪恶玩家对，支持首尾相邻 [Circular Adjacency]；中毒 Chef 不自动给结果，手动输入结果优先；新增 Chef 测试覆盖普通相邻、首尾相邻、中毒抑制和手动覆盖；通过 `go test ./internal/ws`、`go test ./...`、`npm test`、`npm run typecheck` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/chef_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/chef_test.go`。

---

2026-06-13 14:14:39 +08:00 --- 发现 Fortune Teller 恶魔查验 [Demon Check] 仍依赖说书人 [Storyteller] 手动填写 `yes/no` 结果，虽然目标数量校验 [Target Count Validation] 与夜晚行动记录 [Night Action Recording] 已存在，首夜信息自动结算 [First Night Auto-resolution] 还缺少这一核心镇民能力 [Townsfolk Ability] --- 在信息能力自动结算路径 [Information Ability Auto-resolution Path] 中接入 `NightActionCheckDemon`：找到存活且未中毒 [Poisoned] 的 Fortune Teller，检查两个目标中是否有角色类型 [Character Type] 为恶魔 [Demon] 的玩家，有则返回 `yes`，否则返回 `no`；中毒 Fortune Teller 不自动给结果，手动结果优先；新增测试覆盖命中恶魔、未命中恶魔、中毒抑制和手动覆盖；通过 `go test ./internal/ws`、`go test ./...`、`npm test`、`npm run typecheck`、`git diff --check` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/fortune_teller_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/fortune_teller_test.go`。

---

2026-06-13 14:21:20 +08:00 --- 发现 Undertaker 夜晚信息 [Night Information] 仍需要说书人 [Storyteller] 手动填写，后端虽然已记录处决死亡 [Execution Death] 的 `DeathRecord`，但 `NightActionLearnExecuted` 没有自动从当天死亡记录 [Death Records] 中取出被处决玩家的角色 [Character] --- 在信息能力自动结算路径 [Information Ability Auto-resolution Path] 中接入 `NightActionLearnExecuted`：找到存活且未中毒 [Poisoned] 的 Undertaker，从当天 [Current Day] 最近一次处决死亡记录中返回被处决玩家角色名；当天无处决时返回 `none`，角色缺失时返回 `unknown`，手动结果优先；新增测试覆盖处决角色、当天无处决、中毒抑制和手动覆盖；通过 `go test ./internal/ws`、`go test ./...`、`npm test`、`npm run typecheck`、`git diff --check` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/undertaker_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/undertaker_test.go`。

---

2026-06-13 14:26:48 +08:00 --- 发现 Ravenkeeper 夜死触发 [Night-death Trigger] 仍需要说书人 [Storyteller] 手动填写结果，虽然后端夜晚顺序 [Night Wake Order] 已包含 `NightActionLearnDied`，但没有判断 Ravenkeeper 本夜是否真的会被恶魔击杀 [Demon Kill]，也没有自动返回所选玩家角色 [Chosen Player Character] --- 在信息能力自动结算路径 [Information Ability Auto-resolution Path] 中接入 `NightActionLearnDied`：找到存活且未中毒 [Poisoned] 的 Ravenkeeper，检查本夜已记录的击杀行动 [Kill Actions] 是否指向 Ravenkeeper 且没有被 Monk 保护 [Protection] 或 Soldier 免疫 [Immunity] 阻止；若会死亡则返回所选目标角色名，未死亡返回 `none`，目标角色缺失返回 `unknown`，手动结果优先；新增测试覆盖被夜杀自动揭示、未被夜杀、被保护、中毒抑制和手动覆盖；通过 `go test ./internal/ws`、`go test ./...`、`npm test`、`npm run typecheck`、`git diff --check` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/ravenkeeper_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/ravenkeeper_test.go`。

---

2026-06-13 14:32:44 +08:00 --- 发现信息能力自动结算 [Information Ability Auto-resolution] 中 Chef、Empath、Fortune Teller、Undertaker、Ravenkeeper 都重复实现了“找到存活角色 [Living Character] 并判断是否中毒 [Poisoned]”逻辑，后续继续补角色时容易出现不一致 --- 提取 `findLivingCharacterIndexLocked`、`playerIsPoisonedLocked` 与 `characterCanAutoResolveLocked` 三个辅助函数 [Helper Functions]，让现有自动结算函数复用同一角色可结算判断 [Auto-resolve Eligibility Check]，不改变外部行为 [External Behavior] --- 修改了 packages/backend/internal/ws/game_session.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`。

---

2026-06-13 14:48:21 +08:00 --- 发现 Trouble Brewing 首夜主要信息角色 [First-night Information Roles] 中 Washerwoman、Librarian、Investigator 仍只记录说书人 [Storyteller] 手动结果，缺少按目标玩家角色类型 [Character Type] 自动结算，导致最小可玩后端 [Minimum Playable Backend] 的首夜信息链路不完整 --- 新增通用角色类型提示结算 [Character Type Hint Resolution]：存活且未中毒 [Living and Unpoisoned] 时，从两个候选目标 [Candidate Targets] 中返回对应 Townsfolk/Outsider/Minion 的角色名；Librarian 在无 Outsider 在场时返回 `none`；无可判定目标返回 `unknown`；手动结果优先，中毒时不自动结算；新增测试覆盖 Washerwoman、Librarian、Investigator 的正常结果、无 Outsider、中毒抑制和手动覆盖 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/first_night_information_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/first_night_information_test.go`。

---

2026-06-13 14:54:43 +08:00 --- 发现 Butler 规则 [Butler Rule] 尚未落地：夜晚 `learn_master` 只记录行动，没有把选择的主人 [Master] 带到白天投票 [Day Voting]；同时中毒状态 [Poisoned State] 在白天 `PoisonedUntil == dayNumber` 时被视为已失效，早于黄昏 [Dusk] 清理时机 --- 新增 `butlerMasters` 会话状态 [Session State]，在 Butler 夜晚行动中记录主人并纳入快照保存/恢复 [Snapshot Save/Restore]；投票时若存活且未中毒的 Butler 要投赞成票 [Yes Vote]，必须等待主人已投赞成；反对票 [No Vote] 不受限制；将中毒判断改为持续到 `PoisonedUntil >= dayNumber`，与黄昏清理保持一致；新增测试覆盖主人未赞成拦截、主人赞成后放行、反对票放行、中毒 Butler 放行、缺少主人错误和快照恢复 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/persistence_test.go、packages/backend/internal/ws/butler_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/persistence.go packages/backend/internal/ws/persistence_test.go work.md`，并删除 `packages/backend/internal/ws/butler_test.go`。

---

2026-06-13 14:59:29 +08:00 --- 发现 Mayor 终局 [Mayor Endgame] 的触发点不够准确：通用胜负检查 [Generic Win-condition Check] 会在夜杀后刚剩 3 人且 Mayor 存活时过早判定善良胜利，但规则语义应是白天结束且没有处决 [No Execution] 时触发；同时 Saint 被处决邪恶胜利 [Saint Executed Evil Win] 缺少明确回归测试 --- 将 Mayor 胜利移动到 Day→Night 阶段转换 [Phase Transition] 前检查，只有 3 人存活、Mayor 存活、未中毒 [Unpoisoned] 且当天没有处决死亡时才结束游戏；从通用死亡后胜负检查中移除 Mayor 早触发；新增测试覆盖白天结束 Mayor 胜利、夜杀到 3 人不立即胜利、当天有处决时 Mayor 不胜利、中毒 Mayor 不胜利，以及 Saint 被处决时邪恶胜利 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/endgame_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/endgame_test.go`。

---

2026-06-13 15:04:50 +08:00 --- 发现 Imp 自杀传魔 [Imp Self-kill Starpass] 缺失：夜晚 Imp 选择击杀自己时，后端会把它当成普通恶魔死亡 [Demon Death]，若没有 Scarlet Woman 接魔路径就直接判善良胜利，导致最小可玩局 [Minimum Playable Game] 的关键恶魔规则不完整 --- 在夜晚击杀结算 [Night Kill Resolution] 中识别目标为当前 Imp 的自杀击杀；若存在存活爪牙 [Living Minion]，将第一个存活爪牙转换为新的 Imp，并阻止本次恶魔死亡触发善良胜利；若没有存活爪牙，则保留原有恶魔死亡胜利；新增测试覆盖有爪牙传魔继续游戏、无爪牙时善良胜利 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/endgame_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/endgame_test.go work.md`。

---

2026-06-13 15:12:24 +08:00 --- 发现 Spy 视野规则 [Spy Grimoire View] 尚未实现：非说书人 [Non-storyteller] 房间快照 [Room Snapshot] 永远只显示自己角色，导致 Spy 夜晚无法看到魔典 [Grimoire]，邪恶方信息体验不完整 --- 将房间快照的全量可见性判断 [Full Visibility Check] 收敛到 `recipientCanSeeAllLocked`，在夜晚阶段 [Night Phase] 允许存活且未中毒 [Alive and Unpoisoned] 的 Spy 查看全员角色与状态；白天或中毒时继续按普通玩家隐私隐藏其他角色；新增测试覆盖 Spy 夜晚全量可见、白天不可见、中毒夜晚不可见 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go work.md`。

---

2026-06-13 15:16:00 +08:00 --- 发现房间号生成 [Room ID Generation] 在 6 位房间号空间耗尽 [ID Space Exhaustion] 时会直接 `panic`，属于生产后端 [Production Backend] 的硬崩溃点 [Hard Crash]；Hub 创建房间 [Create Room] 路径也没有可恢复错误分支 --- 将 `generateRoomIDUnlocked` 改为返回错误 [Error Return]，`CreateRoom` 在耗尽时返回 `nil`，Hub 将其转换为 `ERROR` 消息而不是让进程崩溃；更新原 panic 测试为错误返回测试，并新增 `CreateRoom` 耗尽返回 nil 的回归测试 [Regression Test] --- 修改了 packages/backend/internal/ws/room_manager.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/room_manager_race_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/room_manager.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/room_manager_race_test.go work.md`。

---

2026-06-13 15:20:25 +08:00 --- 发现特殊登记角色 [Special Registration Roles] 的自动信息结算存在不确定性：Recluse 可能被登记为邪恶/爪牙/恶魔，Spy 可能被登记为善良/镇民/外来者；若后端自动给出确定结果，会替说书人 [Storyteller] 做规则裁定并可能产生错误信息 --- 新增特殊登记保护 [Registration Guard]：当 Washerwoman/Librarian/Investigator 的候选目标、Fortune Teller 的目标、Chef 的在场玩家或 Empath 的相邻玩家包含未中毒的 Spy/Recluse 时，不自动生成结果，返回空结果交给说书人手动填写；中毒的 Spy/Recluse 不触发登记不确定性；新增测试覆盖 Spy/Recluse 影响各类信息结算、以及中毒 Spy 不阻塞自动结果 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/registration_test.go、work.md

撤回方式 [Rollback Strategy]：执行 `git checkout -- packages/backend/internal/ws/game_session.go work.md`，并删除 `packages/backend/internal/ws/registration_test.go`。

---

2026-06-13 15:55:20 +08:00 --- 发现首夜流程 [First Night Flow] 缺少邪恶阵营信息 [Evil Team Information]：5 人局中爪牙 [Minion] 与恶魔 [Demon] 无法通过后端夜晚步骤互知身份，导致最小可玩局 [Minimum Playable Game] 的邪恶方基础信息不完整；现有夜晚步骤 [Night Wake Step] 只能按具体角色 ID 激活，无法表达“所有爪牙/恶魔”这种群体唤醒 [Group Wake Step] --- 为 `NightWakeStep` 增加可选角色类型 [Character Type] 字段，首夜新增“爪牙得知恶魔”和“恶魔得知爪牙”两个群体步骤；后端自动结算 `learn_demon` 与恶魔侧 `learn_minion` 的身份名单；同步核心脚本 [Core Script]、夜晚状态机 [Night State Machine]、前端唤醒列表 [Wake Order UI] 与行动按钮 [Action Selector]；迁移旧测试助手，新增邪恶互知回归测试 [Regression Test] --- 修改了 packages/backend/internal/game/game.go、packages/backend/internal/game/characters.go、packages/backend/internal/game/characters_test.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/evil_team_information_test.go、packages/backend/internal/ws/night_wake_test_helpers_test.go、packages/backend/internal/ws/butler_test.go、packages/backend/internal/ws/chef_test.go、packages/backend/internal/ws/empath_test.go、packages/backend/internal/ws/first_night_information_test.go、packages/backend/internal/ws/fortune_teller_test.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/backend/internal/ws/poison_test.go、packages/backend/internal/ws/ravenkeeper_test.go、packages/backend/internal/ws/undertaker_test.go、packages/backend/internal/ws/ws_test.go、packages/core/src/night-phase/index.ts、packages/core/src/night-phase/__tests__/night-phase.test.ts、packages/core/src/scripts/index.ts、packages/core/src/websocket/index.ts、packages/core/tsconfig.tsbuildinfo、packages/frontend/src/pages/index/index.tsx、packages/frontend/src/pages/scripts/index.tsx、packages/frontend/src/pages/scripts/utils.test.ts、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/game/game.go packages/backend/internal/game/characters.go packages/backend/internal/game/characters_test.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/butler_test.go packages/backend/internal/ws/chef_test.go packages/backend/internal/ws/empath_test.go packages/backend/internal/ws/first_night_information_test.go packages/backend/internal/ws/fortune_teller_test.go packages/backend/internal/ws/hub_game_flow_test.go packages/backend/internal/ws/poison_test.go packages/backend/internal/ws/ravenkeeper_test.go packages/backend/internal/ws/undertaker_test.go packages/backend/internal/ws/ws_test.go packages/core/src/night-phase/index.ts packages/core/src/night-phase/__tests__/night-phase.test.ts packages/core/src/scripts/index.ts packages/core/src/websocket/index.ts packages/core/tsconfig.tsbuildinfo packages/frontend/src/pages/index/index.tsx packages/frontend/src/pages/scripts/index.tsx packages/frontend/src/pages/scripts/utils.test.ts work.md`，并删除 `packages/backend/internal/ws/evil_team_information_test.go` 与 `packages/backend/internal/ws/night_wake_test_helpers_test.go`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 16:00:29 +08:00 --- 发现桌面 txt 中明确要求将 vless 与 tuic 节点 [Proxy Nodes] 加入 `2625_updated_manual_select_no_old_racknerd(1).yaml`，目标 yaml 已有 `自动选择` 锚点 [Anchor] 与 `手动选择` 别名 [Alias] 结构，需要同步更新代理定义和选择列表 --- 从 txt 中提取 `vless-reality-vision-ecs-Slbca` 与 `tuic5-ecs-Slbca` 两个节点，新增到目标 yaml 的 `proxies`，并加入 `自动选择` 与 `🌍选择代理节点` 列表；未导入 txt 中的 vmess、hysteria2、anytls 节点 --- 修改了 C:/Users/Qilia/Desktop/2625_updated_manual_select_no_old_racknerd(1).yaml、work.md

撤回方式 [Rollback Strategy]：从 `C:/Users/Qilia/Desktop/2625_updated_manual_select_no_old_racknerd(1).yaml` 删除本次新增的 `vless-reality-vision-ecs-Slbca` 与 `tuic5-ecs-Slbca` 两个 `proxies` 条目，并从 `自动选择`、`🌍选择代理节点` 两个列表删除同名列表项；若只撤回审计记录 [Audit Log]，执行 `git checkout -- work.md`。

---

2026-06-13 16:31:35 +08:00 --- 发现酒鬼 [Drunk] 只有角色定义 [Character Definition]，后端无法同时表达真实身份 [Actual Character] 与玩家看到的身份 [Shown Character]；夜晚唤醒 [Night Wake] 不会为酒鬼展示的镇民 [Shown Townsfolk] 创建步骤，自动信息结算 [Auto-resolution] 也没有统一的能力失效 [Ability Malfunction] 抽象；另外僧侣 [Monk] 中毒后仍可能通过保护动作生效 --- 为玩家状态 [Player State] 与角色分配事件 [Character Assignment Event] 增加 `shownCharacter`，角色分配命令 [Assignment Command] 增加 `shownCharacters` 并校验酒鬼必须展示未实际分配的镇民；房间视图 [Room View] 对说书人展示真实与展示身份，对酒鬼本人只展示假身份；夜晚步骤把酒鬼展示的镇民纳入唤醒，但自动结果保持为空；新增统一能力失效判断，覆盖信息角色、圣女 [Virgin]、管家 [Butler]、杀手 [Slayer]、僧侣 [Monk]、士兵 [Soldier]、市长 [Mayor]；同步 ProtoBuf、core WebSocket 类型、前端显示与生成类型 --- 修改了 packages/backend/internal/game/game.go、packages/backend/internal/game/characters.go、packages/backend/internal/game/characters_test.go、packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/drunk_test.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/message.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/ravenkeeper_test.go、proto/game.proto、scripts/generate-types.mjs、packages/core/src/types/index.ts、packages/core/src/types/generated/index.ts、packages/core/src/state-machine/index.ts、packages/core/src/websocket/index.ts、packages/core/tsconfig.tsbuildinfo、packages/frontend/src/pages/index/index.tsx、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/game/game.go packages/backend/internal/game/characters.go packages/backend/internal/game/characters_test.go packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/message.go packages/backend/internal/ws/persistence.go packages/backend/internal/ws/ravenkeeper_test.go proto/game.proto scripts/generate-types.mjs packages/core/src/types/index.ts packages/core/src/types/generated/index.ts packages/core/src/state-machine/index.ts packages/core/src/websocket/index.ts packages/core/tsconfig.tsbuildinfo packages/frontend/src/pages/index/index.tsx work.md`，并删除 `packages/backend/internal/ws/drunk_test.go`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 16:47:37 +08:00 --- 发现占卜师 [Fortune Teller] 自动信息结算 [Auto-resolution] 没有红鲱鱼 [Red Herring] 模型：目标不含恶魔 [Demon] 时会直接返回 `no`，但规则要求一个善良玩家 [Good Player] 对占卜师登记为恶魔，缺少该配置会让后端给出错误确定信息 --- 为角色分配命令 [Assignment Command] 增加 `fortuneTellerRedHerringId`，校验红鲱鱼只能在占卜师在场时指定，且必须是非占卜师的善良玩家；会话快照 [Session Snapshot] 持久化红鲱鱼；占卜师自动结算在目标包含恶魔或红鲱鱼时返回 `yes`，若未配置红鲱鱼且目标不含恶魔则返回空结果交给说书人 [Storyteller] 手动裁定；前端示例分配 [Sample Assignment] 在含占卜师时自动选择一个合法红鲱鱼 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/fortune_teller_test.go、packages/backend/internal/ws/hub.go、packages/backend/internal/ws/message.go、packages/backend/internal/ws/persistence.go、packages/backend/internal/ws/persistence_test.go、packages/core/src/websocket/index.ts、packages/frontend/src/pages/index/index.tsx、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/fortune_teller_test.go packages/backend/internal/ws/hub.go packages/backend/internal/ws/message.go packages/backend/internal/ws/persistence.go packages/backend/internal/ws/persistence_test.go packages/core/src/websocket/index.ts packages/frontend/src/pages/index/index.tsx work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 16:55:52 +08:00 --- 发现一次性/被动胜负能力 [Passive Win/Loss Abilities] 的能力失效 [Ability Malfunction] 处理不完整：中毒圣女 [Poisoned Virgin] 第一次被提名不会消耗“第一次”条件，中毒圣徒 [Poisoned Saint] 被处决仍可能让邪恶胜利，且中毒猩红女郎 [Poisoned Scarlet Woman] 仍会在恶魔死亡时接魔 --- 调整提名流程 [Nomination Flow]，圣女第一次被提名总会记录能力已检查，但只有未失效且提名者为镇民 [Townsfolk] 时才处决；胜负检查 [Win Check] 只在处决当天且圣徒能力未失效时触发邪恶胜利，避免毒性过期后追溯触发；猩红女郎接魔前检查自身能力未失效；新增回归测试 [Regression Tests] 覆盖三种中毒场景 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/endgame_test.go、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/endgame_test.go work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 17:00:22 +08:00 --- 发现夜晚恶魔击杀 [Demon Night Kill] 没有检查小恶魔 [Imp] 是否中毒或能力失效 [Ability Malfunction]；毒药师 [Poisoner] 当晚先毒小恶魔后，后端仍会在结算夜晚 [Resolve Night] 时杀死目标，破坏 5 人局核心夜晚规则 --- 在夜晚击杀结算前新增 `nightDemonCanKillLocked`，只有存活且能力未失效的小恶魔才能造成夜晚击杀；保留说书人提交动作 [Submitted Action] 记录，但不产生死亡事件 [Death Event] 或死亡记录 [Death Record]；新增回归测试覆盖毒小恶魔后击杀目标不死亡 --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/poison_test.go、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/poison_test.go work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 17:10:40 +08:00 --- 发现市长 [Mayor] 被恶魔夜晚击杀 [Demon Night Kill] 时，后端会像普通玩家一样自动死亡，缺少规则中的死亡可能转移 [Death Redirection] 裁定路径；同时需要证明中毒/失效市长 [Poisoned or Malfunctioning Mayor] 不应阻止死亡 --- 在夜晚击杀阻止逻辑 [Night Kill Prevention] 中将未失效市长视为自动夜杀被阻止，把实际死亡留给说书人 [Storyteller] 通过手动死亡命令 [Manual Kill Command] 裁定；新增回归测试 [Regression Tests] 覆盖市长自动夜杀不死、中毒市长会死、说书人可手动把夜杀转移到其他目标；已通过 `go test ./...` 与 `go test -race ./internal/ws` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/endgame_test.go、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/endgame_test.go work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 17:16:33 +08:00 --- 发现夜晚动作输入校验 [Night Action Input Validation] 只检查目标数量 [Target Count]、重复目标 [Duplicate Targets] 与目标存在性 [Target Existence]，没有阻止僧侣 [Monk] 保护自己或管家 [Butler] 选择自己为主人，导致非法规则动作可进入生产会话状态 [Production Session State] --- 在 `validateNightTargetsLocked` 后增加自选目标校验 [Self-target Validation]，通过可见角色 [Visible Character] 找到当前唤醒角色，拒绝 `monk cannot protect themself` 与 `butler cannot choose themself as master`；新增回归测试 [Regression Tests] 覆盖两类非法动作，并修正 Ravenkeeper 测试夹具 [Test Fixture] 中过去用于跳步的非法僧侣自保动作；已通过 `go test ./...` 与 `go test -race ./internal/ws` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/game_session_test.go、packages/backend/internal/ws/butler_test.go、packages/backend/internal/ws/ravenkeeper_test.go、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/game_session_test.go packages/backend/internal/ws/butler_test.go packages/backend/internal/ws/ravenkeeper_test.go work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 17:26:54 +08:00 --- 发现 Trouble Brewing 首夜顺序 [First-night Wake Order] 错误包含小恶魔 [Imp] 击杀步骤，导致最小 5 人局 [Five-player Game] 首夜会自动进入死亡/幽灵票 [Ghost Vote] 状态，偏离“小恶魔除首夜外每晚杀人”的核心规则；后端、core 与前端脚本视图 [Script View] 使用平行脚本定义，需要同步修正 --- 从后端 `TroubleBrewingFirstNightOrder` 与 core `TROUBLE_BREWING_FIRST_NIGHT_ORDER` 移除 Imp 首夜 `kill` 步骤，更新 Imp 能力文案 [Ability Text]；调整 Hub/WebSocket/角色测试与测试夹具 [Test Fixtures]，让首夜只结算信息，真正夜杀场景进入第二夜 [Second Night]；持久化恢复测试改为首夜后由说书人 [Storyteller] 手动制造死亡以继续覆盖死亡快照 [Death Snapshot]；已通过 `go test ./...`、`go test -race ./internal/ws`、`pnpm test`、`pnpm typecheck` --- 修改了 packages/backend/internal/game/characters.go、packages/backend/internal/game/characters_test.go、packages/backend/internal/ws/hub_game_flow_test.go、packages/backend/internal/ws/ws_test.go、packages/backend/internal/ws/poison_test.go、packages/backend/internal/ws/undertaker_test.go、packages/backend/internal/ws/ravenkeeper_test.go、packages/core/src/scripts/index.ts、packages/core/src/night-phase/__tests__/night-phase.test.ts、packages/core/tsconfig.tsbuildinfo、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/game/characters.go packages/backend/internal/game/characters_test.go packages/backend/internal/ws/hub_game_flow_test.go packages/backend/internal/ws/ws_test.go packages/backend/internal/ws/poison_test.go packages/backend/internal/ws/undertaker_test.go packages/backend/internal/ws/ravenkeeper_test.go packages/core/src/scripts/index.ts packages/core/src/night-phase/__tests__/night-phase.test.ts packages/core/tsconfig.tsbuildinfo work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 17:34:28 +08:00 --- 发现非说书人 [Non-storyteller] 提交夜晚动作 [Night Action] 时只校验玩家存在与存活，未校验当前唤醒步骤 [Current Wake Step]、动作类型 [Action Type]、目标数量 [Target Count] 与目标合法性 [Target Validity]，导致任意存活玩家可以提交 `kill` 等非法私密动作并广播给说书人，污染会话状态 [Session State] --- 将玩家提交动作也绑定到当前夜晚唤醒步骤：动作类型必须匹配当前步骤，提交者必须匹配该步骤的角色或阵营类型 [Character or Character Type]，并复用目标校验 [Target Validation]；保留合法私密动作只通知提交者与说书人的隐私边界 [Privacy Boundary]；新增 Hub 回归测试 [Regression Tests] 覆盖非当前角色提交、错误动作类型、缺失目标三类拒绝场景；已通过 `go test ./...` 与 `go test -race ./internal/ws` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub_game_flow_test.go、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub_game_flow_test.go work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

2026-06-13 17:38:04 +08:00 --- 发现 `EXECUTE_PLAYER` 处决命令 [Execution Command] 可以在夜晚 [Night Phase] 直接执行，绕过白天提名/投票流程 [Day Nomination/Vote Flow]；项目已有 `KILL_PLAYER` 手动死亡命令 [Manual Death Command] 可承载说书人裁定死亡，夜晚处决会混淆死亡原因 [Death Cause] 与阶段规则 [Phase Rules] --- 将 `ExecutePlayerCmd` 收紧为只能在白天阶段 [Day Phase] 使用，夜晚或其他阶段返回 `players can only be executed during the day phase`；遗留 `executePlayerId` 测试先切换到白天再验证兼容字段，断线重连死亡快照 [Reconnect Death Snapshot] 改用 `KILL_PLAYER`；新增夜晚处决拒绝回归测试 [Regression Test]；已通过 `go test ./...` 与 `go test -race ./internal/ws` --- 修改了 packages/backend/internal/ws/game_session.go、packages/backend/internal/ws/hub_game_flow_test.go、work.md

撤回方式 [Rollback Strategy]：提交前可执行 `git checkout -- packages/backend/internal/ws/game_session.go packages/backend/internal/ws/hub_game_flow_test.go work.md`；提交后使用 `git revert <commit>` 撤回整个切片。

---

## 2026-06-13 18:30 — 前端页面设计方案评审 [Frontend Page Design Review]

### 问题
发现当前前端 `index.tsx` 单页面 1517 行代码，所有游戏状态（setup/day/voting/night/finished）通过条件渲染切换，缺少清晰的页面架构、导航机制、状态管理和组件复用策略，导致代码可维护性差、团队协作困难、UI 一致性弱。

### 解决方案
通过 `/grill-me` 技能 [Skill] 深度拷问前端页面设计的 10 个核心问题，确定最终架构方案：

#### 1. 页面架构 [Page Architecture]
- **多页面架构 [Multi-Page Architecture]**：4 个核心页面 + 1 个辅助页面
  - `/pages/lobby` — 大厅（连接服务、创建/加入房间）
  - `/pages/game-setup` — 游戏准备（设置说书人、分配角色）
  - `/pages/game-play` — 游戏进行（day/voting/night 组件切换）
  - `/pages/game-over` — 游戏结束（结果展示、角色揭示）
  - `/pages/scripts` — 剧本查看（已存在）

#### 2. 导航机制 [Navigation Mechanism]
- **混合模式 [Hybrid Mode]**：
  - 页面入口校验 + 降级 UI（状态不匹配时显示 ErrorState 组件）
  - WebSocket 关键事件自动跳转（gameOver → game-over 页面）
  - 用户操作乐观跳转（点击"开始游戏"后立即跳转）

#### 3. 状态管理 [State Management]
- **Zustand 全局 Store**：
  - 全局状态：client、connectionStatus、roomState、gamePhase、isStoryteller、myCharacter
  - 页面本地状态：UI 临时数据（输入框值、错误提示、日志）
  - 状态分层设计避免过度全局化

#### 4. game-play 页面内部结构 [Game-Play Internal Structure]
- **子组件拆分 [Component Splitting]**：
  - 主容器 `index.tsx`（150 行）
  - `components/PhaseIndicator.tsx` — 阶段指示器
  - `components/PlayerList/` — 玩家列表变体（Setup/Game/Storyteller/GameOver）
  - `components/DayPhase/` — 白天阶段组件
  - `components/VotingPhase/` — 投票阶段组件
  - `components/NightPhase/` — 夜间阶段组件（说书人/玩家）

#### 5. 说书人控制面板布局 [Storyteller Panel Layout]
- **固定栏 + 抽屉混合 [Fixed Bar + Drawer]**：
  - 固定顶栏：阶段切换按钮（始终可见）
  - 滚动区域：玩家列表、唤醒顺序、行动选择器
  - 固定底栏：当前阶段主操作（提交行动/结束夜晚）+ 更多按钮
  - 抽屉：低频操作（宣告死亡、结束游戏、房间设置）

#### 6. 夜间唤醒顺序交互 [Night Wake Order Interaction]
- **折叠列表 + 自动聚焦 [Collapsible List + Auto-focus]**：
  - 已完成步骤折叠、当前步骤高亮、下一步预览、剩余步骤折叠
  - 智能过滤"未在场"步骤
  - 快捷跳过功能（角色未使用能力时）
  - 自动滚动到当前步骤

#### 7. 玩家列表展示层级 [Player List Display Hierarchy]
- **基础组件 + 变体组件 [Base + Variants]**：
  - `PlayerItem.tsx` — 基础组件（昵称、ID、存活状态）
  - `PlayerListSetup.tsx` — Setup 阶段（说书人标记、踢出按钮）
  - `PlayerListGame.tsx` — 游戏中（普通玩家视角）
  - `PlayerListStoryteller.tsx` — 游戏中（说书人视角，显示所有信息）
  - `PlayerListGameOver.tsx` — 结束阶段（揭示所有角色）
  - 提取共享逻辑到自定义 hooks（usePlayerInfo）

#### 8. 状态同步与容错 [State Sync & Fault Tolerance]
- **降级显示 + 部分自动跳转 [Graceful Degradation + Partial Auto-navigation]**：
  - 统一 ErrorState 组件（icon、message、hint、redirectUrl）
  - 页面入口校验（未加入房间、状态不匹配、连接断开）
  - 关键事件自动跳转，异常情况手动引导
  - WebSocket 消息处理器直接触发导航

#### 9. 基础组件库 [Base Component Library]
- **CSS 变量 + 组件封装 [CSS Variables + Component Wrapper]**：
  - **优先级 P0**：Button、Card、ErrorState、Tag
  - **优先级 P1**：Input、Loading、Divider
  - CSS 变量系统：`--color-*`、`--spacing-*`、`--radius-*`、`--shadow-*`
  - 统一样式、类型安全、行为封装

#### 10. 关键设计原则 [Key Design Principles]
1. ✅ 多页面架构 — 职责分离、心智清晰
2. ✅ Zustand 全局状态 — WebSocket 连接跨页面持久化
3. ✅ 组件化拆分 — 单文件 200-300 行，易于维护
4. ✅ 混合导航 — 关键事件自动跳转，异常情况降级 UI
5. ✅ 说书人 UX 优化 — 固定栏 + 抽屉 + 折叠列表
6. ✅ 基础组件库 — 统一样式、类型安全、便于复用
7. ✅ 智能过滤与快捷操作 — 减少说书人操作负担
8. ✅ 防御性编程 — 页面入口校验，容错机制完善

### 文件估算 [File Estimation]
- **页面**：5 个（lobby、game-setup、game-play、game-over、scripts）
- **页面组件**：~15 个（game-play 拆分出的子组件）
- **基础组件**：~8 个（Button、Card、ErrorState、Tag、Input、Loading、Divider、Badge）
- **Hooks**：~5 个（usePlayerInfo、useGameActions、useNavigationGuard 等）
- **总文件数**：~40 个（包括 .tsx + .css）

### 修改文件
- `work.md` — 新增前端页面设计方案评审记录

### 撤回方式 [Rollback Strategy]
执行 `git checkout -- work.md` 并从 work.md 中删除本条 2026-06-13 18:30 记录；前端页面设计方案尚未实施，暂无代码变更需要回滚。


## 2026-07-10 14:17 --- 架构深化机会缺少跨上下文统一视图 --- 使用领域上下文、删除测试 [Deletion Test] 与并行分包探索生成架构评审 --- 修改临时 HTML 报告与 `work.md`

### 发现什么问题
- Backend 的 Hub、RoomManager、GameSession 之间存在权威游戏会话所有权分散、命令编排重复与双写补偿。
- Frontend 页面同时承担实时消息投影、身份持久化、连接生命周期与游戏 UI 状态，接口 [Interface] 接近实现 [Implementation] 复杂度。
- Core 的 Game State Machine 与 Vote、Death、Win Conditions 存在平行状态和重复规则路径。
- ProtoBuf、生成 TypeScript 类型、手写 TypeScript/Go 类型的事实来源所有权不明确。
- Script、Night Phase 与 Backend 角色规则之间存在 Trouble Brewing 知识泄漏。

### 使用什么方式解决
- 读取 `CONTEXT-MAP.md`、三个包的 `CONTEXT.md` 与架构词汇规范；确认当前无 ADR 冲突。
- 并行探索 Core、Frontend、Backend，应用删除测试 [Deletion Test]，按深度 [Depth]、杠杆 [Leverage]、局部性 [Locality] 合并为五个候选。
- 在系统临时目录生成包含 Before/After 图、推荐强度与首要推荐的 HTML 报告；未提出具体 interface 设计，未修改生产代码。
- Backend 探索基线通过 `go test ./...` 与 `go test -race ./internal/ws`。

### 修改了哪些文件
- `C:\Users\Qilia\AppData\Local\Temp\architecture-review-20260710-141729.html` — 新增统一架构评审报告。
- `work.md` — 新增本次架构评审审计记录。

### 撤回方式 [Rollback Strategy]
- 删除临时报告：`Remove-Item -LiteralPath 'C:\Users\Qilia\AppData\Local\Temp\architecture-review-20260710-141729.html'`。
- 撤回日志：从 `work.md` 删除标题以 `## 2026-07-10 14:17` 开头的本节；或在确认没有其他未提交日志后执行 `git checkout -- work.md`。
## 2026-07-10 14:35 --- 后端权威游戏会话缺少成员与连接生命周期不变量 --- 将已确认决策写入 Backend 领域上下文 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 现有领域文档未区分房间成员 [Room Member] 与活动连接 [Active Connection]，无法表达断线保留成员和同一 `playerId` 连接接管语义。

### 使用什么方式解决
- 定义权威游戏会话 [Authoritative Game Session] 为成员关系、说书人、房间设置、重连资格与完整游戏状态的唯一所有者。
- 记录断线不移除成员、同一 `playerId` 可重连、新连接替换并关闭旧连接、仅显式离开或踢出撤销成员资格。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增会话所有权领域术语与四条业务规则。
- `work.md` — 新增本次设计决策审计记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 `Session Ownership` 小节及业务规则 5-8。
- 从 `work.md` 删除标题以 `## 2026-07-10 14:35` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 14:40 --- 命令提交缺少持久化失败语义 --- 将持久化纳入事务式提交 [Transactional Commit] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 现有规则只声明有效事件需要持久化，没有规定内存提交、Snapshot persistence 与广播的顺序，也没有规定持久化失败后的状态。

### 使用什么方式解决
- 确认命令仅在结果状态持久化成功后才算成功。
- 确认持久化失败时恢复命令前的内存状态，并向调用者返回错误。
- 确认广播只能发生在持久化成功后，客户端不得观察到未持久化状态。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增事务式提交、回滚与广播顺序规则。
- `work.md` — 新增本次失败语义决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 9-11。
- 从 `work.md` 删除标题以 `## 2026-07-10 14:40` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 14:45 --- 已提交状态缺少广播失败与恢复语义 --- 采用提交不可回滚与完整快照恢复 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 持久化成功后若部分客户端广播失败，现有领域规则未说明命令结果、连接状态和遗漏事件的恢复方式。

### 使用什么方式解决
- 确认已持久化命令不会因广播失败而回滚或改判失败。
- 广播失败的活动连接 [Active Connection] 会失效并关闭，但房间成员 [Room Member] 保留。
- 玩家以相同 `playerId` 重连后接收最新完整持久化状态，不补发遗漏的增量事件。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增广播失败、连接失效与完整快照恢复规则。
- `work.md` — 新增本次恢复语义决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 12-14。
- 从 `work.md` 删除标题以 `## 2026-07-10 14:45` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 14:50 --- 并发命令缺少排序与隔离模型 --- 采用每房间串行化 [Per-Room Serialization] 与异步广播 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 同一房间并发命令的读取顺序、持久化失败后的后继状态，以及广播是否占用命令执行序列均未定义。

### 使用什么方式解决
- 同一权威游戏会话 [Authoritative Game Session] 按服务器接收顺序串行处理命令，不同会话可并行。
- 后继命令只读取前一条成功持久化的状态；失败命令不推进状态。
- 提交后生成不可变广播负载 [Immutable Broadcast Payload]，广播异步执行且不阻塞后续命令。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增命令排序、会话隔离与广播解耦规则。
- `work.md` — 新增本次并发模型决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 15-18。
- 从 `work.md` 删除标题以 `## 2026-07-10 14:50` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 14:55 --- 房间生命周期与游戏命令可能落入不同并发序列 --- 统一进入权威会话命令序列并隔离连接接管 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- Join、Leave、Kick、房间设置、说书人指派若不与游戏命令共享排序，会导致各命令读取不同成员集合并重新产生双写竞态 [Dual-Write Race]。

### 使用什么方式解决
- 所有改变房间成员、设置或游戏状态的命令统一进入每会话串行命令序列。
- 新连接接管 [Connection Takeover] 仅原子替换活动连接，不改变权威状态，因此由传输 adapter 处理且不触发 Snapshot persistence。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增生命周期命令排序与连接接管例外规则。
- `work.md` — 新增本次命令边界决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 19-20。
- 从 `work.md` 删除标题以 `## 2026-07-10 14:55` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:00 --- 游戏结束、空房间与房间销毁语义混合 --- 引入显式关闭房间 [Close Room] 生命周期 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 当前设计可能在游戏结束或最后成员离开时销毁房间，导致最终状态无法重连查看，也让网络或操作失误造成不可恢复的数据删除。

### 使用什么方式解决
- 区分已结束游戏与已关闭房间 [Closed Room]：游戏结束后会话继续保留并持久化。
- 仅显式 Close Room 命令可从会话注册表 [Session Registry] 与 Snapshot 原子删除房间。
- 空房间不会自动删除；本轮不引入超时清理 [TTL Cleanup]。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Closed Room 术语及房间终止规则。
- `work.md` — 新增本次房间生命周期决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Closed Room 定义与业务规则 21-24。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:00` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:05 --- Close Room 权限与创建者身份缺少持久语义 --- 定义房间创建者 [Room Creator] 的独占权限 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 未定义说书人是否可关闭房间、创建者离开后的权限、创建者能否被踢出，以及创建者身份丢失后的恢复策略。

### 使用什么方式解决
- 定义房间创建者 [Room Creator] 为持久身份，并独占 Close Room 权限。
- 创建者断线或离开成员关系后仍保留关闭权限，且不能成为 Kick Player 目标。
- 本轮不支持所有权转移或管理员恢复；说书人身份不授予关闭权限。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Room Creator 术语及关闭权限规则。
- `work.md` — 新增本次权限决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Room Creator 定义与业务规则 25-27。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:05` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:10 --- 公开 playerId 可被用于身份与房间权限劫持 --- 引入服务器恢复凭证 [Resume Credential] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 客户端提供的公开 `playerId` 无法证明身份控制权；攻击者可借此接管活动连接、恢复创建者身份并执行 Close Room。

### 使用什么方式解决
- 首次创建房间成员身份时由服务器签发不可猜测的恢复凭证 [Resume Credential]。
- 重连和新连接接管必须同时验证 `playerId` 与恢复凭证。
- 凭证随权威游戏会话持久化，仅返回其所有者，禁止向其他成员广播。
- 凭证丢失视为身份控制权丢失，本轮不提供恢复机制。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Resume Credential 术语及身份验证规则。
- `work.md` — 新增本次鉴权决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Resume Credential 定义与业务规则 28-31。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:10` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:15 --- 自动轮换恢复凭证可能因响应丢失锁死身份 --- 固定凭证生命周期并延后撤销机制 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 若每次重连或连接接管自动轮换恢复凭证 [Resume Credential]，持久化成功但响应丢失会导致客户端永久失去新凭证。

### 使用什么方式解决
- 恢复凭证在房间成员 [Room Member] 身份生命周期内保持稳定，重连和连接接管不自动轮换。
- 显式凭证轮换与撤销机制不属于本轮生命周期。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增恢复凭证稳定性与延后范围规则。
- `work.md` — 新增本次安全取舍决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 32-33。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:15` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:20 --- Join Room 可泄漏成员存在性并混淆新加入与重连 --- 统一凭证错误并固定 playerId 语义 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 已占用 `playerId` 的 Join Room 请求若根据成员是否存在返回不同错误，会产生成员枚举 [Member Enumeration]；自动替换 `playerId` 也会破坏客户端身份语义。

### 使用什么方式解决
- 已存在 `playerId` 只有在恢复凭证 [Resume Credential] 正确时才作为重连或连接接管处理。
- 凭证缺失或错误统一返回不暴露成员存在性的凭证错误。
- 服务器不自动生成替代 `playerId`；新成员必须提交房间内未占用的身份。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Join Room 身份冲突与防枚举规则。
- `work.md` — 新增本次身份冲突决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 34-36。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:20` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:25 --- 主动离开与被踢后的重新加入资格未区分 --- 引入保留身份 [Retained Identity] 与房间封禁 [Room Ban] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- Leave Room 与 Kick Player 若都只删除成员，将无法表达自愿离开后恢复身份以及被踢后禁止重新加入的不同业务意图。

### 使用什么方式解决
- 主动离开删除成员资格但保留身份记录和原恢复凭证，可使用相同 `playerId` 再次加入。
- 被踢出会删除成员、撤销凭证并创建房间级封禁 [Room Ban]；本轮不支持解除封禁。
- 关闭房间时统一删除成员、保留身份、凭证和封禁记录。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Retained Identity、Room Ban 术语及重新加入规则。
- `work.md` — 新增本次成员资格决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Retained Identity、Room Ban 定义与业务规则 37-40。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:25` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:30 --- 游戏开始后成员变更会破坏角色与胜负不变量 --- 冻结参与者集合 [Participant Set] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 游戏开始后加入、重新加入、主动离开或踢出会改变角色分配、原始玩家数与胜负规则依赖的参与者集合。

### 使用什么方式解决
- 游戏开始时永久冻结参与者集合 [Participant Set]，游戏结束后仍保持冻结。
- 冻结后禁止新玩家、保留身份重新加入、主动离开与踢出。
- 现有参与者仍可断线、重连和连接接管，因为这些操作不改变参与者集合。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Participant Set 术语及冻结规则。
- `work.md` — 新增本次游戏完整性决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Participant Set 定义与业务规则 41-43。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:30` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:35 --- 离开的创建者在参与者冻结后无法重新加入关闭房间 --- 引入分离式管理命令 [Detached Management Command] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 房间创建者在游戏开始前主动离开后会成为保留身份；参与者集合冻结后无法重新加入，但仍需保有关闭房间的可达路径。

### 使用什么方式解决
- 允许离开的创建者凭 `roomId + playerId + Resume Credential` 执行分离式 Close Room。
- 该命令不恢复成员资格，也不授予其他房间设置或游戏命令权限。
- Close Room 仍进入房间串行命令序列，并以事务式持久化删除完成。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Detached Management Command 术语及关闭房间例外规则。
- `work.md` — 新增本次可达性决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Detached Management Command 定义与业务规则 44-46。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:35` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:40 --- 全局 Snapshot 与跨房间并行提交存在丢失更新风险 --- 改为每房间记录 [Per-Room Record] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 单一全局 Snapshot 在不同房间并行执行读取与写回时可能覆盖彼此更新，违背跨房间并行处理模型。

### 使用什么方式解决
- 每个权威游戏会话独立持久化为房间记录 [Room Record]。
- Create Room 创建记录，成功命令原子替换记录，Close Room 删除记录；启动时枚举恢复全部记录。
- Redis 使用每房间 Key；File adapter 即使共用物理文件，也必须提供按房间原子更新语义。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Room Record 术语及每房间持久化规则。
- `work.md` — 新增本次持久化所有权决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Room Record 定义与业务规则 47-51。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:40` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:45 --- File adapter 无法在单一 JSON 文件中提供真正的每房间原子更新 --- 改为目录式房间文件与原子重命名 [Atomic Rename] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 单一 JSON 文件仍要求重写全部房间，无法隔离并发更新、单房间损坏和关闭删除。

### 使用什么方式解决
- `CLOCKTOWER_SNAPSHOT_PATH` 改为目录，每个房间使用 `<roomId>.json` 独立文件。
- 更新时在同目录写入并刷新临时文件，再原子重命名覆盖；关闭房间删除对应文件。
- 启动扫描目录恢复房间，单个损坏文件记录错误并跳过，不阻断其他房间。
- 本轮不实现旧单文件 Snapshot 自动迁移。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 File adapter 原子更新、恢复与损坏隔离规则。
- `work.md` — 新增本次文件持久化决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 52-56。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:45` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:50 --- Room Record 演进缺少版本与恢复失败策略 --- 引入显式模式版本 [Schema Version] 与逐版本迁移 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 持久化结构变化后，缺少版本字段会迫使恢复代码猜测缺失字段，未知新版本也可能被错误解释。

### 使用什么方式解决
- 每条 Room Record 包含从 1 开始的 `schemaVersion`。
- 未知新版本记录错误并跳过，不做推测转换；旧版本通过显式逐版本迁移函数升级。
- 迁移在内存中执行，并在该房间下一次成功命令时写回最新版本。
- 本轮只建立版本框架，不实现旧 Room Record 迁移。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Room Record 版本与迁移规则。
- `work.md` — 新增本次恢复兼容决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 57-61。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:50` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 15:55 --- Create Room 响应丢失会产生重复或孤儿房间 --- 引入创建请求幂等键 [Idempotency Key] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 房间已持久化但创建响应丢失时，客户端无法获得 `roomId` 与恢复凭证，直接重试会创建第二个房间并遗留不可管理状态。

### 使用什么方式解决
- Create Room 必须携带客户端生成且不可猜测的 `requestId`。
- 首次成功时持久化请求参数及房间、创建者和凭证结果；相同请求重试返回原结果。
- 同一 `requestId` 参数不一致时拒绝；关闭房间时删除对应幂等记录。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Create Request 术语及 Create Room 幂等规则。
- `work.md` — 新增本次创建失败恢复决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Create Request 定义与业务规则 62-66。
- 从 `work.md` 删除标题以 `## 2026-07-10 15:55` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:00 --- 普通命令确认丢失会导致重复执行或依赖规则偶然拒绝 --- 引入单调客户端序号 [Monotonic Client Sequence] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 夜间行动、处决、阶段切换等命令可能已提交但确认响应丢失；重试不能依赖具体游戏规则恰好阻止第二次执行。

### 使用什么方式解决
- 每个房间成员身份使用独立递增的 `clientSequence`，服务器持久化最后接受的序号与命令标识。
- 新命令必须使用期望的下一序号；相同序号与相同内容视为去重重试，不重新执行或广播。
- 相同序号内容不同、旧序号或跳号均拒绝，并返回服务器期望的下一序号。
- 重连响应包含下一可用序号。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Client Sequence 术语及成员命令幂等规则。
- `work.md` — 新增本次重复投递决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Client Sequence 定义与业务规则 67-72。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:00` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:05 --- 客户端序号在离开、重连与分离式命令间的作用域不明确 --- 将序号绑定持久身份 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 若客户端序号 [Client Sequence] 绑定连接或当前成员资格，主动离开后重置序号会允许旧命令重放，分离式 Close Room 也缺少幂等顺序。

### 使用什么方式解决
- Client Sequence 绑定持久玩家身份，保留身份 [Retained Identity] 继续保存最后序号。
- 重新加入和分离式 Close Room 使用下一序号；重连与连接接管不改变权威状态，因此不消耗序号。
- 新身份由服务器返回初始下一序号。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Client Sequence 身份作用域规则。
- `work.md` — 新增本次序号生命周期决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 73-77。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:05` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:10 --- 首次 Join Room 响应丢失会重复创建身份或遗失凭证 --- 引入加入请求幂等键 [Join Request Idempotency] --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 新玩家首次加入时尚无恢复凭证和客户端序号；若持久化成功但响应丢失，重试可能创建重复成员或签发不同凭证。

### 使用什么方式解决
- 首次 Join Room 必须携带不可猜测的 `joinRequestId`。
- 成功时持久化请求参数、新身份、恢复凭证和初始下一客户端序号；相同请求重试返回原结果。
- 同一请求 ID 参数冲突时拒绝；身份创建后改用既有身份的重连或保留身份重新加入规则。
- Join Request 记录随房间关闭删除。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Join Request 术语及首次加入幂等规则。
- `work.md` — 新增本次加入恢复决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Join Request 定义与业务规则 78-83。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:10` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:15 --- 保留身份重新加入与普通断线重连的幂等语义混淆 --- 将重新加入定义为带序号成员命令 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 保留身份 [Retained Identity] 重新加入会改变成员关系，而普通断线重连不会；若两者都不消耗客户端序号，响应丢失后的重复重新加入缺少可靠去重。

### 使用什么方式解决
- 保留身份重新加入携带 `playerId + Resume Credential + clientSequence`，并在同一事务式提交中恢复成员资格和推进序号。
- 相同序号与相同重新加入命令的重试返回当前已提交状态，不再次广播。
- 普通断线重连保持传输操作，不消耗客户端序号。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增保留身份重新加入的序号与去重规则。
- `work.md` — 新增本次重新加入语义决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 84-87。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:15` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:20 --- Leave Room 响应丢失会让客户端误判成员状态 --- 提交后关闭连接并显式返回保留身份状态 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 主动离开已提交但响应丢失时，客户端可能继续把自己视为成员；若重连自动恢复成员资格，又会撤销用户刚完成的离开意图。

### 使用什么方式解决
- Leave Room 作为带客户端序号的状态命令，提交后将成员转为保留身份并关闭活动连接。
- 后续认证只返回保留身份状态和下一期望序号，不自动恢复成员；必须显式执行重新加入命令。
- 重试已接受的 Leave Room 序号返回已离开状态，不重复执行。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Leave Room 提交、连接关闭与恢复规则。
- `work.md` — 新增本次主动离开失败恢复决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 88-92。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:20` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:25 --- Kick Player 的撤权、通知与连接关闭顺序未定义 --- 先持久化撤权再尽力通知 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 若先通知或关闭连接再持久化踢出结果，持久化失败可能让目标已断线但仍具有效身份；若错误信息暴露封禁，又会泄漏房间策略。

### 使用什么方式解决
- Kick Player 在房间串行序列中先原子持久化成员删除、凭证撤销、房间封禁与发起者序号。
- 提交成功后尽力发送 `KICKED`，随后无论通知结果都关闭并移除目标连接。
- 旧凭证后续只收到统一凭证错误，广播不包含凭证或封禁详情。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Kick Player 撤权、通知与隐私规则。
- `work.md` — 新增本次踢出时序决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 93-96。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:25` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:30 --- Close Room 的持久化删除、注册表移除与连接通知顺序未定义 --- 采用持久化优先的关闭事务 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 若先移除内存房间或关闭连接再删除持久化记录，删除失败会造成运行状态与恢复状态分裂；关闭后的错误也可能泄漏历史成员信息。

### 使用什么方式解决
- Close Room 在房间串行序列内校验创建者凭证与下一客户端序号。
- 先持久化删除 Room Record 与 Create Request；失败则保持房间、连接和内存状态不变。
- 删除成功后移出会话注册表，使用预先冻结的连接列表尽力发送 `ROOM_CLOSED` 并关闭连接。
- 关闭后的所有请求统一返回房间不存在。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Close Room 事务、通知与隐私规则。
- `work.md` — 新增本次关闭时序决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 97-101。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:30` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:35 --- 短房间 ID 复用会混淆已关闭房间与新房间身份 --- 使用不可猜测且不复用的持久 Room ID --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 关闭后复用有限短码会让旧链接、缓存请求和历史凭证误指向新房间，也需要墓碑才能区分生命周期。

### 使用什么方式解决
- Room ID 使用足够长、不可猜测的持久标识，并且关闭后不主动复用。
- 创建时对 Room Record 执行条件创建；碰撞则生成新 ID 重试。
- 不保存关闭墓碑；未来如需短邀请码，作为独立可轮换别名设计。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Room ID 术语及不复用、条件创建规则。
- `work.md` — 新增本次标识符决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Room ID 定义与业务规则 102-105。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:35` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:40 --- Redis 持久化可能被误解为支持多实例并发处理 --- 限定单活动后端实例并引入 Room Revision --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 每房间串行化仅在单进程内成立；两个后端实例可能同时读取并覆盖同一 Room Record，Redis 本身不会提供命令排序。

### 使用什么方式解决
- 当前部署明确只支持单活动后端实例 [Single Active Backend Instance]。
- Redis 仅作为持久化 adapter，不承担分布式锁、排序或故障切换协调。
- 每次成功替换 Room Record 递增 Room Revision，用于陈旧写入检测和未来扩展。
- 多实例、房间分片、分布式租约与主动故障切换延后设计，并要求文档与启动日志声明限制。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Room Revision 术语及部署限制规则。
- `work.md` — 新增本次部署模型决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除 Room Revision 定义与业务规则 106-110。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:40` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。
## 2026-07-10 16:45 --- 客户端无法检测广播遗漏、乱序或陈旧状态 --- 将 Room Revision 暴露到协议 --- 修改 `packages/backend/CONTEXT.md` 与 `work.md`

### 发现什么问题
- 仅依赖增量广播时，客户端无法可靠判断是否遗漏消息、收到乱序结果或持有旧状态。

### 使用什么方式解决
- 完整房间状态、成功命令确认和提交广播均包含持久化后的 Room Revision。
- 客户端发现修订号跳跃或倒退时请求完整状态，不自行推断遗漏事件。
- 去重命令响应返回当前最新修订号；修订号只表示提交顺序，与游戏阶段、事件数量及客户端序号区分。
- 本轮后端协议支持该字段，前端自动重同步可后续垂直切片交付。

### 修改了哪些文件
- `packages/backend/CONTEXT.md` — 新增 Room Revision 协议可见性规则。
- `work.md` — 新增本次客户端一致性决策记录。

### 撤回方式 [Rollback Strategy]
- 从 `packages/backend/CONTEXT.md` 删除业务规则 111-115。
- 从 `work.md` 删除标题以 `## 2026-07-10 16:45` 开头的本节；或在确认没有其他未提交修改后执行 `git checkout -- packages/backend/CONTEXT.md work.md`。

## 2026-07-10 16:38 --- 房间成员关系、游戏状态与持久化散落在 Hub 和全局 Snapshot 中 --- 引入权威游戏会话 [Authoritative Game Session] 与每房间事务提交 --- 修改 Backend 会话、存储、WebSocket 与启动文件

### 发现什么问题
- 断线会混淆成员资格与活动连接，无法安全保留身份并重连。
- 同房间命令缺少统一串行事务，持久化失败可能造成内存与存储状态撕裂。
- 全局 Snapshot 扩大故障域，隐私投影、客户端序号和房间修订号没有单一权威来源。

### 使用什么方式解决
- 新增 `internal/session` 深层模块 [Deep Module]，独占成员、保留身份、凭证 nonce、封禁、设置、游戏状态、客户端序号、房间修订号和关闭墓碑。
- 使用候选状态 → Store CAS → 原子发布的事务式提交 [Transactional Commit]；同房间串行，不同房间可并行。
- 新增 Memory、File 与 Redis Room Record Store 适配器 [Adapter]，文件模式改为每房间文件，Redis 改为命名空间与 Lua CAS。
- Hub 收窄为协议与活动连接适配器；连接接管使用连接代次并在执行临界区重新验证。
- 修复普通玩家夜晚进度与角色私密信息投影。

### 修改了哪些文件
- `packages/backend/internal/session/**` — 权威会话、注册表、凭证和模块测试。
- `packages/backend/internal/sessionstore/**` — Store seam 与 Memory/File/Redis adapters、契约测试。
- `packages/backend/internal/ws/active_connections.go`、`packages/backend/internal/ws/hub_v2.go`、`packages/backend/internal/ws/session_engine.go` 及相关测试 — 连接代次、协议编排与旧游戏规则实现适配。
- `packages/backend/internal/ws/hub.go`、`packages/backend/internal/ws/message.go`、`packages/backend/internal/ws/game_session.go` — 生产 v2 路径、协议字段与隐私修复。
- `packages/backend/cmd/server/main.go` — 凭证密钥校验、每房间恢复、单实例日志与健康检查。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/session`、`packages/backend/internal/sessionstore` 和新增的 v2 WebSocket 文件。
- 将 `packages/backend/internal/ws/hub.go`、`packages/backend/internal/ws/message.go`、`packages/backend/internal/ws/game_session.go`、`packages/backend/cmd/server/main.go` 恢复到本次修改前版本；旧 Snapshot 备份可用于整体回滚。

## 2026-07-10 16:38 --- 客户端缺少安全重连、序号去重与修订重同步 --- Core 与 Frontend 同步切换协议 v2 --- 修改 Core WebSocket 与 Frontend 身份生命周期

### 发现什么问题
- 客户端只保存短房间 ID，断线后无法证明身份；命令确认丢失可能重复执行或错误推进后续命令。
- 客户端无法发现房间广播跳跃、倒退或同修订重复。

### 使用什么方式解决
- Core 统一添加协议 v2 信封、恢复身份、创建/加入幂等键、单在途序号队列和自动 `RESUME_ROOM`。
- 恢复时比较服务端下一序号，区分未提交重试与已提交确认丢失。
- Frontend 持久化版本化房间身份，支持保留身份重新加入、关闭房间、被踢清理和完整状态重同步。
- 发现修订缺口时不应用不一致增量负载，先请求权威完整投影。

### 修改了哪些文件
- `packages/core/src/websocket/index.ts`、`packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/frontend/src/lib/utils.ts`、`packages/frontend/src/lib/utils.test.ts`、`packages/frontend/src/pages/index/index.tsx`。

### 撤回方式 [Rollback Strategy]
- 将上述 Core 与 Frontend 文件恢复到本次修改前版本；同时清除客户端存储键 `clocktower.roomIdentity.v2`，避免旧客户端误读 v2 身份记录。

## 2026-07-10 16:38 --- 新架构缺少统一交付门禁和部署事实源 --- 增加协议文档、配置示例、CI 与统一验证脚本 --- 修改文档和工程配置

### 发现什么问题
- 三个上下文文档仍描述旧 WebSocket 与全局 Snapshot，仓库没有 CI 工作流或生产环境变量示例。
- Go、竞态测试、Core 与 Frontend 构建测试没有统一入口。

### 使用什么方式解决
- 新增协议 v2 和 Backend 部署文档，更新三个 `CONTEXT.md`。
- 新增 `.env.example`、`pnpm verify` 与 GitHub Actions 门禁。
- 忽略 TypeScript 增量构建产物，避免验证污染工作树。

### 修改了哪些文件
- `docs/protocol/websocket-v2.md`、`packages/backend/README.md`、`.env.example`。
- `packages/backend/CONTEXT.md`、`packages/core/CONTEXT.md`、`packages/frontend/CONTEXT.md`。
- `package.json`、`.github/workflows/verify.yml`、`.gitignore`、`work.md`。

### 撤回方式 [Rollback Strategy]
- 删除新增文档、配置示例和 CI 文件，并将 `package.json`、`.gitignore` 与三个 `CONTEXT.md` 恢复到本次修改前版本。

## 2026-07-10 17:09 --- 已提交会话投影可能乱序、阻塞或混入下一修订元数据 --- 引入实时投影提交观察器与有序出站队列 --- 修改 Session、Hub、Core 与 Frontend

### 发现什么问题
- Hub 在会话锁释放后同步发送确认和广播，后续命令可能先发送更高修订；慢连接会阻塞其他接收者。
- Join 通过重新查询生成广播，可能跳过 Join 修订；Hub 重新读取元数据可能把投影 N 与元数据 N+1 混合。
- Core 与 Frontend 分散处理修订门禁，过期完整投影仍可能回滚 UI，完整同步也可能残留旧提名和夜间状态。
- 间谍和游戏结束角色公开复用了说书人全视野，可能连带泄漏中毒和夜晚管理状态。

### 使用什么方式解决
- 为 `AuthoritativeGameSession.Execute` 与 `SessionRegistry.Join` 增加提交观察器 [Commit Observer]，在房间串行锁内生成并入队同一已提交视图的确认和接收者投影。
- 将提交元数据随结果返回，Hub 不再从后续视图重建投影元数据，也不再凭凭证重新查询 Join 广播。
- 新增每连接独立、有界的有序出站队列 [Ordered Outbound Queue]；发送异步执行，慢消费者或发送失败只关闭活动连接。
- Core 增加投影门禁 [Projection Gate]，统一忽略重复、过期和不连续投影，并限制同一重同步周期只请求一次完整状态。
- Frontend 在清除或应用完整投影时显式清理提名、死亡、阶段、夜间输入与游戏结束状态。
- 将角色可见、中毒可见和夜间管理可见拆分为投影能力 [Projection Capabilities]。

### 修改了哪些文件
- `packages/backend/internal/session/types.go`、`packages/backend/internal/session/session.go`、`packages/backend/internal/session/registry.go`。
- `packages/backend/internal/ws/outbound_dispatcher.go`、`packages/backend/internal/ws/outbound_dispatcher_test.go`、`packages/backend/internal/ws/hub.go`、`packages/backend/internal/ws/hub_v2.go`、`packages/backend/internal/ws/hub_v2_test.go`、`packages/backend/internal/ws/session_persistence_test.go`。
- `packages/backend/internal/ws/game_session.go`、`packages/backend/internal/ws/game_session_test.go`。
- `packages/core/src/websocket/index.ts`、`packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/frontend/src/pages/index/index.tsx`。
- `packages/backend/CONTEXT.md`、`packages/core/CONTEXT.md`、`packages/frontend/CONTEXT.md`、`docs/protocol/websocket-v2.md`、`work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/ws/outbound_dispatcher.go` 与对应测试。
- 将上述 Session、Hub、游戏投影、Core、Frontend 和文档文件恢复到本节修改前版本；恢复后投影会回到同步广播和客户端分散修订判断模式。

## 2026-07-13 10:03 --- 架构职责分散且迁移残留扩大测试面 --- 使用领域上下文与删除测试生成架构审查报告 --- 修改系统临时报告与 work.md

### 发现什么问题
- 后端旧 `RoomManager`、旧 `GameSession`、旧全局快照与协议 v2 的权威游戏会话 [Authoritative Game Session] 同时存在。
- 前端首页同时承担连接、身份、实时游戏会话投影 [Real-time Game Session Projection]、命令和渲染职责。
- WebSocket 协议 v2 在 Go、TypeScript、ProtoBuf 与 Markdown 中存在多份手写接口 [Interface]。

### 使用什么方式解决
- 读取多上下文领域文档并核查各包架构决策记录 [ADR]；当前没有已落盘 ADR。
- 对候选模块 [Modules] 执行删除测试 [Deletion Test]，按局部性 [Locality]、杠杆效应 [Leverage] 与测试面排序。
- 生成包含三个候选及前后对比图的 HTML 架构审查报告，未修改实现代码。

### 修改了哪些文件
- `C:\Users\Qilia\AppData\Local\Temp\architecture-review-20260713-100024.html`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除系统临时目录中的 `architecture-review-20260713-100024.html`。
- 删除 `work.md` 中标题为 `2026-07-13 10:03` 的本节记录。

## 2026-07-13 10:20 --- 后端同时维护协议 v0 与 v2 两套房间状态栈 --- 收拢为 v2 权威游戏会话并迁移关键测试 --- 修改 Backend WebSocket 模块与 work.md

### 发现什么问题
- `Hub` 同时持有旧 `RoomManager`、旧 `GameSession` 映射、连接映射、全局快照和新的权威游戏会话 [Authoritative Game Session] Registry。
- 生产入口与协议文档只允许 v2，但大量测试仍通过协议 v0 驱动另一套实现。
- 仍被 v2 Engine 使用的 `GameSession` 快照克隆逻辑混在旧 Hub 全局持久化文件中。

### 使用什么方式解决
- 将 `Hub` 深化为协议 v2 传输适配器 [Adapter]，只保留协议校验、Registry、活动连接和有序出站投递。
- 删除协议 v0 处理器、`RoomManager`、旧全局快照与对应测试。
- 将 `GameSession` 快照实现迁到独立模块，保留游戏规则持久化能力。
- 新增 v2 加入投影、身份伪造、连接接管和真实 WebSocket 协议版本测试。

### 修改了哪些文件
- 重写 `packages/backend/internal/ws/hub.go`，修改 `conn.go`、`hub_v2_test.go`。
- 新增 `game_session_snapshot.go`、`game_session_test_helpers_test.go`、`hub_transport_test.go`。
- 删除 `broadcaster.go`、`room_manager.go`、`persistence.go` 及协议 v0 专属测试文件。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 从版本控制恢复本节删除的旧 Hub、RoomManager、全局快照与测试文件。
- 删除三个新增文件，并将 `hub.go`、`conn.go`、`hub_v2_test.go` 恢复到本节修改前版本。
- 删除 `work.md` 中标题为 `2026-07-13 10:20` 的本节记录。

## 2026-07-13 10:45 --- 跨语言协议与前端投影仍由多处手写状态共同维护 --- 生成协议契约并深化房间体验投影模块 --- 修改 ProtoBuf、Core、Frontend、Backend 与文档

### 发现什么问题
- WebSocket v2 信封、枚举和错误码在 ProtoBuf、Go、TypeScript 与文档中重复手写，变更时可能发生契约漂移 [Contract Drift]。
- 前端首页同时维护房间投影、提名、死亡、夜间行动、胜负和身份等并行状态，并解析增量事件重建服务端状态。
- 夜间行动已由后端管理，但完整投影缺少该字段，前端只能依赖瞬时事件；同时需要明确验证玩家投影不会泄漏说书人夜间管理信息。

### 使用什么方式解决
- 扩展 `proto/game.proto`，以 ProtoBuf 作为 WebSocket JSON 契约的规范来源 [Canonical Source]；新增生成器同时输出 Go 与 TypeScript 契约，并以内容哈希和 `proto:check` 防止生成物漂移。
- 新增纯函数房间体验投影模块 [Room Experience Projection Module]，用单一状态原子替换完整服务端投影并派生页面视图；删除页面内重复修订门禁、增量事件重建和并行领域状态。
- 将夜间行动加入接收者特定投影，仅在夜晚向说书人暴露，并新增说书人与间谍投影的隐私边界测试。
- 更新协议说明、三层领域上下文与统一验证入口，使架构约束可发现并可自动检查。

### 修改了哪些文件
- `proto/game.proto`、`proto/README.md`、`scripts/generate-types.mjs`、`scripts/generate-protocol-contracts.mjs`、`package.json`。
- `packages/backend/internal/ws/protocol_generated.go`、`message.go`、`hub.go`、`hub_v2.go`、`game_session.go`、`game_session_test.go`。
- `packages/core/src/websocket/protocol.generated.ts`、`index.ts`、`__tests__/websocket-client.test.ts`、`packages/core/src/types/generated/index.ts`、`packages/core/tsconfig.tsbuildinfo`。
- `packages/frontend/src/lib/room-experience.ts`、`room-experience.test.ts`、`utils.ts`、`pages/index/index.tsx`、`packages/frontend/tsconfig.tsbuildinfo`。
- `docs/protocol/websocket-v2.md`、`docs/prd/mvp-human-storyteller.md`、三个包的 `CONTEXT.md`、`work.md`。

### 撤回方式 [Rollback Strategy]
- 删除协议契约生成器与两个新生成文件，将 `game.proto`、`generate-types.mjs`、`package.json` 和 WebSocket 消费代码恢复到本节修改前版本。
- 删除房间体验模块及测试，将首页和工具函数恢复为原有独立状态与事件解析实现。
- 将 `game_session.go`、对应隐私测试、协议文档、三个 `CONTEXT.md` 与 `work.md` 恢复到本节修改前版本。

## 2026-07-13 11:24 --- WebSocket 包同时承载传输与完整游戏规则 --- 深化 Gameplay Session 并清理测试专用生产代码 --- 修改 Backend 模块、测试、验证脚本与文档

### 发现什么问题
- `internal/ws` 的 32 个文件中有 21 个测试文件，数量本身合理；但 2,180 行的 `game_session.go` 以及 15 个角色规则测试都位于传输包，包名与领域所有权不一致。
- `GameSession` 直接返回协议 `RoomState` 并接收 `roomId`，导致游戏规则模块知道传输元数据。
- `FakeConnection` 只由测试使用却编译进生产包；`session_memory_store.go` 没有调用者。
- 慢消费者队列测试没有等待发送协程真正阻塞，容量断言受调度时序影响。

### 使用什么方式解决
- 新增 Gameplay Session 模块 [Gameplay Session Module]，把游戏状态、快照、角色规则和领域测试迁入 `internal/gameplay`。
- 将接收者视图收敛为不含房间元数据的游戏投影 [Gameplay Projection]；`ws/session_engine.go` 作为适配器 [Adapter] 转换为 ProtoBuf 生成的 `RoomState`。
- 保持 `Apply(Command)` 单一深接口 [Deep Interface]，只按信息计算、白天、夜晚、胜负和投影拆分实现文件，不引入角色插件接缝。
- 将测试连接迁入 `fake_connection_test.go` 并改为非导出实现，删除未使用的内存存储转发文件。
- 为阻塞发送测试增加确定性同步信号，并把 `internal/gameplay` 纳入竞态检测 [Race Detection]。

### 修改了哪些文件
- 新增 `packages/backend/internal/gameplay/`，包含 `game_session.go`、`information.go`、`day.go`、`night.go`、`endgame.go`、`projection.go`、`game_session_snapshot.go` 及角色规则测试。
- 修改 `packages/backend/internal/ws/session_engine.go`、`hub_v2.go`、`conn.go`、`fake_connection_test.go`、`outbound_dispatcher_test.go`、Hub 与连接测试。
- 删除 `packages/backend/internal/ws/session_memory_store.go`，并从 `internal/ws` 迁出 Gameplay Session 与对应测试文件。
- 修改 `package.json`、`packages/backend/CONTEXT.md`、`work.md`。

### 撤回方式 [Rollback Strategy]
- 将 `internal/gameplay` 中迁移的 GameSession、快照和角色测试移回 `internal/ws`，恢复包声明与原 `RoomState` 投影方法。
- 将拆分的 Gameplay Session 实现合并回 `game_session.go`，恢复 `session_engine.go` 对原快照和命令类型的引用。
- 把测试连接实现恢复到 `conn.go`，恢复 `session_memory_store.go` 与原队列测试时序。
- 将 `package.json`、Backend `CONTEXT.md` 和 `work.md` 恢复到本节修改前版本。

## 2026-07-13 11:55 --- 协议漂移检查覆盖不完整且字段可选性重复维护 --- 将 JSON 必填语义收回 ProtoBuf 并停止跟踪构建缓存 --- 修改协议生成流程、文档与 Git 跟踪项

### 发现什么问题
- `proto:check` 只检查 WebSocket 的 Go 与 TypeScript 生成物，没有检查被版本控制跟踪的领域类型 `packages/core/src/types/generated/index.ts`。
- WebSocket 生成器在脚本内维护第二份 JSON 必填字段清单，`game.proto` 的字段存在性 [Field Presence] 变更无法自动同步。
- 两个已被 `.gitignore` 忽略的 `tsconfig.tsbuildinfo` 增量构建缓存 [Incremental Build Cache] 仍被 Git 跟踪。

### 使用什么方式解决
- 为 `generate-types.mjs` 增加只读 `--check` 模式，并让统一的 `proto:check` 入口覆盖全部被跟踪的生成物。
- 在 `game.proto` 定义 `json_required` 自定义字段选项 [Custom Field Option]，由 Go 与 TypeScript 契约生成器直接读取，删除脚本内的必填字段清单。
- 将未被消费且已由 Git 忽略的 protobufjs 中间声明文件排除出 ESLint，避免标准描述类型产生无意义告警。
- 从 Git 索引移除两个 `tsconfig.tsbuildinfo`，本地文件继续由 `.gitignore` 管理。

### 修改了哪些文件
- `proto/game.proto`、`proto/README.md`。
- `scripts/generate-types.mjs`、`scripts/generate-protocol-contracts.mjs`、`package.json`。
- `.eslintrc.json`。
- 重新生成 `packages/core/src/websocket/protocol.generated.ts` 与 `packages/backend/internal/ws/protocol_generated.go`。
- 从版本控制移除 `packages/core/tsconfig.tsbuildinfo`、`packages/frontend/tsconfig.tsbuildinfo`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `game.proto` 中的 `json_required` 扩展与字段标注，并恢复生成器内原有必填字段清单。
- 将 `package.json` 的 `proto:check` 恢复为仅调用 `generate-protocol-contracts.mjs --check`，撤回 `generate-types.mjs` 的检查模式。
- 从 `.eslintrc.json` 删除 protobufjs 中间声明文件的忽略项。
- 使用 `git add -f packages/core/tsconfig.tsbuildinfo packages/frontend/tsconfig.tsbuildinfo` 恢复两个缓存文件的跟踪。
- 重新运行 `pnpm proto:generate`，并删除 `work.md` 中标题为 `2026-07-13 11:55` 的本节记录。
