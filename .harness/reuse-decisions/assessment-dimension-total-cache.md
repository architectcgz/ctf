# Reuse Decision

## Change type

repository / state_store / config / runtime / challenge-event / test

## Existing code searched

- `code/backend/internal/module/assessment/infrastructure/*.go`
- `code/backend/internal/module/assessment/ports/*.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/challenge/infrastructure/solved_count_cache.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/queries/recommendation_service.go`
- `code/backend/internal/module/challenge/contracts/events.go`
- `code/backend/internal/config/config.go`

## Similar implementations found

- `assessment/infrastructure/state_store.go` 已经承接画像锁与推荐缓存，适合继续承接 assessment 维度总分缓存
- `challenge/infrastructure/solved_count_cache.go` 已经提供了当前仓库里最接近的 Redis 计数缓存模式
- `assessment/infrastructure/repository.go` 已经是画像维度得分读取 owner，适合在这里收口“缓存总分 + 查询用户得分”的组合逻辑
- `assessment/application/queries/recommendation_service.go` 已经有 practice / contest 事件删缓存模式，适合复用到 assessment 维度总分失效
- `challenge/application/commands/challenge_service.go` 已经有 event bus 接口和弱发布 helper，适合继续承接题库变更事件

## Decision

refactor_existing

## Reason

这次目标是优化已有 assessment 画像计算路径，不是新增统计表、后台同步任务或新的聚合服务：

- Redis cache 继续放在 assessment 现有 state store 层
- 组合逻辑继续留在 assessment repository
- runtime 继续使用 assessment module 现有装配入口
- challenge 写路径继续使用 challenge module 现有 event bus，而不是给 assessment 反向注入 challenge repository
- 测试继续落在 assessment infrastructure 与 application tests

这样可以用最小改动把“每个学生都重算题库维度总分”收掉，并且把“题库变更后马上准”的要求收在既有事件边界里，同时保持边界不扩散。

## Files to modify

- `.harness/reuse-decisions/assessment-dimension-total-cache.md`
- `docs/plan/archive/impl-plan/2026-05-21-assessment-dimension-total-cache-implementation-plan.md`
- `code/backend/internal/config/config.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/infrastructure/cachekeys/redis_keys.go`
- `code/backend/internal/module/assessment/infrastructure/state_store.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/infrastructure/repository_test.go`
- `code/backend/internal/module/assessment/application/commands/profile_service_test.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/application/commands/dimension_total_cache_invalidation_service.go`
- `code/backend/internal/module/assessment/application/commands/dimension_total_cache_invalidation_service_test.go`
- `code/backend/internal/module/challenge/contracts/events.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`

## After implementation

- 已发布题目各维度总分不再为每个学生重复做 SQL 聚合
- 画像计算只查询学生自己的已解题得分，再和缓存总分组装
- challenge 的 `Update / Delete / Publish / CommitChallengeImport` 会在影响已发布总分时发事件
- assessment 收到题库变更事件后会立刻删除维度总分缓存
- TTL 只保留为 Redis 读写异常场景下的兜底，不再是主失效路径
