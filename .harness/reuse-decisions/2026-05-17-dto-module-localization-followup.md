# Reuse Decision

## Change type

contract / api / mapper / query / report

## Existing code searched

- `code/backend/internal/module/assessment/**/*.go`
- `code/backend/internal/module/challenge/**/*writeup*.go`
- `code/backend/internal/module/practice/**/*manual_review*.go`
- `code/backend/internal/module/teaching_query/**/*.go`
- `code/backend/internal/teaching/classreview/*.go`
- `code/backend/internal/dto/report.go`
- `code/backend/internal/dto/submission.go`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/dto/writeup.go`

## Similar implementations found

- `code/backend/internal/module/assessment/api/http/*request*.go` 已经使用模块内 HTTP request DTO
- `code/backend/internal/module/practice/contracts/*.go` 已经承接模块内事件/契约
- `code/backend/internal/module/challenge/contracts/*.go` 已经承接 writeup / query 契约
- `code/backend/internal/module/teaching_query/application/queries/*.go` 已经承接教师侧 query response model

## Decision

refactor_existing

## Reason

这批 DTO 并不构成可复用的统一共享模型，而是历史上堆在 `internal/dto` 里的多 owner 类型。继续新增或保留全局桶只会放大跨模块耦合。这里沿用仓库现有模式，把 HTTP request DTO 收回 `api/http`，把模块 response/query DTO 收回各自 `contracts` 或 `application/queries`，同时保留真正仍被多链路复用的全局 DTO。

## Files to modify

- `code/backend/internal/dto/report.go`
- `code/backend/internal/dto/submission.go`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/dto/writeup.go`
- `code/backend/internal/module/assessment/api/http/report_request_types.go`
- `code/backend/internal/module/assessment/api/http/report_handler.go`
- `code/backend/internal/module/assessment/api/http/request_mapper.go`
- `code/backend/internal/module/assessment/api/http/request_mapper_gen.go`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/challenge/api/http/request_mapper.go`
- `code/backend/internal/module/challenge/api/http/request_mapper_gen.go`
- `code/backend/internal/module/challenge/api/http/writeup_handler.go`
- `code/backend/internal/module/challenge/application/queries/response_mapper_goverter.go`
- `code/backend/internal/module/challenge/application/queries/response_mapper_goverter_gen.go`
- `code/backend/internal/module/challenge/domain/mappers.go`
- `code/backend/internal/module/challenge/domain/response_mapper_goverter_assign.go`
- `code/backend/internal/module/challenge/contracts/writeup.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/practice/contracts/manual_review.go`
- `code/backend/internal/module/teaching_query/contracts/class_insight.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_response_mapper.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_response_types.go`
- `code/backend/internal/teaching/classreview/review.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/cmd/seed-teaching-review-data/main.go`

## After implementation

- 本次只补任务级 reuse decision，不覆盖已有草稿。
- 若后续继续按模块批量清理 `internal/dto`，再把稳定模式回收到 `harness/reuse/index.yaml` 或 `harness/reuse/history.md`。
