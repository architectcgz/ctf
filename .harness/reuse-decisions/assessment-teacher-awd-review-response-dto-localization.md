# Reuse Decision

## Change type

query / mapper / export / api

## Existing code searched

- `code/backend/internal/module/assessment/application/queries/*teacher_awd_review*.go`
- `code/backend/internal/module/assessment/application/commands/{awd_review_export_builder.go,awd_review_export_renderer.go,report_service.go}`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/dto/teacher_awd_review.go`

## Similar implementations found

- `teaching_query` 的 timeline DTO 已经收回 owning query package
- `assessment` skill profile command/report 链已经使用模块内 contract / query snapshot 供 report/export 消费

## Decision

refactor_existing

## Reason

教师侧 AWD review 的 list / archive / export archive 形状只在 assessment 模块内部流动：query service 生成、handler 返回、runtime 包装转接、report/export 直接消费。最小正确方案是复用现有 query mapper，把响应与导出归档模型收回 `assessment/application/queries`，而不是继续留在全局 DTO 桶。

## Files to modify

- `.harness/reuse-decisions/assessment-teacher-awd-review-response-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-17-assessment-teacher-awd-review-response-dto-localization-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-assessment-teacher-awd-review-response-dto-localization-review.md`
- `code/backend/internal/module/assessment/application/queries/teacher_awd_review_types.go`
- `code/backend/internal/module/assessment/application/queries/teacher_awd_review_service.go`
- `code/backend/internal/module/assessment/application/queries/response_mapper.go`
- `code/backend/internal/module/assessment/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/teacher_awd_review.go`
