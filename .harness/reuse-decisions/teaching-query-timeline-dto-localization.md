# Reuse Decision

## Change type

query / mapper / api

## Existing code searched

- `code/backend/internal/module/teaching_query/application/queries/*.go`
- `code/backend/internal/module/practice/api/http/progress_dto.go`
- `code/backend/internal/dto/timeline.go`

## Similar implementations found

- `practice/api/http` 已经把学生侧 timeline DTO 收回模块边界
- `teaching_query/application/queries` 已经有现成 response mapper，可直接沿用 timeline event 映射

## Decision

refactor_existing

## Reason

教师侧 student timeline 只在 `teaching_query/application/queries` 生成并由 handler 直接返回，不满足继续留在全局 `internal/dto` 的共享条件。最小正确方案是复用现有 query mapper，把 response DTO 收回到 owning query package。

## Files to modify

- `.harness/reuse-decisions/teaching-query-timeline-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-teaching-query-timeline-dto-localization-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-teaching-query-timeline-dto-localization-review.md`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/timeline_types.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/timeline.go`
