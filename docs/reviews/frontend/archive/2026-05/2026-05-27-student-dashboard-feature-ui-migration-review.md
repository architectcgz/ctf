# Student Dashboard Feature UI Migration 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-student-dashboard-feature-ui-migration-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/student-dashboard-feature-ui-migration.md`
    - `docs/plan/impl-plan/2026-05-27-student-dashboard-feature-ui-migration-plan.md`
    - `docs/reviews/frontend/2026-05-27-student-dashboard-feature-ui-migration-review.md`
    - `code/frontend/src/features/student-dashboard/**/*`
    - `code/frontend/src/views/dashboard/DashboardView.vue`
    - `code/frontend/src/components/dashboard/student/StudentOverviewVariantSwitcher.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
    - `code/frontend/src/views/__tests__/studentOverviewEntrypoint.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 student dashboard feature-owned UI 迁移，保留 `useStudentDashboardPage.ts` 为数据 / route owner，只迁 student-only panel 和 panel registry。
- Gate verdict：Pass after targeted verification

## Findings

- None.

## Material findings

- None.

## Senior implementation assessment

- `StudentCategoryProgressPage`、`StudentDifficultyPage`、`StudentOverviewPage`、`StudentRecommendationPage` 和 `studentDashboardPanelRegistry.ts` 已迁入 `features/student-dashboard/ui`，`DashboardView.vue` 现在只通过 `features/student-dashboard` public API 组合 page model 与 panel registry。
- `useStudentDashboardPage.ts` 继续保留 route / query tabs / dashboard data owner，没有因为迁移 UI 文件重新反向依赖 `components/dashboard/student/*Page.vue`。
- `StudentOverviewVariantSwitcher.vue` 已改成直接桥接 `StudentOverviewStyleEditorial.vue`，避免继续依赖被迁出的 `StudentOverviewPage.vue` 旧路径。
- `StudentTimelinePage.vue` 本轮保持在 `components/dashboard/student`，因为 teacher 学员洞察还直接复用它；这让本轮迁移保持在 student-only feature UI 范围内，没有把共享 panel owner 问题和 page migration 混在一起。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/dashboard/__tests__/DashboardView.test.ts src/views/__tests__/studentOverviewEntrypoint.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## Residual risk

- `StudentTimelinePage.vue` 仍是 `legacy component page`，并且被 teacher 学员洞察共享消费；后续如果继续压缩 `legacyComponentPageAllowlist`，需要先把这块抽成更中立的 timeline panel，再决定 feature / teacher 的最终 owner。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中这条 `P1` 在 touched surface 上已经进一步收口，只剩 `StudentTimelinePage.vue` 这一条 student dashboard 残余 page-sized 组件未迁出 `components/**`。
