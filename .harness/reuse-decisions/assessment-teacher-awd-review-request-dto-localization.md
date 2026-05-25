# Reuse Decision

## Change type

api / handler / mapper

## Existing code searched

- `code/backend/internal/module/assessment/api/http/*.go`
- `code/backend/internal/dto/teacher_awd_review.go`
- `code/backend/internal/app/router_test.go`

## Similar implementations found

- `identity/api/http` 与 `assessment/api/http` 已经把模块自有 request DTO 收回 `api/http`
- `teaching_query` 的教师侧 handler 直接绑定本地 query DTO，再映射为 application input

## Decision

refactor_existing

## Reason

教师侧 AWD review 的 contest list / archive request DTO 只被 `assessment/api/http` 的 handler 和 request mapper 消费，不满足继续留在全局 `internal/dto` 的共享条件。最小正确方案是复用现有 request mapper，把请求类型直接收回 handler 所属模块。

## Files to modify

- `.harness/reuse-decisions/assessment-teacher-awd-review-request-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-17-assessment-teacher-awd-review-request-dto-localization-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-assessment-teacher-awd-review-request-dto-localization-review.md`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_request_types.go`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- `code/backend/internal/module/assessment/api/http/request_mapper.go`
- `code/backend/internal/module/assessment/api/http/request_mapper_gen.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/dto/teacher_awd_review.go`
