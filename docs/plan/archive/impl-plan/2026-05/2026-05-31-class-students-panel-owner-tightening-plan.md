# Class Students Panel Owner Tightening 计划

## Objective

- 把班级工作区的 `panel` query owner 从 `ClassStudentsPage.vue` 收回 `useClassStudentsPage.ts`。
- 保持 alias route canonicalize、insight window query owner 和班级工作区加载 owner 不变。

## Non-goals

- 不改 `useClassWorkspaceSection.ts` 的 alias route 规则。
- 不改 `getClassStudents()`、`getClassReview()`、`getClassSummary()`、`getClassTrend()` 的加载与错误处理。
- 不扩到 `student-analysis-workspace`。

## Source Inputs

- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassWorkspaceSection.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformClassStudents.test.ts`

## Plan Review Result

- 这条线不需要新造 panel helper，直接复用 `useRouteQueryTabs()` 更合理，因为 page model 本身已经持有 route-aware owner，继续把 panel query 放在同一个 owner 里更稳。
- 最小改动是 `useClassStudentsPage.ts` 接入 `useRouteQueryTabs()`，`ClassStudentsPage.vue` 改成消费 `activeTab / selectTab / setTabButtonRef / handleTabKeydown`。

## Task Slices

### Slice 1: 收回 page model 的 panel owner

- 目标：让 `useClassStudentsPage.ts` 复用 `useRouteQueryTabs()` 持有 `overview/trend/students/review/insight/action` 的 panel query。
- 风险：
  - 如果和 alias route canonicalize 的 query 规则打架，初始化会出现重复 replace 或默认 panel 漂移。

### Slice 2: 让 ClassStudentsPage 退回纯展示壳

- 目标：删除 `ClassStudentsPage.vue` 中的 `useUrlSyncedTabs()`，通过 props / emits 消费外部 tab 状态和切换行为。
- 风险：
  - 如果 contract 没收清，click / keyboard tab 行为会散落在 UI 壳里。

### Slice 3: 补 teacher / platform 两侧护栏测试

- 目标：同时补上 raw-source 护栏和 panel query 初始恢复 / 点击回写断言。
- 风险：
  - 如果只测点击不测初始 hydrate，后续 `?panel=review` 之类的回归仍可能漏掉。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision class-students-panel-owner-tightening`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherClassStudents.test.ts src/pages/platform/__tests__/PlatformClassStudents.test.ts`
- `git diff --check -- .harness/reuse-decisions/class-students-panel-owner-tightening.md docs/plan/impl-plan/2026-05-31-class-students-panel-owner-tightening-plan.md code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts code/frontend/src/pages/platform/__tests__/PlatformClassStudents.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- `useClassStudentsPage.ts` 是否成为唯一 `panel` query owner。
- `ClassStudentsPage.vue` 是否已退回纯展示壳，不再直接做 query 同步。
- `?panel=review` 初始恢复、点击 tab 后 query 回写，以及 alias route canonicalize 是否保持正确。

## Rollback / Recovery

- 如果 props / emits 命名不清楚，可以继续调整契约命名，但不能回退到 UI 壳直接持有 `useUrlSyncedTabs()`。
