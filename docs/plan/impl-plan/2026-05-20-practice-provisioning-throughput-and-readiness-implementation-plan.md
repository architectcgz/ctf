# practice provisioning throughput and readiness 实施计划

## Objective

缓解本地压测下 practice 实例创建吞吐过低的问题：先只提高 `dev` 环境的 provisioning 调度并发，并为 HTTP readiness probe 增加对裸 TCP/banner 服务的兼容回退，降低 `pending/creating` 积压和错误探活导致的失败。

## Non-goals

- 不修改生产默认配置或 `config.yaml` / `config.prod.yaml`
- 不重做 practice scheduler、runtime provisioning 或实例生命周期状态机
- 不修正题目元数据本身的协议错误；元数据修正留作后续独立任务

## Inputs

- `docs/plan/impl-plan/2026-05-20-runtime-instance-delete-cleanup-hardening-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-20-runtime-network-subnet-allocation-implementation-plan.md`
- `docs/plan/archive/impl-plan/2026-05-13-practice-instance-readiness-probe-phase5-slice17-implementation-plan.md`
- `.harness/reuse-decisions/practice-instance-readiness-probe-phase5-slice17.md`
- `code/backend/configs/config.yaml`
- `code/backend/configs/config.dev.yaml`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- `code/backend/internal/module/practice/infrastructure/instance_readiness_probe.go`

## Ownership evaluation

- `config.dev.yaml` 只负责本地开发环境的调度并发覆盖，不改变通用/生产默认值。
- `practice scheduler` 继续负责 `pending -> creating` 的限流推进和活跃实例数控制；不负责协议探测细节。
- `practice readiness probe` 继续负责 access URL 的协议级可达性检查；这轮只在 HTTP 路径下补兼容 TCP fallback。
- `instance provisioning` 继续负责创建链路编排、超时控制和失败落库；不把探活细节抬回 application。

## Task slices

1. 为 `instance_readiness_probe` 增加一条回归测试，覆盖 “HTTP probe 遇到 `malformed HTTP status code`，但 TCP 可连通” 的场景。
2. 在 probe 实现中为上述错误增加 TCP connect fallback，保持其他 HTTP/TCP 行为不变。
3. 在 `config.dev.yaml` 增加本地调度并发覆盖，提高 batch size、并发启动上限和活跃实例上限。
4. 运行最小 Go 测试验证 probe 与 practice 命令包。
5. 重新启动本地后端，串行执行 50 / 100 规模创建与清理压测，记录创建吞吐、失败数和删除收口情况。

## Data and compatibility impact

- 仅影响本地 `dev` 配置读取结果，不涉及数据库结构变更。
- HTTP readiness probe 在收到非标准 HTTP 响应头时，若目标 TCP 端口可连通，将改判为成功。
- 对真正要求 HTTP 业务已完全 ready 的服务，判定会比之前宽松；这是本地压测场景下的有意折中。

## Validation

- `go test ./internal/module/practice/infrastructure -count=1`
- `go test ./internal/module/practice/application/commands -count=1`
- `./scripts/dev-run.sh --infra --migrate --background`
- `curl http://127.0.0.1:8080/health`
- 50 / 100 规模实例创建与删除压测

## Review focus

- `max_concurrent_starts` 提升后，创建链路是否仍稳定受 `max_active_instances` 约束
- TCP fallback 是否只覆盖 `malformed HTTP status code`，没有吞掉其他 HTTP 失败
- readiness probe 宽松判定是否足以把已监听的 TCP/banner 服务从 `failed` 拉回 `running`
- 删除 worker 在更高创建并发下是否仍能稳定完成 cleanup

## Rollback

如果这刀引入误判或本地资源压力过大，可以先回退 `config.dev.yaml` 的 scheduler 覆盖，并移除 HTTP probe 的 TCP fallback，恢复原有严格探活行为。
