# Contest AWD Round State Store Phase 5 Slice 5 Implementation Plan

## Objective

继续 phase 5 收窄 `contest` application concrete allowlist，把 `AWDRoundUpdater` 里的 Redis 状态面下沉到模块 port / infrastructure adapter：

- 去掉 `contest/application/jobs/awd_round_updater.go` 对 `github.com/redis/go-redis/v9` 的直接依赖
- 去掉 `awd_round_scheduler_runtime.go`、`awd_round_lock.go`、`awd_round_flag_lookup_support.go`、`awd_round_flag_sync.go`、`awd_check_cache_support.go`、`awd_check_writeback.go` 里对 Redis key、`redis.Nil`、pipeline 和 `SetNX` 的直接依赖
- 保持 AWD 轮次调度锁、轮次锁、当前轮次指针、轮次 flag 缓存和服务状态缓存的现有行为不变

## Non-goals

- 不处理 `contest/application/commands/awd_service.go`、`contest/application/queries/scoreboard_service.go` 等其他 Redis 依赖
- 不改 AWD 轮次生成规则、flag 生成算法、checker runner、scoreboard cache rebuild 逻辑或 HTTP 接口
- 不把 `AWDRoundUpdater` 里仍需的数据库查询能力一起改造成新的 repository hierarchy

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/jobs/*.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/infrastructure/*.go`
- `code/backend/internal/pkg/redislock/lock.go`

## Current Baseline

- `AWDRoundUpdater` 当前直接持有 `*redis.Client`
- scheduler lock、round lock、当前轮次指针、round flag 缓存和 live service status cache 都散落在 `application/jobs` 多个文件里直接操作 Redis
- allowlist 当前至少保留：
  - `contest/application/jobs/awd_check_cache_support.go -> github.com/redis/go-redis/v9`
  - `contest/application/jobs/awd_round_flag_lookup_support.go -> github.com/redis/go-redis/v9`
  - `contest/application/jobs/awd_round_updater.go -> github.com/redis/go-redis/v9`
- `AWDRoundUpdater` 还需要保留 scoreboard cache writer 和数据库 repository 注入，但不需要继续知道 Redis key / pipeline / `redislock.Acquire(...)`

## Chosen Direction

把 AWD 轮次调度与运行态缓存表达成 `contest` 自己的 job port：

1. 在 `contest/ports` 新增通用 scheduler lock lease，以及 `AWDRoundStateStore`
2. `AWDRoundUpdater` 只依赖 `AWDRoundStateStore`，保留“什么时候持锁、什么时候刷当前轮次、什么时候更新 live status cache”的编排语义
3. 在 `contest/infrastructure` 新增 Redis adapter，内部封装 AWD scheduler lock、round lock、current round、round flags、service status cache 的 key 和 Redis miss 语义
4. `runtime/module.go` 统一构建 AWD round state store，并注入 `AWDRoundUpdater`
5. 删除本次真正收口的 Redis allowlist，并同步当前架构事实

## Ownership Boundary

- `contest/application/jobs`
  - 负责：决定何时尝试调度、何时申请 round lock、何时切换当前轮次、何时判断 live status cache 是否应覆盖
  - 不负责：知道 Redis key、`redis.Nil`、`TxPipeline`、`SetNX` 或 `redislock.Acquire(...)`
- `contest/infrastructure`
  - 负责：实现 AWD round state store，处理 Redis key、序列化、锁 token 和 miss 语义
  - 不负责：决定轮次调度时机或业务状态转换规则
- `contest/runtime`
  - 负责：把进程级 Redis client 装配成 AWD round state store，并注入 `AWDRoundUpdater`
  - 不负责：把 Redis client 继续暴露回 application/jobs

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-13-contest-awd-round-state-store-phase5-slice5-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-awd-round-state-store-phase5-slice5.md`
- Add: `code/backend/internal/module/contest/infrastructure/awd_round_state_store.go`
- Add: `code/backend/internal/module/contest/ports/awd_round_state_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/contest.go`
- Modify: `code/backend/internal/module/contest/ports/status_update_lock_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_lock.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_flag_sync.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_check_cache_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_check_sync.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_check_writeback.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime_internal_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_testsupport_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: 提取 AWD round state store port 与 Redis adapter

目标：

- `AWDRoundUpdater` 不再持有 `*redis.Client`
- application/jobs 改为通过 `AWDRoundStateStore` 访问 scheduler lock、round lock、current round、round flags 和 live service status cache
- runtime / tests 都改为注入 state store

Validation:

- `cd code/backend && go test ./internal/module/contest/application/jobs -run 'AWDRoundUpdater|AWDRoundUpdaterRefreshesSchedulerLockWhileRunning' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/application/commands -run 'AWDService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus:

- application/jobs 是否已经不再知道 Redis key / client / `redis.Nil` / `SetNX`
- scheduler lock、round lock、current round 和 service status cache 语义是否保持一致

### Slice 2: 删除 allowlist 并同步文档

目标：

- 删除本次真正收口的 AWD round updater Redis allowlist
- 在迁移设计稿和当前模块边界文档里补上 phase 5 当前进展

Validation:

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `timeout 600 bash scripts/check-workflow-complete.sh`

Review focus:

- 只删除本次实际收口的 allowlist
- 文档是否准确描述“AWD round updater 的 Redis 状态面已下沉”，而不是误写成 contest 所有 Redis 依赖都已消失

## Risks

- 如果 state store 没有把 `redis.Nil` 和“不存在”分支表达清楚，会改变 fallback 到 flagSecret 的行为
- 如果 scheduler lock keepalive 没有沿用现有“失锁即停”语义，会放宽单实例调度约束
- 如果 live service status cache 的更新语义变成追加而不是覆盖，会留下旧轮次状态残留

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/jobs -run 'AWDRoundUpdater|AWDRoundUpdaterRefreshesSchedulerLockWhileRunning' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/application/commands -run 'AWDService' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：AWD 轮次调度和缓存覆盖时机继续留在 application/jobs，Redis 细节落回 infrastructure
- reuse point 明确：复用已有 `redislock`、contest runtime wiring 和 status updater 已经落地的 keepalive 模式
- 这刀同时解决行为与结构：保留 AWD round updater 的运行语义，同时删掉 application/jobs 多个 Redis concrete import
