# AWD 宿主机重启恢复演练手册

> 状态：Current
> 事实源：`code/backend/internal/module/instance/application/commands/startup_runtime_recovery_service.go`、`code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`、`code/backend/internal/config/config.go`、`docker/ctf/docker-compose.dev.yml`
> 替代：无

## 定位

这份文档只说明单机 Docker 宿主下，真实“整机停机后 API + Docker 一起起来”的 AWD 恢复演练怎么做、怎么验收、要保留什么证据。

- 负责：给本地 compose 或同类单机部署一套可执行的宿主重启回放步骤。
- 不负责：把单测、组合测试或代码阅读说成已经完成了真实宿主重启演练；也不覆盖多机调度、外部 KMS、退赛/人工 suppress 控制面。

## 当前结论

截至 2026-05-16，仓库里已经具备这些自动恢复能力：

- API 启动后会先竞争 Redis leader lock；只有 leader 会读取 Redis `platform_runtime_state`。真正触发启动恢复的条件是 boot ID 变化；同 `boot_id` 下的 heartbeat gap 只表示 leaderless gap。非 leader 副本会等 leader 写入当前 `boot_id` heartbeat 后再继续启动。
- 检测到 outage 后，会先给活跃 AWD 比赛补 `paused_seconds`，刷新活跃实例 `expires_at`，再执行 active runtime recovery 和 desired runtime reconciliation。
- desired reconcile 对长期坏配置会在 scope 级做 backoff / suppress，不再每个 `desired_reconcile_interval` 固定重试。
- `container.flag_global_secret_file` 会在启动时恢复或自动生成 AWD dynamic flag 全局密钥，只要文件位于持久化卷中，宿主重启后不会丢。

当前还没有随仓库提交一份“已实际执行过这次真实宿主重启”的证据记录；这份文档是 runbook，不是完成证明。

## 前置条件

演练前至少确认下面几项：

- 使用的部署确实把 PostgreSQL、Redis 数据目录和 `/app/storage` 放在持久化卷上。
- `CTF_CONTAINER_FLAG_GLOBAL_SECRET_FILE` 或 `container.flag_global_secret_file` 指向持久化路径。compose dev 默认是 `/app/storage/runtime/flag-global-secret`。
- 演练目标中至少有一场 `running` 或 `frozen` 的 AWD 比赛，并且至少存在一组可见 service。
- 最好提前准备一条“正常 scope”和一条“坏配置 scope”，方便同时观察恢复与 suppress 行为。
- 选定本次观测的 `<contest_id>`，并记住至少一组 `<team_id> + <service_id>`。

## 演练前基线

在仓库根目录执行以下检查，并把输出保存到演练记录里。

1. 记录容器与 API 健康状态。

```bash
docker compose -f docker/ctf/docker-compose.dev.yml ps
curl -fsS http://127.0.0.1:8080/health
```

2. 记录 runtime heartbeat、比赛暂停字段和实例状态。

```bash
docker exec ctf-redis redis-cli -a redis123456 HGETALL ctf:platform:runtime:state
docker exec ctf-postgres psql -U postgres -d ctf -c "
SELECT id, status, paused_seconds, runtime_recovery_key, runtime_recovery_applied_seconds
FROM contests
WHERE id = <contest_id>;
"
docker exec ctf-postgres psql -U postgres -d ctf -c "
SELECT id, status, contest_id, team_id, service_id, container_id, network_id, expires_at
FROM instances
WHERE contest_id = <contest_id>
ORDER BY team_id, service_id, id;
"
```

3. 记录 AWD global secret 文件，确认它已经落在持久化卷里。

```bash
docker exec ctf-api sh -lc 'ls -l /app/storage/runtime/flag-global-secret && cat /app/storage/runtime/flag-global-secret'
```

4. 如果要观察 suppress 行为，再记录目标 scope 的 Redis 状态和最近 operation。

```bash
docker exec ctf-redis redis-cli -a redis123456 HGETALL ctf:awd:desired_reconcile:state:<contest_id>:<team_id>:<service_id>
docker exec ctf-postgres psql -U postgres -d ctf -c "
SELECT id, team_id, service_id, operation_type, status, reason, started_at, finished_at
FROM awd_service_operations
WHERE contest_id = <contest_id>
ORDER BY id DESC
LIMIT 20;
"
```

## 演练步骤

1. 确认当前服务都在运行，并且前面的基线已经留存。
2. 触发真实宿主机重启。

```bash
sudo systemctl reboot
```

如果是在云主机或虚拟机里演练，用平台控制台执行同等语义的整机重启，不要只重启 `ctf-api` 容器。

3. 宿主机起来后，等待 Docker、PostgreSQL、Redis 和 API 全部恢复。

```bash
docker compose -f docker/ctf/docker-compose.dev.yml up -d
docker compose -f docker/ctf/docker-compose.dev.yml ps
curl -fsS http://127.0.0.1:8080/health
```

4. 抓取 API 恢复日志，确认启动恢复链路真的执行过。

```bash
docker logs --since 10m ctf-api | rg 'runtime_outage_detected_for_startup_recovery|reconcile_desired_awd|save_platform_runtime_heartbeat'
```

5. 重复“演练前基线”里的 Redis / PostgreSQL / secret 文件检查。

## 验收点

以下结果同时满足，才算这次真实演练通过：

- `ctf:platform:runtime:state` 的 `boot_id` 或 `last_heartbeat_at` 明显更新。
- `contests.paused_seconds` 比重启前增加，`runtime_recovery_key` 和 `runtime_recovery_applied_seconds` 有对应变化。
- 重启前仍应活跃的 AWD scope，最终要么已经回到 `running`，要么先进入 `pending / creating` 后被调度器拉起。
- `awd_service_operations` 不应出现同一坏配置 scope 每 15 秒稳定新增一条自动 start/recreate 噪声；如果触发了 suppress，Redis `suppressed_until` 应晚于当前时间。
- `/app/storage/runtime/flag-global-secret` 在重启后仍存在且内容不变；如果内容变化，说明持久化卷或 secret 注入策略不成立。

## 失败判读

- `paused_seconds` 没变化：启动恢复服务没有识别到 outage，优先检查 Redis `platform_runtime_state` 是否可写、boot ID 读取是否正常。
- 实例长期停在 `pending`：先查 `ctf-api` 日志里的 provisioning 错误，再查 Docker daemon 是否可用。
- 同一坏配置 scope 仍然每轮都新增 operation：优先检查 Redis 是否可写，以及 `ctf:awd:desired_reconcile:state:<contest_id>:<team_id>:<service_id>` 是否根本没生成。
- global secret 文件丢失或变化：说明 `/app/storage` 没有持久化，或者部署层在每次启动时都覆盖了 `CTF_CONTAINER_FLAG_GLOBAL_SECRET`。

## 证据保留

建议把下面这些内容一起存档到演练记录：

- 重启前后的 `docker compose ps`
- 重启前后的 `platform_runtime_state`
- 重启前后的 `contests.paused_seconds / runtime_recovery_*`
- 目标 scope 的 `instances` 和 `awd_service_operations`
- `ctf-api` 恢复日志摘录
- `flag-global-secret` 文件路径与内容校验结果

## 已知限制

- 这份手册只覆盖真实宿主重启恢复，不替代退赛、停用队伍服务或人工 suppress 某些 scope 的后续设计。
- 目前仓库里没有自动生成这份演练报告的脚本，证据收集仍以人工执行为主。
