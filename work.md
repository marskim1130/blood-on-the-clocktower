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
