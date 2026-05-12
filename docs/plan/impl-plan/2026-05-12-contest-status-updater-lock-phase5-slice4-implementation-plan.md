# Contest Status Updater Lock Phase 5 Slice 4 Implementation Plan

## Objective

继续 phase 5 收窄 `contest` application concrete allowlist，把 `StatusUpdater` 的调度锁 Redis 细节下沉到模块 port / infrastructure adapter：

- 去掉 `contest/application/jobs/status_updater.go` 对 `github.com/redis/go-redis/v9` 的直接依赖
- 让 `status_update_runner.go` 不再直接知道 Redis key、`redislock.Acquire(...)` 和具体 client 类型
- 保持“单实例调度 + 续租 + 丢锁即停 + 结束时释放锁”的现有行为不变

## Non-goals

- 不处理 `scoreboard_admin_service.go` / `scoreboard_admin_score_commands.go` 的排行榜写缓存 Redis 依赖
- 不改 `AWDRoundUpdater` 的调度锁实现
- 不调整竞赛状态迁移规则、回放逻辑或 side-effect replay 机制

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/jobs/status_updater.go`
- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Current Baseline

- `StatusUpdater` 构造函数当前直接接收 `*redis.Client`
- `status_update_runner.go` 直接调用 `redislock.Acquire(...)` 并持有 `rediskeys.ContestStatusUpdateLockKey()`
- `lock_keepalive.go` 直接依赖 `*redislock.Lock`
- allowlist 里仍保留：
  - `contest/application/jobs/status_updater.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把“竞赛状态调度锁”表达成 `contest` 自己的 job port：

1. 在 `contest/ports` 新增状态调度锁 store / lease interface
2. `StatusUpdater` 改为通过 setter 注入调度锁 store，默认无 store 时保持单机直跑语义
3. `status_update_runner.go` 只通过 port 获取 lease 并做 keepalive / release 编排
4. 在 `contest/infrastructure` 新增 Redis lock adapter，内部封装 `redislock.Acquire(...)` 和 `ContestStatusUpdateLockKey()`
5. `runtime` 统一构建调度锁 store，并注入 `StatusUpdater`

## Ownership Boundary

- `contest/application/jobs`
  - 负责：决定何时尝试持锁、续租失败后如何收敛、释放时机和上下文边界
  - 不负责：知道 Redis key、Redis client、`SetNX` / Lua release / refresh 细节
- `contest/infrastructure`
  - 负责：用 Redis / `redislock` 实现状态调度锁获取
  - 不负责：决定拿到锁后如何执行状态扫描
- `contest/runtime`
  - 负责：装配调度锁 store 并注入 job
  - 不负责：把 Redis 细节继续暴露回 application

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-contest-status-updater-lock-phase5-slice4-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-status-updater-lock-phase5-slice4.md`
- Add: `code/backend/internal/module/contest/infrastructure/status_update_lock_store.go`
- Add: `code/backend/internal/module/contest/ports/status_update_lock_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/contest.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_updater.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- Modify: `code/backend/internal/module/contest/application/jobs/lock_keepalive_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: 提取状态调度锁 port

目标：

- `StatusUpdater` 不再持有 Redis client
- application/jobs 改为依赖 lock store / lease interface
- runtime / tests 都改为注入调度锁 store

Validation:

- `cd code/backend && go test ./internal/module/contest/application/jobs -run 'StatusUpdater|RedisLockKeepalive' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus:

- application 层是否已经不再知道 Redis key / client / `redislock.Acquire(...)`
- 持锁、续租、丢锁取消和 release 语义是否与当前实现一致

### Slice 2: 删除 allowlist 并同步文档

目标：

- 删除 `status_updater.go` 的 Redis allowlist
- 在迁移设计稿和当前模块边界文档里补上这条 phase 5 进展

Validation:

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`

Review focus:

- 只删除本次真正收口的 allowlist
- 文档是否明确说明 AWD round scheduler 仍保留自己的 Redis 锁实现，没有误写成整条 jobs 链已经完全去 Redis

## Risks

- 如果 runtime 忘记给 `StatusUpdater` 注入 lock store，分布式锁会退化成单机直跑
- keepalive 如果没有沿用现有的“失锁即停”语义，会放宽单实例调度约束
- 测试如果只覆盖 side-effect store 而没覆盖 lock refresh / held-elsewhere 分支，会遗漏行为回归

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/jobs -run 'StatusUpdater|RedisLockKeepalive' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`

## Architecture-Fit Evaluation

- owner 明确：application/jobs 继续持有调度语义，Redis 锁细节落回 infrastructure
- reuse point 明确：复用 `contest` 现有 lock keepalive 行为、`redislock` 包和 runtime wiring
- 这刀同时解决行为与结构：保持单实例状态调度约束，同时删掉 application 层一条 concrete Redis import
