# ProtoBuf 定义 [ProtoBuf Definitions]

`game.proto` 是 Go 后端与 TypeScript 客户端共享数据结构和 WebSocket v2 信封的规范来源 [Canonical Source]。线上 WebSocket 仍使用 JSON 编码；ProtoBuf 负责字段、枚举和可选性的跨语言契约 [Cross-language Contract]，不改变传输格式。JSON 必填字段通过 `json_required` 字段选项声明，生成器不得维护第二份必填清单。

## 生成契约 [Generate Contracts]

```bash
pnpm proto:generate
```

该命令会生成：

- 游戏领域 TypeScript 类型：`packages/core/src/types/generated/`
- WebSocket TypeScript 契约：`packages/core/src/websocket/protocol.generated.ts`
- WebSocket Go 契约：`packages/backend/internal/ws/protocol_generated.go`

生成文件带有 `game.proto` 内容哈希。不要直接编辑生成文件。

## 验证漂移 [Drift Check]

```bash
pnpm proto:check
```

该命令重新计算所有被版本控制跟踪的预期输出，但不写文件；任何手写修改或未重新生成的 ProtoBuf 变更都会失败。`pnpm verify` 会先执行此检查。

## 修改流程 [Change Workflow]

1. 在 `game.proto` 修改消息或枚举。
2. 运行 `pnpm proto:generate`。
3. 更新消费生成类型的业务逻辑与测试。
4. 运行 `pnpm verify`。
