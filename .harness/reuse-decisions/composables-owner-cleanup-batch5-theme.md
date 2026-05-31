# Reuse Decision

## Change type
refactor_existing / shared-theme-owner / docs / test

## Existing code searched
- `code/frontend/src/composables/useTheme.ts`
- `code/frontend/src/composables/__tests__/useTheme.test.ts`
- `code/frontend/src/App.vue`
- `code/frontend/src/shared/model/layout/topnav/useTopNavViewState.ts`
- `code/frontend/src/shared/ui/charts/*.vue`
- `code/frontend/src/shared/ui/charts/__tests__/EChartsMountGate.test.ts`
- `code/frontend/src/shared/ui/layout/topnav/TopNavBrandPicker.vue`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`

## Similar implementations found
- `shared/model/reporting/useReportStatusPolling.ts` 已承接共享 workflow owner
- `shared/model/layout/*` 已承接共享布局状态 owner
- `useTheme` 当前由 `App.vue` 初始化，并被 topnav、charts 等多个 shared surface 共同读取，明显是全局主题 owner，而不是局部 composable

## Decision
refactor_existing

## Reason
- `useTheme` 管理全局主题、品牌、DOM `data-*` 属性与 localStorage 持久化，owner 很清楚
- 它不属于基础无状态机制层，也不属于某个 feature，最合理的落点是 `shared/model/theme`
- 这批可以把主题入口从历史 `src/composables` 收口出来，同时不改任何主题行为

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch5-theme.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch5-theme-plan.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `code/frontend/src/composables/useTheme.ts`
- `code/frontend/src/composables/__tests__/useTheme.test.ts`
- `code/frontend/src/shared/model/theme/useTheme.ts`
- `code/frontend/src/shared/model/theme/__tests__/useTheme.test.ts`
- `code/frontend/src/App.vue`
- `code/frontend/src/shared/model/layout/topnav/useTopNavViewState.ts`
- `code/frontend/src/shared/ui/charts/LineChart.vue`
- `code/frontend/src/shared/ui/charts/BarChart.vue`
- `code/frontend/src/shared/ui/charts/GaugeChart.vue`
- `code/frontend/src/shared/ui/charts/RadarChart.vue`
- `code/frontend/src/shared/ui/charts/__tests__/EChartsMountGate.test.ts`
- `code/frontend/src/shared/ui/layout/topnav/TopNavBrandPicker.vue`

## After implementation
- 全局主题 owner 从历史 `src/composables` 收口到 `shared/model/theme`
- `src/composables` 继续只剩 router / realtime / easter-eggs 等还待判断 owner 的内容
