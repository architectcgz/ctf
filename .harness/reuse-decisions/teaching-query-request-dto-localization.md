# Reuse Decision

## Change type

teaching query api / handler / query service

## Existing code searched

- `code/backend/internal/module/teaching_query/api/http/*.go`
- `code/backend/internal/module/teaching_query/application/queries/*.go`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- `code/backend/internal/module/assessment/application/queries/teacher_awd_review_input.go`
- `code/backend/internal/dto/teacher.go`

## Similar implementations found

- `assessment/api/http` 已经把教师侧 AWD review request DTO 收回 handler 所属模块
- `runtime/api/http` 已经把教师实例列表 query 收回 HTTP 层，并把跨模块 contract 收口到 owning module

## Decision

refactor_existing

## Reason

这批教师侧 request DTO 只服务于 `teaching_query` 的 HTTP 入口和对应 query service，不满足继续留在全局 `internal/dto` 的共享条件。最小正确方案是复用现有 handler + thin mapper 结构，把：

- 带 `form` / `binding` / `time_format` tag 的 HTTP 输入收回 `teaching_query/api/http`
- 纯应用层输入收回 `teaching_query/application/queries`

这样可以避免把 HTTP 语义带入 application，同时不改外部 query 参数、校验口径和响应结构。

## Files to modify

- `.harness/reuse-decisions/teaching-query-request-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-teaching-query-request-dto-localization-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-teaching-query-request-dto-localization-review.md`
- `code/backend/internal/module/teaching_query/api/http/handler.go`
- `code/backend/internal/module/teaching_query/api/http/request_types.go`
- `code/backend/internal/module/teaching_query/api/http/request_mapper.go`
- `code/backend/internal/module/teaching_query/api/http/handler_test.go`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/input_types.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service_test.go`
- `code/backend/internal/dto/teacher.go`
