# Contest Scoreboard State Store Phase 5 Slice 6 Implementation Plan

## Objective

继续 phase 5 收窄 `contest` application concrete allowlist，把 scoreboard 的 Redis 状态面下沉到模块 port / infrastructure adapter：

- 去掉 `contest/application/queries/scoreboard_service.go`、`scoreboard_support.go`、`scoreboard_rank_query.go`、`scoreboard_list_support.go` 对 `github.com/redis/go-redis/v9` 的直接依赖
- 去掉 `contest/application/commands/scoreboard_admin_service.go`、`scoreboard_admin_score_commands.go` 对 `github.com/redis/go-redis/v9` 的直接依赖
- 保持 live scoreboard、frozen snapshot、队伍分数增量、排行榜重建和 rank 查询的现有行为不变

## Non-goals

- 不处理 `contest/application/commands/awd_service.go`、`contest/application/commands/contest_awd_service_service.go`、`contest/application/commands/awd_status_cache.go` 的 Redis 依赖
- 不改 submission 触发 scoreboard 更新的业务语义
- 不重写 scoreboard HTTP contract、仓储 SQL 或 frozen/running 状态迁移规则

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/application/queries/scoreboard_*.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_*.go`
- `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- `code/backend/internal/module/contest/infrastructure/awd_scoreboard_cache.go`
- `code/backend/internal/module/contest/runtime/module.go`

## Current Baseline

- `ScoreboardService` 当前直接持有 `*redis.Client`
- frozen scoreboard snapshot 的存在判断、从 live 复制 snapshot、全量榜单读取和 team rank 查询都散落在 query 层直接操作 Redis
- `ScoreboardAdminService` 当前直接持有 `*redis.Client`
- live scoreboard 的 `ZIncrBy` 与全量 `TxPipeline + ZAdd` rebuild 都散落在 command 层直接操作 Redis
- `ContestStatusSideEffectStore` 已经承接了 freeze/unfreeze 时的 frozen snapshot create/clear，但 scoreboard query/admin 仍绕过它直接操作相同 key
- allowlist 当前至少保留：
  - `contest/application/queries/scoreboard_list_support.go -> github.com/redis/go-redis/v9`
  - `contest/application/queries/scoreboard_rank_query.go -> github.com/redis/go-redis/v9`
  - `contest/application/queries/scoreboard_service.go -> github.com/redis/go-redis/v9`
  - `contest/application/queries/scoreboard_support.go -> github.com/redis/go-redis/v9`
  - `contest/application/commands/scoreboard_admin_service.go -> github.com/redis/go-redis/v9`
  - `contest/application/commands/scoreboard_admin_score_commands.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把 scoreboard live/frozen 状态面表达成 `contest` 自己的 scoreboard state store：

1. 在 `contest/ports` 新增 scoreboard state store 和 scoreboard entry/rank value object
2. `ScoreboardService` 只依赖 scoreboard store，保留“何时看 frozen、何时补 snapshot、何时过滤非法 member / 缺失 team”的查询编排语义
3. `ScoreboardAdminService` 只依赖 scoreboard store，保留“何时增量改分、何时根据仓储结果重建 live scoreboard”的命令编排语义
4. 在 `contest/infrastructure` 新增 Redis adapter，内部封装 sorted-set key、member 编解码、snapshot copy、rank 读取和 rebuild pipeline
5. `ContestStatusSideEffectStore` 复用同一组 infrastructure helper，不再自己保留第二套 frozen snapshot Redis 细节
6. `runtime/module.go` 统一构建 scoreboard state store，并注入 query / command / side-effect wiring

## Ownership Boundary

- `contest/application/queries`
  - 负责：根据竞赛状态决定 live/frozen 读取策略、过滤非法 member、过滤缺失 team、拼装 scoreboard result
  - 不负责：知道 Redis key、`redis.Nil`、`redis.Z`、pipeline 或 sorted-set copy 细节
- `contest/application/commands`
  - 负责：决定何时增量改分、何时按数据库结果重建 live scoreboard
  - 不负责：知道 Redis key、`ZIncrBy`、`TxPipeline` 或 `redis.Z` 细节
- `contest/infrastructure`
  - 负责：实现 scoreboard state store，处理 Redis key、member 编解码、snapshot create/clear、rank/list/write 语义
  - 不负责：决定 frozen 读取策略或分数更新业务规则
- `contest/runtime`
  - 负责：装配 scoreboard state store，并注入 scoreboard query / command / side-effect owner
  - 不负责：把 Redis client 继续暴露回 application

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-13-contest-scoreboard-state-store-phase5-slice6-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-scoreboard-state-store-phase5-slice6.md`
- Add: `code/backend/internal/module/contest/infrastructure/scoreboard_state_store.go`
- Add: `code/backend/internal/module/contest/ports/scoreboard_state_context_contract_test.go`
- Modify: `code/backend/internal/module/contest/ports/contest.go`
- Modify: `code/backend/internal/module/contest/infrastructure/status_side_effect_store.go`
- Modify: `code/backend/internal/module/contest/application/queries/scoreboard_service.go`
- Modify: `code/backend/internal/module/contest/application/queries/scoreboard_support.go`
- Modify: `code/backend/internal/module/contest/application/queries/scoreboard_rank_query.go`
- Modify: `code/backend/internal/module/contest/application/queries/scoreboard_list_support.go`
- Modify: `code/backend/internal/module/contest/application/queries/scoreboard_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/scoreboard_admin_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/scoreboard_admin_score_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/scoreboard_admin_service_test.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

### Slice 1: 提取 scoreboard state store port 与 Redis adapter

目标：

- scoreboard query / admin 不再持有 `*redis.Client`
- application/query 不再直接依赖 `redis.Nil`、`redis.Z`、`ZRevRangeWithScores`、`Exists`、`ZUnionStore`
- application/command 不再直接依赖 `ZIncrBy`、`TxPipeline`、`ZAdd`
- runtime / tests 都改为注入 scoreboard state store

Validation:

- `cd code/backend && go test ./internal/module/contest/application/queries -run 'ScoreboardService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/application/commands -run 'ScoreboardAdminService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`

Review focus:

- scoreboard query/admin 是否已经不再知道 Redis key / client / `redis.Z` / `redis.Nil`
- frozen snapshot copy、list、rank、rebuild 和增量改分语义是否保持一致

### Slice 2: 删除 allowlist 并同步文档

目标：

- 删除本次真正收口的 scoreboard Redis allowlist
- 在迁移设计稿和当前模块边界文档里补上 phase 5 当前进展

Validation:

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 只删除本次 scoreboard query/admin 实际收口的 allowlist
- 文档是否准确描述“scoreboard live/frozen 状态面已下沉”，而不是误写成 `contest` 所有 Redis 依赖都已消失

## Risks

- 如果 frozen snapshot copy 语义和现有 `CreateFrozenScoreboardSnapshot` 不一致，冻结榜单可能回归
- 如果 list 读取阶段丢掉非法 member 过滤或缺失 team 过滤，现有 scoreboard 容错行为会回归
- 如果 rebuild live scoreboard 时没有保留“先清空再重建”的语义，会留下脏旧成员
- 如果 runtime 继续把旧 Redis client 注回 application，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/queries -run 'ScoreboardService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/contest/application/commands -run 'ScoreboardAdminService' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/contest/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：scoreboard query/admin 继续持有业务编排语义，Redis scoreboard 状态细节落回 infrastructure
- reuse point 明确：复用现有 `contest/domain/scoreboard.go` member 编解码、`status_side_effect_store.go` frozen snapshot 语义和 runtime wiring
- 这刀同时解决行为与结构：保留 frozen/live scoreboard 现有行为，同时删掉 scoreboard query/admin 的 concrete Redis import
