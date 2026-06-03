# 2026-06-03 backend test architecture phase 15 plan

## Objective

- 修复 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 的既有 PDF 断言失败。
- 保持改动只限于 full-router 测试层，不改 AWD review PDF 生产逻辑。

## Non-goals

- 本轮不迁移 `TeacherAWDReviewExportStateMatrix` owner。
- 本轮不改 `RenderAWDReviewReportPDF` 的标题、章节文案或布局。
- 本轮不把 sqlite full-router 测试改造成 PostgreSQL 测试入口。

## Inputs

- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
- `code/backend/internal/module/assessment/application/commands/awd_review_export_renderer.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `.harness/reuse-decisions/backend-test-architecture-phase15.md`

## Root cause summary

- 当前渲染器输出中文报告标题 / 章节名。
- 模块级 PDF 渲染测试 helper 已支持 UTF-16 BE 文本匹配。
- full-router 测试仍断旧英文 token，且自己的 PDF helper 不支持 UTF-16 BE，导致即使报告里有中文 section heading 也提取不到。

## Working design

### Assertion strategy

- 保留：
  - `Content-Type` / `Content-Disposition`
  - `%PDF` 文件头
- 调整：
  - 对齐模块级 helper，补齐 UTF-16 BE 匹配
  - 去掉旧英文 token
  - 改为当前样本里稳定可提取的中文 section heading，例如 `摘要`、`选中轮次摘要`

### Boundary choice

- 只修测试断言。
- 不改报告生成器，不改 seed 数据形态。

## Task slices

### Slice 1：reuse-first 门禁

- Goal：补齐 phase 15 reuse decision / implementation plan，并通过 startup gate。
- Validation：
  - `bash scripts/check-task-intake.sh --reuse-decision backend-test-architecture-phase15`

### Slice 2：修复 AWD review export 既有失败

- Goal：让 `TestFullRouter_TeacherAWDReviewExportStateMatrix` 与当前 PDF 文案/内容口径一致。
- Touched files：
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestFullRouter_TeacherAWDReviewExportStateMatrix' -count=1`

### Slice 3：回归与 workflow gate

- Goal：确认 router / practice 相关回归和 workflow 约束未被破坏。
- Validation：
  - `cd code/backend && go test ./internal/app -run 'TestPracticeFlow_|TestRouter' -count=1`
  - `bash scripts/check-workflow-complete.sh`

## Expected change surface

- 一个 full-router PDF helper
- 一个 full-router 测试函数的 PDF 断言

## Data / API / compatibility impact

- 无生产数据、API 或行为变化。
- 风险主要在于：
  - helper 仍漏掉当前 PDF 的编码形态，导致断言继续假阴性
  - 选错 PDF 可提取 token，导致断言仍不稳定

## Review fit check

- Root cause 已定位：过时英文断言 + full-router helper 落后于模块级 helper。
- 修复最小：只改 full-router 测试层，不动生产逻辑。
- 验证闭环清晰：先定向失败用例，再跑 router 回归。

## Rollback / recovery

- 纯测试断言调整，可直接回退到当前 `full_router_teacher_state_matrix_test.go` 版本。
