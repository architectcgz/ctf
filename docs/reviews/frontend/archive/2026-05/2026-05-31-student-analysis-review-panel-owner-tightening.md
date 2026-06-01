# Student Analysis Panel Owner Tightening Review

## Review target

- Repository: `/home/azhi/workspace/projects/ctf`
- Branch: `main`
- Diff source: working tree uncommitted change set `student-analysis panel owner tightening`
- Files reviewed:
  - `.harness/reuse-decisions/student-analysis-panel-owner-tightening.md`
  - `docs/plan/impl-plan/2026-05-31-student-analysis-panel-owner-tightening-plan.md`
  - `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
  - `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
  - `code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue`
  - `code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue`
  - `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `code/frontend/src/pages/__tests__/studentAnalysisRouteTestSupport.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Validation actually run:
  - `git diff --check -- .harness/reuse-decisions/student-analysis-panel-owner-tightening.md docs/plan/impl-plan/2026-05-31-student-analysis-panel-owner-tightening-plan.md code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts code/frontend/src/pages/__tests__/studentAnalysisRouteTestSupport.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `bash scripts/check-task-intake.sh`

## Classification check

- 结论：同意按 `HARNESS / frontend non-trivial refactor review` 审查。
- 依据：改动集中在 student analysis 的 route/query owner 收口、route page 组合层以及对应护栏测试，没有扩散到 review workspace contract 或数据加载策略本身。

## Gate verdict

- `pass`
- blocker：无

## Findings

- no findings

## Material findings

- 无 material findings。

## Senior implementation assessment

- `useStudentAnalysisPage.ts:27-46,124-143,238-277` 现在同时持有 `panel` query owner、review workspace query owner、route navigation owner 和页面数据装配，owner 边界是清楚的，没有把 query 同步再拆回 UI 壳。
- `StudentAnalysisPage.vue:2-17,264-355` 已退回纯展示壳：tab 列表、props / emits 契约和键盘焦点导航都还留在 UI 层，但 route-aware tab state 已不再直接存在于页面壳里。
- `TeacherStudentAnalysisRoutePage.vue` 与 `PlatformStudentAnalysisRoutePage.vue` 继续只是 route view 组合层，只桥接 feature model 输出到页面组件，没有重新持有 query owner。
- `useRouteQueryTabs.ts:45-67` 的默认行为与这次需求一致：非默认 tab 写入 `panel`，默认 tab 删除 `panel`，且保留其它 query。student analysis 复用共享 helper 比再造一层本地 panel helper 更稳。

## Required re-validation

- 当前无需额外修复后复验。
- 如果后续继续调整这条线，最小充分回归集仍应包括：
  - `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`

## Residual risk

- 本轮没有发现 correctness 或 owner 一致性问题。
- `useRouteQueryTabs()` 的“切回默认 tab 时移除 `panel`”行为由共享 helper实现覆盖，但 student analysis 两侧当前用例没有显式断言这一路径；这更像轻微测试空白，不足以构成当前改动 finding。

## Touched known-debt status

- 已触达并实质收口 `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中 `student analysis panel owner tightening` 这条 debt。
- 当前 touched surface 上，`panel` query owner 已从 `StudentAnalysisPage.vue` 收回 `useStudentAnalysisPage.ts`，UI 壳不再直接持有 `useUrlSyncedTabs()`；这条 owner debt 在本次改动范围内已闭合。
