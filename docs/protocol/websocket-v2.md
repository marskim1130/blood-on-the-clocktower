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

## 房间请求顺序 [Room Command Flow]

同一房间的请求在成功解码并进入服务端接纳入口后按FIFO处理，包含加入、重连接管、已有建房请求的重试、成员/游戏命令和恢复审批；过期检查与关闭也参加该房间顺序。不同房间各自推进。连接读循环保留逐请求背压，尚未被读取和解码的网络数据不属于已接纳请求。

接管等待排在前面的操作完成，包括持久化失败；轮到接管时才读取最新完整已提交状态，因此状态、房间修订与下一个客户端序号对应同一时点。接管不插队、不消耗客户端序号；同一连接重放接管不会关闭自身或使已排队命令失效。

设备断线不取消已接纳命令。命令仍须通过执行时的凭据、权限、序号与游戏规则校验；旧连接在接管或身份撤销之后不能继续执行后续命令。发送失败只影响投递，不回滚成功提交。重连通过完整状态及下一个客户端序号确认结果，重发沿用原有幂等规则。

入队不是成功确认，队列也不持久化；服务重启只恢复已持久化的Room Record。命令响应和投影在同房处理顺序内入队，实际网络发送不阻塞下一条房间操作。

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
