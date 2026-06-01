# Reuse Decision

## Change type
frontend refactor / platform contest manage route page owner cleanup

## Existing code searched
- `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/instance-management/ui/PlatformInstanceManagementPage.vue`
- `code/frontend/src/pages/platform/InstanceManageRoutePage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspacePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `PlatformInstanceManagementPage.vue` 已把平台实例目录的 page model 与 workspace shell 收回 feature 内部，`InstanceManageRoutePage.vue` 只剩单个 feature page 薄壳。
- `StudentAnalysisWorkspacePage.vue` 已证明，共享或跨区块的 page-level overlay 也可以和 page model 一起留在 feature workspace page 中，由 route page 退回纯渲染壳。
- `ContestManageRoutePage.vue` 现在直接组合 `ContestOrchestrationPage`、`ContestAnnouncementManageDrawer`、`PlatformContestFormDialog`、`AWDReadinessOverrideDialog` 与 `useContestManagePage()`；这组职责已经明显超过“route composition surface”的最小边界。

## Decision
refactor_existing

## Reason
上一刀已经把 `ContestOrchestrationPage.vue` 收口成纯 shell。当前更合适的下一步不是继续拆 route page 里的单个 dialog，而是直接把 `ContestManageRoutePage.vue` 退回薄壳，新增 feature 内部的 `PlatformContestManagePage.vue` 统一承接：

- `useContestManagePage()` page model
- `ContestOrchestrationPage` contest manage shell
- `ContestAnnouncementManageDrawer`
- `PlatformContestFormDialog`
- `AWDReadinessOverrideDialog`

这样 route page 就能对齐最近已经采用的 `PlatformInstanceManagementPage.vue` / `StudentAnalysisWorkspacePage.vue` 模式，同时不改 contest manage 的 async workflow owner、query owner 和 overlay contract。

## Files to modify
- `.harness/reuse-decisions/platform-contest-manage-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-platform-contest-manage-page-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-06-01-platform-contest-manage-page-owner-cleanup-review.md`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 平台竞赛目录 route page 会退回只渲染 feature page 的薄壳。
- `platform/contests` 会新增一个明确的 page-level owner，后续继续清这条线时，不需要再让 route page 直接耦合 page model 和 overlay workflow。
- `ContestOrchestrationPage.vue`、`useContestManagePage.ts` 和现有 dialog / drawer 的外部契约保持不变。
