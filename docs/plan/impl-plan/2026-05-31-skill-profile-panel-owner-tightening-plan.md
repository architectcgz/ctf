# Skill Profile Panel Owner Tightening 计划

## Objective

- 把能力画像页面的 `panel` query owner 从 `SkillProfileRoutePage.vue` 收回 `useSkillProfilePage.ts`。
- 保持 route page 薄壳和现有画像 / 推荐 workflow owner 不变。

## Non-goals

- 不改 `getSkillProfile()`、`getRecommendations()` 以及教师代看学员画像的数据加载。
- 不调整 `SkillProfileWorkspaceShell.vue` 的布局、tab 内容和推荐题目跳转。
- 不扩到 `StudentAnalysis`、`ClassStudents` 或 `ChallengeDetail`。

## Source Inputs

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

## Plan Review Result

- `SkillProfileRoutePage.vue` 当前除了 route 组合之外，还额外承担了 `panel` query owner，这和最近几笔 page model 收口方向不一致。
- 最小改动是 `useSkillProfilePage.ts` 接入 `useRouteQueryTransport()`，新增纯 helper 管 panel 解析与构建，route page 只保留 tab 列表与键盘焦点移动。

## Task Slices

### Slice 1: 提取 skill profile panel helper

- 目标：新增纯 helper，统一解析 `analysis/weakness/recommendations` 并构建 query。
- 风险：
  - 如果 `analysis` 默认归一规则不对，默认 URL 会漂移。

### Slice 2: 收回 page model owner

- 目标：让 `useSkillProfilePage.ts` 负责读取与切换 `panel` query。
- 风险：
  - 如果 route page 和 page model 同时保留 active tab，会形成双 owner。

### Slice 3: 让 route page 退回薄组合层

- 目标：删除 `SkillProfileRoutePage.vue` 里的 `useUrlSyncedTabs()`，改为消费 `activePanel` / `switchPanel`。
- 风险：
  - 如果 tab contract 没有收清，点击标签和初始 query 恢复行为会漂。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision skill-profile-panel-owner-tightening`
- `cd code/frontend && npm run test:run -- src/pages/profile/__tests__/SkillProfile.test.ts`
- `git diff --check -- .harness/reuse-decisions/skill-profile-panel-owner-tightening.md docs/plan/impl-plan/2026-05-31-skill-profile-panel-owner-tightening-plan.md code/frontend/src/features/skill-profile/model/skillProfilePanelRoute.ts code/frontend/src/features/skill-profile/model/index.ts code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts code/frontend/src/pages/profile/SkillProfileRoutePage.vue code/frontend/src/pages/profile/__tests__/SkillProfile.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- `useSkillProfilePage.ts` 是否成为唯一 `panel` query owner。
- `SkillProfileRoutePage.vue` 是否已经退回 route 组合层，不再直接做 query 同步。
- 初始 `?panel=recommendations` 和点击顶部标签后的 query 回写是否正确。

## Rollback / Recovery

- 如果 props / emits 或 helper 命名不清楚，可以继续调整命名，但不能回退到 route page 直接持有 query owner。
