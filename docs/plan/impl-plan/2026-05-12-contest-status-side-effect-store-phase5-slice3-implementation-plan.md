# Contest Status Side Effect Store Phase 5 Slice 3 Implementation Plan

## Objective

继续 phase 5 收窄 `contest` application concrete allowlist，把竞赛状态迁移副作用里的 Redis 细节下沉到模块 port / infrastructure adapter：

- 去掉 `contest/application/statusmachine/side_effects.go` 对 `github.com/redis/go-redis/v9` 的直接依赖
- 去掉 `contest/application/commands/contest_service.go` 为构造副作用 runner 持有的 Redis 依赖
- 保持冻结榜快照创建、解冻快照清理、比赛结束时 AWD 运行态缓存清理的现有行为不变

## Non-goals

- 不处理 `contest/application/jobs/status_updater.go` 为分布式锁保留的 Redis 依赖
- 不收口 `scoreboard_admin_service.go` / `scoreboard_admin_score_commands.go` 里的排行榜写缓存 Redis 依赖
- 不改竞赛状态机规则、状态迁移记录表结构或 HTTP 接口

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/statusmachine/side_effects.go`
- `code/backend/internal/module/contest/application/commands/contest_service.go`
- `code/backend/internal/module/contest/application/jobs/status_updater.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Current Baseline

- `statusmachine.SideEffectRunner` 当前直接持有 `*redis.Client`
- `ContestService` 构造时接收 Redis client，只为了创建 `SideEffectRunner`
- `StatusUpdater` 既用 Redis 做 scheduler lock，也用同一个 client 创建 `SideEffectRunner`
- `ScoreboardAdminService` 也通过同样方式创建 `SideEffectRunner`
- allowlist 里仍保留：
  - `contest/application/statusmachine/side_effects.go -> github.com/redis/go-redis/v9`
  - `contest/application/commands/contest_service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把“状态迁移副作用落到缓存”的具体执行能力表达成 `contest` 自己的 port：

1. 在 `contest/ports` 新增状态迁移副作用 store interface
2. `statusmachine.SideEffectRunner` 只依赖这个 port，保留状态判断与编排逻辑
3. 在 `contest/infrastructure` 新增 Redis adapter，实现 frozen snapshot / runtime cache cleanup
4. `ContestService` 改为通过 setter 接收 side-effect store，不再在构造函数里接 Redis client
5. `runtime` 统一构建 side-effect store，并注入 `ContestService`、`ScoreboardAdminService`、`StatusUpdater`

## Ownership Boundary

- `contest/application/statusmachine`
  - 负责：判断什么状态迁移需要什么副作用，以及执行顺序
  - 不负责：知道 Redis key、Redis client 或序列化细节
- `contest/infrastructure`
  - 负责：实现 frozen snapshot 与 runtime state cleanup 的 Redis 读写
  - 不负责：决定何时触发这些副作用
- `contest/runtime`
  - 负责：装配 side-effect store，并把它注入需要的 application service / job
  - 不负责：把 Redis 细节继续暴露回 application 层

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-contest-status-side-effect-store-phase5-slice3-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-status-side-effect-store-phase5-slice3.md`
- Add: `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- Add: `code/backend/internal/module/contest/ports/status_side_effect_store_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/contest.go`
- Modify: `code/backend/internal/module/contest/application/statusmachine/side_effects.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/scoreboard_admin_service.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_updater.go`
- Modify: `code/backend/internal/module/contest/application/commands/contest_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/scoreboard_admin_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_service_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: 提取状态迁移副作用 store port

目标：

- `SideEffectRunner` 不再依赖 Redis client
- infrastructure 有 Redis adapter 承接 snapshot / cleanup
- runtime / tests 都改为注入 side-effect store

Validation:

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'ContestService|ScoreboardAdminService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/application/jobs -run 'StatusUpdater' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus:

- application 层是否已经不再知道 Redis key / client
- setter 注入是否只改变 wiring，没有改变 freeze / unfreeze / ended 的副作用语义

### Slice 2: 删除 allowlist 并同步文档

目标：

- 删除 `statusmachine/side_effects.go` 与 `contest_service.go` 的 Redis allowlist
- 在迁移设计稿和当前模块边界文档里补上这条 phase 5 进展

Validation:

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 只删除本次真正收口的 allowlist
- 文档是否明确说明 `StatusUpdater` 仍保留锁 Redis 依赖，没有误写成整条链完全去 Redis

## Risks

- 如果 runtime 忘记给某条路径注入 side-effect store，freeze / ended 的缓存副作用会静默丢失
- `StatusUpdater` 同时依赖锁 Redis 和 side-effect store，构造调整时容易把两类 Redis 用途混淆
- replay failed side effects 的测试如果没同步更新，会掩盖回放链路回归

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/commands -run 'ContestService|ScoreboardAdminService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/application/jobs -run 'StatusUpdater' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：状态迁移副作用编排留在 application，Redis 细节落回 infrastructure
- reuse point 明确：复用 `contest` 现有 runtime wiring、status transition replay 机制和既有 Redis key
- 这刀同时解决行为与结构：保留 frozen snapshot / runtime cleanup 语义，同时删掉 application 层两条 concrete Redis import
