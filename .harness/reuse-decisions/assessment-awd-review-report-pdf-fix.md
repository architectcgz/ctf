# Reuse Decision

## Change type
service

## Existing code searched
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_pdf_fonts.go`
- `code/backend/internal/module/assessment/application/reportassets/fonts.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`

## Similar implementations found
- `code/backend/internal/module/assessment/application/commands/report_service.go`
  - 已有通用 `newReportPDF`、标题、摘要、表头、表格行等 PDF 骨架
- `code/backend/internal/module/assessment/application/reportassets/fonts.go`
  - 适合作为 report 导出静态字体资源的专门 owner
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
  - 已有 AWD 赛事复盘 PDF 渲染，但内容层次和默认轮次策略仍偏薄
- `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
  - 已有导出 builder，可继续承担“是否需要默认选轮次”的导出装配职责

## Decision
extend_existing

## Reason
这次不是新增第二套教师导出系统，而是把 AWD 报告继续收回现有 report 骨架。最小正确方案是：
- 继续复用 `report_service.go` 里的通用 PDF 骨架，而不是新起独立 renderer 基类
- 把字体二进制资源从 `commands` 目录抽到专门的 `reportassets` owner，避免 command 包继续承担静态资源归属
- 扩展 `awd_review_export_builder.go`，在未显式指定轮次时自动补一个最值得复盘的轮次，避免报告退化成只有概览
- 扩展 `awd_review_export_renderer.go`，补齐关键样本和分析段落，而不是再建并行导出链路
- 扩展现有 `report_service_test.go`，把“只生成 PDF 文件”提升为“报告内容和字体装配可用”的回归保护

## Files to modify
- `code/backend/internal/module/assessment/application/commands/report_pdf_fonts.go`
- `code/backend/internal/module/assessment/application/reportassets/fonts.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_builder.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `docs/architecture/features/AWD教师复盘归档与报告导出设计.md`

## After implementation
- 如果这次默认选轮次和报告段落结构可复用，再考虑补到 `harness/reuse/history.md`。
