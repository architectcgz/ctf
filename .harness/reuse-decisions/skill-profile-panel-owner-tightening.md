# Reuse Decision

## Change type
frontend refactor / skill profile panel owner tightening

## Existing code searched
- `code/frontend/src/features`
- `code/frontend/src/pages/profile/SkillProfileRoutePage.vue`
- `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform/user-management/model/useUserGovernancePanelRoute.ts`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/contests/model/useContestManagePanelRoute.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardPanelRoute.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/shared/lib/keyboard/useTabKeyboardNavigation.ts`
- `code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `usePlatformUserManagePage.ts` 已通过 `useRouteQueryTransport()` 成为目录 panel query 的唯一 owner。
- `useContestManagePage.ts` 与 `useContestManagePanelRoute.ts` 已把 `ContestManage` 的 `panel` query 从 UI 壳收回 page model。
- `useDashboardPage.ts` 与 `teacherDashboardPanelRoute.ts` 也已证明“page model 持有 query owner，UI 只保留 tab 展示 / 键盘交互”这条模式可复用。
- `SkillProfileRoutePage.vue` 当前仍直接使用 `useUrlSyncedTabs()`，让 route page 自己持有 `analysis/weakness/recommendations` 的 panel query owner。

## Decision
refactor_existing

## Reason
当前最小正确切片不是重写能力画像页面，而是把 `panel` query owner 从 route page 收回 page model：

- `useSkillProfilePage.ts` 承接 `panel` query 的读取与切换。
- 新增纯 helper，统一解析 `analysis/weakness/recommendations` 并构建 query。
- `SkillProfileRoutePage.vue` 退回 route 组合层，只消费 `activePanel`、`switchPanel` 和 tab 键盘交互 helper。

本轮不做：

- 不改学生 / 教师画像数据加载与推荐靶场 workflow。
- 不改 `SkillProfileWorkspaceShell.vue` 的布局、视觉和推荐卡片交互。
- 不扩到 `StudentAnalysis`、`ClassStudents` 或 challenge detail 页面。

## Files to modify
- `.harness/reuse-decisions/skill-profile-panel-owner-tightening.md`
- `docs/plan/impl-plan/2026-05-31-skill-profile-panel-owner-tightening-plan.md`
- `code/frontend/src/features/skill-profile/model/skillProfilePanelRoute.ts`
- `code/frontend/src/features/skill-profile/model/index.ts`
- `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
- `code/frontend/src/pages/profile/SkillProfileRoutePage.vue`
- `code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `useSkillProfilePage.ts` 会成为能力画像页面 `panel` query 的唯一 owner。
- `SkillProfileRoutePage.vue` 不再直接依赖 `useUrlSyncedTabs()`。
- 能力画像页面会对齐到最近几笔 panel owner 收口模式：`page model + shared route transport + 纯 helper + UI keyboard helper`。
