# Work Log

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
