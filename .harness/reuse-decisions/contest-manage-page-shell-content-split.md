# Reuse Decision

## Change type
frontend refactor / platform contest manage page shell content split

## Existing code searched
- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform/contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/features/platform/contests/ui/AWDChallengeConfigPanel.vue`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `StudentAnalysisPage.vue` 最近已经通过 `StudentAnalysisWorkspaceTabs.vue` 与 `StudentAnalysisWorkspaceContent.vue` 把 workspace shell 内的 tab owner / content assembly 拆开，父页只保留 shell 与事件桥接。
- `ContestEditWorkspacePanel.vue` 与 `AWDChallengeConfigPanel.vue` 已采用“panel 组件 + 独立 css 文件”的方式，让大块 workspace UI 不再把模板、脚本和样式长期塞在一个 SFC 里。
- `ContestOperationsHubHeroPanel.vue` / `ContestOperationsHubWorkspacePanel.vue` 这组也说明，在 `platform/contests` 内部继续按区块拆分 page-sized UI，而不是把所有区块硬放在单页壳体里，是当前已采用的模式。

## Decision
refactor_existing

## Reason
`ContestOrchestrationPage.vue` 当前约 449 行，虽然 panel query owner 已收回 `useContestManagePage.ts`，但 overview 与 create 两个大区块仍混在同一个 feature page shell 中。下一刀最小正确切口不是继续改 route page 或 page model，而是：

- 保留 `ContestOrchestrationPage.vue` 作为 `platform/contests` 的总 shell 与事件桥接 owner。
- 新增 `ContestManageOverviewPanel.vue`，承接目录 header、summary、filter、empty/loading/table 装配。
- 新增 `ContestManageCreatePanel.vue`，承接创建竞赛工作区与表单桥接。
- 将这一页的局部样式提到 `contestOrchestrationPage.css`，避免新增子组件后再把大量 page-level class 继续留在父 SFC。

这轮不改变 `ContestManageRoutePage.vue` 对 `useContestManagePage()` 的组合方式，不新建 route-level workspace page，也不触碰 create/edit dialog、公告抽屉或 AWD override workflow owner。

## Files to modify
- `.harness/reuse-decisions/contest-manage-page-shell-content-split.md`
- `docs/plan/impl-plan/2026-06-01-contest-manage-page-shell-content-split-plan.md`
- `docs/reviews/frontend/2026-06-01-contest-manage-page-shell-content-split-review.md`
- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform/contests/ui/ContestManageOverviewPanel.vue`
- `code/frontend/src/features/platform/contests/ui/ContestManageCreatePanel.vue`
- `code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.css`
- `code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.types.ts`
- `code/frontend/src/features/platform/contests/ui/index.ts`
- `code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `ContestOrchestrationPage.vue` 会退回纯 shell / active panel 切换 / 事件桥接 owner。
- `platform/contests` 的 contest manage surface 会新增两个内部 panel owner，后续继续收口该页时不需要再在单一 SFC 里混改 overview 与 create。
- route page、page model 和现有 async workflow owner 保持不变，回归面集中在 feature 内部 UI 装配与相邻 source-boundary 护栏。
