# Teaching Review Optimization Review

- Review target:
  - Repository: `ctf`
  - Branch: `main`
  - Diff source: local working tree changes for `teaching-review-optimization`
  - Files reviewed:
    - `code/frontend/src/api/teacher/students.ts`
    - `code/frontend/src/features/teacher-student-analysis/model/useTeacherReviewWorkspace.ts`
    - `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
    - `code/frontend/src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.vue`
    - `code/frontend/src/widgets/teacher-student-review-workspace/model/presentation.ts`
    - `code/backend/internal/module/assessment/ports/ports.go`
    - `code/backend/internal/module/assessment/application/commands/report_service.go`
    - `code/backend/internal/module/assessment/infrastructure/report_repository.go`

- Classification check: agree with non-trivial implementation gate. This change spans frontend state ownership, backend read-model alignment, API contract changes, and archive consistency.
- Gate verdict: pass with minor issues

## Findings

### Minor

1. `useTeacherStudentAnalysisPage.ts`
   - Risk: 当前只在 `initialize()` 时从路由 query 回填复盘筛选。浏览器前进/后退如果只变更 `reviewMode`、`reviewResult`、`reviewChallengeId`，页面状态不会即时跟随。
   - Impact: 主要影响可分享/可回退的筛选体验，不影响当前功能正确性。
   - Fix direction: 后续可补一个只监听复盘相关 query 的 watcher，并在值变化时同步 `sessionQuery` 与对应 reload。

## Material findings

无。review 中唯一有直接交互影响的问题是“题目筛选后选项被过滤收窄”，已在本轮修复，通过缓存 `reviewChallengeOptions` 保持可切换性。

## Senior Implementation Assessment

- 当前实现方向是对的：没有继续把状态和模板堆进 `StudentInsightPanel.vue`，而是把复盘工作台拆到 widget/composable，并把筛选编排放在页面模型层。
- 后端归档对齐采取了保守方案：先统一证据查询契约、AWD 事件字段和代理请求 meta，再逐步推进到更彻底的共享构建器。这比现在就做跨模块大重构更稳。
- 仍有一个架构余量：`assessment` 现在直接依赖了 `teaching_readmodel` 的 `EvidenceQuery`。这次改动成本低、风险也可控，但如果后续继续扩展，建议再抽一个真正共享的 evidence query contract，避免读模型包反向渗透。

## Required Re-validation

- `cd code/frontend && npm run test:run -- src/api/__tests__/teacher.test.ts src/widgets/teacher-student-review-workspace/model/presentation.test.ts src/widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd code/backend && go test ./internal/module/assessment/... ./internal/app -count=1`

## Residual Risk

- 复盘筛选的路由 query 目前只做了写回和初始化回填，还没有做到对浏览器历史导航的全量响应。
- 归档证据与实时证据已经对齐了主要事件类型和字段口径，但尚未完全共享同一个 builder，后续继续扩展事件源时仍需要注意双边同步。
