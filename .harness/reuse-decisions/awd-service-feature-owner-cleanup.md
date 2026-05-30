# Reuse Decision

## Change type
component / test / docs

## Existing code searched
- `code/frontend/src/components/platform/awd-service/AwdChallengeLibrarySection.vue`
- `code/frontend/src/components/platform/awd-service/AwdChallengeWorkspaceHeader.vue`
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- `code/frontend/src/features/platform/awd-challenges/ui/AWDChallengeLibraryPage.vue`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/components.d.ts`

## Similar implementations found
- `features/challenge-list/ui/ChallengeDirectoryPanel.vue` 与 `features/challenge-detail/ui/*` 已经承接“原本遗留在 components、但实际只服务单一 feature”的 owner 收口模式。
- `features/platform/awd-challenges/ui/AwdChallengeImportSection.vue` 已经是 `platform/awd-challenges` 内部子块，说明 AWD 题库页内部块继续留在 `components/platform/awd-service` 只是历史目录残留。

## Decision
refactor_existing

## Reason
- `AwdChallengeLibrarySection.vue` 与 `AwdChallengeWorkspaceHeader.vue` 的运行时 consumer 只有 `AWDChallengeLibraryPage.vue`，owner 已经清楚落在 `features/platform/awd-challenges/ui`。
- 它们不是跨 feature 共享组件，也不是 `entities/*` 稳定对象表达；继续留在 `components/platform/awd-service` 会保留历史中间态。
- 本轮只收这两个 feature-internal 子块，不扩展到其它 AWD 页面组件，保持改动面最小。

## Files to modify
- `.harness/reuse-decisions/awd-service-feature-owner-cleanup.md`
- `code/frontend/src/features/platform/awd-challenges/ui/AwdChallengeLibrarySection.vue`
- `code/frontend/src/features/platform/awd-challenges/ui/AwdChallengeWorkspaceHeader.vue`
- `code/frontend/src/features/platform/awd-challenges/ui/AWDChallengeLibraryPage.vue`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- `code/frontend/src/components.d.ts`
- `docs/plan/impl-plan/2026-05-30-awd-service-feature-owner-cleanup-plan.md`
- `docs/reviews/frontend/2026-05-30-awd-service-feature-owner-cleanup-review.md`

## After implementation
- `components/platform/awd-service/` 应不再保留这两个仅服务 `platform/awd-challenges` 的组件。
- `platform/awd-challenges` 的页面壳与内部子块将统一落在同一 feature owner 下。
