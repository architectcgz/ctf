# Teacher Dashboard Panel Owner Tightening 计划

## Objective

- 把教师概览 `panel` query owner 从 `TeacherDashboardPage.vue` 收回 `useDashboardPage.ts`。
- 保留 UI 壳的 tab 键盘导航和现有展示结构。

## Non-goals

- 不改教师概览数据加载、班级管理 route target 或各子 panel 内容。
- 不调整 `useTeacherOverviewData()` 与 metric builders。
- 不扩到其他 teacher 页面。

## Source Inputs

- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/shared/lib/keyboard/useTabKeyboardNavigation.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`

## Plan Review Result

- `TeacherDashboardPage.vue` 里真正属于 UI 的是 tab 列表和键盘焦点移动；route query 读写不该继续留在这里。
- 最小改动是 page model 接入 `useRouteQueryTransport()`，UI 改为消费 `activePanel` / `switchPanel`。

## Task Slices

### Slice 1: 提取 teacher dashboard panel helper

- 目标：新增纯 helper，统一解析 panel 并构建 query。
- 风险：
  - 如果 `overview` 归一规则不对，默认页签 URL 会漂移。

### Slice 2: 收回 page model owner

- 目标：让 `useDashboardPage.ts` 负责 query 读写。
- 风险：
  - 如果 UI 和 model 同时持有 active tab，会形成双 owner。

### Slice 3: 保留 UI 壳键盘交互 owner

- 目标：让 `TeacherDashboardPage.vue` 只保留 `useTabKeyboardNavigation()`。
- 风险：
  - 如果事件 contract 没收清，可能影响键盘切 tab 和 query 同步。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision teacher-dashboard-panel-owner-tightening`
- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `git diff --check -- .harness/reuse-decisions/teacher-dashboard-panel-owner-tightening.md docs/plan/impl-plan/2026-05-31-teacher-dashboard-panel-owner-tightening-plan.md code/frontend/src/features/teacher/dashboard/model/teacherDashboardPanelRoute.ts code/frontend/src/features/teacher/dashboard/model/index.ts code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue code/frontend/src/pages/teacher/TeacherDashboardRoutePage.vue code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- `useDashboardPage.ts` 是否成为教师概览唯一 panel query owner。
- `TeacherDashboardPage.vue` 是否已退回纯 UI shell，不再直接做 route query 同步。
- 初始 `?panel=portrait` 和点击 tab 后的 query 回写是否保持正确。

## Rollback / Recovery

- 如果交互 contract 名称不够清楚，可以调整 props / emits，但不能回退到 UI 壳直接持有 query owner。
