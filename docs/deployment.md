# 私人服务器部署与回滚

本配置运行一个 Go 后端、一个 Redis 和一个提供 H5 的 Caddy，面向已确定的私人使用。后端仅支持单活动实例，Redis 不提供跨后端协调；不要扩容 backend。公网只开放 80/443，Redis 与 8080 不映射到宿主机。

## 首次部署（Linux 服务器）

准备 Docker Engine 与 Compose v2，将域名 A/AAAA 指向服务器，确认防火墙允许 80/TCP、443/TCP（443/UDP 可选），且端口未被其他网站占用。在仓库根目录执行：

```sh
cp .env.example .env
chmod 600 .env
openssl rand -hex 32
```

用编辑器填写 `.env` 的 `CLOCKTOWER_DOMAIN`、`CLOCKTOWER_CREDENTIAL_KEY` 和本次 `CLOCKTOWER_RELEASE`（例如 `20260907-1`）。密钥使用上面生成的值，保持稳定并离线备份。不要把密钥放入构建参数、仓库或截图。Compose 强制使用内部 Redis 地址；`.env.example` 的 localhost Redis 地址仅供不使用 Compose 的本地开发。

```sh
docker compose config --quiet
docker compose build --pull
docker compose run --rm --no-deps web caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
docker compose up -d
docker compose ps
docker compose logs --tail=80 backend web
```

Caddy 使用域名自动申请 HTTPS 证书，并将 `/ws` 转发至后端；H5 构建时写入同域 `wss://域名/ws`。变更域名必须重新构建 web。不要在工单中贴完整的 `docker compose config` 输出，因为其中包含密钥。Redis 启用 AOF、每次写入同步刷盘，使用独立持久卷；磁盘仍需可靠存储和备份。

## 验收

```sh
curl --fail https://你的域名/health
docker compose exec redis redis-cli ping
```

预期分别为 `ok` 和 `PONG`。用至少两部手机打开 HTTPS 页面，实际创建房间、邀请加入、设置说书人、准备、分配角色并确认身份。检查每个手机仅显示自己的私密信息；继续验证夜间行动、投票和结束页。游戏中刷新页面以及 `docker compose restart backend` 后，确认原手机可恢复房间与身份。`/health` 仅证明后端基本状态，不能代替游戏流程验收。

微信内置浏览器可直接打开 H5 链接。微信小程序是独立交付：在本机构建时设置 `TARO_APP_WS_URL=wss://你的域名/ws`，执行 `pnpm --filter @clocktower/core build` 和 `pnpm --filter @clocktower/frontend build`，用微信开发者工具导入前端项目并配置自己的 AppID 与 socket 合法域名。正式发布仍需在微信后台完成相应配置与审核。

## 升级、备份与回滚

先通知玩家暂停操作，记录当前版本标签与代码提交号，并在仓库外的受保护位置备份 `.env`。停止写入并停止 Redis 后，备份整个 Redis 数据目录（AOF 包含多个文件，不能只复制 dump.rdb）：

```sh
mkdir -p backups
chmod 700 backups
docker compose stop web backend redis
docker run --rm -v clocktower_redis_data:/data:ro -v "$PWD/backups:/backup" alpine:3.22 tar -czf /backup/redis-before-upgrade.tgz -C /data .
docker compose start redis backend web
```

每次备份使用独立文件名，避免覆盖上一份。备份中包含私密角色和身份凭据，限制读取权限。Caddy 的证书保存在 `clocktower_caddy_data` 卷，保留该卷即可避免不必要的重新申请。

把 `.env` 中版本标签改为新值，构建新镜像，执行 `docker compose up -d` 并重复验收。旧镜像在确认稳定前不要清理。如果仅需回滚应用且数据结构兼容，把标签改回旧值并执行：

```sh
docker compose up -d --no-build --pull never
```

如果版本间数据结构不兼容，应先停服，保留当前卷作为回退副本，将升级前完整 Redis 备份恢复至一个新卷，再通过 Compose override 把 `redis_data` 指向该新卷，以旧版代码、镜像与原密钥启动。切勿直接覆盖唯一的当前数据卷。恢复备份会丢失备份之后的游戏操作，需在操作前明确恢复时间点。

停止服务使用 `docker compose down`，该命令默认保留卷。不要使用 `down -v`，它会删除房间和证书持久卷。密钥更换会使已有身份恢复失败，不能当作普通更新处理。

## 验证范围与依据

当前开发环境没有 Docker 命令，不能在此宣称镜像构建、容器启动、证书签发或公网手机验收已完成。以上命令应在服务器执行；SSH 尚未提供，因此未修改公网服务。

配置依据：[Caddy WebSocket 反代](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy)、[Compose 健康依赖启动顺序](https://docs.docker.com/compose/how-tos/startup-order/)、[Redis AOF 持久化](https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/)。
