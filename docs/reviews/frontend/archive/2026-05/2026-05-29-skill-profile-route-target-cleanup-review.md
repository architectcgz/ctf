# Skill Profile Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-skill-profile-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/skill-profile-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-skill-profile-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-skill-profile-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/components/navigation/AppRouteLink.vue`
  - `code/frontend/src/features/skill-profile/model/index.ts`
  - `code/frontend/src/features/skill-profile/model/skillProfileRoutes.ts`
  - `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
  - `code/frontend/src/views/profile/SkillProfile.vue`
  - `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
  - `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- Classification check：同意按 `skill-profile` feature 内“薄导航 route target cleanup”处理；`useSkillProfilePage.ts` 的 router 依赖只剩“去做题 / 去题目详情”，不应继续保留为 reviewed route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `skillProfileRoutes.ts` 现在集中提供六维画像页的题库页和题目详情 route target contract。
- `useSkillProfilePage.ts` 已退回纯六维画像数据、推荐数据、教师学员选择、刷新和错误态 owner，不再直接 import `vue-router`；“去做题 / 去题目详情”改成 `challengesRoute` 与 `buildChallengeRoute()` contract。
- `SkillProfileWorkspaceShell.vue` 已通过共享 `AppRouteLink.vue` 渲染“去做题”和推荐题目卡片导航；tab、刷新和教师学员选择交互仍保持原 owner，没有顺手迁走。
- `SkillProfile.test.ts` 已从 mock `router.push()` 切到真实 router 导航断言，同时补上“page model 不再 import vue-router、workspace shell 直接消费 AppRouteLink”的 raw-source 护栏。
- `featureRouterImportAllowlist` 已再减少 1 条：`features/skill-profile/model/useSkillProfilePage.ts -> vue-router`。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/profile/__tests__/SkillProfile.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/skill-profile-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-skill-profile-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-skill-profile-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/skill-profile/model/index.ts code/frontend/src/features/skill-profile/model/skillProfileRoutes.ts code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts code/frontend/src/views/profile/SkillProfile.vue code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `SkillProfile.vue` 仍保留 `useUrlSyncedTabs()` 的 window query owner，这轮不处理。
