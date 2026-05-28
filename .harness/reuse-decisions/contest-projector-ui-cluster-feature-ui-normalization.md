# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/views/platform/ContestProjector.vue
- code/frontend/src/features/contest-projector/index.ts
- code/frontend/src/features/contest-projector/model/useContestProjectorPage.ts
- code/frontend/src/features/contest-projector/model/useContestProjectorBoundary.test.ts
- code/frontend/src/components/platform/contest/projector/ContestProjectorAttackMap.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorAttackDetailOverlay.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorFocusOverlay.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorHero.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorToolbar.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorTraffic.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorLeaderboard.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorServiceMatrix.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorEvents.vue
- code/frontend/src/components/platform/contest/projector/ContestProjectorAttackMap.css
- code/frontend/src/components/platform/contest/projector/ContestProjectorAttackMapResponsive.css
- code/frontend/src/components/platform/contest/projector/ContestProjectorAttackDetailOverlay.css
- code/frontend/src/components/platform/contest/projector/__tests__/ContestProjectorAttackMap.test.ts
- code/frontend/src/components/platform/contest/projector/__tests__/ContestProjectorFocusOverlay.test.ts
- code/frontend/src/components/platform/contest/projector/__tests__/ContestProjectorServiceMatrix.test.ts
- code/frontend/src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts
- code/frontend/src/components.d.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `features/awd-inspector/ui` 已整体承接 `AWDRoundInspector.vue`、`AWDTrafficPanel.vue`、`AWDServiceStatusPanel.vue` 等单一 capability UI cluster，说明这种“feature model 已在、旧 components 目录仍挂 UI”的情况，最小正确收口方式是整簇迁入 feature `ui`。
- `features/contest-awd-admin/ui` 最近几刀已连续承接 `AWDOperationsPanel.vue` 及其 runtime 子件，说明当前前端 migration 基线已经接受“route / page 继续保留组合 owner，单一 feature UI cluster 从旧 `components/**` 收回 feature owner”的切法。
- `features/contest-projector/model/*` 已经持有 projector types、formatters、page model 和派生数据构建，不再需要继续让 projector UI 依赖旧 `components/platform/contest/projector/contestProjectorTypes.ts` 与 `contestProjectorFormatters.ts`。

## Decision
refactor_existing

## Reason
`ContestProjector.vue` 当前已经通过 `useContestProjectorPage()` 走 `features/contest-projector` page model，但 route view 仍直接引用旧 `components/platform/contest/projector/*` UI cluster。这会把单一 feature 的 UI owner 留在历史共用组件目录里。最小正确改动是：

- 把 projector UI cluster 整体迁入 `features/contest-projector/ui`
- `ContestProjector.vue` 改为通过 `@/features/contest-projector` public API 组合 page model 与 UI
- UI 组件内部改为复用 `features/contest-projector/model/*` 的 types / formatters
- 同步更新 `components.d.ts`、raw-source / component test 路径和 backlog 记录

本轮不继续调整 `useContestProjectorPage()` 的 route / fullscreen / auto refresh owner，也不继续拆 `ContestProjectorAttackMap.vue` 的内部展示分区。

## Files to modify
- .harness/reuse-decisions/contest-projector-ui-cluster-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-contest-projector-ui-cluster-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-contest-projector-ui-cluster-feature-ui-normalization-review.md
- code/frontend/src/features/contest-projector/index.ts
- code/frontend/src/features/contest-projector/ui/index.ts
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackDetailOverlay.css
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackDetailOverlay.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMap.css
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMapResponsive.css
- code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMap.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorEvents.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorFocusOverlay.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorHero.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorLeaderboard.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorServiceMatrix.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorToolbar.vue
- code/frontend/src/features/contest-projector/ui/ContestProjectorTraffic.vue
- code/frontend/src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts
- code/frontend/src/features/contest-projector/ui/__tests__/ContestProjectorFocusOverlay.test.ts
- code/frontend/src/features/contest-projector/ui/__tests__/ContestProjectorServiceMatrix.test.ts
- code/frontend/src/views/platform/ContestProjector.vue
- code/frontend/src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts
- code/frontend/src/components.d.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- projector route view 不再直接引用旧 `components/platform/contest/projector/*` 路径。
- `contest-projector` capability 会同时持有 model 和 UI owner。
- `components/platform/contest/projector/*` 目录里的单一 feature UI cluster 残片会显著缩小，旧 types / formatter 回流点也会收掉。
