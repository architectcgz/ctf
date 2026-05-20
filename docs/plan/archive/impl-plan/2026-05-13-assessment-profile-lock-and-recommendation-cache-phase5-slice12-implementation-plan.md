# Assessment Profile Lock And Recommendation Cache Phase 5 Slice 12 Implementation Plan

## Objective

继续 phase 5 收窄 `assessment` application concrete allowlist，把画像锁和推荐缓存的 Redis 状态面下沉到模块 port / infrastructure adapter：

- 去掉 `assessment/application/commands/profile_service.go -> github.com/redis/go-redis/v9`
- 去掉 `assessment/application/queries/recommendation_service.go -> github.com/redis/go-redis/v9`
- 保持画像锁、推荐缓存命中和事件触发缓存失效的现有行为不变

## Non-goals

- 不处理 `assessment/application/commands/report_service.go -> gorm.io/gorm`
- 不改画像计算公式、推荐弱项判定、推荐列表排序或事件名
- 不改 `assessment` HTTP contract、数据库 schema 或 challenge recommendation 查询 SQL

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/runtime/module.go`

## Current Baseline

- `profile_service.go` 当前直接持有 `*redis.Client`
- 画像增量更新与全量计算的 lock key、`SetNX` / `Del` 和锁失败 fallback 都散落在 command 层
- `recommendation_service.go` 当前直接持有 `*redis.Client`
- recommendation cache 的 key、JSON 编解码、`Get/Set/Del` 和事件触发 cache invalidation 都散落在 query 层
- allowlist 当前保留：
  - `assessment/application/commands/profile_service.go -> github.com/redis/go-redis/v9`
  - `assessment/application/queries/recommendation_service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

把 assessment 剩余 Redis 状态面表达成两个 assessment 自己的窄 state store：

1. 在 `assessment/ports` 新增 `AssessmentProfileLockStore` 和 `AssessmentRecommendationCacheStore`
2. `profile_service.go` 只依赖画像锁 port，保留“何时尝试加锁、何时回退已有画像”的业务编排语义
3. `recommendation_service.go` 只依赖推荐缓存 port，保留“何时默认 limit 走缓存、何时事件触发失效”的业务编排语义
4. `assessment/infrastructure/state_store.go` 提供同文件 Redis adapter，内部封装 lock key、recommendation key、TTL、JSON 编解码和 `Get/Set/Del/SetNX`
5. `assessment/runtime/module.go` 统一构建这两个 store，并注入 profile / recommendation wiring

## Ownership Boundary

- `assessment/application/commands/profile_service.go`
  - 负责：画像增量更新和全量计算的业务编排，决定何时尝试加锁、何时回退已有画像
  - 不负责：知道 lock key、`SetNX`、`Del` 或 Redis client 细节
- `assessment/application/queries/recommendation_service.go`
  - 负责：推荐查询编排、默认 limit 缓存命中、事件触发失效
  - 不负责：知道 recommendation cache key、JSON 编解码、TTL 或 Redis client 细节
- `assessment/infrastructure/state_store.go`
  - 负责：实现画像锁和推荐缓存的 Redis 状态读写
  - 不负责：决定画像计算时机、弱项判定或推荐内容生成规则
- `assessment/runtime/module.go`
  - 负责：装配两个 state store 并注入 assessment command/query service
  - 不负责：把 Redis client 继续暴露回 assessment application surface

## Change Surface

- Add: `.harness/reuse-decisions/assessment-profile-lock-and-recommendation-cache-phase5-slice12.md`
- Add: `docs/plan/impl-plan/2026-05-13-assessment-profile-lock-and-recommendation-cache-phase5-slice12-implementation-plan.md`
- Add: `code/backend/internal/module/assessment/infrastructure/state_store.go`
- Add: `code/backend/internal/module/assessment/ports/state_store_context_contract_test.go`
- Modify: `code/backend/internal/module/assessment/ports/ports.go`
- Modify: `code/backend/internal/module/assessment/application/commands/profile_service.go`
- Modify: `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- Modify: `code/backend/internal/module/assessment/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`

## Task Slices

- [x] Slice 1: 提取画像锁 store 和推荐缓存 store
  - 目标：assessment profile / recommendation application surface 不再持有 `*redis.Client`
  - 验证：
    - `cd code/backend && go test ./internal/module/assessment/application/commands -run 'ProfileService|CalculateSkillProfile|GetSkillProfile' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/assessment/application/queries -run 'RecommendationService' -count=1 -timeout 5m`
    - `cd code/backend && go test ./internal/module/assessment/runtime -run '^$' -count=1 -timeout 5m`
  - Review focus：画像锁和 recommendation cache 是否已经都不再知道 Redis concrete；锁失败 fallback、缓存命中和事件失效行为是否保持一致

- [x] Slice 2: 删除 allowlist 并同步文档
  - 目标：删除 assessment 剩余两条 Redis allowlist，并更新 phase 5 当前事实
  - 验证：
    - `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
    - `python3 scripts/check-docs-consistency.py`
    - `bash scripts/check-consistency.sh`
  - Review focus：只删除本次 assessment 实际收口的 allowlist；文档准确描述 phase 5 现在还剩哪些 concrete surface

## Risks

- 如果画像锁 key 形状变化，现有并发保护会漂移
- 如果 recommendation cache 的 JSON 结构或 TTL 变化，现有缓存命中会回退成 miss
- 如果事件消费后没有清理缓存，推荐结果会长期陈旧
- 如果 runtime 继续把 Redis client 注回 application，allowlist 不会真正收口

## Verification Plan

1. `cd code/backend && go test ./internal/module/assessment/application/commands -run 'ProfileService|CalculateSkillProfile|GetSkillProfile' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/assessment/application/queries -run 'RecommendationService' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module/assessment/runtime -run '^$' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：profile command 继续持有画像锁的业务编排语义，recommendation query 继续持有缓存命中和失效的业务编排语义
- reuse point 明确：复用 phase5 已验证的模块内窄 store + runtime wiring 模式，不引入新的宽 assessment cache service
- 这刀同时解决行为与结构：assessment 最后两条 Redis concrete allowlist 一次收口，但仍保持 lock 和 recommendation cache 两个独立 owner surface
