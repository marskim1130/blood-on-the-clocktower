# Work Log

## 2026-08-04 12:00:50 +08:00 --- 仓库缺少开源项目必需的三要素：无 LICENSE、无根级 README.md、Go 模块名为占位符 `github.com/your-org`（go.mod、41 处 import、协议生成器模板均引用该占位符，导致 `go get` 无法工作） --- 添加 MIT LICENSE（版权人 金琦亮）；编写根级 README.md（功能特性、技术栈、项目结构、快速开始、验证、文档导航、许可）；将 `go.mod` module 行、35 个 `.go` 文件的 import、`scripts/generate-protocol-contracts.mjs` 生成模板中的模块路径统一改为 `github.com/marskim1130/blood-on-the-clocktower`，重新运行 `pnpm proto:generate` 生成契约，`pnpm proto:check` 无漂移，`go build` / `go vet` / `go test ./...` 全部通过 --- 修改了 `LICENSE`（新增）、`README.md`（新增）、`packages/backend/go.mod`、`scripts/generate-protocol-contracts.mjs`、`packages/backend` 下 35 个 `.go` 文件、重新生成的 `packages/backend/internal/ws/protocol_generated.go` 与 `packages/core/src/websocket/protocol.generated.ts`、`work.md`

### 撤回方式 [Rollback Strategy]
删除 `LICENSE` 与 `README.md`；将 `go.mod` 的 `module` 行、全部 `.go` 文件 import、生成器脚本中的模块路径改回 `github.com/your-org/blood-on-the-clocktower`，重新运行 `pnpm proto:generate` 恢复生成文件，并删除本节 `work.md` 记录。

## 2026-07-15 09:10:08 +08:00 --- 缺少 Go 后端学习材料，前端开发者难以把 WebSocket 操作映射到服务端执行链 --- 基于真实 Backend、协议和前端传输代码制作一节请求链课程与速查表，并补充官方 Go 资源 --- 修改 `learning/lessons/0002-golang-backend-request-flow.html`、`learning/reference/golang-backend-cheatsheet.html`、`learning/RESOURCES.md`、`work.md`

### 撤回方式 [Rollback Strategy]
删除新增课程与速查表，移除 `learning/RESOURCES.md` 中“Go 与后端知识”小节，并删除本节 `work.md` 记录；不改动 `learning/MISSION.md` 或学习记录。

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

## 2026-07-13 16:38 --- 开放 Issues 长期未随实现与架构演进更新 --- 按 triage 状态机归并队列并补齐发布垂直切片 --- 修改 GitHub Issues、`.out-of-scope/public-room-directory.md` 与 `work.md`

### 发现什么问题
- 32 个开放 Issue 全部只带 `ready-for-agent`，缺少必需的 `bug` 或 `enhancement` 类别，也没有代理简报 [Agent Brief]。
- `#2-#22` 多数已被协议 v2、权威游戏会话 [Authoritative Game Session]、Gameplay Session 和现有前端实现完成或取代，仍保持开放会制造错误待办。
- `#31-#40` 与旧前端切片重叠，部分正文依赖不存在的命令、公开房间列表或与当前权威投影架构冲突的增量状态方案。
- 当前队列没有覆盖生产部署、完整多人端到端回归 [End-to-end Regression] 和后端发布加固三个真实上线缺口。

### 使用什么方式解决
- 以 `enhancement` 类别关闭 12 个已完成 Issue：`#2 #3 #4 #6 #7 #9 #10 #12 #13 #16 #20 #22`。
- 以 `enhancement + duplicate` 关闭 12 个已取代 Issue：`#5 #8 #11 #14 #15 #17 #18 #19 #21 #31 #38 #40`；每项评论均指出承接 Issue 或现行架构决策。
- 将公开房间列表 `#33` 标记为 `enhancement + wontfix` 并关闭；把邀请制房间决策写入 `.out-of-scope/public-room-directory.md`。
- 保留并重新分诊 `#1 #32 #34 #35 #36 #37 #39`：其中五个代理切片补充行为型 Agent Brief，`#1` 转为 `ready-for-human` 发布追踪项，`#36` 转为 `needs-triage` 等待投票语义决策。
- 新建 `#41` 后端安全、可观测性与负载基线，`#42` 协议 v2 一名说书人与五名玩家完整对局回归，`#43` 生产部署、WSS 与微信发布配置。
- 所有新 Issue 和分诊评论均以 triage AI 声明开头；最终开放队列为 10 项，每项恰好包含一个类别标签和一个状态标签。

### 修改了哪些文件或外部状态
- GitHub Issues：更新 `#1-#22`、`#31-#40` 的状态、标签或评论；新增 `#41 #42 #43`。
- `.out-of-scope/public-room-directory.md` — 记录不提供公开房间目录、改用邀请分享或可轮换短邀请码的长期范围决策。
- `work.md` — 新增本次 Issue 队列整理记录。

### 撤回方式 [Rollback Strategy]
- 重新打开 `#2 #3 #4 #5 #6 #7 #8 #9 #10 #11 #12 #13 #14 #15 #16 #17 #18 #19 #20 #21 #22 #31 #33 #38 #40`，移除本次新增的 `enhancement`、`duplicate`、`wontfix`，并恢复原先唯一的 `ready-for-agent` 标签。
- 将 `#1 #32 #34 #35 #36 #37 #39` 的标签恢复为仅 `ready-for-agent`；如需完全撤销审计历史，再按评论 URL 删除本次 AI triage 评论。
- GitHub 不支持删除普通 Issue；将新增的 `#41 #42 #43` 关闭为 `not planned` 可撤销其待办影响。
- 删除 `.out-of-scope/public-room-directory.md`，并删除 `work.md` 中标题以 `2026-07-13 16:38` 开头的本节。

## 2026-07-13 17:48 --- 角色配置缺少可测试的合法随机与手动交换能力 --- 新增纯函数角色设置模型与确定性测试 --- 修改 Core 剧本模块及测试

### 发现什么问题
- 前端只能生成固定示例分配，无法在 5-15 人、男爵修正、酒鬼展示身份和占卜师干扰项规则下安全随机。
- 手动调整角色时缺少保持一一映射 [Bijection] 的共享操作，页面容易产生重复或丢失角色。

### 使用什么方式解决
- 新增 `ScriptSetup` 与 `validateScriptSetup`，集中校验玩家集合、角色计数、酒鬼和占卜师特殊设置。
- 新增可注入随机源的 Fisher-Yates 洗牌 [Shuffle]，先抽取恶魔与爪牙，再按男爵结果调整镇民与外来者。
- 新增不可变角色交换函数，并以 11 个聚焦测试覆盖 5/15 人和特殊规则。

### 修改了哪些文件
- `packages/core/src/scripts/index.ts`。
- `packages/core/src/scripts/__tests__/scripts.test.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 从剧本模块删除 `ScriptSetup`、`validateScriptSetup`、`randomizeScriptAssignments`、`swapScriptAssignments` 及对应辅助函数和导出。
- 删除测试文件中本次新增的角色设置测试，并删除 `work.md` 中标题为 `2026-07-13 17:48` 的本节记录。

## 2026-07-13 17:48 --- 拆页会销毁页面级 WebSocket 且终端原因会被二次清空 --- 建立应用级会话仓库和权威阶段路由 --- 修改前端会话基础与构建配置

### 发现什么问题
- 原首页在卸载清理函数中断开网络套接字 [WebSocket]，拆分 Setup、Game-Play 与 Game-Over 后会在每次跳转丢失连接。
- 身份持久化、协议错误、命令待处理和终端状态散落在页面中，`KICKED`、`ROOM_CLOSED` 等消息可能被后续清理覆盖。

### 使用什么方式解决
- 新增单例 Zustand 会话仓库 [Session Store]，在应用级初始化一次客户端并统一处理自动恢复、身份持久化、命令错误和终端消息。
- 新增阶段路由纯函数与协议错误翻译，保留权威投影 [Authoritative Projection] 作为唯一已提交游戏状态。
- 构建配置注入服务端点和开发标记；生产端点缺失时进入明确错误状态，开发环境保留本地回退。
- 新增 7 个测试覆盖单次初始化、恢复、跨房间拦截、身份清理和阶段路由。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.ts`、`protocol-feedback.ts`、`session-routing.ts` 及对应测试。
- `packages/frontend/src/app.tsx`、`src/env.d.ts`、`config/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除新增的会话仓库、协议反馈、阶段路由、环境声明及对应测试文件。
- 将 `app.tsx` 恢复为仅返回页面子节点，将 `config/index.ts` 的 `defineConstants` 恢复为空对象。
- 删除 `work.md` 中第二个标题为 `2026-07-13 17:48` 的本节记录；不要撤回更早的 Issue 整理记录。

## 2026-07-14 11:11:29 +08:00 --- 房间契约尚未覆盖容量、说书人、冻结名单、阶段编号和终局离开语义 --- 扩展 ProtoBuf 契约并在权威会话 [Authoritative Session] 中落实校验、投影与恢复规则 --- 修改协议、Backend、Core 与契约测试

### 发现什么问题
- 协议缺少 `ROOM_FULL`、`INVALID_COMMAND`、说书人姓名、夜晚编号和占卜师干扰项，客户端无法稳定区分容量错误与非法命令。
- 角色发放失败也可能冻结成员，说书人权限、5-15 名实际玩家上限和重复发放没有形成服务端闭环。
- 首夜/首日编号、夜间死亡归属、中毒失效和终局普通成员离开规则与产品约定不一致。
- 恢复消息缺少下一客户端序号 [Client Sequence] 与名单冻结状态，创建/加入中的断线重放可能丢命令或重复身份。

### 使用什么方式解决
- 将新增错误码和房间字段写入 ProtoBuf 规范来源 [Canonical Source]，重新生成 Go/TypeScript 协议契约；占卜师干扰项仅投影给说书人。
- 在会话注册表和游戏会话中统一校验剧本、容量、房主权限、说书人选择和一次性角色分配；仅在成功分配后冻结成员。
- 统一首夜为 `dayNumber=0/nightNumber=1`、首日为第 1 天，并让夜间死亡和中毒按即将到来的白天结算。
- 普通成员仅在终局撤销凭证并保留最终快照，房主执行关闭；恢复投影携带冻结状态与下一序号，客户端可安全重放待处理命令。
- 将无目标夜间行动规范化为 `targetIds: []`，避免 JSON `null` 造成前端运行时崩溃。

### 修改了哪些文件
- `proto/game.proto`、`scripts/generate-protocol-contracts.mjs`。
- `packages/backend/internal/gameplay/day.go`、`game_session.go`、`game_session_snapshot.go`、`night.go`、`projection.go`、`poison_test.go`。
- `packages/backend/internal/session/registry.go`、`session.go`、`session_test.go`、`types.go`。
- `packages/backend/internal/ws/hub_v2.go`、`hub_v2_test.go`、`protocol_generated.go`、`session_engine.go`。
- `packages/core/src/websocket/index.ts`、`protocol.generated.ts`、`__tests__/websocket-client.test.ts`。
- 新增 `packages/backend/internal/gameplay/day_number_test.go`、`projection_contract_test.go`、`packages/backend/internal/ws/protocol_v2_contract_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 将上述 ProtoBuf、生成器、Backend、Core 协议和测试文件恢复到本节修改前版本，并删除三个新增契约测试文件。
- 重新运行 `pnpm proto:generate`，确保生成物与恢复后的 `game.proto` 一致；删除 `work.md` 中本节记录。

## 2026-07-14 11:11:29 +08:00 --- 单页调试界面无法在路由切换中保留连接，也未形成设置、游玩和终局闭环 --- 建立应用级会话仓库 [Session Store]、权威阶段路由与四页面本地 MVP --- 修改 Frontend 页面、会话模块、展示适配和测试

### 发现什么问题
- 页面卸载会断开 WebSocket，邀请链接携带过多内部状态，恢复、被踢、关闭和凭证失效由多个页面重复处理。
- 角色设置缺少不可撤销确认、特殊角色约束和隐私视图；日间提名仍依赖内部 ID；夜间存在本地唤醒顺序回退。
- 终局页缺少完整角色、酒鬼展示身份、死亡时间线、胜负原因映射和刷新恢复降级。
- 后端英文角色资料、行动提示和终局描述直接出现在中文界面，协议未知错误也可能导致渲染失败。

### 使用什么方式解决
- 使用单例 Zustand 仓库唯一持有连接、身份、修订、权威投影、待处理命令和错误，应用启动时初始化，页面路由只消费投影。
- 保留大厅 `/pages/index/index`，邀请仅预填 `roomId`；新增 Game-Setup、Game-Play、Game-Over 页面并按权威阶段重定向。
- 设置页接入合法随机、逐玩家交换、酒鬼展示身份和占卜师干扰项；游玩页仅使用服务端唤醒步骤，并按姓名与座位提名。
- 终局页展示完整身份、胜方、六类原因、死亡时间线和离开/关闭流程；刷新先恢复最终投影。
- 增加角色、能力、剧本、夜间结果与协议错误本地化映射，并以显式文本颜色避免 H5 与 RN 的按钮文字差异。

### 修改了哪些文件
- `packages/frontend/src/app.tsx`、`app.config.ts`、`app.css`、`env.d.ts`、`config/index.ts`。
- 新增 `packages/frontend/src/components/session-shell.tsx`、`loading-state.tsx`。
- 新增 `packages/frontend/src/lib/room-session-store.ts`、`session-routing.ts`、`use-session-route.ts`、`protocol-feedback.ts`、`character-display.ts` 及对应测试。
- 修改 `packages/frontend/src/lib/room-experience.ts`、`room-experience.test.ts`。
- 重写 `packages/frontend/src/pages/index/index.tsx`、`index.css`；新增 `pages/game-setup/`、`pages/game-play/`、`pages/game-over/` 页面与样式。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除本节新增的组件、会话/路由/展示模块及三个游戏阶段页面目录。
- 将应用入口、配置、房间体验模块和大厅页面恢复到本节修改前版本；删除 `work.md` 中本节记录。

## 2026-07-14 11:11:29 +08:00 --- 缺少真实六连接覆盖完整对局与投票断线恢复的回归保护 --- 新增协议 v2 六客户端端到端测试 [End-to-End Test] --- 修改 Backend WebSocket 测试

### 发现什么问题
- 单元测试没有证明一名说书人和五名玩家能通过真实 WebSocket 完成首夜、提名、三票处决恶魔和善方胜利。
- 投票中断线恢复时，序号、修订、投票快照和后续命令连续性缺少精确断言。
- 玩家角色隐私、说书人夜间信息、首日编号和终局公开身份没有处于同一完整回归链路中。

### 使用什么方式解决
- 使用真实测试服务器建立六条连接，固定 Slayer、Soldier、Mayor、Poisoner、Imp 身份并完成完整关键路径 [Critical Path]。
- 在投票中主动断开并恢复一名玩家，断言凭证、客户端序号、服务器修订和投票状态连续。
- 对玩家私密投影、说书人全视图、首夜/首日编号、终局公开身份和胜利原因进行接收者级精确断言。

### 修改了哪些文件
- 新增 `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/ws/protocol_v2_game_flow_test.go`，并删除 `work.md` 中本节记录；生产代码不需要额外回滚。

## 2026-07-14 11:11:29 +08:00 --- H5、微信小程序与 React Native 缺少统一可构建配置，原界面在移动端存在溢出和调试信息泄露风险 --- 补齐三端构建入口并实施响应式视觉回归 [Responsive Visual Regression] --- 修改构建配置、依赖与跨端样式

### 发现什么问题
- React Native 缺少入口、Metro 配置和必要依赖，Taro RN 旧启动器在 Node 24/Windows 下出现 `spawn EINVAL`。
- 原 CSS 使用 RN 不支持的布局能力；`!important` 会被 RN 转换器拼入颜色值，导致无效原生样式。
- H5 按设计宽度转换字体后在宽桌面被放大，输入框与按钮文字可能溢出；生产界面还可能暴露端点和内部调试操作。

### 使用什么方式解决
- 增加 RN 入口、Metro 配置和依赖，使用 React Native 官方 bundle 命令生成 Android Bundle；保留微信和 H5 独立构建脚本。
- 将页面主体改为 Flexbox，固定输入/按钮/计数器尺寸；H5 将像素转换目标设为 `px`，保证 390×844 与 1280×720 字号一致。
- 用显式 `Text` 类表达深色、浅色和选中按钮文字，移除会被 RN 错误解析的 `!important` 与无效 `box-sizing`。
- 在两种视口完成大厅、设置、首夜、首日、终局、刷新恢复和房主关闭检查；生产构建隐藏服务器地址、内部 ID 与强制重置入口。

### 修改了哪些文件
- `packages/frontend/package.json`、`packages/frontend/index.js`、`packages/frontend/metro.config.js`、`pnpm-lock.yaml`。
- `packages/core/package.json`。
- `packages/frontend/config/index.ts`、`src/app.css`、各页面 CSS 与涉及显式按钮文本的 TSX 文件。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/frontend/index.js` 与 `metro.config.js`，恢复 Frontend/Core 的 `package.json` 和 `pnpm-lock.yaml`。
- 将 H5 配置、全局及页面样式和显式按钮文本恢复到本节修改前版本；删除 `work.md` 中本节记录。

## 2026-07-14 15:03:55 +08:00 --- MVP 交付后需要重新核对模块深度与架构摩擦 --- 使用领域上下文、删除测试 [Deletion Test] 和并行只读探索生成架构审查报告 --- 修改系统临时报告与 `work.md`

### 发现什么问题
- Core 同时暴露失活客户端状态机和孤立旧 Proto 类型，两者没有运行时消费者，却扩大接口 [Interface] 并制造第二权威来源。
- `Registry.JoinObserved` 直接操作权威游戏会话的锁、提交视图、持久化和投影投递，Join 事务跨模块 [Module] 接缝 [Seam] 泄漏。
- Room Experience 能原子替换权威投影，但阶段、权限、投票和夜间合法动作仍由三个页面各自解释，局部性 [Locality] 不足。
- 根目录与三个包均不存在架构决策记录 [ADR]，本次候选没有已记录决策冲突。

### 使用什么方式解决
- 读取 `CONTEXT-MAP.md`、三个包的 `CONTEXT.md`、技能语言规范和 HTML 报告模板。
- 对 Backend、Core、Frontend 并行执行只读探索，并以删除测试排除已经具备深度 [Depth] 的 Session Store、GameWebSocketClient 和 RoomRecordStore。
- 将三个最高确定性候选、前后对照图、推荐强度与首选顺序写入系统临时 HTML；未提出具体新接口，也未修改实现代码。

### 修改了哪些文件
- `C:\Users\Qilia\AppData\Local\Temp\architecture-review-20260714-145900.html`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `C:\Users\Qilia\AppData\Local\Temp\architecture-review-20260714-145900.html`。
- 删除 `work.md` 中标题以 `2026-07-14 15:03:55` 开头的本节记录。

## 2026-07-14 16:12:47 +08:00 --- Core 暴露无运行时消费者的客户端状态机与旧 Proto 类型，过期工作流仍会引导代理重新使用它们 --- 删除失活模块并让现有协议生成器成为唯一契约入口 --- 修改 Core、生成流程、领域文档与工作流

### 发现什么问题
- `packages/core/src/state-machine/` 没有运行时调用者，却维护与后端权威投影平行的 `GameState`/`GameEvent`。
- `packages/core/src/types/generated/` 只有自测使用；真实客户端使用 `websocket/protocol.generated.ts`，形成两套“生成类型”接口 [Interface]。
- `scripts/generate-types.mjs` 只为旧类型目录调用 `protobufjs-cli`，然后转发到现有协议生成器，属于浅模块 [Shallow Module]。
- 6 个已完成的 `workflows/*.md` 仍要求修改已失活状态机，会破坏后续代理的局部性 [Locality]。
- Core 上下文、Frontend 上下文、PRD 和代理说明仍宣称客户端状态机是权威状态来源。

### 使用什么方式解决
- 删除客户端状态机、旧生成类型及自测，移除根导出和 `./state-machine` 子路径；保留 Vote、Death、Win、Night 与 Scripts 规则模块。
- 将共享类型收窄为仍被规则模块消费的 `PlayerId`、`Character`、`Team`、`DeathCause` 与构造函数。
- 让 `proto:generate`/`proto:check` 直接调用 `generate-protocol-contracts.mjs`，删除转发脚本、Core 的 `zustand` 与根 `protobufjs-cli` 依赖。
- 删除 6 个过期工作流，并更新 Context Map、Core/Frontend 上下文、ProtoBuf 文档、PRD、AGENTS/CLAUDE 描述和旧代码注释。
- 清理旧 Core 构建产物并重新生成依赖锁文件 [Lockfile]；未增加弃用层 [Deprecation Layer] 或新适配器 [Adapter]。

### 修改了哪些文件
- 删除 `packages/core/src/state-machine/index.ts`、`packages/core/src/types/generated/index.ts` 及其测试、`scripts/generate-types.mjs`。
- 删除 `workflows/death-system.md`、`frontend-integration.md`、`mvp-development.md`、`night-phase.md`、`vote-engine.md`、`win-conditions.md`。
- 修改根与 Core `package.json`、`pnpm-lock.yaml`、`.gitignore`、`.eslintrc.json`。
- 修改 `packages/core/src/index.ts`、`types/index.ts`、`vote-engine/index.ts`。
- 修改 `AGENTS.md`、`CLAUDE.md`、`CONTEXT-MAP.md`、Core/Frontend `CONTEXT.md`、`proto/README.md`、`docs/prd/mvp-human-storyteller.md`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 从版本控制恢复本节删除的状态机、旧生成类型、生成脚本和 6 个工作流文件。
- 将本节列出的清单、Core 导出、共享类型、依赖配置和领域文档恢复到修改前版本。
- 运行 `pnpm install --no-frozen-lockfile` 恢复锁文件与依赖链接，并删除 `work.md` 中标题以 `2026-07-14 16:12:47` 开头的本节记录。

## 2026-07-14 13:53:56 +08:00 --- Codex 尚未配置 Ponytail 插件市场且未安装该插件 --- 按官方 Codex 插件流程添加市场并安装后校验 --- 修改用户级 Codex 插件状态与安装审计记录

### 发现什么问题
- `codex plugin marketplace list` 中不存在 `ponytail` 市场，`codex plugin list` 中不存在 `ponytail@ponytail`。
- Ponytail 的生命周期钩子 [Lifecycle Hooks] 依赖 Node.js；本机 `node v24.18.0` 已满足要求。

### 使用什么方式解决
- 按官方 README 执行 `codex plugin marketplace add DietrichGebert/ponytail`。
- 执行 `codex plugin add ponytail@ponytail`，并通过插件清单、版本、安装路径和钩子配置进行安装后校验。

### 修改了哪些文件
- 新增用户级市场快照 `C:\Users\Qilia\.codex\.tmp\marketplaces\ponytail`。
- 新增用户级插件缓存 `C:\Users\Qilia\.codex\plugins\cache\ponytail\ponytail\4.8.4`，并更新 Codex 用户级插件状态。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 若插件钩子已经运行，先执行 `node C:\Users\Qilia\.codex\plugins\cache\ponytail\ponytail\4.8.4\scripts\uninstall.js` 清理 Ponytail 运行状态。
- 依次执行 `codex plugin remove ponytail@ponytail` 与 `codex plugin marketplace remove ponytail`；最后删除 `work.md` 中本节记录。

## 2026-07-14 16:52:07 +08:00 --- Registry 穿透权威会话并自行提交加入事务，修订冲突被错误降级为存储不可用 --- 将加入事务收回权威游戏会话 [Authoritative Game Session] 并补充冲突回归测试 --- 修改 Backend 会话、注册表、WebSocket Hub、测试与 `work.md`

### 发现什么问题
- `Registry.JoinObserved` 直接访问会话锁、已提交视图、引擎克隆、持久化和观察器投递，注册表 [Registry] 不再只是定位与生命周期管理模块。
- 加入时遇到存储修订冲突 [Revision Conflict] 会返回 `ErrPersistenceUnavailable`，既没有暴露正确语义，也没有将可能分叉的会话标记为不健康。
- 加入事务缺少直接验证候选成员不会泄漏、观察器不会提前触发的会话层回归测试 [Regression Test]。

### 使用什么方式解决
- 新增 `AuthoritativeGameSession.JoinObserved`，由会话独占成员校验、凭证签发、引擎克隆、持久化提交、权威投影发布和观察器投递。
- `Registry` 删除 `Join`/`JoinObserved`，继续只负责创建、查找、恢复和移除；WebSocket Hub 先通过 `Registry.Get` 定位，再调用会话加入接口。
- 从 `JoinInput` 删除冗余 `RoomID`，统一使用会话已提交记录中的房间标识。
- 将 `sessionstore.ErrRevisionConflict` 映射为 `ErrPersistenceConflict` 并标记会话不健康；持久化失败前不发布候选状态，也不调用观察器。
- 新增单一冲突测试，断言修订不前进、候选成员不可查询、已提交投影人数不变、观察器未触发且健康检查失败；未引入共享事务框架 [Transaction Framework]。

### 修改了哪些文件
- `packages/backend/internal/session/types.go`。
- `packages/backend/internal/session/registry.go`。
- `packages/backend/internal/session/session.go`。
- `packages/backend/internal/session/session_test.go`。
- `packages/backend/internal/ws/hub_v2.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 将 `JoinInput`/`JoinResult` 与 `Join`/`JoinObserved` 恢复到 `registry.go`，恢复 `RoomID` 输入，并从 `session.go` 删除会话加入方法。
- 将 `hub_v2.go` 恢复为直接调用 `Registry.JoinObserved`。
- 从 `session_test.go` 删除 `conflictReplace` 测试开关与 `TestJoinRevisionConflictRollsBackAndMarksSessionUnhealthy`。
- 删除 `work.md` 中标题以 `2026-07-14 16:52:07` 开头的本节记录。

## 2026-09-03 12:00:48 +08:00 --- 本机缺少 Go 与项目依赖，项目无法启动 --- 使用 Scoop 安装 Go、按锁文件安装依赖、构建共享包并启动后端与微信小程序监听 --- 修改本机工具链、生成目录与 `work.md`

### 发现什么问题
- 本机没有可用的 `go` 命令，项目要求 Go 1.22+。
- 仓库尚未安装 pnpm 工作区依赖，也没有 `@clocktower/core` 的 `dist` 入口产物，无法直接启动前端。
- 后端启动要求进程内提供至少 32 字节的 `CLOCKTOWER_CREDENTIAL_KEY`。

### 使用什么方式解决
- 通过既有 Scoop 安装 Go 1.27.1（windows/amd64），并以 `go version` 验证安装结果。
- 使用项目锁定的 pnpm 9.15.0 执行冻结锁文件安装，随后直接调用本地 TypeScript 编译器构建 `@clocktower/core`。
- 运行全部 Go 后端测试；使用仅存在于进程内的随机开发凭证启动后端，并启动 Taro 微信小程序监听构建。
- 验证 `GET http://localhost:8080/health` 返回 `200 ok`，前端首次 Webpack 编译成功并生成微信小程序产物。

### 修改了哪些文件
- Scoop 安装目录 `D:\Scoop\apps\go\1.27.1`、`D:\Scoop\apps\go\current` 与 Go/gofmt shim。
- Go 用户模块/构建缓存、Taro 用户配置缓存。
- 工作区忽略目录 `node_modules/`、`packages/core/dist/`、`packages/frontend/dist/`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 在对应开发终端按 `Ctrl+C` 停止后端与前端监听进程。
- 执行 `scoop uninstall go` 卸载本次安装的 Go；如不再需要缓存，可执行 `go clean -cache -modcache`。
- 删除本次生成的 `node_modules/`、`packages/core/dist/` 与 `packages/frontend/dist/`；若 Taro 用户配置目录此前不存在且不再需要，可删除 `C:\Users\Asurf\.taro3.7`。
- 删除 `work.md` 中标题以 `2026-09-03 12:00:48` 开头的本节记录。

## 2026-09-03 14:08:27 +08:00 --- 完整可玩性审计需要核实现有工程基线与页面闭环 --- 按仓库固定版本恢复依赖并执行类型、测试、构建与后端健康检查 --- 刷新忽略的依赖/构建产物并修改 `work.md`

### 发现什么问题
- 当前五个前端路由可以编译，创建/加入、发身份、昼夜、提名投票与结算也已连接权威 WebSocket 状态；“不可玩”主要来自产品流程和规则模型缺口，而非源码无法构建。
- Codex fallback pnpm 与仓库锁定的 pnpm 9.15.0 依赖布局不一致，首次验证时触发依赖目录重建，导致链接包内容不完整。

### 使用什么方式解决
- 使用 `corepack pnpm` 按 `packageManager` 固定的 pnpm 9.15.0 和冻结锁文件强制恢复依赖，没有改动清单或锁文件。
- 直接运行 TypeScript build、Go 全包测试、Vitest 与 Taro 微信小程序构建；检查本地后端 `/health` 返回 `ok`。
- 最终结果：TypeScript 类型检查通过，Go 全包测试通过，Vitest 13 个测试文件共 207 项通过，微信小程序构建成功。

### 修改了哪些文件
- 工作区忽略目录 `node_modules/`、`packages/frontend/dist/` 与 TypeScript/Taro 构建缓存。
- `work.md`。
- 未修改任何产品源码、协议、测试源码或锁文件。

### 撤回方式 [Rollback Strategy]
- 删除本次刷新的忽略目录 `node_modules/`、`packages/frontend/dist/` 及相关构建缓存；需要继续开发时再执行 `corepack pnpm install --frozen-lockfile` 和对应构建命令。
- 删除 `work.md` 中标题以 `2026-09-03 14:08:27` 开头的本节记录。

## 2026-09-03 14:55:26 +08:00 --- 房间缺少可由房主明确设置的顺时针座位顺序 --- 以 WebSocket v2 公开接口完成首个 RED→GREEN 行为切片 --- 修改协议、权威会话、玩法座次、生成契约、集成测试与 `work.md`

### 发现什么问题
- 现有系统把加入顺序隐式当作物理座位顺序，房主无法在身份锁定前按真实围坐顺序调整座位。
- 共情者、厨师以及后续顺时针投票都需要稳定且对所有接收者一致的座位顺序。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 集成测试，描述房主提交完整顺时针座位顺序后，所有成员收到相同权威玩家顺序的外部行为。
- RED 阶段确认测试因缺少 `SET_SEAT_ORDER` 和 `seatOrder` 协议字段而失败。
- GREEN 阶段新增房主专属座位命令；Gameplay 验证提交内容是当前全部玩家 ID 的无重复完整排列，然后直接重排权威 `players` 切片。既有快照、重连与接收者投影自然保留该顺序，不建立平行座位状态。
- 重新生成 Go/TypeScript 协议契约，并通过目标集成测试、协议漂移检查、Go 全包测试及 TypeScript build。

### 修改了哪些文件
- `packages/backend/internal/ws/seat_order_test.go`。
- `proto/game.proto`。
- `packages/backend/internal/session/types.go`。
- `packages/backend/internal/session/session.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/ws/seat_order_test.go`。
- 从 `proto/game.proto` 删除 `SET_SEAT_ORDER` 与 `seat_order`，再运行协议生成器恢复生成契约。
- 从 Session、Gameplay 与 WebSocket Hub 删除 `CommandSetSeatOrder`/`SetSeatOrderCmd` 的分发、权限与重排实现。
- 删除 `work.md` 中标题以 `2026-09-03 14:55:26` 开头的本节记录。

## 2026-09-03 14:58:57 +08:00 --- Core 客户端尚不能发送座位顺序命令 --- 完成单个公开客户端 RED→GREEN 行为切片 --- 修改 WebSocket 客户端、测试与 `work.md`

### 发现什么问题
- 后端协议已接受 `SET_SEAT_ORDER`，但 `GameWebSocketClient` 没有供前端调用的类型安全方法。

### 使用什么方式解决
- 新增一个且仅一个客户端集成测试，要求 `setSeatOrder` 发送带当前客户端序号和完整 `seatOrder` 的权威房间命令。
- RED 阶段确认 `client.setSeatOrder is not a function`。
- GREEN 阶段新增类型安全的 `setSeatOrder` 公共方法，通过既有单航班序列队列发送 `SET_SEAT_ORDER`。
- 目标 Vitest 已通过。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `sends the complete clockwise seat order as a sequenced room command`。
- 删除 `GameWebSocketClient.setSeatOrder`。
- 删除 `work.md` 中标题以 `2026-09-03 14:58:57` 开头的本节记录。

## 2026-09-03 15:00:38 +08:00 --- 前端会话 Store 尚未暴露座位顺序操作 --- 完成单个 Store 公开行为 RED→GREEN 切片 --- 修改 Store、测试与 `work.md`

### 发现什么问题
- 页面层尚无通过应用级会话 Store 调用 Core `setSeatOrder` 的入口。
- 座位提交不能在前端先行篡改权威投影，必须等待服务端提交后的完整 `RoomState`。

### 使用什么方式解决
- 新增一个且仅一个 Store 行为测试，要求完整顺序被转交给 Core、操作进入 `settings` pending 状态且本地权威投影保持不变。
- RED 阶段确认 `setSeatOrder is not a function`。
- GREEN 阶段把 Core `setSeatOrder` 纳入 SessionClient 接缝，并由 RoomSessionState 复用 `settings` pending 通道发送；不维护乐观座位副本。
- 目标 Frontend Vitest 已通过。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 FakeClient 的 `setSeatOrder` 测试接缝和测试 `forwards the complete seat order without changing the authoritative projection locally`。
- 从 `SessionClient`、`RoomSessionState` 和 Store 实现删除 `setSeatOrder`。
- 删除 `work.md` 中标题以 `2026-09-03 15:00:38` 开头的本节记录。

## 2026-09-03 15:02:28 +08:00 --- 设置页会把换座误判为成员变化并重抽角色草稿 --- 完成单个成员集合 RED→GREEN 行为切片 --- 修改设置页、工具、测试与 `work.md`

### 发现什么问题
- `game-setup` 使用保留数组顺序的 `playerKey` 驱动随机配角 effect；权威座位重排会改变该键并无提示地覆盖未提交草稿。

### 使用什么方式解决
- 新增一个且仅一个纯行为测试，要求同一批玩家无论座次如何排列都产生相同成员集合键。
- RED 阶段确认设置页成员工具不存在。
- GREEN 阶段新增顺序无关的 `membershipKey`，设置页只在成员集合真正变化时重新生成角色草稿，单纯换座不再覆盖草稿。
- 目标 Vitest 已通过。

### 修改了哪些文件
- `packages/frontend/src/pages/game-setup/utils.test.ts`。
- `packages/frontend/src/pages/game-setup/utils.ts`。
- `packages/frontend/src/pages/game-setup/index.tsx`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/frontend/src/pages/game-setup/utils.test.ts`。
- 删除 `packages/frontend/src/pages/game-setup/utils.ts`，并恢复 `game-setup/index.tsx` 原有保序 `playerKey` 计算。
- 删除 `work.md` 中标题以 `2026-09-03 15:02:28` 开头的本节记录。

## 2026-09-03 15:08:00 +08:00 --- 设置页缺少可提交的顺时针座位编辑界面 --- 以本地表单草稿接入权威 `SET_SEAT_ORDER` 垂直切片 --- 修改座位工具、编辑组件、设置页样式与测试

### 发现什么问题
- `RoomState.players` 已承载权威顺时针座次，Core 与会话 Store 也已提供提交入口，但设置页仍只能只读显示加入顺序，房主无法在发身份前按线下围坐情况调整座次。
- 前端不得在提交命令后乐观改写权威房间投影，否则拒绝、断线重放或并发更新时会显示并未提交的座次。

### 使用什么方式解决
- RED 阶段只新增一个 `moveSeatClockwise` 行为测试，验证末位玩家顺时针移动后循环到首位；亲自运行确认因函数不存在而失败。
- GREEN 阶段新增最小纯函数实现，并新增 `SeatOrderEditor`：编辑过程只维护明确标注的本机草稿，可还原；确认时一次提交完整玩家 ID 顺序，组件不把草稿写入 Room Experience，等待服务器返回的新 `RoomState.players` 后才重置为权威顺序。
- 设置页仅在“当前身份是房主、已指定说书人、参与者未冻结”三个条件同时满足时呈现编辑器；其他成员继续只读查看权威座位列表。
- 使用可换行的 Flex 布局和窄屏媒体规则兼容微信小程序与 H5。目标测试 2 项、Frontend 全量测试 58 项、TypeScript project build 与微信小程序构建均通过；微信构建在沙箱内受 pnpm 上级目录读取权限限制后，于获准的沙箱外环境成功完成。

### 修改了哪些文件
- `packages/frontend/src/pages/game-setup/utils.test.ts`。
- `packages/frontend/src/pages/game-setup/utils.ts`。
- `packages/frontend/src/pages/game-setup/seat-order-editor.tsx`。
- `packages/frontend/src/pages/game-setup/index.tsx`。
- `packages/frontend/src/pages/game-setup/index.css`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 从 `utils.test.ts` 删除 `moveSeatClockwise` 导入及其“末位循环到首位”测试，并从 `utils.ts` 删除该函数，保留既有 `membershipKey` 行为。
- 删除 `seat-order-editor.tsx`，从 `game-setup/index.tsx` 删除对应导入、`setSeatOrder` selector 与仅房主可见的“调整顺时针座次”区块。
- 从 `game-setup/index.css` 删除所有 `seatOrder*`、`seatMoveButton` 样式及本次窄屏媒体规则。
- 删除 `work.md` 中标题以 `2026-09-03 15:08:00` 开头的本节记录。

## 2026-09-03 15:09:26 +08:00 --- 房间没有玩家准备状态 --- 完成单个 WebSocket 权威投影 RED→GREEN 行为切片 --- 修改协议、玩家模型、玩法命令、生成器、生成契约、集成测试与 `work.md`

### 发现什么问题
- 实际入座玩家无法公开标记自己是否准备完成，说书人也无法从权威状态判断发身份条件。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 集成测试，要求 `p1` 设置准备后，自己、说书人和另一玩家的同一提交投影中只有 `p1` 为已准备。
- RED 阶段确认 `SET_READY`、`ready` 请求字段和 `Player.IsReady` 均不存在。
- GREEN 阶段新增 `SET_READY`，请求必须明确携带布尔值；玩法层只允许未发身份的实际入座玩家修改自己的准备状态。
- 准备状态直接属于权威玩家对象，沿既有克隆、快照、持久化和接收者投影传播，不建立第二份状态。
- 目标 WebSocket 集成测试已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/ready_state_test.go`。
- `proto/game.proto`。
- `scripts/generate-protocol-contracts.mjs`。
- `packages/backend/internal/game/game.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/ws/ready_state_test.go`。
- 从 Proto、生成器覆盖、玩家模型、Gameplay 与 Hub 删除 `SET_READY`/`ready`/`isReady`，再重新生成协议契约。
- 删除 `work.md` 中标题以 `2026-09-03 15:09:26` 开头的本节记录。

## 2026-09-03 15:11:44 +08:00 --- 说书人可以在仍有玩家未准备时发放身份 --- 完成单个 WebSocket 开局门禁 RED→GREEN 切片 --- 修改玩法门禁、集成测试、既有测试夹具与 `work.md`

### 发现什么问题
- `ASSIGN_CHARACTERS` 目前只验证角色组合，不验证所有实际玩家是否完成准备。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 测试：五名玩家中四人准备、一人未准备时，合法角色组合仍必须以稳定 `INVALID_COMMAND` 拒绝，且说书人序号不前进。
- RED 阶段确认未准备的合法角色配置被错误提交并冻结房间。
- GREEN 阶段在 Gameplay 角色分配入口要求所有实际入座玩家均已准备。
- 既有 Gameplay/协议测试夹具显式标记为已准备或通过真实 `SET_READY` 命令准备，不放宽生产规则；TypeScript 房间投影夹具补齐新增必填字段。
- 目标测试与 Go 全包测试均已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/ready_state_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/game_session_test_helpers_test.go`。
- `packages/backend/internal/gameplay/*_test.go` 中所有直接分配角色的既有夹具。
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `packages/backend/internal/ws/protocol_v2_contract_test.go`。
- `packages/frontend/src/lib/room-experience.test.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestProtocolV2StorytellerCannotAssignCharactersUntilEveryPlayerIsReady`。
- 删除角色分配的全员准备门禁和 `markAllPlayersReady` 测试辅助，并撤回既有测试夹具中的准备步骤。
- 从 TypeScript 投影夹具移除为本轮兼容新增的 `isReady` 字段（仅在同时撤回准备协议字段时）。
- 删除 `work.md` 中标题以 `2026-09-03 15:11:44` 开头的本节记录。

## 2026-09-03 15:17:31 +08:00 --- 换座后旧准备状态仍然有效 --- 完成单个准备失效 RED→GREEN 行为切片 --- 修改玩法座次、准备状态集成测试与 `work.md`

### 发现什么问题
- 玩家确认准备所依据的邻接座位发生变化后，现有实现仍保留旧准备标记。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 测试，要求房主真正改变顺时针座次后清空所有实际玩家的准备状态。
- RED 阶段确认换座后的权威投影仍把两名玩家标为已准备。
- GREEN 阶段检测提交顺序是否真正发生变化；仅在真实换座时统一清空所有准备标记，相同顺序的幂等提交不误清空。
- 目标 WebSocket 测试已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/ready_state_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestProtocolV2ChangingSeatOrderClearsEveryPlayersReadyState`。
- 删除 `clearReadyLocked` 及座位变更检测/清空调用，恢复直接重排。
- 删除 `work.md` 中标题以 `2026-09-03 15:17:31` 开头的本节记录。

## 2026-09-03 15:19:16 +08:00 --- 新玩家加入后旧准备状态仍然有效 --- 以单个成员变化测试完成 RED-GREEN --- 修改准备状态集成测试、游戏会话与 `work.md`

### 发现什么问题
- 已有玩家确认准备后，房间仍允许新玩家加入，但原准备标记不会失效。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 测试，要求新成员加入后，重连投影中的全部实际玩家均恢复未准备。
- RED 阶段确认 P2 加入后 P1 仍为 `isReady=true`，测试按预期失败。
- GREEN 阶段仅在 `GameSession.AddPlayer` 的实际新增分支追加全员准备状态清理；已有 ID 的重连/改名分支提前返回，不会误清准备状态。
- 定向测试 `go test ./internal/ws -run TestProtocolV2JoiningPlayerClearsExistingReadyState -count=1` 已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/ready_state_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestProtocolV2JoiningPlayerClearsExistingReadyState`。
- 删除 `GameSession.AddPlayer` 新增玩家分支中的 `gs.clearReadyLocked()` 调用。
- 删除 `work.md` 中标题以 `2026-09-03 15:19:16` 开头的本节记录。

## 2026-09-03 15:23:19 +08:00 --- 玩家离开后剩余准备状态可能过期 --- 以单个成员移除测试完成 RED-GREEN --- 修改准备状态集成测试、游戏会话与 `work.md`

### 发现什么问题
- 玩家确认准备后，其他玩家仍可能在角色分配前离开；现有实现尚未证明剩余玩家会重新确认当前成员构成。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 测试：P1 准备后 P2 离开，P1 重连投影必须显示未准备。
- 首次运行发现通用命令夹具错误地要求主动离房回执携带房间状态；已校准为校验离房者无状态成功回执，并从仍在房间的 P1 广播投影断言业务行为。
- RED 阶段确认 P2 离开后 P1 仍为 `isReady=true`，测试按预期失败。
- GREEN 阶段仅在 `GameSession.RemovePlayer` 确实移除一名座位玩家时清空剩余玩家准备状态；无匹配 ID 时不产生副作用。
- 定向测试 `go test ./internal/ws -run TestProtocolV2LeavingPlayerClearsRemainingReadyState -count=1` 已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/ready_state_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestProtocolV2LeavingPlayerClearsRemainingReadyState`。
- 删除 `GameSession.RemovePlayer` 移除分支中的 `gs.clearReadyLocked()` 调用。
- 删除 `work.md` 中标题以 `2026-09-03 15:23:19` 开头的本节记录。

## 2026-09-03 15:25:48 +08:00 --- 选择说书人会改变实际座位玩家集合 --- 以单个说书人选择测试完成 RED-GREEN --- 修改准备状态集成测试、游戏会话与 `work.md`

### 发现什么问题
- 协议允许玩家在房主选择说书人前确认准备；被选者随后从实际玩家列表移除，因此其他人的旧准备确认必须失效。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 测试：P1 先准备，房主再把 P2 设为说书人，权威投影中的 P1 必须恢复未准备。
- RED 阶段确认 P2 成为说书人并退出玩家集合后，P1 仍为 `isReady=true`，测试按预期失败。
- GREEN 阶段仅在 `applySetStoryteller` 成功重建座位玩家列表后清空全员准备状态。
- 定向测试 `go test ./internal/ws -run TestProtocolV2SelectingStorytellerClearsRemainingReadyState -count=1` 已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/ready_state_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestProtocolV2SelectingStorytellerClearsRemainingReadyState`。
- 删除 `applySetStoryteller` 成功分支中的 `gs.clearReadyLocked()` 调用。
- 删除 `work.md` 中标题以 `2026-09-03 15:25:48` 开头的本节记录。

## 2026-09-03 15:27:50 +08:00 --- Core 尚未暴露准备/取消准备命令 --- 以单个 `false` 序列化测试完成 RED-GREEN --- 修改 WebSocket 客户端及测试与 `work.md`

### 发现什么问题
- 后端已有 `SET_READY`，但共享 Core 客户端没有公共方法；尤其取消准备的 `false` 值容易被条件展开错误遗漏。

### 使用什么方式解决
- 新增一个且仅一个 Core WebSocket 公共接口测试，要求 `setReady(false)` 发出带客户端序号且显式包含 `ready:false` 的命令。
- RED 阶段确认 `client.setReady` 不存在，测试按预期失败。
- GREEN 阶段新增 `setReady(ready)`，直接通过 `sendSequenced` 发送完整布尔值，不会丢失 `false`。
- 定向 Vitest 已通过（1 passed，20 skipped）。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `sends false ready state as a sequenced room command`。
- 删除 `GameWebSocketClient.setReady` 方法。
- 删除 `work.md` 中标题以 `2026-09-03 15:27:50` 开头的本节记录。

## 2026-09-03 15:29:39 +08:00 --- 前端 Store 尚未转发准备命令 --- 以单个权威投影测试完成 RED-GREEN --- 修改 Store、测试与 `work.md`

### 发现什么问题
- Core 已能发送准备命令，但页面状态层尚无 `setReady` 操作；实现还必须避免在服务器确认前乐观改写权威玩家投影。

### 使用什么方式解决
- 为测试替身新增 `setReady` 调用记录能力。
- 新增一个且仅一个 Store 公共接口测试：转发 `true`、设置 `pendingCommand='ready'`，同时保持服务器投影中的 `isReady=false` 不变。
- RED 阶段确认 Store 没有 `setReady` 方法，测试按预期失败。
- GREEN 阶段增加 `ready` 待处理类型、Core 客户端能力及 `setReady` 公共操作；实现只调用 `send`，没有乐观修改房间投影。
- 定向 Vitest 已通过（1 passed，11 skipped）。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards readiness without changing the authoritative projection locally` 及测试替身的 `setReady` 方法。
- 从 `PendingCommand`、`SessionClient`、`RoomSessionState` 和 Store 返回对象中删除准备状态接线。
- 删除 `work.md` 中标题以 `2026-09-03 15:29:39` 开头的本节记录。

## 2026-09-03 15:31:37 +08:00 --- 准备 UI 缺少可测试的权威汇总策略 --- 以单个纯函数测试完成 RED-GREEN --- 修改 setup 工具、测试与 `work.md`

### 发现什么问题
- setup 页面需要同时判断本人准备状态、已准备人数与是否全员准备；直接散落在 JSX 中会让发放身份门禁难以独立验证。

### 使用什么方式解决
- 新增一个且仅一个纯函数测试，要求两名玩家中一人准备时返回 `1/2`、`allReady=false`，并识别当前玩家已准备。
- RED 阶段确认 `summarizeReadiness` 不存在，测试按预期失败。
- GREEN 阶段实现纯函数：计算准备人数/总人数，空房不视为全员准备，并在本人不属于实际座位玩家时返回 `selfReady=null`。
- 定向 Vitest 已通过（1 passed，2 skipped）。

### 修改了哪些文件
- `packages/frontend/src/pages/game-setup/utils.test.ts`。
- `packages/frontend/src/pages/game-setup/utils.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `summarizes authoritative readiness for the current player and the room` 及新增导入。
- 删除 `summarizeReadiness` 函数。
- 删除 `work.md` 中标题以 `2026-09-03 15:31:37` 开头的本节记录。

## 2026-09-03 15:33:50 +08:00 --- 页面发放按钮尚未受全员准备约束 --- 以单个发放决策测试完成 RED-GREEN 并接入准备 UI --- 修改 setup 工具、页面、样式、测试与 `work.md`

### 发现什么问题
- 即使后端已有硬门禁，前端在角色配置合法时仍会显示可发放，无法向说书人准确表达“仍在等待玩家准备”。

### 使用什么方式解决
- 新增一个且仅一个纯决策测试，要求角色配置合法但未全员准备时仍禁止发放。
- RED 阶段确认 `canAssignCharacters` 不存在，测试按预期失败。
- GREEN 阶段实现“配置合法且全员准备”决策，并把已测试的准备汇总/决策薄接入 setup 页面。
- 实际座位玩家现在可准备或取消；成员列表公开显示准备徽标；说书人看到人数进度且全员准备前无法发放身份。所有状态与按钮文案只读取服务器权威投影，没有本地乐观改写。
- 依据 Context7 返回的 Taro 官方文档，沿用跨微信小程序/H5 的 `Button disabled` 与 `onClick` 接口。
- 准备流程相关 Vitest 共 37 条通过，全仓 TypeScript 构建检查通过；Taro H5 与微信小程序构建均成功。H5 仅保留既有入口资源体积警告（`app.js` 约 256 KiB），无编译错误。

### 修改了哪些文件
- `packages/frontend/src/pages/game-setup/utils.test.ts`。
- `packages/frontend/src/pages/game-setup/utils.ts`。
- `packages/frontend/src/pages/game-setup/index.tsx`。
- `packages/frontend/src/pages/game-setup/index.css`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `does not allow assignment while any seated player is unready` 及新增导入。
- 删除 `canAssignCharacters`，并从 setup 页面移除 `setReady`、准备汇总/徽标/操作区及全员准备发放门禁。
- 删除 `index.css` 中 `.readyBadge*`、`.readiness*` 样式。
- 删除 `work.md` 中标题以 `2026-09-03 15:33:50` 开头的本节记录。

## 2026-09-03 15:37:44 +08:00 --- 身份发放后缺少玩家确认闭环 --- 以单个接收者投影集成测试完成 RED-GREEN --- 新增身份确认协议、领域实现、测试并修改 `work.md`

### 发现什么问题
- 服务器会私下投影角色，但没有“玩家已查看并确认”的持久状态，说书人无法判断能否安全开始首夜。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 公共接口测试：P1 确认自己的已分配角色后，本人、说书人和其他玩家均看到确认状态；P2 仍看不到 P1 的角色且不会被代确认。
- RED 阶段确认 `CONFIRM_CHARACTER` 与 `hasConfirmedCharacter` 均不存在，测试按预期编译失败。
- GREEN 阶段在 Proto 中增加必填公开确认字段与无载荷序列命令，重新生成 Go/TypeScript 契约；领域命令仅允许 setup 阶段、已获身份的实际座位玩家确认自己。
- 分配角色时显式重置确认标记；接收者投影继续沿用既有角色脱敏，仅公开确认状态。
- 定向测试 `go test ./internal/ws -run TestProtocolV2PlayerConfirmsOwnAssignedCharacter -count=1` 已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/character_confirmation_test.go`。
- `proto/game.proto`。
- `packages/backend/internal/game/game.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/protocol_generated.go`（生成产物）。
- `packages/core/src/websocket/protocol.generated.ts`（生成产物）。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/ws/character_confirmation_test.go`。
- 从 Proto 删除 `Player.has_confirmed_character` 与 `CONFIRM_CHARACTER`，重新运行 `node scripts/generate-protocol-contracts.mjs`。
- 从 `game.Player`、`GameSession` 和 WebSocket 映射中删除对应字段、命令、分配时重置和确认处理函数。
- 删除 `work.md` 中标题以 `2026-09-03 15:37:44` 开头的本节记录。

## 2026-09-03 15:54:52 +08:00 --- 说书人仍可在玩家未确认身份时开局 --- 以单个全员确认门禁测试完成 RED-GREEN --- 修改开局规则、身份确认测试与 `work.md`

### 发现什么问题
- 确认状态已经可持久化和投影，但 `START_GAME` 尚未使用它，仍可能在最后一名玩家未看清身份时进入首夜。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 测试：五名玩家中仅四人确认时，`START_GAME` 必须返回 `INVALID_COMMAND`，且失败命令不能消耗说书人的客户端序号。
- RED 阶段确认未确认的 P5 不会阻止开局，测试按预期失败。
- GREEN 阶段在现有角色完整性检查中追加全员 `HasConfirmedCharacter` 硬门禁，不改变其他开局逻辑。
- 定向测试 `go test ./internal/ws -run TestProtocolV2StorytellerCannotStartUntilEveryPlayerConfirmsCharacter -count=1` 已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/character_confirmation_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestProtocolV2StorytellerCannotStartUntilEveryPlayerConfirmsCharacter`。
- 删除 `applyStartGame` 中的 `HasConfirmedCharacter` 检查。
- 删除 `work.md` 中标题以 `2026-09-03 15:54:52` 开头的本节记录。

## 2026-09-03 15:59:41 +08:00 --- 新开局门禁使旧规则测试缺少明确前置条件 --- 显式补齐测试身份确认并完成全量 Go 回归 --- 修改 gameplay/ws 测试夹具与 `work.md`

### 发现什么问题
- 全量 Go 回归显示，旧的夜间/角色规则测试在分配身份后直接开局，未表达“玩家已经查看并确认身份”的新公共前置条件；完整 WebSocket 流程的修订号与玩家序号也因此需要顺延。

### 使用什么方式解决
- 新增测试辅助函数 `markAllPlayersConfirmed`，只在意图测试开局后规则的 gameplay 用例中，于 `StartGameCmd` 前显式确认全部现有玩家。
- 完整 WebSocket 与协议契约用例通过真实 `CONFIRM_CHARACTER` 命令确认每名玩家；同步把新增五次提交后的固定修订号 `20/23` 调整为 `25/28`，P2 序号由 `2` 调整为 `3`。
- 未修改或放宽生产门禁；`go test ./...` 已全部通过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/game_session_test_helpers_test.go`。
- `packages/backend/internal/gameplay/drunk_test.go`。
- `packages/backend/internal/gameplay/chef_test.go`。
- `packages/backend/internal/gameplay/empath_test.go`。
- `packages/backend/internal/gameplay/first_night_information_test.go`。
- `packages/backend/internal/gameplay/fortune_teller_test.go`。
- `packages/backend/internal/gameplay/game_session_test.go`。
- `packages/backend/internal/gameplay/poison_test.go`。
- `packages/backend/internal/gameplay/ravenkeeper_test.go`。
- `packages/backend/internal/gameplay/undertaker_test.go`。
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `packages/backend/internal/ws/protocol_v2_contract_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `markAllPlayersConfirmed` 及上述 gameplay 测试中紧邻 `StartGameCmd` 的对应调用。
- 删除两个 WebSocket 测试中的确认命令循环，并把完整流程固定修订号/序号恢复为 `20/23` 与 `2`。
- 删除 `work.md` 中标题以 `2026-09-03 15:59:41` 开头的本节记录。

## 2026-09-03 16:00:37 +08:00 --- 新必填身份确认字段暴露三个旧 TS 夹具缺口 --- 显式补充未确认初值并通过类型检查 --- 修改前端测试夹具与 `work.md`

### 发现什么问题
- 生成契约把 `hasConfirmedCharacter` 正确设为必填后，全仓 TypeScript 检查定位到三个旧 `RoomPlayer` 夹具没有声明该字段。

### 使用什么方式解决
- 在对应的初始/死亡/准备状态测试玩家中显式加入 `hasConfirmedCharacter:false`，不把协议字段降级为可选。
- `.\\node_modules\\.bin\\tsc.cmd --build` 已通过。

### 修改了哪些文件
- `packages/frontend/src/lib/room-experience.test.ts`。
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 从三个测试玩家夹具中删除 `hasConfirmedCharacter:false`（仅应与整个身份确认协议一并撤回）。
- 删除 `work.md` 中标题以 `2026-09-03 16:00:37` 开头的本节记录。

## 2026-09-03 16:00:56 +08:00 --- Core 尚未暴露身份确认命令 --- 以单个序列命令测试完成 RED-GREEN --- 修改 WebSocket 客户端、测试与 `work.md`

### 发现什么问题
- 服务端已支持当前连接玩家确认自己的角色，但共享客户端没有公共方法供 H5/微信页面调用。

### 使用什么方式解决
- 新增一个且仅一个 Core WebSocket 测试，要求 `confirmCharacter()` 发出带当前身份和客户端序号的无目标 `CONFIRM_CHARACTER` 命令。
- RED 阶段确认 `client.confirmCharacter` 不存在，测试按预期失败。
- GREEN 阶段新增无参数方法并复用 `sendSequenced`；玩家身份只取当前认证连接，不允许客户端指定被确认者。
- 定向 Vitest 已通过（1 passed，21 skipped）。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `confirms the current players character as a sequenced room command`。
- 删除 `GameWebSocketClient.confirmCharacter` 方法。
- 删除 `work.md` 中标题以 `2026-09-03 16:00:56` 开头的本节记录。

## 2026-09-03 16:02:02 +08:00 --- 前端 Store 尚未转发身份确认 --- 以单个权威投影测试完成 RED-GREEN --- 修改 Store、测试与 `work.md`

### 发现什么问题
- 页面状态层尚不能发起身份确认；实现还必须保证服务器拒绝或断线时不会乐观显示为已确认。

### 使用什么方式解决
- 为测试替身增加 `confirmCharacter` 调用记录。
- 新增一个且仅一个 Store 测试：转发无参数确认、设置 `pendingCommand='confirm-character'`，同时保持权威投影中的确认值为 `false`。
- RED 阶段确认 Store 没有 `confirmCharacter` 方法，测试按预期失败。
- GREEN 阶段新增待处理类型、Core 能力白名单与 Store 公共方法；只转发请求，不改本地房间状态。
- 定向 Vitest 已通过（1 passed，12 skipped）。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards character confirmation without changing the authoritative projection locally` 及测试替身方法。
- 从 `PendingCommand`、`SessionClient`、`RoomSessionState` 与 Store 返回对象中删除身份确认接线。
- 删除 `work.md` 中标题以 `2026-09-03 16:02:02` 开头的本节记录。

## 2026-09-03 16:03:56 +08:00 --- 私密角色卡缺少可验证的遮罩状态模型 --- 以单个初始/揭示时限测试完成 RED-GREEN --- 新增可见性模块、测试并修改 `work.md`

### 发现什么问题
- 当前角色会直接明文显示；在接入触摸和页面生命周期前，需要先固定“默认遮罩、首次按住揭示、最长 30 秒”的纯状态语义。

### 使用什么方式解决
- 新增一个且仅一个纯函数测试，要求初始状态不可见且未查看；按下时标记已查看并记录当前时间后 30 秒的强制遮罩时刻。
- RED 阶段确认可见性模块不存在，测试按预期加载失败。
- GREEN 阶段新增纯状态模型、默认遮罩工厂、30 秒常量和按下揭示转换。
- 定向 Vitest 已通过（1 passed）。

### 修改了哪些文件
- `packages/frontend/src/components/private-character-visibility.test.ts`。
- `packages/frontend/src/components/private-character-visibility.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/frontend/src/components/private-character-visibility.test.ts`。
- 删除 `packages/frontend/src/components/private-character-visibility.ts`。
- 删除 `work.md` 中标题以 `2026-09-03 16:03:56` 开头的本节记录。

## 2026-09-03 16:05:16 +08:00 --- 私密卡尚无统一立即遮罩转换 --- 以单个隐藏状态测试完成 RED-GREEN --- 修改可见性模块、测试与 `work.md`

### 发现什么问题
- 松手、触摸取消、切后台和超时都需要同一种安全隐藏语义；若隐藏时把“已查看”也清掉，玩家将无法在松手后确认身份。

### 使用什么方式解决
- 新增一个且仅一个纯函数测试，要求隐藏后 `revealed=false`、截止时间清空，同时保留 `hasRevealed=true`。
- RED 阶段确认 `concealPrivateCharacter` 不存在，测试按预期失败。
- GREEN 阶段新增统一隐藏转换，供松手、取消、后台与超时路径共同调用。
- 定向 Vitest 已通过（1 passed，1 skipped）。

### 修改了哪些文件
- `packages/frontend/src/components/private-character-visibility.test.ts`。
- `packages/frontend/src/components/private-character-visibility.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `conceals immediately without forgetting that the character was viewed` 及新增导入。
- 删除 `concealPrivateCharacter` 函数。
- 删除 `work.md` 中标题以 `2026-09-03 16:05:16` 开头的本节记录。

## 2026-09-03 16:06:38 +08:00 --- setup 缺少身份确认进度决策 --- 以单个权威确认汇总测试完成 RED-GREEN --- 修改 setup 工具、测试与 `work.md`

### 发现什么问题
- 服务端会拒绝未全员确认时开局，但 setup 页面还无法提前显示确认进度和禁用开局操作。

### 使用什么方式解决
- 新增一个且仅一个纯函数测试，要求两名玩家中一人确认时返回 `1/2`、`allConfirmed=false`，并识别当前玩家未确认。
- RED 阶段确认 `summarizeCharacterConfirmation` 不存在，测试按预期失败。
- GREEN 阶段实现确认人数/总人数、非空全员确认与本人确认状态汇总。
- 定向 Vitest 已通过（1 passed，4 skipped）。

### 修改了哪些文件
- `packages/frontend/src/pages/game-setup/utils.test.ts`。
- `packages/frontend/src/pages/game-setup/utils.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `summarizes authoritative character confirmations before the first night` 及新增导入。
- 删除 `summarizeCharacterConfirmation` 函数。
- 删除 `work.md` 中标题以 `2026-09-03 16:06:38` 开头的本节记录。

## 2026-09-03 16:08:18 +08:00 --- setup/play 明文显示玩家身份且缺少确认交互 --- 接入跨端按住揭示私密卡与权威确认进度 --- 新增组件/样式并修改两页与 `work.md`

### 发现什么问题
- setup 成员行、setup 身份面板、play 页面标题/成员行/侧栏会持续明文展示玩家角色；没有松手隐藏、后台隐藏、30 秒强制隐藏、水印或“看过后确认”交互。

### 使用什么方式解决
- 新增复用 `PrivateCharacterCard`：遮罩状态下不渲染角色名称/能力；触摸按下才显示，触摸结束/取消、Taro `useDidHide` 和 30 秒定时器共用立即遮罩转换。
- 揭示层叠加房间号与玩家名水印，不触发震动或声音；明确提示截图无法从技术上阻止的屏幕私密风险。
- setup 玩家至少查看一次后才能提交确认，按钮状态只读服务器 `hasConfirmedCharacter`；说书人看到确认进度，全员确认前“开始游戏”禁用。
- setup/play 的非说书人标题和成员行不再明文泄漏本人角色；说书人仍保留完整角色视图。
- Context7 的 Taro 4 官方文档确认 `useDidHide` 覆盖微信小程序与 H5；本地 Taro 4 类型确认 `View` 支持 `onTouchStart/onTouchEnd/onTouchCancel`。
- 相关 Vitest 共 42 条通过，全仓 TypeScript 与 Proto 生成漂移检查通过；Taro H5/微信小程序构建均成功。H5 仅有既有入口资源体积警告，无编译错误。
- 本地真实 H5 多会话验收：一个说书人页面与五个独立来源玩家页面共同建房/入房，五人准备实时汇总为 `5/5`，身份成功发放；玩家 DOM 仅包含遮罩提示，不含“间谍”等角色明文，确认按钮在未查看前确实为禁用；390×844 移动端视口截图显示卡片和操作区无横向溢出。

### 修改了哪些文件
- `packages/frontend/src/components/private-character-card.tsx`。
- `packages/frontend/src/components/private-character-card.css`。
- `packages/frontend/src/pages/game-setup/index.tsx`。
- `packages/frontend/src/pages/game-play/index.tsx`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `private-character-card.tsx/.css`，恢复 setup/play 原有角色面板、标题与成员行显示。
- 从 setup 页面移除确认汇总、`confirmCharacter` 接线和开局按钮确认门禁。
- 删除 `work.md` 中标题以 `2026-09-03 16:08:18` 开头的本节记录。

## 2026-09-03 16:22:39 +08:00 --- 死亡玩家投反对也会消耗幽灵票 --- 以单个幽灵票保留测试完成 RED-GREEN --- 新增投票规则测试并修改投票实现与 `work.md`

### 发现什么问题
- `applyCastVote` 在读取赞成/反对决定前就标记死亡玩家已使用幽灵票，违反“只有死者投赞成才消耗一次性票”的规则。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 公共命令测试：死者投反对后投票成功、内部未标记消耗，公开投影仍列出该玩家的可用幽灵票。
- RED 阶段确认反对票会立即把 `ghostVotesUsed` 标记为真，测试按预期失败。
- GREEN 阶段把幽灵票可用性检查和消耗限定为 `!IsAlive && Decision`；死者仍可在后续提名中投反对。
- 定向 Go 测试已通过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/gameplay/voting_rules_test.go` 中 `TestDeadPlayerVotingNoKeepsGhostVote`。
- 把 `applyCastVote` 的死者幽灵票条件恢复为不区分赞成/反对。
- 删除 `work.md` 中标题以 `2026-09-03 16:22:39` 开头的本节记录。

## 2026-09-03 16:24:12 +08:00 --- 被拒绝的死者重复赞成票仍会消耗幽灵票 --- 以单个无副作用测试完成 RED-GREEN --- 修改投票规则测试、投票实现与 `work.md`

### 发现什么问题
- 幽灵票消耗发生在重复投票校验之前，导致命令最终报错但一次性资源已经被扣除。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：已投反对的死者再投赞成必须被拒绝，且 `ghostVotesUsed` 仍为假。
- RED 阶段确认重复赞成票报错前已经消耗幽灵票，测试按预期失败。
- GREEN 阶段把幽灵票检查/扣除移动到重复投票和管家能力校验之后，只有通过全部校验的死者赞成票才会扣除资源。
- 定向 Go 测试已通过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestRejectedDuplicateYesVoteDoesNotSpendGhostVote`。
- 把 `applyCastVote` 的幽灵票检查/扣除块移回重复投票校验之前。
- 删除 `work.md` 中标题以 `2026-09-03 16:24:12` 开头的本节记录。

## 2026-09-03 16:26:41 +08:00 --- 过半票会被错误地立即处决并结束白天 --- 以单个“上台而非处决”测试完成 RED-GREEN --- 修改投票历史/投影/快照、规则测试与 `work.md`

### 发现什么问题
- 现有 `ResolveNomination` 把达到存活半数直接等同于处决、死亡和进入夜晚，无法继续当天后续提名，也无法比较全日最高票。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：5 人局获 3 票者应成为唯一待处决候选并记录票数，但仍存活、无死亡记录、阶段回到白天。
- RED 阶段确认待处决候选投影不存在，且旧结算会立即杀人入夜。
- GREEN 阶段新增不可变 `NominationResult` 历史（含逐人票型、赞成/反对/门槛和天数）并纳入快照；候选/最高票/并列从当天历史派生，避免散字段漂移。
- 提名结算现在只追加公开记录并回到白天，不再杀死被提名者；定向 Go 测试已通过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/game/game.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `packages/backend/internal/gameplay/projection.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestResolvedMajorityPutsNomineeOnBlockWithoutExecuting`。
- 删除 `game.NominationResult`、GameSession/快照中的提名历史、投影候选字段及派生函数，并恢复 `applyResolveNomination` 的即时处决/入夜逻辑。
- 删除 `work.md` 中标题以 `2026-09-03 16:26:41` 开头的本节记录。

## 2026-09-03 16:29:59 +08:00 --- 日终缺少“系统候选→说书人确认→处决”事务 --- 以单个日终确认测试完成 RED-GREEN --- 修改日终命令/规则、投票测试与 `work.md`

### 发现什么问题
- 提名已能产生候选，但没有单一、显式的说书人日终确认命令来执行唯一候选并推进夜晚；复用普通切阶段会绕过人工确认语义。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：说书人提交 `FinalizeDayCmd` 后，唯一候选产生处决死亡记录/事件，随后进入夜晚。
- RED 阶段确认 `FinalizeDayCmd` 不存在，测试按预期编译失败。
- GREEN 阶段新增显式日终命令：仅说书人、白天且无活动提名时可提交；唯一候选在该事务内处决并检查胜负，否则检查市长无处决胜利，未结束才进入夜晚。
- 定向 Go 测试已通过；普通 `CHANGE_PHASE` 的绕过路径将在下一条独立测试中关闭。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestStorytellerFinalizesDayByExecutingUniqueCandidateAndStartingNight`。
- 删除 `FinalizeDayCmd`、命令标记/分派与 `applyFinalizeDay`。
- 删除 `work.md` 中标题以 `2026-09-03 16:29:59` 开头的本节记录。

## 2026-09-03 16:31:35 +08:00 --- 普通切阶段仍可绕过日终确认事务 --- 以单个旁路拒绝测试完成 RED-GREEN 并清理旧路径 --- 修改投票规则测试、阶段实现与 `work.md`

### 发现什么问题
- 新增 `FinalizeDayCmd` 后，旧 `ChangePhaseCmd(day→night)` 仍能直接切夜晚，跳过候选处决、市长判断和日终人工确认。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试，要求白天到夜晚的普通切阶段命令被拒绝且阶段保持白天。
- RED 阶段确认 `CHANGE_PHASE` 仍可直接进入夜晚，测试按预期失败。
- GREEN 阶段明确要求使用 `FINALIZE_DAY`；重构删除旧切阶段中的市长/中毒/开夜副作用，并把黄昏中毒清理迁入唯一日终事务。
- 定向 Go 测试已通过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestChangePhaseCannotBypassDayFinalization`。
- 恢复 `applyChangePhase` 的 day→night 分支及其中市长、黄昏清理和开夜逻辑，并从 `applyFinalizeDay` 删除迁移的黄昏清理。
- 删除 `work.md` 中标题以 `2026-09-03 16:31:35` 开头的本节记录。

## 2026-09-03 16:33:52 +08:00 --- 日终候选与票史尚未进入 WebSocket 公共契约 --- 以单个五连接纵切面测试完成 RED-GREEN --- 新增协议/映射/投影与投票流程测试并修改 `work.md`

### 发现什么问题
- 领域层已有正确候选和日终命令，但 Proto/Hub/接收者投影尚无对应字段与命令，H5/微信客户端无法显示或确认。

### 使用什么方式解决
- 新增一个且仅一个 WebSocket v2 纵切面测试：五名玩家完成首夜，P2 获 3 票后投影公开唯一候选与一条完整票史；说书人 `FINALIZE_DAY` 后 P2 才死亡并进入夜晚。
- RED 阶段确认 RoomState 候选/票史字段和 `FINALIZE_DAY` 均不存在，测试按预期编译失败。
- GREEN 阶段在 Proto 增加 `NominationResult`、候选/最高票/并列投影及日终命令，更新生成器的领域类型映射，重新生成 Go/TypeScript 契约并接入 Hub/SessionEngine。
- 候选 ID、票数与并列被建模为同一组可选状态：无候选时可整体省略，客户端按 `0/false` 解释；避免要求所有非投票房间夹具伪造候选值。全仓 TypeScript 检查通过。
- 定向五连接 WebSocket 测试已通过。

### 修改了哪些文件
- `packages/backend/internal/ws/voting_flow_test.go`。
- `proto/game.proto`。
- `scripts/generate-protocol-contracts.mjs`。
- `packages/backend/internal/ws/protocol_generated.go`（生成产物）。
- `packages/core/src/websocket/protocol.generated.ts`（生成产物）。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/session_engine.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/ws/voting_flow_test.go`。
- 从 Proto 删除 `NominationResult`、RoomState 新字段及 `FINALIZE_DAY`，恢复生成器映射后重新生成契约。
- 删除 Hub 的日终映射与 SessionEngine 的候选/票史投影赋值。
- 删除 `work.md` 中标题以 `2026-09-03 16:33:52` 开头的本节记录。

## 2026-09-03 16:37:47 +08:00 --- 旧规则测试仍通过普通切阶段或期待过半即结束 --- 迁移到显式日终确认并完成全量 Go 回归 --- 修改 gameplay/ws 测试与 `work.md`

### 发现什么问题
- 全量回归中，夜间轮转、市长、中毒、守鸦人、入殓师等旧测试仍用 `ChangePhaseCmd(day→night)`；完整 WebSocket 流程则仍期待 `RESOLVE_NOMINATION` 直接处决恶魔。

### 使用什么方式解决
- 将意图为“结束白天”的测试调用迁移到 `FinalizeDayCmd`，保留专门的旁路拒绝测试继续使用 `ChangePhaseCmd`。
- 完整 WebSocket 流程先断言恶魔成为候选，再显式 `FINALIZE_DAY`，最终修订号由 28 顺延至 29。
- `go test ./...` 已全部通过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/endgame_test.go`。
- `packages/backend/internal/gameplay/day_number_test.go`。
- `packages/backend/internal/gameplay/game_session_test.go`。
- `packages/backend/internal/gameplay/poison_test.go`。
- `packages/backend/internal/gameplay/ravenkeeper_test.go`。
- `packages/backend/internal/gameplay/undertaker_test.go`。
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 将上述旧测试的 `FinalizeDayCmd` 调用恢复为原 `ChangePhaseCmd(day→night)`；删除完整流程新增候选断言/日终命令并把最终修订号恢复为 28（仅应与整套日终事务一并撤回）。
- 删除 `work.md` 中标题以 `2026-09-03 16:37:47` 开头的本节记录。

## 2026-09-03 16:39:27 +08:00 --- Core 尚未暴露日终确认命令 --- 以单个序列命令测试完成 RED-GREEN --- 修改 WebSocket 客户端、测试与 `work.md`

### 发现什么问题
- `FINALIZE_DAY` 已存在于服务端协议，但共享 Core 客户端尚无公共调用方法。

### 使用什么方式解决
- 新增一个且仅一个 Core WebSocket 测试，要求 `finalizeDay()` 发出带当前认证身份和客户端序号的无载荷命令。
- RED 阶段确认 `client.finalizeDay` 不存在，测试按预期失败。
- GREEN 阶段新增 `finalizeDay()` 并复用序列命令队列。
- 定向 Vitest 已通过（1 passed，22 skipped）。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `finalizes the day as a sequenced room command`。
- 删除 `GameWebSocketClient.finalizeDay` 方法。
- 删除 `work.md` 中标题以 `2026-09-03 16:39:27` 开头的本节记录。

## 2026-09-03 16:40:29 +08:00 --- 前端 Store 尚未转发日终确认 --- 以单个待处理态测试完成 RED-GREEN --- 修改 Store、测试与 `work.md`

### 发现什么问题
- Core 已支持 `FINALIZE_DAY`，但页面状态层没有对应动作和独立待处理态，无法防止日终确认重复提交。

### 使用什么方式解决
- 为测试替身增加 `finalizeDay` 调用记录。
- 新增一个且仅一个 Store 测试，要求调用被无参数转发且 `pendingCommand='finalize-day'`。
- RED 阶段确认 Store 没有 `finalizeDay` 方法，测试按预期失败。
- GREEN 阶段增加独立待处理类型、Core 能力白名单与 Store 公共转发方法。
- 定向 Vitest 已通过（1 passed，13 skipped）。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards day finalization as its own pending command` 及测试替身方法。
- 从 Store 的待处理类型、客户端能力、公开接口和返回对象中删除 `finalizeDay` 接线。
- 删除 `work.md` 中标题以 `2026-09-03 16:40:29` 开头的本节记录。

## 2026-09-03 16:41:48 +08:00 --- 日终 UI 缺少“候选不是即时处决”表达 --- 以单个候选文案测试完成 RED-GREEN 并接入日终确认 UI --- 新增 game-play 工具/测试并修改页面、样式与 `work.md`

### 发现什么问题
- 后端已把投票与处决分离，但前端尚无统一文案向全桌说明当前最高票只是等待说书人确认的提案。

### 使用什么方式解决
- 新增一个且仅一个纯函数测试：唯一候选文案包含玩家名、票数，并明确“确认日终后才会被处决”。
- RED 阶段确认 `executionProposalText` 模块不存在，测试按预期加载失败。
- GREEN 阶段实现唯一候选、并列、无人上台三种文案；play 页面公开显示当前提案和当天逐轮票数/门槛。
- 只有说书人能点击“确认日终并进入夜晚”，二次确认弹窗后才调用 `finalizeDay`；删除旧 `changePhase('night')` 页面入口。
- 定向 Vitest 已通过。
- TypeScript 首轮检查发现异步确认函数闭包中的 `room` 未被控制流永久收窄；改用可选访问后，全仓 `tsc --build` 已通过。

### 修改了哪些文件
- `packages/frontend/src/pages/game-play/utils.test.ts`。
- `packages/frontend/src/pages/game-play/utils.ts`。
- `packages/frontend/src/pages/game-play/index.tsx`。
- `packages/frontend/src/pages/game-play/index.css`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/frontend/src/pages/game-play/utils.test.ts`。
- 删除 `utils.ts`，移除 play 页的处决提案/票史/确认弹窗，恢复说书人工具中的 `changePhase('night')` 按钮并删除新增样式。
- 删除 `work.md` 中标题以 `2026-09-03 16:41:48` 开头的本节记录。

## 2026-09-03 16:44:40 +08:00 --- 投票没有从被提名者开始的顺时针权威顺序 --- 以单个座次顺序测试完成 RED-GREEN --- 修改提名领域/快照/规则测试与 `work.md`

### 发现什么问题
- 当前提名只保存双方和票表，任何玩家都可任意时刻投票；无法实现从被提名者开始逐席 3 秒的面杀节奏。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：座次 P1→P2→P3→P4、P2 提名 P4 时，权威投票顺序必须为 P4→P1→P2→P3，并从索引 0 开始；死亡 P3 仍保留席位。
- RED 阶段确认 `VoterOrder/CurrentVoterIndex` 不存在，测试按预期编译失败。
- GREEN 阶段在活动提名中持久化完整顺时针玩家 ID 顺序和当前索引；从被提名者的权威座次开始循环，快照克隆会深拷贝顺序。
- 定向 Go 测试已通过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/game/game.go`。
- `packages/backend/internal/gameplay/day.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestNominationBuildsClockwiseVoterOrderFromNominee`。
- 从 `game.Nomination` 删除顺序/索引，删除 `clockwiseVoterOrderLocked` 与提名初始化/快照克隆接线。
- 删除 `work.md` 中标题以 `2026-09-03 16:44:40` 开头的本节记录。

## 2026-09-03 16:46:18 +08:00 --- 玩家仍可越过顺时针当前席抢先投票 --- 以单个失败测试驱动服务端轮次门禁完成 RED→GREEN --- 修改投票实现、规则测试与 `work.md`

### 发现什么问题
- 虽然提名已记录投票顺序，`applyCastVote` 尚未读取当前索引，非当前席仍可写入票表。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：当前应由 P2 投票时，P1 的赞成票必须被拒绝，票表和当前索引保持不变。
- 已运行定向测试并确认 RED：P1 的越序投票被旧实现接纳。
- 在写票、校验管家和消耗幽灵票之前校验当前索引与当前玩家；投票序列完成后也拒绝新增票，保证失败请求不改变任何状态。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestOnlyCurrentClockwiseSeatCanVote`。
- 从 `applyCastVote` 删除当前索引范围和 `currentVoterID` 校验。
- 删除 `work.md` 中标题以 `2026-09-03 16:46:18` 开头的本节记录。

## 2026-09-03 16:49:33 +08:00 --- 当前席投票后轮次不会推进 --- 以单个失败测试驱动索引推进完成 RED→GREEN --- 修改投票实现、规则测试与 `work.md`

### 发现什么问题
- 服务端已能拒绝越序投票，但合法当前席写票后 `CurrentVoterIndex` 仍停在原位，下一席永远无法获得投票权。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：P2 作为当前席投赞成后，票必须落盘且索引从 0 推进到 1。
- 已运行定向测试并确认 RED：合法票写入后当前索引仍为 0。
- 在合法票完成全部校验并写入后将当前索引推进一席；任何失败请求仍不会推进。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestCurrentClockwiseVoteAdvancesToNextSeat`。
- 从 `applyCastVote` 删除写票后的 `CurrentVoterIndex` 自增。
- 删除 `work.md` 中标题以 `2026-09-03 16:49:33` 开头的本节记录。

## 2026-09-03 16:51:58 +08:00 --- 说书人可以提前结算未完成的票圈 --- 以单个失败测试驱动完整性门禁完成 RED→GREEN --- 修改投票实现、规则测试与 `work.md`

### 发现什么问题
- 当前仍有玩家未表态时，`RESOLVE_NOMINATION` 可直接把不完整票表写成公开结果，既绕过顺时针流程，也无法保证“无响应视为否”。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：三席票圈只完成首席后，提前结算必须失败，阶段、提名和历史保持原样。
- 已运行定向测试并确认 RED：只完成第一席时旧实现仍接受结算。
- 结算前要求投票顺序非空且当前索引已走完全部席位；拒绝时不写公开历史、不切换阶段。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestStorytellerCannotResolveBeforeEverySeatHasARecordedDecision`。
- 从 `applyResolveNomination` 删除票圈顺序和索引完整性校验。
- 删除 `work.md` 中标题以 `2026-09-03 16:51:58` 开头的本节记录。

## 2026-09-03 16:53:23 +08:00 --- 当前席无响应时票圈无法继续 --- 以单个失败测试驱动说书人代记命令完成 RED→GREEN --- 修改领域命令、投票实现、规则测试与 `work.md`

### 发现什么问题
- 玩家断线或三秒内无响应时，没有权威方式为当前席记录默认“否”，顺时针票圈会永久卡住。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：说书人为当前死亡席代记“否”后，票表写入 false、索引推进且不消耗幽灵票。
- 已运行定向测试并确认 RED：`RecordVoteCmd` 尚不存在，测试按预期编译失败。
- 新增只允许当前说书人调用的 `RecordVoteCmd`；它以目标玩家身份复用唯一的投票写入路径，因此轮次、重复票、管家和幽灵票规则不会分叉。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `TestStorytellerCanRecordNoForTheCurrentUnresponsiveSeat`。
- 删除 `RecordVoteCmd` 类型、命令标记与 `Apply` 分派。
- 删除 `applyRecordVote`。
- 删除 `work.md` 中标题以 `2026-09-03 16:53:23` 开头的本节记录。

## 2026-09-03 16:54:59 +08:00 --- 顺时针轮次与说书人代记尚未穿透 WebSocket v2 --- 以单条真实多端失败测试驱动协议接线完成 RED→GREEN --- 修改协议、网关、生成契约、流程测试与 `work.md`

### 发现什么问题
- 领域层已有轮次与代记能力，但协议没有 `RECORD_VOTE`，客户端也看不到当前索引，真实多连接流程无法使用。
- 原投票流程测试按 P1、P2、P3 抢投，已不符合从被提名者 P2 开始的顺时针顺序。

### 使用什么方式解决
- 只改造一条真实 WebSocket 流程测试：P2、P3 依次赞成，说书人为 P4、P5 代记反对，P1 最后赞成；并断言代记后票表与当前索引公开同步。
- 已运行定向测试并确认 RED：`MsgRecordVote` 尚不存在，测试按预期编译失败。
- 协议新增 `RECORD_VOTE`，复用 `targetPlayerId + decision` 表达说书人代记；活动提名公开 `voterOrder/currentVoterIndex`。
- WebSocket v2 将新命令纳入鉴权、序号、持久化和广播的统一命令管线，并校验必填参数。
- 重新生成 Go/TypeScript 契约，避免手写协议漂移。

### 修改了哪些文件
- `packages/backend/internal/ws/voting_flow_test.go`。
- `proto/game.proto`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 将 `TestProtocolV2StorytellerFinalizesTheUniqueExecutionCandidate` 的投票段恢复为 P1/P2/P3 直接投票。
- 从 proto 删除 `RECORD_VOTE` 与提名的 `voter_order/current_voter_index` 后重新生成契约。
- 从 Hub 的命令白名单与转换分支删除 `MsgRecordVote`。
- 删除 `work.md` 中标题以 `2026-09-03 16:54:59` 开头的本节记录。

## 2026-09-03 16:56:50 +08:00 --- Core 客户端无法发送说书人代记票 --- 以单个失败测试驱动受序号保护的方法完成 RED→GREEN --- 修改 WebSocket 客户端与测试、`work.md`

### 发现什么问题
- 服务端协议已支持 `RECORD_VOTE`，但共享 Core 客户端没有对应方法，前端无法通过受序号保护的命令管线调用。

### 使用什么方式解决
- 新增一个且仅一个 Core WebSocket 测试：`recordVote('p4', false)` 必须发送含玩家、目标、决定和客户端序号的 `RECORD_VOTE`。
- 已运行定向测试并确认 RED：运行时明确报告 `client.recordVote is not a function`。
- 新增 `recordVote(targetPlayerId, decision)`，通过现有 `sendSequenced` 统一携带身份、恢复凭据、客户端序号并参与重放保护。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `lets the storyteller record a decision for the current voter`。
- 删除 `GameWebSocketClient.recordVote`。
- 删除 `work.md` 中标题以 `2026-09-03 16:56:50` 开头的本节记录。

## 2026-09-03 16:58:03 +08:00 --- 前端会话仓库无法调用说书人代记票 --- 以单个失败测试驱动仓库动作完成 RED→GREEN --- 修改会话仓库与测试、`work.md`

### 发现什么问题
- Core 客户端方法已经可用，但 Zustand 会话仓库没有动作和独立 pending 状态，页面无法调用或防止重复提交。

### 使用什么方式解决
- 新增一个且仅一个仓库测试：`recordVote('p4', false)` 必须原样转发给客户端，并将 pending 标记为 `record-vote`。
- 已运行定向测试并确认 RED：运行时明确报告 `recordVote is not a function`。
- 将 `recordVote` 纳入客户端能力、公开仓库动作和 `record-vote` pending 类型，通过现有统一 `send` 包装转发并复用错误恢复行为。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards a storyteller proxy decision as its own pending command`。
- 从测试 FakeClient、SessionClient、RoomSessionState、PendingCommand 和仓库实现删除 `recordVote` 接线。
- 删除 `work.md` 中标题以 `2026-09-03 16:58:03` 开头的本节记录。

## 2026-09-03 16:59:03 +08:00 --- 投票页不能从权威索引识别唯一当前席 --- 以单个失败测试驱动纯函数完成 RED→GREEN --- 修改投票页工具与测试、`work.md`

### 发现什么问题
- 协议已下发票圈顺序和当前索引，但页面仍仅按“自己是否投过”启用按钮，无法保证只有当前席操作。

### 使用什么方式解决
- 新增一个且仅一个纯函数测试：有效索引返回对应玩家，索引走出票圈后返回 `undefined`。
- 已运行定向测试并确认 RED：运行时明确报告 `currentVoterId is not a function`。
- 新增只读取权威 `voterOrder/currentVoterIndex` 的纯函数；非法或完成后的索引安全返回 `undefined`。

### 修改了哪些文件
- `packages/frontend/src/pages/game-play/utils.test.ts`。
- `packages/frontend/src/pages/game-play/utils.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `clockwise ballot` 测试组和 `currentVoterId` 导入。
- 删除 `currentVoterId` 实现。
- 删除 `work.md` 中标题以 `2026-09-03 16:59:03` 开头的本节记录。

## 2026-09-03 17:00:12 +08:00 --- 游戏页仍以全员抢投呈现新票圈 --- 接入权威当前席、代记否与完整性门禁 --- 修改游戏页、样式与 `work.md`

### 发现什么问题
- 页面未消费 `voterOrder/currentVoterIndex`，所有未投玩家仍同时看到可用按钮，说书人也可提前结算。
- 死者投反对的提示和门禁错误：幽灵票只应在赞成时消耗，幽灵票用尽仍必须能在轮到时记录反对。

### 使用什么方式解决
- 以已转绿的 `currentVoterId` 纯函数为页面门禁，只允许权威当前席操作；赞成额外检查存活或仍有幽灵票，反对始终允许当前席提交。
- 逐席显示已投、当前、等待状态和进度；说书人可为当前无响应席代记反对，只有全部席位完成后才可结算。
- 修正文案，明确死者仅投赞成才消耗幽灵票。

### 修改了哪些文件
- `packages/frontend/src/pages/game-play/index.tsx`。
- `packages/frontend/src/pages/game-play/index.css`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复页面原 `canVote` 条件、按票表渲染和无条件结算按钮，移除 `recordVote/currentVoterId` 接线。
- 删除新增 `.voteTurn`、`.voteDecisionWaiting`、`.voteRecordCurrent` 样式。
- 删除 `work.md` 中标题以 `2026-09-03 17:00:12` 开头的本节记录。

## 2026-09-03 17:01:23 +08:00 --- 提名协议新增必填字段后旧前端夹具无法类型检查 --- 补齐权威票圈字段 --- 修改房间体验测试夹具与 `work.md`

### 发现什么问题
- TypeScript 全量构建发现旧 `RoomNomination` 测试夹具缺少新必填的 `voterOrder/currentVoterIndex`。

### 使用什么方式解决
- 为该替换语义测试补入两席顺序与初始索引，不改变测试原本要验证的“完整投影替换可选状态”行为。

### 修改了哪些文件
- `packages/frontend/src/lib/room-experience.test.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 从该测试的 nomination 夹具删除 `voterOrder/currentVoterIndex`。
- 删除 `work.md` 中标题以 `2026-09-03 17:01:23` 开头的本节记录。

## 2026-09-03 17:02:53 +08:00 --- 旧后端测试夹具绕过完整顺时针票圈 --- 迁移普通提名夹具到权威顺序模型 --- 修改 gameplay 测试与 `work.md`

### 发现什么问题
- 全量 gameplay 回归显示旧辅助函数只让任意一名玩家投票便结算，静态提名也没有顺序与完成索引；新完整性门禁正确拒绝了这些旧夹具。

### 使用什么方式解决
- `resolveNominationWithoutExecutionBy` 改为读取活动提名的权威当前席，逐席记录反对直到票圈完成，再请求结算。
- 死者幽灵票和多数票静态夹具补齐顺序、索引及明确的反对票，使测试表达完整票圈而非绕过生产约束。
- 管家规则测试暂不混入本次机械迁移，留作独立语义追踪子弹。

### 修改了哪些文件
- `packages/backend/internal/gameplay/game_session_test.go`。
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 将 `resolveNominationWithoutExecutionBy` 恢复为只用传入玩家投一张反对票。
- 从三处静态提名夹具删除顺序/索引，并移除多数票夹具的 P4/P5 反对票。
- 删除 `work.md` 中标题以 `2026-09-03 17:02:53` 开头的本节记录。

## 2026-09-03 17:04:02 +08:00 --- 管家服务端门禁与公开顺时针票圈冲突 --- 依据官方裁定以单个失败测试移除泄密门禁完成 RED→GREEN --- 修改投票实现、管家测试与 `work.md`

### 发现什么问题
- 现实现要求主人已经被计为赞成后才接纳管家赞成；当顺时针标记先经过管家时会永久卡住，也会通过错误暴露隐藏身份。
- 官方管家裁定明确说明座次先后不重要、说书人应照常计入误投，守规则责任在玩家本人；同桌程序不应在服务端拒绝并泄密。

### 使用什么方式解决
- 只改写一条管家测试：说书人先为被提名席代记反对使标记到达管家，主人的席位尚未经过时，当前管家的赞成必须照常计入并推进。
- 已运行定向测试并确认 RED：旧实现以“主人尚未赞成”拒绝当前管家，正好复现卡死与身份泄漏。
- 删除服务器对管家投票的隐藏角色校验；同桌玩家依照桌面手势自行遵守能力，服务端只执行公开轮次、存活/幽灵票和重复票规则。

### 修改了哪些文件
- `packages/backend/internal/gameplay/butler_test.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复原 `TestButlerCannotVoteYesBeforeMasterVotesYes` 测试内容。
- 恢复 `applyCastVote` 对 `validateButlerVoteLocked` 的调用与该校验函数。
- 删除 `work.md` 中标题以 `2026-09-03 17:04:02` 开头的本节记录。

## 2026-09-03 17:05:18 +08:00 --- 其余管家测试仍假定旧服务端隐藏角色门禁 --- 迁移到公开轮次与私密状态分离模型 --- 修改管家测试与 `work.md`

### 发现什么问题
- 首条官方语义测试转绿后，其余管家测试仍越过被提名席抢投，且有一条期待服务器以缺少主人为由公开拒绝投票。
- 快照测试把“主人映射是否持久化”与已不成立的乱序投票流程耦合。

### 使用什么方式解决
- 添加测试辅助方法，由说书人为 P6 代记反对后再让当前 P1 管家表态。
- 将缺少主人场景改为验证公开投票不泄露私密能力状态；快照场景直接验证主人映射恢复。

### 修改了哪些文件
- `packages/backend/internal/gameplay/butler_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `advanceVoteMarkerToButler` 及各调用，恢复缺少主人时应报错和快照后的旧投票断言。
- 删除 `work.md` 中标题以 `2026-09-03 17:05:18` 开头的本节记录。

## 2026-09-03 17:06:36 +08:00 --- 完整 WebSocket 对局测试仍从非当前席抢投 --- 按被提名席起始完成五席票圈并更新修订断言 --- 修改完整对局测试与 `work.md`

### 发现什么问题
- 后端全量测试只剩完整对局场景失败：P5 被提名后测试直接让 P1 投票，违反新权威顺序，也少了 P4/P5 的明确决定。

### 使用什么方式解决
- 让说书人先为当前 P5 代记反对，再由 P1、恢复后的 P2、P3 依次赞成，最后为 P4 代记反对，形成完整五席票圈。
- 两次新增提交使恢复点修订号和终局修订号分别从 25/29 更新为 26/31；仍验证断线恢复保持序号和票表。

### 修改了哪些文件
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复原 P1/P2/P3 三票流程、票表断言及修订号 25/29。
- 删除 `work.md` 中标题以 `2026-09-03 17:06:36` 开头的本节记录。

## 2026-09-03 17:08:13 +08:00 --- 玩家夜间选择提交后没有待说书人审核状态 --- 以单个失败测试驱动可持久化待审核状态完成 RED→GREEN --- 修改夜间领域、快照、测试与 `work.md`

### 发现什么问题
- 玩家提交当前角色的夜间选择会直接追加到混合行动列表，却没有可由说书人审核的唯一待处理状态；效果虽未生效，但无法形成可靠确认流程。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：投毒者本人选择目标后应生成 `pendingNightAction`，但唤醒索引、已确认行动和中毒状态都不得改变。
- 已运行定向测试并确认 RED：`GameSession` 不存在 `pendingNightAction`，测试按预期编译失败。
- 玩家合法提交时只写唯一 `pendingNightAction` 并广播“已提交”事件，不推进唤醒索引、不写确认行动、不应用中毒等效果；已有待审核选择时拒绝重复提交。
- 待审核选择深拷贝进入会话快照，开新夜晚或结算夜晚时清空，保证服务器重启后不会丢失或串夜。
- 测试使用第二夜，使投毒者确为首个活动步骤；首夜爪牙互认步骤应先于投毒，未通过篡改索引绕过。

### 修改了哪些文件
- `packages/backend/internal/gameplay/night_player_flow_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/information.go`。
- `packages/backend/internal/gameplay/night.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `night_player_flow_test.go`。
- 删除 `pendingNightAction` 会话/快照字段及其克隆、清理逻辑，恢复玩家提交直接追加 `nightActions` 的旧流程。
- 删除 `work.md` 中标题以 `2026-09-03 17:08:13` 开头的本节记录。

## 2026-09-03 17:10:21 +08:00 --- 说书人无法审核并确认玩家夜间选择 --- 以单个失败测试驱动可持久化确认状态完成 RED→GREEN --- 修改夜间领域、快照、测试与 `work.md`

### 发现什么问题
- 玩家选择已有待审核槽位，但说书人没有独立确认命令，也无法在确认时调整目标/结果；更没有“效果已确认但等待玩家阅知”的状态。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：说书人确认待处理投毒后，应清除 pending、写入 confirmed 与确认行动、应用中毒，但在玩家确认已阅前不推进唤醒索引。
- 已运行定向测试并确认 RED：确认命令和 `confirmedNightAction` 字段均不存在，测试按预期编译失败。
- 新增仅说书人可用的确认命令：复核当前步骤、可调整目标/结果、自动生成信息结果，随后才写入权威行动并应用投毒/管家等效果。
- 确认后的行动进入可持久化 `confirmedNightAction`，在玩家已阅前不推进步骤；这期间拒绝新提交或重复确认。
- 已确认的玩家选择与说书人代办统一视为权威行动，夜间保护/击杀/守鸦人判断不再依赖 `ActorID == storyteller`。

### 修改了哪些文件
- `packages/backend/internal/gameplay/night_player_flow_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/night.go`。
- `packages/backend/internal/gameplay/information.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `TestStorytellerConfirmationAppliesPendingChoiceButWaitsForPlayerAcknowledgement`。
- 删除确认命令、`confirmedNightAction` 会话/快照字段与夜间结果/效果辅助函数，恢复只处理说书人 ActorID 的旧结算筛选。
- 删除 `work.md` 中标题以 `2026-09-03 17:10:21` 开头的本节记录。

## 2026-09-03 17:14:16 +08:00 --- 玩家无法确认已阅并推进夜间步骤 --- 以单个失败测试驱动可持久化已阅推进完成 RED→GREEN --- 修改夜间领域、快照、测试与 `work.md`

### 发现什么问题
- 说书人确认后会正确停在当前步骤，但没有玩家“已阅”命令，夜晚仍无法从审核态继续。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：当前投毒者确认已阅后，唤醒索引推进、confirmed 清空，同时保留已确认行动和已应用效果。
- 已运行定向测试并确认 RED：`AcknowledgeNightActionCmd` 尚不存在，测试按预期编译失败。
- 新增玩家已阅命令，只允许当前步骤匹配的存活角色确认；每名角色只能确认一次。
- 单人步骤确认后立即推进并清理 confirmed；阵营级多人步骤会等待所有匹配玩家确认，已阅集合随快照持久化并在换夜/结算时清理。

### 修改了哪些文件
- `packages/backend/internal/gameplay/night_player_flow_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/night.go`。
- `packages/backend/internal/gameplay/information.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `TestActingPlayerAcknowledgementAdvancesToTheNextNightStep`。
- 删除已阅命令、`nightAcknowledged` 会话/快照字段和 `applyAcknowledgeNightAction`。
- 删除 `work.md` 中标题以 `2026-09-03 17:14:16` 开头的本节记录。

## 2026-09-03 17:16:00 +08:00 --- 当前夜间角色拿不到私密步骤且待审核选择无法投影 --- 以单个失败测试驱动接收者投影完成 RED→GREEN --- 修改夜间投影、测试与 `work.md`

### 发现什么问题
- 领域状态已存在，但 `ProjectionFor` 仍只给说书人夜间步骤；当前角色手机无法知道自己被唤醒，旁观隐私边界也没有测试保护。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 投影测试：当前投毒者收到 `awaiting_player` 与步骤，旁观者完全看不到；提交后投毒者收到 `awaiting_storyteller` 与自己的待审核选择，旁观者仍为空。
- 已运行定向测试并确认 RED：夜间状态常量及投影字段均不存在，测试按预期编译失败。
- 投影新增三态 `awaiting_player/awaiting_storyteller/awaiting_acknowledgement` 以及当前 pending/confirmed 行动。
- 说书人获得完整步骤与审核数据；仅匹配当前步骤的存活角色获得当前步骤，只有提交者看自己的 pending，确认结果可给当前相关角色；旁观者仍为空。

### 修改了哪些文件
- `packages/backend/internal/gameplay/night_player_flow_test.go`。
- `packages/backend/internal/gameplay/projection.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `TestOnlyTheActingPlayerReceivesThePrivateNightTurnProjection`。
- 删除夜间投影三态常量、字段和按接收者分支。
- 删除 `work.md` 中标题以 `2026-09-03 17:16:00` 开头的本节记录。

## 2026-09-03 17:17:46 +08:00 --- 玩家确认前断线会卡住整夜 --- 以单个失败测试驱动说书人跳过兜底完成 RED→GREEN --- 修改夜间领域、测试与 `work.md`

### 发现什么问题
- 当前角色若在说书人确认结果后断线，所有匹配角色无法完成已阅，唤醒索引永久停留。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：说书人跳过已确认步骤的玩家已阅后，应推进并清理临时态，但保留已确认行动与已生效的投毒结果。
- 已运行定向测试并确认 RED：`SkipNightActionCmd` 尚不存在，测试按预期编译失败。
- 新增仅说书人可用的跳过命令：无论当前处于等待玩家、待审核或待已阅，都推进一席并清理临时选择/确认/已阅集合；已确认并已生效的行动保留。

### 修改了哪些文件
- `packages/backend/internal/gameplay/night_player_flow_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/night.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `TestStorytellerCanSkipAConfirmedStepWhenTheActorCannotAcknowledge`。
- 删除 `SkipNightActionCmd`、命令分派和 `applySkipNightAction`。
- 删除 `work.md` 中标题以 `2026-09-03 17:17:46` 开头的本节记录。

## 2026-09-03 17:19:20 +08:00 --- 间谍旧投影测试禁止其看到自己正在参与的爪牙步骤 --- 区分当前私密步骤与说书人管理数据 --- 修改 gameplay 投影测试与 `work.md`

### 发现什么问题
- gameplay 全量回归只剩间谍投影断言失败：首夜当前正是爪牙互认，间谍作为存活爪牙现在应收到自己的当前步骤；旧测试把它误当成泄漏说书人面板。

### 使用什么方式解决
- 改为要求间谍只看到当前 `learn_demon` 私密步骤，同时仍看不到完整唤醒列表、夜间行动历史与中毒状态。

### 修改了哪些文件
- `packages/backend/internal/gameplay/game_session_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复“CurrentNightWakeStep 必须为空”的旧断言。
- 删除 `work.md` 中标题以 `2026-09-03 17:19:20` 开头的本节记录。

## 2026-09-03 17:20:06 +08:00 --- 夜间玩家状态机尚未穿透 WebSocket v2 --- 以单个真实六端失败测试驱动协议接线完成 RED→GREEN --- 修改协议、生成器、网关、投影适配、测试与 `work.md`

### 发现什么问题
- 领域层已有提交、审核、已阅和隐私投影，但协议没有确认/已阅/跳过命令，也没有投影字段，微信/H5 客户端无法参与。

### 使用什么方式解决
- 新增一个且仅一个真实 WebSocket 流程测试：五名玩家与说书人完成开局，当前爪牙独享首夜步骤并提交，说书人看到待审核且确认，只有行动者收到结果，行动者已阅后推进至恶魔步骤。
- 同时断言旁观者在选择前后均看不到步骤或待审核数据。
- 已运行定向测试并确认 RED：RoomState 字段和确认/已阅消息常量均不存在，测试按预期编译失败。
- 协议新增确认、已阅、跳过三条命令，以及夜间状态、pending、confirmed 三个接收者特定投影字段。
- Hub 将新命令纳入统一鉴权/序号/事务管线，SessionEngine 透传领域隐私投影；生成器为两个 NightAction 指针补显式 Go 类型并重新生成双端契约。
- 生成后的 Go RoomState 直接复用强类型 `game.NightActionType`，测试按该枚举比较而非错误转换为 string。

### 修改了哪些文件
- `packages/backend/internal/ws/night_player_flow_test.go`。
- `proto/game.proto`。
- `scripts/generate-protocol-contracts.mjs`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/session_engine.go`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `packages/backend/internal/ws/night_player_flow_test.go`。
- 从 proto 删除三条夜间命令和三个 RoomState 字段，删除生成器类型覆盖后重新生成。
- 从 Hub 与 SessionEngine 删除相应命令/投影接线。
- 删除 `work.md` 中标题以 `2026-09-03 17:20:06` 开头的本节记录。

## 2026-09-03 17:22:16 +08:00 --- Core 客户端缺少夜间审核生命周期方法 --- 以单个失败测试驱动三条有序方法完成 RED→GREEN --- 修改 Core WebSocket 客户端、测试与 `work.md`

### 发现什么问题
- 服务端协议纵切已完成，但共享客户端无法发送说书人确认、玩家已阅或说书人跳过命令。

### 使用什么方式解决
- 新增一个且仅一个客户端测试：确认（含调整目标/结果）、已阅、跳过必须依次经序号 1/2/3 发出，并等待前一命令确认后再发送下一条。
- 已运行定向测试并确认 RED：运行时明确报告 `confirmNightAction is not a function`。
- 新增确认、已阅、跳过三个方法，全部通过 `sendSequenced` 排队；确认携带说书人最终目标和去空白结果。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `sends the night review lifecycle through sequenced commands`。
- 删除三个 GameWebSocketClient 夜间生命周期方法。
- 删除 `work.md` 中标题以 `2026-09-03 17:22:16` 开头的本节记录。

## 2026-09-03 17:23:16 +08:00 --- Zustand 会话仓库缺少夜间审核生命周期动作 --- 以单个失败测试驱动三动作转发完成 RED→GREEN --- 修改会话仓库、测试与 `work.md`

### 发现什么问题
- Core 已支持确认、已阅和跳过，但页面使用的会话仓库尚无对应动作及 pending 分类。

### 使用什么方式解决
- 新增一个且仅一个仓库测试：三个动作必须按参数原样转发，最后一条跳过将 pending 标记为 `skip-night-action`。
- 已运行定向测试并确认 RED：运行时明确报告 `confirmNightAction is not a function`。
- 将确认、已阅、跳过加入 SessionClient/RoomSessionState 和三个独立 pending 类型，统一通过现有 `send` 错误边界转发。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards the three night review lifecycle actions`。
- 从 FakeClient、能力 Pick、仓库接口/实现和 PendingCommand 删除三动作接线。
- 删除 `work.md` 中标题以 `2026-09-03 17:23:16` 开头的本节记录。

## 2026-09-03 17:24:55 +08:00 --- 夜间玩家与说书人仍停留在旧单机表单 --- 接入三阶段私密交互、断线兜底与横屏面板 --- 修改游戏页、样式与 `work.md`

### 发现什么问题
- 非说书人夜晚仍只显示闭眼等待，无法使用已经贯通的私密步骤；说书人表单也不会区分等待玩家、待复核、待玩家已阅。

### 使用什么方式解决
- 当前角色手机按权威状态显示：私密选人/提交、等待说书人、裁定结果/已阅；阵营级步骤支持一人提交后相关角色共同等待与分别已阅。
- 说书人面板自动带入玩家选择，可调整目标和最终结果后确认；无玩家提交时仍可代办，任一阶段均可在二次确认后跳过断线步骤。
- 已确认效果在跳过“已阅”时保留；玩家提交不允许填写裁定结果。
- 新增夜间高对比私密视觉，并在宽屏横向设备将主流程与说书人侧栏改为双栏、侧栏吸顶。

### 修改了哪些文件
- `packages/frontend/src/pages/game-play/index.tsx`。
- `packages/frontend/src/pages/game-play/index.css`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复夜间仅说书人直接提交、非说书人统一等待的两段旧 JSX，并移除三个仓库动作选择器/审核辅助函数。
- 删除新增夜间审核、玩家唤醒和横屏媒体查询样式。
- 删除 `work.md` 中标题以 `2026-09-03 17:24:55` 开头的本节记录。

## 2026-09-03 17:27:26 +08:00 --- 完整对局旧隐私断言禁止当前角色看自己的夜间步骤 --- 收紧为角色匹配的私密投影断言 --- 修改 WebSocket 完整对局测试与 `work.md`

### 发现什么问题
- 后端全量回归只剩旧断言失败：它要求所有玩家的 `CurrentNightWakeStep` 永远为空，与新“仅当前角色手机亮起”的目标冲突。

### 使用什么方式解决
- 继续严格禁止完整唤醒列表、行动历史、pending/confirmed 和中毒状态泄漏。
- 若玩家收到单个当前步骤，则验证其真实角色或阵营类型确实匹配，并且状态只能是 `awaiting_player`；无步骤时也不得单独泄漏状态。

### 修改了哪些文件
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复玩家 `CurrentNightWakeStep` 必须为空的旧断言并删除 gameplay 导入。
- 删除 `work.md` 中标题以 `2026-09-03 17:27:26` 开头的本节记录。

## 2026-09-03 17:29:17 +08:00 --- 随机配包没有生成小恶魔三张伪装身份 --- 以单个失败测试驱动配置与验证完成 RED→GREEN --- 修改剧本核心、测试与 `work.md`

### 发现什么问题
- 暗流涌动首夜小恶魔必须得知三张不在场善良身份，当前 `ScriptSetup` 和随机配包完全没有该数据。

### 使用什么方式解决
- 新增一个且仅一个 Core 测试：随机配置必须包含恰好三张、不重复、不在实际配置中且阵营为善良的伪装身份。
- 已运行定向测试并确认 RED：随机配置的 `demonBluffCharacterIds` 为 undefined。
- `ScriptSetup` 新增三张伪装身份；随机器从未实际在场、未作为酒鬼展示且阵营为善良的角色中无重复抽取三张。
- 配置验证新增 `INVALID_DEMON_BLUFFS`，拒绝数量、重复、在场、酒鬼展示占用、邪恶或未知角色；补齐原手写合法夹具。

### 修改了哪些文件
- `packages/core/src/scripts/__tests__/scripts.test.ts`。
- `packages/core/src/scripts/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `generates three unique out-of-play good characters as Demon bluffs`。
- 从 ScriptSetup、验证和随机器删除伪装身份字段/规则，恢复两个手写夹具。
- 删除 `work.md` 中标题以 `2026-09-03 17:29:17` 开头的本节记录。

## 2026-09-03 17:31:13 +08:00 --- 后端首夜恶魔信息仍只有爪牙名单 --- 以单个失败测试驱动持久化伪装信息完成 RED→GREEN --- 修改配包领域、快照、信息计算、测试与 `work.md`

### 发现什么问题
- 即使前端配置将生成伪装身份，后端会话尚未保存，首夜恶魔信息仍只返回爪牙。

### 使用什么方式解决
- 只改写一条已有首夜测试：恶魔必须同时得到爪牙名单与确定的三张不在场善良身份（厨师、共情者、占卜师）。
- 已运行定向测试并确认 RED：旧结果只有 `Minions: P4 (Poisoner)`。
- 角色发放命令与会话快照新增三张伪装身份；显式输入严格校验为不重复、不在场/不占用酒鬼展示且为善良角色。
- 为兼容旧保存与测试，未显式传入时按剧本顺序确定性补足三张合法伪装；首夜恶魔信息同时输出爪牙和伪装身份。

### 修改了哪些文件
- `packages/backend/internal/gameplay/evil_team_information_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `packages/backend/internal/gameplay/information.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 将首夜恶魔期望恢复为仅 `Minions: P4 (Poisoner)`。
- 删除发放命令/会话/快照的伪装字段、`resolveDemonBluffCharacterIDs` 和首夜结果拼接。
- 删除 `work.md` 中标题以 `2026-09-03 17:31:13` 开头的本节记录。

## 2026-09-03 17:32:53 +08:00 --- 小恶魔伪装身份尚未穿透 WebSocket 发放与投影 --- 以单个完整对局失败测试驱动协议接线完成 RED→GREEN --- 修改协议、领域投影、网关、生成契约、测试与 `work.md`

### 发现什么问题
- 领域层能保存并生成伪装信息，但协议发放消息没有字段，说书人也无法在最终配置投影核对。

### 使用什么方式解决
- 只改造完整 WebSocket 对局测试：发放明确的厨师/共情者/占卜师三张伪装，断言说书人投影保留，首夜恶魔结果完整包含它们。
- 已运行定向测试并确认 RED：ClientMessage 不接受伪装字段，RoomState 也无对应投影。
- 协议在角色发放和 RoomState 中新增三张伪装身份；Hub 传入领域命令，Projection 只向说书人公开原始配置，SessionEngine 透传并重新生成双端契约。

### 修改了哪些文件
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `proto/game.proto`。
- `packages/backend/internal/gameplay/projection.go`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/session_engine.go`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复发放命令不带伪装、删除说书人投影断言，并恢复旧恶魔结果字符串和 slices 导入。
- 删除协议字段、Hub/Projection/SessionEngine 接线并重新生成契约。
- 删除 `work.md` 中标题以 `2026-09-03 17:32:53` 开头的本节记录。

## 2026-09-03 17:34:44 +08:00 --- Core 角色发放方法无法携带三张恶魔伪装 --- 以单个失败测试驱动出站字段完成 RED→GREEN --- 修改 Core WebSocket 客户端、测试与 `work.md`

### 发现什么问题
- 协议已支持伪装字段，但 Core `assignCharacters` 仍只有角色、酒鬼展示和红鲱鱼三个配置参数。

### 使用什么方式解决
- 新增一个且仅一个客户端测试：第四个参数的三张伪装必须原序写入 `ASSIGN_CHARACTERS.demonBluffCharacterIds`。
- 已运行定向测试并确认 RED：ASSIGN_CHARACTERS 消息缺少 `demonBluffCharacterIds`。
- `assignCharacters` 新增可选第四参数，非空时原序写入受序号保护的发放消息。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `sends the three configured Demon bluffs with character assignment`。
- 删除 `assignCharacters` 第四参数及消息字段。
- 删除 `work.md` 中标题以 `2026-09-03 17:34:44` 开头的本节记录。

## 2026-09-03 17:35:44 +08:00 --- Zustand 角色发放动作会丢弃恶魔伪装 --- 以单个失败测试驱动第四参数转发完成 RED→GREEN --- 修改会话仓库、测试与 `work.md`

### 发现什么问题
- Core 发放方法已扩展，但前端仓库接口仍只有三个参数，设置页的三张伪装无法送达服务端。

### 使用什么方式解决
- 新增一个且仅一个仓库测试：第四参数三张伪装必须与角色、酒鬼展示、红鲱鱼一起原样转发。
- 已运行定向测试并确认 RED：FakeClient 只收到前三个发放参数，伪装数组被丢弃。
- 扩展 RoomSessionState 签名和实现，将第四参数原样交给 Core 客户端。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards Demon bluffs with the character setup`。
- 删除仓库发放动作第四参数及转发。
- 删除 `work.md` 中标题以 `2026-09-03 17:35:44` 开头的本节记录。

## 2026-09-03 17:37:14 +08:00 --- 说书人换角后原恶魔伪装可能变成在场或重复 --- 以单个失败测试驱动规范化工具完成 RED→GREEN --- 修改设置工具、测试与 `work.md`

### 发现什么问题
- 随机配置初始合法，但说书人逐席换角或改变酒鬼展示后，原三张伪装可能不再可用；直接保留会导致发放被服务端拒绝。

### 使用什么方式解决
- 新增一个且仅一个纯函数测试：保留仍合法的现有选择，去掉在场/重复项，再按候选顺序补足三张。
- 已运行定向测试并确认 RED：运行时明确报告 `normalizeDemonBluffs is not a function`。
- 新增纯函数：按原顺序保留候选集中且不重复的当前选择，再按候选顺序补到最多三张。

### 修改了哪些文件
- `packages/frontend/src/pages/game-setup/utils.test.ts`。
- `packages/frontend/src/pages/game-setup/utils.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除规范化测试和导入。
- 删除 `normalizeDemonBluffs`。
- 删除 `work.md` 中标题以 `2026-09-03 17:37:14` 开头的本节记录。

## 2026-09-03 17:38:20 +08:00 --- 设置页无法核对或调整小恶魔三张伪装 --- 接入合法候选、冲突修复、发放与首夜中文展示 --- 修改设置页、结果显示与 `work.md`

### 发现什么问题
- 随机配置虽已有伪装数据，但页面不显示、不允许调整，发放动作也未传入；换角后可能保留冲突选择。
- 首夜英文结果中的 `Bluffs` 段没有中文本地化。

### 使用什么方式解决
- 设置页列出所有未在场、未作为酒鬼展示的善良身份，允许说书人保持三选；换角/酒鬼展示时用已测试纯函数保留合法项并补足。
- 发放时携带三张伪装，身份锁定后的最终配置仅向说书人展示以便核对。
- 私密首夜结果把 `Bluffs` 本地化为“伪装身份”。

### 修改了哪些文件
- `packages/frontend/src/pages/game-setup/index.tsx`。
- `packages/frontend/src/lib/character-display.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 eligible/toggle/normalize 逻辑、两处伪装身份 UI 和发放第四参数。
- 恢复 `nightResultDisplay` 不处理 Bluffs 段。
- 删除 `work.md` 中标题以 `2026-09-03 17:38:20` 开头的本节记录。

## 2026-09-03 17:39:18 +08:00 --- 公共投影泄露死亡原因和击杀者 --- 以单个失败测试驱动服务端脱敏完成 RED→GREEN --- 修改投影实现、契约测试与 `work.md`

### 发现什么问题
- 当前 `Deaths` 原样发给每个玩家，夜杀、能力致死和 `killedBy` 会泄露说书人私密裁定，违反已确认的黎明只公布姓名规则。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 投影测试：普通玩家只能得到死亡玩家和天数，死因/击杀者为空；说书人仍得到完整记录。
- 已运行定向测试并确认 RED：玩家收到完整 `night_kill` 与击杀者 P1。
- 接收者能力新增 `seeDeathCauses`，仅说书人/内部可信投影可见；普通玩家的死亡记录保留玩家与天数，但清空死因和击杀者。

### 修改了哪些文件
- `packages/backend/internal/gameplay/projection_contract_test.go`。
- `packages/backend/internal/gameplay/projection.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `TestPlayerProjectionHidesPrivateDeathCauseAndKiller`。
- 删除死亡记录副本脱敏与 `seeDeathCauses` 能力。
- 删除 `work.md` 中标题以 `2026-09-03 17:39:18` 开头的本节记录。

## 2026-09-03 17:40:48 +08:00 --- 页面仍把脱敏后的空死因当作公开标签 --- 统一玩家端只显示死亡日，说书人保留原因 --- 修改进行中页、结果页与 `work.md`

### 发现什么问题
- 服务端已脱敏，但进行中与结果页仍按死因字典渲染，空值会产生残缺文案，且旧代码语义仍鼓励公开死因。

### 使用什么方式解决
- 普通玩家的状态行与时间线只显示“第 N 天死亡”；进行中说书人视图额外显示私密死因。
- 结果页的公开身份卡与死亡时间线同样只保留死亡日，移除死因字典。

### 修改了哪些文件
- `packages/frontend/src/pages/game-play/index.tsx`。
- `packages/frontend/src/pages/game-over/index.tsx`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复两页按 `DEATH_CAUSE_LABELS` 公开渲染的旧文案和结果页字典。
- 删除 `work.md` 中标题以 `2026-09-03 17:40:48` 开头的本节记录。

## 2026-09-03 17:41:23 +08:00 --- 游戏结束瞬间自动向全员揭示身份 --- 以单个失败测试驱动说书人揭幕门禁完成 RED→GREEN --- 修改终局领域、快照、投影、测试与 `work.md`

### 发现什么问题
- `Finished` 阶段自动启用全角色可见，绕过已确认的说书人揭幕控制，胜负产生与公开魔典无法分开。

### 使用什么方式解决
- 只改写一条终局投影测试：结束后普通玩家仍看不到他人身份；说书人执行 `PublishGrimoireCmd` 后才向其揭示小恶魔。
- 已运行定向测试并确认 RED：`PublishGrimoireCmd` 尚不存在（旧投影同时会提前揭示）。
- 新增仅说书人可在已结束且有胜负时执行的幂等揭幕命令；`grimoireRevealed` 随快照持久化并进入投影。
- 普通玩家只有在揭幕标记为真后才获得全角色；说书人始终保留其私密全局视图。

### 修改了哪些文件
- `packages/backend/internal/gameplay/game_session_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/endgame.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `packages/backend/internal/gameplay/projection.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复 `TestGameSessionFinishedStateRevealsCharactersToPlayers` 的结束即揭示断言。
- 删除揭幕命令、状态/快照/投影字段，并恢复 Finished 自动 `seeAllCharacters`。
- 删除 `work.md` 中标题以 `2026-09-03 17:41:23` 开头的本节记录。

## 2026-09-03 17:43:23 +08:00 --- 说书人揭幕门禁尚未穿透真实 WebSocket 对局 --- 以单个完整对局失败测试驱动协议接线完成 RED→GREEN --- 修改协议、网关、投影适配、生成契约、测试与 `work.md`

### 发现什么问题
- 领域门禁已完成，但协议无发布命令/标记；原完整对局还假定结束广播自动带全角色与公开死因。

### 使用什么方式解决
- 改造唯一完整对局测试：终局时说书人仍看全局、每位玩家只看自己且 `grimoireRevealed=false`；发布命令提交后全员才看到全部角色，修订号推进到 32。
- 同时让公共终局断言只检查死亡玩家与天数，不再要求已被脱敏的死因。
- 已运行定向测试并确认 RED：发布命令常量与 RoomState 揭幕标记均不存在。
- 协议新增 `PUBLISH_GRIMOIRE` 与必填 `grimoire_revealed`；Hub 进入统一命令事务，SessionEngine 透传领域标记并重新生成双端契约。

### 修改了哪些文件
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `proto/game.proto`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/session_engine.go`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复终局立即对所有广播调用全角色断言、修订号 31 和公开处决死因断言，删除发布步骤。
- 删除协议发布命令/标记、Hub/SessionEngine 接线并重新生成契约。
- 删除 `work.md` 中标题以 `2026-09-03 17:43:23` 开头的本节记录。

## 2026-09-03 17:44:51 +08:00 --- Core 客户端不能发送发布魔典命令 --- 以单个失败测试驱动受序号保护的方法完成 RED→GREEN --- 修改 Core WebSocket 客户端、测试与 `work.md`

### 发现什么问题
- 服务端发布命令已贯通，但共享客户端没有方法，结果页无法调用。

### 使用什么方式解决
- 新增一个且仅一个客户端测试：`publishGrimoire()` 必须发送携带身份与客户端序号的 `PUBLISH_GRIMOIRE`。
- 已运行定向测试并确认 RED：运行时明确报告 `publishGrimoire is not a function`。
- 新增 `publishGrimoire()` 并通过统一 `sendSequenced` 管线发送。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `publishes the grimoire through a sequenced command`。
- 删除 `GameWebSocketClient.publishGrimoire`。
- 删除 `work.md` 中标题以 `2026-09-03 17:44:51` 开头的本节记录。

## 2026-09-03 17:45:48 +08:00 --- Zustand 仓库无法触发终局揭幕 --- 以单个失败测试驱动动作与 pending 完成 RED→GREEN --- 修改会话仓库、测试与 `work.md`

### 发现什么问题
- 结果页只能访问会话仓库，Core 方法不能被页面直接安全调用。

### 使用什么方式解决
- 新增一个且仅一个仓库测试：发布动作必须调用客户端并设置独立 `publish-grimoire` pending 状态。
- 已运行定向测试并确认 RED：运行时明确报告 `publishGrimoire is not a function`。
- 将客户端能力、公开仓库动作和 `publish-grimoire` pending 分类完整接线。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards grimoire publication as its own pending command`。
- 从 FakeClient、SessionClient、RoomSessionState、PendingCommand 与仓库实现删除发布接线。
- 删除 `work.md` 中标题以 `2026-09-03 17:45:48` 开头的本节记录。

## 2026-09-03 17:47:07 +08:00 --- 结果页没有揭幕控制且遗漏投票复盘 --- 接入说书人发布魔典与完整提名时间线 --- 修改结果页与 `work.md`

### 发现什么问题
- 后端已区分胜负与揭幕，但结果页仍直接渲染“身份缺失”卡片，也没有发布入口。
- 已持久化的 `nominationResults` 未在最终复盘展示。

### 使用什么方式解决
- 说书人结束后先看到私密魔典，并可通过不可逆二次确认向全员公开；普通玩家公开前只看到胜负、等待提示和公共时间线。
- 揭幕后统一显示身份卡；新增按天展示提名者、被提名者、赞成、反对和门槛的投票时间线。
- 顶部动作同时容纳发布魔典与关闭/离开。

### 修改了哪些文件
- `packages/frontend/src/pages/game-over/index.tsx`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复结果页固定标题与无条件身份列表，删除发布动作、等待视图和投票时间线。
- 删除 `work.md` 中标题以 `2026-09-03 17:47:07` 开头的本节记录。

## 2026-09-03 17:48:21 +08:00 --- 夜间死亡缺少黎明批量审核关口 --- 以单个失败测试驱动建议/编辑/确认状态机完成 RED→GREEN --- 修改夜间领域、快照、测试与 `work.md`

### 发现什么问题
- 当前结束夜晚会立即应用击杀并广播，不能先让说书人核对保护、士兵、镇长转移等结果，也不能确认零人或多人死亡。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：系统先建议 P1 死亡且不改变阶段/存活/死亡历史；说书人改为 P1、P2 两人并确认后，才作为同一黎明批次进入第 1 天。
- 已运行定向测试并确认 RED：准备/确认命令及待审核字段均不存在，测试按预期编译失败。
- 新增 `PrepareDawnCmd`：全部夜间步骤完成后，根据恶魔存活/中毒、僧侣保护、士兵与镇长规则生成零/多名建议，仅写私密待审状态。
- 新增 `ConfirmDawnCmd`：说书人可提交任意不重复的存活玩家批次，确认后原子写入死亡、推进天数和胜负；待审状态随快照持久化。
- 旧 `ResolveNightCmd` 保留为兼容快速路径，但一旦进入审核态便不能绕过 `CONFIRM_DAWN`。
- 测试为四名善良玩家补齐角色阵营，避免只有小恶魔被计入存活人数而误触“最终两人”胜负。

### 修改了哪些文件
- `packages/backend/internal/gameplay/dawn_review_test.go`。
- `packages/backend/internal/gameplay/game_session.go`。
- `packages/backend/internal/gameplay/night.go`。
- `packages/backend/internal/gameplay/information.go`。
- `packages/backend/internal/gameplay/game_session_snapshot.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `dawn_review_test.go`。
- 删除两条黎明命令、待审会话/快照字段与建议/确认辅助函数，恢复 ResolveNight 内联直接结算。
- 删除 `work.md` 中标题以 `2026-09-03 17:48:21` 开头的本节记录。

## 2026-09-03 17:51:29 +08:00 --- 黎明审核尚未穿透 WebSocket 且隐私未验证 --- 以单个完整对局失败测试驱动协议接线完成 RED→GREEN --- 修改协议、领域投影、网关、生成契约、测试与 `work.md`

### 发现什么问题
- 新领域路径没有协议消息或 RoomState 字段，页面无法使用；也未证明死亡建议只对说书人可见。

### 使用什么方式解决
- 将完整六端对局的首夜从直接 Resolve 改为 `PREPARE_DAWN → CONFIRM_DAWN`：首夜建议为空且仍在夜晚，只对说书人可见；确认后进入第一天。
- 多一次持久提交使后续恢复、终局和揭幕修订号整体加一（27/32/33）。
- 已运行定向测试并确认 RED：准备/确认消息和 RoomState 待审字段均不存在。
- 协议新增 `PREPARE_DAWN/CONFIRM_DAWN`，确认复用 `targetIds` 传零/多名最终死亡；RoomState 新增待审标记与建议名单。
- 领域投影只在说书人夜间管理能力下填充建议；Hub 进入统一事务，SessionEngine 透传并重新生成双端契约。

### 修改了哪些文件
- `packages/backend/internal/ws/protocol_v2_game_flow_test.go`。
- `proto/game.proto`。
- `packages/backend/internal/gameplay/projection.go`。
- `packages/backend/internal/ws/hub_v2.go`。
- `packages/backend/internal/ws/session_engine.go`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复单条 `RESOLVE_NIGHT`，删除建议隐私断言，并恢复修订号 26/31/32。
- 删除黎明协议消息/字段、Projection/Hub/SessionEngine 接线并重新生成契约。
- 删除 `work.md` 中标题以 `2026-09-03 17:51:29` 开头的本节记录。

## 2026-09-03 17:53:21 +08:00 --- Core 客户端不能调用黎明审核协议 --- 以单个失败测试驱动两条有序方法完成 RED→GREEN --- 修改 Core WebSocket 客户端、测试与 `work.md`

### 发现什么问题
- 服务端已有准备/确认消息，但 Core 客户端仍只能调用旧的直接 `resolveNight`。

### 使用什么方式解决
- 新增一个且仅一个客户端测试：先发送序号 1 的 `PREPARE_DAWN`，确认后再发送序号 2 且携带 P2/P4 的 `CONFIRM_DAWN`。
- 已运行定向测试并确认 RED：运行时明确报告 `prepareDawn is not a function`。
- 新增 `prepareDawn()` 与 `confirmDawn(deathPlayerIds)`，均通过序号队列发送，确认使用 targetIds 保留零/多人批次。

### 修改了哪些文件
- `packages/core/src/websocket/__tests__/websocket-client.test.ts`。
- `packages/core/src/websocket/index.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `sends dawn preparation and the edited death batch in sequence`。
- 删除两个 GameWebSocketClient 黎明方法。
- 删除 `work.md` 中标题以 `2026-09-03 17:53:21` 开头的本节记录。

## 2026-09-03 17:54:22 +08:00 --- Zustand 仓库不能调用黎明审核方法 --- 以单个失败测试驱动两动作转发完成 RED→GREEN --- 修改会话仓库、测试与 `work.md`

### 发现什么问题
- 结果已到 Core，但说书人页面仍不能经 Zustand 调用准备和确认。

### 使用什么方式解决
- 新增一个且仅一个仓库测试：准备无参数，确认原样转发 P2/P4，并将最后 pending 标记为 `confirm-dawn`。
- 已运行定向测试并确认 RED：运行时明确报告 `prepareDawn is not a function`。
- 将两个客户端能力、仓库动作及 `prepare-dawn/confirm-dawn` pending 分类完整接线。

### 修改了哪些文件
- `packages/frontend/src/lib/room-session-store.test.ts`。
- `packages/frontend/src/lib/room-session-store.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除测试 `forwards dawn preparation and edited death confirmation`。
- 从 FakeClient、SessionClient、RoomSessionState、PendingCommand 与实现删除两条黎明动作。
- 删除 `work.md` 中标题以 `2026-09-03 17:54:22` 开头的本节记录。

## 2026-09-03 17:55:37 +08:00 --- 说书人页面仍用旧的一键结束夜晚 --- 接入私密建议、多选编辑与确认黎明 --- 修改游戏页、样式与 `work.md`

### 发现什么问题
- 页面仍调用兼容用的 `resolveNight`，会绕过新审核；没有零/多人死亡编辑界面或最终确认文案。

### 使用什么方式解决
- 全部唤醒步骤完成后只开放“生成黎明死亡建议”；说书人投影返回建议后自动预选。
- 显示全部存活玩家作为可切换死亡批次，支持清空为零人或追加多人；弹窗列出最终名单，确认后才发送 `CONFIRM_DAWN` 并天亮。
- 旁观玩家在整个待审阶段仍维持普通夜晚等待视图。

### 修改了哪些文件
- `packages/frontend/src/pages/game-play/index.tsx`。
- `packages/frontend/src/pages/game-play/index.css`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复 `resolveNight` 选择器和单个“结束夜晚”按钮，删除黎明本地状态、同步 effect、切换/确认函数与审核面板。
- 删除 `.dawnReviewPanel` 样式。
- 删除 `work.md` 中标题以 `2026-09-03 17:55:37` 开头的本节记录。

## 2026-09-03 17:57:39 +08:00 --- 提名后直接开放投票，缺少控方陈述阶段 --- 以单个失败测试驱动定时阶段与投票门禁完成 RED→GREEN --- 修改游戏模型、白天流程、测试与 `work.md`

### 发现什么问题
- 当前提名立刻进入全席投票，已确认的“控方 30 秒→辩方 30 秒→逐席 3 秒”前两阶段不存在。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：提名必须以带服务端截止时间的 accusation 阶段开始，当前被提名席此时也不能投票。
- 已运行定向测试并确认 RED：Nomination 阶段、常量与截止字段均不存在。
- 提名模型新增 accusation/defense/voting、服务端毫秒截止、暂停与剩余时间字段；新提名以 30 秒控方阶段开始。
- CastVote 仅在 voting 阶段接纳；空阶段仅作为旧快照/手写夹具兼容，不影响新生产提名。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `packages/backend/internal/game/game.go`。
- `packages/backend/internal/gameplay/day.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `TestNominationStartsWithTimedAccusationBeforeVotesOpen`。
- 删除 Nomination 定时字段/阶段常量和新提名初始化/投票门禁。
- 删除 `work.md` 中标题以 `2026-09-03 17:57:39` 开头的本节记录。

## 2026-09-03 17:59:02 +08:00 --- 控方阶段无法推进到辩方和逐席投票 --- 新增单个两段推进测试进入 RED --- 修改投票规则测试与 `work.md`

### 发现什么问题
- 新提名会停在 accusation，没有说书人提前/到时推进命令，无法进入辩方与正式投票。

### 使用什么方式解决
- 新增一个且仅一个 gameplay 测试：说书人连续推进后依次进入带未来截止时间的 defense 和 voting（首席三秒）阶段。
- 当前处于 TDD RED 阶段，`AdvanceNominationStageCmd` 尚不存在。

### 修改了哪些文件
- `packages/backend/internal/gameplay/voting_rules_test.go`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 删除 `TestStorytellerAdvancesAccusationToDefenseAndVoting` 和 time 导入。
- 删除 `work.md` 中标题以 `2026-09-03 17:59:02` 开头的本节记录。

## 2026-09-03 17:56:27 +08:00 --- 新揭幕标记作为必填字段破坏旧投影夹具兼容 --- 改为缺省即未揭幕并重新生成契约 --- 修改 proto、生成契约与 `work.md`

### 发现什么问题
- TypeScript 全量构建显示所有旧 RoomState 夹具都必须机械补 `grimoireRevealed=false`；历史服务投影缺字段时也会产生不必要的不兼容。

### 使用什么方式解决
- 将协议字段改为 optional，缺失与 false 都表示未揭幕；页面和后端领域仍使用明确布尔语义。
- 重新生成 Go/TypeScript 契约，避免批量污染无关测试夹具。

### 修改了哪些文件
- `proto/game.proto`。
- `packages/backend/internal/ws/protocol_generated.go`。
- `packages/core/src/websocket/protocol.generated.ts`。
- `work.md`。

### 撤回方式 [Rollback Strategy]
- 恢复 `grimoire_revealed` 的 `json_required` 选项并重新生成契约。
- 删除 `work.md` 中标题以 `2026-09-03 17:56:27` 开头的本节记录。

## 2026-09-07 10:28 +08:00 --- 投票计时断点与暂停/超时缺失 --- 单个失败测试驱动暂停、恢复、重开与带截止令牌的超时默认否 --- 修改 game_session.go、day.go、新增 nomination_timer.go 和 nomination_timer_test.go

- 已验证 RED（命令不存在）；实现仅说书人控制计时，参与者可请求服务端验证过期，旧截止令牌不能推进下一席。补接上次已写入的控辩推进逻辑。
- 撤回：删除两个新文件及两个命令/分派、暂停门禁；保留此前用户修改。删除本节记录。

## 2026-09-07 10:33 +08:00 --- 控辩计时未贯通页面 --- 同步协议与命令、迁移旧对局测试到先控辩后投票、页面倒计时与暂停/重开 --- 修改 proto、生成器/生成契约、hub_v2.go、Core 客户端、会话仓库、game-play/index.tsx 与投票测试辅助

- 计时以服务端截止时间为准；说书人前台提交到期请求，服务端校验截止令牌且默认记否。说书人离线期间保留状态，重连后继续。
- 撤回：移除本次三条计时协议和双端接线，恢复测试旧投票流程后重新生成协议。

## 2026-09-07 10:33 +08:00 --- 缺少私人服务器部署方案 --- 子代理补 Docker 多阶段构建、Redis 持久卷和 Caddy HTTPS/WSS --- 修改 Dockerfile、compose.yaml、deploy/Caddyfile、.dockerignore、.env.example、docs/deployment.md、.gitignore

- YAML 静态解析通过；当前机器没有 Docker，容器与公网验收仍待部署环境执行。
- 撤回：删除新部署文件，仅撤回 .env.example/.gitignore 新段落，保留真实 .env 与数据卷。

## 2026-09-07 10:39 +08:00 --- 终局房间不能复用且房主不能移交 --- Session 事务权限校验与 Gameplay 重赛清理 --- 修改 session/types.go、session.go、新 lifecycle_test.go、gameplay/restart.go/restart_test.go、WS adapter/hub 与 proto

- 两个 Session 主行为及 gameplay 清理先 RED 后 GREEN，验证权限、幂等、持久失败回滚。重赛保留当前成员顺时针座位与说书人，清空角色、投票、死亡、毒、夜间临时信息及揭幕状态。
- 撤回：删除新重赛文件、Session 两分支及接口，删除协议命令与适配接线后重新生成。

## 2026-09-07 10:39 +08:00 --- 手机页面样式割裂且夜间结果裸露 --- 子代理统一主题和私密结果遮罩 --- 修改 app.css、五个页面 CSS、私密角色 CSS；新增 private-night-result.tsx/test.tsx；修改 night.go/information.go/projection.go/night_player_flow_test.go

- 统一暗色金色主题、46px 触控区、安全区与横屏双栏；信息角色只请求信息，由说书人选择线索；已阅玩家不再收到秘密；结果卡长按查看，后台/松手/30秒隐藏。
- 验证 CSS 解析 7/7，私密组件 3 条测试和夜间行为测试通过。撤回：移除 CSS 末尾主题块和新增组件，逆向移除信息角色/已阅投影逻辑，不触及此前改动。

## 2026-09-07 10:37 +08:00 --- 恢复身份入口缺失、夜间结果未接隐私组件 --- 测试驱动恢复并接线私密结果 --- 修改 room-session-store.ts/test.ts、game-play/index.tsx、character-display.ts；新增 room-recovery-panel.tsx/css、room-recovery.ts/test.ts、room-invite.ts/test.ts；修改首页和 setup 邀请

- 恢复码解析与导入公开接口先 RED 后 GREEN；导出默认遮罩主动确认，导入密码输入后台清空；复用原座位凭据而非重新加入。当前为持凭据恢复，不包含说书人审批。
- 邪恶互认仅显示玩家姓名，补齐爪牙同伴，修改 information.go、night_player_flow_test.go、evil_team_information_test.go、protocol_v2_game_flow_test.go，相关 Go 回归通过。
- 撤回：移除新增恢复文件、导入入口及其测试，恢复邀请路径拼接和夜间明文块；逆向移除互认名单改动及相应断言，不回滚其他工作。

## 2026-09-07 10:40 +08:00 --- 重赛/转房主无前端入口，旧夜晚命令可绕过审核 --- 补全客户端入口并禁用旧客户端夜晚结算 --- 修改 core websocket index/test、frontend store/test、setup/index.tsx、game-over/index.tsx、ws/hub_v2.go；新增 ws/legacy_night_test.go

- 生命周期测试先红后绿，Core/Store 共 50 项通过，TypeScript 编译通过。转让和再开均二次确认，等待服务器权威投影；旧 RESOLVE_NIGHT 回归先失败后禁用，使用生成建议/确认黎明两步。
- 撤回：仅移除转让/重赛方法、对应类型和入口、新增测试；恢复旧命令映射。保留之前房间与夜间代码。

## 2026-09-07 10:44 +08:00 --- 夜间规则边界错误、验收范围不透明 --- 修复并文档化实际交付边界 --- 修改 gameplay/information.go、night.go 与夜间测试及 WS fixture；README.md；新增 docs/playtest-checklist.md；修改 SessionShell、恢复组件、app.css、room-invite.ts

- 共情跨死亡邻座、空信息确认拒绝、5–6人无邪恶互认、只有实际小恶魔自杀才传位，四项先红后绿。
- 房内接恢复面板，CSS 单点导入消除微信顺序警告；H5/微信构建和 tsc 已通过；19 文件 243 条前端测试通过，协议漂移检查通过。
- 验收文档明确未实现的审批/撤销/清理及未验证公网/真机，不将构建成功等同产品全部完成。
- 撤回：逐项逆向撤销上述逻辑及断言；移除新验收文档和 README 本轮说明；移除 Shell 面板，恢复组件 CSS import。不整体还原共享文件。

## 2026-09-07 10:49 +08:00 --- 手机实测样式缺陷、投票顺序与重赛边界、间谍魔典缺失 --- 子代理实测并修复，扩展完整对局回归 --- 修改 app.css、恢复面板CSS、准备页CSS/TSX、game-play/index.tsx、私密角色CSS；gameplay/day.go、game_session.go及投票/重赛测试；WS完整对局测试；game/game.go、characters.go、gameplay/night.go/information.go/projection.go及夜间测试；core night-phase/scripts及测试

- 390×844 实测创建、指定说书人、独立origin加入、准备/取消、刷新恢复和恢复遮罩通过，无横向溢出。修复 disabled="false" 被误灰化、2×2 概况卡盒模型和弹层按钮间距。
- 官方顺时针票序以被提名人最后投；再开清空秘密后可换说书人，原说书人回到目标座位。完整 WS 测试覆盖结束→公开→再开→换说书人。
- 间谍每晚专属步骤复用私密结果卡，不再整夜下发所有人的角色字段；中毒必须手填假魔典。增加中文动作名称与换行显示。
- Go 全量与 -race 通过，tsc 通过；间谍相关16个TS测试通过。最终双端构建另行复验。
- 撤回：仅逐项逆向移除上述选择器/盒模型/间距、票序和交换ST分支及其测试、新间谍动作/步骤/结果格式和投影限制；不要整体还原共享文件或删除其他工作。

## 2026-09-07 10:49 +08:00 --- 需要最终构建与验收状态 --- 顺序构建 weapp 后 H5 并补实测范围 --- 修改 docs/playtest-checklist.md

- 前端/核心244条测试通过，最终 weapp 构建8.17秒无警告，H5构建9.96秒仅包体积提示；最终dist为本地预览H5，连接127.0.0.1:8080，不是公网发布配置。
- 浏览器验收限于已列出的房间/恢复路径，完整游戏自动化另列，避免混淆。撤回：移除清单新增验证段与间谍入口说明；构建产物可用对应构建命令重新生成。

## 2026-09-07 10:53 +08:00 --- 不活跃房间长期保留、到期提示不准确 --- 事务活动时间与七天安全清理、补提示并复验 --- 修改 session/types.go、registry.go、session.go；新增 session/expiry.go/test、ws/expiry.go/test；修改 cmd/server/main.go、backend README/CONTEXT、frontend room-session-store.ts、docs/playtest-checklist.md、gameplay/restart.go注释

- 成功持久化才续期；恰好七天、未来时间及旧无时间戳记录不直接删除。旧记录CAS写入完整宽限期，过期房间CAS删除成功后才关闭连接，失败保留。启动及每小时扫描，进程信号触发HTTP关闭。
- 新增边界、失败、恢复、已连接玩家通知回归，最终Go全量和-race通过，tsc通过，weapp 9.97秒无警告，H5 12.14秒仅体积提示。最新后端健康检查200。
- 重启本轮临时内存预览后端，测试房间已清除且不可恢复；没有删除真实持久化存档。正式清理删除的数据需从部署备份恢复。
- 撤回：删除新增expiry文件与LastActiveAt写入/扫描接线、恢复文档及提示的本轮段落；不要删除真实数据卷。restart注释可直接逆向恢复。

## 2026-09-07 11:09 +08:00 --- 恢复审批可被旧凭据绕过、缺少撤销和守鸦人触发 --- 独立签名恢复凭据及持久授权、有限历史事务、延迟黎明发布 --- 修改 proto/game.proto及生成器/生成文件；session/types.go/session.go，新增recovery.go/test及recovery_approval_test.go；ws/hub.go/hub_v2.go/session_engine.go，新增recovery.go/test、history.go/test；修改core websocket及测试、frontend恢复码/store/utils及测试/组件；新增game-history-panel.tsx/test；修改game-play、SessionShell；修改gameplay夜间/信息与守鸦人测试；backend README

- CT3仅导出独立recoveryCredential；签名不能用于RESUME，旧CT2拒绝。批准前不换连接、不保存新身份、不下发角色投影；ST审批、恢复ST由房主审批、唯一审批者不能自批。批准原子轮换nonce并保留10分钟签名绑定授权用于丢包/重启重试，旧设备凭据失效并收到替换通知。
- Session真实HMAC回归覆盖权限、签名、持久失败、nonce轮换、重启、幂等及二次轮换；WS申请→等待→批准→旧凭据失效先红后绿。
- 日志仅ST可见，最近100条/30步；只回滚游戏状态与锁定标记，成员、凭据和序列不变；跨阶段二次确认，恢复投票默认暂停，成员/座位改变清历史边界。私密面板测试先红后绿。
- 守鸦人仅在ST锁定其死亡后唤醒，私密选择/审核/已阅后统一天亮；多人死亡和快照恢复回归通过，无新proto字段。
- 撤回：按上述文件本轮差异逐块逆向移除，不整体覆盖共享改动；历史额外snapshot字段旧加载器可忽略。协议与客户端/服务端须共同回滚；旧恢复码不能混用，回滚后需重新导出。保留真实持久化数据及备份。

## 2026-09-07 11:15 +08:00 --- 审批重放误踢、拒绝改判、恢复竞态与等待页刷新丢请求 --- 边缘回归和有界恢复票据 --- 修改 ws/recovery.go、hub_v2.go、expiry.go；新增 ws/recovery_edges_test.go；修改core/websocket与frontend/store及测试、docs/playtest-checklist.md

- 五组WS边缘回归先红后绿：重复批准只重发结果、终态不允许新序列改判；恢复认证和连接接管与撤销互斥；每分钟清理过期恢复请求；批准丢包及服务端重启授权可重放而不带投影。
- 前端只保存requestId/roomId/playerId非秘密票据，刷新后重贴同码可重用原请求；未批准身份和恢复凭据不落盘，终态删除票据。
- Session真实签名测试及全Go、-race通过；前端新增测试和tsc通过。重启本轮临时预览后端到最新版本，仅清理自动验收内存房间，无真实存档删除。
- 撤回：逐块移除终态与重放/互斥/恢复票据逻辑及新测试，恢复相应文档；不整体回滚共享文件。不建议单独撤回安全修复，协议两端需保持一致。

## 2026-09-07 11:26 +08:00 --- 真机尺寸实测发现弹窗长度与复制备用入口问题、旧领域文档过时 --- 三端实测修复并最终验收 --- 修改恢复面板TSX/CSS、game-play/index.tsx；新增lib/dialog-contract.test.ts；修改README、三包CONTEXT、docs/playtest-checklist.md

- Taro确认按钮最多4字；复制恢复码、确认并天亮、无处决入夜三处5字已改短，新增全页面弹窗契约检查先红后绿。测试使用NodeURL避免Taro全局URL类型冲突。
- 默认仍遮罩，新增主动显示恢复码备用入口：二次确认，30秒/关闭/后台隐藏，不额外落盘恢复凭据。实测复制成功提示、可见CT3不含直接resume凭据，30秒从DOM移除。
- 三独立origin实测：准备→撤销→重做双端同步；申请→刷新重贴同码保持同请求→ST批准→新端原座位与ready保留→旧端退出→新码轮换。未读取隐藏存储绕过流程，测试标签关闭且工具变量中测试码清空。
- 最终22文件255条前端/核心测试通过，Go全量和-race通过，tsc/协议漂移/git diff --check通过；weapp 5.90秒无警告，H5 9.82秒仅体积提示。最终dist保留临时本地H5预览；公网Docker/Redis/Caddy与真实微信设备尚未验收，没有执行公网部署或开源发布。
- 撤回：仅逆向上述文案、备用显示入口、测试及文档补充；保留CT3分离审批安全约束。产物可重新构建。改动均保留本地，未执行git reset/覆盖用户既有更改。

## 2026-09-07 11:54 +08:00 --- 页面缺乏官网氛围且管理表单堆叠成长页 --- 官网视觉参考与桌游App分区布局 --- 修改SessionShell；新增room-navigation.tsx/test.tsx；修改准备/对局/结算/图鉴页面TSX，新增pages/layout.test.tsx、game-play/workspace.test.tsx

- 参考官网实际呈现的深紫黑、羊皮纸、复古衬线层次，以及BGA官方移动端布局指引中的主桌优先/窄屏重排；不复制官方图片、标志或字体文件。
- 公共框架改顶栏+独立滚动区+底部主桌/图鉴/记录/更多，辅助日志/审批/恢复从主桌移出；记录仍只使用权威投影，ST私密日志权限保留。准备和对局改局部标签工作区，图鉴搜索筛选，结算身份卡阵列与折叠时间线。
- 导航SSR先RED后GREEN；准备/图鉴/结算及默认行动/私密卡边界测试通过，TypeScript检查通过。CSS与最终视觉验收继续进行。
- 撤回：仅逆向本轮布局/局部UI状态/props差异，删除新增导航和本轮布局测试；不要整体还原页面或触动此前完成的业务逻辑、协议、存储与权限。
- 工具简报：Context7查询Taro4 ScrollView固定高度及导航限制；CUA查看官网与旧本地界面；web读取BGA官方移动端指引。时间2026-09-07。

## 2026-09-07 12:01 +08:00 --- 大厅重复输入与各页样式叠加不统一 --- 单表单大厅及全局视觉系统重写 --- 修改首页TSX、pages/layout.test.tsx、app.config.ts；重写app.css、5页CSS、私密身份CSS、恢复面板CSS

- 大厅只保留一份昵称，创建/加入切换；邀请只切加入并预填不自动提交。增加原创CSS钟楼介绍区和剧本辅卡，开发工具移更多。默认及邀请页面SSR先红后绿。
- 统一深紫黑/羊皮纸/酒红/古金、中文衬线标题与易读正文；去除旧主题叠加覆盖。共享座位卡/标签/主次栏，底部导航与操作栏分层，滚动区固定高度与安全区留白。原生导航栏背景同步主题。
- 263条前端/核心测试及TypeScript通过，8份CSS解析通过；开始手机/桌面实际视觉验收。
- 撤回：仅逆向本轮首页/样式/app.config差异及新增首页测试；保留上一轮游戏功能、审批和存储协议。不整体回滚仓库。

## 2026-09-07 12:05 +08:00 --- 手机截图出现标签横滚、白色外滚条、概况过高 --- 按实测元素尺寸修正布局 --- 修改app.css、game-setup/index.css及index.tsx；新增docs/ui-design.md

- pageTab显式width:auto/min-width:0/flex收缩，三标签同屏；H5外层taro_page关闭滚动，仅保留Shell固定高度滚动区；手机隐藏内部滚条。概况卡紧凑四列、ST长名省略，菱形分隔居中。
- 准备页按权限将准备/配置/开始主操作固定底部，私密卡前置且仍需长按确认；少于5玩家时明确提示邀请。新增设计规范与来源说明。
- 撤回：仅逆向上述微调及新增设计文档，保留私密机制与业务回调。首轮H5构建通过，正在复验双端构建及截图。

## 2026-09-07 12:15 +08:00 --- Taro路由后加载样式覆盖滚动约束、图鉴刷新后返回失效 --- 提高外层选择器优先级并验证页面栈导航 --- 修改app.css、SessionShell；新增lib/table-navigation.ts/test.ts

- 实测390px有效页面宽仅375px，CSSOM确认后加载.taro_page覆盖；改为.taro_router直接子页选择器，实际外层overflow:hidden、页面宽390、三标签各115px同屏。
- 图鉴刷新后navigateBack可能成功返回但无路由变化；使用官方getCurrentPages寻找目标主桌并保留有效页面栈，否则reLaunch权威状态对应路由。新增导航边界测试先RED后GREEN；tsc通过。
- 撤回：仅逆向选择器与导航helper接线、删除helper/test；不修改服务器或现有房间身份。
- 工具简报：CUA检查CSSOM、375/390截图、搜索与标签切换；Context7查询Taro4 getCurrentPages/navigateBack/reLaunch，2026-09-07。

## 2026-09-07 12:23 +08:00 --- 全站视觉与布局交付验收 --- 自动化检查及真实浏览器响应式验证 --- 修改work.md（验收记录）

- 最终26个测试文件、264条前端/核心测试通过，TypeScript构建通过，H5与微信构建通过；H5仅有包体积提示，git diff --check通过。
- 浏览器验证375/390px手机及1440px桌面布局；大厅创建/加入切换、图鉴搜索与夜序、更多、房间设置和记录入口正常；图鉴刷新后可返回大厅，页面没有横向溢出。
- 对局与结算以页面测试、类型及构建验证，未对用户当前零玩家房间模拟开局；未修改房间成员、后端、协议或游戏规则，未进行微信真机或公网部署验收。
- 保留新版大厅预览http://localhost:10124/#/pages/index/index；原房间回到默认座位分区。设计来源与约束见docs/ui-design.md。
- 撤回：本条仅为验收记录，可删除本节；代码撤回按本轮各条日志逐项逆向补丁，不整体还原既有未提交成果。
- 工具简报：Context7核对Taro4布局和导航约束；CUA完成页面交互与响应式实测；终端执行测试、类型检查、双端构建及差异检查。时间2026-09-07。

## 2026-09-07 13:30 +08:00 --- 当前阶段成果尚未提交 --- 按用户要求创建本地Git提交 --- 修改work.md并提交当前源码、测试、部署模板与文档

- 范围包含此前房间/对局/恢复流程成果及本轮全站布局优化；检查新增文件清单和部署环境模板，常见私钥及令牌模式扫描未命中，真实.env和构建产物不纳入提交。
- 本次仅提交到当前main分支，不推送远端；复用上轮已完成的264条测试、类型和双端构建验收，本次另执行差异空白检查。
- 撤回：需要撤销本次成果时使用git revert对本次提交生成反向提交；本条日志可随反向提交撤回，不使用硬重置或删除工作区。
