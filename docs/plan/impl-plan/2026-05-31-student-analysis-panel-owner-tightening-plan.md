# Student Analysis Panel Owner Tightening 计划

## Objective

- 把学员分析工作区的 `panel` query owner 从 `StudentAnalysisPage.vue` 收回 `useStudentAnalysisPage.ts`。
- 保持 review workspace query sync、导航 route target owner 和页面级数据加载 owner 不变。

## Non-goals

- 不改 `useStudentAnalysisReviewQuerySync.ts` 的 `reviewMode/reviewResult/reviewChallengeId` query 规则。
- 不改 student analysis 的数据加载、错误态和 review workspace 刷新策略。
- 不扩到 `instance`、`student review archive` 或更深层 insight section 拆分。

## Source Inputs

- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/pages/__tests__/studentAnalysisRouteTestSupport.ts`

## Plan Review Result

- 这条线不需要新造 panel helper，直接复用 `useRouteQueryTabs()` 更合理，因为 `useStudentAnalysisPage.ts` 已经是 route-aware page owner，继续把 workspace tab query 放在同一个 owner 里更稳。
- 最小改动是 `useStudentAnalysisPage.ts` 接入 `useRouteQueryTabs()`，`StudentAnalysisPage.vue` 改成消费 `activeWorkspaceTab / selectWorkspaceTab / setWorkspaceTabButtonRef / handleWorkspaceTabKeydown`。

## Task Slices

### Slice 1: 收回 page model 的 panel owner

- 目标：让 `useStudentAnalysisPage.ts` 复用 `useRouteQueryTabs()` 持有 `overview/recommendations/writeups/evidence/timeline` 的 panel query。
- 风险：
  - 如果和现有 review workspace query sync 混在一起，可能出现 tab query 与 review query 互相覆盖或回写丢字段。

### Slice 2: 让 StudentAnalysisPage 退回纯展示壳

- 目标：删除 `StudentAnalysisPage.vue` 中的 `useUrlSyncedTabs()`，通过 props / emits 消费外部 tab 状态和切换行为。
- 风险：
  - 如果 contract 没收清，click / keyboard tab 行为仍会散落在 UI 壳里。

### Slice 3: 补 teacher / platform 两侧护栏测试

- 目标：补上 raw-source 护栏和 panel query 初始恢复 / 点击回写断言。
- 风险：
  - 如果只测 source 不测运行态，后续 `?panel=writeups` 之类的回归仍可能漏掉。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision student-analysis-panel-owner-tightening`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `git diff --check -- .harness/reuse-decisions/student-analysis-panel-owner-tightening.md docs/plan/impl-plan/2026-05-31-student-analysis-panel-owner-tightening-plan.md code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts code/frontend/src/pages/__tests__/studentAnalysisRouteTestSupport.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- `useStudentAnalysisPage.ts` 是否成为唯一 `panel` query owner。
- `StudentAnalysisPage.vue` 是否已退回纯展示壳，不再直接做 query 同步。
- `?panel=writeups` 初始恢复、点击 tab 后 query 回写，以及 review workspace query sync 是否保持正确。

## Rollback / Recovery

- 如果 props / emits 命名不清楚，可以继续调整契约命名，但不能回退到 UI 壳直接持有 `useUrlSyncedTabs()`。
