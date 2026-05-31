# Reuse Decision

## Change type
frontend refactor / student analysis panel owner tightening

## Existing code searched
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/pages/__tests__/studentAnalysisRouteTestSupport.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `useRouteQueryTabs.ts` 已经是仓库里共享的 query-tab route owner，`class-students`、`teacher-dashboard`、`skill-profile` 最近都在 page model 里直接复用它承接 `panel` query。
- `useStudentAnalysisPage.ts` 当前已经持有 `useRouteQueryTransport()`、`useRouteNavigationTransport()`、review workspace query sync 和页面级导航 target owner，继续把 workspace tab query 收到同一 owner 更一致。
- `StudentAnalysisPage.vue` 当前仍直接使用 `useUrlSyncedTabs()`，让 UI 壳自己持有 `overview/recommendations/writeups/evidence/timeline` 的 panel query owner。

## Decision
refactor_existing

## Reason
当前最小正确切片不是继续拆 student insight 子区块，而是把 `panel` query owner 从 `StudentAnalysisPage.vue` 收回 `useStudentAnalysisPage.ts`，并直接复用已有 `useRouteQueryTabs()`：

- `useStudentAnalysisPage.ts` 继续统一持有 route query transport、review workspace query sync、导航 route target 和 student analysis workspace 的 panel query。
- `StudentAnalysisPage.vue` 退回纯 props / emits 展示壳，不再自己读写 `panel` query。
- 不新增新的 panel helper；这条线已有 `useRouteQueryTabs()`，直接复用比再造一层 student-analysis 专用 helper 更小，也和最近几轮 page owner 收口保持一致。

本轮不做：

- 不改 student progress / profile / recommendations / timeline / evidence / attack sessions 的加载与错误处理。
- 不改 `useStudentAnalysisReviewQuerySync.ts` 当前的 review workspace filter query owner。
- 不扩到 `instance`、`review archive` 或其他 student-analysis 子 feature 的结构拆分。

## Files to modify
- `.harness/reuse-decisions/student-analysis-panel-owner-tightening.md`
- `docs/plan/impl-plan/2026-05-31-student-analysis-panel-owner-tightening-plan.md`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/pages/__tests__/studentAnalysisRouteTestSupport.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `useStudentAnalysisPage.ts` 会成为 student analysis workspace `panel` query 的唯一 owner。
- `StudentAnalysisPage.vue` 不再直接依赖 `useUrlSyncedTabs()`。
- `student-analysis-workspace` 会对齐到当前仓库的 route-aware page owner 模式：`page model + shared route query owner + 纯 UI shell`。
