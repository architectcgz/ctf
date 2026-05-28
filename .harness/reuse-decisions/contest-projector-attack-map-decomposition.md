# Reuse Decision

## Change type
frontend refactor / feature-owned contest projector attack map decomposition

## Existing code searched
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMap.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackDetailOverlay.vue
- code/frontend/src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts
- code/frontend/src/views/platform/ContestProjector.vue
- code/frontend/src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts
- docs/reviews/frontend/2026-05-28-contest-projector-ui-cluster-feature-ui-normalization-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `PlatformContestFormPanel.vue`、`ContestChallengeOrchestrationPanel.vue`、`AWDOperationsPanel.vue` 这几轮都已经按“父层保留唯一 owner，稳定展示分区下沉，必要的 DOM / workflow mechanics 收进局部子 owner”完成过收口；`ContestProjectorAttackMap.vue` 也应沿同一模式推进。
- `ContestProjectorAttackDetailOverlay.vue` 已经是 attack map drilldown 的独立 detail owner，说明这条线已经接受“父 map 只组合 panel / overlay，而不是继续把所有 drilldown UI 混在一个文件里”的模式。
- `useContestProjectorPage()` 已经把 route-level projector lifecycle、fullscreen 和 focus panel owner 收到 page model；本轮不应把这些 route owner 再重新拉回 attack map 内部。

## Decision
refactor_existing

## Reason
`ContestProjectorAttackMap.vue` 当前约 `779` 行，问题不只是模板大，而是两类 owner 混在一起：

- map view-model owner：`visibleEdges`、`teamPanels`、`rankingRows`、detail panel open/close
- board DOM mechanics owner：drag offsets、`ResizeObserver`、beam path 计算、localStorage 持久化、team/service DOM ref

最小正确改动不是只切一点 template，也不是把全部逻辑塞进新的“大子组件”，而是：

- 保持 `ContestProjectorAttackMap.vue` 继续做唯一 map view-model owner 和 detail overlay owner
- 新增 `ContestProjectorAttackMapTeamSidebar.vue` 承接左侧 legend / first blood / team list drilldown
- 新增 `ContestProjectorAttackBoard.vue` 承接中央 board shell、beam layer、team node、recent events
- 新增 `ContestProjectorAttackMapStatsSidebar.vue` 承接右侧 ranking / attacks drilldown
- 新增 `useProjectorAttackBoard.ts` 承接 board DOM mechanics owner
- 如有需要，新增 projector attack map support 文件承接 `AttackMapDetailPanel`、service display/icon 等共享 helper，避免 overlay 与 board 再各写一份

本轮不调整 `ContestProjector.vue` 的 route shell contract，不改 `useContestProjectorPage()` 的 page owner，也不改 `ContestProjectorAttackDetailOverlay.vue` 的对外行为。

## Files to modify
- .harness/reuse-decisions/contest-projector-attack-map-decomposition.md
- docs/plan/impl-plan/2026-05-28-contest-projector-attack-map-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-contest-projector-attack-map-decomposition-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMap.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackBoard.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMapTeamSidebar.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMapStatsSidebar.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackDetailOverlay.vue
- code/frontend/src/features/contest-projector/ui/index.ts
- code/frontend/src/features/contest-projector/model/projectorTypes.ts
- code/frontend/src/features/contest-projector/model/useProjectorAttackBoard.ts
- code/frontend/src/features/contest-projector/model/index.ts
- code/frontend/src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts
- code/frontend/src/features/contest-projector/ui/__tests__/contestProjectorAttackMapExtraction.test.ts

## After implementation
- `ContestProjectorAttackMap.vue` 会回到 “map view-model owner + detail overlay owner” 这一层，不再继续内联左右侧栏和中央 board 壳体。
- board drag / beam / observer / storage 这类 DOM-heavy mechanics 会落到独立 composable，后续如果 projector board 继续增长，可以在正确 owner 下继续处理。
- backlog 里 `ContestProjectorAttackMap.vue` 这一项会从“超大组件待拆”转成已显著收口，P2 剩余重点继续收敛。
