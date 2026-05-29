# Reuse Decision

## Change type
frontend refactor / feature router owner cleanup

## Existing code searched
- code/frontend/src/features/contest-awd-config/model/useAwdChallengeSelection.ts
- code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts

## Similar implementations found
- `useContestAwdConfigPage.ts` 已经是 AWD 配置 route 的 page owner，当前天然持有 `useRoute()`、`useRouter()` 和返回平台赛事工作台的导航动作。
- `useAwdChallengeSelection.ts` 当前主要负责服务选择与 query 对齐，router.replace 混在这里属于 selection helper 越权。

## Decision
refactor_existing

## Reason
`featureRouterImportAllowlist` 中，`features/contest-awd-config/model/useAwdChallengeSelection.ts -> vue-router` 不是合理长期例外。这个文件是 service selection owner，不应直接认识 `Router` 或 `RouteLocationNormalizedLoaded`。

最小正确改动是：

- 让 `useAwdChallengeSelection()` 改为消费 `readServiceQuery` / `replaceServiceQuery` callback
- 保留 `useContestAwdConfigPage.ts` 作为唯一 route-aware page owner
- 删除对应 allowlist 条目，并补 source guardrail，防止 router 再漂回 selection helper

本轮不做：

- 不改 `useContestAwdConfigPage.ts` 继续作为 route-aware page owner 的身份
- 不调整 AWD checker 草稿、预览、保存和数据加载 owner
- 不处理 `featureRouterImportAllowlist` 其它剩余条目

## Files to modify
- .harness/reuse-decisions/awd-challenge-selection-router-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-awd-challenge-selection-router-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-awd-challenge-selection-router-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/contest-awd-config/model/useAwdChallengeSelection.ts
- code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts
- code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts

## After implementation
- `useAwdChallengeSelection.ts` 不再 import `vue-router`
- AWD 服务选择的 query/router owner 明确回到 `useContestAwdConfigPage.ts`
- `featureRouterImportAllowlist` 缩小一条
