# Reuse Decision

## Change type

component / hook / test

## Existing code searched

- `code/frontend/src/components/charts/LineChart.vue`
- `code/frontend/src/components/charts/BarChart.vue`
- `code/frontend/src/components/charts/GaugeChart.vue`
- `code/frontend/src/components/charts/RadarChart.vue`
- `code/frontend/src/components/charts/echartsMountGate.ts`
- `code/frontend/src/components/charts/radarVisuals.ts`
- `code/frontend/src/composables/useTheme.ts`
- `code/frontend/src/views/dashboard/DashboardView.vue`
- `code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue`

## Similar implementations found

- `useTheme()` already provides the reactive `theme` ref used across the app shell.
- Existing chart wrappers already centralize `vue-echarts` mounting and option generation.
- `radarVisuals.ts` already centralizes the radar canvas color derivation.

## Decision

extend_existing

## Reason

The bug is in the existing shared chart wrappers, not in a missing new chart abstraction. The smallest safe fix is to extend the current wrappers so their `option` computed values explicitly depend on `theme.value`, which makes theme changes re-run the option builders and refresh canvas colors.

## Files to modify

- `code/frontend/src/components/charts/LineChart.vue`
- `code/frontend/src/components/charts/BarChart.vue`
- `code/frontend/src/components/charts/GaugeChart.vue`
- `code/frontend/src/components/charts/RadarChart.vue`
- `code/frontend/src/components/charts/__tests__/EChartsMountGate.test.ts`

## After implementation

- Keep the theme dependency in the shared chart wrappers so future chart pages reuse the same reactivity behavior.
- The regression test should remain on the shared chart test file instead of being duplicated in `student/dashboard`.
