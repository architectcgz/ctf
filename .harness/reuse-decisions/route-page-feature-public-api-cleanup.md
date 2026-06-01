# Reuse Decision

## Change type
frontend refactor / route page public api cleanup

## Existing code searched
- `code/frontend/src/pages/teacher/TeacherDashboardRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformOverviewRoutePage.vue`
- `code/frontend/src/features/teacher/dashboard/index.ts`
- `code/frontend/src/features/teacher/dashboard/model/index.ts`
- `code/frontend/src/features/platform/overview/index.ts`
- `code/frontend/src/features/platform/overview/model/index.ts`
- `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `features/teacher/dashboard` 和 `features/platform/overview` 根出口都已经转发各自的 page model。
- 现有 route page 已普遍通过 feature 根入口消费 page owner，只剩少数页面还在直指 `model/use*Page.ts`。

## Decision
refactor_existing

## Reason
这次不是要重做 dashboard / overview 的 page model，而是把 route page 最后一层内部路径引用收干净。

当前 `TeacherDashboardRoutePage.vue` 和 `PlatformOverviewRoutePage.vue` 虽然行为上已经是薄壳，但仍直接 import `@/features/*/model/use*Page`。这会让 route page 绕过 feature public API，也让“page 只组合，feature 负责暴露 owner”这条边界缺少自动护栏。

最小正确改动是：

- 把两个 route page 改成只从各自 feature 根入口拿 page component 与 page model
- 在 `routePageArchitectureBoundary.test.ts` 新增一条窄护栏，禁止 route page 直接 import `@/features/*/(model|ui|lib|api|types)/*`
- 同步更新对应 raw-source 测试和 backlog 进展

本轮不做：

- 不改 `useDashboardPage()`、`usePlatformOverviewPage()` 的内部 workflow
- 不继续拆 teacher / platform dashboard 的其它 feature owner
- 不扩展到 widget / entity 的更宽公共出口规则

## Files to modify
- `.harness/reuse-decisions/route-page-feature-public-api-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-route-page-feature-public-api-cleanup-plan.md`
- `docs/reviews/frontend/2026-06-01-route-page-feature-public-api-cleanup-review.md`
- `code/frontend/src/pages/teacher/TeacherDashboardRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformOverviewRoutePage.vue`
- `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `TeacherDashboardRoutePage.vue` 与 `PlatformOverviewRoutePage.vue` 都只通过 feature 根入口消费 page owner。
- route page 层会多一条自动护栏，后续新页面不能再直接深导入 feature internal modules。
- backlog 中这条 route-page public API 残片会记为已收口进展，而不是继续混在更大的 dashboard owner 叙事里。
