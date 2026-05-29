# Reuse Decision

## Change type
frontend refactor / feature-owned AWD operations panel owner cleanup

## Existing code searched
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPreRuntimeStage.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsRuntimeStage.vue
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsTabs.vue
- code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.ts
- code/frontend/src/components/layout/notification-drawer/useNotificationDrawerViewState.ts
- code/frontend/src/components/layout/sidebar/useSidebarNavigationViewState.ts
- code/frontend/src/components/layout/topnav/useTopNavViewState.ts
- code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts
- docs/plan/impl-plan/2026-05-28-awd-operations-panel-decomposition-plan.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- layout 线最近已经把 `NotificationDrawer.vue`、`Sidebar.vue`、`TopNav.vue` 的本地 view-state / interaction owner 收口到 `use*ViewState.ts`，父组件只保留 bridge 和稳定组合壳。
- `AWDOperationsPanel.vue` 上一轮已经完成 stage shell / dialog hub 拆分；当前剩余债和 layout 线一致，主要不是模板还没拆，而是父层继续混放 panel view-state、dialog open-close 和 action hint owner。

## Decision
refactor_existing

## Reason
`AWDOperationsPanel.vue` 当前约 `503` 行。模板层的大块 stage / dialog 已经拆出，但父组件仍同时承担：

- panel view-state：`selectedContest`、`runtimeStageReady`、`activePanel`、`visibleOperationTabs`、`runtimeContent` 派生和 tab keyboard wiring
- dialog workflow：三个 dialog 的 open/close、保存后关闭、override dialog close 桥接
- action capability：`nextRoundNumber`、service check / attack log 的可用性和 hint 推导
- 剩余父层样式：`studio-ops-shell`、`studio-ops-content`

最小正确改动不是继续新建一层展示壳，也不是把这些逻辑散回 stage 子组件，而是：

- 保持 `AWDOperationsPanel.vue` 继续作为唯一 feature workflow owner，负责接 `usePlatformContestAwd()`、对外 props / emits contract，以及 stage / dialog 组合
- 新增 `useAwdOperationsPanelViewState.ts`，收口 contest 选择、runtime stage、panel tab 和 keyboard navigation 相关本地 owner
- 新增 `useAwdOperationsDialogState.ts`，收口 dialog open-close、保存后关闭、override dialog close，以及 round / service / attack action capability hint
- 新增 `awdOperationsPanel.css`，承接父层剩余 shell 样式
- 护栏改成覆盖“父层已改为组合 + composable owner”，而不是只检查 stage 子组件是否存在

## Files to modify
- .harness/reuse-decisions/awd-operations-panel-owner-cleanup.md
- docs/plan/impl-plan/2026-05-28-awd-operations-panel-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-28-awd-operations-panel-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue
- code/frontend/src/features/contest-awd-admin/ui/awdOperationsPanel.css
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.test.ts
- code/frontend/src/features/contest-awd-admin/ui/useAwdOperationsDialogState.test.ts
- code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts

## After implementation
- `AWDOperationsPanel.vue` 会退回到“feature workflow wiring + stage/dialog composition + outward emit bridge”这一层。
- stage / dialog 子组件不会被再次塞入业务 owner；本地 view-state 和 dialog capability owner 会明确落在 panel 专属 composable。
- backlog 中 `AWDOperationsPanel.vue` 这条 residual debt 会从“更深层 workflow handler / dialog cluster”推进到已收口状态，后续 residual 会进一步回到 layout shell 或 `AWDInstanceOrchestrationPanel.vue` 这类尚未单独开刀的 surface。
