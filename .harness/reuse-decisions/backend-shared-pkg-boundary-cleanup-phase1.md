# Reuse Decision

## Change type
infrastructure / cache / middleware / composition

## Existing code searched

- `code/backend/internal/module/*/infrastructure/`
- `code/backend/internal/infrastructure/`
- `code/backend/internal/app/`
- `code/backend/internal/middleware/`
- `code/backend/pkg/`
- `docs/tasks/backend-task-breakdown.md`

## Similar implementations found

- `code/backend/internal/module/contest/infrastructure/cachekeys/redis_keys.go`
- `code/backend/internal/module/practice/infrastructure/cachekeys/redis_keys.go`
- `code/backend/internal/module/runtime/infrastructure/cachekeys/redis_keys.go`
- `code/backend/internal/infrastructure/redislock/lock.go`
- `code/backend/internal/module/challenge/infrastructure/solved_count_cache.go`

## Decision
refactor_existing

## Reason

这刀不是新增新的缓存 / 日志 / 限流模式，而是延续前面 DTO、model、redis key 边界清理里已经形成的 owner 规则：

- 单模块私有能力下沉到模块自己的 `infrastructure`
- 只服务后端内部的共享实现迁回 `internal/infrastructure`
- 不再让 `pkg/*` 继续充当“看起来通用、实际只在 internal 使用”的历史共享桶

当前 `challenge` solved-count key 只被 challenge 自己使用，复用已有模块内 `cachekeys` 模式即可；`logger` 与 `ratelimit` 都只用于内部 wiring / middleware，复用已有共享基础设施包的放置方式最小且一致，不需要重新设计新的抽象层。

## Files to modify

- `code/backend/internal/module/challenge/infrastructure/cachekeys/redis_keys.go`
- `code/backend/internal/module/challenge/infrastructure/solved_count_cache.go`
- `code/backend/internal/infrastructure/logger/logger.go`
- `code/backend/internal/infrastructure/ratelimit/ratelimit.go`
- `code/backend/internal/bootstrap/run.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/middleware/ratelimit.go`
- `code/backend/internal/middleware/ratelimit_test.go`
- `code/backend/pkg/cache/keys.go`
- `code/backend/pkg/logger/logger.go`
- `code/backend/pkg/ratelimit/ratelimit.go`
- `docs/plan/archive/impl-plan/2026-05-19-backend-shared-pkg-boundary-cleanup-implementation-plan.md`

## After implementation

- 若后续 `pkg/response`、`pkg/websocket`、`pkg/crypto`、`pkg/errcode` 继续按 owner 规则拆完，再把“共享 pkg 迁移判定规则”沉淀到 `harness/reuse/history.md` 或长期索引。
