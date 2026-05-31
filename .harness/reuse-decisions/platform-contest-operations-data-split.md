# Reuse Decision

## Change type
frontend refactor / platform contest operations data split

## Existing code searched
- `code/frontend/src/features/platform/contests/model/useContestOperationsPage.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`
- `code/frontend/src/features/platform/contests/model/useContestAnnouncementsData.ts`

## Similar implementations found
- `useContestOperationsPage.ts` 当前同时持有单场竞赛请求、loading 状态、breadcrumb side effect 和 toast side effect。
- `useContestAnnouncementsData.ts` 已经证明 `platform/contests` 下的页面级数据 owner 可以先拆到独立 data composable。

## Decision
refactor_existing

## Reason
当前最小正确切片是把单场运维页拆成：

- `useContestOperationsData`：承接单场竞赛请求、loading 和错误状态。
- `useContestOperationsPage`：保留 breadcrumb、toast 和 runtime 内容派生。

这样可以：

- 去掉 `useContestOperationsPage` 里的 API 请求 owner
- 保持 breadcrumb / toast 这类页面壳 side effect 仍在 page model

本轮不做：

- 不改运维页 UI
- 不改 breadcrumb hook 或 toast hook
- 不调整 `AWDOperationsPanel` contract

## Files to modify
- `.harness/reuse-decisions/platform-contest-operations-data-split.md`
- `docs/plan/impl-plan/2026-05-31-platform-contest-operations-data-split-plan.md`
- `code/frontend/src/features/platform/contests/model/useContestOperationsData.ts`
- `code/frontend/src/features/platform/contests/model/useContestOperationsData.test.ts`
- `code/frontend/src/features/platform/contests/model/useContestOperationsPage.ts`
- `code/frontend/src/features/platform/contests/model/index.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestOperations.test.ts`

## After implementation
- 单场运维页的数据加载 owner 会集中到 `useContestOperationsData`。
- `useContestOperationsPage` 只保留页面壳 side effect 和 runtime 内容派生。
