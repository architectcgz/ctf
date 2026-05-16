# Reuse Decision

## Change type
service / repository / config / job / script / doc

## Existing code searched
- `code/backend/internal/module/contest/application/commands/contest_update_support.go`
- `code/backend/internal/module/contest/application/commands/challenge_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/config/config.go`
- `code/backend/internal/pkg/redis/keys.go`
- `docker/ctf/docker-compose.dev.yml`
- `docs/architecture/backend/{03-container-architecture.md,05-key-flows.md}`

## Similar implementations found
- `challenge_support.go`
  - 已有“比赛进入 running / frozen / ended 后冻结题目配置”的 owner，可直接复用到 AWD service 配置入口
- `awd_desired_runtime_reconciler.go` + `instance_provisioning.go`
  - 已经把“缺失 scope 补齐”和“实例失败标记”拆成两个 owner，适合在两处之间补失败回退状态，而不是重写调和器
- `platform_runtime_state_store.go`
  - 已有 Redis 持久化运行态状态的实现模式，适合作为 desired reconcile 抑噪状态的轻量持久化参考
- `config.Load()`
  - 已经是容器运行配置、环境变量和默认值的唯一 owner，适合把 `flag_global_secret` 的自动持久化读写收口在这里

## Decision
extend_existing

## Reason
这次不是新增一套 AWD 控制面，而是把已有 AWD 运行链路补成更可长期运行的形态。比赛配置冻结应继续落在现有 contest command service 边界；desired reconcile 抑噪应建立在现有 reconciler / provisioning / Redis key 模式上；`flag_global_secret` 的自动持久化也应继续由配置加载 owner 承担。这样能最小化改动面，同时不给后续“人工 suppress / scope 停用 / 退赛”预埋错误 owner。

宿主重启回放里暴露出的后续问题也仍然属于同一条运行链路：AWD failed 实例的 host port 复用、失败清理的上下文语义，以及 workspace companion 的自动恢复，都应继续收口在现有 practice restart / runtime maintenance owner 下，而不是另起独立恢复器。

## Files to modify
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- `code/backend/internal/module/practice/application/commands/repository_stub_test.go`
- `code/backend/internal/module/practice/application/commands/service_lifecycle_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/infrastructure/repository.go`
- `code/backend/internal/module/practice/infrastructure/repository_test.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/instance/application/commands/maintenance_service.go`
- `code/backend/internal/module/runtime/infrastructure/repository.go`
- `code/backend/internal/module/runtime/application/commands/runtime_maintenance_service_test.go`
- `code/backend/internal/module/practice/infrastructure/desired_awd_reconcile_state_store.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/config/config.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/internal/pkg/redis/keys.go`
- `code/backend/configs/config.yaml`
- `code/backend/configs/config.prod.yaml`
- `docker/ctf/docker-compose.dev.yml`
- `docs/plan/impl-plan/2026-05-16-awd-runtime-hardening-implementation-plan.md`
- `docs/architecture/backend/03-container-architecture.md`
- `docs/architecture/backend/05-key-flows.md`
- `docs/operations/awd-host-reboot-recovery-drill.md`

## After implementation
- 如果“AWD 自动调和失败抑噪”后续扩展为人工 suppress / scope 停用的统一 owner，再把当前决策追加到 `harness/reuse/history.md`
