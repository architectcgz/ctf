# Reuse Decision

## Change type
backend test fix / full-router teacher awd review export stale assertion

## Existing code searched
- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app`
  - `code/backend/tests`
  - `docs/plan`

## Similar implementations found
- `RenderAWDReviewReportPDF` 当前标题与章节文案是中文，例如 `教师 AWD 复盘报告`、`选中轮次摘要`。
- `TestRenderAWDReviewReportPDFIncludesSelectedRoundSummary` 的 `pdfContainsText` helper 已支持 UTF-16 BE 检索。
- `TestFullRouter_TeacherAWDReviewExportStateMatrix` 仍在断旧的英文 token，且 full-router 自己的 `fullRouterPDFContainsText` helper 不支持 UTF-16 BE。

## Decision
refactor_existing

## Reason
本轮不改生产逻辑，最小修复是先补齐 full-router PDF helper 的文本提取能力，再同步 full-router 测试断言口径：

- 对齐模块级测试 helper，补上 UTF-16 BE PDF 文本匹配
- 删除过时的英文 PDF token 断言
- 改为断当前报告中稳定可提取的中文 section heading

这样能修复既有失败，同时保持 full-router 仍然覆盖“导出的是有效 PDF 且包含关键结构内容”。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase15.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase15-plan.md`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`

## After implementation
- `TestFullRouter_TeacherAWDReviewExportStateMatrix` 不再依赖已废弃的英文报告文案。
- full-router PDF 断言 helper 能匹配 UTF-16 BE 文本。
- AWD review report export 仍验证 PDF header、下载元信息和关键 section heading。
