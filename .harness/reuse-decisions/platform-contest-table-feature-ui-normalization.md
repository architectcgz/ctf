# Reuse Decision

## Change type
frontend architecture / feature-owned UI normalization

## Existing code searched
- code/frontend/src/components/platform/contest/PlatformContestTable.vue
- code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/features/platform-contests/index.ts
- code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/components.d.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `PlatformContestFormPanel.vue`、`PlatformContestFormDialog.vue` 已迁入 `features/platform-contests/ui`，说明 `platform-contests` 线上的 create / edit / list 配套 UI 已按 feature owner 收口。
- `ContestEditTopbarPanel.vue`、`AWDChallengeConfigPanel.vue` 已迁入同一目录，当前 `platform-contests/ui` 已经是 contest manage / edit surface 的稳定 UI owner。
- `PlatformContestTable.vue` 当前只被 `ContestOrchestrationPage.vue` 消费，没有 shared route 或跨 feature 调用面。

## Decision
refactor_existing

## Reason
`PlatformContestTable.vue` 是 `ContestOrchestrationPage.vue` 目录页的单一 feature 展示 UI，不是通用平台表格，也没有独立 shared owner。继续把它留在 `components/platform/contest/*`，会让 `platform-contests` 这条线保留一块历史遗留展示面，和最近几刀已经完成的 `platform-contests/ui` 收口模式不一致。最小正确改动是：

- 把 `PlatformContestTable.vue` 迁入 `features/platform-contests/ui`
- `ContestOrchestrationPage.vue` 改成 feature 内部相对 import
- `features/platform-contests/ui/index.ts` 补充 table export，保持 feature UI 聚合入口一致
- 更新 `components.d.ts`、raw-source 测试、组件测试与 backlog 记录

这样可以继续清掉 `platform-contests` 线上残余的 legacy component surface，并让 contest 目录页的 table owner 与 form / topbar / awd config 一致地落在同一 feature。

## Files to modify
- .harness/reuse-decisions/platform-contest-table-feature-ui-normalization.md
- docs/plan/impl-plan/2026-05-28-platform-contest-table-feature-ui-normalization-plan.md
- docs/reviews/frontend/2026-05-28-platform-contest-table-feature-ui-normalization-review.md
- code/frontend/src/components.d.ts
- code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue
- code/frontend/src/features/platform-contests/ui/ContestOrchestrationPage.vue
- code/frontend/src/features/platform-contests/ui/index.ts
- code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## After implementation
- `PlatformContestTable.vue` 会归 `features/platform-contests/ui` 持有。
- `ContestOrchestrationPage.vue` 与相关 raw-source / component 测试不再引用旧 `components/platform/contest/PlatformContestTable.vue` 路径。
- `platform-contests` 这条 feature 在 contest manage surface 上的低风险遗留 UI 会继续缩小。
