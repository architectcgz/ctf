# Teaching Readmodel Class Insight Query Surface Phase 4 Slice 3 Implementation Plan

## Objective

继续完成 phase 4，但只处理 `teaching_readmodel` 剩余宽 query surface 里最适合独立 owner 的下一组查询：

- 把 `GetClassSummary`、`GetClassTrend`、`GetClassReview` 从宽 `application/queries.Service` 拆到独立 `ClassInsightService`
- 让 `api/http.Handler` 不再通过剩余宽接口承接班级详情洞察查询
- 保持 `GET /api/v1/teacher/classes/:name/{summary,trend,review}` 的路径、响应结构和权限口径不变

## Non-goals

- 不改 `ListClasses`、`ListStudents`、`ListClassStudents` 的 owner
- 不改学生复盘、时间线、证据链和推荐题查询 contract
- 不重写 `teaching_readmodel/infrastructure.Repository` 的底层 SQL
- 不进入 phase 5，不处理 Redis / GORM concrete allowlist

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/features/教师教学概览聚合架构.md`
- `docs/design/教学复盘建议优化方案.md`
- `code/backend/internal/module/teaching_readmodel/application/queries/{contracts.go,overview_service.go,service.go}`
- `code/backend/internal/module/teaching_readmodel/api/http/{handler.go,handler_contract_test.go}`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`

## Current Baseline

- `teaching_readmodel/application/queries/contracts.go` 目前把目录查询、班级详情洞察和学生复盘仍挂在同一个 `Service` 接口下。
- `GetClassSummary`、`GetClassTrend`、`GetClassReview` 共用同一套教师班级访问判断，但仍混在 `service.go`，和学生复盘 owner 没有物理边界。
- `api/http.Handler` 只把 overview 拆成了单独依赖，班级详情洞察仍通过 `Handler.service` 进入剩余宽 surface。

## Chosen Direction

1. 保留 `teaching_readmodel` 作为物理模块，不新建模块或新路由。
2. 在 `application/queries` 内新增独立 `ClassInsightService` 与 `ClassInsightQueryService`，只负责班级详情洞察三条查询。
3. 现有 `Service` 接口收窄为目录与学生复盘查询，不再承接 class insight。
4. `api/http.Handler` 改为分别依赖 `Service`、`OverviewService`、`ClassInsightService`。
5. `runtime/module.go` 在 wiring 边缘分别构造三类 query service，并同步更新 guardrail 与架构事实。

## Ownership Boundary

- `ClassInsightQueryService`
  - 负责：`GetClassSummary`、`GetClassTrend`、`GetClassReview`、班级级访问校验后的洞察拼装、教学建议输入整形、推荐题 fallback
  - 不负责：overview 聚合、学生个人复盘、目录分页与排序
- `application/queries.Service`
  - 负责：目录、班级成员列表、学生进度、推荐、时间线、证据链与攻击会话查询
  - 不负责：继续持有 class insight query owner
- `api/http.Handler`
  - 负责：把 `classes/:name/{summary,trend,review}` 绑定到 `ClassInsightService`
  - 不负责：通过剩余宽接口承接班级详情洞察

## Change Surface

- Add: `.harness/reuse-decisions/teaching-readmodel-class-insight-query-surface-phase4-slice3.md`
- Add: `docs/plan/impl-plan/2026-05-12-teaching-readmodel-class-insight-query-surface-phase4-slice3-implementation-plan.md`
- Add: `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- Add: `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service_test.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/contracts.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- Modify: `code/backend/internal/module/teaching_readmodel/api/http/handler_contract_test.go`
- Modify: `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- Modify: `code/backend/internal/module/teaching_readmodel/architecture_test.go`
- Modify: `docs/design/backend-module-boundary-target.md`
- Modify: `docs/architecture/backend/04-api-design.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/architecture/features/教师教学概览聚合架构.md`

## Task Slices

### Slice 1: 抽出 class insight owner

目标：

- `GetClassSummary`、`GetClassTrend`、`GetClassReview` 和 class-review helper 从 `service.go` 移到独立 `class_insight_service.go`
- `contracts.go` 单独声明 `ClassInsightService`
- shared access helper 保持单点 owner，不在 `QueryService` 和 `ClassInsightQueryService` 间复制

Validation:

- `cd code/backend && go test ./internal/module/teaching_readmodel/application/queries -run 'ClassInsight|Overview' -count=1 -timeout 120s`

Review focus:

- 班级访问校验是否保持原口径
- review 生成链路是否仍只读取 readmodel 与 recommendation provider，不引入新跨模块依赖

### Slice 2: 收窄 handler 与 runtime wiring

目标：

- `api/http.Handler` 分别依赖 `Service`、`OverviewService`、`ClassInsightService`
- `runtime/module.go` 不再把班级详情洞察继续挂在剩余宽 query service 上
- `handler_contract_test.go` 与 `architecture_test.go` 表达新的边界

Validation:

- `cd code/backend && go test ./internal/module/teaching_readmodel/... -count=1 -timeout 120s`

Review focus:

- `/teacher/classes/:name/{summary,trend,review}` 是否都走独立 query owner
- runtime wiring 是否没有继续把 class insight 暗中暴露给宽 `Service`

### Slice 3: 回收 phase 4 当前事实

目标：

- phase 4 当前状态更新到 slice 3
- API 与 feature 文档同步指向新的 class insight owner

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否准确描述“继续拆 query surface”，而不是错误宣称 phase 4 已完成

## Risks

- 如果访问校验 helper 拆错，教师与管理员的班级访问范围可能回归
- 如果 handler 仍经由剩余宽 `Service` 间接进入 class insight，这刀只会改名不会真正收口
- 如果 class review helper 仍散落在 `service.go`，后续 student review 再拆时会继续反复碰同一块残留逻辑

## Verification Plan

1. `cd code/backend && go test ./internal/module/teaching_readmodel/... -count=1 -timeout 120s`
2. `python3 scripts/check-docs-consistency.py`
3. `bash scripts/check-consistency.sh`
4. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：class insight 现在有独立 query owner，不再挂在剩余宽 `Service` 上
- reuse point 明确：复用现有 repository、DTO、handler 路由和 recommendation provider，不新增模块或并行 API
- 这刀同时解决行为与结构：外部接口不变，但 `ClassInsightService`、handler 依赖和 runtime wiring 一起收口
- 本切片不会制造“做完班级详情后还要立刻重拆 handler”的二次返工；phase 4 剩余范围会继续收敛到 student review surface
