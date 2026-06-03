# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 10

## Existing code searched
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterconteststate/contest_state.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouterconteststate/contest_state.go` 已经证明可以把长 HTTP 状态矩阵 owner 迁出，同时把数据库 seed 和等待 helper 留在 `internal/app`。
- `full_router_state_matrix_integration_test.go` 里只有 `TestFullRouter_ReportPreviewAndDownloadStateMatrix` 适合本轮单独迁移；另外两个 module builder 测试和通用 helper owner 暂不适合一起抽动。

## Decision
refactor_existing

## Reason
`full_router_state_matrix_integration` 本轮选择更小且更安全的切片：

- 在 `code/backend/tests/system/http/fullrouterreportstate/` 建立可导入的场景断言 package
- 只迁 `TestFullRouter_ReportPreviewAndDownloadStateMatrix`
- module builder 测试和 `createReportRecord` / `waitForReportStatus` 等共享 helper 继续留在 `internal/app`

这样能继续缩小 `internal/app` 的系统测试 owner，同时避免把本轮 scope 扩成 helper owner 抽取。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase10.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase10-plan.md`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/tests/system/http/fullrouterreportstate/*.go`
- `code/backend/tests/README.md`

## After implementation
- `ReportPreviewAndDownloadStateMatrix` 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `full_router_state_matrix_integration_test.go` 暂时仍保留 module builder 测试和共享 helper owner。
- 报表相关共享 helper 是否进一步抽取，后续单独切片处理。
