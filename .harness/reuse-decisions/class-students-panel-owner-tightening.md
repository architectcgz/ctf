# Reuse Decision

## Change type
frontend refactor / class students panel owner tightening

## Existing code searched
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassWorkspaceSection.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformClassStudents.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `useRouteQueryTabs.ts` 已经是当前仓库里共享的 query-tab route owner，`contest-detail` 也在 page model 里直接复用它承接 panel query。
- `usePlatformUserManagePage.ts`、`useContestManagePage.ts`、`useDashboardPage.ts` 已证明“panel/query owner 收回 page model，UI 壳退回纯展示契约”是当前收口方向。
- `ClassStudentsPage.vue` 当前仍直接使用 `useUrlSyncedTabs()`，让 UI 壳自己持有 `overview/trend/students/review/insight/action` 的 panel query owner。

## Decision
refactor_existing

## Reason
当前最小正确切片不是改班级工作台的数据加载，而是把 panel query owner 从 UI 壳收回 page model，并直接复用已有 `useRouteQueryTabs()`：

- `useClassStudentsPage.ts` 继续统一持有 alias route canonicalize、`from_date/to_date` insight window query、班级工作区加载和 panel query。
- `ClassStudentsPage.vue` 退回纯 props / emits 展示壳，不再自己读写 `panel` query。
- 不新增新的 panel helper；这条线已有 `useRouteQueryTabs()`，继续复用比再造一套纯 helper 更小、更一致。

本轮不做：

- 不改 `useClassWorkspaceSection.ts` 的 alias route canonical target 规则。
- 不改学生目录加载、班级复盘 / 趋势 / summary 数据请求和 stale request 处理。
- 不扩到 `student-analysis-workspace`。

## Files to modify
- `.harness/reuse-decisions/class-students-panel-owner-tightening.md`
- `docs/plan/impl-plan/2026-05-31-class-students-panel-owner-tightening-plan.md`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformClassStudents.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `useClassStudentsPage.ts` 会成为班级工作区 `panel` query 的唯一 owner。
- `ClassStudentsPage.vue` 不再直接依赖 `useUrlSyncedTabs()`。
- `class-students-workspace` 会对齐到当前仓库的 route-aware page owner 模式：`page model + shared route query owner + 纯 UI shell`。
