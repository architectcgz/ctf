# Reuse Decision

## Change type
+component / feature / refactor / docs

## Existing code searched
- `code/frontend/src/pages/platform/CheatDetectionRoutePage.vue`
- `code/frontend/src/features/platform/overview/model/useCheatDetectionPage.ts`
- `code/frontend/src/features/platform/overview/model/platformOverviewRoutes.ts`
- `code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionSummaryPanel.vue`
- `code/frontend/src/pages/platform/__tests__/CheatDetection.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`

## Similar implementations found
- 当前 `CheatDetection` 已经部分进入 `features/platform/overview`，但只迁了 page model，UI 仍挂在 `components/platform/cheat/*`，形成 feature owner 与展示 owner 裂开的问题。
- `PlatformOverview` 本身只需要保留总览页自己的 route helper 与 dashboard workflow，不应继续承载独立的作弊检测工作台 owner。
- 同类已完成迁移的页面遵循“route page 只做组合，feature 自己持有 model + ui”的结构，不保留 `components/platform/*` 作为过渡层。

## Decision
refactor_existing

## Reason
- 这次目标不是简单搬路径，而是把 `CheatDetection` 收成独立的 `features/platform/cheat-detection` feature。
- `overview` 对 `cheat detection` 的 model owner 属于历史中间态；如果继续把 `useCheatDetectionPage` 留在 `overview`，即使 UI 迁走，边界仍然不完整。
- 旧 `components/platform/cheat/*` 没有继续保留的理由，应在迁移后直接删除，避免残留双入口。

## Files to modify
- `.harness/reuse-decisions/cheat-detection-feature-owner-cleanup.md`
- `code/frontend/src/pages/platform/CheatDetectionRoutePage.vue`
- `code/frontend/src/features/platform/overview/model/index.ts`
- `code/frontend/src/features/platform/overview/index.ts`
- `code/frontend/src/features/platform/overview/model/useCheatDetectionPage.ts`
- `code/frontend/src/features/platform/cheat-detection/index.ts`
- `code/frontend/src/features/platform/cheat-detection/model/index.ts`
- `code/frontend/src/features/platform/cheat-detection/model/cheatDetectionRoutes.ts`
- `code/frontend/src/features/platform/cheat-detection/model/useCheatDetectionPage.ts`
- `code/frontend/src/features/platform/cheat-detection/ui/index.ts`
- `code/frontend/src/features/platform/cheat-detection/ui/CheatDetectionWorkspacePanel.vue`
- `code/frontend/src/features/platform/cheat-detection/ui/CheatDetectionHeroPanel.vue`
- `code/frontend/src/features/platform/cheat-detection/ui/CheatDetectionReviewPanels.vue`
- `code/frontend/src/features/platform/cheat-detection/ui/CheatDetectionSummaryPanel.vue`
- `code/frontend/src/pages/platform/__tests__/CheatDetection.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/pages/platform/__tests__/journalPlatformShellStyles.test.ts`
- `code/frontend/src/pages/platform/__tests__/cheatDetectionSurfaceAlignment.test.ts`
- `code/frontend/src/pages/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/__tests__/journalNoteStyles.test.ts`
- `code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue`
- `code/frontend/src/components/platform/cheat/CheatDetectionSummaryPanel.vue`

## After implementation
- `CheatDetection` route page 只依赖 `features/platform/cheat-detection` 公共出口。
- `features/platform/overview` 只保留总览页相关 owner，不再导出 `useCheatDetectionPage`。
- 所有 raw-source style tests 改为读取 `features/platform/cheat-detection/ui/*`。
- `components/platform/cheat/*` 目录删除，不保留桥接壳。
