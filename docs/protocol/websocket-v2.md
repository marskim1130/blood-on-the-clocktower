# WebSocket 协议 v2 [WebSocket Protocol v2]

后端只接受 `protocolVersion: 2`。协议 v1、无凭证重连和旧全局快照不兼容。

## 客户端信封 [Client Envelope]

所有消息包含 `protocolVersion` 与 `type`。身份相关消息按需包含 `roomId`、`playerId`、`resumeCredential`、`clientSequence`、`requestId` 或 `joinRequestId`。

- `CREATE_ROOM`：使用不可猜测且可重试的 `requestId`。
- `JOIN_ROOM`：使用不可猜测且可重试的 `joinRequestId`。
- `RESUME_ROOM`：只恢复或接管活动连接，不消耗客户端序号。
- `REJOIN_ROOM`：把保留身份 [Retained Identity] 恢复为房间成员，消耗客户端序号。
- `GET_ROOM_STATE`：读取当前接收者投影，不持久化、不广播、不消耗序号。
- `CLOSE_ROOM`：创建者专属，可使用分离式管理命令 [Detached Management Command] 执行。

所有改变权威状态的身份命令必须携带单调客户端序号 [Monotonic Client Sequence]。

## 服务端信封 [Server Envelope]

成功消息按需包含 `state` 或互斥的 `identityStatus`，以及同一已提交视图中的 `roomRevision`、`acceptedSequence` 与 `nextClientSequence`。首次创建或加入还返回 `resumeCredential`。

状态消息包括 `CREATE_ROOM_RESULT`、`JOIN_ROOM_RESULT`、`RESUME_ROOM_RESULT`、`COMMAND_RESULT`、`ROOM_STATE`、`ROOM_STATE_CHANGED`、`KICKED` 与 `ROOM_CLOSED`。

## 稳定错误码 [Stable Error Code]

`UNSUPPORTED_PROTOCOL`、`INVALID_MESSAGE`、`ROOM_NOT_FOUND`、`INVALID_CREDENTIAL`、`STALE_CONNECTION`、`FORBIDDEN`、`PARTICIPANT_SET_FROZEN`、`UNEXPECTED_SEQUENCE`、`SEQUENCE_CONFLICT`、`IDEMPOTENCY_CONFLICT`、`PERSISTENCE_UNAVAILABLE`、`PERSISTENCE_CONFLICT`、`INTERNAL`。

凭证错误、身份不存在、已离开和被封禁不得通过公开错误文本相互区分。

## 修订同步 [Revision Synchronization]

`roomRevision` 只表示持久化提交顺序。客户端只应用连续的新修订；相同修订视为幂等重复。发现跳跃或倒退时，请求 `GET_ROOM_STATE` 并等待完整接收者投影。

## 实时投影投递 [Real-time Projection Delivery]

实时游戏会话投影 [Real-time Game Session Projection] 是从单个已提交视图生成的接收者特定完整状态。加入和命令提交在房间串行顺序内生成整批投影；网络发送不参与事务。

每个活动连接使用独立的有序出站队列 [Ordered Outbound Queue]。同一连接按入队顺序发送，不同连接互不阻塞。发送失败或慢消费者队列溢出时，服务端只关闭活动连接，房间成员资格保持；重连后通过完整状态恢复，不补发历史事件。

## 契约生成 [Contract Generation]

`proto/game.proto` 是客户端信封、服务端信封、稳定错误码和房间投影字段的规范来源 [Canonical Source]。WebSocket 继续使用 JSON 编码，不使用 ProtoBuf 二进制传输。

`pnpm proto:generate` 生成 Go 与 TypeScript 契约；`pnpm proto:check` 检查生成物漂移 [Generated Artifact Drift]。生成文件不得手动维护。
