# Contest Manage Panel Owner Tightening 计划

## Objective

- 把 `ContestManage` 的 `overview/create` panel query owner 从 `ContestOrchestrationPage.vue` 收回 `useContestManagePage.ts`。
- 保持 route page 薄壳和现有 contest workflow owner 不变。

## Non-goals

- 不改赛事目录请求、状态筛选、创建 / 编辑保存和公告抽屉。
- 不改 `usePlatformContests()` 内部状态组合。
- 不扩到 `ContestEdit`、`ContestOperationsHub` 或其他 contest 页面。

## Source Inputs

- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform/user-management/model/useUserGovernancePanelRoute.ts`
- `code/frontend/src/shared/model/navigation/useRouteQueryTransport.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`

## Plan Review Result

- 推荐直接复用 `UserManage` 的模式：page model 读写 query，单独 helper 管 panel 解析，UI 壳只收 props / emits。
- `useUrlSyncedTabs()` 仍可留在其他确实需要键盘 tab 状态 owner 的位置，本轮不把它当成 contest manage 的正确 owner。

## Task Slices

### Slice 1: 提取 contest manage panel query helper

- 目标：新增纯 helper，统一解析 `overview/create/list` 并构建 query。
- 风险：
  - 如果 query normalize 不兼容旧 `panel=list`，创建成功后的回退会出错。

### Slice 2: 收回 page model owner

- 目标：让 `useContestManagePage.ts` 成为唯一 panel owner，并通过 route transport 切换 query。
- 风险：
  - 如果把 UI 壳和 page model 同时保留 panel 状态，会形成双 owner。

### Slice 3: 让 ContestOrchestrationPage 退回纯展示壳

- 目标：删除 UI 壳里的 `useUrlSyncedTabs()` 和 watch，同步 route page / 测试。
- 风险：
  - 如果事件 contract 不清楚，会把创建按钮、返回工作台按钮的切换行为打散。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision contest-manage-panel-owner-tightening`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `git diff --check -- .harness/reuse-decisions/contest-manage-panel-owner-tightening.md docs/plan/impl-plan/2026-05-31-contest-manage-panel-owner-tightening-plan.md code/frontend/src/features/platform/contests/model/useContestManagePage.ts code/frontend/src/features/platform/contests/model/useContestManagePanelRoute.ts code/frontend/src/features/platform/contests/model/index.ts code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Review Focus

- `ContestManage` 是否只保留一个 panel query owner。
- `ContestOrchestrationPage.vue` 是否已经退回纯 props / emits 壳，不再直接碰 route query。
- 创建竞赛成功后的回退和“打开创建面板 / 返回工作台”行为是否保持不变。

## Rollback / Recovery

- 如果 helper 命名或事件 contract 可读性差，可以调整命名，但不能回退到 UI 壳直接持有 query owner。
