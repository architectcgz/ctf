# Teaching Readmodel Overview Query Surface Phase 4 Slice 2 Implementation Plan

## Objective

继续完成 phase 4 的下一刀，但只处理 `teaching_readmodel` 里一个最小、可审查的 query surface：

- 把教师总览 `GetOverview` 从宽 `application/queries.Service` 里拆到独立 `OverviewService`
- 让 `api/http.Handler` 不再通过单个大一统 query 接口承接 overview
- 保持 `GET /api/v1/teacher/overview` 的路径、响应结构和教师权限口径不变

## Non-goals

- 不改 `/api/v1/teacher/classes/:name/{summary,trend,review}` 的查询 owner
- 不改学生复盘、时间线、证据链和推荐题查询 contract
- 不重写 `teaching_readmodel/infrastructure.Repository` 的底层 SQL
- 不进入 phase 5，不处理 Redis / GORM concrete allowlist

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/features/教师教学概览聚合架构.md`
- `code/backend/internal/module/teaching_readmodel/application/queries/{contracts.go,service.go,service_overview_test.go}`
- `code/backend/internal/module/teaching_readmodel/api/http/{handler.go,handler_contract_test.go}`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`

## Current Baseline

- `teaching_readmodel/application/queries/contracts.go` 目前把目录、班级洞察、overview、学生复盘全部挂在同一个 `Service` 接口下。
- `teaching_readmodel/api/http.Handler` 也只持有一个 `Service` 字段，`GetOverview` 与班级详情/学生复盘复用同一条 query surface。
- `GetOverview` 的实现和 helper 目前直接混在 `application/queries/service.go` 里，和其他教师查询 owner 没有物理边界。

## Chosen Direction

1. 保留 `teaching_readmodel` 作为物理模块，不新建新模块或新路由。
2. 在 `application/queries` 内新增独立 `OverviewService` 与 `OverviewQueryService`，只负责教师总览。
3. 现有 `Service` 接口收窄为“除 overview 外的教师查询能力”。
4. `api/http.Handler` 改为分别依赖 `Service` 与 `OverviewService`，让 `/teacher/overview` 不再经过宽接口。
5. `runtime/module.go` 只在 wiring 边缘分别构造两类 query service，并同步更新 guardrail 与架构事实。

## Ownership Boundary

- `OverviewQueryService`
  - 负责：overview scope 的可访问班级求值、聚合摘要、聚合趋势、重点班级与重点学生组装
  - 不负责：班级详情 review、学生证据链和推荐题查询
- `application/queries.Service`
  - 负责：目录、班级详情和学生复盘剩余查询
  - 不负责：继续持有 overview query owner
- `api/http.Handler`
  - 负责：把 overview route 绑定到 `OverviewService`
  - 不负责：通过单个宽接口承接所有教师读模型

## Change Surface

- Add: `.harness/reuse-decisions/teaching-readmodel-overview-query-surface-phase4-slice2.md`
- Add: `docs/plan/impl-plan/2026-05-12-teaching-readmodel-overview-query-surface-phase4-slice2-implementation-plan.md`
- Add: `code/backend/internal/module/teaching_readmodel/application/queries/overview_service.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/contracts.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service_overview_test.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler_contract_test.go`
- Modify: `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- Modify: `code/backend/internal/module/teaching_readmodel/architecture_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/architecture/features/教师教学概览聚合架构.md`

## Task Slices

### Slice 1: 把 overview 迁到独立 query service

目标：

- `GetOverview` 和它的 helper 从 `service.go` 移到独立 `overview_service.go`
- `contracts.go` 单独声明 `OverviewService`
- `service_overview_test.go` 改为直接覆盖独立 overview owner

Validation:

- `cd code/backend && go test ./internal/module/teaching_readmodel/application/queries -run 'Overview' -count=1 -timeout 120s`

Review focus:

- overview 的可访问班级口径、摘要计算和趋势映射是否保持不变
- overview helper 是否真的从宽 query owner 中移出，而不是继续留一个未使用的旧实现

### Slice 2: 收窄 handler 与 runtime wiring

目标：

- `api/http.Handler` 分别依赖 `Service` 与 `OverviewService`
- `runtime/module.go` 不再暴露宽 `Query` 字段
- `handler_contract_test.go` 与 `architecture_test.go` 表达新的边界

Validation:

- `cd code/backend && go test ./internal/module/teaching_readmodel/... -count=1 -timeout 120s`

Review focus:

- handler 是否只有 overview 走独立 query owner
- runtime wiring 是否没有继续把宽 query surface 往外暴露

### Slice 3: 文档回收 phase 4 当前事实

目标：

- phase 4 当前状态更新到 slice 2
- feature 文档与模块边界事实源同步指向新的 overview owner

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否把这刀写成“query surface 拆分”而不是“phase 4 已整体完成”

## Risks

- 如果 handler 仍错误调用宽 `Service`，这刀只会把 owner 改名，不会真正收口
- 如果 `GetOverview` helper 迁移不完整，后续 class insight / student review 继续拆时会重复碰同一批残留逻辑
- 如果 runtime 还保留宽 `Query` 出口，后续调用点仍可能绕过新边界

## Verification Plan

1. `cd code/backend && go test ./internal/module/teaching_readmodel/... -count=1 -timeout 120s`
2. `python3 scripts/check-docs-consistency.py`
3. `bash scripts/check-consistency.sh`
4. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：overview 现在有独立 query owner，不再挂在宽 `Service` 上
- reuse point 明确：复用现有 repository、DTO、handler 路由和教师权限口径，不新增模块或并行 API
- 这刀同时解决行为与结构：外部接口不变，但 `OverviewService`、handler 依赖和 runtime wiring 一起收口
- 本切片不会制造“做完 overview 还得立刻重拆 handler”的二次返工；剩余 phase 4 范围已经明确转入 class insight / student review query surface
