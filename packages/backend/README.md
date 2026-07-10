# Backend 部署 [Backend Deployment]

## 配置

- `CLOCKTOWER_CREDENTIAL_KEY`：所有生产启动模式必需，至少 32 字节，必须跨重启稳定；当前不支持轮换或恢复。
- `CLOCKTOWER_REDIS_URL`：启用 Redis 每房间持久化。
- `CLOCKTOWER_REDIS_KEY`：Redis 命名空间前缀，生成 `<prefix>:rooms` 与 `<prefix>:room:<roomId>`。
- `CLOCKTOWER_SNAPSHOT_PATH`：文件持久化目录，每房间一个 `<roomId>.json`；若指向旧普通文件则启动失败。

Redis 优先于文件配置；两者均未设置时使用内存存储，数据不跨重启。当前只支持单活动后端实例 [Single Active Backend Instance]。

## 启动

```powershell
$env:CLOCKTOWER_CREDENTIAL_KEY='replace-with-a-persistent-secret-of-at-least-32-bytes'
go run ./cmd/server
```

`GET /health` 正常返回 `200`；检测到房间修订 CAS 冲突后返回 `503`。旧全局 Snapshot 不自动迁移，升级前应保留原存储用于整体回滚。
