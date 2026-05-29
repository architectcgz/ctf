# Feature Router Second Batch Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-feature-router-second-batch-cleanup-plan.md`
- Scope：
  - `useNotificationDetailPage.ts`
  - `useScoreboardRoutePage.ts`
  - `usePlatformChallengeDetailRoutePage.ts`
  - `useRouteQueryTabs.ts`
- Classification check：同意按“第二批中复杂度 route owner cleanup”处理；通知详情是 route-aware page owner，`scoreboard / platform challenge detail` 则是同构 query-tab owner，不需要扩大到 `ChallengeList` 那种 query/filter 迁移。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `notification detail` 不应把站内跳转继续留成 `router.push()`；route props 与 route target 已经是当前仓库更稳定的 owner 形态。
- `scoreboard route page` 与 `platform challenge detail route page` 如果只是把 `useRoute/useRouter` 从 feature model 平移到另一个 wrapper，并没有真正改善 contract。更稳的实现是直接把 `useRouteQueryTabs()` 收口成共享 query-tab route owner。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/notifications/__tests__/NotificationDetail.test.ts src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/__tests__/routeQueryTabsAdoption.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/feature-router-second-batch-cleanup.md docs/plan/impl-plan/2026-05-29-feature-router-second-batch-cleanup-plan.md docs/reviews/frontend/2026-05-29-feature-router-second-batch-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/composables/useRouteQueryTabs.ts code/frontend/src/router/routes/studentRoutes.ts code/frontend/src/features/notifications/model/useNotificationDetailPage.ts code/frontend/src/features/scoreboard/model/useScoreboardRoutePage.ts code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts code/frontend/src/views/notifications/NotificationDetail.vue code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts code/frontend/src/views/__tests__/routeQueryTabsAdoption.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `ChallengeList` 仍未处理。
- 这份 review 是同上下文 self-review；由于本轮工具策略不允许未经用户明确授权的 sub-agent delegation，独立 reviewer gate 仍待满足。
