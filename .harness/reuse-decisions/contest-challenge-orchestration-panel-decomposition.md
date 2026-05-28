# Reuse Decision

## Change type
frontend refactor / feature-owned orchestration panel decomposition

## Existing code searched
- code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeEditorDialog.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeFilterStrip.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeSummaryStrip.vue
- code/frontend/src/features/contest-workbench/model/useContestChallengeOrchestration.ts
- code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts
- code/frontend/src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase24.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `PlatformContestFormPanel.vue` 刚完成的 section cluster 收口，适合作为“父层保留 workflow owner，稳定展示分区下沉”的最近模式参考。
- `ContestChallengeSummaryStrip.vue` 与 `ContestChallengeFilterStrip.vue` 已经是 `contest-workbench` 线上的稳定子区块，本轮继续沿同一套路把 header 和 directory section 下沉，比重写 model 或 route shell 更符合最小改动。
- `useContestChallengeOrchestration.ts` 已经拥有题目池加载、AWD 目录查询、保存、删除和对话框状态 owner，本轮不应该把这些 owner 再挪回 route 或新增 store。

## Decision
refactor_existing

## Reason
`ContestChallengeOrchestrationPanel.vue` 当前约 `560` 行，真正需要保留在父层的逻辑主要是：

- `useContestChallengeOrchestration()` 的 model wiring
- summary / filter strip 的组合顺序
- dialog props / emits 桥接

剩余大部分代码属于稳定的 header、目录表、空态 / 加载态和局部样式，适合按 feature 内 section 下沉。最小正确改动不是改 `useContestChallengeOrchestration()` contract，也不是把路由或数据 owner 上提，而是：

- 保持 `ContestChallengeOrchestrationPanel.vue` 继续做唯一 orchestration model owner 和 dialog owner
- 新增 `ContestChallengeOrchestrationHeader.vue` 承接标题、副标题和主操作按钮
- 新增 `ContestChallengeDirectorySection.vue` 承接加载态、空态、题目表与行操作
- 把随这些展示区块绑定的样式一起从父组件迁走

本轮不改变 `ContestChallengeEditorDialog.vue`、`useContestChallengeOrchestration.ts`、`ContestEditWorkspacePanel.vue` 的对外 contract，不新增 route/store owner。

## Files to modify
- .harness/reuse-decisions/contest-challenge-orchestration-panel-decomposition.md
- docs/plan/impl-plan/2026-05-28-contest-challenge-orchestration-panel-decomposition-plan.md
- docs/reviews/frontend/2026-05-28-contest-challenge-orchestration-panel-decomposition-review.md
- code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationHeader.vue
- code/frontend/src/features/contest-workbench/ui/ContestChallengeDirectorySection.vue
- code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts
- code/frontend/src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase24.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `ContestChallengeOrchestrationPanel.vue` 会回到 “model wiring + summary/filter strip + dialog owner” 这一层，不再继续内联 header、目录表、空态 / 加载态和整块样式。
- `contest-workbench/ui` 会形成更完整的 orchestration panel cluster，后续继续调整 AWD challenge config 或 contest edit workspace 时，可以直接复用这些稳定分区。
- 当前 backlog 里 `ContestChallengeOrchestrationPanel.vue` 这条 P2 会从“超大组件待拆”转成已收口或至少显著收口。
