# Reuse Decision

## Change type

teaching_readmodel / query service / handler / runtime / architecture

## Existing code searched

- `code/backend/internal/module/teaching_readmodel/application/queries/{contracts.go,service.go,overview_service.go,class_insight_service.go}`
- `code/backend/internal/module/teaching_readmodel/api/http/{handler.go,handler_contract_test.go}`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/cmd/seed-teaching-review-data/main.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/features/教师教学概览聚合架构.md`

## Similar implementations found

- `code/backend/internal/module/teaching_readmodel/application/queries/overview_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- `code/backend/internal/module/assessment/api/http/report_handler.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少学生复盘能力，而是 `GetStudentProgress`、`GetStudentRecommendations`、`GetStudentTimeline`、`GetStudentEvidence`、`GetStudentAttackSessions` 仍和目录查询挂在同一个剩余宽 surface 上。最小合理改法不是新建模块或改底层 SQL，而是复用现有 `teaching_readmodel` 目录、repository、DTO 和路由，把学生复盘收成独立 `StudentReviewService`，并让 HTTP handler 不再通过剩余目录接口承接 `/teacher/students/:id/*`。

## Files to modify

- `code/backend/internal/module/teaching_readmodel/application/queries/contracts.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/student_review_service_test.go`
- `code/backend/internal/module/teaching_readmodel/api/http/handler.go`
- `code/backend/internal/module/teaching_readmodel/api/http/handler_contract_test.go`
- `code/backend/internal/module/teaching_readmodel/runtime/module.go`
- `code/backend/internal/module/teaching_readmodel/architecture_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/cmd/seed-teaching-review-data/main.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/features/教师教学概览聚合架构.md`
- `docs/plan/impl-plan/2026-05-12-teaching-readmodel-student-review-query-surface-phase4-slice4-implementation-plan.md`

## After implementation

- 如果后续还需要继续把目录查询显式命名为更窄 owner，可在不改学生复盘 owner 的前提下继续收口
- 如果学生复盘以后再补新的聚合视图，继续沿 query-surface owner 的方式演进，不回退到新的大一统 `Service`
