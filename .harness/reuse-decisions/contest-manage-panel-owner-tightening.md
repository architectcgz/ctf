# Reuse Decision

## Change type
frontend refactor / contest manage panel owner tightening

## Existing code searched
- `code/frontend/src/features`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform/user-management/model/useUserGovernancePanelRoute.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/shared/model/navigation/useUrlSyncedTabs.ts`
- `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `usePlatformUserManagePage.ts` 已经通过 `useRouteQueryTransport()` 成为 `panel` query 的唯一 owner。
- `useUserGovernancePanelRoute.ts` 已经证明“UI 壳层退回纯 props / emits，query helper 降为纯函数”这条模式在当前仓库里可复用。
- `ContestOrchestrationPage.vue` 目前仍直接使用 `useUrlSyncedTabs()` 维护 `overview/create` panel query，这让 route-aware panel owner 还停在 UI shell。

## Decision
refactor_existing

## Reason
当前最小正确切片不是重写 `ContestManage` 工作台，而是把 panel query owner 从 UI shell 收回 page model：

- 保留 `ContestOrchestrationPage.vue` 作为工作台壳和目录 / 创建面板的展示 owner。
- 在 `useContestManagePage.ts` 里统一读取 / 切换 `panel` query。
- 新增纯 helper 收 `overview/create/list` 的 panel 解析与 query 构建。
- 让 route page 继续只做 feature 组合，不自己持有 panel 逻辑。

本轮不做：

- 不改赛事列表加载、状态筛选、创建保存、编辑弹窗和公告抽屉 workflow。
- 不继续拆 `usePlatformContests()`。
- 不触碰 `ContestEdit`、`ContestOperationsHub`、`ContestProjector`。

## Files to modify
- `.harness/reuse-decisions/contest-manage-panel-owner-tightening.md`
- `docs/plan/impl-plan/2026-05-31-contest-manage-panel-owner-tightening-plan.md`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/contests/model/useContestManagePanelRoute.ts`
- `code/frontend/src/features/platform/contests/model/index.ts`
- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `useContestManagePage.ts` 会成为 `ContestManage` 的唯一 panel query owner。
- `ContestOrchestrationPage.vue` 不再直接持有 `useUrlSyncedTabs()`。
- `platform/contests` 的目录页 panel owner 会和 `UserManage` 一样收敛到 page model + 纯 query helper 模式。
