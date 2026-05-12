# Reuse Decision

## Change type

teaching_readmodel / query service / handler / runtime / architecture

## Existing code searched

- `code/backend/internal/module/teaching_readmodel/application/queries/{contracts.go,service.go,service_overview_test.go}`
- `code/backend/internal/module/teaching_readmodel/api/http/{handler.go,handler_contract_test.go}`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/features/教师教学概览聚合架构.md`

## Similar implementations found

- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺 overview 功能，而是 `teaching_readmodel` 已经只剩真实跨 owner 聚合后，`GetOverview` 仍挂在一个覆盖目录、班级洞察和学生复盘的大一统 query surface 上。最小合理改法不是新建模块，而是复用现有 `teaching_readmodel` 目录与 repository，把 overview 抽成独立 `OverviewService`，同时让 HTTP handler 不再通过单个宽接口承接所有教师查询。

## Files to modify

- `code/backend/internal/module/teaching_readmodel/application/queries/contracts.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/overview_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/service_overview_test.go`
- `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- `code/backend/internal/module/teaching_readmodel/api/http/handler_contract_test.go`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- `code/backend/internal/module/teaching_readmodel/architecture_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/features/教师教学概览聚合架构.md`
- `docs/plan/impl-plan/2026-05-12-teaching-readmodel-overview-query-surface-phase4-slice2-implementation-plan.md`

## After implementation

- 后续如果继续拆 `GetClassSummary/GetClassTrend/GetClassReview` 或学生复盘查询，再沿用同样的 query-surface owner 方式继续收口，不回到宽 `Service`。
