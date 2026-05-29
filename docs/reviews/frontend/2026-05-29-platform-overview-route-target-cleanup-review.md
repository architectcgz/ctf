# Platform Overview Route Target Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-platform-overview-route-target-cleanup-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/platform-overview-route-target-cleanup.md`
  - `docs/plan/impl-plan/2026-05-29-platform-overview-route-target-cleanup-plan.md`
  - `docs/reviews/frontend/2026-05-29-platform-overview-route-target-cleanup-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/features/platform-overview/model/index.ts`
  - `code/frontend/src/features/platform-overview/model/platformOverviewRoutes.ts`
  - `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts`
  - `code/frontend/src/features/platform-overview/model/useCheatDetectionPage.ts`
  - `code/frontend/src/features/platform-overview/ui/PlatformOverviewPage.vue`
  - `code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue`
  - `code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue`
  - `code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue`
  - `code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue`
  - `code/frontend/src/views/platform/PlatformOverview.vue`
  - `code/frontend/src/views/platform/CheatDetection.vue`
  - `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
  - `code/frontend/src/views/platform/__tests__/CheatDetection.test.ts`
- Classification check：同意按 `platform-overview` feature 内同类 route target cleanup 处理；`usePlatformOverviewPage.ts` 与 `useCheatDetectionPage.ts` 的 router 依赖都只是单次跳转，不应继续保留为 reviewed route-aware owner。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `platformOverviewRoutes.ts` 现在统一承接 overview / cheat detection 两条面上的导航 target，避免在同一 feature 内保留两份“数据 owner + 顺手 push”的重复模式。
- `usePlatformOverviewPage.ts` 与 `useCheatDetectionPage.ts` 已去掉 `vue-router`，只保留数据加载、错误状态、刷新和快捷动作数据 owner。
- `PlatformOverviewHeroPanel.vue` 与 `CheatDetection*` 组件群已直接通过 `RouterLink` 消费 route target；刷新与加载 owner 仍留在原 page model，没有被顺手下沉。
- `featureRouterImportAllowlist` 已从 `platform-overview` 这一组净减少 2 条。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/CheatDetection.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-overview-route-target-cleanup.md docs/plan/impl-plan/2026-05-29-platform-overview-route-target-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-overview-route-target-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-overview/model/index.ts code/frontend/src/features/platform-overview/model/platformOverviewRoutes.ts code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts code/frontend/src/features/platform-overview/model/useCheatDetectionPage.ts code/frontend/src/features/platform-overview/ui/PlatformOverviewPage.vue code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue code/frontend/src/views/platform/PlatformOverview.vue code/frontend/src/views/platform/CheatDetection.vue code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts code/frontend/src/views/platform/__tests__/CheatDetection.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `useAuditLogPage.ts` 仍保留 route-aware owner，这轮不一并处理。
