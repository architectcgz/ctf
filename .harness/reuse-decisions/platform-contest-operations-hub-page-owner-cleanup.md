# Reuse Decision

## Change type
frontend refactor / platform contest operations hub route page owner cleanup

## Existing code searched
- `code/frontend/src/pages/platform/contests/ContestOperationsHubRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestOperationsHubPage.ts`
- `code/frontend/src/features/platform/contests/ui/ContestOperationsHubHeroPanel.vue`
- `code/frontend/src/features/platform/contests/ui/ContestOperationsHubWorkspacePanel.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/instance-management/ui/PlatformInstanceManagementPage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperationsHub.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `PlatformContestManagePage.vue` 刚刚已经承接了 contest manage 的 page model 和 route-level overlay/section 组合，`ContestManageRoutePage.vue` 退回薄壳。
- `PlatformInstanceManagementPage.vue` 也已经证明，平台后台的目录路由页可以只渲染 feature page，而把 page model 与 panel 组合留在 feature 内部。
- `ContestOperationsHubRoutePage.vue` 当前只组合 `useContestOperationsHubPage()`、hero panel 和 workspace panel，没有额外 overlay，是这组 route page 中更小、更直接的 owner 收口候选。

## Decision
refactor_existing

## Reason
当前最小正确改动不是继续拆 `useContestOperationsHubPage.ts` 或 panel 本身，而是新增 `PlatformContestOperationsHubPage.vue`，把：

- `useContestOperationsHubPage()`
- `ContestOperationsHubHeroPanel`
- `ContestOperationsHubWorkspacePanel`

统一收回 `features/platform/contests/ui` 内部，再让 `ContestOperationsHubRoutePage.vue` 退回真正的 feature public API 薄壳。这样能继续压 `platform/contests` 的 route-level owner 面，同时保持请求、分页、preferred contest 和 route target contract 全部不变。

## Files to modify
- `.harness/reuse-decisions/platform-contest-operations-hub-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-platform-contest-operations-hub-page-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-06-01-platform-contest-operations-hub-page-owner-cleanup-review.md`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsHubPage.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/pages/platform/contests/ContestOperationsHubRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperationsHub.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 赛事运维目录 route page 会退回薄壳。
- `platform/contests` 内会新增明确的 operations hub page-level owner，后续继续处理 `ContestAnnouncements` / `ContestEdit` 时可复用同一模式。
- `useContestOperationsHubPage.ts` 和现有 hero/workspace panel 的输入输出契约保持不变。
