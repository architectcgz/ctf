# AWD Runtime Hardening Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 AWD runtime hardening 相关 diff，重点覆盖 `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`、`code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`、`code/backend/internal/module/practice/infrastructure/desired_awd_reconcile_state_store.go`、`code/backend/internal/config/config.go` 及对应测试 / 文档接线
- Files reviewed:
  - `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
  - `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
  - `code/backend/internal/module/contest/application/commands/awd_service_test.go`
  - `code/backend/internal/module/practice/ports/ports.go`
  - `code/backend/internal/module/practice/application/commands/service.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
  - `code/backend/internal/module/practice/application/commands/instance_provisioning_scheduler.go`
  - `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - `code/backend/internal/module/practice/infrastructure/desired_awd_reconcile_state_store.go`
  - `code/backend/internal/module/practice/runtime/module.go`
  - `code/backend/internal/pkg/redis/keys.go`
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/configs/config.yaml`
  - `code/backend/configs/config.prod.yaml`
  - `code/backend/Dockerfile`
  - `docker/ctf/docker-compose.dev.yml`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Classification check: agree with pipeline，属于 non-trivial backend implementation + review gate
- Gate verdict: blocked

## Findings

1. `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler.go:150` + `code/backend/internal/module/practice/infrastructure/desired_awd_reconcile_state_store.go:41`
   - material / blocker
   - 新增的 desired reconcile state 只要 Redis 读取失败，或 hash 里任一字段解析失败，就会把错误直接返回给 `ReconcileDesiredAWDInstances()`。这不只是让后台循环打 warn，因为宿主重启恢复链路也会同步调用它；`startup_runtime_recovery_service.go:212` 会把这个错误继续往上抛，而 `http_server.go:143` 会因此让后台任务启动失败并阻断 API 启动。
   - 结果是：一次 Redis 短暂抖动，或者一个损坏的 `ctf:awd:desired_reconcile:state:*` hash，就可能把“宿主重启后自动恢复”降级成“API 起不来”。这和本次改动的目标相反。
   - 修复方向：desired reconcile state 应该是 best-effort 降噪层，不该成为 startup recovery 的 hard dependency。至少要把读取 / 解析失败降级成“忽略 backoff 状态并继续 reconcile”，同时记录 warn。

2. `code/backend/internal/pkg/redis/keys.go:164` + `code/backend/internal/module/practice/infrastructure/desired_awd_reconcile_state_store.go:74`
   - medium
   - 新增的 scope 状态明确是“TTL: 无过期”，而当前实现也没有在 contest `ended` side effect、service 删除或其他生命周期出口清理这些 key。恢复 active 时会删，但“比赛已经结束”“scope 永远坏掉”“服务不再参与 reconcile”这几类路径不会自动收口。
   - 结果是 Redis 会长期积累 `ctf:awd:desired_reconcile:state:<contest>:<team>:<service>` 脏 key。它不会立刻打坏功能，但这是永久状态泄漏，不是短期缓存。
   - 修复方向：要么给这类 key 明确 TTL，要么在 contest ended/runtime cleanup 路径补统一删除。

## Material Findings

- Finding 1：把 desired reconcile state 读取 / 解析失败从 startup blocker 改成 best-effort 降级，并补测试证明 Redis 故障或坏数据不会阻断 API 启动恢复。

## Senior Implementation Assessment

- AWD service 冻结放在 `ContestAWDServiceService.ensureMutableAWDContest()` 是合理的 owner 位置，没有把状态机约束散到 handler 或 repository。
- desired reconcile 的 Redis state 放在 `practice/ports` + `practice/infrastructure`，分层本身也是对的；问题不在“接到哪里”，而在错误语义太强，导致一个降噪层反过来卡住启动恢复主链路。
- `container.flag_global_secret_file` 的能力本身可行，但当前把“生成 / 持久化 runtime secret”放进 `config.Load()`，让 config 读取从纯解析变成了带副作用的运行态初始化。它暂时没形成明确 blocker，但后续如果还要扩 secret 生命周期，建议迁回 bootstrap / runtime 初始化层，避免所有 CLI 共享这个副作用面。

## Required Re-validation

- `cd code/backend && go test ./internal/module/practice/application/commands -run 'TestReconcileDesiredAWDInstances' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/instance/application/commands -run 'TestStartupRuntimeRecoveryService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/app/composition -count=1 -timeout 5m`
- `bash scripts/check-architecture.sh --quick`

## Residual Risk

- 本次 review 没有执行真实宿主重启演练；运行恢复相关判断仍主要依赖 package tests、composition guardrail 和代码路径复核。
- `docker/ctf/docker-compose.dev.yml` 当前还有明显的非本任务 diff，review 只按 AWD runtime hardening 相关 hunk 判断，没有把同文件里其他工作区改动一并当作本次结论。

## Touched Known-Debt Status

- 本次 touched surface 未命中当前 fact source 中要求“只要触达就必须顺手拆掉”的已知结构债；本轮 blocker 来自新增错误语义与状态生命周期，而不是旧债未收口。
