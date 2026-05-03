# Review Archive Evidence Contract Review

- Review target:
  - Repository: `ctf`
  - Worktree: `ctf/.worktrees/review-archive-evidence-contract`
  - Branch: `codex/review-archive-evidence-contract`
  - Diff source: local working tree changes for review archive evidence contract cleanup
  - Files reviewed:
    - `code/backend/internal/teaching/evidence/query.go`
    - `code/backend/internal/module/teaching_readmodel/ports/query.go`
    - `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
    - `code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`
    - `code/backend/internal/module/assessment/ports/ports.go`
    - `code/backend/internal/module/assessment/application/commands/report_service.go`
    - `code/backend/internal/module/assessment/infrastructure/report_repository.go`
    - `code/backend/internal/module/assessment/application/commands/report_service_test.go`
    - `code/backend/internal/module/assessment/infrastructure/report_repository_test.go`
    - `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
    - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

- Classification check: agree with non-trivial classification. This is a cross-module dependency cleanup with regression-sensitive review archive behavior.
- Gate verdict: pass

## Findings

无。

## Material findings

无。

## Senior Implementation Assessment

- 当前实现是更稳妥的收敛方式：共享的是稳定查询契约 `evidence.Query`，不是把 `assessment` 直接拉进 `teaching_readmodel` 的内部边界。
- 这比现在就强行合并 evidence builder 更合适。builder 全共享仍然是后续架构项，但不应为了修依赖方向而扩大本次变更面。
- 前端“路由 query 回退后复盘筛选不同步”的遗留项已经由页面层 watcher 和回归测试覆盖，不需要重复改动。

## Required Re-validation

- `cd code/backend && go test ./internal/module/assessment/... ./internal/module/teaching_readmodel/... -count=1`
- `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Residual Risk

- evidence builder 本体仍然分布在 readmodel repository 和 review archive repository 中，后续若继续扩展事件源，仍要注意双边同步。
- 本次前端验证使用主工作区现成依赖运行，因为新 worktree 未安装前端依赖；但前端源码在这次收尾中未发生变更，验证目标是确认既有 watcher 与回归测试覆盖仍成立。
