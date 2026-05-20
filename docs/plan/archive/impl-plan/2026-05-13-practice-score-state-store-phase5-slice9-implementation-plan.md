# Practice Score State Store Phase 5 Slice 9 Implementation Plan

## Objective

继续 phase 5 收窄 `practice` application concrete allowlist，把用户得分锁、用户得分缓存和排行榜 Redis 状态面下沉到模块 port / infrastructure adapter：

- 去掉 `practice/application/commands/score_service.go -> github.com/redis/go-redis/v9`
- 去掉 `practice/application/queries/score_service.go -> github.com/redis/go-redis/v9`
- 保持用户计分锁、用户得分缓存和排行榜同步的现有行为不变

## Non-goals

- 不处理 `practice/application/commands/service.go -> github.com/redis/go-redis/v9` 这条 flag submit 限流依赖
- 不改 `practice/application/queries/score_service.go -> gorm.io/gorm` 的 record-not-found 判定
- 不改 practice 分数计算规则、HTTP contract 或 SQL 查询语义

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/practice/application/commands/score_service.go`
- `code/backend/internal/module/practice/application/queries/score_service.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/runtime/module.go`

## Current Baseline

- `practice/application/commands/score_service.go` 当前直接持有 `*redis.Client`
- 用户计分锁的 `SetNX + Eval(compare-and-del)`、用户得分缓存写入和排行榜 `ZAdd` 都散落在 score command 里
- `practice/application/queries/score_service.go` 当前直接持有 `*redis.Client`
- 用户得分缓存读取与 fallback 后回填缓存都散落在 score query 里
- allowlist 当前保留：
  - `practice/application/commands/score_service.go -> github.com/redis/go-redis/v9`
  - `practice/application/queries/score_service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把 practice 用户得分的 Redis 状态面表达成 `practice` 自己的 score state store：

1. 在 `practice/ports` 新增 `PracticeScoreStateStore` 和窄锁 lease interface
2. `practice/application/commands/ScoreService` 只依赖该 store，保留“何时加锁、何时根据 solved challenge 计算总分、何时同步缓存与排行榜”的编排语义
3. `practice/application/queries/ScoreService` 只依赖该 store，保留“何时查缓存、何时 fallback 数据库、何时回填缓存”的查询编排语义
4. `practice/infrastructure` 提供 Redis adapter，内部封装 score lock、user score cache 和 ranking sorted-set 细节
5. `practice/runtime/module.go` 统一构建 score state store，并注入 score command/query wiring

## Ownership Boundary

- `practice/application/commands/score_service.go`
  - 负责：用户得分计算、锁持有期内的编排、用户总分与 solved count 的持久化时机
  - 不负责：知道 Redis key、Lua compare-and-del、`ZAdd` 或 JSON 缓存格式
- `practice/application/queries/score_service.go`
  - 负责：缓存优先查询、record-not-found fallback 和用户名补全
  - 不负责：知道 Redis key、`redis.Nil` 或缓存序列化细节
- `practice/infrastructure/score_state_store.go`
  - 负责：实现用户计分锁、用户得分缓存和排行榜 sorted-set 状态读写
  - 不负责：决定分数计算规则、缓存命中策略或用户名补全逻辑
- `practice/runtime/module.go`
  - 负责：装配 score state store 并注入 score command/query owner
  - 不负责：把 Redis client 继续暴露回 score application surface

## Change Surface

- Add: `.harness/reuse-decisions/practice-score-state-store-phase5-slice9.md`
- Add: `docs/plan/impl-plan/2026-05-13-practice-score-state-store-phase5-slice9-implementation-plan.md`
- Add: `code/backend/internal/module/practice/infrastructure/score_state_store.go`
- Add: `code/backend/internal/module/practice/ports/score_state_context_contract_test.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/module/practice/application/commands/score_service.go`
- Modify: `code/backend/internal/module/practice/application/commands/score_service_test.go`
- Modify: `code/backend/internal/module/practice/application/queries/score_service.go`
- Modify: `code/backend/internal/module/practice/application/queries/score_service_test.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [x] Slice 1: 提取 practice score state store port 与 Redis adapter
  - 目标：score command/query 不再持有 `*redis.Client`，计分锁、用户得分缓存和排行榜同步行为保持一致
  - 验证：
    - `cd code/backend && go test ./internal/module/practice/application/commands -run 'ScoreService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/practice/application/queries -run 'ScoreService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
  - Review focus：score command/query 是否已经不再知道 Redis concrete；score lock release、cache hit/fallback、ranking sync 是否保持不变

- [x] Slice 2: 删除 allowlist 并同步文档
  - 目标：删除本次真正收口的 practice score Redis allowlist，并更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除 `practice/application/commands/score_service.go` 和 `practice/application/queries/score_service.go` 两条例外；文档明确说明 `practice/application/commands/service.go` 的 submission rate-limit Redis 依赖仍未处理

## Risks

- 如果 score lock release 语义回归，会导致用户计分锁遗留或错误释放
- 如果缓存序列化字段变化，现有 `GetUserScore` cache hit 行为会回归
- 如果 ranking sync 没保留与缓存写入同批次执行的语义，用户分数与排行榜可能短暂漂移
- 如果 runtime 继续把旧 Redis client 注回 score command/query，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/application/commands -run 'ScoreService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/practice/application/queries -run 'ScoreService' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/practice/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：practice score command/query 继续持有业务编排语义，Redis 锁/缓存/排行榜状态细节落回 infrastructure
- reuse point 明确：复用既有 `cache.ScoreLockKey`、`cache.UserScoreKey`、`cache.RankingKey`，并沿用 phase5 已验证的 port + state store 模式
- 这刀同时解决行为与结构：保留 practice score 现有行为，同时删掉 score command/query 的 concrete Redis import
