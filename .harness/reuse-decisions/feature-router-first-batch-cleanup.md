# Reuse Decision

## Change type
frontend refactor / route target cleanup batch

## Existing code searched
- code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts
- code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts
- code/frontend/src/features/auth/model/useRegisterPage.ts
- code/frontend/src/features/challenge-package-import/model/useChallengeImportManagePage.ts
- code/frontend/src/features/challenge-package-import/model/useChallengeImportPreviewPage.ts
- code/frontend/src/views/scoreboard/ScoreboardDetail.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/views/auth/RegisterView.vue
- code/frontend/src/views/platform/ChallengeImportManage.vue
- code/frontend/src/views/platform/ChallengeImportPreview.vue
- code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts
- code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts
- code/frontend/src/views/auth/__tests__/RegisterView.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeImportPreview.test.ts
- code/frontend/src/router/routes/studentRoutes.ts
- code/frontend/src/router/routes/platformRoutes.ts
- code/frontend/src/router/routes/authRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/platform-contests/model/useContestOperationsPage.ts`
- `features/platform-contests/model/useContestEditPage.ts`
- `features/platform-contests/model/contestManageRoutes.ts`
- `components/navigation/AppRouteLink.vue`
- `components/navigation/AppRouteRedirect.vue`

## Decision
refactor_existing

## Reason
这批五条 allowlist 都是同一类低风险收口：

- `useScoreboardDetailPage.ts` 只是读取单个 route param
- `useAwdChallengeLibraryPage.ts` 只有一条“去导入页”的薄导航
- `useRegisterPage.ts` 只有注册成功后的跳转
- `useChallengeImportManagePage.ts` / `useChallengeImportPreviewPage.ts` 主要是导入页、预览页和题库目录之间的薄导航

当前仓库的主模式已经明确：

- route param 通过 route props 显式下沉
- 薄导航改成 route target contract，让 UI 直接消费 `AppRouteLink`
- mutation 成功后的跳转由独立 redirect transport 承接，而不是继续把 `router.push()` 留在 feature page model

因此这轮不再新增新的 router helper 桶或把 `router.push()` 平移到别的 feature model，而是直接按现有 route target / redirect transport 模式批量收口第一批低复杂度条目。

## Files to modify
- .harness/reuse-decisions/feature-router-first-batch-cleanup.md
- docs/plan/impl-plan/2026-05-29-feature-router-first-batch-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-feature-router-first-batch-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/router/routes/authRoutes.ts
- code/frontend/src/router/routes/studentRoutes.ts
- code/frontend/src/router/routes/platformRoutes.ts
- code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts
- code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts
- code/frontend/src/features/auth/model/useRegisterPage.ts
- code/frontend/src/features/challenge-package-import/model/index.ts
- code/frontend/src/features/challenge-package-import/model/challengeImportRoutes.ts
- code/frontend/src/features/challenge-package-import/model/useChallengeImportManagePage.ts
- code/frontend/src/features/challenge-package-import/model/useChallengeImportPreviewPage.ts
- code/frontend/src/views/scoreboard/ScoreboardDetail.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/views/auth/RegisterView.vue
- code/frontend/src/views/platform/ChallengeImportManage.vue
- code/frontend/src/views/platform/ChallengeImportPreview.vue
- code/frontend/src/components/platform/awd-service/AwdChallengeWorkspaceHeader.vue
- code/frontend/src/components/platform/challenge/ChallengeImportHeroPanel.vue
- code/frontend/src/components/platform/challenge/ChallengeImportQueuePanel.vue
- code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue
- code/frontend/src/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue
- code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue
- code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts
- code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts
- code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts
- code/frontend/src/views/auth/__tests__/RegisterView.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeImportPreview.test.ts

## After implementation
- `useScoreboardDetailPage.ts` 不再 import `vue-router`
- `useAwdChallengeLibraryPage.ts` 不再 import `vue-router`
- `useRegisterPage.ts` 不再 import `vue-router`
- `useChallengeImportManagePage.ts` 不再 import `vue-router`
- `useChallengeImportPreviewPage.ts` 不再 import `vue-router`
- 第一批五条 `featureRouterImportAllowlist` 完整收掉
