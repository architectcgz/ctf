# Reuse Decision

## Change type
frontend refactor / router owner cleanup

## Existing code searched
- code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts
- code/frontend/src/features/challenge-package-import/model/index.ts
- code/frontend/src/features/challenge-package-import/index.ts
- code/frontend/src/views/platform/ChallengePackageFormat.vue
- code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useAwdReviewIndex.ts`、`useAuth.ts` 这类 wrapper 已改成纯 route target / page wrapper，不再保留只做一次 `router.push` 的通用 helper。
- 仓库的 route view boundary 允许 `RouterLink`，但不允许 view 直接 `useRouter()` / `router.push()`。

## Decision
refactor_existing

## Reason
`useChallengePackageFormatPage.ts` 当前只做一件事：点击“返回导入题目包”时执行一次 `router.push({ name: 'PlatformChallengeImportManage' })`。

这类一次性返回动作没有必要继续保留成 route-aware composable。更合理的边界是：

- feature 只暴露纯 route target contract
- route view 直接通过 `RouterLink` 消费这个 contract

这样可以直接收掉一条 allowlist，而不是仅仅移动 owner。

本轮不做：

- 不调整 `ChallengePackageFormatGuidePanel.vue` 的内容和样式
- 不处理 `useChallengeImportManagePage.ts` / `useChallengeImportPreviewPage.ts`
- 不改变“返回导入题目包”的目标路由

## Files to modify
- .harness/reuse-decisions/challenge-package-format-router-target-cleanup.md
- docs/plan/impl-plan/2026-05-29-challenge-package-format-router-target-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-challenge-package-format-router-target-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/challenge-package-import/model/useChallengePackageFormatPage.ts
- code/frontend/src/features/challenge-package-import/model/index.ts
- code/frontend/src/features/challenge-package-import/index.ts
- code/frontend/src/views/platform/ChallengePackageFormat.vue
- code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts

## After implementation
- `useChallengePackageFormatPage.ts` 不再 import `vue-router`
- `ChallengePackageFormat.vue` 通过 `RouterLink` + 纯 route target contract 返回导入页
- `featureRouterImportAllowlist` 收掉 `useChallengePackageFormatPage.ts` 这一条
