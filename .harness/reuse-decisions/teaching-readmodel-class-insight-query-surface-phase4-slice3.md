# Reuse Decision

## Change type

teaching_readmodel / query service / handler / runtime / architecture

## Existing code searched

- `code/backend/internal/module/teaching_readmodel/application/queries/{contracts.go,overview_service.go,service.go,service_overview_test.go}`
- `code/backend/internal/module/teaching_readmodel/api/http/{handler.go,handler_contract_test.go}`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- `code/backend/internal/module/teaching_readmodel/ports/query.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/features/教师教学概览聚合架构.md`

## Similar implementations found

- `code/backend/internal/module/teaching_readmodel/application/queries/overview_service.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/handler.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺班级详情能力，而是 `GetClassSummary`、`GetClassTrend`、`GetClassReview` 仍和目录查询、学生复盘混在同一个宽 query surface 里。最小合理改法不是新建模块，也不是改底层 repository，而是复用现有 `teaching_readmodel` 目录、DTO、repository 和路由，把 class insight 抽成独立 `ClassInsightService`，并让 HTTP handler 不再通过剩余宽接口承接班级详情洞察。

## Files to modify

- `code/backend/internal/module/teaching_readmodel/application/queries/contracts.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- `code/backend/internal/module/teaching_readmodel/api/http/handler_contract_test.go`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- `code/backend/internal/module/teaching_readmodel/architecture_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/features/教师教学概览聚合架构.md`
- `docs/plan/impl-plan/2026-05-12-teaching-readmodel-class-insight-query-surface-phase4-slice3-implementation-plan.md`

## After implementation

- 如果后续继续拆学生复盘查询，继续沿用 query-surface owner 的收口方式，不回退到新的大一统 `Service`
- 如果这组拆分证明可复用，再补到 `harness/reuse/history.md` 与 `harness/reuse/index.yaml`
