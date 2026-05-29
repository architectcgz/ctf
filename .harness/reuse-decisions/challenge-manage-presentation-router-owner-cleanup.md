# Reuse Decision

## Change type
frontend refactor / feature router owner cleanup

## Existing code searched
- code/frontend/src/features/platform-challenges/model/useChallengeManagePresentation.ts
- code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts

## Similar implementations found
- `useChallengeManagePage.ts` 已经是这组平台题目管理的 route-aware page owner，本身持有 `useRouter()`。
- `useChallengeManagePresentation.ts` 当前只负责展示派生、菜单开关和动作编排，把导航动作回调化后更符合 presentation owner。

## Decision
refactor_existing

## Reason
`featureRouterImportAllowlist` 中，`features/platform-challenges/model/useChallengeManagePresentation.ts -> vue-router` 不是合理长期例外。这个文件的职责是 presentation / action menu owner，不应该直接认识 `Router`。

最小正确改动是：

- 让 `useChallengeManagePresentation()` 改为接收导航 callback，而不是 `Router`
- 保留 `useChallengeManagePage()` 作为 route-aware page owner
- 删除对应 allowlist 条目

本轮不做：

- 不改 `useChallengeManagePage()` 的 route owner 身份
- 不处理 `featureRouterImportAllowlist` 的其它条目
- 不改题目管理页的业务流程、API、排序或筛选逻辑

## Files to modify
- .harness/reuse-decisions/challenge-manage-presentation-router-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-challenge-manage-presentation-router-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-challenge-manage-presentation-router-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-challenges/model/useChallengeManagePresentation.ts
- code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts
- code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts

## After implementation
- `useChallengeManagePresentation.ts` 不再 import `vue-router`
- 题目管理这组导航动作明确回到 `useChallengeManagePage.ts`
- `featureRouterImportAllowlist` 缩小一条
