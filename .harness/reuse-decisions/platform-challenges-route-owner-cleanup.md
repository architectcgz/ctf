# Reuse Decision

## Change type
frontend refactor / platform challenges route owner cleanup

## Existing code searched
- code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts
- code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeWriteupPage.ts
- code/frontend/src/features/platform-challenges/model/useChallengeWriteupViewPage.ts
- code/frontend/src/features/platform-challenges/model/index.ts
- code/frontend/src/composables/routeNavigationTransport.ts
- code/frontend/src/composables/routeQueryTransport.ts
- code/frontend/src/components/navigation/routeTarget.ts
- code/frontend/src/features/class-students-workspace/model/classStudentsRoutes.ts
- code/frontend/src/features/student-analysis-workspace/model/studentAnalysisRoutes.ts
- code/frontend/src/features/student-dashboard/model/studentDashboardRoutes.ts
- code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts
- code/frontend/src/__tests__/architectureAllowlist.ts

## Similar implementations found
- `composables/routeNavigationTransport.ts`
- `composables/routeQueryTransport.ts`
- `features/class-students-workspace/model/classStudentsRoutes.ts`
- `features/student-analysis-workspace/model/studentAnalysisRoutes.ts`
- `features/student-dashboard/model/studentDashboardRoutes.ts`

## Decision
refactor_existing

## Reason
`platform-challenges` 目前有两类 router 职责混在 feature model 里：

- `useChallengeManagePage.ts` 直接拼 import preview、challenge detail、topology、writeup、import manage 几条薄导航
- `usePlatformChallengeRoutePage.ts` 既读取 `challengeId` params，又负责返回详情和跳转题解编辑页

如果只是在这两个文件外面再包一层 router helper，allowlist 虽然可能下降，但会留下一个“只是把 router 再包一层”的中间态，owner 反而更模糊。更合理的收口方式是：

- route `params` 读取下沉到共享 `routeQueryTransport`
- `push()` 下沉到共享 `routeNavigationTransport`
- `platform-challenges` 自己的 route target 落到本地 `platformChallengeRoutes.ts`
- challenge 管理页和 topology / writeup route page 继续保留自己的 workflow owner，不把业务语义塞进 shared

这样既能真实减少 `featureRouterImportAllowlist`，也能保持 `platform-challenges` 的页面路由语义仍由本 feature 自己描述。

## Files to modify
- .harness/reuse-decisions/platform-challenges-route-owner-cleanup.md
- docs/plan/impl-plan/2026-05-29-platform-challenges-route-owner-cleanup-plan.md
- docs/reviews/frontend/2026-05-29-platform-challenges-route-owner-cleanup-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/features/platform-challenges/model/index.ts
- code/frontend/src/features/platform-challenges/model/platformChallengeRoutes.ts
- code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts
- code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts
- code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts
- code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts

## After implementation
- `useChallengeManagePage.ts` 不再 import `vue-router`
- `usePlatformChallengeRoutePage.ts` 不再 import `vue-router`
- platform challenge 的详情、拓扑、题解、导入相关导航统一走本地 route target helper + shared navigation transport
- `challengeId` params 改由 `routeQueryTransport` 读取
- `featureRouterImportAllowlist` 再收掉 `platform-challenges` 这两条 router import
