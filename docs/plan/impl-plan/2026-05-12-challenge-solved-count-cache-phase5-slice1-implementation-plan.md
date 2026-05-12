# Challenge Solved Count Cache Phase 5 Slice 1 Implementation Plan

## Objective

完成后端模块边界迁移 phase 5 的一个最小切片：

- 去掉 `challenge/application/queries/challenge_service.go` 对 `github.com/redis/go-redis/v9` 的直接依赖
- 把“题目解出人数 solved-count 缓存”收口成 `challenge` 自己的窄 cache port
- 保持 `GetPublishedChallenge` 的外部行为不变：有缓存时优先读缓存，缓存 miss 或异常时回退数据库

## Non-goals

- 不处理 `challenge/application/queries/challenge_service.go -> gorm.io/gorm` 这条 allowlist
- 不收口 `practice`、`contest` 里的 Redis 直接依赖
- 不改题目列表、题目详情 DTO 字段或 HTTP 路由
- 不重做通用缓存框架，也不新增跨模块共享 cache 包装

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`

## Current Baseline

- `challenge/application/queries/ChallengeService` 当前直接持有 `*redis.Client`
- solved-count 缓存键、JSON 编解码、Redis `Nil` 分支和 TTL 都写在 application query service 里
- `challenge/runtime/module.go` 直接把进程级 Redis client 注入 query service
- `architecture_allowlist_test.go` 里仍保留 `challenge/application/queries/challenge_service.go -> github.com/redis/go-redis/v9`

## Chosen Direction

沿 `challenge` 模块自身收口一个用例级 cache port：

1. 在 `challenge/ports` 新增 solved-count cache interface，只暴露 query service 真正需要的读写能力
2. 在 `challenge/infrastructure` 新增 Redis adapter，承接 key、JSON 编解码和 Redis miss 细节
3. `ChallengeService` 只依赖 repository + solved-count cache port + config，不再 import Redis client
4. `challenge/runtime` 负责构建 Redis adapter 并注入 query service
5. 删除 allowlist 例外，并同步最小行为测试

## Ownership Boundary

- `challenge/application/queries`
  - 负责：决定何时走缓存、何时回退 DB、缓存失败时如何降级
  - 不负责：知道 Redis client、Redis `Nil`、key 拼装或 JSON 存储格式
- `challenge/infrastructure`
  - 负责：实现 solved-count cache port，处理 Redis key、序列化和 miss 语义
  - 不负责：决定业务回退策略或题目详情组装逻辑
- `challenge/runtime`
  - 负责：把进程级 Redis client 装配成模块内 cache adapter
  - 不负责：把 Redis 细节重新暴露回 application 层

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-challenge-solved-count-cache-phase5-slice1-implementation-plan.md`
- Add: `.harness/reuse-decisions/challenge-solved-count-cache-phase5-slice1.md`
- Add: `code/backend/internal/module/challenge/infrastructure/solved_count_cache.go`
- Add: `code/backend/internal/module/challenge/ports/solved_count_cache_context_contract_test.go`
- Modify: `code/backend/internal/module/challenge/application/queries/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/queries/challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: 提取 solved-count cache port 与 Redis adapter

目标：

- `challenge/ports` 有明确的 solved-count cache port
- Redis 细节下沉到 `challenge/infrastructure`
- `ChallengeService` 不再 import `github.com/redis/go-redis/v9`

Validation:

- `cd code/backend && go test ./internal/module/challenge/application/queries -run 'SolvedCount|GetPublishedChallenge|HonorsCancellation' -count=1 -timeout 5m`

Review focus:

- cache port 是否足够窄，只表达 solved-count 读写，而没有泄漏 Redis API
- context 是否沿原调用链继续向下传递，没有新建 background context

### Slice 2: 收口 allowlist 与文档

目标：

- 删除 `challenge/application/queries/challenge_service.go -> github.com/redis/go-redis/v9`
- 把 phase 5 当前进展回收到设计/架构文档

Validation:

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否只更新当前已落地事实，没有把未完成的 phase 5 范围写成已完成
- allowlist 是否只删除本次真正收口的例外

## Integration Checks

- 已发布题目详情仍能返回 `SolvedCount`
- 缓存 miss 时仍会回退数据库
- 已取消的 context 仍会在 solved-count 查询链路上得到 `context.Canceled`
- `challenge/application/queries/challenge_service.go` 不再直接 import Redis client

## Rollback / Recovery Notes

- 本切片无 schema 变更，可整体代码回退
- 若 Redis adapter 行为与原缓存语义不一致，应整体回退本切片，而不是保留“application 一半自己读 Redis、一半走 port”的混合状态

## Risks

- cache port 过宽会把 phase 5 的 concrete dependency 换成新的宽接口，边界仍然不清
- 如果 adapter 把缓存异常误判成命中，题目详情 `SolvedCount` 可能返回错误值
- 如果 runtime 忘记注入 adapter，题目详情会退化成全量查库；功能可用，但会丢掉缓存收益

## Verification Plan

1. `cd code/backend && go test ./internal/module/challenge/application/queries -run 'SolvedCount|GetPublishedChallenge|HonorsCancellation' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`
5. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：application 负责回退策略，infrastructure 负责 Redis 适配
- reuse point 明确：复用现有 `cache.ChallengeSolvedCountKey`、challenge runtime wiring、仓库内既有 Redis adapter 模式
- 这刀同时解决行为与结构：保留 solved-count 缓存能力，同时删掉 query service 的 Redis concrete import
- touched surface 上的已知 debt 正是 `challenge_service.go -> redis` 这条 allowlist，本切片直接收口，不把它留成后续说明
