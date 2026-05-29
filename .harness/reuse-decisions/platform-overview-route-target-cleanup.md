# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts
- code/frontend/src/features/platform-overview/model/useCheatDetectionPage.ts
- code/frontend/src/features/platform-overview/model/index.ts
- code/frontend/src/features/platform-overview/ui/PlatformOverviewPage.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue
- code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue
- code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue
- code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue
- code/frontend/src/views/platform/PlatformOverview.vue
- code/frontend/src/views/platform/CheatDetection.vue
- code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts
- code/frontend/src/views/platform/__tests__/CheatDetection.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useChallengePackageFormatPage.ts` 已从单次 `router.push()` wrapper 收口成纯 route target contract。
- `useNotificationListPage.ts` 已去掉 `vue-router`，由 route view 直接消费 route target helper。

## Decision
refactor_existing

## Reason
`platform-overview` feature 里当前有两条同类 router 例外：

- `usePlatformOverviewPage.ts`
- `useCheatDetectionPage.ts`

它们都不读取 route params / query，也不承担 query-tab、alias redirect 或 role redirect，只是顺手跳到：

- `AuditLog`
- `CheatDetection`

最小正确改动是：

- 新增本地 route target contract helper，统一生成审计日志与作弊检测跳转目标
- `usePlatformOverviewPage.ts` / `useCheatDetectionPage.ts` 改成只返回 route target
- `PlatformOverviewPage.vue` / `CheatDetectionWorkspacePanel.vue` 及其下游子组件直接通过 `RouterLink` 消费 route target

这样可以在同一个 feature 内一次收掉 2 条 `featureRouterImportAllowlist`，避免留下半收口状态。

本轮不做：

- 不处理 `useAuditLogPage.ts`
- 不改 overview / cheat detection 的数据加载 owner
- 不继续拆 dashboard / cheat 组件层级

## Files to modify
- .harness/reuse-decisions/platform-overview-route-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-platform-overview-route-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-platform-overview-route-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-overview/model/index.ts
- code/frontend/src/features/platform-overview/model/platformOverviewRoutes.ts
- code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts
- code/frontend/src/features/platform-overview/model/useCheatDetectionPage.ts
- code/frontend/src/features/platform-overview/ui/PlatformOverviewPage.vue
- code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue
- code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue
- code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue
- code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue
- code/frontend/src/views/platform/PlatformOverview.vue
- code/frontend/src/views/platform/CheatDetection.vue
- code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts
- code/frontend/src/views/platform/__tests__/CheatDetection.test.ts

## After implementation
- `usePlatformOverviewPage.ts` 不再 import `vue-router`
- `useCheatDetectionPage.ts` 不再 import `vue-router`
- `platform-overview` feature 内导航改为 route target contract + `RouterLink`
- `featureRouterImportAllowlist` 再减少 2 条
