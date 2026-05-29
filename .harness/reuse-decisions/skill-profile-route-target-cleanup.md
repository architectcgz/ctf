# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts
- code/frontend/src/views/profile/SkillProfile.vue
- code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue
- code/frontend/src/views/profile/__tests__/SkillProfile.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components/navigation/AppRouteLink.vue
- code/frontend/src/features/teacher-dashboard/model/teacherDashboardRoutes.ts

## Similar implementations found
- `teacher-dashboard-route-target-cleanup` 已示范“page model 暴露 route target，page shell / workspace shell 通过 `AppRouteLink` 消费”的模式。
- `student-management-route-target-cleanup` 已示范“列表行 / header action 的声明式导航改成 route target，不动数据和弹窗 owner”的做法。

## Decision
refactor_existing

## Reason
`useSkillProfilePage.ts` 当前没有 `useRoute`、没有 query sync，也没有 redirect guard；router 依赖只剩两类声明式导航：

- `goToChallenges()`
- `goToChallenge(id)`

真正需要留在 page model 里的 owner 是：

- 学员选择
- 六维画像数据加载
- 推荐靶场数据加载
- 错误态和刷新

最小正确改动是：

- 给 `skill-profile` 补本地 route target helper
- `useSkillProfilePage.ts` 去掉 `vue-router`，改为暴露 `challengesRoute` 与 `buildChallengeRoute()`
- `SkillProfile.vue` 继续只组合 feature page model 与 workspace shell
- `SkillProfileWorkspaceShell.vue` 通过共享 `AppRouteLink` 消费 route target

这样可以收掉：

- `features/skill-profile/model/useSkillProfilePage.ts -> vue-router`

本轮不做：

- 不处理 `SkillProfile.vue` 本身的 tab query owner
- 不迁移 `components/profile/SkillProfileWorkspaceShell.vue` 到 `features/*/ui`
- 不改六维画像数据加载、推荐算法或教师视角学员选择

## Files to modify
- .harness/reuse-decisions/skill-profile-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-skill-profile-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-skill-profile-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/skill-profile/model/index.ts
- code/frontend/src/features/skill-profile/model/skillProfileRoutes.ts
- code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts
- code/frontend/src/views/profile/SkillProfile.vue
- code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue
- code/frontend/src/views/profile/__tests__/SkillProfile.test.ts

## After implementation
- `useSkillProfilePage.ts` 不再 import `vue-router`
- 六维画像页“去做题”和推荐题目跳转改成 route target + `AppRouteLink`
- `featureRouterImportAllowlist` 再减少 1 条
