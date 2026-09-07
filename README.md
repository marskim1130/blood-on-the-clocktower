# Blood on the Clocktower

[Blood on the Clocktower](https://bloodontheclocktower.com/) 桌游的数字实现 [Digital Implementation] —— 支持人类 Storyteller [Human Storyteller] 模式的社交推理游戏。玩家通过微信小程序加入房间，一位玩家担任 Storyteller 控制游戏流程，系统自动处理投票、状态同步等机械工作，让 Storyteller 专注于游戏决策。

> 本项目是桌游《血染钟楼》的爱好者数字实现，与 The Pandemonium Institute 无关联。游戏规则版权归原作者所有。

## 功能特性 [Features]

- **实时房间 [Real-time Rooms]**：创建 / 加入 / 离开 / 踢人 / 关闭房间，基于 WebSocket 的实时状态广播
- **人类 Storyteller 模式**：Storyteller 分配角色、控制昼夜阶段、记录夜间行动、判定胜负
- **Trouble Brewing 角色规则**：内置镇民 / 外来者 / 爪牙 / 恶魔的规则校验与夜间行动处理；特殊裁定仍由真人说书人负责。
- **端到端类型同步 [End-to-end Type Safety]**：`proto/game.proto` 作为单一事实源，生成 TypeScript 与 Go 双向契约，编译期验证类型兼容
- **断线恢复 [Session Resume]**：凭据签名 [Resume Credential] + 幂等命令序列，断线重连不丢进度
- **多级持久化 [Persistence]**：Redis / 文件 / 内存三种存储适配器，服务器重启可恢复房间
- **手机前端**：本轮验收目标是 H5 与微信小程序；包含大厅、准备、对局、结算和剧本查询。React Native 尚未纳入本轮验收。
- **面杀流程**：准备/确认身份、私密夜间信息、控辩与逐席投票、黎明确认、说书人揭幕、同房再开。
- **恢复与纠错**：独立 CT3 恢复码经说书人/房主审批后轮换凭据；说书人私密操作日志、有限撤销/重做与跨阶段确认；七天不活跃房间清理。

当前验收范围与未完成项见 [手机面杀验收清单](docs/playtest-checklist.md)，公网部署见 [部署说明](docs/deployment.md)。

## 技术栈 [Tech Stack]

| 层 | 技术 |
|---|---|
| 前端 | Taro 4.x (React)、Zustand、TypeScript |
| 后端 | Go 1.22、gorilla/websocket、gRPC、ProtoBuf |
| 类型契约 | ProtoBuf（`proto/game.proto` → TS / Go 代码生成） |
| 包管理 | pnpm workspaces |

## 项目结构 [Project Structure]

```text
blood-on-the-clocktower/
├── packages/
│   ├── core/       # @clocktower/core —— 共享规则、WebSocket 客户端、生成的协议类型
│   ├── frontend/   # @clocktower/frontend —— Taro 4.x 前端（微信小程序 / RN）
│   └── backend/    # @clocktower/backend —— Go 后端服务
├── proto/          # ProtoBuf 契约定义（game.proto）
├── docs/           # PRD、协议、代理协作文档
└── scripts/        # 协议契约生成脚本
```

## 快速开始 [Quick Start]

### 环境要求

- Node.js >= 20，pnpm >= 9
- Go 1.22+

### 安装依赖

```bash
pnpm install
```

### 启动后端

后端需要至少 32 字节的持久凭证密钥 [Credential Signing Key]：

```powershell
$env:CLOCKTOWER_CREDENTIAL_KEY='replace-with-a-persistent-secret-of-at-least-32-bytes'
cd packages/backend
go run ./cmd/server
```

- WebSocket 端点：`ws://localhost:8080/ws`
- 健康检查：`GET http://localhost:8080/health`

可选持久化配置（见 `packages/backend/README.md`）：

| 环境变量 | 作用 |
|---|---|
| `CLOCKTOWER_REDIS_URL` | 启用 Redis 每房间持久化 |
| `CLOCKTOWER_REDIS_KEY` | Redis 命名空间前缀 |
| `CLOCKTOWER_SNAPSHOT_PATH` | 文件持久化目录 |

> 当前部署模型只支持**单个活动后端实例 [Single Active Backend Instance]**。

### 启动前端

```bash
# 微信小程序（配合微信开发者工具打开 dist 目录）
pnpm dev:frontend

# H5 版本
pnpm --filter @clocktower/frontend run dev:h5
```

前端通过 `TARO_APP_WS_URL` 环境变量配置 WebSocket 地址。

### 一键开发

```bash
pnpm dev        # 同时启动后端与前端
pnpm verify     # 完整验证：proto 漂移检查 + 测试 + 构建
```

## 验证与测试 [Verification]

```bash
pnpm verify     # 一体化验证入口（CI 使用同一命令）
```

包含：ProtoBuf 契约漂移检查 → 后端 Go 测试（含 `-race`）→ core 构建与测试 → 前端构建与测试。

## 协议与架构 [Protocol & Architecture]

- `docs/protocol/websocket-v2.md` —— WebSocket v2 协议（序列号排序、会话恢复、状态同步）
- `proto/README.md` —— ProtoBuf 契约生成与漂移检查流程
- `CONTEXT-MAP.md` —— 各包的领域文档导航（每包含 `CONTEXT.md` 与架构决策记录 [ADR]）

## 文档 [Documentation]

| 文档 | 说明 |
|---|---|
| [PRD](docs/prd/mvp-human-storyteller.md) | MVP 产品需求（人类 Storyteller 模式） |
| [后端部署](packages/backend/README.md) | 后端配置与启动说明 |
| [WebSocket v2 协议](docs/protocol/websocket-v2.md) | 实时通信协议细节 |
| [ProtoBuf 契约](proto/README.md) | 类型同步与代码生成 |

## 许可 [License]

[MIT](LICENSE)
