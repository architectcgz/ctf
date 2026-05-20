# Practice Readmodel Boundary Phase 4 Slice 1 Implementation Plan

## Objective

完成后端模块边界迁移 phase 4 的第一刀：

- 复核 `practice_readmodel` 后确认 `/users/me/progress`、`/users/me/timeline` 只读取 `practice` 自有事实
- 把这两条查询与 HTTP 入口并回 `practice` 模块
- 删除 `practice_readmodel` 物理模块与 app/composition 装配
- 同时收口当前 touched surface 上 `practice_readmodel/application/queries/service.go -> github.com/redis/go-redis/v9` 这条 allowlist

## Non-goals

- 不处理 `teaching_readmodel` 的跨 owner 聚合能力
- 不改 `practice` 写路径、实例生命周期或排行榜查询 owner
- 不收口 `practice/application/commands/service.go`、`score_service.go` 里的 Redis 直接依赖
- 不重做用户进度缓存 key、TTL、DTO 字段或外部路由路径

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/{01-system-architecture.md,07-modular-monolith-refactor.md}`
- `code/backend/internal/app/{router.go,router_routes.go,router_test.go,full_router_integration_test.go,practice_flow_integration_test.go}`
- `code/backend/internal/app/composition/{practice_module.go,practice_readmodel_module.go}`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice_readmodel/**`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Current Baseline

- `practice_readmodel` 当前只服务 `GET /api/v1/users/me/progress` 和 `GET /api/v1/users/me/timeline`
- 这两条查询只直接读取 `submissions`、`challenges`、`instances`、`audit_logs`，没有 import `challenge/contracts`、`contest/contracts` 或其他写模块 contract
- `practice_readmodel/application/queries/QueryService` 直接持有 `*redis.Client`，导致 `architecture_allowlist_test.go` 保留 `practice_readmodel/application/queries/service.go -> github.com/redis/go-redis/v9`
- app 层仍把 `practice` 和 `practice_readmodel` 作为两个独立 builder 装配，并在用户路由里从 `practiceReadmodel.Handler` 挂载 progress / timeline

## Chosen Direction

直接把个人进度 / 时间线收口回 `practice`：

1. 在 `practice/ports` 定义 progress/timeline 查询仓储与 progress cache 的窄接口
2. 在 `practice/infrastructure` 实现查询仓储与 Redis cache adapter，沿用现有 key / TTL / JSON 语义
3. 在 `practice/application/queries` 新增 progress/timeline query service，保持当前 DTO、缓存降级和上下文传播不变
4. 在 `practice/api/http` 复用现有 handler 结构，直接挂 `GetProgress` / `GetTimeline`
5. `practice/runtime` 统一装配 command + ranking + progress/timeline query，不再需要独立 `practice_readmodel` runtime
6. 删除 `practice_readmodel` 模块目录、composition builder、allowlist 例外和相关架构叙述

## Ownership Boundary

- `practice`
  - 负责：个人进度与时间线查询、缓存回退策略、用户态只读视图
  - 不负责：把只读用户态查询再拆成独立 readmodel 模块
- `practice/infrastructure`
  - 负责：读库聚合 progress / timeline，以及 progress cache 的 Redis 适配
  - 不负责：决定查询降级策略或 HTTP 响应结构
- `app/composition`
  - 负责：只装配 `practice` 一个 owner
  - 不负责：继续保留 `practice_readmodel` builder 作为并行壳

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-practice-readmodel-phase4-slice1-implementation-plan.md`
- Add: `.harness/reuse-decisions/practice-readmodel-phase4-slice1.md`
- Add: `code/backend/internal/module/practice/infrastructure/progress_query_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/timeline_query_repository.go`
- Add: `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- Add: `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
- Add: `code/backend/internal/module/practice/application/queries/progress_timeline_progress.go`
- Add: `code/backend/internal/module/practice/application/queries/progress_timeline_timeline.go`
- Add: `code/backend/internal/module/practice/application/queries/progress_timeline_context_test.go`
- Add: `code/backend/internal/module/practice/ports/progress_timeline_context_contract_test.go`
- Modify: `code/backend/internal/module/practice/api/http/handler.go`
- Modify: `code/backend/internal/module/practice/runtime/module.go`
- Modify: `code/backend/internal/module/practice/ports/ports.go`
- Modify: `code/backend/internal/app/router.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/app/full_router_integration_test.go`
- Modify: `code/backend/internal/app/practice_flow_integration_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `docs/architecture/backend/{01-system-architecture.md,07-modular-monolith-refactor.md}`
- Modify: `docs/design/backend-module-boundary-target.md`
- Delete: `code/backend/internal/app/composition/practice_readmodel_module.go`
- Delete: `code/backend/internal/module/practice_readmodel/**`

## Task Slices

### Slice 1: 把 progress/timeline query 并回 practice

目标：

- `practice` 自己提供 `GetProgress` / `GetTimeline`
- Redis cache 通过 `practice` 模块内窄 port 表达
- `practice_readmodel` 不再参与路由装配

Validation:

- `cd code/backend && go test ./internal/module/practice/... -run 'Progress|Timeline|PracticeHandler|ScoreService' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/app/... -run 'Practice|Router|FullRouter' -count=1 -timeout 5m`

Review focus:

- progress/timeline 是否仍只读取 practice 自有事实
- cache adapter 是否只暴露 progress query 真正需要的能力
- handler 是否保持原有用户路由和响应契约

### Slice 2: 删除 practice_readmodel 壳与 allowlist

目标：

- app 不再装配 `practice_readmodel`
- 删除 `practice_readmodel/application/queries/service.go -> github.com/redis/go-redis/v9`
- 删除 `practice_readmodel` 相关架构守卫和模块 builder 断言

Validation:

- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/app/... -run 'Practice|Router|FullRouter' -count=1 -timeout 5m`

Review focus:

- 是否真的删掉了并行 owner，而不是把旧模块换壳保留
- allowlist 是否只删除这次已收口的例外

### Slice 3: 文档回收当前事实

目标：

- phase 4 当前进展写回架构事实源
- 目标设计稿明确 `practice_readmodel` 已并回 `practice`

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否准确区分“phase 4 第一刀已完成”和 `teaching_readmodel` 仍保留的事实
- 是否避免继续把 `practice_readmodel` 写成当前模块版图

## Risks

- 如果 `practice` handler 注入方式改坏，`/api/v1/users/me/progress` 和 `/api/v1/users/me/timeline` 会在路由层失效
- 如果 progress cache adapter 行为与原逻辑不一致，用户进度会出现缓存 miss / 反序列化后的回退差异
- 删除 `practice_readmodel` 后，若仍有测试或文档假设其独立 builder 存在，会暴露成编译或 guardrail 失败

## Verification Plan

1. `cd code/backend && go test ./internal/module/practice/... -run 'Progress|Timeline|PracticeHandler|ScoreService' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/app/... -run 'Practice|Router|FullRouter' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
4. `python3 scripts/check-docs-consistency.py`
5. `bash scripts/check-consistency.sh`
6. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：个人进度 / 时间线只属于 `practice`，不再平行挂在 `practice_readmodel`
- reuse point 明确：复用现有 `practice` handler / runtime / repository 结构，以及已有 progress cache key 和 DTO
- 这刀同时解决行为与结构：保留原路由与响应，同时删掉独立 readmodel 壳和该壳上的 Redis concrete allowlist
- touched surface 上的已知 debt 就是 `practice_readmodel` 边界不稳定和 query service 直持 Redis；本切片一起收口，不留回流入口
