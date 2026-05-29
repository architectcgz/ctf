# Reuse Decision

## Change type
frontend refactor / challenge list route owner cleanup

## Existing code searched
- code/frontend/src/features/challenge-list/model/useChallengeListPage.ts
- code/frontend/src/views/challenges/ChallengeList.vue
- code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts
- code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue
- code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue
- code/frontend/src/features/skill-profile/model/skillProfileRoutes.ts
- code/frontend/src/features/platform-user-management/model/useUserGovernancePanelRoute.ts
- code/frontend/src/composables/useRouteQueryTabs.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `features/skill-profile/model/skillProfileRoutes.ts`
- `features/challenge-package-import/model/challengeImportRoutes.ts`
- `features/platform-user-management/model/useUserGovernancePanelRoute.ts`
- `components/navigation/AppRouteLink.vue`

## Decision
refactor_existing

## Reason
`useChallengeListPage.ts` 当前混有两种 route owner：

- `category / difficulty` 的 query sync
- 返回仪表盘、进入能力画像、进入题目详情的 3 条薄导航

这条如果继续把完整 `vue-router` 留在 page model 里，allowlist 不会下降；如果只是把 `router` 平移到另一个 feature wrapper，也只是换壳不换 owner。更合理的做法是：

- 导航改成显式 route target contract，让 `ChallengeList.vue` 和目录行直接通过 `AppRouteLink` 消费
- query sync 继续留在 challenge list page owner，但把 `useRoute/useRouter` transport 下沉成共享 query transport，而不是继续散在 feature page model 内

这样能把 allowlist 真正减掉一条，同时保持筛选 query、刷新策略和数据加载仍由 challenge list page owner 自己负责。

## Files to modify
- .harness/reuse-decisions/challenge-list-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-challenge-list-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-challenge-list-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/features/challenge-list/model/challengeListRoutes.ts
- code/frontend/src/features/challenge-list/model/index.ts
- code/frontend/src/features/challenge-list/model/useChallengeListPage.ts
- code/frontend/src/views/challenges/ChallengeList.vue
- code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue
- code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue
- code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts

## After implementation
- `useChallengeListPage.ts` 不再 import `vue-router`
- challenge list 的 header 与目录行直接消费 route target
- `featureRouterImportAllowlist` 再收掉 `features/challenge-list/model/useChallengeListPage.ts -> vue-router`
