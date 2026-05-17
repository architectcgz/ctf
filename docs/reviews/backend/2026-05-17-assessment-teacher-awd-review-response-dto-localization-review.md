# Assessment Teacher AWD Review Response DTO Localization Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 assessment teacher AWD review response DTO 模块内化相关 diff，重点覆盖 `assessment/application/queries`、`assessment/runtime`、`assessment/application/commands`、`full_router_state_matrix_integration_test.go` 与删除 `internal/dto/teacher_awd_review.go`
- Files reviewed:
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
  - `docs/architecture/backend/04-api-design.md`
  - `docs/plan/impl-plan/2026-05-17-assessment-teacher-awd-review-response-dto-localization-implementation-plan.md`
  - `.harness/reuse-decisions/assessment-teacher-awd-review-response-dto-localization.md`
- Classification check: agree with pipeline，属于 non-trivial backend refactor + review gate
- Initial gate verdict: pass with minor issues

## Current Status

- 2026-05-17 补充修复状态：已完成。独立 review 没有发现 blocker；补充项主要集中在导出兼容性验证深度不足，当前已新增 `TestRenderAWDReviewArchiveZipPreservesTeacherAWDReviewJSONFields`、`TestRenderAWDReviewReportPDFIncludesSelectedRoundSummary`，并加强 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 对 `manifest.json`、`selected-round.json` 和 PDF 文本内容的断言。

## Findings

- no findings

## Validation Evidence

- `cd code/backend && go test ./internal/module/assessment/application/queries -run 'TestTeacherAWDReviewService' -count=1`
- `cd code/backend && go test ./internal/module/assessment/application/commands -run 'TestReportServiceCreateAWDReviewArchiveExportStartsProcessingTask|TestReportServiceCreateAWDReviewReportExportRejectsRunningContest|TestWriteJSONReportCreatesJSONFile|TestRenderAWDReviewArchiveZipPreservesTeacherAWDReviewJSONFields|TestRenderAWDReviewReportPDFIncludesSelectedRoundSummary' -count=1`
- `cd code/backend && go test ./internal/app -run 'TestFullRouter_TeacherAWDReviewExportStateMatrix|TestFullRouter_ContestAndReviewArchiveExportStateMatrix' -count=1`

## Final Review Verdict

- Gate verdict after follow-up tests: pass
- 结论：教师侧 AWD review 的 contest list / archive 响应与导出归档模型已收口到 `assessment/application/queries`，旧的全局 `internal/dto/teacher_awd_review.go` 已退出；`GET /api/v1/teacher/awd/reviews`、`GET /api/v1/teacher/awd/reviews/:id`、archive ZIP 与 report PDF 的外部字段和内容语义保持不变，assessment 内部链路统一消费 query snapshot。
