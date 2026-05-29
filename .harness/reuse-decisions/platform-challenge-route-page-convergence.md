# Reuse Decision

## Change type
frontend refactor / router owner convergence

## Existing code searched
- code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeWriteupPage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeWriteupViewPage.ts
- code/frontend/src/features/platform-challenges/model/index.ts
- code/frontend/src/features/platform-challenges/index.ts
- code/frontend/src/views/platform/ChallengeTopologyStudio.vue
- code/frontend/src/views/platform/ChallengeWriteup.vue
- code/frontend/src/views/platform/ChallengeWriteupView.vue
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `useAwdReviewIndexPage(scope)` 已作为显式 route-aware page wrapper，承接共享 feature 的角色路由。
- `useChallengePackageFormatPage.ts` 已从单次 `router.push` wrapper 收口成纯 route target contract。

## Decision
refactor_existing

## Reason
`platform-challenges/model` 里目前有三份极薄的 route wrapper：

- `useChallengeTopologyStudioRoutePage.ts`
- `useChallengeWriteupPage.ts`
- `useChallengeWriteupViewPage.ts`

它们重复持有：

- `useRoute()` 读取 `challengeId`
- `useRouter()` 返回题目详情
- 仅在 query / 额外跳转目标上有细微差别

最小正确改动是：

- 新增一个显式的 `usePlatformChallengeRoutePage(mode)` route-aware page wrapper
- 旧三个 public API 文件改为纯委托层，不再各自 import `vue-router`
- route view 继续保持现有 public API，不改调用面

这样可以在不删文件的前提下，把 3 条 `featureRouterImportAllowlist` 收拢到 1 条。

本轮不做：

- 不改 `ChallengeTopologyStudio.vue` / `ChallengeWriteup.vue` / `ChallengeWriteupView.vue` 的模板结构
- 不继续处理 `usePlatformChallengeDetailRoutePage.ts`
- 不动题解编辑 / 查看和拓扑工作台自身业务流程

## Files to modify
- .harness/reuse-decisions/platform-challenge-route-page-convergence.md
- docs/plan/impl-plan/2026-05-29-platform-challenge-route-page-convergence-plan.md
- docs/reviews/frontend/2026-05-29-platform-challenge-route-page-convergence-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeWriteupPage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeWriteupViewPage.ts
- code/frontend/src/features/platform-challenges/model/index.ts
- code/frontend/src/features/platform-challenges/index.ts
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts

## After implementation
- 只有 `usePlatformChallengeRoutePage.ts` 继续持有 `vue-router`
- 旧三个 route wrapper 变成纯委托层
- `featureRouterImportAllowlist` 从这组条目净减少 2 条
