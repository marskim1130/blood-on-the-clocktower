# PRD: Blood on the Clocktower MVP — Human Storyteller Mode

## Problem Statement

桌游爱好者想要一个数字化的 Blood on the Clocktower 体验，以便在无法线下聚会时也能和朋友一起玩这款社交推理游戏。现有的数字版本要么功能不完整，要么用户体验差，无法替代实体桌游的流畅感。

## Solution

构建一个基于微信小程序的 Blood on the Clocktower 数字版本，支持人类 Storyteller 模式。玩家通过微信小程序加入房间，一个玩家担任 Storyteller 角色控制游戏流程，系统自动处理投票、状态同步等机械性工作，让 Storyteller 专注于游戏决策而非繁琐的记录工作。

## User Stories

### 房间管理

1. As a 玩家, I want to 创建一个新游戏房间, so that 我可以邀请朋友开始一局游戏
2. As a 玩家, I want to 通过房间号加入已有房间, so that 我可以和朋友一起玩
3. As a 玩家, I want to 在房间内看到所有已加入的玩家列表, so that 我知道还有谁在房间里
4. As a 玩家, I want to 设置自己的昵称, so that 其他玩家能识别我
5. As a 玩家, I want to 在游戏开始前离开房间, so that 我可以取消不想玩的游戏
6. As a 房主, I want to 将某个玩家踢出房间, so that 我可以管理不守规矩的玩家
7. As a 房主, I want to 设置游戏参数（玩家人数范围）, so that 游戏符合我的预期
8. As a 系统, I want to 在游戏结束后自动销毁房间, so that 服务器资源得到释放

### Storyteller 控制

9. As a Storyteller, I want to 被指定为本局游戏的主持人, so that 我可以控制游戏流程
10. As a Storyteller, I want to 为每个玩家分配角色, so that 游戏可以开始
11. As a Storyteller, I want to 查看所有玩家的角色分配, so that 我了解当前局势
12. As a Storyteller, I want to 进入夜间阶段并唤醒特定玩家, so that 他们可以使用角色能力
13. As a Storyteller, I want to 记录夜间行动的结果, so that 游戏状态正确更新
14. As a Storyteller, I want to 宣布某个玩家死亡, so that 游戏状态反映实际情况
15. As a Storyteller, I want to 判定游戏胜负, so that 游戏可以正常结束

### 游戏流程

16. As a 玩家, I want to 看到当前游戏阶段（白天/夜晚/投票）, so that 我知道现在该做什么
17. As a 玩家, I want to 在白天阶段看到所有存活玩家, so that 我可以进行讨论
18. As a 玩家, I want to 提名其他玩家接受处决投票, so that 我可以推动游戏进程
19. As a 玩家, I want to 对提名进行投票, so that 我表达我的立场
20. As a 玩家, I want to 看到投票结果, so that 我知道谁被处决了
21. As a 玩家, I want to 看到哪些玩家已经死亡, so that 我了解游戏局势
22. As a 玩家, I want to 在投票阶段收到通知, so that 我不会错过投票

### 角色系统

23. As a 玩家, I want to 看到自己的角色信息, so that 我知道我的能力
24. As a 玩家, I want to 看到自己角色的能力描述, so that 我知道如何使用能力
25. As a Storyteller, I want to 从 Trouble Brewing 角色池中选择角色, so that 游戏使用正确的角色组合
26. As a Storyteller, I want to 查看每个角色的能力说明, so that 我正确执行夜间行动
27. As a 系统, I want to 验证角色分配的合法性（村民/外来者/爪牙/恶魔数量）, so that 游戏配置符合规则

### 状态同步

28. As a 玩家, I want to 实时看到其他玩家的状态变化, so that 我了解最新局势
29. As a 玩家, I want to 在断线重连后恢复游戏状态, so that 我不会因为网络问题丢失进度
30. As a 系统, I want to 在服务器重启后恢复进行中的游戏, so that 玩家不会丢失游戏进度
31. As a 系统, I want to 向所有玩家广播游戏状态变化, so that 所有人看到一致的画面

### 用户体验

32. As a 玩家, I want to 在微信小程序中流畅运行, so that 我无需下载额外应用
33. As a 玩家, I want to 通过微信分享房间链接, so that 我可以快速邀请朋友
34. As a 玩家, I want to 看到清晰的 UI 提示当前应该做什么, so that 我不会迷茫
35. As a 玩家, I want to 收到重要事件的推送通知, so that 我不会错过关键行动

## Implementation Decisions

### 架构决策

1. **Monorepo 结构**：使用 pnpm workspaces 管理 `@clocktower/core`、`@clocktower/frontend`、`@clocktower/backend` 三个包
2. **端到端类型安全**：ProtoBuf 定义数据结构，自动生成 Go 和 TypeScript 类型
3. **状态管理**：后端权威游戏会话持有已提交状态，Frontend Zustand 仓库保存接收者特定完整投影
4. **通信协议**：WebSocket v2 用于房间管理、游戏命令与实时完整状态同步

### 模块划分

**Core Game Logic（需要单元测试）：**
- **Vote Engine**：提名管理、投票收集、计票逻辑、处决判定
- **Character System**：Trouble Brewing 角色定义、能力描述、角色合法性验证

**Backend Services（需要单元测试）：**
- **Room Manager**：房间 CRUD、玩家加入/离开、房间参数配置
- **Storyteller Controller**：角色分配、夜间行动执行、死亡宣告、胜负判定
- **WebSocket Gateway**：连接生命周期管理、消息路由、房间广播

**Frontend UI（不需要单元测试）：**
- **Lobby**：创建/加入房间界面
- **Game Board**：游戏主界面、玩家列表、投票 UI
- **Storyteller Panel**：Storyteller 控制面板

### 数据模型

`proto/game.proto` 生成 WebSocket v2 的 `RoomState`、`ClientMessage` 与 `ServerMessage`。Frontend 只保存通过修订门禁 [Revision Gate] 的完整 `RoomState`，不维护平行的 `GameState`/`GameEvent` 模型。

### API 契约

**WebSocket 消息格式：**
```json
// 客户端 → 服务器
{ "protocolVersion": 2, "type": "JOIN_ROOM", "roomId": "...", "playerId": "...", "playerName": "Alice", "joinRequestId": "..." }
{ "protocolVersion": 2, "type": "CAST_VOTE", "roomId": "...", "playerId": "...", "resumeCredential": "...", "clientSequence": 3, "targetPlayerId": "..." }

// 服务器 → 客户端
{ "type": "JOIN_ROOM_RESULT", "roomId": "...", "roomRevision": 2, "resumeCredential": "...", "state": { ... } }
{ "type": "ROOM_STATE_CHANGED", "roomId": "...", "roomRevision": 3, "state": { ... } }
```

服务端只发送接收者特定的完整权威投影 [Recipient-specific Authoritative Projection]；客户端不通过增量事件重建已提交状态。完整契约见 `docs/protocol/websocket-v2.md`。

### 关键交互流程

1. **创建房间**：玩家 A 创建房间 → 获得房间号 → 分享给朋友
2. **加入房间**：玩家 B/C/D 输入房间号加入 → 等待开始
3. **指定 Storyteller**：房主指定自己或其他人为 Storyteller
4. **分配角色**：Storyteller 为每个玩家分配 Trouble Brewing 角色
5. **白天阶段**：玩家讨论 → 提名 → 投票 → 处决
6. **夜间阶段**：Storyteller 唤醒玩家 → 执行能力 → 记录结果
7. **游戏结束**：满足胜利条件 → Storyteller 宣布结果 → 保留最终投影供成员重连查看；房主可显式关闭房间

## Testing Decisions

### 测试原则

- 只测试外部行为，不测试实现细节
- 测试应覆盖边界情况和错误路径
- 使用真实的领域术语，避免 mock 泄露实现细节

### Core Game Logic 测试

1. **WebSocket Client 测试**
   - 验证断线恢复、客户端序号与命令重放
   - 验证房间修订门禁与完整状态重同步
   - 验证身份终止后不再自动恢复

2. **Vote Engine 测试**
   - 验证提名规则（只能提名存活玩家）
   - 验证投票规则（每人只能投一次）
   - 验证计票逻辑（多数票通过）
   - 验证平票处理

3. **Character System 测试**
   - 验证角色分配合法性（村民+外来者+爪牙+恶魔数量正确）
   - 验证角色能力描述完整性

### Backend Services 测试

1. **Room Manager 测试**
   - 验证房间创建/销毁
   - 验证玩家加入/离开
   - 验证房间人数限制

2. **Storyteller Controller 测试**
   - 验证角色分配权限（只有 Storyteller 能分配）
   - 验证夜间行动执行顺序
   - 验证死亡宣告逻辑

3. **WebSocket Gateway 测试**
   - 验证连接生命周期
   - 验证消息广播范围（只广播给房间内玩家）
   - 验证断线重连

## Out of Scope

以下功能不在 MVP 范围内，将在后续迭代中实现：

1. **AI Storyteller 模式** — 系统自动执行 Storyteller 职责
2. **Bad Moon Rising / Sects & Violets 角色集** — 只支持 Trouble Brewing
3. **自定义脚本** — 玩家自定义角色组合
4. **游戏回放** — 游戏结束后查看回放
5. **排行榜/统计** — 玩家胜率、游戏历史
6. **语音/视频聊天** — 集成实时通信
7. **React Native 版本** — 只支持微信小程序
8. **用户注册系统** — 使用匿名房间制
9. **付费功能** — 高级角色、自定义主题等

## Further Notes

### 服务器配置

- 硬件：2c2g 腾讯云
- 并发目标：20-30 局同时运行
- 持久化：Redis
- 玩家规模：每局 5-15 人

### 技术栈

- 前端：Taro 4.x (React) + Zustand
- 后端：Go + gRPC/WebSocket
- 类型同步：ProtoBuf
- 包管理：pnpm workspaces

### 后续迭代方向

1. **Phase 2**：AI Storyteller 模式
2. **Phase 3**：Bad Moon Rising / Sects & Violets 角色集
3. **Phase 4**：自定义脚本系统
4. **Phase 5**：React Native 跨平台
5. **Phase 6**：游戏回放、排行榜、社交功能
