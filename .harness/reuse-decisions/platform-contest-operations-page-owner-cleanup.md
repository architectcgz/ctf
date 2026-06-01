# Reuse Decision

## Change type
frontend refactor / platform contest operations route page owner cleanup

## Existing code searched
- `code/frontend/src/pages/platform/contests/ContestOperationsRoutePage.vue`
- `code/frontend/src/features/platform/contests/model/useContestOperationsPage.ts`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestManagePage.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsHubPage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `PlatformContestManagePage.vue` 已承接 contest manage 的 page model、page shell 与 overlay owner，route page 退回薄壳。
- `PlatformContestOperationsHubPage.vue` 已承接 operations hub 的 page model 与 workspace panel 组合，route page 只保留 feature public API 薄壳。
- `ContestOperationsRoutePage.vue` 当前也只接收一个 `contestId` route prop，但仍直接组合 `useContestOperationsPage()`、`AWDOperationsPanel` 与服务告警插槽，因此适合沿同一模式收口。

## Decision
refactor_existing

## Reason
这轮最小正确改动不是去碰 `useContestOperationsPage()` 的单场赛事加载或 breadcrumb owner，而是新增 `PlatformContestOperationsPage.vue`，把：

- `useContestOperationsPage(toRef(props, 'contestId'))`
- `AWDOperationsPanel`
- `AWDServiceAlertBanner`
- 本地 loading shell 与 inspector workspace 布局

统一收回 `features/platform/contests/ui` 内部。`ContestOperationsRoutePage.vue` 继续只保留 route prop 输入并渲染 feature page，从而让 route-level owner 回到最小，同时保持 `contestId` prop contract 不变。

## Files to modify
- `.harness/reuse-decisions/platform-contest-operations-page-owner-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-platform-contest-operations-page-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-06-01-platform-contest-operations-page-owner-cleanup-review.md`
- `code/frontend/src/features/platform/contests/ui/PlatformContestOperationsPage.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/pages/platform/contests/ContestOperationsRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 单场竞赛运维 route page 会退回 “route prop -> feature page” 的薄壳。
- `platform/contests` 内会新增明确的 operations page-level owner，后续如果继续清 `ContestEditRoutePage.vue`，可以继续复用同一模式。
- `useContestOperationsPage.ts` 的 breadcrumb、toast 与单场赛事查询 owner 保持不变。
