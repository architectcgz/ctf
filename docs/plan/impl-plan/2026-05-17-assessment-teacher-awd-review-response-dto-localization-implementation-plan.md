# assessment teacher AWD review response DTO 模块内化实现方案

## Objective

把教师侧 AWD review 的 contest list / archive 响应与导出归档模型从 `internal/dto/teacher_awd_review.go` 收回 `assessment/application/queries`，保持教师侧查询和导出外部契约不变。

## Non-goals

- 不改 `teacher_awd_review` 的 request DTO
- 不改教师侧路由、权限、查询参数与分页语义
- 不处理 `teacher.go`、`timeline.go` 或 runtime/teaching_query 的其他教师侧 DTO

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/plan/impl-plan/2026-05-17-teacher-aggregate-dto-localization-implementation-plan.md`
- `docs/plan/impl-plan/2026-05-17-assessment-teacher-awd-review-request-dto-localization-implementation-plan.md`
- `code/backend/internal/module/assessment/application/queries/{teacher_awd_review_service.go,response_mapper.go,response_mapper_gen.go}`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/application/commands/{awd_review_export_builder.go,awd_review_export_renderer.go,report_service.go,report_service_test.go}`
- `code/backend/internal/dto/teacher_awd_review.go`

## Task Slices

1. 在 `assessment/application/queries` 定义本地 AWD review response / archive 类型
   - 包括 contest page、archive、scope、overview、round、selected round、team/service/attack/traffic

2. 收口 query service 与 mapper
   - `TeacherAWDReviewService` 改为返回模块内类型
   - 重新生成 `response_mapper_gen.go`

3. 收口 assessment 模块内部消费者
   - `assessment/api/http` handler interface
   - `assessment/runtime` 包装层
   - `assessment/application/commands` 的 export builder / renderer / report service

4. 调整测试与删除全局 DTO
   - `full_router_state_matrix_integration_test.go` 改为解码模块内类型
   - `report_service_test.go` 改为构造模块内 archive
   - 若无剩余引用则删除 `internal/dto/teacher_awd_review.go`

## Expected Changes

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

## Validation

- `go generate ./internal/module/assessment/application/queries`
- `go test ./internal/module/assessment/application/queries -run 'TestTeacherAWDReviewService' -count=1`
- `go test ./internal/module/assessment/application/commands -run 'TestReportServiceCreateAWDReviewArchiveExportStartsProcessingTask|TestReportServiceCreateAWDReviewReportExportMarksReady|TestWriteJSONReportCreatesJSONFile' -count=1`
- `go test ./internal/app -run 'TestFullRouter_TeacherAWDReviewExportStateMatrix|TestFullRouter_ContestAndReviewArchiveExportStateMatrix' -count=1`

## Review Focus

- AWD review 响应/归档模型 owner 是否已经回到 `assessment/application/queries`
- 导出 ZIP/PDF 与教师侧 archive JSON 字段是否保持不变
- 删除全局 `teacher_awd_review.go` 后是否没有把 `dto` 漏回 assessment 内部链路
