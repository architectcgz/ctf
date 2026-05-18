# Reuse Decision

## Change type

api / query / report / runtime / contract / practice / contest

## Existing code searched

- `code/backend/internal/module/teaching_query/**/*.go`
- `code/backend/internal/module/runtime/**/*.go`
- `code/backend/internal/module/assessment/**/*teacher_awd_review*.go`
- `code/backend/internal/module/practice/**/*submission*.go`
- `code/backend/internal/module/contest/**/*submission*.go`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/dto/timeline.go`
- `code/backend/internal/dto/teacher_awd_review.go`
- `code/backend/internal/dto/submission.go`
- `code/backend/internal/dto/report.go`

## Similar implementations found

- `practice` 已经把学生侧 `progress / timeline` HTTP DTO 收回模块边界
- `practice` 和 `contest` 都已经用各自 `api/http` + `application/commands` 承接 submission 流程
- `assessment` recommendation / skill profile 已经使用 contract + HTTP 本地 DTO 模式
- `ops` dashboard 已经把 query snapshot 与 HTTP DTO 分层

## Decision

refactor_existing

## Reason

这批类型并不是一个真正共享的统一领域，而是历史上堆在 `teacher.go` / `timeline.go` / `teacher_awd_review.go` / `submission.go` / `report.go` 里的多 owner 聚合。最小正确方案不是继续保留全局桶，而是按 `teaching_query`、`runtime`、`assessment`、`practice`、`contest` 五个 owner 拆回模块边界。

## Target owners

- `code/backend/internal/module/teaching_query/**`
- `code/backend/internal/module/runtime/**`
- `code/backend/internal/module/assessment/**`
- `code/backend/internal/module/practice/**`
- `code/backend/internal/module/contest/**`
- `code/backend/internal/app/**test.go`
- `code/backend/cmd/seed-teaching-review-data/*.go`
- `docs/plan/impl-plan/2026-05-17-teacher-aggregate-dto-localization-implementation-plan.md`
- `.harness/reuse-decisions/teacher-aggregate-dto-localization.md`

## Legacy DTO sources to delete after localization

- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/dto/timeline.go`
- `code/backend/internal/dto/teacher_awd_review.go`
- `code/backend/internal/dto/submission.go`
- `code/backend/internal/dto/report.go`

## Files to modify

- `.harness/reuse-decisions/teacher-aggregate-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-teacher-aggregate-dto-localization-implementation-plan.md`
- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/cmd/seed-teaching-review-data/main_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/dto/report.go`
- `code/backend/internal/dto/submission.go`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/module/assessment/api/http/report_handler.go`
- `code/backend/internal/module/assessment/api/http/response_mapper_gen.go`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- `code/backend/internal/module/assessment/application/commands/report_output.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/assessment/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/assessment/architecture_test.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/api/http/submission_handler.go`
- `code/backend/internal/module/practice/api/http/submission_types.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/practice/application/commands/submission_history_service.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/application/commands/submission_output.go`
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/architecture_test.go`
- `code/backend/internal/module/teaching_query/api/http/handler_test.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_response_mapper.go`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/overview_service.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/teacher_response_types.go`
- `code/backend/internal/module/teaching_query/architecture_test.go`
- `code/backend/internal/module/teaching_query/contracts/class_insight.go`
- `code/backend/internal/module/teaching_query/contracts/teacher_response.go`
