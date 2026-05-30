# Reuse Decision

## Change type
frontend architecture / feature-owner migration

## Existing code searched
- code/frontend/src/components/platform/challenge/ChallengeDescriptionPanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportHeroPanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportQueuePanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportUploadResultsPanel.vue
- code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue
- code/frontend/src/components/platform/challenge/ChallengePackageImportEntry.vue
- code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue
- code/frontend/src/pages/platform/challenges/ChallengeImportManageRoutePage.vue
- code/frontend/src/pages/platform/challenges/ChallengeImportPreviewRoutePage.vue
- code/frontend/src/pages/platform/challenges/ChallengePackageFormatRoutePage.vue
- code/frontend/src/features/platform/challenge-package-import/**
- code/frontend/src/features/platform/challenge-detail/ui/AdminChallengeProfilePanel.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupViewPage.vue
- code/frontend/src/features/platform/awd-challenges/ui/AwdChallengeImportSection.vue
- code/frontend/src/entities/challenge/**
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `contest-detail`、`contest-awd-workspace` 这两轮刚证明：单 route / 单 capability 的历史 page shell 应直接回到 owning feature 的 `ui/`。
- `ChallengeDescriptionPanel.vue` 不是单 feature 私有壳，而是被 challenge import preview、platform challenge detail、writeup view 三处共享的 challenge 内容展示块，更适合下沉到 `entities/challenge/ui`。
- `features/platform/challenge-package-import` 目前已有 page model public API，但还没有对应 `ui/` owner，这轮正好补齐。

## Decision
refactor_existing

## Reason
这轮把 `components/platform/challenge/*` 拆成两类 owner：

- 下沉到 `entities/challenge/ui`
  - `ChallengeDescriptionPanel.vue`
- 收口到 `features/platform/challenge-package-import/ui`
  - `ChallengeImportHeroPanel.vue`
  - `ChallengeImportPreviewWorkspacePanel.vue`
  - `ChallengeImportQueuePanel.vue`
  - `ChallengeImportUploadResultsPanel.vue`
  - `ChallengePackageFormatGuidePanel.vue`
  - `ChallengePackageImportEntry.vue`
  - `ChallengePackageImportReview.vue`

这样可以让：

- `ChallengeImportManageRoutePage.vue`、`ChallengeImportPreviewRoutePage.vue`、`ChallengePackageFormatRoutePage.vue` 全部从 `@/features/platform/challenge-package-import` 读取 UI
- `AdminChallengeProfilePanel.vue`、`ChallengeWriteupViewPage.vue`、`AwdChallengeImportSection.vue` 不再依赖 `components/platform/challenge/ChallengeDescriptionPanel.vue`
- `components/platform/challenge/*` 退出活动 owner

不做：

- 不改 challenge import / preview / format 的 page model
- 不改 platform challenge detail / writeup view 的行为
- 不碰 `components/platform/contest/*`

## Files to modify
- .harness/reuse-decisions/platform-challenge-feature-owner-migration-batch1.md
- docs/plan/impl-plan/2026-05-30-platform-challenge-feature-owner-migration-batch1-plan.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/components.d.ts
- code/frontend/src/components/platform/challenge/ChallengeDescriptionPanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportHeroPanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportQueuePanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportUploadResultsPanel.vue
- code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue
- code/frontend/src/components/platform/challenge/ChallengePackageImportEntry.vue
- code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue
- code/frontend/src/entities/challenge/index.ts
- code/frontend/src/entities/challenge/ui/ChallengeDescriptionPanel.vue
- code/frontend/src/features/platform/challenge-package-import/index.ts
- code/frontend/src/features/platform/challenge-package-import/model/challengeImportErrorSupport.ts
- code/frontend/src/features/platform/challenge-package-import/model/challengeImportRoutes.ts
- code/frontend/src/features/platform/challenge-package-import/model/challengeImportUploadFlow.ts
- code/frontend/src/features/platform/challenge-package-import/model/index.ts
- code/frontend/src/features/platform/challenge-package-import/model/useChallengeImportManagePage.ts
- code/frontend/src/features/platform/challenge-package-import/model/useChallengeImportPreviewPage.ts
- code/frontend/src/features/platform/challenge-package-import/model/useChallengePackageFormatPage.ts
- code/frontend/src/features/platform/challenge-package-import/model/useChallengePackageImport.test.ts
- code/frontend/src/features/platform/challenge-package-import/model/useChallengePackageImport.ts
- code/frontend/src/features/platform/challenge-package-import/model/useChallengePackageImportBoundary.test.ts
- code/frontend/src/features/platform/challenge-package-import/ui/index.ts
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengeImportHeroPanel.vue
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengeImportPreviewWorkspacePanel.vue
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengeImportQueuePanel.vue
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengeImportUploadResultsPanel.vue
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengePackageFormatGuidePanel.vue
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengePackageImportEntry.vue
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengePackageImportReview.test.ts
- code/frontend/src/features/platform/challenge-package-import/ui/ChallengePackageImportReview.vue
- code/frontend/src/pages/platform/challenges/ChallengeImportManageRoutePage.vue
- code/frontend/src/pages/platform/challenges/ChallengeImportPreviewRoutePage.vue
- code/frontend/src/pages/platform/challenges/ChallengePackageFormatRoutePage.vue
- code/frontend/src/pages/platform/challenges/__tests__/ChallengeImportManage.test.ts
- code/frontend/src/pages/platform/challenges/__tests__/ChallengePackageFormat.test.ts
- code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts
- code/frontend/src/features/platform/challenge-detail/ui/AdminChallengeProfilePanel.vue
- code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupViewPage.vue
- code/frontend/src/features/platform/awd-challenges/ui/AwdChallengeImportSection.vue

## After implementation
- `components/platform/challenge/*` 应不再保留活动运行时代码 owner。
- `features/platform/challenge-package-import/ui` 会成为导入页壳的唯一 owner。
- `ChallengeDescriptionPanel` 会作为 challenge 实体展示块，从 `entities/challenge` 对外暴露。
